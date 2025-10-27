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
	"github.com/casdoor/casdoor/util"
)

type WeCom struct {
	Id          string `xorm:"varchar(100) notnull pk" json:"id"`
	Owner       string `xorm:"varchar(100)" json:"owner"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	ServerName   string `xorm:"varchar(100)" json:"serverName"`
	CorpId       string `xorm:"varchar(200)" json:"corpId"`
	CorpSecret   string `xorm:"varchar(500)" json:"corpSecret"`
	DepartmentId string `xorm:"varchar(100)" json:"departmentId"` // Optional: specific department to sync
	SubType      string `xorm:"varchar(100)" json:"subType"`      // "Internal" or "Third-party"

	AutoSync int    `json:"autoSync"`
	LastSync string `xorm:"varchar(100)" json:"lastSync"`
}

func AddWeCom(weCom *WeCom) (bool, error) {
	if len(weCom.Id) == 0 {
		weCom.Id = util.GenerateId()
	}

	if len(weCom.CreatedTime) == 0 {
		weCom.CreatedTime = util.GetCurrentTime()
	}

	affected, err := ormer.Engine.Insert(weCom)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func CheckWeComExist(weCom *WeCom) (bool, error) {
	var result []*WeCom
	err := ormer.Engine.Find(&result, &WeCom{
		Owner:      weCom.Owner,
		CorpId:     weCom.CorpId,
		CorpSecret: weCom.CorpSecret,
	})
	if err != nil {
		return false, err
	}

	if len(result) > 0 {
		return true, nil
	}

	return false, nil
}

func GetWeComs(owner string) ([]*WeCom, error) {
	var weComs []*WeCom
	err := ormer.Engine.Desc("created_time").Find(&weComs, &WeCom{Owner: owner})
	if err != nil {
		return weComs, err
	}

	return weComs, nil
}

func GetWeCom(id string) (*WeCom, error) {
	if util.IsStringsEmpty(id) {
		return nil, nil
	}

	weCom := WeCom{Id: id}
	existed, err := ormer.Engine.Get(&weCom)
	if err != nil {
		return &weCom, nil
	}

	if existed {
		return &weCom, nil
	} else {
		return nil, nil
	}
}

func GetMaskedWeCom(weCom *WeCom, errs ...error) (*WeCom, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	if weCom == nil {
		return nil, nil
	}

	if weCom.CorpSecret != "" {
		weCom.CorpSecret = "***"
	}

	return weCom, nil
}

func GetMaskedWeComs(weComs []*WeCom, errs ...error) ([]*WeCom, error) {
	if len(errs) > 0 && errs[0] != nil {
		return nil, errs[0]
	}

	var err error
	for _, weCom := range weComs {
		weCom, err = GetMaskedWeCom(weCom)
		if err != nil {
			return nil, err
		}
	}
	return weComs, nil
}

func UpdateWeCom(weCom *WeCom) (bool, error) {
	var w *WeCom
	var err error
	if w, err = GetWeCom(weCom.Id); err != nil {
		return false, nil
	} else if w == nil {
		return false, nil
	}

	if weCom.CorpSecret == "***" {
		weCom.CorpSecret = w.CorpSecret
	}

	affected, err := ormer.Engine.ID(weCom.Id).AllCols().Update(weCom)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteWeCom(weCom *WeCom) (bool, error) {
	affected, err := ormer.Engine.ID(weCom.Id).Delete(&WeCom{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func GetExistIds(owner string, ids []string) ([]string, error) {
	var existIds []string

	err := ormer.Engine.Table("user").Where("owner = ?", owner).
		In("id", ids).Select("DISTINCT id").Find(&existIds)
	if err != nil {
		return existIds, err
	}

	return existIds, nil
}
