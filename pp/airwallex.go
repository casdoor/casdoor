package pp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/conf"
)

type AirwallexPaymentProvider struct {
	Client *AirwallexClient
}

func NewAirwallexPaymentProvider(clientId string, apiKey string) (*AirwallexPaymentProvider, error) {
	isProd := conf.GetConfigString("runmode") == "prod"
	apiEndpoint := "https://api-demo.airwallex.com/api/v1"
	apiCheckout := "https://checkout-demo.airwallex.com/#/standalone/checkout?"
	if isProd {
		apiEndpoint = "https://api.airwallex.com/api/v1"
		apiCheckout = "https://checkout.airwallex.com/#/standalone/checkout?"
	}
	client := &AirwallexClient{
		ClientId:    clientId,
		APIKey:      apiKey,
		APIEndpoint: apiEndpoint,
		APICheckout: apiCheckout,
		client:      &http.Client{Timeout: 15 * time.Second},
	}
	pp := &AirwallexPaymentProvider{
		Client: client,
	}
	return pp, nil
}

func (pp *AirwallexPaymentProvider) Pay(r *PayReq) (*PayResp, error) {
	// Create a payment intent
	intent, err := pp.Client.CreateIntent(r)
	if err != nil {
		return nil, err
	}
	payUrl, err := pp.Client.GetCheckoutUrl(intent, r)
	if err != nil {
		return nil, err
	}
	return &PayResp{
		PayUrl:  payUrl,
		OrderId: intent.MerchantOrderId,
	}, nil
}

func (pp *AirwallexPaymentProvider) Notify(body []byte, orderId string) (*NotifyResult, error) {
	notifyResult := &NotifyResult{}
	intent, err := pp.Client.GetIntentByOrderId(orderId)
	if err != nil {
		return nil, err
	}
	// Check intent status
	switch intent.Status {
	case "PENDING", "REQUIRES_PAYMENT_METHOD", "REQUIRES_CUSTOMER_ACTION", "REQUIRES_CAPTURE":
		notifyResult.PaymentStatus = PaymentStateCreated
		return notifyResult, nil
	case "CANCELLED":
		notifyResult.PaymentStatus = PaymentStateCanceled
		return notifyResult, nil
	case "EXPIRED":
		notifyResult.PaymentStatus = PaymentStateTimeout
		return notifyResult, nil
	case "SUCCEEDED":
		// Skip
	default:
		notifyResult.PaymentStatus = PaymentStateError
		notifyResult.NotifyMessage = fmt.Sprintf("unexpected airwallex checkout status: %v", intent.Status)
		return notifyResult, nil
	}
	// Check attempt status
	if intent.PaymentStatus != "" {
		switch intent.PaymentStatus {
		case "CANCELLED", "EXPIRED", "RECEIVED", "AUTHENTICATION_REDIRECTED", "AUTHORIZED", "CAPTURE_REQUESTED":
			notifyResult.PaymentStatus = PaymentStateCreated
			return notifyResult, nil
		case "PAID", "SETTLED":
			// Skip
		default:
			notifyResult.PaymentStatus = PaymentStateError
			notifyResult.NotifyMessage = fmt.Sprintf("unexpected airwallex checkout payment status: %v", intent.PaymentStatus)
			return notifyResult, nil
		}
	}
	// The Payment has succeeded.
	var productDisplayName, productName, providerName string
	if description, ok := intent.Metadata["description"]; ok {
		productName, productDisplayName, providerName, _ = parseAttachString(description.(string))
	}
	orderId = intent.MerchantOrderId
	return &NotifyResult{
		PaymentName:        orderId,
		PaymentStatus:      PaymentStatePaid,
		ProductName:        productName,
		ProductDisplayName: productDisplayName,
		ProviderName:       providerName,
		Price:              priceStringToFloat64(intent.Amount.String()),
		Currency:           intent.Currency,
		OrderId:            orderId,
	}, nil
}

func (pp *AirwallexPaymentProvider) GetInvoice(paymentName, personName, personIdCard, personEmail, personPhone, invoiceType, invoiceTitle, invoiceTaxId string) (string, error) {
	return "", nil
}

func (pp *AirwallexPaymentProvider) GetResponseError(err error) string {
	if err == nil {
		return "success"
	}
	return "fail"
}

/*
 * Airwallex Client implementation (to be removed upon official SDK release)
 */

type AirwallexClient struct {
	ClientId    string
	APIKey      string
	APIEndpoint string
	APICheckout string
	client      *http.Client
	tokenCache  *AirWallexTokenInfo
	tokenMutex  sync.RWMutex
}

type AirWallexTokenInfo struct {
	Token           string `json:"token"`
	ExpiresAt       string `json:"expires_at"`
	parsedExpiresAt time.Time
}

type AirWallexIntentResp struct {
	Id              string `json:"id"`
	ClientSecret    string `json:"client_secret"`
	MerchantOrderId string `json:"merchant_order_id"`
}

func (c *AirwallexClient) GetToken() (string, error) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	if c.tokenCache != nil && time.Now().Before(c.tokenCache.parsedExpiresAt) {
		return c.tokenCache.Token, nil
	}
	req, _ := http.NewRequest("POST", c.APIEndpoint+"/authentication/login", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("x-client-id", c.ClientId)
	req.Header.Set("x-api-key", c.APIKey)
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result AirWallexTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Token == "" {
		return "", fmt.Errorf("invalid token response")
	}
	expiresAt := strings.Replace(result.ExpiresAt, "+0000", "+00:00", 1)
	result.parsedExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
	c.tokenCache = &result
	return result.Token, nil
}

func (c *AirwallexClient) authRequest(method, url string, body interface{}) (map[string]interface{}, error) {
	token, err := c.GetToken()
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(body)
	var reqBody io.Reader
	if method != "GET" {
		reqBody = bytes.NewBuffer(b)
	}
	req, _ := http.NewRequest(method, url, reqBody)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *AirwallexClient) CreateIntent(r *PayReq) (*AirWallexIntentResp, error) {
	description := joinAttachString([]string{r.ProductName, r.ProductDisplayName, r.ProviderName})
	orderId := r.PaymentName
	intentReq := map[string]interface{}{
		"currency":          r.Currency,
		"amount":            r.Price,
		"merchant_order_id": orderId,
		"request_id":        orderId,
		"descriptor":        strings.ReplaceAll(string([]rune(description)[:32]), "\x00", ""),
		"metadata":          map[string]interface{}{"description": description},
		"order":             map[string]interface{}{"products": []map[string]interface{}{{"name": r.ProductDisplayName, "quantity": 1, "desc": r.ProductDescription, "image_url": r.ProductImage}}},
		"customer":          map[string]interface{}{"merchant_customer_id": r.PayerId, "email": r.PayerEmail, "first_name": r.PayerName, "last_name": r.PayerName},
	}
	intentUrl := fmt.Sprintf("%s/pa/payment_intents/create", c.APIEndpoint)
	intentRes, err := c.authRequest("POST", intentUrl, intentReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %v", err)
	}
	return &AirWallexIntentResp{
		Id:              intentRes["id"].(string),
		ClientSecret:    intentRes["client_secret"].(string),
		MerchantOrderId: intentRes["merchant_order_id"].(string),
	}, nil
}

type AirwallexIntent struct {
	Amount               json.Number `json:"amount"`
	Currency             string      `json:"currency"`
	Id                   string      `json:"id"`
	Status               string      `json:"status"`
	Descriptor           string      `json:"descriptor"`
	MerchantOrderId      string      `json:"merchant_order_id"`
	LatestPaymentAttempt struct {
		Status string `json:"status"`
	} `json:"latest_payment_attempt"`
	Metadata map[string]interface{} `json:"metadata"`
}

type AirwallexIntents struct {
	Items []AirwallexIntent `json:"items"`
}

type AirWallexIntentInfo struct {
	Amount          json.Number
	Currency        string
	Id              string
	Status          string
	Descriptor      string
	MerchantOrderId string
	PaymentStatus   string
	Metadata        map[string]interface{}
}

func (c *AirwallexClient) GetIntentByOrderId(orderId string) (*AirWallexIntentInfo, error) {
	intentUrl := fmt.Sprintf("%s/pa/payment_intents/?merchant_order_id=%s", c.APIEndpoint, orderId)
	intentRes, err := c.authRequest("GET", intentUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment intent: %v", err)
	}
	items := intentRes["items"].([]interface{})
	if len(items) == 0 {
		return nil, fmt.Errorf("no payment intent found for order id: %s", orderId)
	}
	var intent AirwallexIntent
	if b, err := json.Marshal(items[0]); err == nil {
		json.Unmarshal(b, &intent)
	}
	return &AirWallexIntentInfo{
		Id:              intent.Id,
		Amount:          intent.Amount,
		Currency:        intent.Currency,
		Status:          intent.Status,
		Descriptor:      intent.Descriptor,
		MerchantOrderId: intent.MerchantOrderId,
		PaymentStatus:   intent.LatestPaymentAttempt.Status,
		Metadata:        intent.Metadata,
	}, nil
}

func (c *AirwallexClient) GetCheckoutUrl(intent *AirWallexIntentResp, r *PayReq) (string, error) {
	return fmt.Sprintf("%sintent_id=%s&client_secret=%s&mode=payment&currency=%s&amount=%v&requiredBillingContactFields=%s&successUrl=%s&failUrl=%s&logoUrl=%s",
		c.APICheckout,
		intent.Id,
		intent.ClientSecret,
		r.Currency,
		r.Price,
		url.QueryEscape(`["address"]`),
		r.ReturnUrl,
		r.ReturnUrl,
		"data:image/gif;base64,R0lGODlhAQABAAD/ACwAAAAAAQABAAACADs=", // replace default logo
	), nil
}
