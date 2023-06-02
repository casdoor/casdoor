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
	"strings"

	"github.com/casdoor/casdoor/conf"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Group struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	Id            string    `xorm:"varchar(100) not null index" json:"id"`
	DisplayName   string    `xorm:"varchar(100)" json:"displayName"`
	Manager       string    `xorm:"varchar(100)" json:"manager"`
	ContactEmail  string    `xorm:"varchar(100)" json:"contactEmail"`
	Type          string    `xorm:"varchar(100)" json:"type"`
	ParentGroupId string    `xorm:"varchar(100)" json:"parentGroupId"`
	Users         *[]string `xorm:"-" json:"users"`

	IsEnabled bool `json:"isEnabled"`
}

func GetGroupCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Group{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetGroups(owner string) []*Group {
	groups := []*Group{}
	err := adapter.Engine.Desc("created_time").Find(&groups, &Group{Owner: owner})
	if err != nil {
		panic(err)
	}

	return groups
}

func GetPaginationGroups(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Group {
	groups := []*Group{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&groups)
	if err != nil {
		panic(err)
	}

	return groups
}

func getGroup(owner string, name string) *Group {
	if owner == "" || name == "" {
		return nil
	}

	group := Group{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&group)
	if err != nil {
		panic(err)
	}

	if existed {
		return &group
	} else {
		return nil
	}
}

func GetGroup(id string) *Group {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getGroup(owner, name)
}

func UpdateGroup(id string, group *Group) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	oldGroup := getGroup(owner, name)
	if oldGroup == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(group)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddGroup(group *Group) bool {
	if group.Id == "" {
		group.Id = util.GenerateId()
	}

	affected, err := adapter.Engine.Insert(group)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddGroups(groups []*Group) bool {
	if len(groups) == 0 {
		return false
	}
	affected, err := adapter.Engine.Insert(groups)
	if err != nil {
		if !strings.Contains(err.Error(), "Duplicate entry") {
			panic(err)
		}
	}
	return affected != 0
}

func AddGroupsInBatch(groups []*Group) bool {
	batchSize := conf.GetConfigBatchSize()

	if len(groups) == 0 {
		return false
	}

	affected := false
	for i := 0; i < (len(groups)-1)/batchSize+1; i++ {
		start := i * batchSize
		end := (i + 1) * batchSize
		if end > len(groups) {
			end = len(groups)
		}

		tmp := groups[start:end]
		// TODO: save to log instead of standard output
		// fmt.Printf("Add users: [%d - %d].\n", start, end)
		if AddGroups(tmp) {
			affected = true
		}
	}

	return affected
}

func DeleteGroup(group *Group) bool {
	affected, err := adapter.Engine.ID(core.PK{group.Owner, group.Name}).Delete(&Group{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (group *Group) GetId() string {
	return fmt.Sprintf("%s/%s", group.Owner, group.Name)
}
