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

	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Invitation struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Code      string `xorm:"varchar(100) index" json:"code"`
	IsRegexp  bool   `json:"isRegexp"`
	Quota     int    `json:"quota"`
	UsedCount int    `json:"usedCount"`

	Application string `xorm:"varchar(100)" json:"application"`
	Username    string `xorm:"varchar(100)" json:"username"`
	Email       string `xorm:"varchar(100)" json:"email"`
	Phone       string `xorm:"varchar(100)" json:"phone"`

	SignupGroup string `xorm:"varchar(100)" json:"signupGroup"`
	DefaultCode string `xorm:"varchar(100)" json:"defaultCode"`

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

func GetInvitationByCode(code string, organizationName string, lang string) (*Invitation, string) {
	invitations, err := GetInvitations(organizationName)
	if err != nil {
		return nil, err.Error()
	}
	errMsg := ""
	for _, invitation := range invitations {
		if isValid, msg := invitation.SimpleCheckInvitationCode(code, lang); isValid {
			return invitation, msg
		} else if msg != "" && errMsg == "" {
			errMsg = msg
		}
	}

	if errMsg != "" {
		return nil, errMsg
	} else {
		return nil, i18n.Translate(lang, "check:Invitation code is invalid")
	}
}

func GetMaskedInvitation(invitation *Invitation) *Invitation {
	if invitation == nil {
		return nil
	}

	invitation.CreatedTime = ""
	invitation.UpdatedTime = ""
	invitation.Code = "***"
	invitation.DefaultCode = "***"
	invitation.IsRegexp = false
	invitation.Quota = -1
	invitation.UsedCount = -1
	invitation.SignupGroup = ""

	return invitation
}

func UpdateInvitation(id string, invitation *Invitation, lang string) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if p, err := getInvitation(owner, name); err != nil {
		return false, err
	} else if p == nil {
		return false, nil
	}

	if isRegexp, err := util.IsRegexp(invitation.Code); err != nil {
		return false, err
	} else {
		invitation.IsRegexp = isRegexp
	}

	err := CheckInvitationDefaultCode(invitation.Code, invitation.DefaultCode, lang)
	if err != nil {
		return false, err
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(invitation)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddInvitation(invitation *Invitation, lang string) (bool, error) {
	if isRegexp, err := util.IsRegexp(invitation.Code); err != nil {
		return false, err
	} else {
		invitation.IsRegexp = isRegexp
	}

	err := CheckInvitationDefaultCode(invitation.Code, invitation.DefaultCode, lang)
	if err != nil {
		return false, err
	}

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

func (invitation *Invitation) SimpleCheckInvitationCode(invitationCode string, lang string) (bool, string) {
	if matched, err := util.IsInvitationCodeMatch(invitation.Code, invitationCode); err != nil {
		return false, err.Error()
	} else if !matched {
		return false, ""
	}

	if invitation.State != "Active" {
		return false, i18n.Translate(lang, "check:Invitation code suspended")
	}
	if invitation.UsedCount >= invitation.Quota {
		return false, i18n.Translate(lang, "check:Invitation code exhausted")
	}

	// Determine whether the invitation code is in the form of a regular expression other than pure numbers and letters
	if invitation.IsRegexp {
		user, _ := GetUserByInvitationCode(invitation.Owner, invitationCode)
		if user != nil {
			return false, i18n.Translate(lang, "check:The invitation code has already been used")
		}
	}
	return true, ""
}

func (invitation *Invitation) IsInvitationCodeValid(application *Application, invitationCode string, username string, email string, phone string, lang string) (bool, string) {
	if isValid, msg := invitation.SimpleCheckInvitationCode(invitationCode, lang); !isValid {
		return false, msg
	}
	if application.IsSignupItemRequired("Username") && invitation.Username != "" && invitation.Username != username {
		return false, i18n.Translate(lang, "check:Please register using the username corresponding to the invitation code")
	}
	if application.IsSignupItemRequired("Email") && invitation.Email != "" && invitation.Email != email {
		return false, i18n.Translate(lang, "check:Please register using the email corresponding to the invitation code")
	}
	if application.IsSignupItemRequired("Phone") && invitation.Phone != "" && invitation.Phone != phone {
		return false, i18n.Translate(lang, "check:Please register using the phone  corresponding to the invitation code")
	}
	return true, ""
}
