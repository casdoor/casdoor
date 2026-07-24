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
	"testing"
)

func TestShouldDiscardGetRecordLogPostOnly(t *testing.T) {
	old := logPostOnly
	logPostOnly = true
	defer func() { logPostOnly = old }()

	if shouldDiscardGetRecord(&Record{Method: "GET", Action: "logout"}) {
		t.Error("GET logout should be kept when logPostOnly=true")
	}

	if !shouldDiscardGetRecord(&Record{Method: "GET", Action: "get-account"}) {
		t.Error("ordinary GET should be discarded when logPostOnly=true")
	}

	if shouldDiscardGetRecord(&Record{Method: "POST", Action: "logout"}) {
		t.Error("POST logout should not be discarded by logPostOnly")
	}
}

func TestShouldDiscardGetRecordLogPostOnlyDisabled(t *testing.T) {
	old := logPostOnly
	logPostOnly = false
	defer func() { logPostOnly = old }()

	if shouldDiscardGetRecord(&Record{Method: "GET", Action: "get-account"}) {
		t.Error("GET should not be discarded when logPostOnly=false")
	}
}
