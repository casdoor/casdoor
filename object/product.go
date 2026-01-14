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

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Product struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Image                 string    `xorm:"varchar(100)" json:"image"`
	Detail                string    `xorm:"varchar(1000)" json:"detail"`
	Description           string    `xorm:"varchar(200)" json:"description"`
	Tag                   string    `xorm:"varchar(100)" json:"tag"`
	Currency              string    `xorm:"varchar(100)" json:"currency"`
	Price                 float64   `json:"price"`
	Quantity              int       `json:"quantity"`
	Sold                  int       `json:"sold"`
	IsRecharge            bool      `json:"isRecharge"`
	RechargeOptions       []float64 `xorm:"varchar(500)" json:"rechargeOptions"`
	DisableCustomRecharge bool      `json:"disableCustomRecharge"`
	Providers             []string  `xorm:"varchar(255)" json:"providers"`
	SuccessUrl            string    `xorm:"varchar(1000)" json:"successUrl"`

	State string `xorm:"varchar(100)" json:"state"`

	ProviderObjs []*Provider `xorm:"-" json:"providerObjs"`
}

func GetProductCount(owner, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Product{})
}

func GetProducts(owner string) ([]*Product, error) {
	products := []*Product{}
	err := ormer.Engine.Desc("created_time").Find(&products, &Product{Owner: owner})
	if err != nil {
		return products, err
	}

	return products, nil
}

func GetPaginationProducts(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Product, error) {
	products := []*Product{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&products)
	if err != nil {
		return products, err
	}

	return products, nil
}

func getProduct(owner string, name string) (*Product, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	product := Product{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&product)
	if err != nil {
		return &product, nil
	}

	if existed {
		return &product, nil
	} else {
		return nil, nil
	}
}

func GetProduct(id string) (*Product, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getProduct(owner, name)
}

func UpdateProductStock(products []Product) error {
	var (
		affected int64
		err      error
	)
	for _, product := range products {
		if product.IsRecharge {
			affected, err = ormer.Engine.ID(core.PK{product.Owner, product.Name}).
				Incr("sold", 1).
				Update(&Product{})
		} else {
			affected, err = ormer.Engine.ID(core.PK{product.Owner, product.Name}).
				Where("quantity > 0").
				Decr("quantity", 1).
				Incr("sold", 1).
				Update(&Product{})
		}

		if err != nil {
			return err
		}
		if affected == 0 {
			if product.IsRecharge {
				return fmt.Errorf("failed to update stock for product: %s", product.Name)
			}
			return fmt.Errorf("insufficient stock for product: %s", product.Name)
		}
	}
	return nil
}

func UpdateProduct(id string, product *Product) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if p, err := getProduct(owner, name); err != nil {
		return false, err
	} else if p == nil {
		return false, nil
	}

	err = checkProduct(product)
	if err != nil {
		return false, err
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(product)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddProduct(product *Product) (bool, error) {
	err := checkProduct(product)
	if err != nil {
		return false, err
	}

	affected, err := ormer.Engine.Insert(product)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func checkProduct(product *Product) error {
	if product == nil {
		return fmt.Errorf("the product not exist")
	}

	for _, providerName := range product.Providers {
		provider, err := getProvider(product.Owner, providerName)
		if err != nil {
			return err
		}
		if provider != nil && provider.Type == "Alipay" && product.Currency != "CNY" {
			return fmt.Errorf("alipay provider only supports CNY, got: %s", product.Currency)
		}
	}
	return nil
}

func DeleteProduct(product *Product) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{product.Owner, product.Name}).Delete(&Product{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (product *Product) GetId() string {
	return fmt.Sprintf("%s/%s", product.Owner, product.Name)
}

func (product *Product) isValidProvider(provider *Provider) error {
	if provider.Type == "Alipay" && product.Currency != "CNY" {
		return fmt.Errorf("alipay provider only supports CNY, got: %s", product.Currency)
	}

	providerMatched := false
	for _, providerName := range product.Providers {
		if providerName == provider.Name {
			providerMatched = true
			break
		}
	}
	if !providerMatched {
		return fmt.Errorf("the payment provider: %s is not valid for the product: %s", provider.Name, product.Name)
	}

	return nil
}

func (product *Product) getProvider(providerName string) (*Provider, error) {
	provider, err := getProvider(product.Owner, providerName)
	if err != nil {
		return nil, err
	}

	if provider == nil {
		return nil, fmt.Errorf("the payment provider: %s does not exist", providerName)
	}

	if err := product.isValidProvider(provider); err != nil {
		return nil, err
	}

	return provider, nil
}

func ExtendProductWithProviders(product *Product) error {
	if product == nil {
		return nil
	}

	product.ProviderObjs = []*Provider{}

	m, err := getProviderMap(product.Owner)
	if err != nil {
		return err
	}

	for _, providerItem := range product.Providers {
		if provider, ok := m[providerItem]; ok {
			product.ProviderObjs = append(product.ProviderObjs, provider)
		}
	}

	return nil
}

func CreateProductForPlan(plan *Plan) *Product {
	product := &Product{
		Owner:       plan.Owner,
		Name:        fmt.Sprintf("product_%v", util.GetRandomName()),
		DisplayName: fmt.Sprintf("Product for Plan %v/%v/%v", plan.Name, plan.DisplayName, plan.Period),
		CreatedTime: plan.CreatedTime,

		Image:       "https://cdn.casbin.org/img/casdoor-logo_1185x256.png", // TODO
		Detail:      fmt.Sprintf("This product was auto created for plan %v(%v), subscription period is %v", plan.Name, plan.DisplayName, plan.Period),
		Description: plan.Description,
		Tag:         "auto_created_product_for_plan",
		Price:       plan.Price,
		Currency:    plan.Currency,

		Quantity:   999,
		Sold:       0,
		IsRecharge: false,

		Providers: plan.PaymentProviders,
		State:     "Published",
	}
	if product.Providers == nil {
		product.Providers = []string{}
	}
	return product
}

func UpdateProductForPlan(plan *Plan, product *Product) {
	product.Owner = plan.Owner
	product.DisplayName = fmt.Sprintf("Product for Plan %v/%v/%v", plan.Name, plan.DisplayName, plan.Period)
	product.Detail = fmt.Sprintf("This product was auto created for plan %v(%v), subscription period is %v", plan.Name, plan.DisplayName, plan.Period)
	product.Price = plan.Price
	product.Currency = plan.Currency
	product.Providers = plan.PaymentProviders
}

func getOrderProducts(owner string, productNames []string) ([]Product, error) {
	if len(productNames) == 0 {
		return []Product{}, nil
	}

	names := make([]string, 0, len(productNames))
	for _, productName := range productNames {
		names = append(names, productName)
	}

	var products []Product
	err := ormer.Engine.
		Where("owner = ?", owner).
		In("name", names).
		Find(&products)
	if err != nil {
		return nil, err
	}
	return products, nil
}
