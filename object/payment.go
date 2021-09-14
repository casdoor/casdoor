package object

import (
	"github.com/plutov/paypal/v4"
	"xorm.io/core"
)

type Payment struct {
	Id          string                        `xorm:"varchar(100) notnull pk" json:"id"`
	Invoice     string                        `xorm:"varchar(100)" json:"invoice"`
	Application string                        `xorm:"varchar(100)" json:"application"`
	PayItem     PayItem                       `xorm:"json varchar(1000)" json:"pay_item"`
	Payer       *paypal.PayerWithNameAndPhone `xorm:"json varchar(1000)" json:"payer"`
	Purchase    []paypal.CapturedPurchaseUnit `xorm:"varchar(10000)" json:"purchase"`
	Status      string                        `xorm:"varchar(100)" json:"status"`
	CreateTime  string                        `xorm:"varchar(100) created" json:"create_time"`
	UpdateTime  string                        `xorm:"varchar(100) updated" json:"update_time"`
	Callback    string                        `xorm:"varchar(1000)" json:"callback"`
}

func AddPayment(pay *Payment) bool {
	affected, err := adapter.Engine.Insert(pay)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetPayments() []*Payment {
	pays := []*Payment{}
	err := adapter.Engine.Desc("create_time").Find(&pays)
	if err != nil {
		panic(err)
	}

	return pays
}

func GetPayment(id string) *Payment {
	pay := Payment{Id: id}
	existed, err := adapter.Engine.Get(&pay)
	if err != nil {
		panic(err)
	}

	if existed {
		return &pay
	} else {
		return nil
	}
}

func DeletePayment(payment *Payment) bool {
	affected, err := adapter.Engine.ID(core.PK{payment.Id}).Delete(&Payment{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func UpdatePay(id string, pay *Payment) bool {
	if GetPayment(id) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{id}).AllCols().Update(pay)
	if err != nil {
		panic(err)
	}

	return affected != 0
}
