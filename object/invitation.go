// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

type Invitation struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Code      string `xorm:"varchar(100)" json:"code"`
	Quota     int    `json:"quota"`
	UsedCount int    `json:"usedCount"`

	Application string `xorm:"varchar(100)" json:"application"`
	Username    string `xorm:"varchar(100)" json:"username"`
	Email       string `xorm:"varchar(100)" json:"email"`
	Phone       string `xorm:"varchar(100)" json:"phone"`

	SignupGroup string `xorm:"varchar(100)" json:"signupGroup"`

	State string `xorm:"varchar(100)" json:"state"`
}

func GetInvitationCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Invitation{})
}

func GetInvitations(owner string) ([]*Invitation, error) {
	invitations := []*Invitation{}
	err := ormer.Engine.Desc("created_time").Find(&invitations, &Invitation{Owner: owner})
	if err != nil {
		return invitations, err
	}

	return invitations, nil
}

func GetPaginationInvitations(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Invitation, error) {
	invitations := []*Invitation{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&invitations)
	if err != nil {
		return invitations, err
	}

	return invitations, nil
}

func getInvitation(owner string, name string) (*Invitation, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	invitation := Invitation{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&invitation)
	if err != nil {
		return &invitation, nil
	}

	if existed {
		return &invitation, nil
	} else {
		return nil, nil
	}
}

func GetInvitation(id string) (*Invitation, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getInvitation(owner, name)
}

func UpdateInvitation(id string, invitation *Invitation) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if p, err := getInvitation(owner, name); err != nil {
		return false, err
	} else if p == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(invitation)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddInvitation(invitation *Invitation) (bool, error) {
	affected, err := ormer.Engine.Insert(invitation)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteInvitation(invitation *Invitation) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{invitation.Owner, invitation.Name}).Delete(&Invitation{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (invitation *Invitation) GetId() string {
	return fmt.Sprintf("%s/%s", invitation.Owner, invitation.Name)
}

func VerifyInvitation(id string) (payment *Payment, attachInfo map[string]interface{}, err error) {
	return nil, nil, fmt.Errorf("the invitation: %s does not exist", id)
}
