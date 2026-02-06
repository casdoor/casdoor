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

func TestDingtalkUserToOriginalUser(t *testing.T) {
	provider := &DingtalkSyncerProvider{
		Syncer: &Syncer{},
	}

	// Test case 1: Full DingTalk user with all fields
	dtUser := &DingtalkUser{
		UserId:     "user123",
		UnionId:    "union456",
		Name:       "张三",
		Department: []int{1, 2, 3},
		Position:   "Software Engineer",
		Mobile:     "13800138000",
		Email:      "zhangsan@example.com",
		Avatar:     "https://example.com/avatar.jpg",
		JobNumber:  "E001",
		Active:     true,
	}

	originalUser := provider.dingtalkUserToOriginalUser(dtUser)

	// Verify basic fields
	if originalUser.Id != "user123" {
		t.Errorf("Expected Id to be 'user123', got '%s'", originalUser.Id)
	}
	if originalUser.Name != "union456" {
		t.Errorf("Expected Name to be 'union456' (unionId), got '%s'", originalUser.Name)
	}
	if originalUser.DisplayName != "张三" {
		t.Errorf("Expected DisplayName to be '张三', got '%s'", originalUser.DisplayName)
	}
	if originalUser.Email != "zhangsan@example.com" {
		t.Errorf("Expected Email to be 'zhangsan@example.com', got '%s'", originalUser.Email)
	}
	if originalUser.Phone != "13800138000" {
		t.Errorf("Expected Phone to be '13800138000', got '%s'", originalUser.Phone)
	}
	if originalUser.Avatar != "https://example.com/avatar.jpg" {
		t.Errorf("Expected Avatar to be 'https://example.com/avatar.jpg', got '%s'", originalUser.Avatar)
	}
	if originalUser.Title != "Software Engineer" {
		t.Errorf("Expected Title to be 'Software Engineer', got '%s'", originalUser.Title)
	}
	if originalUser.IsForbidden != false {
		t.Errorf("Expected IsForbidden to be false for active user, got %v", originalUser.IsForbidden)
	}

	// Verify department groups are populated
	if len(originalUser.Groups) != 3 {
		t.Errorf("Expected 3 groups (departments), got %d", len(originalUser.Groups))
	}
	expectedGroups := []string{"1", "2", "3"}
	for i, expectedGroup := range expectedGroups {
		if i >= len(originalUser.Groups) || originalUser.Groups[i] != expectedGroup {
			t.Errorf("Expected group[%d] to be '%s', got '%s'", i, expectedGroup, originalUser.Groups[i])
		}
	}

	// Test case 2: Inactive user (should be forbidden)
	inactiveUser := &DingtalkUser{
		UserId:     "user456",
		UnionId:    "union789",
		Name:       "李四",
		Department: []int{1},
		Active:     false,
	}

	inactiveOriginalUser := provider.dingtalkUserToOriginalUser(inactiveUser)
	if inactiveOriginalUser.IsForbidden != true {
		t.Errorf("Expected IsForbidden to be true for inactive user, got %v", inactiveOriginalUser.IsForbidden)
	}

	// Test case 3: User without UnionId (should fallback to UserId for name)
	noUnionIdUser := &DingtalkUser{
		UserId:     "user789",
		UnionId:    "",
		Name:       "王五",
		Department: []int{},
		Active:     true,
	}

	noUnionIdOriginalUser := provider.dingtalkUserToOriginalUser(noUnionIdUser)
	if noUnionIdOriginalUser.Name != "user789" {
		t.Errorf("Expected Name to fallback to UserId 'user789', got '%s'", noUnionIdOriginalUser.Name)
	}
	if len(noUnionIdOriginalUser.Groups) != 0 {
		t.Errorf("Expected 0 groups for user with no departments, got %d", len(noUnionIdOriginalUser.Groups))
	}
}

func TestDingtalkDepartmentToOriginalGroup(t *testing.T) {
	provider := &DingtalkSyncerProvider{
		Syncer: &Syncer{},
	}

	// Test case 1: Department with parent
	dept := &DingtalkDepartment{
		DeptId:          123,
		Name:            "Engineering",
		ParentId:        1,
		CreateDeptGroup: true,
		AutoAddUser:     true,
	}

	originalGroup := provider.dingtalkDepartmentToOriginalGroup(dept)

	// Verify all fields
	if originalGroup.Id != "123" {
		t.Errorf("Expected Id to be '123', got '%s'", originalGroup.Id)
	}
	if originalGroup.Name != "123" {
		t.Errorf("Expected Name to be '123', got '%s'", originalGroup.Name)
	}
	if originalGroup.DisplayName != "Engineering" {
		t.Errorf("Expected DisplayName to be 'Engineering', got '%s'", originalGroup.DisplayName)
	}
	if originalGroup.Type != "department" {
		t.Errorf("Expected Type to be 'department', got '%s'", originalGroup.Type)
	}

	// Test case 2: Root department (parent = 0)
	rootDept := &DingtalkDepartment{
		DeptId:   1,
		Name:     "Company",
		ParentId: 0,
	}

	rootOriginalGroup := provider.dingtalkDepartmentToOriginalGroup(rootDept)
	if rootOriginalGroup.Id != "1" {
		t.Errorf("Expected root department Id to be '1', got '%s'", rootOriginalGroup.Id)
	}
	if rootOriginalGroup.DisplayName != "Company" {
		t.Errorf("Expected root department DisplayName to be 'Company', got '%s'", rootOriginalGroup.DisplayName)
	}
}

func TestGetSyncerProviderDingTalk(t *testing.T) {
	syncer := &Syncer{
		Type: "DingTalk",
		User: "test_app_key",
		Password: "test_app_secret",
	}

	provider := GetSyncerProvider(syncer)

	if _, ok := provider.(*DingtalkSyncerProvider); !ok {
		t.Errorf("Expected DingtalkSyncerProvider for type 'DingTalk', got %T", provider)
	}
}

func TestDingtalkSyncerProviderEmptyMethods(t *testing.T) {
	provider := &DingtalkSyncerProvider{
		Syncer: &Syncer{},
	}

	// Test AddUser returns error (read-only syncer)
	_, err := provider.AddUser(&OriginalUser{})
	if err == nil {
		t.Error("Expected AddUser to return error for read-only syncer")
	}

	// Test UpdateUser returns error (read-only syncer)
	_, err = provider.UpdateUser(&OriginalUser{})
	if err == nil {
		t.Error("Expected UpdateUser to return error for read-only syncer")
	}

	// Test Close returns no error
	err = provider.Close()
	if err != nil {
		t.Errorf("Expected Close to return nil, got error: %v", err)
	}

	// Test InitAdapter returns no error
	err = provider.InitAdapter()
	if err != nil {
		t.Errorf("Expected InitAdapter to return nil, got error: %v", err)
	}
}
