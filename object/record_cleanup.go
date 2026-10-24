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
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// recordCleanupBatchSize limits how many audit rows are deleted by a single statement,
// so that cleaning up a "record" table that has grown for years does not lock it for
// a long time.
const recordCleanupBatchSize = 1000

// recordCleanupBatchInterval spreads the deletes over time, so that the binlog they generate
// does not fill up the database's disk at once.
const recordCleanupBatchInterval = 100 * time.Millisecond

// getOrgRecordRetentionDays returns a map from organization name to its configured
// record retention period in days. Organizations that keep their records forever
// (the default, i.e. a non-positive value) are not included, so audit rows are never
// deleted unless the retention has been explicitly configured.
func getOrgRecordRetentionDays() (map[string]int, error) {
	organizations, err := GetOrganizationsByFields("admin", "name", "record_retention_days")
	if err != nil {
		return nil, fmt.Errorf("failed to load organizations for record cleanup: %w", err)
	}

	res := map[string]int{}
	for _, organization := range organizations {
		if organization.RecordRetentionDays > 0 {
			res[organization.Name] = organization.RecordRetentionDays
		}
	}
	return res, nil
}

// cleanupOrgRecords deletes the records of one organization that were created before
// cutoffTime, batch by batch, and returns how many rows were deleted.
func cleanupOrgRecords(owner string, cutoffTime string) (int64, error) {
	deletedCount := int64(0)

	for {
		records := []*Record{}
		err := ormer.Engine.Cols("id").Where("owner = ?", owner).And("created_time < ?", cutoffTime).Limit(recordCleanupBatchSize).Find(&records)
		if err != nil {
			return deletedCount, fmt.Errorf("failed to query expired records of organization %s: %w", owner, err)
		}
		if len(records) == 0 {
			break
		}

		ids := []int{}
		for _, record := range records {
			ids = append(ids, record.Id)
		}

		affected, err := ormer.Engine.In("id", ids).Delete(&Record{})
		if err != nil {
			return deletedCount, fmt.Errorf("failed to delete expired records of organization %s: %w", owner, err)
		}
		deletedCount += affected

		if len(records) < recordCleanupBatchSize {
			break
		}
		time.Sleep(recordCleanupBatchInterval)
	}

	return deletedCount, nil
}

// cleanupRecordsByGlobalRetention deletes the records created before cutoffTime whose organizations
// do not set their own retention. The records are scanned in the order of their ids, because
// "created_time" has no index and the ids grow with it.
func cleanupRecordsByGlobalRetention(cutoffTime string, retentionDaysMap map[string]int) (int64, error) {
	deletedCount := int64(0)
	lastId := 0

	for {
		records := []*Record{}
		err := ormer.Engine.Cols("id", "owner", "created_time").Where("id > ?", lastId).Asc("id").Limit(recordCleanupBatchSize).Find(&records)
		if err != nil {
			return deletedCount, fmt.Errorf("failed to query expired records: %w", err)
		}
		if len(records) == 0 {
			break
		}

		ids := []int{}
		reachedCutoff := false
		for _, record := range records {
			if record.CreatedTime >= cutoffTime {
				reachedCutoff = true
				break
			}

			lastId = record.Id
			if _, ok := retentionDaysMap[record.Owner]; !ok {
				ids = append(ids, record.Id)
			}
		}

		if len(ids) != 0 {
			affected, err := ormer.Engine.In("id", ids).Delete(&Record{})
			if err != nil {
				return deletedCount, fmt.Errorf("failed to delete expired records: %w", err)
			}
			deletedCount += affected
		}

		if reachedCutoff || len(records) < recordCleanupBatchSize {
			break
		}
		time.Sleep(recordCleanupBatchInterval)
	}

	return deletedCount, nil
}

// owners of the rows written before unauthenticated requests were attributed, see AddRecord()
var legacyOrphanRecordOwners = []string{"", "anonymous"}

func addLegacyOrphanRetentionDays(retentionDaysMap map[string]int) {
	retentionDays, ok := retentionDaysMap["built-in"]
	if !ok {
		return
	}

	for _, owner := range legacyOrphanRecordOwners {
		if _, ok := retentionDaysMap[owner]; !ok {
			retentionDaysMap[owner] = retentionDays
		}
	}
}

func CleanupRecords() error {
	retentionDaysMap, err := getOrgRecordRetentionDays()
	if err != nil {
		return err
	}
	addLegacyOrphanRetentionDays(retentionDaysMap)

	currentTime := time.Now()
	for owner, retentionDays := range retentionDaysMap {
		// "record"'s "owner" column is the organization that the record belongs to,
		// see AddRecord().
		cutoffTime := currentTime.AddDate(0, 0, -retentionDays).Format(time.RFC3339)

		deletedCount, err := cleanupOrgRecords(owner, cutoffTime)
		if err != nil {
			return err
		}

		if deletedCount != 0 {
			fmt.Printf("Deleted [%d] expired records | Org: %s | Created before: %s\n", deletedCount, owner, cutoffTime)
		}
	}

	setting, err := GetSetting()
	if err != nil {
		return err
	}
	if setting == nil || setting.RecordRetentionDays <= 0 {
		return nil
	}

	cutoffTime := currentTime.AddDate(0, 0, -setting.RecordRetentionDays).Format(time.RFC3339)
	deletedCount, err := cleanupRecordsByGlobalRetention(cutoffTime, retentionDaysMap)
	if err != nil {
		return err
	}

	if deletedCount != 0 {
		fmt.Printf("Deleted [%d] expired records | Global setting | Created before: %s\n", deletedCount, cutoffTime)
	}

	return nil
}

func InitCleanupRecords() {
	schedule := "0 0 * * *"

	go func() {
		if err := CleanupRecords(); err != nil {
			fmt.Printf("Error cleaning up records at startup: %v\n", err)
		}
	}()

	cronJob := cron.New()
	_, err := cronJob.AddFunc(schedule, func() {
		if err := CleanupRecords(); err != nil {
			fmt.Printf("Error cleaning up records: %v\n", err)
		}
	})
	if err != nil {
		fmt.Printf("Error scheduling record cleanup: %v\n", err)
		return
	}
	cronJob.Start()
}
