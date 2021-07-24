// Copyright 2021 The casbin Authors. All Rights Reserved.
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

package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"golang.org/x/net/proxy"
)

var httpClient *http.Client

func InitHttpClient() {
	httpProxy := beego.AppConfig.String("httpProxy")
	if httpProxy == "" {
		httpClient = &http.Client{}
		return
	}

	// https://stackoverflow.com/questions/33585587/creating-a-go-socks5-client
	dialer, err := proxy.SOCKS5("tcp", httpProxy, nil, proxy.Direct)
	if err != nil {
		panic(err)
	}

	tr := &http.Transport{Dial: dialer.Dial}
	httpClient = &http.Client{
		Transport: tr,
	}

	//resp, err2 := httpClient.Get("https://google.com")
	//if err2 != nil {
	//	panic(err2)
	//}
	//defer resp.Body.Close()
	//println("Response status: %s", resp.Status)
}

func (c *ApiController) ResponseError(error string) {
	c.Data["json"] = Response{Status: "error", Msg: error}
	c.ServeJSON()
}

func (c *ApiController) ResponseErrorWithData(error string, data interface{}) {
	c.Data["json"] = Response{Status: "error", Msg: error, Data: data}
	c.ServeJSON()
}

func (c *ApiController) RequireSignedIn() (string, bool) {
	userId := c.GetSessionUsername()
	if userId == "" {
		resp := Response{Status: "error", Msg: "Please sign in first"}
		c.Data["json"] = resp
		c.ServeJSON()
		return "", false
	}
	return userId, true
}

type RequestTokenResp struct {
	OauthToken           string `json:"oauth_token"`
	OauthTokenSecret     string `json:"oauth_token_secret"`
	OauthCallbackConfirm string `json:"oauth_callback_confirm"`
}

func (c *ApiController) GetRequestToken() {
	oauthCallback := url.QueryEscape(c.Input().Get("oauth_callback"))
	clientId := c.Input().Get("client_id")
	p := object.GetProviderByClientId(clientId)

	nonce, ts := util.GenerateClientSecret(), strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	hParams := fmt.Sprintf("oauth_nonce=%s,oauth_callback=%s,oauth_signature_method=HMAC-SHA1,oauth_timestamp=%s,oauth_consumer_key=%s", nonce, oauthCallback, ts, clientId)
	v := url.Values{
		"oauth_consumer_key": []string{clientId},
		"oauth_callback": []string{oauthCallback},
		"oauth_nonce": []string{nonce},
		"oauth_signature_method": []string{"HMAC-SHA1"},
		"oauth_timestamp": []string{ts},
		"oauth_version": []string{"1.0"},
	}

	sig := c.createSignature("POST", url.QueryEscape("https://api.twitter.com/oauth/request_token"), v.Encode(), url.QueryEscape(p.ClientSecret)+"&")
	r, _ := http.NewRequest("POST", "https://api.twitter.com/oauth/request_token?oauth_callback="+oauthCallback, nil)
	r.Header.Set("OAuth", fmt.Sprintf("%s,oauth_signature=%s,oauth_version=1.0", hParams, url.QueryEscape(sig)))

	fmt.Printf("\nOauth Header:%s\n", r.Header.Get("OAuth"))

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Println(err)
		c.ResponseError(err.Error())
		return
	}

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		c.ResponseError("Internal Error")
	}
	defer resp.Body.Close()

	fmt.Printf("\n\nresponse:%s\n\n",string(all))

	token := &RequestTokenResp{}
	if err = json.Unmarshal(all, token); err != nil {
		log.Println(err)
		c.ResponseError(err.Error())
		return
	}

	fmt.Printf("\n\n%s,%s,%s\n\n", token.OauthToken, token.OauthTokenSecret, token.OauthCallbackConfirm)

	if token.OauthCallbackConfirm != oauthCallback {
		c.ResponseError("inconsistent callback url")
		return 
	}

	c.Data["json"] = token
	c.ServeJSON()
}

func (c *ApiController) createSignature(method, uri, param, key string) string {
	param = strings.ReplaceAll(param, "&", "%26")
	fmt.Printf("\nkey:%s\n", key)

	base := fmt.Sprintf("%s&%s&%s", method, uri, param)
	base = strings.ReplaceAll(base, "=", "%3D")
	fmt.Printf("\nbase:%s\n", base)

	sig := util.EncodeSHA1AndBase64(base, key)

	return sig
}