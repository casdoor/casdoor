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

package email

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	importanceNormal  = "normal"
	sendEmailEndpoint = "/emails:send"
	apiVersion        = "2023-03-31"
)

type Email struct {
	Recipients    Recipients     `json:"recipients"`
	SenderAddress string         `json:"senderAddress"`
	Content       Content        `json:"content"`
	Headers       []CustomHeader `json:"headers"`
	Tracking      bool           `json:"disableUserEngagementTracking"`
	Importance    string         `json:"importance"`
	ReplyTo       []EmailAddress `json:"replyTo"`
	Attachments   []Attachment   `json:"attachments"`
}

type Recipients struct {
	To  []EmailAddress `json:"to"`
	CC  []EmailAddress `json:"cc"`
	BCC []EmailAddress `json:"bcc"`
}

type EmailAddress struct {
	DisplayName string `json:"displayName"`
	Address     string `json:"address"`
}

type Content struct {
	Subject   string `json:"subject"`
	HTML      string `json:"html"`
	PlainText string `json:"plainText"`
}

type CustomHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Attachment struct {
	Content        string `json:"contentBytesBase64"`
	AttachmentType string `json:"attachmentType"`
	Name           string `json:"name"`
}

type ErrorResponse struct {
	Error CommunicationError `json:"error"`
}

// CommunicationError contains the error code and message
type CommunicationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type AzureACSEmailProvider struct {
	AccessKey string
	Endpoint  string
}

func NewAzureACSEmailProvider(accessKey string, endpoint string) *AzureACSEmailProvider {
	return &AzureACSEmailProvider{
		AccessKey: accessKey,
		Endpoint:  endpoint,
	}
}

func newEmail(fromAddress string, toAddress string, subject string, content string) *Email {
	return &Email{
		Recipients: Recipients{
			To: []EmailAddress{
				{
					DisplayName: toAddress,
					Address:     toAddress,
				},
			},
		},
		SenderAddress: fromAddress,
		Content: Content{
			Subject: subject,
			HTML:    content,
		},
		Importance: importanceNormal,
	}
}

func (a *AzureACSEmailProvider) sendEmail(e *Email) error {
	postBody, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("email JSON marshall failed: %s", err)
	}

	bodyBuffer := bytes.NewBuffer(postBody)

	endpoint := strings.TrimSuffix(a.Endpoint, "/")
	url := fmt.Sprintf("%s/emails:send?api-version=2023-03-31", endpoint)
	req, err := http.NewRequest("POST", url, bodyBuffer)
	if err != nil {
		return fmt.Errorf("error creating AzureACS API request: %s", err)
	}

	// Sign the request using the AzureACS access key and HMAC-SHA256
	err = signRequestHMAC(a.AccessKey, req)
	if err != nil {
		return fmt.Errorf("error signing AzureACS API request: %s", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Some important header
	req.Header.Set("repeatability-request-id", uuid.New().String())
	req.Header.Set("repeatability-first-sent", time.Now().UTC().Format(http.TimeFormat))

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending AzureACS API request: %s", err)
	}
	defer resp.Body.Close()

	// Response error Handling
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized {
		commError := ErrorResponse{}

		err = json.NewDecoder(resp.Body).Decode(&commError)
		if err != nil {
			return err
		}

		return fmt.Errorf("error sending email: %s", commError.Error.Message)
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("error sending email: status: %d", resp.StatusCode)
	}

	return nil
}

func signRequestHMAC(secret string, req *http.Request) error {
	method := req.Method
	host := req.URL.Host
	pathAndQuery := req.URL.Path

	if req.URL.RawQuery != "" {
		pathAndQuery = pathAndQuery + "?" + req.URL.RawQuery
	}

	var content []byte
	var err error
	if req.Body != nil {
		content, err = io.ReadAll(req.Body)
		if err != nil {
			// return err
			content = []byte{}
		}
	}

	req.Body = io.NopCloser(bytes.NewBuffer(content))

	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return fmt.Errorf("error decoding secret: %s", err)
	}

	timestamp := time.Now().UTC().Format(http.TimeFormat)
	contentHash := GetContentHashBase64(content)
	stringToSign := fmt.Sprintf("%s\n%s\n%s;%s;%s", strings.ToUpper(method), pathAndQuery, timestamp, host, contentHash)
	signature := GetHmac(stringToSign, key)

	req.Header.Set("x-ms-content-sha256", contentHash)
	req.Header.Set("x-ms-date", timestamp)

	req.Header.Set("Authorization", "HMAC-SHA256 SignedHeaders=x-ms-date;host;x-ms-content-sha256&Signature="+signature)

	return nil
}

func GetContentHashBase64(content []byte) string {
	hasher := sha256.New()
	hasher.Write(content)

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}

func GetHmac(content string, key []byte) string {
	hmac := hmac.New(sha256.New, key)
	hmac.Write([]byte(content))

	return base64.StdEncoding.EncodeToString(hmac.Sum(nil))
}

func (a *AzureACSEmailProvider) Send(fromAddress string, fromName string, toAddress string, subject string, content string) error {
	e := newEmail(fromAddress, toAddress, subject, content)

	return a.sendEmail(e)
}
