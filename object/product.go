// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
	"xorm.io/core"
)

type Product struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Image     string   `xorm:"varchar(100)" json:"image"`
	Detail    string   `xorm:"varchar(100)" json:"detail"`
	Tag       string   `xorm:"varchar(100)" json:"tag"`
	Currency  string   `xorm:"varchar(100)" json:"currency"`
	Price     float64  `json:"price"`
	Quantity  int      `json:"quantity"`
	Sold      int      `json:"sold"`
	Providers []string `xorm:"varchar(100)" json:"providers"`

	State string `xorm:"varchar(100)" json:"state"`
}

func GetProductCount(owner, field, value string) int {
	session := GetSession(owner, -1, -1, field, value, "", "")
	count, err := session.Count(&Product{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetProducts(owner string) []*Product {
	products := []*Product{}
	err := adapter.Engine.Desc("created_time").Find(&products, &Product{Owner: owner})
	if err != nil {
		panic(err)
	}

	return products
}

func GetPaginationProducts(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Product {
	products := []*Product{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&products)
	if err != nil {
		panic(err)
	}

	return products
}

func getProduct(owner string, name string) *Product {
	if owner == "" || name == "" {
		return nil
	}

	product := Product{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&product)
	if err != nil {
		panic(err)
	}

	if existed {
		return &product
	} else {
		return nil
	}
}

func GetProduct(id string) *Product {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getProduct(owner, name)
}

func UpdateProduct(id string, product *Product) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getProduct(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(product)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddProduct(product *Product) bool {
	affected, err := adapter.Engine.Insert(product)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteProduct(product *Product) bool {
	affected, err := adapter.Engine.ID(core.PK{product.Owner, product.Name}).Delete(&Product{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (product *Product) GetId() string {
	return fmt.Sprintf("%s/%s", product.Owner, product.Name)
}

func (product *Product) isValidProvider(provider *Provider) bool {
	for _, providerName := range product.Providers {
		if providerName == provider.Name {
			return true
		}
	}
	return false
}

func (product *Product) getProvider(providerId string) (*Provider, error) {
	provider := getProvider(product.Owner, providerId)
	if provider == nil {
		return nil, fmt.Errorf("the payment provider: %s does not exist", providerId)
	}

	if !product.isValidProvider(provider) {
		return nil, fmt.Errorf("the payment provider: %s is not valid for the product: %s", providerId, product.Name)
	}

	return provider, nil
}

func BuyProduct(id string, providerId string, host string) (string, error) {
	product := GetProduct(id)
	if product == nil {
		return "", fmt.Errorf("the product: %s does not exist", id)
	}

	provider, err := product.getProvider(providerId)
	if err != nil {
		return "", err
	}

	cert := getCert(product.Owner, provider.Cert)
	if cert == nil {
		return "", fmt.Errorf("the cert: %s does not exist", provider.Cert)
	}

	pProvider := pp.GetPaymentProvider(provider.Type, provider.ClientId, cert.PublicKey, cert.PrivateKey, cert.AuthorityPublicKey, cert.AuthorityRootPublicKey)
	if pProvider == nil {
		return "", fmt.Errorf("the payment provider type: %s is not supported", provider.Type)
	}

	paymentId := util.GenerateTimeId()

	originFrontend, originBackend := getOriginFromHost(host)
	returnUrl := fmt.Sprintf("%s/payments/%s", originFrontend, paymentId)
	notifyUrl := fmt.Sprintf("%s/api/notify-payment", originBackend)

	payUrl, err := pProvider.Pay(product.DisplayName, product.Name, provider.Name, paymentId, product.Price, returnUrl, notifyUrl)
	return payUrl, err
}
