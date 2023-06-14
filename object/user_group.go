package object

import (
	"errors"
	"fmt"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
)

type UserGroupRelation struct {
	UserId  string `xorm:"varchar(100) notnull pk" json:"userId"`
	GroupId string `xorm:"varchar(100) notnull pk" json:"groupId"`

	CreatedTime string `xorm:"created" json:"createdTime"`
	UpdatedTime string `xorm:"updated" json:"updatedTime"`
}

func updateUserGroupRelation(session *xorm.Session, user *User) (int64, error) {
	physicalGroupCount, err := session.Where("type = ?", "Physical").In("id", user.Groups).Count(Group{})
	if err != nil {
		return 0, err
	}
	if physicalGroupCount > 1 {
		return 0, errors.New("user can only be in one physical group")
	}

	groups := []*Group{}
	err = session.In("id", user.Groups).Find(&groups)
	if err != nil {
		return 0, err
	}
	if len(groups) != len(user.Groups) {
		return 0, errors.New("group not found")
	}

	_, err = session.Delete(&UserGroupRelation{UserId: user.Id})
	if err != nil {
		return 0, err
	}

	relations := []*UserGroupRelation{}
	for _, group := range groups {
		relations = append(relations, &UserGroupRelation{UserId: user.Id, GroupId: group.Id})
	}
	if len(relations) == 0 {
		return 1, nil
	}
	_, err = session.Insert(relations)
	if err != nil {
		return 0, err
	}

	return 1, nil
}

func RemoveUserFromGroup(owner, name, groupId string) (bool, error) {
	user, err := getUser(owner, name)
	if err != nil {
		return false, err
	}

	groups := []string{}
	for _, group := range user.Groups {
		if group != groupId {
			groups = append(groups, group)
		}
	}
	user.Groups = groups

	_, err = UpdateUser(util.GetId(owner, name), user, []string{"groups"}, false)
	if err != nil {
		return false, err
	}
	return true, nil
}

func deleteUserGroupRelation(session *xorm.Session, userId, groupId string) (int64, error) {
	affected, err := session.ID(core.PK{userId, groupId}).Delete(&UserGroupRelation{})
	return affected, err
}

func deleteRelationByUser(id string) (int64, error) {
	affected, err := adapter.Engine.Delete(&UserGroupRelation{UserId: id})
	return affected, err
}

func GetGroupUserCount(id string, field, value string) (int64, error) {
	group, err := GetGroup(id)
	if group == nil || err != nil {
		return 0, err
	}

	if field == "" && value == "" {
		return adapter.Engine.Count(UserGroupRelation{GroupId: group.Id})
	} else {
		return adapter.Engine.Table("user").
			Join("INNER", []string{"user_group_relation", "r"}, "user.id = r.user_id").
			Where("r.group_id = ?", group.Id).
			And(fmt.Sprintf("user.%s LIKE ?", util.CamelToSnakeCase(field)), "%"+value+"%").
			Count()
	}
}

func GetPaginationGroupUsers(id string, offset, limit int, field, value, sortField, sortOrder string) ([]*User, error) {
	group, err := GetGroup(id)
	if group == nil || err != nil {
		return nil, err
	}

	users := []*User{}
	session := adapter.Engine.Table("user").
		Join("INNER", []string{"user_group_relation", "r"}, "user.id = r.user_id").
		Where("r.group_id = ?", group.Id)

	if offset != -1 && limit != -1 {
		session.Limit(limit, offset)
	}

	if field != "" && value != "" {
		session = session.And(fmt.Sprintf("user.%s LIKE ?", util.CamelToSnakeCase(field)), "%"+value+"%")
	}

	if sortField == "" || sortOrder == "" {
		sortField = "created_time"
	}
	if sortOrder == "ascend" {
		session = session.Asc(fmt.Sprintf("user.%s", util.SnakeString(sortField)))
	} else {
		session = session.Desc(fmt.Sprintf("user.%s", util.SnakeString(sortField)))
	}

	err = session.Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func GetGroupUsers(id string) ([]*User, error) {
	group, err := GetGroup(id)
	if group == nil || err != nil {
		return []*User{}, err
	}

	users := []*User{}
	err = adapter.Engine.Table("user_group_relation").Join("INNER", []string{"user", "u"}, "user_group_relation.user_id = u.id").
		Where("user_group_relation.group_id = ?", group.Id).
		Find(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
}
