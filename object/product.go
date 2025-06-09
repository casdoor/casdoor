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

	"github.com/casdoor/casdoor/idp"

	"github.com/casdoor/casdoor/pp"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Product struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	DisplayName string `xorm:"varchar(100)" json:"displayName"`

	Image       string   `xorm:"varchar(100)" json:"image"`
	Detail      string   `xorm:"varchar(1000)" json:"detail"`
	Description string   `xorm:"varchar(200)" json:"description"`
	Tag         string   `xorm:"varchar(100)" json:"tag"`
	Currency    string   `xorm:"varchar(100)" json:"currency"`
	Price       float64  `json:"price"`
	Quantity    int      `json:"quantity"`
	Sold        int      `json:"sold"`
	IsRecharge  bool     `json:"isRecharge"`
	Providers   []string `xorm:"varchar(255)" json:"providers"`
	ReturnUrl   string   `xorm:"varchar(1000)" json:"returnUrl"`
	ReturnType  string   `xorm:"varchar(100)" json:"returnType"`

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
	owner, name := util.GetOwnerAndNameFromId(id)
	return getProduct(owner, name)
}

func UpdateProduct(id string, product *Product) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if p, err := getProduct(owner, name); err != nil {
		return false, err
	} else if p == nil {
		return false, nil
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(product)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddProduct(product *Product) (bool, error) {
	affected, err := ormer.Engine.Insert(product)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
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

func (product *Product) isValidProvider(provider *Provider) bool {
	for _, providerName := range product.Providers {
		if providerName == provider.Name {
			return true
		}
	}
	return false
}

func (product *Product) getProvider(providerName string) (*Provider, error) {
	provider, err := getProvider(product.Owner, providerName)
	if err != nil {
		return nil, err
	}

	if provider == nil {
		return nil, fmt.Errorf("the payment provider: %s does not exist", providerName)
	}

	if !product.isValidProvider(provider) {
		return nil, fmt.Errorf("the payment provider: %s is not valid for the product: %s", providerName, product.Name)
	}

	return provider, nil
}

func BuyProduct(id string, user *User, providerName, pricingName, planName, host, paymentEnv string, customPrice float64, remark string) (payment *Payment, attachInfo map[string]interface{}, err error) {
	product, err := GetProduct(id)
	if err != nil {
		return nil, nil, err
	}
	if product == nil {
		return nil, nil, fmt.Errorf("the product: %s does not exist", id)
	}

	if product.IsRecharge {
		if customPrice <= 0 {
			return nil, nil, fmt.Errorf("the custom price should bigger than zero")
		} else {
			product.Price = customPrice
		}
	}

	provider, err := product.getProvider(providerName)
	if err != nil {
		return nil, nil, err
	}

	pProvider, err := GetPaymentProvider(provider)
	if err != nil {
		return nil, nil, err
	}

	owner := product.Owner
	payerName := fmt.Sprintf("%s | %s", user.Name, user.DisplayName)
	paymentName := fmt.Sprintf("payment_%v", util.GenerateTimeId())

	originFrontend, originBackend := getOriginFromHost(host)

	returnUrl := fmt.Sprintf("%s/payments/%s/%s/result?source=pay", originFrontend, owner, paymentName)
	if product.ReturnType == "directRedirect" {
		returnUrl = product.ReturnUrl
	}

	notifyUrl := fmt.Sprintf("%s/api/notify-payment/%s/%s", originBackend, owner, paymentName)
	if user.Type == "paid-user" {
		// Create a subscription for `paid-user`
		if pricingName != "" && planName != "" {
			plan, err := GetPlan(util.GetId(owner, planName))
			if err != nil {
				return nil, nil, err
			}
			if plan == nil {
				return nil, nil, fmt.Errorf("the plan: %s does not exist", planName)
			}
			sub := NewSubscription(owner, user.Name, plan.Name, paymentName, plan.Period)
			_, err = AddSubscription(sub)
			if err != nil {
				return nil, nil, err
			}
			returnUrl = fmt.Sprintf("%s/buy-plan/%s/%s/result?subscription=%s", originFrontend, owner, pricingName, sub.Name)
		}
	}
	// Create an order
	payReq := &pp.PayReq{
		ProviderName:       providerName,
		ProductName:        product.Name,
		PayerName:          payerName,
		PayerId:            user.Id,
		PayerEmail:         user.Email,
		PaymentName:        paymentName,
		ProductDisplayName: product.DisplayName,
		ProductDescription: product.Description,
		ProductImage:       product.Image,
		Price:              product.Price,
		Currency:           product.Currency,
		ReturnUrl:          returnUrl,
		NotifyUrl:          notifyUrl,
		PaymentEnv:         paymentEnv,
	}

	// custom process for WeChat & WeChat Pay
	if provider.Type == "WeChat Pay" {
		payReq.PayerId, err = getUserExtraProperty(user, "WeChat", idp.BuildWechatOpenIdKey(provider.ClientId2))
		if err != nil {
			return nil, nil, err
		}
	} else if provider.Type == "Balance" {
		payReq.PayerId = user.GetId()
	}

	payResp, err := pProvider.Pay(payReq)
	if err != nil {
		return nil, nil, err
	}
	// Create a Payment linked with Product and Order
	payment = &Payment{
		Owner:       product.Owner,
		Name:        paymentName,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: paymentName,

		Provider: provider.Name,
		Type:     provider.Type,

		ProductName:        product.Name,
		ProductDisplayName: product.DisplayName,
		Detail:             product.Detail,
		Tag:                product.Tag,
		Currency:           product.Currency,
		Price:              product.Price,
		ReturnUrl:          product.ReturnUrl,
		ReturnType:         product.ReturnType,
		IsRecharge:         product.IsRecharge,

		User:       user.Name,
		PayUrl:     payResp.PayUrl,
		SuccessUrl: returnUrl,
		State:      pp.PaymentStateCreated,
		OutOrderId: payResp.OrderId,
		Remark:     remark,
	}

	transaction := &Transaction{
		Owner:       payment.Owner,
		Name:        payment.Name,
		DisplayName: payment.DisplayName,
		Provider:    provider.Name,
		Category:    provider.Category,
		Type:        provider.Type,

		ProductName:        product.Name,
		ProductDisplayName: product.DisplayName,
		Detail:             product.Detail,
		Tag:                product.Tag,
		Currency:           product.Currency,
		Amount:             payment.Price,
		ReturnUrl:          payment.ReturnUrl,

		User:        payment.User,
		Application: owner,
		Payment:     payment.GetId(),

		State: pp.PaymentStateCreated,
	}

	if provider.Type == "Dummy" {
		payment.State = pp.PaymentStatePaid
		err = UpdateUserBalance(user.Owner, user.Name, payment.Price)
		if err != nil {
			return nil, nil, err
		}
	} else if provider.Type == "Balance" {
		if product.Price > user.Balance {
			return nil, nil, fmt.Errorf("insufficient user balance")
		}
		transaction.Amount = -transaction.Amount
		err = UpdateUserBalance(user.Owner, user.Name, -product.Price)
		if err != nil {
			return nil, nil, err
		}
		payment.State = pp.PaymentStatePaid
		transaction.State = pp.PaymentStatePaid
	}

	affected, err := AddPayment(payment)
	if err != nil {
		return nil, nil, err
	}

	if !affected {
		return nil, nil, fmt.Errorf("failed to add payment: %s", util.StructToJson(payment))
	}

	if product.IsRecharge || provider.Type == "Balance" {
		affected, err = AddTransaction(transaction)
		if err != nil {
			return nil, nil, err
		}
		if !affected {
			return nil, nil, fmt.Errorf("failed to add transaction: %s", util.StructToJson(payment))
		}
	}

	return payment, payResp.AttachInfo, nil
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

type BuyProductBody struct {
	Remark string `json:"remark"`
}
