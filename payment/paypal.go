package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/casbin/casdoor/object"
	"github.com/plutov/paypal/v4"
)

var client = GetClient()

func GetClient() *paypal.Client {
	c, err := paypal.NewClient(beego.AppConfig.String("paypalClientId"), beego.AppConfig.String("paypalSecret"), paypal.APIBaseSandBox)
	if err != nil {
		panic(err)
	}
	return c
}

func Paypal(payItem object.PayItem, clientId string, redirectUri string) string {

	application := object.GetApplicationByClientId(clientId)
	if application == nil {
		return "Invalid client_id"
	}
	applicationName := fmt.Sprintf("%s/%s", application.Owner, application.Name)
	if payItem.Currency == "" {
		payItem.Currency = "USD"
	}

	_, err := client.GetAccessToken(context.Background())
	if err != nil {
		panic(err)
	}
	appContext := &paypal.ApplicationContext{
		ReturnURL: "http://localhost:7001/pay/success", //回调链接
		CancelURL: "https://www.baidu.com",
	}

	purchaseUnits := make([]paypal.PurchaseUnitRequest, 1)
	purchaseUnits[0] = paypal.PurchaseUnitRequest{
		Amount: &paypal.PurchaseUnitAmount{
			Currency: payItem.Currency, //收款类型
			Value:    payItem.Price,    //收款数量
		},
		InvoiceID:   payItem.Invoice,
		Description: payItem.Description,
	}

	order, err := client.CreateOrder(context.Background(),
		paypal.OrderIntentCapture,
		purchaseUnits,
		&paypal.CreateOrderPayer{},
		appContext)
	if err != nil {
		panic(err)
	}

	newPay := object.Payment{
		Id:          order.ID,
		Invoice:     payItem.Invoice,
		PayItem:     payItem,
		Application: applicationName,
		Status:      order.Status,
		Callback:    redirectUri,
	}

	success := object.AddPayment(&newPay)
	if success {
		links := order.Links
		for _, link := range links {
			fmt.Println(link.Rel)
			if link.Rel == "approve" {
				return link.Href
			}
		}
	}

	return "Add Order to Database false"
}

func SuccessPay(token string) string {
	_, err := client.GetAccessToken(context.Background())
	if err != nil {
		panic(err)
	}
	captureOrder, err := client.CaptureOrder(context.Background(), token, paypal.CaptureOrderRequest{})
	if err != nil {
		panic(err)
	}
	pay := object.GetPayment(captureOrder.ID)
	pay.Purchase = captureOrder.PurchaseUnits
	pay.Payer = captureOrder.Payer
	pay.UpdateTime = time.Now().String()
	pay.Status = captureOrder.Status
	object.UpdatePay(captureOrder.ID, pay)
	if captureOrder.Status == "COMPLETED" {
		return fmt.Sprintf("%s?paymentId=%s", pay.Callback, token)
	}
	return ""

}
