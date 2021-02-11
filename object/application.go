// Copyright 2020 The casbin Authors. All Rights Reserved.
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
	"xorm.io/core"
)

type Application struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName  string   `xorm:"varchar(100)" json:"displayName"`
	Logo         string   `xorm:"varchar(100)" json:"logo"`
	Organization string   `xorm:"varchar(100)" json:"organization"`
	Providers    []string `xorm:"varchar(100)" json:"providers"`
}

func GetApplications(owner string) []*Application {
	applications := []*Application{}
	err := adapter.engine.Desc("created_time").Find(&applications, &Application{Owner: owner})
	if err != nil {
		panic(err)
	}

	return applications
}

func getApplication(owner string, name string) *Application {
	application := Application{Owner: owner, Name: name}
	existed, err := adapter.engine.Get(&application)
	if err != nil {
		panic(err)
	}

	if existed {
		return &application
	} else {
		return nil
	}
}

func GetApplication(id string) *Application {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getApplication(owner, name)
}

func UpdateApplication(id string, application *Application) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getApplication(owner, name) == nil {
		return false
	}

	_, err := adapter.engine.Id(core.PK{owner, name}).AllCols().Update(application)
	if err != nil {
		panic(err)
	}

	//return affected != 0
	return true
}

func AddApplication(application *Application) bool {
	affected, err := adapter.engine.Insert(application)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteApplication(application *Application) bool {
	affected, err := adapter.engine.Id(core.PK{application.Owner, application.Name}).Delete(&Application{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}
