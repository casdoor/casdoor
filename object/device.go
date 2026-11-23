// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	DeviceTrustSignal   = "Signal"
	DeviceTrustRequired = "Required"

	DeviceStatusEnrolling = "Enrolling"
	DeviceStatusPending   = "Pending"
	DeviceStatusTrusted   = "Trusted"
	DeviceStatusBlocked   = "Blocked"

	DeviceAssertionHeader = "X-Casdoor-Device-Assertion"

	// SigninMethodDevice signs in with a device enrolled with the Casdoor device agent, without a password
	SigninMethodDevice = "Trusted device"

	deviceEnrollTimeout   = 15 * time.Minute
	deviceSignatureWindow = 2 * time.Minute
)

// DevicePosture is what the device agent reports about the security of the device
type DevicePosture struct {
	DiskEncrypted     bool   `json:"diskEncrypted"`
	FirewallEnabled   bool   `json:"firewallEnabled"`
	ScreenLockEnabled bool   `json:"screenLockEnabled"`
	OsName            string `json:"osName"`
	OsVersion         string `json:"osVersion"`
	Hostname          string `json:"hostname"`
	Serial            string `json:"serial"`
	ReportedTime      string `json:"reportedTime"`
}

// Device is a computer enrolled with the Casdoor device agent, bound to the user who approved its enrollment
type Device struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	User         string         `xorm:"varchar(100) index" json:"user"`
	Platform     string         `xorm:"varchar(100)" json:"platform"`
	AgentVersion string         `xorm:"varchar(100)" json:"agentVersion"`
	PublicKey    string         `xorm:"varchar(200)" json:"publicKey"`
	Posture      *DevicePosture `xorm:"json" json:"posture"`
	Status       string         `xorm:"varchar(20)" json:"status"`
	Managed      bool           `json:"managed"`
	IsCompliant  bool           `xorm:"-" json:"isCompliant"`
	Problems     []string       `xorm:"-" json:"problems"`

	EnrollCode       string `xorm:"varchar(100) index" json:"-"`
	EnrollExpireTime string `xorm:"varchar(100)" json:"-"`

	LastSeenTime string `xorm:"varchar(100)" json:"lastSeenTime"`
	LastIp       string `xorm:"varchar(100)" json:"lastIp"`
}

func (device *Device) GetId() string {
	return fmt.Sprintf("%s/%s", device.Owner, device.Name)
}

// DeviceCheck is what device trust says about the device a sign-in comes from
type DeviceCheck struct {
	Trusted      bool
	Untrusted    bool
	SatisfiesMfa bool
	Denied       string
	Device       *Device
}

// DeviceTrustPolicy is how an organization trusts the computers enrolled with the Casdoor device agent
type DeviceTrustPolicy struct {
	// "" for off, "Signal" for untrusted devices to count as a risk of adaptive risk, "Required" for only the trusted
	// devices to sign in
	Mode             string   `json:"mode"`
	Applications     []string `json:"applications"`
	RequireEncrypted bool     `json:"requireEncrypted"`
	RequireFirewall  bool     `json:"requireFirewall"`
	RequireLock      bool     `json:"requireLock"`
	AutoApprove      bool     `json:"autoApprove"`
	Passwordless     bool     `json:"passwordless"`
	SatisfiesMfa     bool     `json:"satisfiesMfa"`
}

type cachedDeviceTrustPolicy struct {
	policy    *DeviceTrustPolicy
	fetchTime time.Time
}

var deviceTrustPolicyCache sync.Map

func clearDeviceTrustPolicyCache(owner string) {
	deviceTrustPolicyCache.Delete(owner)
}

// GetDeviceTrustPolicy returns the device trust of the organization when it is on, cached for a minute
func GetDeviceTrustPolicy(owner string) *DeviceTrustPolicy {
	if owner == "" {
		return nil
	}
	var policy *DeviceTrustPolicy
	if value, ok := deviceTrustPolicyCache.Load(owner); ok && time.Since(value.(*cachedDeviceTrustPolicy).fetchTime) < time.Minute {
		policy = value.(*cachedDeviceTrustPolicy).policy
	} else {
		organization, err := LookupOrganization("admin", owner)
		if err == nil && organization != nil {
			policy = organization.DeviceTrust
		}
		deviceTrustPolicyCache.Store(owner, &cachedDeviceTrustPolicy{policy: policy, fetchTime: time.Now()})
	}
	if policy == nil || policy.Mode == "" {
		return nil
	}
	return policy
}

// getDeviceProblems lists what the posture of the device misses of the organization's requirements
func getDeviceProblems(policy *DeviceTrustPolicy, posture *DevicePosture) []string {
	problems := []string{}
	if posture == nil {
		return []string{"no posture reported"}
	}
	if policy.RequireEncrypted && !posture.DiskEncrypted {
		problems = append(problems, "disk not encrypted")
	}
	if policy.RequireFirewall && !posture.FirewallEnabled {
		problems = append(problems, "firewall off")
	}
	if policy.RequireLock && !posture.ScreenLockEnabled {
		problems = append(problems, "screen lock off")
	}
	return problems
}

func fillDeviceCompliance(devices ...*Device) {
	for _, device := range devices {
		policy := GetDeviceTrustPolicy(device.Owner)
		if policy == nil {
			device.Problems = []string{}
			device.IsCompliant = true
			continue
		}
		device.Problems = getDeviceProblems(policy, device.Posture)
		device.IsCompliant = len(device.Problems) == 0
	}
}

func GetDeviceCount(owner string, user string, field string, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	if user != "" {
		session = session.And("user = ?", user)
	}
	return session.Where("status <> ?", DeviceStatusEnrolling).Count(&Device{})
}

func GetPaginationDevices(owner string, user string, offset int, limit int, field string, value string, sortField string, sortOrder string) ([]*Device, error) {
	devices := []*Device{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	if user != "" {
		session = session.And("user = ?", user)
	}
	err := session.Where("status <> ?", DeviceStatusEnrolling).Find(&devices)
	if err != nil {
		return nil, err
	}
	fillDeviceCompliance(devices...)
	return devices, nil
}

func GetDevice(id string) (*Device, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	device := Device{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&device)
	if err != nil || !existed {
		return nil, err
	}
	fillDeviceCompliance(&device)
	return &device, nil
}

// UpdateDevice saves what an admin changes of a device: its name, status and whether it is managed
func UpdateDevice(id string, device *Device) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if !util.InSlice([]string{DeviceStatusPending, DeviceStatusTrusted, DeviceStatusBlocked}, device.Status) {
		return false, fmt.Errorf("unknown device status: %s", device.Status)
	}
	device.UpdatedTime = util.GetCurrentTime()
	affected, err := ormer.Engine.ID(core.PK{owner, name}).Cols("display_name", "status", "managed", "updated_time").Update(device)
	return affected != 0, err
}

func DeleteDevice(device *Device) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{device.Owner, device.Name}).Delete(&Device{})
	return affected != 0, err
}

// DeviceEnrollRequest is what the device agent sends to start its enrollment
type DeviceEnrollRequest struct {
	Organization string         `json:"organization"`
	PublicKey    string         `json:"publicKey"`
	DisplayName  string         `json:"displayName"`
	Platform     string         `json:"platform"`
	AgentVersion string         `json:"agentVersion"`
	Posture      *DevicePosture `json:"posture"`
}

func decodeDevicePublicKey(text string) (ed25519.PublicKey, error) {
	data, err := base64.StdEncoding.DecodeString(text)
	if err != nil || len(data) != ed25519.PublicKeySize {
		return nil, errors.New("the public key must be a base64 Ed25519 public key")
	}
	return data, nil
}

func newEnrollCode() string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	code := make([]byte, 0, 9)
	for i, b := range buf {
		if i == 4 {
			code = append(code, '-')
		}
		code = append(code, letters[int(b)%len(letters)])
	}
	return string(code)
}

// StartDeviceEnrollment registers a device that waits for a signed-in user to approve its enrollment code
func StartDeviceEnrollment(request *DeviceEnrollRequest, clientIp string) (*Device, error) {
	if GetDeviceTrustPolicy(request.Organization) == nil {
		return nil, fmt.Errorf("the organization: %s doesn't use device trust", request.Organization)
	}
	if _, err := decodeDevicePublicKey(request.PublicKey); err != nil {
		return nil, err
	}
	now := time.Now()
	if request.Posture != nil {
		request.Posture.ReportedTime = util.GetCurrentTime()
	}
	device := &Device{
		Owner:            request.Organization,
		Name:             util.GenerateId(),
		CreatedTime:      util.GetCurrentTime(),
		UpdatedTime:      util.GetCurrentTime(),
		DisplayName:      TruncateString(request.DisplayName, 100),
		Platform:         TruncateString(request.Platform, 100),
		AgentVersion:     TruncateString(request.AgentVersion, 100),
		PublicKey:        request.PublicKey,
		Posture:          request.Posture,
		Status:           DeviceStatusEnrolling,
		EnrollCode:       newEnrollCode(),
		EnrollExpireTime: util.Time2String(now.Add(deviceEnrollTimeout)),
		LastSeenTime:     util.GetCurrentTime(),
		LastIp:           clientIp,
	}
	_, err := ormer.Engine.Insert(device)
	if err != nil {
		return nil, err
	}
	// the enrollments that nobody approved are dropped
	_, _ = ormer.Engine.Where("status = ? and enroll_expire_time < ?", DeviceStatusEnrolling, util.Time2String(now.Add(-time.Hour))).Delete(&Device{})
	return device, nil
}

// GetDeviceEnrollment finds the device that waits for the enrollment code, for the user to see what they approve
func GetDeviceEnrollment(owner string, code string) (*Device, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) == 8 {
		code = code[:4] + "-" + code[4:]
	}
	device := Device{}
	existed, err := ormer.Engine.Where("owner = ? and enroll_code = ? and status = ?", owner, code, DeviceStatusEnrolling).Get(&device)
	if err != nil || !existed {
		return nil, err
	}
	if util.String2Time(device.EnrollExpireTime).Before(time.Now()) {
		return nil, nil
	}
	fillDeviceCompliance(&device)
	return &device, nil
}

// ApproveDeviceEnrollment binds the device to the user, it is trusted right away when the policy approves the
// compliant devices automatically, otherwise it waits for an admin
func ApproveDeviceEnrollment(user *User, code string) (*Device, error) {
	policy := GetDeviceTrustPolicy(user.Owner)
	if policy == nil {
		return nil, fmt.Errorf("the organization: %s doesn't use device trust", user.Owner)
	}
	device, err := GetDeviceEnrollment(user.Owner, code)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, errors.New("the enrollment code is wrong or has expired")
	}

	device.User = user.Name
	device.Status = DeviceStatusPending
	if policy.AutoApprove && len(getDeviceProblems(policy, device.Posture)) == 0 {
		device.Status = DeviceStatusTrusted
	}
	device.EnrollCode = ""
	device.UpdatedTime = util.GetCurrentTime()
	_, err = ormer.Engine.ID(core.PK{device.Owner, device.Name}).Cols("user", "status", "enroll_code", "updated_time").Update(device)
	if err != nil {
		return nil, err
	}
	fillDeviceCompliance(device)
	return device, nil
}

func getDeviceSignedMessage(parts ...string) []byte {
	return []byte("casdoor-device|" + strings.Join(parts, "|"))
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func verifyDeviceSignature(device *Device, message []byte, signature string) error {
	publicKey, err := decodeDevicePublicKey(device.PublicKey)
	if err != nil {
		return err
	}
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil || !ed25519.Verify(publicKey, message, sig) {
		return errors.New("the signature of the device is invalid")
	}
	return nil
}

func checkDeviceTimestamp(timestamp string) error {
	unix, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("the timestamp is invalid")
	}
	if diff := time.Since(time.Unix(unix, 0)); diff > deviceSignatureWindow || diff < -deviceSignatureWindow {
		return errors.New("the timestamp is out of date, check the clock of the device")
	}
	return nil
}

// AuthenticateDeviceRequest checks a request that the device agent signed with its key: the headers
// X-Casdoor-Device-Id ("owner/name"), X-Casdoor-Device-Timestamp and X-Casdoor-Device-Signature, which signs
// "casdoor-device|<id>|<timestamp>|<sha256 of the body>"
func AuthenticateDeviceRequest(request *http.Request, body []byte) (*Device, error) {
	id := request.Header.Get("X-Casdoor-Device-Id")
	timestamp := request.Header.Get("X-Casdoor-Device-Timestamp")
	if err := checkDeviceTimestamp(timestamp); err != nil {
		return nil, err
	}
	device, err := GetDevice(id)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, errors.New("the device is not enrolled")
	}
	message := getDeviceSignedMessage(id, timestamp, sha256Hex(body))
	if err = verifyDeviceSignature(device, message, request.Header.Get("X-Casdoor-Device-Signature")); err != nil {
		return nil, err
	}
	return device, nil
}

// ReportDevicePosture saves the posture that an enrolled device reports
func ReportDevicePosture(device *Device, posture *DevicePosture, agentVersion string, clientIp string) error {
	if posture != nil {
		posture.ReportedTime = util.GetCurrentTime()
		device.Posture = posture
	}
	if agentVersion != "" {
		device.AgentVersion = TruncateString(agentVersion, 100)
	}
	device.LastSeenTime = util.GetCurrentTime()
	device.LastIp = clientIp
	_, err := ormer.Engine.ID(core.PK{device.Owner, device.Name}).Cols("posture", "agent_version", "last_seen_time", "last_ip").Update(device)
	return err
}

var (
	deviceChallenges = map[string]time.Time{}
	deviceLock       sync.Mutex
)

// NewDeviceChallenge returns a one-time nonce that the device agent signs to prove the device a sign-in comes from
func NewDeviceChallenge() string {
	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	nonce := base64.RawURLEncoding.EncodeToString(buf)

	deviceLock.Lock()
	defer deviceLock.Unlock()
	now := time.Now()
	for key, expire := range deviceChallenges {
		if now.After(expire) {
			delete(deviceChallenges, key)
		}
	}
	deviceChallenges[nonce] = now.Add(deviceSignatureWindow)
	return nonce
}

func consumeDeviceChallenge(nonce string) bool {
	deviceLock.Lock()
	defer deviceLock.Unlock()
	expire, ok := deviceChallenges[nonce]
	delete(deviceChallenges, nonce)
	return ok && time.Now().Before(expire)
}

// DeviceAssertion is what the device agent signs for the sign-in page, sent base64 JSON in the header
// X-Casdoor-Device-Assertion. The signature covers "casdoor-device|<id>|<origin>|<nonce>|<timestamp>|<sha256 of
// the posture JSON>"
type DeviceAssertion struct {
	Device    string `json:"device"`
	Origin    string `json:"origin"`
	Nonce     string `json:"nonce"`
	Timestamp string `json:"timestamp"`
	Posture   string `json:"posture"`
	Signature string `json:"signature"`
}

type verifiedAssertion struct {
	device *Device
	err    error
	time   time.Time
}

// verifiedAssertions keeps the result per request, the nonce of an assertion can be used only once but a sign-in
// checks its device more than once
var verifiedAssertions sync.Map

// verifyDeviceAssertion returns the device that signed the assertion of the request, nil without an assertion
func verifyDeviceAssertion(request *http.Request) (*Device, error) {
	if value, ok := verifiedAssertions.Load(request); ok {
		result := value.(*verifiedAssertion)
		return result.device, result.err
	}
	device, err := doVerifyDeviceAssertion(request)
	now := time.Now()
	verifiedAssertions.Range(func(key, value interface{}) bool {
		if now.Sub(value.(*verifiedAssertion).time) > time.Minute {
			verifiedAssertions.Delete(key)
		}
		return true
	})
	if request.Header.Get(DeviceAssertionHeader) != "" {
		verifiedAssertions.Store(request, &verifiedAssertion{device: device, err: err, time: now})
	}
	return device, err
}

func doVerifyDeviceAssertion(request *http.Request) (*Device, error) {
	header := request.Header.Get(DeviceAssertionHeader)
	if header == "" {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, errors.New("the device assertion is invalid")
	}
	assertion := DeviceAssertion{}
	if err = json.Unmarshal(data, &assertion); err != nil {
		return nil, errors.New("the device assertion is invalid")
	}
	if !consumeDeviceChallenge(assertion.Nonce) {
		return nil, errors.New("the device challenge is unknown or has expired")
	}
	if err = checkDeviceTimestamp(assertion.Timestamp); err != nil {
		return nil, err
	}
	// the agent answers only the origins of the Casdoor it enrolled with, the browser tells which page asked it
	if origin := request.Header.Get("Origin"); origin != "" && !strings.EqualFold(strings.TrimRight(origin, "/"), strings.TrimRight(assertion.Origin, "/")) {
		return nil, errors.New("the device assertion is for another site")
	}

	device, err := GetDevice(assertion.Device)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, errors.New("the device is not enrolled")
	}
	message := getDeviceSignedMessage(assertion.Device, assertion.Origin, assertion.Nonce, assertion.Timestamp, sha256Hex([]byte(assertion.Posture)))
	if err = verifyDeviceSignature(device, message, assertion.Signature); err != nil {
		return nil, err
	}

	posture := &DevicePosture{}
	if assertion.Posture != "" && json.Unmarshal([]byte(assertion.Posture), posture) == nil {
		_ = ReportDevicePosture(device, posture, "", util.GetClientIpFromRequest(request))
	}
	fillDeviceCompliance(device)
	return device, nil
}

func isDeviceTrustedFor(device *Device, user *User) bool {
	return device != nil && device.Owner == user.Owner && device.User == user.Name && device.Status == DeviceStatusTrusted && device.IsCompliant
}

// CheckSigninDevice tells whether the sign-in of the user comes from a trusted device, and refuses it when the
// application needs one. An organization admin is never locked out by it, the device only counts as a risk then
func CheckSigninDevice(user *User, application string, request *http.Request) *DeviceCheck {
	check := &DeviceCheck{}
	if user == nil {
		return check
	}
	policy := GetDeviceTrustPolicy(user.Owner)
	if policy == nil {
		return check
	}

	device, err := verifyDeviceAssertion(request)
	if err != nil {
		device = nil
	}
	check.Device = device
	check.Trusted = isDeviceTrustedFor(device, user)
	check.SatisfiesMfa = check.Trusted && policy.SatisfiesMfa

	applies := len(policy.Applications) == 0 || application == "" || util.InSlice(policy.Applications, application)
	if !applies || check.Trusted {
		return check
	}
	check.Untrusted = true
	if policy.Mode == DeviceTrustRequired && !user.IsAdmin && !user.IsGlobalAdmin() {
		check.Denied = "auth:This application can only be signed in to from a trusted device, install the Casdoor device agent and enroll this device"
	}
	return check
}

// CheckDeviceSignin signs in the user that the trusted device of the request is bound to, without a password
func CheckDeviceSignin(request *http.Request, organization string) (*User, error) {
	policy := GetDeviceTrustPolicy(organization)
	if policy == nil || !policy.Passwordless {
		return nil, errors.New("signing in with a trusted device is not enabled for the organization")
	}
	device, err := verifyDeviceAssertion(request)
	if err != nil {
		return nil, err
	}
	if device == nil || device.Owner != organization || device.User == "" {
		return nil, errors.New("this device is not enrolled to the organization")
	}
	if device.Status != DeviceStatusTrusted {
		return nil, errors.New("this device is not trusted yet, ask your administrator to approve it")
	}
	if !device.IsCompliant {
		return nil, fmt.Errorf("this device doesn't meet the security requirements: %s", strings.Join(device.Problems, ", "))
	}
	user, err := GetUser(util.GetId(device.Owner, device.User))
	if err != nil {
		return nil, err
	}
	if user == nil || user.IsDeleted || user.IsForbidden {
		return nil, errors.New("the user of the device is forbidden to sign in")
	}
	return user, nil
}
