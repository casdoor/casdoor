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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Entry struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Provider    string `xorm:"varchar(100)" json:"provider"`
	Application string `xorm:"varchar(100)" json:"application"`

	Type    string `xorm:"varchar(100)" json:"type"`
	Message string `xorm:"mediumtext" json:"message"`
}


func GetEntries(owner string) ([]*Entry, error) {
	entries := []*Entry{}
	err := ormer.Engine.Desc("created_time").Find(&entries, &Entry{Owner: owner})
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func getEntry(owner string, name string) (*Entry, error) {
	entry := Entry{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&entry)
	if err != nil {
		return nil, err
	}

	if existed {
		return &entry, nil
	}
	return nil, nil
}

func GetEntry(id string) (*Entry, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getEntry(owner, name)
}

func UpdateEntry(id string, entry *Entry) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if e, err := getEntry(owner, name); err != nil {
		return false, err
	} else if e == nil {
		return false, nil
	}

	entry.UpdatedTime = util.GetCurrentTime()

	_, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(entry)
	if err != nil {
		return false, err
	}

	return true, nil
}

func AddEntry(entry *Entry) (bool, error) {
	affected, err := ormer.Engine.Insert(entry)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteEntry(entry *Entry) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{entry.Owner, entry.Name}).Delete(&Entry{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (entry *Entry) GetId() string {
	return fmt.Sprintf("%s/%s", entry.Owner, entry.Name)
}

func GetEntryCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Entry{})
}

func GetPaginationEntries(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Entry, error) {
	entries := []*Entry{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&entries)
	if err != nil {
		return entries, err
	}

	return entries, nil
}
