// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/robfig/cron/v3"
)

// GetSecurityPolicy returns the organization's security policy, or nil when there is none or the license
// gives a lower edition than the commercial one
func GetSecurityPolicy(organization *object.Organization) *object.SecurityPolicy {
	if organization == nil || organization.SecurityPolicy == nil || !edition.Has(edition.Commercial) {
		return nil
	}
	return organization.SecurityPolicy
}

func GetSecurityPolicyByOwner(owner string) *object.SecurityPolicy {
	if !edition.Has(edition.Commercial) {
		return nil
	}
	organization, err := object.LookupOrganization("admin", owner)
	if err != nil {
		return nil
	}
	return GetSecurityPolicy(organization)
}

type cachedSecurityPolicy struct {
	policy    *object.SecurityPolicy
	fetchTime time.Time
}

var securityPolicyCache sync.Map

// GetCachedSecurityPolicy is GetSecurityPolicyByOwner cached for a minute, for the checks on every request
func GetCachedSecurityPolicy(owner string) *object.SecurityPolicy {
	if !edition.Has(edition.Commercial) || owner == "" {
		return nil
	}
	if value, ok := securityPolicyCache.Load(owner); ok {
		cached := value.(*cachedSecurityPolicy)
		if time.Since(cached.fetchTime) < time.Minute {
			return cached.policy
		}
	}
	policy := GetSecurityPolicyByOwner(owner)
	securityPolicyCache.Store(owner, &cachedSecurityPolicy{policy: policy, fetchTime: time.Now()})
	return policy
}

// getOrganizationEmailProvider picks the email provider of the organization's default application, then any
// email provider owned by the organization, then the default Email provider of the global setting
func getOrganizationEmailProvider(organization *object.Organization) (*object.Provider, error) {
	if organization.DefaultApplication != "" {
		application, err := object.LookupApplication("admin", organization.DefaultApplication)
		if err != nil {
			return nil, err
		}
		if application != nil {
			provider, err := application.GetProviderByCategoryAndRule("Email", "", "All")
			if err != nil {
				return nil, err
			}
			if provider != nil {
				return provider, nil
			}
		}
	}

	providers := []*object.Provider{}
	err := object.GetOrmer().Engine.Where("category = ? and (owner = ? or owner = ?)", "Email", organization.Name, "admin").Find(&providers)
	if err != nil {
		return nil, err
	}
	for _, provider := range providers {
		if provider.Owner == organization.Name {
			return provider, nil
		}
	}

	provider, err := object.GetDefaultEmailProvider()
	if err != nil || provider != nil {
		return provider, err
	}
	if len(providers) > 0 {
		return providers[0], nil
	}
	return nil, nil
}

func sendOrganizationEmail(organization *object.Organization, title string, content string, dest []string) error {
	if len(dest) == 0 {
		return nil
	}

	provider, err := getOrganizationEmailProvider(organization)
	if err != nil {
		return err
	}
	if provider == nil {
		return fmt.Errorf("the organization: %s has no email provider", organization.Name)
	}

	sender := organization.DisplayName
	if sender == "" {
		sender = organization.Name
	}
	return object.SendEmail(organization.Name, provider, title, content, dest, sender)
}

// SendSecurityAlert emails the organization's alert addresses
func SendSecurityAlert(organization *object.Organization, title string, content string) {
	policy := GetSecurityPolicy(organization)
	if policy == nil || len(policy.AlertEmails) == 0 {
		return
	}

	err := sendOrganizationEmail(organization, title, content, policy.AlertEmails)
	if err != nil {
		logs.Warn("security alert of the organization: %s is not sent: %s", organization.Name, err.Error())
	}
}

func parseAnyTime(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05.999999", "2006-01-02T15:04:05", "2006-01-02 15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// DisableInactiveUsers forbids the users who haven't signed in for the organization's DisableInactiveUserDays,
// counting from the creation time for the users who never signed in. Administrators are kept.
func DisableInactiveUsers() error {
	organizations, err := object.GetOrganizations("admin")
	if err != nil {
		return err
	}

	for _, organization := range organizations {
		policy := GetSecurityPolicy(organization)
		if policy == nil || policy.DisableInactiveUserDays <= 0 {
			continue
		}

		cutoff := time.Now().Add(-time.Duration(policy.DisableInactiveUserDays) * 24 * time.Hour)
		users := []*object.User{}
		err = object.GetOrmer().Engine.Cols("owner", "name", "created_time", "last_signin_time", "is_admin").
			Where("owner = ? and is_forbidden = ? and is_deleted = ? and is_admin = ?", organization.Name, false, false, false).Find(&users)
		if err != nil {
			return err
		}

		names := []string{}
		for _, user := range users {
			lastTime, ok := parseAnyTime(user.LastSigninTime)
			if !ok {
				lastTime, ok = parseAnyTime(user.CreatedTime)
			}
			if !ok || lastTime.After(cutoff) {
				continue
			}

			_, err = object.GetOrmer().Engine.Where("owner = ? and name = ?", user.Owner, user.Name).Cols("is_forbidden", "updated_time").
				Update(&object.User{IsForbidden: true, UpdatedTime: util.GetCurrentTime()})
			if err != nil {
				return err
			}
			names = append(names, user.Name)
		}

		if len(names) > 0 {
			logs.Info("disabled %d users of the organization: %s who haven't signed in for %d days", len(names), organization.Name, policy.DisableInactiveUserDays)
			shown := names
			if len(shown) > 50 {
				shown = shown[:50]
			}
			SendSecurityAlert(organization, fmt.Sprintf("[%s] %d inactive accounts disabled", organization.Name, len(names)),
				fmt.Sprintf("These accounts haven't signed in for %d days and have been disabled: %s", policy.DisableInactiveUserDays, strings.Join(shown, ", ")))
		}
	}
	return nil
}

func runSecurityJob(name string, job func() error) {
	if !edition.Has(edition.Commercial) {
		return
	}
	if err := job(); err != nil {
		logs.Warn("security job %s failed: %s", name, err.Error())
	}
}

// InitSecurityPolicyJobs runs the scheduled jobs of the security policies
func InitSecurityPolicyJobs() {
	go func() {
		time.Sleep(time.Minute)
		runSecurityJob("disable inactive users", DisableInactiveUsers)
		runSecurityJob("check certs", CheckCertExpiration)
	}()

	cronJob := cron.New()
	_, err := cronJob.AddFunc("30 3 * * *", func() {
		runSecurityJob("disable inactive users", DisableInactiveUsers)
		runSecurityJob("check certs", CheckCertExpiration)
	})
	if err != nil {
		logs.Warn("security jobs are not scheduled: %s", err.Error())
		return
	}
	cronJob.Start()
}
