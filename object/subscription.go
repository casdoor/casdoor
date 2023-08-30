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
	"time"

	"github.com/casdoor/casdoor/pp"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type SubscriptionState string

const (
	SubStatePending   SubscriptionState = "Pending"
	SubStateError     SubscriptionState = "Error"
	SubStateSuspended SubscriptionState = "Suspended" // suspended by the admin

	SubStateActive   SubscriptionState = "Active"
	SubStateUpcoming SubscriptionState = "Upcoming"
	SubStateExpired  SubscriptionState = "Expired"
)

type Subscription struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	Description string `xorm:"varchar(100)" json:"description"`

	User    string `xorm:"varchar(100)" json:"user"`
	Pricing string `xorm:"varchar(100)" json:"pricing"`
	Plan    string `xorm:"varchar(100)" json:"plan"`
	Payment string `xorm:"varchar(100)" json:"payment"`

	StartTime time.Time         `json:"startTime"`
	EndTime   time.Time         `json:"endTime"`
	Period    string            `xorm:"varchar(100)" json:"period"`
	State     SubscriptionState `xorm:"varchar(100)" json:"state"`
}

func (sub *Subscription) GetId() string {
	return fmt.Sprintf("%s/%s", sub.Owner, sub.Name)
}

func (sub *Subscription) UpdateState() error {
	preState := sub.State
	// update subscription state by payment state
	if sub.State == SubStatePending {
		if sub.Payment == "" {
			return nil
		}
		payment, err := GetPayment(util.GetId(sub.Owner, sub.Payment))
		if err != nil {
			return err
		}
		if payment == nil {
			sub.Description = fmt.Sprintf("payment: %s does not exist", sub.Payment)
			sub.State = SubStateError
		} else {
			if payment.State == pp.PaymentStatePaid {
				sub.State = SubStateActive
			} else if payment.State != pp.PaymentStateCreated {
				// other states: Canceled, Timeout, Error
				sub.Description = fmt.Sprintf("payment: %s state is %v", sub.Payment, payment.State)
				sub.State = SubStateError
			}
		}
	}

	if sub.State == SubStateActive || sub.State == SubStateUpcoming || sub.State == SubStateExpired {
		if sub.EndTime.Before(time.Now()) {
			sub.State = SubStateExpired
		} else if sub.StartTime.After(time.Now()) {
			sub.State = SubStateUpcoming
		} else {
			sub.State = SubStateActive
		}
	}

	if preState != sub.State {
		_, err := UpdateSubscription(sub.GetId(), sub)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewSubscription(owner, userName, planName, paymentName, period string) *Subscription {
	startTime, endTime := GetDuration(period)
	id := util.GenerateId()[:6]
	return &Subscription{
		Owner:       owner,
		Name:        "sub_" + id,
		DisplayName: "New Subscription - " + id,
		CreatedTime: util.GetCurrentTime(),

		User:    userName,
		Plan:    planName,
		Payment: paymentName,

		StartTime: startTime,
		EndTime:   endTime,
		Period:    period,
		State:     SubStatePending, // waiting for payment complete
	}
}

func GetSubscriptionCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Subscription{})
}

func GetSubscriptions(owner string) ([]*Subscription, error) {
	subscriptions := []*Subscription{}
	err := ormer.Engine.Desc("created_time").Find(&subscriptions, &Subscription{Owner: owner})
	if err != nil {
		return subscriptions, err
	}
	for _, sub := range subscriptions {
		err = sub.UpdateState()
		if err != nil {
			return nil, err
		}
	}
	return subscriptions, nil
}

func GetSubscriptionsByUser(owner, userName string) ([]*Subscription, error) {
	subscriptions := []*Subscription{}
	err := ormer.Engine.Desc("created_time").Find(&subscriptions, &Subscription{Owner: owner, User: userName})
	if err != nil {
		return subscriptions, err
	}
	// update subscription state
	for _, sub := range subscriptions {
		err = sub.UpdateState()
		if err != nil {
			return subscriptions, err
		}
	}
	return subscriptions, nil
}

func GetPaginationSubscriptions(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Subscription, error) {
	subscriptions := []*Subscription{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&subscriptions)
	if err != nil {
		return subscriptions, err
	}
	for _, sub := range subscriptions {
		err = sub.UpdateState()
		if err != nil {
			return nil, err
		}
	}
	return subscriptions, nil
}

func getSubscription(owner string, name string) (*Subscription, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	subscription := Subscription{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&subscription)
	if err != nil {
		return nil, err
	}

	if existed {
		return &subscription, nil
	} else {
		return nil, nil
	}
}

func GetSubscription(id string) (*Subscription, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getSubscription(owner, name)
}

func UpdateSubscription(id string, subscription *Subscription) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if s, err := getSubscription(owner, name); err != nil {
		return false, err
	} else if s == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(subscription)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddSubscription(subscription *Subscription) (bool, error) {
	affected, err := ormer.Engine.Insert(subscription)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteSubscription(subscription *Subscription) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{subscription.Owner, subscription.Name}).Delete(&Subscription{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}
