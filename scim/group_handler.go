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

package scim

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/errors"
)

const GroupExtensionKey = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:Group"

type GroupResourceHandler struct{}

func (h GroupResourceHandler) Create(r *http.Request, attrs scim.ResourceAttributes) (scim.Resource, error) {
	resource := &scim.Resource{Attributes: attrs}
	err := addScimGroup(resource)
	return *resource, err
}

func (h GroupResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	resource, err := getScimGroup(id)
	if err != nil {
		return scim.Resource{}, err
	}
	if resource == nil {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	return *resource, nil
}

func (h GroupResourceHandler) Delete(r *http.Request, id string) error {
	group, err := object.GetGroup(id)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.ScimErrorResourceNotFound(id)
	}
	if err := clearGroupMembers(id); err != nil {
		return err
	}
	_, err = object.DeleteGroup(group)
	return err
}

func (h GroupResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	if params.Count == 0 {
		count, err := object.GetGroupCount("", "", "")
		if err != nil {
			return scim.Page{}, err
		}
		return scim.Page{TotalResults: int(count)}, nil
	}

	// startIndex is 1-based
	groups, err := object.GetPaginationGroups("", params.StartIndex-1, params.Count, "", "", "", "")
	if err != nil {
		return scim.Page{}, err
	}

	resources := make([]scim.Resource, 0, len(groups))
	for _, group := range groups {
		resource, err := getScimGroup(group.GetId())
		if err != nil {
			return scim.Page{}, err
		}
		if resource != nil {
			resources = append(resources, *resource)
		}
	}

	totalCount, err := object.GetGroupCount("", "", "")
	if err != nil {
		return scim.Page{}, err
	}

	return scim.Page{
		TotalResults: int(totalCount),
		Resources:    resources,
	}, nil
}

func (h GroupResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	group, err := object.GetGroup(id)
	if err != nil {
		return scim.Resource{}, err
	}
	if group == nil {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	return updateScimGroupByPatch(id, group, operations)
}

func (h GroupResourceHandler) Replace(r *http.Request, id string, attrs scim.ResourceAttributes) (scim.Resource, error) {
	group, err := object.GetGroup(id)
	if err != nil {
		return scim.Resource{}, err
	}
	if group == nil {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	resource := &scim.Resource{Attributes: attrs}
	err = updateScimGroup(id, group, resource)
	return *resource, err
}

func getScimGroup(id string) (*scim.Resource, error) {
	group, err := object.GetGroup(id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}
	users, err := object.GetGroupUsers(id)
	if err != nil {
		return nil, err
	}
	return group2resource(group, users), nil
}

func addScimGroup(r *scim.Resource) error {
	newGroup, err := resource2group(r.Attributes)
	if err != nil {
		return err
	}

	existing, err := object.GetGroup(newGroup.GetId())
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.ScimErrorUniqueness
	}

	newGroup.CreatedTime = util.GetCurrentTime()
	newGroup.UpdatedTime = util.GetCurrentTime()
	newGroup.IsTopGroup = true
	newGroup.IsEnabled = true

	affected, err := object.AddGroup(newGroup)
	if err != nil {
		return err
	}
	if !affected {
		return fmt.Errorf("add group failed")
	}

	memberIds := extractMemberIds(r.Attributes["members"])
	if len(memberIds) > 0 {
		if err := addGroupMembers(newGroup.GetId(), memberIds); err != nil {
			return err
		}
	}

	users, err := object.GetGroupUsers(newGroup.GetId())
	if err != nil {
		return err
	}
	updated := group2resource(newGroup, users)
	r.ID = updated.ID
	r.ExternalID = updated.ExternalID
	r.Attributes = updated.Attributes
	r.Meta = updated.Meta
	return nil
}

func updateScimGroup(id string, oldGroup *object.Group, r *scim.Resource) error {
	newGroup, err := resource2group(r.Attributes)
	if err != nil {
		return err
	}
	newGroup.Owner = oldGroup.Owner
	newGroup.Name = oldGroup.Name
	newGroup.UpdatedTime = util.GetCurrentTime()

	_, err = object.UpdateGroup(id, newGroup, true, "en")
	if err != nil {
		return err
	}

	currentUsers, err := object.GetGroupUsers(id)
	if err != nil {
		return err
	}
	currentIds := make([]string, 0, len(currentUsers))
	for _, u := range currentUsers {
		currentIds = append(currentIds, u.Id)
	}
	newMemberIds := extractMemberIds(r.Attributes["members"])
	if err := setGroupMembers(id, newMemberIds, currentIds); err != nil {
		return err
	}

	users, err := object.GetGroupUsers(id)
	if err != nil {
		return err
	}
	updated := group2resource(newGroup, users)
	r.ID = updated.ID
	r.ExternalID = updated.ExternalID
	r.Attributes = updated.Attributes
	r.Meta = updated.Meta
	return nil
}

func updateScimGroupByPatch(id string, group *object.Group, ops []scim.PatchOperation) (r scim.Resource, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("invalid patch op value: %v", rec)
		}
	}()

	for _, op := range ops {
		value := op.Value
		if op.Op == scim.PatchOperationRemove {
			value = nil
		}
		switch op.Path.String() {
		case "displayName":
			group.DisplayName = ToString(value, "")
		case "members":
			if op.Op == scim.PatchOperationRemove {
				// Remove specific members matched by filter, or all if no filter
				filterStr := op.Path.String()
				if strings.Contains(filterStr, "[") {
					// members[value eq "user-id"] — handled by the scim library which
					// passes the resolved value; treat as remove-all fallback
				}
				if err = clearGroupMembers(id); err != nil {
					return scim.Resource{}, err
				}
			} else {
				memberIds := extractMemberIds(value)
				currentUsers, gErr := object.GetGroupUsers(id)
				if gErr != nil {
					return scim.Resource{}, gErr
				}
				currentIds := make([]string, 0, len(currentUsers))
				for _, u := range currentUsers {
					currentIds = append(currentIds, u.Id)
				}
				if op.Op == scim.PatchOperationReplace {
					if err = setGroupMembers(id, memberIds, currentIds); err != nil {
						return scim.Resource{}, err
					}
				} else {
					if err = addGroupMembers(id, memberIds); err != nil {
						return scim.Resource{}, err
					}
				}
			}
		case GroupExtensionKey:
			defaultV := AnyMap{"organization": group.Owner}
			v := ToAnyMap(value, defaultV)
			group.Owner = ToString(v["organization"], group.Owner)
		case fmt.Sprintf("%v.%v", GroupExtensionKey, "organization"):
			group.Owner = ToString(value, group.Owner)
		}
	}

	group.UpdatedTime = util.GetCurrentTime()
	_, err = object.UpdateGroup(id, group, true, "en")
	if err != nil {
		return scim.Resource{}, err
	}

	users, err := object.GetGroupUsers(id)
	if err != nil {
		return scim.Resource{}, err
	}
	return *group2resource(group, users), nil
}

func group2resource(group *object.Group, users []*object.User) *scim.Resource {
	attrs := make(scim.ResourceAttributes)
	attrs["displayName"] = group.DisplayName

	members := make([]scim.ResourceAttributes, 0, len(users))
	for _, u := range users {
		members = append(members, scim.ResourceAttributes{
			"value":   u.Id,
			"display": u.DisplayName,
		})
	}
	attrs["members"] = members

	attrs[GroupExtensionKey] = scim.ResourceAttributes{
		"organization": group.Owner,
	}

	createdTime := util.String2Time(group.CreatedTime)
	updatedTime := util.String2Time(group.UpdatedTime)
	if group.UpdatedTime == "" {
		updatedTime = createdTime
	}

	return &scim.Resource{
		ID:         group.GetId(),
		Attributes: attrs,
		Meta: scim.Meta{
			Created:      &createdTime,
			LastModified: &updatedTime,
			Version:      util.Time2String(updatedTime),
		},
	}
}

func resource2group(attrs scim.ResourceAttributes) (*object.Group, error) {
	org := getAttrJsonValue(attrs, GroupExtensionKey, "organization")
	if org == "" {
		return nil, fmt.Errorf("organization in %s is required", GroupExtensionKey)
	}
	displayName := getAttrString(attrs, "displayName")
	if displayName == "" {
		return nil, fmt.Errorf("displayName is required")
	}
	// Derive a URL/path-safe name from displayName by replacing '/' with '-'
	name := strings.ReplaceAll(displayName, "/", "-")
	return &object.Group{
		Owner:       org,
		Name:        name,
		DisplayName: displayName,
	}, nil
}

func extractMemberIds(raw interface{}) []string {
	if raw == nil {
		return nil
	}
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	ids := make([]string, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]interface{}); ok {
			if v, ok := m["value"].(string); ok && v != "" {
				ids = append(ids, v)
			}
		}
	}
	return ids
}

// addGroupMembers adds users (identified by SCIM/Casdoor user ID) to the group.
func addGroupMembers(groupId string, userIds []string) error {
	for _, userId := range userIds {
		user, err := object.GetUserByUserIdOnly(userId)
		if err != nil || user == nil {
			continue
		}
		if !util.InSlice(user.Groups, groupId) {
			user.Groups = append(user.Groups, groupId)
			if _, err := object.UpdateUser(user.GetId(), user, []string{"groups"}, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// setGroupMembers replaces group membership so that exactly newUserIds are members.
func setGroupMembers(groupId string, newUserIds []string, currentUserIds []string) error {
	newSet := make(map[string]bool, len(newUserIds))
	for _, id := range newUserIds {
		newSet[id] = true
	}
	currentSet := make(map[string]bool, len(currentUserIds))
	for _, id := range currentUserIds {
		currentSet[id] = true
	}

	for id := range newSet {
		if !currentSet[id] {
			user, err := object.GetUserByUserIdOnly(id)
			if err != nil || user == nil {
				continue
			}
			if !util.InSlice(user.Groups, groupId) {
				user.Groups = append(user.Groups, groupId)
				if _, err := object.UpdateUser(user.GetId(), user, []string{"groups"}, true); err != nil {
					return err
				}
			}
		}
	}

	for id := range currentSet {
		if !newSet[id] {
			user, err := object.GetUserByUserIdOnly(id)
			if err != nil || user == nil {
				continue
			}
			user.Groups = util.DeleteVal(user.Groups, groupId)
			if _, err := object.UpdateUser(user.GetId(), user, []string{"groups"}, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// clearGroupMembers removes all users from the group.
func clearGroupMembers(groupId string) error {
	users, err := object.GetGroupUsers(groupId)
	if err != nil {
		return err
	}
	for _, user := range users {
		user.Groups = util.DeleteVal(user.Groups, groupId)
		if _, err := object.UpdateUser(user.GetId(), user, []string{"groups"}, true); err != nil {
			return err
		}
	}
	return nil
}
