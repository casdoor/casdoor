// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const (
	RiskDecisionAllow = "Allow"
	RiskDecisionMfa   = "MFA"
	RiskDecisionDeny  = "Deny"

	RiskResultSucceeded = "Succeeded"

	RiskSignalTrustedNetwork   = "trusted-network"
	RiskSignalBlockedCountry   = "blocked-country"
	RiskSignalNewDevice        = "new-device"
	RiskSignalNewLocation      = "new-location"
	RiskSignalNewCountry       = "new-country"
	RiskSignalImpossibleTravel = "impossible-travel"
	RiskSignalIpFailures       = "ip-failures"
	RiskSignalUserFailures     = "user-failures"
	RiskSignalSuspiciousClient = "suspicious-client"
	RiskSignalUntrustedDevice  = "untrusted-device"
	RiskSignalFlow             = "flow"

	defaultRiskMfaScore          = 40
	defaultRiskDenyScore         = 100
	defaultImpossibleTravelHours = 2

	riskFailureWindow = 30 * time.Minute
)

// DefaultRiskScores are the scores of the signals that the organization's policy doesn't set
var DefaultRiskScores = map[string]int{
	RiskSignalBlockedCountry:   100,
	RiskSignalNewDevice:        20,
	RiskSignalNewLocation:      30,
	RiskSignalNewCountry:       40,
	RiskSignalImpossibleTravel: 70,
	RiskSignalIpFailures:       40,
	RiskSignalUserFailures:     30,
	RiskSignalSuspiciousClient: 20,
	RiskSignalUntrustedDevice:  40,
}

var suspiciousClientRegex = regexp.MustCompile(`(?i)curl/|wget/|python-requests|python-urllib|aiohttp|go-http-client|libwww-perl|scrapy|httpie|headlesschrome|phantomjs|java/\d`)

// SigninRiskEvent is a sign-in that adaptive risk found a signal in, with what it decided
type SigninRiskEvent struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100) index" json:"createdTime"`

	User        string   `xorm:"varchar(100) index" json:"user"`
	Application string   `xorm:"varchar(100)" json:"application"`
	Ip          string   `xorm:"varchar(100)" json:"ip"`
	Location    string   `xorm:"varchar(200)" json:"location"`
	Country     string   `xorm:"varchar(100)" json:"country"`
	DeviceId    string   `xorm:"varchar(200)" json:"deviceId"`
	UserAgent   string   `xorm:"varchar(500)" json:"userAgent"`
	Score       int      `json:"score"`
	Signals     []string `xorm:"mediumtext" json:"signals"`
	Decision    string   `xorm:"varchar(20)" json:"decision"`
	Result      string   `xorm:"varchar(20)" json:"result"`
}

// RiskInput is what a sign-in tells about where it comes from
type RiskInput struct {
	User            *object.User
	Application     string
	Ip              string
	DeviceId        string
	UserAgent       string
	Header          http.Header
	TrustedDevice   bool
	UntrustedDevice bool
}

// SigninAssessment is the risk adaptive risk found in a sign-in
type SigninAssessment struct {
	Score     int      `json:"score"`
	Signals   []string `json:"signals"`
	Decision  string   `json:"decision"`
	Location  string   `json:"location"`
	Country   string   `json:"country"`
	EventName string   `json:"eventName"`
}

// ToRisk is the assessment as the flows see it
func (assessment *SigninAssessment) ToRisk() *object.SigninRisk {
	if assessment == nil {
		return nil
	}
	return &object.SigninRisk{Score: assessment.Score, Signals: assessment.Signals, Decision: assessment.Decision, Location: assessment.Location}
}

func (assessment *SigninAssessment) HasSignal(signal string) bool {
	return assessment != nil && util.InSlice(assessment.Signals, signal)
}

// GetAdaptiveRiskPolicy returns the organization's security policy when its adaptive risk is on
func GetAdaptiveRiskPolicy(owner string) *object.SecurityPolicy {
	policy := GetCachedSecurityPolicy(owner)
	if policy == nil || !policy.AdaptiveRisk {
		return nil
	}
	return policy
}

func getRiskScore(policy *object.SecurityPolicy, signal string) int {
	if score, ok := policy.RiskScores[signal]; ok {
		return score
	}
	return DefaultRiskScores[signal]
}

func getRiskThresholds(policy *object.SecurityPolicy) (int, int) {
	mfaScore, denyScore := policy.RiskMfaScore, policy.RiskDenyScore
	if mfaScore <= 0 {
		mfaScore = defaultRiskMfaScore
	}
	if denyScore <= 0 {
		denyScore = defaultRiskDenyScore
	}
	return mfaScore, denyScore
}

// getIpCountry names the country of an IP from the configured IP database, or from the header a CDN in front of
// Casdoor adds (such as CF-IPCountry) when the policy names one
func getIpCountry(clientIp string, header http.Header, policy *object.SecurityPolicy) string {
	if country := object.GetIpDbCountry(clientIp); country != "" {
		return country
	}
	parsed := net.ParseIP(clientIp)
	if parsed == nil || parsed.IsLoopback() || parsed.IsPrivate() {
		return ""
	}
	if policy != nil && policy.RiskCountryHeader != "" && header != nil {
		country := strings.ToUpper(strings.TrimSpace(header.Get(policy.RiskCountryHeader)))
		if country != "" && country != "XX" && country != "T1" {
			return country
		}
	}
	return ""
}

func isCountryBlocked(country string, blocked []string) bool {
	for _, item := range blocked {
		if item != "" && strings.EqualFold(strings.TrimSpace(item), country) {
			return true
		}
	}
	return false
}

type riskFailureRecord struct {
	ipAccounts   map[string]map[string]time.Time
	userFailures map[string][]time.Time
}

var (
	riskFailures = map[string]*riskFailureRecord{}
	riskLock     sync.Mutex
)

// RecordSigninFailure remembers a failed password of the user from the IP, for the failure signals of adaptive risk
func RecordSigninFailure(organization string, clientIp string, username string) {
	if GetAdaptiveRiskPolicy(organization) == nil || username == "" {
		return
	}

	riskLock.Lock()
	defer riskLock.Unlock()
	record := riskFailures[organization]
	if record == nil {
		record = &riskFailureRecord{ipAccounts: map[string]map[string]time.Time{}, userFailures: map[string][]time.Time{}}
		riskFailures[organization] = record
	}
	now := time.Now()
	pruneRiskFailures(record, now)

	if clientIp != "" {
		if record.ipAccounts[clientIp] == nil {
			record.ipAccounts[clientIp] = map[string]time.Time{}
		}
		record.ipAccounts[clientIp][username] = now
	}
	record.userFailures[username] = append(record.userFailures[username], now)
}

func pruneRiskFailures(record *riskFailureRecord, now time.Time) {
	for clientIp, accounts := range record.ipAccounts {
		for name, t := range accounts {
			if now.Sub(t) > riskFailureWindow {
				delete(accounts, name)
			}
		}
		if len(accounts) == 0 {
			delete(record.ipAccounts, clientIp)
		}
	}
	for name, times := range record.userFailures {
		kept := times[:0]
		for _, t := range times {
			if now.Sub(t) <= riskFailureWindow {
				kept = append(kept, t)
			}
		}
		if len(kept) == 0 {
			delete(record.userFailures, name)
		} else {
			record.userFailures[name] = kept
		}
	}
}

func getRiskFailureCounts(organization string, clientIp string, username string) (int, int) {
	riskLock.Lock()
	defer riskLock.Unlock()
	record := riskFailures[organization]
	if record == nil {
		return 0, 0
	}
	pruneRiskFailures(record, time.Now())
	return len(record.ipAccounts[clientIp]), len(record.userFailures[username])
}

// AssessSignin scores a sign-in under the adaptive risk of the user's organization, returns nil when it is off
func AssessSignin(input *RiskInput) *SigninAssessment {
	user := input.User
	if user == nil {
		return nil
	}
	policy := GetAdaptiveRiskPolicy(user.Owner)
	if policy == nil {
		return nil
	}

	assessment := &SigninAssessment{
		Location: getIpLocation(input.Ip),
		Country:  getIpCountry(input.Ip, input.Header, policy),
		Signals:  []string{},
	}

	if object.IsIpInNetworks(input.Ip, policy.RiskTrustedNetworks) {
		assessment.Signals = append(assessment.Signals, RiskSignalTrustedNetwork)
		assessment.Decision = RiskDecisionAllow
		return assessment
	}

	add := func(signal string) {
		assessment.Signals = append(assessment.Signals, signal)
		assessment.Score += getRiskScore(policy, signal)
	}

	if assessment.Country != "" && isCountryBlocked(assessment.Country, policy.RiskBlockedCountries) {
		add(RiskSignalBlockedCountry)
	}

	traces, err := getSigninTraces(user.Owner, user.Name)
	if err != nil {
		logs.Warn("the sign-in traces of the user: %s are not read: %s", user.GetId(), err.Error())
	}
	hasKind := map[string]bool{}
	known := map[string]bool{}
	var lastCountry *SigninTrace
	for _, trace := range traces {
		hasKind[trace.Kind] = true
		known[trace.Kind+"/"+trace.TraceKey] = true
		if trace.Kind == "country" && (lastCountry == nil || trace.LastSeenTime > lastCountry.LastSeenTime) {
			lastCountry = trace
		}
	}

	deviceId := object.TruncateString(input.DeviceId, 200)
	if !input.TrustedDevice && deviceId != "" && hasKind["device"] && !known["device/"+deviceId] {
		add(RiskSignalNewDevice)
	}
	if assessment.Location != "" && assessment.Location != "intranet" && hasKind["location"] && !known["location/"+assessment.Location] {
		add(RiskSignalNewLocation)
	}
	if assessment.Country != "" && hasKind["country"] && !known["country/"+assessment.Country] {
		add(RiskSignalNewCountry)
	}
	if assessment.Country != "" && lastCountry != nil && lastCountry.TraceKey != assessment.Country {
		hours := policy.ImpossibleTravelHours
		if hours <= 0 {
			hours = defaultImpossibleTravelHours
		}
		if lastSeen := util.String2Time(lastCountry.LastSeenTime); time.Since(lastSeen) < time.Duration(hours)*time.Hour {
			add(RiskSignalImpossibleTravel)
		}
	}

	ipAccounts, userFailures := getRiskFailureCounts(user.Owner, input.Ip, user.Name)
	if ipAccounts >= 3 {
		add(RiskSignalIpFailures)
	}
	if userFailures >= 3 {
		add(RiskSignalUserFailures)
	}

	if !input.TrustedDevice && suspiciousClientRegex.MatchString(input.UserAgent) {
		add(RiskSignalSuspiciousClient)
	}
	if input.UntrustedDevice {
		add(RiskSignalUntrustedDevice)
	}
	if assessment.Score > 100 {
		assessment.Score = 100
	}
	mfaScore, denyScore := getRiskThresholds(policy)
	switch {
	case assessment.Score >= denyScore:
		assessment.Decision = RiskDecisionDeny
	case assessment.Score >= mfaScore:
		assessment.Decision = RiskDecisionMfa
	default:
		assessment.Decision = RiskDecisionAllow
	}

	if assessment.Score > 0 {
		event := &SigninRiskEvent{
			Owner:       user.Owner,
			Name:        util.GenerateId(),
			CreatedTime: util.GetCurrentTime(),
			User:        user.Name,
			Application: input.Application,
			Ip:          input.Ip,
			Location:    assessment.Location,
			Country:     assessment.Country,
			DeviceId:    deviceId,
			UserAgent:   object.TruncateString(input.UserAgent, 500),
			Score:       assessment.Score,
			Signals:     assessment.Signals,
			Decision:    assessment.Decision,
		}
		if _, err = object.GetOrmer().Engine.Insert(event); err != nil {
			logs.Warn("the sign-in risk event of the user: %s is not saved: %s", user.GetId(), err.Error())
		} else {
			assessment.EventName = event.Name
		}
	}
	return assessment
}

// MarkSigninRiskByFlow records that a pre-sign-in flow decided the sign-in, on the risk event or a new one
func MarkSigninRiskByFlow(owner string, assessment *SigninAssessment, decision string) {
	if assessment == nil {
		return
	}
	if !util.InSlice(assessment.Signals, RiskSignalFlow) {
		assessment.Signals = append(assessment.Signals, RiskSignalFlow)
	}
	assessment.Decision = decision
	if assessment.EventName == "" {
		return
	}
	_, err := object.GetOrmer().Engine.Cols("decision", "signals").
		Update(&SigninRiskEvent{Decision: decision, Signals: assessment.Signals}, &SigninRiskEvent{Owner: owner, Name: assessment.EventName})
	if err != nil {
		logs.Warn("the sign-in risk event: %s/%s is not updated: %s", owner, assessment.EventName, err.Error())
	}
}

// CompleteSigninRiskEvent marks the risk event of a sign-in as signed in, after its MFA when it asked for one
func CompleteSigninRiskEvent(owner string, name string, application string) {
	if owner == "" || name == "" {
		return
	}
	_, err := object.GetOrmer().Engine.Cols("result", "application").
		Update(&SigninRiskEvent{Result: RiskResultSucceeded, Application: application}, &SigninRiskEvent{Owner: owner, Name: name})
	if err != nil {
		logs.Warn("the sign-in risk event: %s/%s is not updated: %s", owner, name, err.Error())
	}
}

func GetSigninRiskEventCount(owner string, field string, value string) (int64, error) {
	session := object.GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&SigninRiskEvent{})
}

func GetPaginationSigninRiskEvents(owner string, offset int, limit int, field string, value string, sortField string, sortOrder string) ([]*SigninRiskEvent, error) {
	events := []*SigninRiskEvent{}
	if sortField == "" {
		sortField, sortOrder = "created_time", "descend"
	}
	session := object.GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&events)
	return events, err
}

// SigninRiskSummary counts the risk events of the last days by decision and by signal
type SigninRiskSummary struct {
	Days      int            `json:"days"`
	Total     int            `json:"total"`
	Decisions map[string]int `json:"decisions"`
	Signals   map[string]int `json:"signals"`
	TopUsers  []*RiskCount   `json:"topUsers"`
	TopIps    []*RiskCount   `json:"topIps"`
}

type RiskCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func GetSigninRiskSummary(owner string, days int) (*SigninRiskSummary, error) {
	if days <= 0 {
		days = 7
	}
	since := util.Time2String(time.Now().AddDate(0, 0, -days))
	events := []*SigninRiskEvent{}
	err := object.GetOrmer().Engine.Where("owner = ? and created_time >= ?", owner, since).Find(&events)
	if err != nil {
		return nil, err
	}

	summary := &SigninRiskSummary{Days: days, Total: len(events), Decisions: map[string]int{}, Signals: map[string]int{}}
	users := map[string]int{}
	ips := map[string]int{}
	for _, event := range events {
		summary.Decisions[event.Decision]++
		for _, signal := range event.Signals {
			summary.Signals[signal]++
		}
		if event.Decision != RiskDecisionAllow {
			users[event.User]++
			ips[event.Ip]++
		}
	}
	summary.TopUsers = topRiskCounts(users, 5)
	summary.TopIps = topRiskCounts(ips, 5)
	return summary, nil
}

func topRiskCounts(counts map[string]int, limit int) []*RiskCount {
	res := []*RiskCount{}
	for name, count := range counts {
		if name != "" {
			res = append(res, &RiskCount{Name: name, Count: count})
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Count != res[j].Count {
			return res[i].Count > res[j].Count
		}
		return res[i].Name < res[j].Name
	})
	if len(res) > limit {
		res = res[:limit]
	}
	return res
}
