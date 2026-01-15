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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Cart struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	User        string `xorm:"varchar(100)" json:"user"`
	ProductName string `xorm:"varchar(100)" json:"productName"`
	Quantity    int    `json:"quantity"`

	ProductObj *Product `xorm:"-" json:"productObj"`
}

func GetCartCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Cart{})
}

func GetCarts(owner string) ([]*Cart, error) {
	carts := []*Cart{}
	err := ormer.Engine.Desc("created_time").Find(&carts, &Cart{Owner: owner})
	if err != nil {
		return carts, err
	}

	return carts, nil
}

func GetUserCarts(owner, user string) ([]*Cart, error) {
	carts := []*Cart{}
	err := ormer.Engine.Desc("created_time").Find(&carts, &Cart{Owner: owner, User: user})
	if err != nil {
		return carts, err
	}

	return carts, nil
}

func GetPaginationCarts(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Cart, error) {
	carts := []*Cart{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&carts)
	if err != nil {
		return carts, err
	}

	return carts, nil
}

func getCart(owner string, name string) (*Cart, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	cart := Cart{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&cart)
	if err != nil {
		return &cart, err
	}

	if existed {
		return &cart, nil
	} else {
		return nil, nil
	}
}

func GetCart(id string) (*Cart, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getCart(owner, name)
}

func UpdateCart(id string, cart *Cart) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if c, err := getCart(owner, name); err != nil {
		return false, err
	} else if c == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(cart)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddCart(cart *Cart) (bool, error) {
	affected, err := ormer.Engine.Insert(cart)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteCart(cart *Cart) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{cart.Owner, cart.Name}).Delete(&Cart{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (cart *Cart) GetId() string {
	return fmt.Sprintf("%s/%s", cart.Owner, cart.Name)
}

func ExtendCartWithProduct(cart *Cart) error {
	if cart == nil {
		return nil
	}

	if cart.ProductName != "" {
		product, err := getProduct(cart.Owner, cart.ProductName)
		if err != nil {
			return err
		}
		cart.ProductObj = product
	}

	return nil
}
