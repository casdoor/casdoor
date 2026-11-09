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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/i18n"
)

var smsAlertDays sync.Map

func countSmsRecords(where string, args ...interface{}) (int64, error) {
	return ormer.Engine.Where("type = ?", "SMS").And(where, args...).Count(&VerificationRecord{})
}

// getPhonePrefix is the number segment that SMS bombing walks through: the first 7 digits of a Chinese mobile
// number, the country code and first 4 digits otherwise
func getPhonePrefix(phone string) string {
	if strings.HasPrefix(phone, "+86") && len(phone) >= 10 {
		return phone[:10]
	}
	if len(phone) >= 8 {
		return phone[:len(phone)-4]
	}
	return ""
}

func startOfToday() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
}

// checkSmsSecurityPolicy refuses the SMS codes that the organization's limits don't allow: per phone number, per
// number segment, international numbers and the daily total
func checkSmsSecurityPolicy(organization *Organization, phone string, lang string) error {
	policy := GetSecurityPolicy(organization)
	if policy == nil {
		return nil
	}

	if policy.SmsDomesticOnly && !strings.HasPrefix(phone, "+86") {
		return errors.New(i18n.Translate(lang, "verification:SMS codes can only be sent to mainland China numbers"))
	}

	now := time.Now().Unix()
	limits := []struct {
		limit int
		where string
		args  []interface{}
	}{
		{policy.SmsPerPhoneHourly, "receiver = ? and time > ?", []interface{}{phone, now - 3600}},
		{policy.SmsPerPhoneDaily, "receiver = ? and time > ?", []interface{}{phone, now - 86400}},
	}
	if prefix := getPhonePrefix(phone); prefix != "" {
		limits = append(limits, struct {
			limit int
			where string
			args  []interface{}
		}{policy.SmsPerPrefixHourly, "receiver like ? and time > ?", []interface{}{prefix + "%", now - 3600}})
	}
	for _, item := range limits {
		if item.limit <= 0 {
			continue
		}
		count, err := countSmsRecords(item.where, item.args...)
		if err != nil {
			return err
		}
		if count >= int64(item.limit) {
			return errors.New(i18n.Translate(lang, "verification:Too many SMS codes have been sent to this number, please try again later"))
		}
	}

	if policy.SmsDailyLimit > 0 {
		count, err := countSmsRecords("owner = ? and time >= ?", organization.Name, startOfToday())
		if err != nil {
			return err
		}
		if count >= int64(policy.SmsDailyLimit) {
			alertSmsVolume(organization, count, true)
			return errors.New(i18n.Translate(lang, "verification:The SMS codes of today have run out, please try again tomorrow or use another way"))
		}
	}
	return nil
}

// checkSmsVolume alerts the administrators once a day when the SMS codes sent today reach SmsDailyAlertCount
func checkSmsVolume(organization *Organization) {
	policy := GetSecurityPolicy(organization)
	if policy == nil || policy.SmsDailyAlertCount <= 0 {
		return
	}
	count, err := countSmsRecords("owner = ? and time >= ?", organization.Name, startOfToday())
	if err == nil && count >= int64(policy.SmsDailyAlertCount) {
		alertSmsVolume(organization, count, false)
	}
}

func alertSmsVolume(organization *Organization, count int64, isStopped bool) {
	key := fmt.Sprintf("%s/%s/%v", organization.Name, time.Now().Format("2006-01-02"), isStopped)
	if _, sent := smsAlertDays.LoadOrStore(key, true); sent {
		return
	}

	go func() {
		content := fmt.Sprintf("%d SMS codes have been sent today, reaching the alert threshold. Check the verification records for SMS bombing, and raise the limits in the organization's security settings if the traffic is real.", count)
		if isStopped {
			content = fmt.Sprintf("%d SMS codes have been sent today, reaching the daily limit. No more SMS codes will be sent today.", count)
		}
		SendSecurityAlert(organization, fmt.Sprintf("[%s] SMS volume alert", organization.Name), content)
	}()
}
