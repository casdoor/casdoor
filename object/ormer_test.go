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
	"testing"
)

func TestGetReadEngine(t *testing.T) {
	InitConfig()

	// Test that GetReadEngine returns write engine when read engine is not configured
	readEngine := ormer.GetReadEngine()
	if readEngine == nil {
		t.Error("GetReadEngine() returned nil")
	}

	// When no read engine is configured, it should return the write engine
	if ormer.readEngine == nil && readEngine != ormer.Engine {
		t.Error("GetReadEngine() should return write engine when read engine is not configured")
	}

	// When read engine is configured, it should return the read engine
	if ormer.readEngine != nil && readEngine != ormer.readEngine {
		t.Error("GetReadEngine() should return read engine when configured")
	}
}

func TestIsTransactionPoolingEnabled(t *testing.T) {
	InitConfig()

	// Test that the flag can be read
	enabled := ormer.IsTransactionPoolingEnabled()
	// The default should be false
	t.Logf("Transaction pooling enabled: %v", enabled)
}

func TestGetSessionWithoutPrepare(t *testing.T) {
	InitConfig()

	// Test that GetSession works and doesn't panic
	session := GetSession("", -1, -1, "", "", "", "")
	if session == nil {
		t.Error("GetSession() returned nil")
	}

	// Test that GetSessionForUser works and doesn't panic
	sessionForUser := GetSessionForUser("", -1, -1, "", "", "", "")
	if sessionForUser == nil {
		t.Error("GetSessionForUser() returned nil")
	}
}
