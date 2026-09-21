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
	"time"
)

func TestGetRetentionCutoffs(t *testing.T) {
	now := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)

	cutoff := func(days int) string {
		return now.AddDate(0, 0, -days).Format(time.RFC3339)
	}

	t.Run("no retention configured deletes nothing", func(t *testing.T) {
		if got := getRetentionCutoffs(map[string]int{}, now); len(got) != 0 {
			t.Fatalf("expected no cutoffs, got %v", got)
		}
	})

	t.Run("empty owner covered by single org retention", func(t *testing.T) {
		got := getRetentionCutoffs(map[string]int{"org1": 30}, now)
		if got["org1"] != cutoff(30) {
			t.Fatalf("expected org1 cutoff %s, got %s", cutoff(30), got["org1"])
		}
		if got[""] != cutoff(30) {
			t.Fatalf("expected empty-owner cutoff %s, got %s", cutoff(30), got[""])
		}
	})

	t.Run("empty owner uses strictest retention", func(t *testing.T) {
		got := getRetentionCutoffs(map[string]int{"org1": 30, "org2": 7}, now)
		if got[""] != cutoff(7) {
			t.Fatalf("expected empty-owner cutoff %s, got %s", cutoff(7), got[""])
		}
		if got["org1"] != cutoff(30) || got["org2"] != cutoff(7) {
			t.Fatalf("per-org cutoffs changed: %v", got)
		}
	})
}
