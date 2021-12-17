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

	"github.com/casbin/casdoor/util"
	"xorm.io/core"
)

type Syncer struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Organization string `xorm:"varchar(100)" json:"organization"`
	Type         string `xorm:"varchar(100)" json:"type"`

	Host             string `xorm:"varchar(100)" json:"host"`
	Port             int    `json:"port"`
	User             string `xorm:"varchar(100)" json:"user"`
	Password         string `xorm:"varchar(100)" json:"password"`
	Database         string `xorm:"varchar(100)" json:"database"`
	Table            string `xorm:"varchar(100)" json:"table"`
	AffiliationTable string `xorm:"varchar(100)" json:"affiliationTable"`
	AvatarBaseUrl    string `xorm:"varchar(100)" json:"avatarBaseUrl"`
	SyncInterval     int    `json:"syncInterval"`
	IsEnabled        bool   `json:"isEnabled"`

	Adapter *Adapter `xorm:"-" json:"-"`
}

func GetSyncerCount(owner string) int {
	count, err := adapter.Engine.Count(&Syncer{Owner: owner})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetSyncers(owner string) []*Syncer {
	syncers := []*Syncer{}
	err := adapter.Engine.Desc("created_time").Find(&syncers, &Syncer{Owner: owner})
	if err != nil {
		panic(err)
	}

	return syncers
}

func GetPaginationSyncers(owner string, offset, limit int) []*Syncer {
	syncers := []*Syncer{}
	err := adapter.Engine.Desc("created_time").Limit(limit, offset).Find(&syncers, &Syncer{Owner: owner})
	if err != nil {
		panic(err)
	}

	return syncers
}

func getSyncer(owner string, name string) *Syncer {
	if owner == "" || name == "" {
		return nil
	}

	syncer := Syncer{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&syncer)
	if err != nil {
		panic(err)
	}

	if existed {
		return &syncer
	} else {
		return nil
	}
}

func GetSyncer(id string) *Syncer {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getSyncer(owner, name)
}

func GetMaskedSyncer(syncer *Syncer) *Syncer {
	if syncer == nil {
		return nil
	}

	if syncer.Password != "" {
		syncer.Password = "***"
	}
	return syncer
}

func GetMaskedSyncers(syncers []*Syncer) []*Syncer {
	for _, syncer := range syncers {
		syncer = GetMaskedSyncer(syncer)
	}
	return syncers
}

func UpdateSyncer(id string, syncer *Syncer) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getSyncer(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(syncer)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddSyncer(syncer *Syncer) bool {
	affected, err := adapter.Engine.Insert(syncer)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteSyncer(syncer *Syncer) bool {
	affected, err := adapter.Engine.ID(core.PK{syncer.Owner, syncer.Name}).Delete(&Syncer{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (syncer *Syncer) GetId() string {
	return fmt.Sprintf("%s/%s", syncer.Owner, syncer.Name)
}
