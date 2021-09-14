package controllers

import (
	"encoding/json"

	"github.com/casbin/casdoor/object"
	"github.com/casbin/casdoor/payment"
)

func (c *ApiController) PaypalPay() {
	clientId := c.Input().Get("clientId")
	redirectUri := c.Input().Get("redirectUri")
	var payItem object.PayItem
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &payItem)
	if err != nil {
		panic(err)
	}

	msg := payment.Paypal(payItem, clientId, redirectUri)

	c.Data["json"] = msg
	c.ServeJSON()
}

func (c *ApiController) GetPayments() {
	c.Data["json"] = object.GetPayments()
	c.ServeJSON()
}

func (c *ApiController) DeletePayment() {
	var payment object.Payment
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &payment)
	if err != nil {
		panic(err)
	}

	c.Data["json"] = wrapActionResponse(object.DeletePayment(&payment))
	c.ServeJSON()
}

func (c *ApiController) SuccessPay() {
	token := c.Input().Get("paymentId")
	c.Data["json"] = payment.SuccessPay(token)
	c.ServeJSON()
}
