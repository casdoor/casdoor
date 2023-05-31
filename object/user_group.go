package object

import (
	"errors"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
	"github.com/xorm-io/xorm"
)

type UserGroupRelation struct {
	UserOwner string `xorm:"varchar(100) notnull index(user_idx)" json:"userOwner"`
	UserName  string `xorm:"varchar(100) notnull index(user_idx)" json:"userName"`
	GroupName string `xorm:"varchar(100) notnull unique(group_idx)" json:"groupName"`

	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
}

func removeUserFromGroup(userId, groupName string) error {
	owner, name := util.GetOwnerAndNameFromId(userId)
	_, err := adapter.Engine.Delete(&UserGroupRelation{UserOwner: owner, UserName: name, GroupName: groupName})
	return err
}

func addUserToGroup(userId, groupName, groupType string) error {
	owner, name := util.GetOwnerAndNameFromId(userId)
	if groupType == "physical" {
		// 删除用户在其他实体组的记录
		_, err := adapter.Engine.Delete(&UserGroupRelation{UserOwner: owner, UserName: name})
		if err != nil {
			return err
		}
	}

	// 添加用户到新组
	relation := &UserGroupRelation{UserOwner: owner, UserName: name, GroupName: groupName}
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

	groupNames := user.Groups

	physicalGroupCount, err := session.Where("type = ?", "physical").Count(Group{})
	if err != nil {
		return 0, err
	}
	if physicalGroupCount > 1 {
		return 0, errors.New("user can only be in one physical group")
	}

	groups := []*Group{}
	err = session.In("name", groupNames).Find(&groups)
	if err != nil {
		return 0, err
	}

	groupMap := make(map[string]bool)
	for _, group := range groups {
		groupMap[group.Name] = true
	}
	for _, groupName := range groupNames {
		if _, ok := groupMap[groupName]; !ok {
			return 0, errors.New("groupNames not exist")
		}
	}

	affected, err := updateGroupRelation(session, oldUser.GetId(), user.GetId(), groupNames)
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

func updateGroupRelation(session *xorm.Session, oldUserId, userId string, newGroupNames []string) (int64, error) {
	oldOwner, oldName := util.GetOwnerAndNameFromId(userId)
	_, err := session.Delete(&UserGroupRelation{UserOwner: oldOwner, UserName: oldName})
	if err != nil {
		return 0, err
	}

	owner, name := util.GetOwnerAndNameFromId(userId)
	relations := []*UserGroupRelation{}
	for _, groupName := range newGroupNames {
		relations = append(relations, &UserGroupRelation{UserOwner: owner, UserName: name, GroupName: groupName})
	}
	_, err = session.Insert(relations)
	if err != nil {
		return 0, err
	}

	return 1, nil
}
