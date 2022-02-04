package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

type InfoflowInternalIdProvider struct {
	Client  *http.Client
	Config  *oauth2.Config
	AgentId string
}

func NewInfoflowInternalIdProvider(clientId string, clientSecret string, appId string, redirectUrl string) *InfoflowInternalIdProvider {
	idp := &InfoflowInternalIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config
	idp.AgentId = appId
	return idp
}

func (idp *InfoflowInternalIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *InfoflowInternalIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	var config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type InfoflowInterToken struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
}

func (idp *InfoflowInternalIdProvider) GetToken(code string) (*oauth2.Token, error) {
	pTokenParams := &struct {
		CorpId     string `json:"corpid"`
		Corpsecret string `json:"corpsecret"`
	}{idp.Config.ClientID, idp.Config.ClientSecret}
	resp, err := idp.Client.Get(fmt.Sprintf("https://qy.im.baidu.com/api/gettoken?corpid=%s&corpsecret=%s", pTokenParams.CorpId, pTokenParams.Corpsecret))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	pToken := &InfoflowInterToken{}
	err = json.Unmarshal(data, pToken)
	if err != nil {
		return nil, err
	}
	if pToken.Errcode != 0 {
		return nil, fmt.Errorf("pToken.Errcode = %d, pToken.Errmsg = %s", pToken.Errcode, pToken.Errmsg)
	}
	token := &oauth2.Token{
		AccessToken: pToken.AccessToken,
	}

	raw := make(map[string]interface{})
	raw["code"] = code
	token = token.WithExtra(raw)

	return token, nil
}

type InfoflowInternalUserResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	UserId  string `json:"UserId"`
}

type InfoflowInternalUserInfo struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Name    string `json:"name"`
	Email   string `json:"email"`
}

func (idp *InfoflowInternalIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	//Get userid first
	accessToken := token.AccessToken
	code := token.Extra("code").(string)
	resp, err := idp.Client.Get(fmt.Sprintf("https://qy.im.baidu.com/api/user/getuserinfo?access_token=%s&code=%s&agentid=%s", accessToken, code, idp.AgentId))
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	userResp := &InfoflowInternalUserResp{}
	err = json.Unmarshal(data, userResp)
	if err != nil {
		return nil, err
	}
	if userResp.Errcode != 0 {
		return nil, fmt.Errorf("userIdResp.Errcode = %d, userIdResp.Errmsg = %s", userResp.Errcode, userResp.Errmsg)
	}
	//Use userid and accesstoken to get user information
	resp, err = idp.Client.Get(fmt.Sprintf("https://api.im.baidu.com/api/user/get?access_token=%s&userid=%s", accessToken, userResp.UserId))
	if err != nil {
		return nil, err
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	infoResp := &InfoflowInternalUserInfo{}
	err = json.Unmarshal(data, infoResp)
	if err != nil {
		return nil, err
	}
	if infoResp.Errcode != 0 {
		return nil, fmt.Errorf("userInfoResp.errcode = %d, userInfoResp.errmsg = %s", infoResp.Errcode, infoResp.Errmsg)
	}
	userInfo := UserInfo{
		Username:    infoResp.Name,
		DisplayName: infoResp.Name,
		Email:       infoResp.Email,
	}

	if userInfo.Id == "" {
		userInfo.Id = userInfo.Username
	}
	return &userInfo, nil
}
