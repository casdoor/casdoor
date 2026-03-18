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

type Server struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Url         string `xorm:"varchar(500)" json:"url"`
	Application string `xorm:"varchar(100)" json:"application"`
}

func GetServers(owner string) ([]*Server, error) {
	servers := []*Server{}
	err := ormer.Engine.Desc("created_time").Find(&servers, &Server{Owner: owner})
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func getServer(owner string, name string) (*Server, error) {
	server := Server{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&server)
	if err != nil {
		return nil, err
	}

	if existed {
		return &server, nil
	}
	return nil, nil
}

func GetServer(id string) (*Server, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	return getServer(owner, name)
}

func UpdateServer(id string, server *Server) (bool, error) {
	owner, name := util.GetOwnerAndNameFromIdNoCheck(id)
	if s, err := getServer(owner, name); err != nil {
		return false, err
	} else if s == nil {
		return false, nil
	}

	server.UpdatedTime = util.GetCurrentTime()

	_, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(server)
	if err != nil {
		return false, err
	}

	return true, nil
}

func AddServer(server *Server) (bool, error) {
	affected, err := ormer.Engine.Insert(server)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteServer(server *Server) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{server.Owner, server.Name}).Delete(&Server{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (server *Server) GetId() string {
	return fmt.Sprintf("%s/%s", server.Owner, server.Name)
}

func GetServerCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Server{})
}

func GetPaginationServers(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Server, error) {
	servers := []*Server{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&servers)
	if err != nil {
		return servers, err
	}

	return servers, nil
}
