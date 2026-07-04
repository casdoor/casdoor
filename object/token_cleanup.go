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

func CleanupTokens(tokenRetentionIntervalAfterExpiry int) error {
	currentTime := time.Now()
	// A token can only be eligible for cleanup if it was created before this cutoff,
	// since createdTime + expiresIn (the token's expiry) must be even earlier than that
	// for it to have been expired for longer than the retention interval.
	cutoffTime := currentTime.Add(-time.Duration(tokenRetentionIntervalAfterExpiry) * time.Second).Format(time.RFC3339)

	var sessions []*Token
	err := ormer.Engine.Where("created_time < ?", cutoffTime).Find(&sessions)
	if err != nil {
		return fmt.Errorf("failed to query expired tokens: %w", err)
	}

	deletedCount := 0

	for _, session := range sessions {
		isExpired, expireTime := util.IsTokenExpired(session.CreatedTime, session.ExpiresIn)
		if !isExpired {
			continue
		}

		expireTimeObj := util.String2Time(expireTime)
		tokenAfterExpiry := currentTime.Sub(expireTimeObj).Seconds()
		if tokenAfterExpiry > float64(tokenRetentionIntervalAfterExpiry) {
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

func getTokenRetentionInterval(days int) int {
	if days <= 0 {
		days = 30
	}
	return days * 24 * 3600
}

func InitCleanupTokens() {
	schedule := "0 0 * * *"
	interval := getTokenRetentionInterval(30)

	go func() {
		if err := CleanupTokens(interval); err != nil {
			fmt.Printf("Error cleaning up tokens at startup: %v\n", err)
		}
	}()

	cronJob := cron.New()
	_, err := cronJob.AddFunc(schedule, func() {
		if err := CleanupTokens(interval); err != nil {
			fmt.Printf("Error cleaning up tokens: %v\n", err)
		}
	})
	if err != nil {
		fmt.Printf("Error scheduling token cleanup: %v\n", err)
		return
	}
	cronJob.Start()
}
