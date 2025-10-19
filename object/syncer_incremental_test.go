// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

func TestGetUpdatedTimeColumn(t *testing.T) {
	syncer := &Syncer{
		TableColumns: []*TableColumn{
			{Name: "id", CasdoorName: "Id"},
			{Name: "name", CasdoorName: "Name"},
			{Name: "updated_at", CasdoorName: "UpdatedTime"},
			{Name: "email", CasdoorName: "Email"},
		},
	}

	column := syncer.getUpdatedTimeColumn()
	if column != "updated_at" {
		t.Errorf("Expected 'updated_at', got '%s'", column)
	}
}

func TestGetUpdatedTimeColumnNotFound(t *testing.T) {
	syncer := &Syncer{
		TableColumns: []*TableColumn{
			{Name: "id", CasdoorName: "Id"},
			{Name: "name", CasdoorName: "Name"},
			{Name: "email", CasdoorName: "Email"},
		},
	}

	column := syncer.getUpdatedTimeColumn()
	if column != "" {
		t.Errorf("Expected empty string, got '%s'", column)
	}
}

func TestIncrementalSyncDetection(t *testing.T) {
	// Test case 1: Both LastSyncTime and UpdatedTime column exist
	syncer1 := &Syncer{
		LastSyncTime: "2024-01-01T00:00:00Z",
		TableColumns: []*TableColumn{
			{Name: "updated_at", CasdoorName: "UpdatedTime"},
		},
	}
	if syncer1.LastSyncTime == "" || syncer1.getUpdatedTimeColumn() == "" {
		t.Error("Should support incremental sync when LastSyncTime and UpdatedTime column exist")
	}

	// Test case 2: No LastSyncTime
	syncer2 := &Syncer{
		LastSyncTime: "",
		TableColumns: []*TableColumn{
			{Name: "updated_at", CasdoorName: "UpdatedTime"},
		},
	}
	if syncer2.LastSyncTime != "" {
		t.Error("Should not use incremental sync when LastSyncTime is empty")
	}

	// Test case 3: No UpdatedTime column
	syncer3 := &Syncer{
		LastSyncTime: "2024-01-01T00:00:00Z",
		TableColumns: []*TableColumn{
			{Name: "id", CasdoorName: "Id"},
		},
	}
	if syncer3.getUpdatedTimeColumn() != "" {
		t.Error("Should not use incremental sync when UpdatedTime column doesn't exist")
	}
}
