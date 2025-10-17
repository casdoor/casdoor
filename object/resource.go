// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/v2/util"
	"github.com/xorm-io/core"
)

type Resource struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(180) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	User        string `xorm:"varchar(100)" json:"user"`
	Provider    string `xorm:"varchar(100)" json:"provider"`
	Application string `xorm:"varchar(100)" json:"application"`
	Tag         string `xorm:"varchar(100)" json:"tag"`
	Parent      string `xorm:"varchar(100)" json:"parent"`
	FileName    string `xorm:"varchar(255)" json:"fileName"`
	FileType    string `xorm:"varchar(100)" json:"fileType"`
	FileFormat  string `xorm:"varchar(100)" json:"fileFormat"`
	FileSize    int    `json:"fileSize"`
	Url         string `xorm:"varchar(500)" json:"url"`
	Description string `xorm:"varchar(255)" json:"description"`
}

func GetResourceCount(owner, user, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Resource{User: user})
}

func GetResources(owner string, user string) ([]*Resource, error) {
	if owner == "built-in" || owner == "" {
		owner = ""
		user = ""
	}

	resources := []*Resource{}
	err := ormer.Engine.Desc("created_time").Find(&resources, &Resource{Owner: owner, User: user})
	if err != nil {
		return resources, err
	}

	return resources, err
}

func GetPaginationResources(owner, user string, offset, limit int, field, value, sortField, sortOrder string) ([]*Resource, error) {
	if owner == "built-in" || owner == "" {
		owner = ""
		user = ""
	}

	resources := []*Resource{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&resources, &Resource{User: user})
	if err != nil {
		return resources, err
	}

	return resources, nil
}

func getResource(owner string, name string) (*Resource, error) {
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}

	resource := Resource{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&resource)
	if err != nil {
		return &resource, err
	}

	if existed {
		return &resource, nil
	}

	return nil, nil
}

func GetResource(id string) (*Resource, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getResource(owner, name)
}

func UpdateResource(id string, resource *Resource) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if r, err := getResource(owner, name); err != nil {
		return false, err
	} else if r == nil {
		return false, nil
	}

	_, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(resource)
	if err != nil {
		return false, err
	}

	// return affected != 0
	return true, nil
}

func AddResource(resource *Resource) (bool, error) {
	affected, err := ormer.Engine.Insert(resource)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteResource(resource *Resource) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{resource.Owner, resource.Name}).Delete(&Resource{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (resource *Resource) GetId() string {
	return fmt.Sprintf("%s/%s", resource.Owner, resource.Name)
}

func AddOrUpdateResource(resource *Resource) (bool, error) {
	if r, err := getResource(resource.Owner, resource.Name); err != nil {
		return false, err
	} else if r == nil {
		return AddResource(resource)
	} else {
		return UpdateResource(resource.GetId(), resource)
	}
}
