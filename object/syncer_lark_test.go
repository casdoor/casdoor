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

import "testing"

func TestLarkUserToOriginalUserDefaultEmail(t *testing.T) {
	provider := &LarkSyncerProvider{Syncer: &Syncer{}}

	user := provider.larkUserToOriginalUser(&LarkUser{
		UserId:          "u-1",
		Name:            "Alice",
		Email:           "alice@example.com",
		EnterpriseEmail: "alice@corp.example.com",
	})

	if user.Email != "alice@example.com" {
		t.Fatalf("default mapping must use email, got %q", user.Email)
	}
}

func TestLarkUserToOriginalUserEnterpriseEmail(t *testing.T) {
	provider := &LarkSyncerProvider{Syncer: &Syncer{
		TableColumns: []*TableColumn{
			{Name: "enterprise_email", CasdoorName: "Email"},
		},
	}}

	user := provider.larkUserToOriginalUser(&LarkUser{
		UserId:          "u-1",
		Name:            "Alice",
		Email:           "alice@example.com",
		EnterpriseEmail: "alice@corp.example.com",
	})

	if user.Email != "alice@corp.example.com" {
		t.Fatalf("enterprise_email mapping must override email, got %q", user.Email)
	}
	// Fields not listed in TableColumns keep their default mapping.
	if user.Name != "u-1" || user.DisplayName != "Alice" {
		t.Fatalf("default fields must be preserved, got name=%q displayName=%q", user.Name, user.DisplayName)
	}
}

func TestLarkUserToOriginalUserUnknownColumnIsIgnored(t *testing.T) {
	provider := &LarkSyncerProvider{Syncer: &Syncer{
		TableColumns: []*TableColumn{
			{Name: "enterprise_mail", CasdoorName: "Email"}, // typo on purpose
		},
	}}

	user := provider.larkUserToOriginalUser(&LarkUser{
		UserId:          "u-1",
		Name:            "Alice",
		Email:           "alice@example.com",
		EnterpriseEmail: "alice@corp.example.com",
	})

	if user.Email != "alice@example.com" {
		t.Fatalf("unknown column must not overwrite the field, got %q", user.Email)
	}
}

func TestLarkUserFieldValue(t *testing.T) {
	provider := &LarkSyncerProvider{Syncer: &Syncer{}}
	u := &LarkUser{UserId: "u-1", EnterpriseEmail: "alice@corp.example.com"}

	if v, ok := provider.getLarkUserFieldValue(u, "enterprise_email"); !ok || v != "alice@corp.example.com" {
		t.Fatalf("enterprise_email = %q, %v", v, ok)
	}
	if _, ok := provider.getLarkUserFieldValue(u, "enterprise_mail"); ok {
		t.Fatal("unknown field must return ok=false")
	}
}
