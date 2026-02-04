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

package idp

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"golang.org/x/oauth2"
)

type AlipayIdProvider struct {
	Client           *http.Client
	Config           *oauth2.Config
	AppCertSN        string // Application certificate SN
	AlipayRootCertSN string // Alipay root certificate SN
}

// NewAlipayIdProvider ...
func NewAlipayIdProvider(clientId string, clientSecret string, redirectUrl string) *AlipayIdProvider {
	idp := &AlipayIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// NewAlipayIdProviderWithCert creates a new AlipayIdProvider with certificate mode support
func NewAlipayIdProviderWithCert(clientId string, clientSecret string, redirectUrl string, appCert string, alipayRootCert string) *AlipayIdProvider {
	idp := NewAlipayIdProvider(clientId, clientSecret, redirectUrl)

	// Calculate certificate SNs if certificates are provided
	if appCert != "" {
		sn, err := calculateCertSN(appCert)
		if err != nil {
			logs.Warning("[Alipay] Failed to calculate app_cert_sn: %v", err)
		} else {
			idp.AppCertSN = sn
			logs.Info("[Alipay] Calculated app_cert_sn: %s", sn)
		}
	}

	if alipayRootCert != "" {
		sn, err := calculateRootCertSN(alipayRootCert)
		if err != nil {
			logs.Warning("[Alipay] Failed to calculate alipay_root_cert_sn: %v", err)
		} else {
			idp.AlipayRootCertSN = sn
			logs.Info("[Alipay] Calculated alipay_root_cert_sn: %s", sn)
		}
	}

	return idp
}

// SetHttpClient ...
func (idp *AlipayIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *AlipayIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm",
		TokenURL: "https://openapi.alipay.com/gateway.do",
	}

	config := &oauth2.Config{
		Scopes:       []string{"", ""},
		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type AlipayAccessToken struct {
	Response      AlipaySystemOauthTokenResponse `json:"alipay_system_oauth_token_response"`
	ErrorResponse *AlipayErrorResponse           `json:"error_response"`
	Sign          string                         `json:"sign"`
	// Legacy error response fields (fallback)
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

type AlipayErrorResponse struct {
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

type AlipaySystemOauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	AlipayUserId string `json:"alipay_user_id"`
	ExpiresIn    int    `json:"expires_in"`
	ReExpiresIn  int    `json:"re_expires_in"`
	RefreshToken string `json:"refresh_token"`
	UserId       string `json:"user_id"`
	// Error response fields (can also appear in nested response)
	Code    string `json:"code"`
	Msg     string `json:"msg"`
	SubCode string `json:"sub_code"`
	SubMsg  string `json:"sub_msg"`
}

// GetToken use code to get access_token
func (idp *AlipayIdProvider) GetToken(code string) (*oauth2.Token, error) {
	// Build request parameters
	params := map[string]string{
		"app_id":     idp.Config.ClientID,
		"charset":    "utf-8",
		"code":       code,
		"grant_type": "authorization_code",
		"method":     "alipay.system.oauth.token",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
	}

	// Add certificate SNs if using certificate mode
	if idp.AppCertSN != "" {
		params["app_cert_sn"] = idp.AppCertSN
	}
	if idp.AlipayRootCertSN != "" {
		params["alipay_root_cert_sn"] = idp.AlipayRootCertSN
	}

	data, err := idp.postWithParams(params, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	pToken := &AlipayAccessToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, fmt.Errorf("failed to parse alipay response: %w, response: %s", err, strings.TrimSpace(string(data)))
	}

	// Check for error_response (newer format)
	if pToken.ErrorResponse != nil && pToken.ErrorResponse.Code != "" && pToken.ErrorResponse.Code != "10000" {
		logs.Warning("[Alipay] Token API error response: %s", strings.TrimSpace(string(data)))
		var errMsg string
		if pToken.ErrorResponse.Msg != "" && pToken.ErrorResponse.SubMsg != "" {
			// If both msg and sub_msg exist, combine them
			errMsg = fmt.Sprintf("%s - %s", pToken.ErrorResponse.Msg, pToken.ErrorResponse.SubMsg)
		} else if pToken.ErrorResponse.Msg != "" {
			errMsg = pToken.ErrorResponse.Msg
		} else if pToken.ErrorResponse.SubMsg != "" {
			errMsg = pToken.ErrorResponse.SubMsg
		} else {
			errMsg = fmt.Sprintf("code: %s", pToken.ErrorResponse.Code)
		}
		return nil, fmt.Errorf("alipay API error: %s", errMsg)
	}

	// Check for error response at root level (legacy format)
	if pToken.Code != "" && pToken.Code != "10000" {
		logs.Warning("[Alipay] Token API error response (root): %s", strings.TrimSpace(string(data)))
		errMsg := pToken.Msg
		if pToken.SubMsg != "" {
			errMsg = fmt.Sprintf("%s (sub_code: %s, sub_msg: %s)", pToken.Msg, pToken.SubCode, pToken.SubMsg)
		}
		return nil, fmt.Errorf("alipay API error: %s", errMsg)
	}

	// Check for error response in nested response object
	if pToken.Response.Code != "" && pToken.Response.Code != "10000" {
		logs.Warning("[Alipay] Token API error response (nested): %s", strings.TrimSpace(string(data)))
		errMsg := pToken.Response.Msg
		if pToken.Response.SubMsg != "" {
			errMsg = fmt.Sprintf("%s (sub_code: %s, sub_msg: %s)", pToken.Response.Msg, pToken.Response.SubCode, pToken.Response.SubMsg)
		}
		return nil, fmt.Errorf("alipay API error: %s", errMsg)
	}

	// Check if access token is empty
	if pToken.Response.AccessToken == "" {
		logs.Warning("[Alipay] Token API missing access_token, response: %s", strings.TrimSpace(string(data)))
		return nil, fmt.Errorf("alipay API returned empty access token, response: %s", strings.TrimSpace(string(data)))
	}

	token := &oauth2.Token{
		AccessToken: pToken.Response.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+int64(pToken.Response.ExpiresIn), 0),
	}
	return token, nil
}

/*
{
    "alipay_user_info_share_response":{
        "code":"10000",
        "msg":"Success",
        "avatar":"https:\/\/tfs.alipayobjects.com\/images\/partner\/T1.QxFXk4aXXXXXXXX",
        "nick_name":"zhangsan",
        "user_id":"2099222233334444"
    },
    "sign":"m8rWJeqfoa5tDQRRVnPhRHcpX7NZEgjIPTPF1QBxos6XXXXXXXXXXXXXXXXXXXXXXXXXX"
}
*/

type AlipayUserResponse struct {
	AlipayUserInfoShareResponse AlipayUserInfoShareResponse `json:"alipay_user_info_share_response"`
	Sign                        string                      `json:"sign"`
}

type AlipayUserInfoShareResponse struct {
	Code     string `json:"code"`
	Msg      string `json:"msg"`
	SubCode  string `json:"sub_code"`
	SubMsg   string `json:"sub_msg"`
	Avatar   string `json:"avatar"`
	NickName string `json:"nick_name"`
	UserId   string `json:"user_id"`
	OpenId   string `json:"open_id"`
}

// GetUserInfo Use access_token to get UserInfo
func (idp *AlipayIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	atUserInfo := &AlipayUserResponse{}
	accessToken := token.AccessToken

	// Build request parameters
	params := map[string]string{
		"app_id":     idp.Config.ClientID,
		"charset":    "utf-8",
		"auth_token": accessToken,
		"method":     "alipay.user.info.share",
		"sign_type":  "RSA2",
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"version":    "1.0",
	}

	// Add certificate SNs if using certificate mode
	if idp.AppCertSN != "" {
		params["app_cert_sn"] = idp.AppCertSN
	}
	if idp.AlipayRootCertSN != "" {
		params["alipay_root_cert_sn"] = idp.AlipayRootCertSN
	}

	data, err := idp.postWithParams(params, idp.Config.Endpoint.TokenURL)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, atUserInfo)
	if err != nil {
		return nil, err
	}

	// Check for error response
	if atUserInfo.AlipayUserInfoShareResponse.Code != "" && atUserInfo.AlipayUserInfoShareResponse.Code != "10000" {
		logs.Warning("[Alipay] UserInfo API error response: %s", strings.TrimSpace(string(data)))
		var errMsg string
		if atUserInfo.AlipayUserInfoShareResponse.Msg != "" && atUserInfo.AlipayUserInfoShareResponse.SubMsg != "" {
			// If both msg and sub_msg exist, combine them
			errMsg = fmt.Sprintf("%s - %s", atUserInfo.AlipayUserInfoShareResponse.Msg, atUserInfo.AlipayUserInfoShareResponse.SubMsg)
		} else if atUserInfo.AlipayUserInfoShareResponse.Msg != "" {
			errMsg = atUserInfo.AlipayUserInfoShareResponse.Msg
		} else if atUserInfo.AlipayUserInfoShareResponse.SubMsg != "" {
			errMsg = atUserInfo.AlipayUserInfoShareResponse.SubMsg
		} else {
			errMsg = fmt.Sprintf("code: %s", atUserInfo.AlipayUserInfoShareResponse.Code)
		}
		return nil, fmt.Errorf("alipay API error: %s", errMsg)
	}

	// Determine unique ID: prefer user_id, fallback to open_id
	userId := atUserInfo.AlipayUserInfoShareResponse.UserId
	if userId == "" {
		userId = atUserInfo.AlipayUserInfoShareResponse.OpenId
	}

	userInfo := UserInfo{
		Id:          userId,
		Username:    userId,
		DisplayName: atUserInfo.AlipayUserInfoShareResponse.NickName,
		AvatarUrl:   atUserInfo.AlipayUserInfoShareResponse.Avatar,
	}

	return &userInfo, nil
}

func (idp *AlipayIdProvider) postWithBody(body interface{}, targetUrl string) ([]byte, error) {
	bs, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	bodyJson := make(map[string]interface{})
	err = json.Unmarshal(bs, &bodyJson)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for k := range bodyJson {
		params[k] = bodyJson[k].(string)
	}

	return idp.postWithParams(params, targetUrl)
}

func (idp *AlipayIdProvider) postWithParams(params map[string]string, targetUrl string) ([]byte, error) {
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	stringToSign := getStringToSign(formData)
	sign, err := rsaSignWithRSA256(stringToSign, idp.Config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign alipay request: %w", err)
	}

	formData.Set("sign", sign)

	resp, err := idp.Client.Post(targetUrl, "application/x-www-form-urlencoded;charset=utf-8", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("alipay API http status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	return data, nil
}

// get the string to sign, see https://opendocs.alipay.com/common/02kf5q
// According to Alipay docs, the string to sign should use raw (non-URL-encoded) parameter values
func getStringToSign(formData url.Values) string {
	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		if k == "sign" || len(formData[k]) == 0 || formData[k][0] == "" {
			continue
		}
		// Alipay signature string should use raw values WITHOUT URL encoding
		// See: https://opendocs.alipay.com/common/02kf5q
		parts = append(parts, k+"="+formData[k][0])
	}

	return strings.Join(parts, "&")
}

// use privateKey to sign the content
func rsaSignWithRSA256(signContent string, privateKey string) (string, error) {
	privateKey = formatPrivateKey(privateKey)

	if len(privateKey) == 0 {
		return "", fmt.Errorf("private key is empty after formatting")
	}

	// Try to parse and sign with the formatted key (PKCS#8 format by default)
	signature, _, err := trySignWithKey(privateKey, signContent)
	if err == nil {
		return signature, nil
	}

	// If PKCS#8 fails, try PKCS#1 format as fallback
	privateKeyPKCS1 := convertToPKCS1Format(privateKey)

	signature, _, err2 := trySignWithKey(privateKeyPKCS1, signContent)
	if err2 == nil {
		return signature, nil
	}

	// Both formats failed, return the original PKCS#8 error
	return "", fmt.Errorf("failed to sign with private key: PKCS#8 error: %v, PKCS#1 error: %v", err, err2)
}

// trySignWithKey attempts to sign content with the given PEM-formatted key
// returns (signature, usedPKCS8, error)
func trySignWithKey(pemKey string, signContent string) (string, bool, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return "", false, fmt.Errorf("failed to parse PEM block")
	}

	h := sha256.New()
	h.Write([]byte(signContent))
	hashed := h.Sum(nil)

	// Try PKCS8 first
	usedPKCS8 := true
	privateKeyRSA, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// Fall back to PKCS1 if PKCS8 parsing fails
		parsedKey, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err2 != nil {
			return "", false, fmt.Errorf("both PKCS8 and PKCS1 parsing failed: PKCS8: %v, PKCS1: %v", err, err2)
		}
		privateKeyRSA = parsedKey
		usedPKCS8 = false
	}

	rsaKey, ok := privateKeyRSA.(*rsa.PrivateKey)
	if !ok {
		return "", false, fmt.Errorf("parsed key is not an RSA private key, type: %T", privateKeyRSA)
	}

	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed)
	if err != nil {
		return "", false, fmt.Errorf("failed to sign with RSA: %v", err)
	}

	return base64.StdEncoding.EncodeToString(signature), usedPKCS8, nil
}

// convertToPKCS1Format converts a key from PKCS#8 format to PKCS#1 format
// by extracting the base64 content and re-wrapping with RSA PRIVATE KEY headers
func convertToPKCS1Format(pkcs8Key string) string {
	// Extract the base64 content between BEGIN and END
	lines := strings.Split(strings.TrimSpace(pkcs8Key), "\n")
	var keyContent strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "-----") {
			keyContent.WriteString(line)
		}
	}

	// Rewrap with PKCS#1 format
	base64Content := keyContent.String()
	if len(base64Content) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("-----BEGIN RSA PRIVATE KEY-----\n")

	for i := 0; i < len(base64Content); i += 64 {
		end := i + 64
		if end > len(base64Content) {
			end = len(base64Content)
		}
		result.WriteString(base64Content[i:end])
		result.WriteString("\n")
	}

	result.WriteString("-----END RSA PRIVATE KEY-----")
	return result.String()
}

// privateKey in database is a string, format it to PEM style
func formatPrivateKey(privateKey string) string {
	// Trim all leading and trailing whitespace
	privateKey = strings.TrimSpace(privateKey)

	// If already in valid PEM format, validate and return as-is
	if strings.HasPrefix(privateKey, "-----BEGIN") && strings.HasSuffix(privateKey, "-----") {
		// Verify it's a complete PEM structure
		lines := strings.Split(privateKey, "\n")
		if len(lines) >= 3 {
			return privateKey
		}
	}

	// Remove all whitespace including newlines and carriage returns
	privateKey = strings.ReplaceAll(privateKey, "\n", "")
	privateKey = strings.ReplaceAll(privateKey, "\r", "")
	privateKey = strings.ReplaceAll(privateKey, " ", "")
	privateKey = strings.ReplaceAll(privateKey, "\t", "")

	// Remove PEM headers/footers if embedded in the key
	privateKey = strings.TrimPrefix(privateKey, "-----BEGINPRIVATEKEY-----")
	privateKey = strings.TrimPrefix(privateKey, "-----BEGINRSAPRIVATEKEY-----")
	privateKey = strings.TrimSuffix(privateKey, "-----ENDPRIVATEKEY-----")
	privateKey = strings.TrimSuffix(privateKey, "-----ENDRSAPRIVATEKEY-----")

	// If empty after stripping, something went wrong
	if len(privateKey) == 0 {
		return ""
	}

	// Format with proper line breaks (64 chars per line, which is standard)
	// Try PKCS#8 format first (-----BEGIN PRIVATE KEY-----)
	var formattedKey strings.Builder
	formattedKey.WriteString("-----BEGIN PRIVATE KEY-----\n")

	for i := 0; i < len(privateKey); i += 64 {
		end := i + 64
		if end > len(privateKey) {
			end = len(privateKey)
		}
		formattedKey.WriteString(privateKey[i:end])
		formattedKey.WriteString("\n")
	}

	formattedKey.WriteString("-----END PRIVATE KEY-----")
	return formattedKey.String()
}

// calculateCertSN calculates the certificate serial number for Alipay
// SN = MD5(issuer + serialNumber) in hex format
// See: https://opendocs.alipay.com/common/057k2t
func calculateCertSN(certPEM string) (string, error) {
	cert, err := parseCertificate(certPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Get issuer DN in RFC2253 format and serial number
	issuer := cert.Issuer.String()
	serialNumber := cert.SerialNumber.String()

	// Calculate MD5 hash of issuer + serialNumber
	data := issuer + serialNumber
	hash := md5.Sum([]byte(data))
	sn := hex.EncodeToString(hash[:])

	return sn, nil
}

// calculateRootCertSN calculates the root certificate SN for Alipay
// For root cert, we need to extract all RSA certificates and concatenate their SNs with "_"
// See: https://opendocs.alipay.com/common/057k2t
func calculateRootCertSN(rootCertPEM string) (string, error) {
	var sns []string

	// Parse all certificates in the PEM file
	remaining := []byte(rootCertPEM)
	for {
		var block *pem.Block
		block, remaining = pem.Decode(remaining)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		// Only include RSA certificates (signature algorithm contains RSA)
		sigAlg := cert.SignatureAlgorithm.String()
		if !strings.Contains(sigAlg, "RSA") {
			continue
		}

		// Calculate SN for this certificate
		issuer := cert.Issuer.String()
		serialNumber := cert.SerialNumber.String()
		data := issuer + serialNumber
		hash := md5.Sum([]byte(data))
		sn := hex.EncodeToString(hash[:])

		sns = append(sns, sn)
	}

	if len(sns) == 0 {
		return "", fmt.Errorf("no valid RSA certificates found in root cert")
	}

	return strings.Join(sns, "_"), nil
}

// parseCertificate parses a PEM-encoded certificate
func parseCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, nil
}
