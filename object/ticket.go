// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

type TicketMessage struct {
	Author    string `json:"author"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
	IsAdmin   bool   `json:"isAdmin"`
}

type Ticket struct {
	Owner       string           `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string           `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string           `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string           `xorm:"varchar(100)" json:"updatedTime"`
	DisplayName string           `xorm:"varchar(100)" json:"displayName"`
	User        string           `xorm:"varchar(100) index" json:"user"`
	Title       string           `xorm:"varchar(200)" json:"title"`
	Content     string           `xorm:"mediumtext" json:"content"`
	State       string           `xorm:"varchar(50)" json:"state"`
	Messages    []*TicketMessage `xorm:"mediumtext json" json:"messages"`
}

func GetTicketCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Ticket{})
}

func GetTickets(owner string) ([]*Ticket, error) {
	tickets := []*Ticket{}
	err := ormer.Engine.Desc("created_time").Find(&tickets, &Ticket{Owner: owner})
	if err != nil {
		return tickets, err
	}

	return tickets, nil
}

func GetPaginationTickets(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Ticket, error) {
	tickets := []*Ticket{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&tickets)
	if err != nil {
		return tickets, err
	}

	return tickets, nil
}

func GetUserTickets(owner, user string) ([]*Ticket, error) {
	tickets := []*Ticket{}
	err := ormer.Engine.Desc("created_time").Find(&tickets, &Ticket{Owner: owner, User: user})
	if err != nil {
		return tickets, err
	}

	return tickets, nil
}

func getTicket(owner string, name string) (*Ticket, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	ticket := Ticket{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&ticket)
	if err != nil {
		return &ticket, err
	}

	if existed {
		return &ticket, nil
	}

	return nil, nil
}

func GetTicket(id string) (*Ticket, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getTicket(owner, name)
}

func UpdateTicket(id string, ticket *Ticket) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if t, err := getTicket(owner, name); err != nil {
		return false, err
	} else if t == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(ticket)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddTicket(ticket *Ticket) (bool, error) {
	affected, err := ormer.Engine.Insert(ticket)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteTicket(ticket *Ticket) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{ticket.Owner, ticket.Name}).Delete(&Ticket{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (ticket *Ticket) GetId() string {
	return fmt.Sprintf("%s/%s", ticket.Owner, ticket.Name)
}

func AddTicketMessage(id string, message *TicketMessage) (bool, error) {
	ticket, err := GetTicket(id)
	if err != nil {
		return false, err
	}
	if ticket == nil {
		return false, fmt.Errorf("ticket not found: %s", id)
	}

	if ticket.Messages == nil {
		ticket.Messages = []*TicketMessage{}
	}

	ticket.Messages = append(ticket.Messages, message)
	return UpdateTicket(id, ticket)
}
