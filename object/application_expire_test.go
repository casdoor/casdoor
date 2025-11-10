// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"time"
)

func TestApplicationExpireInHoursFloat64(t *testing.T) {
	// Test that ExpireInHours accepts float64 values
	app := &Application{
		ExpireInHours:        0.25, // 15 minutes
		RefreshExpireInHours: 0.5,  // 30 minutes
	}

	// Verify the values are stored correctly
	if app.ExpireInHours != 0.25 {
		t.Errorf("Expected ExpireInHours to be 0.25, got %f", app.ExpireInHours)
	}

	if app.RefreshExpireInHours != 0.5 {
		t.Errorf("Expected RefreshExpireInHours to be 0.5, got %f", app.RefreshExpireInHours)
	}

	// Test time calculation for JWT token expiration
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Duration(app.ExpireInHours * float64(time.Hour)))
	expectedDuration := 15 * time.Minute

	actualDuration := expireTime.Sub(nowTime)
	// Allow 1 second tolerance for time calculation
	if actualDuration < expectedDuration-time.Second || actualDuration > expectedDuration+time.Second {
		t.Errorf("Expected duration to be approximately %v, got %v", expectedDuration, actualDuration)
	}

	// Test time calculation for refresh token expiration
	refreshExpireTime := nowTime.Add(time.Duration(app.RefreshExpireInHours * float64(time.Hour)))
	expectedRefreshDuration := 30 * time.Minute

	actualRefreshDuration := refreshExpireTime.Sub(nowTime)
	if actualRefreshDuration < expectedRefreshDuration-time.Second || actualRefreshDuration > expectedRefreshDuration+time.Second {
		t.Errorf("Expected refresh duration to be approximately %v, got %v", expectedRefreshDuration, actualRefreshDuration)
	}
}

func TestApplicationExpireInHoursToSeconds(t *testing.T) {
	// Test conversion to seconds for OAuth token
	app := &Application{
		ExpireInHours: 0.25, // 15 minutes
	}

	hourSeconds := int(time.Hour / time.Second)
	expiresInSeconds := int(app.ExpireInHours * float64(hourSeconds))

	expectedSeconds := 15 * 60 // 15 minutes = 900 seconds
	if expiresInSeconds != expectedSeconds {
		t.Errorf("Expected %d seconds, got %d", expectedSeconds, expiresInSeconds)
	}

	// Test with 1 hour
	app.ExpireInHours = 1.0
	expiresInSeconds = int(app.ExpireInHours * float64(hourSeconds))
	expectedSeconds = 3600
	if expiresInSeconds != expectedSeconds {
		t.Errorf("Expected %d seconds, got %d", expectedSeconds, expiresInSeconds)
	}

	// Test with fractional hours (2.5 hours)
	app.ExpireInHours = 2.5
	expiresInSeconds = int(app.ExpireInHours * float64(hourSeconds))
	expectedSeconds = 9000 // 2.5 hours = 9000 seconds
	if expiresInSeconds != expectedSeconds {
		t.Errorf("Expected %d seconds, got %d", expectedSeconds, expiresInSeconds)
	}
}
