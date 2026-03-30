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

func GetEntries(owner string, agentName string) ([]*Entry, error) {
	entries := []*Entry{}

	session := ormer.Engine.Desc("created_time")
	if owner != "" {
		session = session.Where("owner = ?", owner)
	}
	if agentName != "" {
		if owner != "" {
			session = session.And("agent_name = ?", agentName)
		} else {
			session = session.Where("agent_name = ?", agentName)
		}
	}

	err := session.Find(&entries)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func addEntry(entry *Entry) (bool, error) {
	affected, err := ormer.Engine.Insert(entry)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
