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
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

const (
	settingOwner = "admin"
	settingName  = "setting-built-in"
)

// Setting is the single row (admin/setting-built-in) of the site-wide settings edited by global admins
type Setting struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	// 0 keeps the database/sql default
	DbMaxOpenConns    int `json:"dbMaxOpenConns"`
	DbMaxIdleConns    int `json:"dbMaxIdleConns"`
	DbConnMaxLifetime int `json:"dbConnMaxLifetime"`

	// applies to the organizations that do not set their own record retention days, 0 keeps the records forever
	RecordRetentionDays int `json:"recordRetentionDays"`

	// hides the links to Casdoor Cloud that the console shows its admins
	HideCloudPromotion bool `json:"hideCloudPromotion"`
}

func GetSetting() (*Setting, error) {
	setting := Setting{Owner: settingOwner, Name: settingName}
	existed, err := ormer.Engine.Get(&setting)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}
	return &setting, nil
}

func UpdateSetting(setting *Setting) (bool, error) {
	oldSetting, err := GetSetting()
	if err != nil {
		return false, err
	}
	if oldSetting == nil {
		return false, nil
	}

	err = setting.check()
	if err != nil {
		return false, err
	}

	setting.Owner = settingOwner
	setting.Name = settingName
	setting.CreatedTime = oldSetting.CreatedTime
	setting.UpdatedTime = util.GetCurrentTime()

	affected, err := ormer.Engine.ID(core.PK{settingOwner, settingName}).AllCols().Update(setting)
	if err != nil {
		return false, err
	}

	setting.apply()
	return affected != 0, nil
}

func (setting *Setting) check() error {
	if setting.DbMaxOpenConns < 0 || setting.DbMaxIdleConns < 0 || setting.DbConnMaxLifetime < 0 {
		return fmt.Errorf("the connection pool settings cannot be negative")
	}
	if setting.DbMaxOpenConns > 0 && setting.DbMaxIdleConns > setting.DbMaxOpenConns {
		return fmt.Errorf("the max idle connections cannot be more than the max open connections")
	}
	if setting.RecordRetentionDays < 0 {
		return fmt.Errorf("the record retention days cannot be negative")
	}
	return nil
}

// apply only touches the main database's pool, the syncers' databases keep their own
func (setting *Setting) apply() {
	ormer.Engine.SetMaxOpenConns(setting.DbMaxOpenConns)

	maxIdleConns := setting.DbMaxIdleConns
	if maxIdleConns == 0 {
		// 0 would keep no idle connection at all, while database/sql keeps 2 by default
		maxIdleConns = 2
	}
	ormer.Engine.SetMaxIdleConns(maxIdleConns)

	ormer.Engine.SetConnMaxLifetime(time.Duration(setting.DbConnMaxLifetime) * time.Second)

	conf.SetHideCloudPromotion(setting.HideCloudPromotion)
}

func initBuiltInSetting() *Setting {
	setting, err := GetSetting()
	if err != nil {
		panic(err)
	}
	if setting != nil {
		return setting
	}

	currentTime := util.GetCurrentTime()
	setting = &Setting{
		Owner:       settingOwner,
		Name:        settingName,
		CreatedTime: currentTime,
		UpdatedTime: currentTime,
	}
	_, err = ormer.Engine.Insert(setting)
	if err != nil {
		panic(err)
	}
	return setting
}

// InitSetting applies the setting at startup, then rereads it every minute so that a change saved on
// another Casdoor instance sharing the same database also takes effect here
func InitSetting() {
	initBuiltInSetting().apply()

	util.SafeGoroutine(func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			setting, err := GetSetting()
			if err != nil {
				fmt.Printf("InitSetting() error: %s\n", err.Error())
				continue
			}
			if setting != nil {
				setting.apply()
			}
		}
	})
}
