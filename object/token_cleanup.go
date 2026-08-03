// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/robfig/cron/v3"
)

const defaultTokenRetentionDays = 30

// getTokenRetentionInterval converts a retention period in days into seconds,
// falling back to the default when the configured value is non-positive.
func getTokenRetentionInterval(days int) int {
	if days <= 0 {
		days = defaultTokenRetentionDays
	}
	return days * 24 * 3600
}

// getOrgTokenRetentionIntervals returns a map from organization name to its
// configured token retention interval (in seconds), so that each organization
// can control how long its expired tokens are kept before cleanup.
func getOrgTokenRetentionIntervals() (map[string]int, error) {
	organizations, err := GetOrganizationsByFields("admin", "name", "token_retention_days")
	if err != nil {
		return nil, fmt.Errorf("failed to load organizations for token cleanup: %w", err)
	}

	intervals := make(map[string]int, len(organizations))
	for _, organization := range organizations {
		intervals[organization.Name] = getTokenRetentionInterval(organization.TokenRetentionDays)
	}
	return intervals, nil
}

func CleanupTokens() error {
	currentTime := time.Now()

	orgIntervals, err := getOrgTokenRetentionIntervals()
	if err != nil {
		return err
	}

	// Use the smallest retention interval (across all organizations) to compute the
	// query cutoff, so that no token eligible for cleanup under any organization's
	// setting is missed. A token can only be eligible if it was created before this
	// cutoff, since createdTime + expiresIn (the token's expiry) must be even earlier
	// than that for it to have been expired for longer than the retention interval.
	minInterval := getTokenRetentionInterval(defaultTokenRetentionDays)
	for _, interval := range orgIntervals {
		if interval < minInterval {
			minInterval = interval
		}
	}
	cutoffTime := currentTime.Add(-time.Duration(minInterval) * time.Second).Format(time.RFC3339)

	var sessions []*Token
	err = ormer.Engine.Where("created_time < ?", cutoffTime).Find(&sessions)
	if err != nil {
		return fmt.Errorf("failed to query expired tokens: %w", err)
	}

	deletedCount := 0

	for _, session := range sessions {
		isExpired, expireTime := util.IsTokenExpired(session.CreatedTime, session.ExpiresIn)
		if !isExpired {
			continue
		}

		retentionInterval, ok := orgIntervals[session.Organization]
		if !ok {
			// The token's organization no longer exists (or has no configured value);
			// fall back to the default retention period.
			retentionInterval = getTokenRetentionInterval(defaultTokenRetentionDays)
		}

		expireTimeObj := util.String2Time(expireTime)
		tokenAfterExpiry := currentTime.Sub(expireTimeObj).Seconds()
		if tokenAfterExpiry > float64(retentionInterval) {
			_, err = ormer.Engine.Delete(session)
			if err != nil {
				return fmt.Errorf("failed to delete expired token %s: %w", session.Name, err)
			}
			fmt.Printf("[%d] Deleted expired token: %s | Created: %s | Org: %s | App: %s | User: %s\n",
				deletedCount, session.Name, session.CreatedTime, session.Organization, session.Application, session.User)
			deletedCount++
		}
	}
	return nil
}

func InitCleanupTokens() {
	schedule := "0 0 * * *"

	go func() {
		if err := CleanupTokens(); err != nil {
			fmt.Printf("Error cleaning up tokens at startup: %v\n", err)
		}
	}()

	cronJob := cron.New()
	_, err := cronJob.AddFunc(schedule, func() {
		if err := CleanupTokens(); err != nil {
			fmt.Printf("Error cleaning up tokens: %v\n", err)
		}
	})
	if err != nil {
		fmt.Printf("Error scheduling token cleanup: %v\n", err)
		return
	}
	cronJob.Start()
}
