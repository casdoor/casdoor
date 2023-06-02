package object

import (
	"errors"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
)

type UserGroupRelation struct {
	UserId string `xorm:"varchar(100) notnull pk" json:"userId"`
	GroupId string `xorm:"varchar(100) notnull pk" json:"groupId"`

	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
}

func removeUserFromGroup(userId, groupId string) error {
	_, err := adapter.Engine.Delete(&UserGroupRelation{UserId: userId, GroupId: groupId})
	return err
}

func addUserToGroup(userId, groupId, groupType string) error {
	if groupType == "physical" {
		// 删除用户在其他实体组的记录
		_, err := adapter.Engine.Delete(&UserGroupRelation{UserId: userId})
		if err != nil {
			return err
		}
	}

	// 添加用户到新组
	relation := &UserGroupRelation{UserId: userId, GroupId: groupId}
	_, err := adapter.Engine.Insert(relation)
	return err
}

func updateUser(oldUser, user *User, columns []string) (int64, error) {
	session := adapter.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return 0, err
	}

	groupIds := user.Groups

	physicalGroupCount, err := session.Where("type = ?", "physical").Count(Group{})
	if err != nil {
		return 0, err
	}
	if physicalGroupCount > 1 {
		return 0, errors.New("user can only be in one physical group")
	}

	groups := []*Group{}
	err = session.In("name", groupIds).Find(&groups)
	if err != nil {
		return 0, err
	}

	groupMap := make(map[string]bool)
	for _, group := range groups {
		groupMap[group.Name] = true
	}
	for _, groupId := range groupIds {
		if _, ok := groupMap[groupId]; !ok {
			return 0, errors.New("groupIds not exist")
		}
	}

	affected, err := updateGroupRelation(session, oldUser.GetId(), user.GetId(), groupIds)
	if err != nil {
		session.Rollback()
		return affected, err
	}

	// 更新用户信息
	affected, err = session.ID(core.PK{oldUser.Owner, oldUser.Name}).Cols(columns...).Update(user)
	if err != nil {
		session.Rollback()
		return affected, err
	}

	err = session.Commit()
	if err != nil {
		session.Rollback()
		return 0, err
	}

	return affected, nil
}

func updateGroupRelation(session *xorm.Session, oldUserId, userId string, newGroupIds []string) (int64, error) {
	_, err := session.Delete(&UserGroupRelation{UserId:oldUserId})
	if err != nil {
		return 0, err
	}

	relations := []*UserGroupRelation{}
	for _, groupId := range newGroupIds {
		relations = append(relations, &UserGroupRelation{UserId: userId, GroupId: groupId})
	}
	_, err = session.Insert(relations)
	if err != nil {
		return 0, err
	}

	return 1, nil
}
