// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

type Order struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	// Product Info
	ProductName string   `xorm:"varchar(100)" json:"productName"`
	Products    []string `xorm:"varchar(1000)" json:"products"` // Future support for multiple products per order. Using varchar(1000) for simple JSON array storage; can be refactored to separate table if needed

	// User Info
	User string `xorm:"varchar(100)" json:"user"`

	// Payment Info
	Payment string `xorm:"varchar(100)" json:"payment"`

	// Order State
	State   string `xorm:"varchar(100)" json:"state"`
	Message string `xorm:"varchar(2000)" json:"message"`

	// Order Duration
	StartTime string `xorm:"varchar(100)" json:"startTime"`
	EndTime   string `xorm:"varchar(100)" json:"endTime"`
}

func GetOrderCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Order{Owner: owner})
}

func GetOrders(owner string) ([]*Order, error) {
	orders := []*Order{}
	err := ormer.Engine.Desc("created_time").Find(&orders, &Order{Owner: owner})
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func GetUserOrders(owner, user string) ([]*Order, error) {
	orders := []*Order{}
	err := ormer.Engine.Desc("created_time").Find(&orders, &Order{Owner: owner, User: user})
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func GetPaginationOrders(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Order, error) {
	orders := []*Order{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&orders, &Order{Owner: owner})
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func getOrder(owner string, name string) (*Order, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	order := Order{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&order)
	if err != nil {
		return nil, err
	}

	if existed {
		return &order, nil
	} else {
		return nil, nil
	}
}

func GetOrder(id string) (*Order, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getOrder(owner, name)
}

func UpdateOrder(id string, order *Order) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if o, err := getOrder(owner, name); err != nil {
		return false, err
	} else if o == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(order)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddOrder(order *Order) (bool, error) {
	affected, err := ormer.Engine.Insert(order)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteOrder(order *Order) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{order.Owner, order.Name}).Delete(&Order{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (order *Order) GetId() string {
	return fmt.Sprintf("%s/%s", order.Owner, order.Name)
}

func PlaceOrder(productId string, user *User) (*Order, error) {
	product, err := GetProduct(productId)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, fmt.Errorf("the product: %s does not exist", productId)
	}

	orderName := fmt.Sprintf("order_%v", util.GenerateTimeId())
	order := &Order{
		Owner:       product.Owner,
		Name:        orderName,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: fmt.Sprintf("Order for %s", product.DisplayName),
		ProductName: product.Name,
		Products:    []string{product.Name},
		User:        user.Name,
		Payment:     "", // Payment will be set when user pays
		State:       "Created",
		Message:     "",
		StartTime:   util.GetCurrentTime(),
		EndTime:     "",
	}

	affected, err := AddOrder(order)
	if err != nil {
		return nil, err
	}
	if !affected {
		return nil, fmt.Errorf("failed to add order: %s", util.StructToJson(order))
	}

	return order, nil
}

func GetOrderByPayment(owner string, paymentName string) (*Order, error) {
	order := &Order{Owner: owner, Payment: paymentName}
	existed, err := ormer.Engine.Get(order)
	if err != nil {
		return nil, err
	}
	if existed {
		return order, nil
	}
	return nil, nil
}

func CancelOrder(orderId string) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(orderId)
	if err != nil {
		return false, err
	}

	order, err := getOrder(owner, name)
	if err != nil {
		return false, err
	}
	if order == nil {
		return false, fmt.Errorf("the order: %s does not exist", orderId)
	}

	// Only allow cancellation of unpaid orders
	if order.State != "Created" {
		return false, fmt.Errorf("cannot cancel order in state: %s", order.State)
	}

	order.State = "Canceled"
	order.Message = "Canceled by user"
	return UpdateOrder(orderId, order)
}
