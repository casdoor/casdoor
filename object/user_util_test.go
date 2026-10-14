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
	"strings"
	"testing"

	"github.com/casdoor/casdoor/util"
)

func TestGetDefaultAccountItemsIncludesIsVerified(t *testing.T) {
	var item *AccountItem
	for _, it := range GetDefaultAccountItems() {
		if it.Name == "Is verified" {
			item = it
			break
		}
	}
	if item == nil {
		t.Fatal("GetDefaultAccountItems() does not contain an \"Is verified\" item, the switch on the user page cannot follow it")
	}
	if !item.Visible {
		t.Error("\"Is verified\" item Visible = false, want true")
	}
	if item.ViewRule != "Self" {
		t.Errorf("\"Is verified\" item ViewRule = %q, want %q", item.ViewRule, "Self")
	}
	if item.ModifyRule != "Admin" {
		t.Errorf("\"Is verified\" item ModifyRule = %q, want %q, a user must not be able to verify themselves", item.ModifyRule, "Admin")
	}
}

func TestIsVerifiedAccountItemModifyRule(t *testing.T) {
	item := &AccountItem{Name: "Is verified", Visible: true, ViewRule: "Self", ModifyRule: "Admin"}

	pass, msg := CheckAccountItemModifyRule(item, false, "en")
	if pass {
		t.Fatal("CheckAccountItemModifyRule() = pass for a non-admin on the \"Is verified\" item, want refusal")
	}
	if !strings.Contains(msg, "Is verified") {
		t.Errorf("CheckAccountItemModifyRule() msg = %q, want it to name the \"Is verified\" item", msg)
	}

	pass, msg = CheckAccountItemModifyRule(item, true, "en")
	if !pass {
		t.Fatalf("CheckAccountItemModifyRule() refused an admin on the \"Is verified\" item: %s", msg)
	}
}

func TestGetDefaultUserUpdateColumnsIsVerified(t *testing.T) {
	adminColumns := GetDefaultUserUpdateColumns(true)
	if !util.InSlice(adminColumns, "is_verified") {
		t.Error("GetDefaultUserUpdateColumns(isAdmin = true) is missing the \"is_verified\" column, an admin update cannot persist IsVerified")
	}

	selfColumns := GetDefaultUserUpdateColumns(false)
	if util.InSlice(selfColumns, "is_verified") {
		t.Error("GetDefaultUserUpdateColumns(isAdmin = false) contains the \"is_verified\" column, a self update must not be able to write the verification flag")
	}
}
