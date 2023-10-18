// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"log"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
)

type AnyMap map[string]interface{}

type AnyArray []interface{}

func ToString(v interface{}, defaultV ...interface{}) string {
	if v == nil {
		if len(defaultV) > 0 {
			v = defaultV[0]
		}
	}
	return v.(string)
}

func ToAnyMap(v interface{}, defaultV ...interface{}) AnyMap {
	if v == nil {
		if len(defaultV) > 0 {
			v = defaultV[0]
		}
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		m = v.(AnyMap)
	}
	return m
}

func ToAnyArray(v interface{}, defaultV ...interface{}) AnyArray {
	if v == nil {
		if len(defaultV) > 0 {
			v = defaultV[0]
		}
	}
	m, ok := v.([]interface{})
	if !ok {
		m = v.(AnyArray)
	}
	return m
}

func newStringParams(name string, required, unique bool) schema.SimpleParams {
	uniqueness := schema.AttributeUniquenessNone()
	if unique {
		uniqueness = schema.AttributeUniquenessServer()
	}
	return schema.SimpleStringParams(schema.StringParams{
		Name:       name,
		Required:   required,
		Uniqueness: uniqueness,
	})
}

func newComplexParams(name string, required bool, multi bool, subAttributes []schema.SimpleParams) schema.ComplexParams {
	return schema.ComplexParams{
		Name:          name,
		Required:      required,
		MultiValued:   multi,
		SubAttributes: subAttributes,
	}
}

func buildExternalId(user *object.User) optional.String {
	if user.ExternalId != "" {
		return optional.NewString(user.ExternalId)
	} else {
		return optional.String{}
	}
}

func buildMeta(user *object.User) scim.Meta {
	createdTime := util.String2Time(user.CreatedTime)
	updatedTime := util.String2Time(user.UpdatedTime)
	if user.UpdatedTime == "" {
		updatedTime = createdTime
	}
	return scim.Meta{
		Created:      &createdTime,
		LastModified: &updatedTime,
		Version:      util.Time2String(updatedTime),
	}
}

func getAttrString(attrs scim.ResourceAttributes, key string) string {
	if attrs[key] == nil {
		return ""
	} else {
		return attrs[key].(string)
	}
}

func getAttrJson(attrs scim.ResourceAttributes, key string) scim.ResourceAttributes {
	if attrs[key] == nil {
		return nil
	} else {
		if v, ok := attrs[key].(map[string]interface{}); ok {
			return v
		} else if v, ok := attrs[key].([]interface{}); ok {
			if len(v) > 0 {
				return v[0].(map[string]interface{})
			} else {
				return nil
			}
		} else {
			panic("invalid attribute type")
		}
	}
}

func getAttrJsonValue(attrs scim.ResourceAttributes, key1 string, key2 string) string {
	attr := getAttrJson(attrs, key1)
	if attr == nil {
		return ""
	} else {
		return getAttrString(attr, key2)
	}
}

func user2resource(user *object.User) *scim.Resource {
	attrs := make(map[string]interface{})
	// Singular attributes
	attrs["userName"] = user.Name
	// The cleartext value or the hashed value of a password SHALL NOT be returnable by a service provider.
	// attrs["password"] = user.Password
	formatted := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	if user.FirstName == "" {
		formatted = user.LastName
	}
	if user.LastName == "" {
		formatted = user.FirstName
	}
	attrs["name"] = scim.ResourceAttributes{
		"formatted":  formatted,
		"familyName": user.LastName,
		"givenName":  user.FirstName,
	}
	attrs["displayName"] = user.DisplayName
	attrs["nickName"] = user.DisplayName
	attrs["userType"] = user.Type
	attrs["profileUrl"] = user.Homepage
	attrs["active"] = !user.IsForbidden && !user.IsDeleted

	// Multi-Valued attributes
	attrs["emails"] = []scim.ResourceAttributes{
		{
			"value": user.Email,
		},
	}
	attrs["phoneNumbers"] = []scim.ResourceAttributes{
		{
			"value": user.Phone,
		},
	}
	attrs["photos"] = []scim.ResourceAttributes{
		{
			"value": user.Avatar,
		},
	}
	attrs["addresses"] = []scim.ResourceAttributes{
		{
			"locality": user.Location,    // e.g. Hollywood
			"region":   user.Region,      // e.g. CN
			"country":  user.CountryCode, // e.g. USA
		},
	}

	// Enterprise user schema extension
	attrs[UserExtensionKey] = scim.ResourceAttributes{
		"organization": user.Owner,
	}

	return &scim.Resource{
		ID:         user.Id,
		ExternalID: buildExternalId(user),
		Attributes: attrs,
		Meta:       buildMeta(user),
	}
}

func resource2user(attrs scim.ResourceAttributes) (user *object.User, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("failed to parse attrs: %v", r)
			err = fmt.Errorf("%v", r)
		}
	}()
	user = &object.User{
		ExternalId:  getAttrString(attrs, "externalId"),
		Name:        getAttrString(attrs, "userName"),
		Password:    getAttrString(attrs, "password"),
		DisplayName: getAttrString(attrs, "displayName"),
		Homepage:    getAttrString(attrs, "profileUrl"),
		Type:        getAttrString(attrs, "userType"),

		Owner:       getAttrJsonValue(attrs, UserExtensionKey, "organization"),
		FirstName:   getAttrJsonValue(attrs, "name", "givenName"),
		LastName:    getAttrJsonValue(attrs, "name", "familyName"),
		Email:       getAttrJsonValue(attrs, "emails", "value"),
		Phone:       getAttrJsonValue(attrs, "phoneNumbers", "value"),
		Avatar:      getAttrJsonValue(attrs, "photos", "value"),
		Location:    getAttrJsonValue(attrs, "addresses", "locality"),
		Region:      getAttrJsonValue(attrs, "addresses", "region"),
		CountryCode: getAttrJsonValue(attrs, "addresses", "country"),

		CreatedTime: util.GetCurrentTime(),
		UpdatedTime: util.GetCurrentTime(),
	}

	if user.Owner == "" {
		err = fmt.Errorf("organization in %s is required", UserExtensionKey)
	}
	return
}
