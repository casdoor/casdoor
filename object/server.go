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

	Url         string `xorm:"varchar(100)" json:"url"`
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
	err := session.Where("owner = ? or owner = ?", "admin", owner).Find(&servers)
	if err != nil {
		return servers, err
	}

	return servers, nil
}
