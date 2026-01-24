// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"github.com/casdoor/casdoor/conf"
)

const (
	// PostgreSQL parameter limit
	// See: https://www.postgresql.org/docs/current/limits.html
	postgresMaxParameters = 65535

	// Database field counts for batch operations
	// These are the number of fields with xorm tags (excluding xorm:"-")
	// To verify: grep -c '`xorm:' object/<struct>.go | grep -v 'xorm:"-"'
	userDBFields       = 156 // User struct database fields
	roleDBFields       = 9   // Role struct database fields
	permissionDBFields = 19  // Permission struct database fields
)

// calculateSafeBatchSize calculates a safe batch size that respects both
// the configured batch size and PostgreSQL's parameter limit.
// PostgreSQL has a limit of 65535 parameters per query, and each record
// in a batch insert uses N parameters (where N is the number of fields).
func calculateSafeBatchSize(fieldsPerRecord int) int {
	// Guard against invalid input
	if fieldsPerRecord <= 0 {
		return 1
	}

	configuredBatchSize := conf.GetConfigBatchSize()

	// Calculate maximum batch size based on PostgreSQL parameter limit
	// Leave some margin for safety (use 90% of the limit)
	// Using integer arithmetic to avoid floating-point precision issues
	maxSafeBatchSize := (postgresMaxParameters * 9) / (10 * fieldsPerRecord)

	// Ensure we always have at least batch size of 1
	if maxSafeBatchSize < 1 {
		maxSafeBatchSize = 1
	}

	// Use the smaller of configured batch size and safe batch size
	if configuredBatchSize < maxSafeBatchSize {
		return configuredBatchSize
	}
	return maxSafeBatchSize
}
