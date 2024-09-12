package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

type CSTNETIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

func NewCSTNETIdProvider(clientId string, clientSecret string, redirectUrl string) *CSTNETIdProvider {
	idp := &CSTNETIdProvider{}
	idp.Config = idp.getConfig(clientId, clientSecret, redirectUrl)
	return idp
}

func (idp *CSTNETIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *CSTNETIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	endpoint := oauth2.Endpoint{
		AuthURL:  "https://passport.escience.cn/oauth2/authorize",
		TokenURL: "https://passport.escience.cn/oauth2/token",
	}

	config := &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Scopes:       []string{"all"},
		Endpoint:     endpoint,
	}

	return config
}

func (idp *CSTNETIdProvider) GetToken(code string) (*oauth2.Token, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("client_id", idp.Config.ClientID)
	values.Set("client_secret", idp.Config.ClientSecret)
	values.Set("redirect_uri", idp.Config.RedirectURL)

	resp, err := idp.Client.PostForm(idp.Config.Endpoint.TokenURL, values)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		UserInfo     string `json:"userInfo"`
	}

	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		Expiry:       time.Unix(time.Now().Unix()+int64(tokenResp.ExpiresIn), 0),
		RefreshToken: tokenResp.RefreshToken,
	}
	extraInfo := map[string]interface{}{
		"UserInfo": tokenResp.UserInfo,
	}
	newToken := token.WithExtra(extraInfo)

	return newToken, nil
}

/*
{"truename":"刘侃","umtId":"11253866","isPhoneVerified":"true","cstnetId":"liukan@wbgcas.cn","passwordType":"password_core_mail","type":"coreMail","cstnetIdStatus":"active","isIdCardVerified":false}
*/
type CSTNETUserInfo struct {
	Truename         string   `json:"truename"`
	UmtId            string   `json:"umtId"`
	IsPhoneVerified  string   `json:"isPhoneVerified"`
	CstnetId         string   `json:"cstnetId"`
	PasswordType     string   `json:"passwordType"`
	Type             string   `json:"type"`
	CstnetIdStatus   string   `json:"cstnetIdStatus"`
	IsIdCardVerified bool     `json:"isIdCardVerified"`
	SecurityEmail    string   `json:"securityEmail"`
	SecondaryEmails  []string `json:"secondaryEmails"`
}

func (idp *CSTNETIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	extraString := token.Extra("UserInfo").(string)
	/*
		{"truename":"刘侃","umtId":"11253866","isPhoneVerified":"true","cstnetId":"liukan@wbgcas.cn","passwordType":"password_core_mail","type":"coreMail","cstnetIdStatus":"active","isIdCardVerified":false}
	*/

	cstnetInfo := &CSTNETUserInfo{}
	if err := json.Unmarshal([]byte(extraString), cstnetInfo); err != nil {
		return nil, err
	}
	gravatarUrl := ""
	/*
		type UserInfo struct {
			Id          string
			Username    string
			DisplayName string
			UnionId     string
			Email       string
			Phone       string
			CountryCode string
			AvatarUrl   string
			Extra       map[string]string
		}
	*/
	extraUInfo := map[string]string{
		"isPhoneVerified":  cstnetInfo.IsPhoneVerified,
		"passwordType":     cstnetInfo.PasswordType,
		"type":             cstnetInfo.Type,
		"cstnetIdStatus":   cstnetInfo.CstnetIdStatus,
		"isIdCardVerified": fmt.Sprintf("%v", cstnetInfo.IsIdCardVerified),
		"SecurityEmail":    "",
		"SecondaryEmails":  "[]",
	}
	userInfo := &UserInfo{
		Id:          cstnetInfo.CstnetId,
		Username:    cstnetInfo.CstnetId,
		DisplayName: cstnetInfo.Truename,
		UnionId:     cstnetInfo.UmtId,
		Email:       cstnetInfo.CstnetId,
		AvatarUrl:   gravatarUrl, // CSTNET doesn't provide avatar information
		Extra:       extraUInfo,
	}

	return userInfo, nil
}
