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

package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/context"
)

func TestSetContextUser(t *testing.T) {
	// Create a test request and context
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Test setting user in context
	testUserId := "built-in/admin"
	setContextUser(ctx, testUserId)

	// Verify user is stored in context
	retrievedUser := getContextUser(ctx)
	if retrievedUser != testUserId {
		t.Errorf("Expected user %s, got %s", testUserId, retrievedUser)
	}
}

func TestGetContextUserEmpty(t *testing.T) {
	// Create a test request and context
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(w, req)

	// Test getting user from empty context
	retrievedUser := getContextUser(ctx)
	if retrievedUser != "" {
		t.Errorf("Expected empty string, got %s", retrievedUser)
	}
}

func TestContextUserIsolation(t *testing.T) {
	// Create first context
	req1 := httptest.NewRequest(http.MethodGet, "/test1", nil)
	w1 := httptest.NewRecorder()
	ctx1 := context.NewContext()
	ctx1.Reset(w1, req1)

	// Create second context
	req2 := httptest.NewRequest(http.MethodGet, "/test2", nil)
	w2 := httptest.NewRecorder()
	ctx2 := context.NewContext()
	ctx2.Reset(w2, req2)

	// Set different users in different contexts
	user1 := "built-in/admin"
	user2 := "built-in/user"

	setContextUser(ctx1, user1)
	setContextUser(ctx2, user2)

	// Verify isolation
	retrieved1 := getContextUser(ctx1)
	retrieved2 := getContextUser(ctx2)

	if retrieved1 != user1 {
		t.Errorf("Context 1: Expected user %s, got %s", user1, retrieved1)
	}
	if retrieved2 != user2 {
		t.Errorf("Context 2: Expected user %s, got %s", user2, retrieved2)
	}
}
