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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserRoles(t *testing.T) {
	testOrmer, err := NewAdapter("sqlite", "file:user_roles_test?mode=memory&cache=shared", "")
	require.NoError(t, err)
	require.NoError(t, testOrmer.Engine.Sync2(new(User), new(Group), new(Role)))

	previousOrmer := ormer
	ormer = testOrmer
	t.Cleanup(func() {
		ormer = previousOrmer
		require.NoError(t, testOrmer.Engine.Close())
	})

	userId := "built-in/alice"
	groupId := "built-in/developers"
	_, err = testOrmer.Engine.Insert(&Group{
		Owner:     "built-in",
		Name:      "developers",
		IsEnabled: true,
	})
	require.NoError(t, err)

	_, err = testOrmer.Engine.Insert(&User{
		Owner: "built-in",
		Name:  "alice",
	})
	require.NoError(t, err)
	_, err = testOrmer.Engine.Insert(&User{Owner: "built-in", Name: "bob"})
	require.NoError(t, err)

	_, err = testOrmer.Engine.Insert([]*Role{
		{Owner: "built-in", Name: "editor", Users: []string{userId}, IsEnabled: true},
		{Owner: "built-in", Name: "viewer", Groups: []string{groupId}, IsEnabled: true},
		{Owner: "built-in", Name: "shared", Users: []string{userId}, Groups: []string{groupId}, IsEnabled: true},
		{Owner: "built-in", Name: "disabled", Users: []string{userId}, IsEnabled: false},
		{Owner: "other", Name: "foreign", Users: []string{userId}, IsEnabled: true},
	})
	require.NoError(t, err)

	beforeGroupAssignment, err := GetUserRoles("built-in", "alice")
	require.NoError(t, err)
	assert.Equal(t, []string{"editor", "shared"}, beforeGroupAssignment.DirectRoles)
	assert.Empty(t, beforeGroupAssignment.GroupRoles)

	_, err = testOrmer.Engine.Where("owner = ? AND name = ?", "built-in", "alice").
		Cols("groups").Update(&User{Groups: []string{groupId}})
	require.NoError(t, err)

	afterGroupAssignment, err := GetUserRoles("built-in", "alice")
	require.NoError(t, err)
	assert.Equal(t, []string{"editor", "shared"}, afterGroupAssignment.DirectRoles)
	assert.Equal(t, []string{"shared", "viewer"}, afterGroupAssignment.GroupRoles)

	withoutRoles, err := GetUserRoles("built-in", "bob")
	require.NoError(t, err)
	assert.Empty(t, withoutRoles.DirectRoles)
	assert.NotNil(t, withoutRoles.DirectRoles)
	assert.Empty(t, withoutRoles.GroupRoles)
	assert.NotNil(t, withoutRoles.GroupRoles)
}
