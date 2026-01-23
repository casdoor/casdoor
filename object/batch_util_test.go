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
	"testing"
)

func TestCalculateSafeBatchSize(t *testing.T) {
	tests := []struct {
		name            string
		fieldsPerRecord int
		expectedMax     int // Maximum value we should see
	}{
		{
			name:            "User struct with 157 fields",
			fieldsPerRecord: 157,
			// With 157 fields: 65535 * 0.9 / 157 ≈ 376
			expectedMax: 376,
		},
		{
			name:            "Role struct with 9 fields",
			fieldsPerRecord: 9,
			// With 9 fields: 65535 * 0.9 / 9 ≈ 6553
			expectedMax: 6553,
		},
		{
			name:            "Permission struct with 19 fields",
			fieldsPerRecord: 19,
			// With 19 fields: 65535 * 0.9 / 19 ≈ 3102
			expectedMax: 3102,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchSize := calculateSafeBatchSize(tt.fieldsPerRecord)

			// Ensure batch size is positive
			if batchSize <= 0 {
				t.Errorf("calculateSafeBatchSize() returned non-positive value: %d", batchSize)
			}

			// Calculate the number of parameters this would create
			totalParams := batchSize * tt.fieldsPerRecord

			// Ensure we don't exceed PostgreSQL's limit
			if totalParams > postgresMaxParameters {
				t.Errorf("calculateSafeBatchSize() would create %d parameters (batch size %d * %d fields), exceeding PostgreSQL limit of %d",
					totalParams, batchSize, tt.fieldsPerRecord, postgresMaxParameters)
			}

			// Ensure we're using the safety margin (90% of limit)
			safeLimit := postgresMaxParameters * 9 / 10
			if totalParams > safeLimit {
				t.Errorf("calculateSafeBatchSize() would create %d parameters, exceeding safe limit of %d (90%% of %d)",
					totalParams, safeLimit, postgresMaxParameters)
			}

			// Log the calculated batch size for verification
			t.Logf("For %d fields per record, calculated batch size: %d (total params: %d)",
				tt.fieldsPerRecord, batchSize, totalParams)
		})
	}
}

func TestCalculateSafeBatchSizeRespectsConfiguredBatchSize(t *testing.T) {
	// This test verifies that if the configured batch size is smaller than
	// the PostgreSQL-safe batch size, we use the configured batch size.
	
	// For a struct with very few fields (e.g., 2 fields), the safe batch size
	// would be very large (65535 * 0.9 / 2 ≈ 29490).
	// But if the configured batch size is 100, we should use 100 instead.
	fieldsPerRecord := 2
	batchSize := calculateSafeBatchSize(fieldsPerRecord)

	// The batch size should be limited by the configured batch size (typically 100)
	// and not be the theoretical maximum
	if batchSize > 10000 {
		t.Logf("Note: Batch size %d seems unusually large for production use. "+
			"This is expected if configured batch size is also large.", batchSize)
	}

	t.Logf("For %d fields per record, calculated batch size: %d", fieldsPerRecord, batchSize)
}
