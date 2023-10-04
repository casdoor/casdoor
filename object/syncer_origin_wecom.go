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

package object

import (
	"fmt"
	"strconv"

	wework "github.com/go-laoji/wecom-go-sdk/v2"
)

type WeComSyncer struct {
	Client        wework.IWeWork
	CorpId        string
	AppSecret     string
	ContactSecret string
	Organization  string
}

func NewWeComSyncer(corpId string, appSecret string, contactSecret string, organization string) (*WeComSyncer, error) {
	wecom := wework.NewWeWork(wework.WeWorkConfig{
		CorpId: corpId,
	})
	wecom.SetAppSecretFunc(func(id uint) (string, string, bool) {
		if id == 0 {
			return corpId, contactSecret, true
		} else {
			return corpId, appSecret, true
		}
	})

	return &WeComSyncer{
		Client:        wecom,
		CorpId:        corpId,
		AppSecret:     appSecret,
		ContactSecret: contactSecret,
		Organization:  organization,
	}, nil
}

func (wSyncer *WeComSyncer) getOriginalUsers() ([]*OriginalUser, error) {
	wSyncer.Client.SetAppSecretFunc(func(id uint) (string, string, bool) {
		if id == 0 {
			return wSyncer.CorpId, wSyncer.AppSecret, true
		} else {
			return wSyncer.CorpId, wSyncer.ContactSecret, true
		}
	})

	userList := make([]wework.User, 0)
	var nextCursor string
	for {
		listRsp := wSyncer.Client.UserListId(1, nextCursor, 10000)
		if listRsp.ErrCode != 0 {
			return nil, fmt.Errorf(listRsp.ErrorMsg)
		}
		nextCursor = listRsp.NextCursor
		wecomUsers := listRsp.DeptUser

		for i := 0; i < len(wecomUsers); i++ {
			userRsp := wSyncer.Client.UserGet(0, wecomUsers[i].UserId)
			if userRsp.ErrCode != 0 {
				return nil, fmt.Errorf(userRsp.ErrorMsg)
			}
			userList = append(userList, userRsp.User)
		}
		if nextCursor == "" {
			break
		}
	}
	return wSyncer.getOriginalUsersFromWeCom(userList)
}

func (wSyncer *WeComSyncer) GetOriginalUserMap() ([]*OriginalUser, map[string]*OriginalUser, error) {
	users, err := wSyncer.getOriginalUsers()
	if err != nil {
		return nil, nil, err
	}

	m := map[string]*OriginalUser{}
	for _, user := range users {
		m[user.Id] = user
	}
	return users, m, nil
}

func (wSyncer *WeComSyncer) UpdateUser(oUser *OriginalUser) (bool, error) {
	wecomUser := wSyncer.getWeComUserFromOriginalUser(oUser)

	updateRsp := wSyncer.Client.UserUpdate(0, wecomUser)
	if updateRsp.ErrCode != 0 {
		return false, fmt.Errorf(updateRsp.ErrorMsg)
	}

	return true, nil
}

func (wSyncer *WeComSyncer) AddUser(oUser *OriginalUser) (bool, error) {
	wecomUser := wSyncer.getWeComUserFromOriginalUser(oUser)

	createRsp := wSyncer.Client.UserCreate(0, wecomUser)
	if createRsp.ErrCode != 0 {
		return false, fmt.Errorf(createRsp.ErrorMsg)
	}

	return true, nil
}

func (wSyncer *WeComSyncer) getOriginalGroups() ([]*OriginalGroup, error) {
	groupList := make([]wework.Department, 0)

	listRsp := wSyncer.Client.DepartmentList(0, 0)
	if listRsp.ErrCode != 0 {
		return nil, fmt.Errorf(listRsp.ErrorMsg)
	}
	groupList = append(groupList, listRsp.Department...)

	return wSyncer.getOriginalGroupFromWecom(groupList), nil
}

func (wSyncer *WeComSyncer) GetOriginalGroupMap() ([]*OriginalGroup, map[string]*OriginalGroup, error) {
	groups, err := wSyncer.getOriginalGroups()
	if err != nil {
		return nil, nil, err
	}

	m := map[string]*OriginalGroup{}
	for _, group := range groups {
		m[group.Key] = group
	}
	return groups, m, nil
}

func (wSyncer *WeComSyncer) UpdateGroup(oGroup *OriginalGroup) (bool, error) {
	wecomGroup, err := wSyncer.getWeComGroupFromOriginalGroup(oGroup)
	if err != nil {
		return false, err
	}

	updateRsp := wSyncer.Client.DepartmentUpdate(0, wecomGroup)
	if updateRsp.ErrCode != 0 {
		return false, fmt.Errorf(updateRsp.ErrorMsg)
	}

	return true, nil
}

func (wSyncer *WeComSyncer) AddGroup(oGroup *OriginalGroup) (bool, error) {
	wecomGroup, err := wSyncer.getWeComGroupFromOriginalGroup(oGroup)
	if err != nil {
		return false, err
	}

	createRsp := wSyncer.Client.DepartmentCreate(0, wecomGroup)
	if createRsp.ErrCode != 0 {
		return false, fmt.Errorf(createRsp.ErrorMsg)
	}

	return true, nil
}

func (wSyncer *WeComSyncer) GetAffiliationMap() ([]*Affiliation, map[int]string, error) {
	return nil, nil, nil
}

func (wSyncer *WeComSyncer) getWeComGroupFromOriginalGroup(oGroup *OriginalGroup) (wework.Department, error) {
	wParentId, err := strconv.Atoi(oGroup.ParentKey)
	if err != nil {
		return wework.Department{}, fmt.Errorf("Parse parentKey error")
	}

	return wework.Department{
		Name:     oGroup.Name,
		ParentId: int32(wParentId),
	}, nil
}

func (wSyncer *WeComSyncer) getOriginalGroupFromWecom(groupList []wework.Department) []*OriginalGroup {
	groups := make([]*OriginalGroup, 0)
	m := make(map[int32]string)

	for _, wecomGroup := range groupList {
		m[wecomGroup.Id] = wecomGroup.Name
		oGroup := &OriginalGroup{
			Key:         string(wecomGroup.Id),
			Name:        wecomGroup.Name,
			DisplayName: wecomGroup.Name,
			ParentKey:   string(wecomGroup.ParentId),
		}

		if parentId, ok := m[wecomGroup.ParentId]; ok {
			oGroup.ParentId = wSyncer.Organization + "/" + parentId
		}

		groups = append(groups, oGroup)
	}
	return groups
}

func (wSyncer *WeComSyncer) getWeComUserFromOriginalUser(oUsher *OriginalUser) wework.User {
	return wework.User{
		Userid:   oUsher.Id,
		Name:     oUsher.Name,
		Alias:    oUsher.DisplayName,
		Avatar:   oUsher.Avatar,
		BizEmail: oUsher.Email,
		Mobile:   oUsher.Phone,
		Address:  oUsher.Location,
		Gender:   oUsher.Gender,
	}
}

func (wSyncer *WeComSyncer) getOriginalUsersFromWeCom(weComUserList []wework.User) ([]*OriginalUser, error) {
	_, m, err := wSyncer.GetOriginalGroupMap()
	if err != nil {
		return nil, err
	}

	users := make([]*OriginalUser, 0)
	for _, wecomUser := range weComUserList {
		oUser := &OriginalUser{
			Id:          wecomUser.Userid,
			Name:        wecomUser.Name,
			DisplayName: wecomUser.Alias,
			Avatar:      wecomUser.Avatar,
			Email:       wecomUser.BizEmail,
			Phone:       wecomUser.Mobile,
			Location:    wecomUser.Address,
			Gender:      wecomUser.Gender,
			Groups:      wSyncer.getUserGroups(wecomUser.Department, m),
		}
		users = append(users, oUser)
	}
	return users, nil
}

func (wSyncer *WeComSyncer) getUserGroups(groupIdList []int32, m map[string]*OriginalGroup) []string {
	groupNameList := make([]string, 0)
	for _, groupId := range groupIdList {
		groupName := wSyncer.Organization + "/" + m[string(groupId)].Name
		groupNameList = append(groupNameList, groupName)
	}

	return groupNameList
}
