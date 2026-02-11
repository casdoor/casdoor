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
	"strings"
	"testing"
)

// TestPolicySyncWithoutRedis tests that policy synchronization gracefully handles
// the case when Redis is not configured (no panic, no errors for basic operations)
func TestPolicySyncWithoutRedis(t *testing.T) {
	// Save original values
	origRedisClient := redisClient
	origPodId := podId
	
	// Reset values after test
	t.Cleanup(func() {
		redisClient = origRedisClient
		podId = origPodId
	})

	// Ensure redisClient is nil (no Redis configured)
	redisClient = nil
	podId = "test-pod"

	// Test that publishing policy changes works without Redis (should be no-op)
	err := publishPolicyChange("test-enforcer", "add")
	if err != nil {
		t.Errorf("publishPolicyChange should not return error when Redis is not configured: %v", err)
	}

	err = publishPolicyChange("test-enforcer", "remove")
	if err != nil {
		t.Errorf("publishPolicyChange should not return error when Redis is not configured: %v", err)
	}

	err = publishPolicyChange("test-enforcer", "update")
	if err != nil {
		t.Errorf("publishPolicyChange should not return error when Redis is not configured: %v", err)
	}
}

// TestPodIdGeneration tests that pod ID generation works correctly
func TestPodIdGeneration(t *testing.T) {
	// Save original function
	originalGetEnvVar := getEnvVar
	
	// Reset after test
	t.Cleanup(func() {
		getEnvVar = originalGetEnvVar
	})

	// Mock getEnvVar to return empty string for testing
	getEnvVar = func(key string) string {
		return ""
	}

	podId := generatePodId()
	if podId == "" {
		t.Error("generatePodId should return a non-empty string")
	}
	if !strings.HasPrefix(podId, "pod-") {
		t.Errorf("generatePodId should return a string starting with 'pod-', got: %s", podId)
	}
}
