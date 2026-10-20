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
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/robfig/cron/v3"
	"github.com/xorm-io/core"
)

func (sub *Subscription) isActiveAt(now time.Time) bool {
	if sub.State != SubStateActive && sub.State != SubStateUpcoming && sub.State != SubStateExpired {
		return false
	}

	startTime, err := time.Parse(time.RFC3339, sub.StartTime)
	if err != nil {
		return false
	}

	endTime, err := time.Parse(time.RFC3339, sub.EndTime)
	if err != nil {
		return false
	}

	return !startTime.After(now) && !endTime.Before(now)
}

func getPlanRoleId(owner string, planName string) (string, error) {
	plan, err := getPlan(owner, planName)
	if err != nil || plan == nil || plan.Role == "" {
		return "", err
	}

	if strings.Contains(plan.Role, "/") {
		return plan.Role, nil
	}
	return util.GetId(plan.Owner, plan.Role), nil
}

// syncSubscriptionRole gives the plan's role to the user or group only while one of their subscriptions to it is active
func syncSubscriptionRole(sub *Subscription) error {
	if sub == nil {
		return nil
	}

	roleId, err := getPlanRoleId(sub.Owner, sub.Plan)
	if err != nil || roleId == "" {
		return err
	}

	if sub.User != "" {
		err = syncSubscriptionRoleMember(sub.Owner, roleId, "user", sub.User)
		if err != nil {
			return err
		}
	}

	if sub.Group != "" {
		err = syncSubscriptionRoleMember(sub.Owner, roleId, "group", sub.Group)
		if err != nil {
			return err
		}
	}

	return nil
}

func syncSubscriptionRoleMember(owner string, roleId string, column string, memberName string) error {
	subscriptions := []*Subscription{}
	err := ormer.Engine.Where("owner = ?", owner).And(fmt.Sprintf("%s = ?", quoteColumn(column)), memberName).Find(&subscriptions)
	if err != nil {
		return err
	}

	isActive := false
	now := time.Now()
	for _, sub := range subscriptions {
		if !sub.isActiveAt(now) {
			continue
		}

		subRoleId, err := getPlanRoleId(sub.Owner, sub.Plan)
		if err != nil {
			return err
		}
		if subRoleId == roleId {
			isActive = true
			break
		}
	}

	role, err := GetRole(roleId)
	if err != nil || role == nil {
		return err
	}

	memberId := util.GetId(owner, memberName)
	members, roleColumn := role.Users, "users"
	if column == "group" {
		members, roleColumn = role.Groups, "groups"
	}

	if util.InSlice(members, memberId) == isActive {
		return nil
	}

	if isActive {
		members = append(members, memberId)
	} else {
		members = util.DeleteVal(members, memberId)
	}

	if column == "group" {
		role.Groups = members
	} else {
		role.Users = members
	}

	_, err = ormer.Engine.ID(core.PK{role.Owner, role.Name}).Cols(roleColumn).Update(role)
	return err
}

func updatePaymentSubscriptions(payment *Payment) error {
	subscriptions := []*Subscription{}
	err := ormer.Engine.Where("owner = ? and payment = ?", payment.Owner, payment.Name).Find(&subscriptions)
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		err = sub.UpdateState()
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateSubscriptionStates refreshes the unfinished subscriptions so that their plans' roles are granted and taken back on time
func UpdateSubscriptionStates() error {
	subscriptions := []*Subscription{}
	err := ormer.Engine.In("state", string(SubStatePending), string(SubStateUpcoming), string(SubStateActive)).Find(&subscriptions)
	if err != nil {
		return err
	}

	for _, sub := range subscriptions {
		err = sub.UpdateState()
		if err != nil {
			fmt.Printf("UpdateSubscriptionStates() error for subscription %s: %v\n", sub.GetId(), err)
			continue
		}

		if sub.State == SubStateActive {
			err = syncSubscriptionRole(sub)
			if err != nil {
				fmt.Printf("UpdateSubscriptionStates() error for subscription %s: %v\n", sub.GetId(), err)
			}
		}
	}

	return nil
}

func InitUpdateSubscriptionStates() {
	go func() {
		if err := UpdateSubscriptionStates(); err != nil {
			fmt.Printf("Error updating subscription states at startup: %v\n", err)
		}
	}()

	cronJob := cron.New()
	_, err := cronJob.AddFunc("0 * * * *", func() {
		if err := UpdateSubscriptionStates(); err != nil {
			fmt.Printf("Error updating subscription states: %v\n", err)
		}
	})
	if err != nil {
		fmt.Printf("Error scheduling subscription state updates: %v\n", err)
		return
	}
	cronJob.Start()
}
