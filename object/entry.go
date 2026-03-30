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

import "strconv"

type Entry struct {
	Id int64 `xorm:"pk autoincr" json:"id"`

	Owner       string `xorm:"varchar(100)" json:"owner"`
	Name        string `xorm:"varchar(100)" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	User         string `xorm:"varchar(100)" json:"user"`
	Language     string `xorm:"varchar(100)" json:"language"`

	Time      string `xorm:"varchar(100)" json:"time"`
	Message   string `xorm:"mediumtext" json:"message"`
	AgentName string `xorm:"varchar(100) index" json:"agent"`
}

func getEntry(id int64) (*Entry, error) {
	if id == 0 {
		return nil, nil
	}

	entry := Entry{Id: id}
	existed, err := ormer.Engine.Get(&entry)
	if err != nil {
		return nil, err
	}

	if existed {
		return &entry, nil
	}
	return nil, nil
}

func GetEntry(id int64) (*Entry, error) {
	return getEntry(id)
}

func GetEntries(owner string, agentName string) ([]*Entry, error) {
	entries := []*Entry{}
	session := GetSession(owner, -1, -1, "", "", "", "")
	if agentName != "" {
		session = session.And("agent_name = ?", agentName)
	}

	err := session.Find(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func UpdateEntry(id int64, entry *Entry) (bool, error) {
	if e, err := getEntry(id); err != nil {
		return false, err
	} else if e == nil {
		return false, nil
	}

	entry.Id = id
	affected, err := ormer.Engine.ID(id).AllCols().Update(entry)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddEntry(entry *Entry) (bool, error) {
	affected, err := ormer.Engine.Insert(entry)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteEntry(entry *Entry) (bool, error) {
	affected, err := ormer.Engine.ID(entry.Id).Delete(&Entry{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (entry *Entry) GetId() string {
	return strconv.FormatInt(entry.Id, 10)
}

func GetEntryCount(owner, agentName, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	if agentName != "" {
		session = session.And("agent_name = ?", agentName)
	}

	return session.Count(&Entry{})
}

func GetPaginationEntries(owner, agentName string, offset, limit int, field, value, sortField, sortOrder string) ([]*Entry, error) {
	entries := []*Entry{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	if agentName != "" {
		session = session.And("agent_name = ?", agentName)
	}

	err := session.Find(&entries)
	if err != nil {
		return entries, err
	}

	return entries, nil
}
