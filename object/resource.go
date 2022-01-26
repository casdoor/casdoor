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

	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Resource struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	User        string `xorm:"varchar(100)" json:"user"`
	Provider    string `xorm:"varchar(100)" json:"provider"`
	Application string `xorm:"varchar(100)" json:"application"`
	Tag         string `xorm:"varchar(100)" json:"tag"`
	Parent      string `xorm:"varchar(100)" json:"parent"`
	FileName    string `xorm:"varchar(100)" json:"fileName"`
	FileType    string `xorm:"varchar(100)" json:"fileType"`
	FileFormat  string `xorm:"varchar(100)" json:"fileFormat"`
	FileSize    int    `json:"fileSize"`
	Url         string `xorm:"varchar(1000)" json:"url"`
	Description string `xorm:"varchar(1000)" json:"description"`
}

func GetResourceCount(owner, user, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Resource{User: user})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetResources(owner string, user string) []*Resource {
	if owner == "built-in" {
		owner = ""
		user = ""
	}

	resources := []*Resource{}
	err := adapter.Engine.Desc("created_time").Find(&resources, &Resource{Owner: owner, User: user})
	if err != nil {
		panic(err)
	}

	return resources
}

func GetPaginationResources(owner, user string, offset, limit int, field, value, sortField, sortOrder string) []*Resource {
	if owner == "built-in" {
		owner = ""
		user = ""
	}

	resources := []*Resource{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&resources, &Resource{User: user})
	if err != nil {
		panic(err)
	}

	return resources
}

func getResource(owner string, name string) *Resource {
	resource := Resource{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&resource)
	if err != nil {
		panic(err)
	}

	if existed {
		return &resource
	}

	return nil
}

func GetResource(id string) *Resource {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getResource(owner, name)
}

func UpdateResource(id string, resource *Resource) bool {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if getResource(owner, name) == nil {
		return false
	}

	_, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(resource)
	if err != nil {
		panic(err)
	}

	//return affected != 0
	return true
}

func AddResource(resource *Resource) bool {
	affected, err := adapter.Engine.Insert(resource)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteResource(resource *Resource) bool {
	affected, err := adapter.Engine.ID(core.PK{resource.Owner, resource.Name}).Delete(&Resource{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (resource *Resource) GetId() string {
	return fmt.Sprintf("%s/%s", resource.Owner, resource.Name)
}

func AddOrUpdateResource(resource *Resource) bool {
	if getResource(resource.Owner, resource.Name) == nil {
		return AddResource(resource)
	} else {
		return UpdateResource(resource.GetId(), resource)
	}
}
