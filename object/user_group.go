package object

import (
	"errors"

	"github.com/xorm-io/xorm"
)

type UserGroupRelation struct {
	UserId  string `xorm:"varchar(100) notnull pk" json:"userId"`
	GroupId string `xorm:"varchar(100) notnull pk" json:"groupId"`

	CreatedTime string `xorm:"created" json:"createdTime"`
	UpdatedTime string `xorm:"updated" json:"updatedTime"`
}

func updateGroupRelation(session *xorm.Session, user *User) (int64, error) {
	groupIds := user.Groups

	physicalGroupCount, err := session.Where("type = ?", "physical").In("id", user.Groups).Count(Group{})
	if err != nil {
		return 0, err
	}
	if physicalGroupCount > 1 {
		return 0, errors.New("user can only be in one physical group")
	}

	groups := []*Group{}
	err = session.In("id", groupIds).Find(&groups)
	if err != nil {
		return 0, err
	}
	if len(groups) == 0 || len(groups) != len(groupIds) {
		return 0, nil
	}

	_, err = session.Delete(&UserGroupRelation{UserId: user.Id})
	if err != nil {
		return 0, err
	}

	relations := []*UserGroupRelation{}
	for _, group := range groups {
		relations = append(relations, &UserGroupRelation{UserId: user.Id, GroupId: group.Id})
	}
	_, err = session.Insert(relations)
	if err != nil {
		return 0, err
	}

	return 1, nil
}
