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
	"fmt"
	"sort"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/util"
)

func hasWechatUnionId(user *User, unionId string) bool {
	if user.WeChat == unionId || user.Properties[UserPropertiesWechatUnionId] == unionId {
		return true
	}
	for key, value := range user.Properties {
		if strings.HasPrefix(key, "oauth_") && strings.HasSuffix(key, "_extra") && strings.Contains(value, fmt.Sprintf(`"wechat_unionid":"%s"`, unionId)) {
			return true
		}
	}
	return false
}

// GetUserByWechatUnionId finds the member that signed in through another WeChat app (a mini program, an official
// account or a website app) of the same WeChat open platform account, when the organization's policy merges members
// by unionid. The account found gets the unionid as its WeChat ID, so the next sign-ins find it directly. When
// several accounts carry the unionid, the earliest one is used and the administrators are alerted.
func GetUserByWechatUnionId(owner string, unionId string) (*User, error) {
	if unionId == "" {
		return nil, nil
	}
	policy := GetCachedSecurityPolicy(owner)
	if policy == nil || !policy.MergeByWechatUnionId {
		return nil, nil
	}

	candidates := []*User{}
	err := ormer.Engine.Where("owner = ? and is_deleted = ?", owner, false).
		And("wechat = ? or properties like ?", unionId, "%"+unionId+"%").Find(&candidates)
	if err != nil {
		return nil, err
	}

	users := []*User{}
	for _, user := range candidates {
		if hasWechatUnionId(user, unionId) {
			users = append(users, user)
		}
	}
	if len(users) == 0 {
		return nil, nil
	}

	sort.Slice(users, func(i, j int) bool { return users[i].CreatedTime < users[j].CreatedTime })
	user := users[0]
	if len(users) > 1 {
		names := []string{}
		for _, u := range users {
			names = append(names, u.Name)
		}
		logs.Warn("the WeChat unionid %s belongs to several members of the organization: %s: %s, signing in as %s", unionId, owner, strings.Join(names, ", "), user.Name)
		if organization, err := getOrganization("admin", owner); err == nil && organization != nil {
			SendSecurityAlert(organization, fmt.Sprintf("[%s] Members to merge: %s", owner, strings.Join(names, ", ")),
				fmt.Sprintf("These members signed in through WeChat apps of the same open platform account (unionid %s), so they are the same person. The sign-ins go to %s, the earliest one. Merge or delete the others in the console.", unionId, user.Name))
		}
	}

	if user.WeChat != unionId {
		if user.Properties == nil {
			user.Properties = map[string]string{}
		}
		if user.WeChat != "" && user.Properties[UserPropertiesWechatOpenId] == "" {
			user.Properties[UserPropertiesWechatOpenId] = user.WeChat
		}
		user.Properties[UserPropertiesWechatUnionId] = unionId
		user.WeChat = unionId
		user.UpdatedTime = util.GetCurrentTime()
		_, err = ormer.Engine.Where("owner = ? and name = ?", user.Owner, user.Name).Cols("wechat", "properties", "updated_time").Update(user)
		if err != nil {
			return nil, err
		}
	}
	return user, nil
}
