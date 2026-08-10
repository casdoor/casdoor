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

	"github.com/xorm-io/builder"
)

func TestBuildUserSearchCondEmpty(t *testing.T) {
	if cond := BuildUserSearchCond("", ""); cond != nil {
		t.Fatalf("expected nil cond for empty search, got %#v", cond)
	}
	if cond := BuildUserSearchCond("", "a."); cond != nil {
		t.Fatalf("expected nil cond for empty search with prefix, got %#v", cond)
	}
}

func TestBuildUserSearchCondOrFields(t *testing.T) {
	cond := BuildUserSearchCond("ivan", "")
	if cond == nil {
		t.Fatal("expected non-nil cond")
	}

	sql, args, err := builder.ToSQL(cond)
	if err != nil {
		t.Fatalf("builder.ToSQL: %v", err)
	}

	lowerSQL := strings.ToLower(sql)
	for _, col := range []string{"name", "display_name", "email"} {
		if !strings.Contains(lowerSQL, col) {
			t.Fatalf("sql %q should contain column %q", sql, col)
		}
	}
	if !strings.Contains(lowerSQL, " or ") {
		t.Fatalf("sql %q should contain OR between fields", sql)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %#v", len(args), args)
	}
	for _, arg := range args {
		if arg != "%ivan%" {
			t.Fatalf("expected arg %%ivan%%, got %#v", arg)
		}
	}
}

func TestBuildUserSearchCondColumnPrefix(t *testing.T) {
	cond := BuildUserSearchCond("ivan", "a.")
	if cond == nil {
		t.Fatal("expected non-nil cond")
	}

	sql, _, err := builder.ToSQL(cond)
	if err != nil {
		t.Fatalf("builder.ToSQL: %v", err)
	}

	for _, col := range []string{"a.name", "a.display_name", "a.email"} {
		if !strings.Contains(sql, col) {
			t.Fatalf("sql %q should contain prefixed column %q", sql, col)
		}
	}
}
