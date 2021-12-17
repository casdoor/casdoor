// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"strconv"
	"strings"

	"github.com/casbin/casdoor/util"
)

func (syncer *Syncer) getFullAvatarUrl(avatar string) string {
	if !strings.HasPrefix(avatar, "https://") {
		return fmt.Sprintf("%s%s", syncer.AvatarBaseUrl, avatar)
	}
	return avatar
}

func (syncer *Syncer) getPartialAvatarUrl(avatar string) string {
	if strings.HasPrefix(avatar, syncer.AvatarBaseUrl) {
		return avatar[len(syncer.AvatarBaseUrl):]
	}
	return avatar
}

func (syncer *Syncer) createUserFromOriginalUser(originalUser *DbUser, affiliationMap map[int]string) *User {
	affiliation := ""
	if originalUser.SchoolId != 0 {
		var ok bool
		affiliation, ok = affiliationMap[originalUser.SchoolId]
		if !ok {
			panic(fmt.Sprintf("SchoolId not found: %d", originalUser.SchoolId))
		}
	}

	user := &User{
		Owner:         syncer.Organization,
		Name:          strconv.Itoa(originalUser.Id),
		CreatedTime:   util.GetCurrentTime(),
		Id:            strconv.Itoa(originalUser.Id),
		Type:          "normal-user",
		Password:      originalUser.Password,
		DisplayName:   originalUser.Name,
		Avatar:        syncer.getFullAvatarUrl(originalUser.Avatar),
		Email:         "",
		Phone:         originalUser.Cellphone,
		Address:       []string{},
		Affiliation:   affiliation,
		Score:         originalUser.SchoolId,
		IsAdmin:       false,
		IsGlobalAdmin: false,
		IsForbidden:   originalUser.Deleted != 0,
		IsDeleted:     false,
		Properties:    map[string]string{},
	}
	return user
}

func (syncer *Syncer) createOriginalUserFromUser(user *User) *DbUser {
	deleted := 0
	if user.IsForbidden {
		deleted = 1
	}

	originalUser := &DbUser{
		Id:        util.ParseInt(user.Id),
		Name:      user.DisplayName,
		Password:  user.Password,
		Cellphone: user.Phone,
		SchoolId:  user.Score,
		Avatar:    syncer.getPartialAvatarUrl(user.Avatar),
		Deleted:   deleted,
	}
	return originalUser
}
