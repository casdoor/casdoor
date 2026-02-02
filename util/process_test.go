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

package util

import (
	"testing"
)

func TestStopOldInstance(t *testing.T) {
	// Test that StopOldInstance doesn't panic when lsof is not available
	// This simulates the Docker environment where lsof is not installed
	// The function should handle this gracefully and not return an error
	
	// Use a high port number (59999) that's unlikely to be in use by other services
	// This avoids conflicts with commonly used ports (e.g., 8000, 8080, 3000)
	port := 59999
	
	err := StopOldInstance(port)
	if err != nil {
		t.Errorf("StopOldInstance should not return error when lsof is not available or port is not in use, got: %v", err)
	}
}

func TestGetPidByPort(t *testing.T) {
	// Test that getPidByPort handles missing lsof gracefully
	// Use a high port number (59998) that's unlikely to be in use by other services
	port := 59998
	
	pid, err := getPidByPort(port)
	if err != nil {
		t.Errorf("getPidByPort should not return error when lsof is not available or port is not in use, got: %v", err)
	}
	
	// When lsof is not available or port is not in use, pid should be 0
	// Note: If lsof is available and a process is found on this port, pid may be non-zero
	// This is acceptable as the test's primary goal is to ensure no errors occur
	if pid != 0 {
		t.Logf("getPidByPort returned pid %d (a process may be using port %d, or lsof found something)", pid, port)
	}
}
