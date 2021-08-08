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
	"net/http"

	"github.com/astaxie/beego"
	"golang.org/x/net/proxy"
)

var defaultHttpClient *http.Client
var proxyHttpClient *http.Client

// InitHttpClient ...
func InitHttpClient() {
	// not use proxy
	defaultHttpClient = http.DefaultClient

	// use proxy
	httpProxy := beego.AppConfig.String("httpProxy")
	if httpProxy == "" {
		proxyHttpClient = &http.Client{}
		return
	}

	// https://stackoverflow.com/questions/33585587/creating-a-go-socks5-client
	dialer, err := proxy.SOCKS5("tcp", httpProxy, nil, proxy.Direct)
	if err != nil {
		panic(err)
	}

	tr := &http.Transport{Dial: dialer.Dial}
	proxyHttpClient = &http.Client{
		Transport: tr,
	}

	//resp, err2 := proxyHttpClient.Get("https://google.com")
	//if err2 != nil {
	//	panic(err2)
	//}
	//defer resp.Body.Close()
	//println("Response status: %s", resp.Status)
}

// ResponseOk ...
func (c *ApiController) ResponseOk(data ...interface{}) {
	resp := Response{Status: "ok"}
	switch len(data) {
	case 2:
		resp.Data2 = data[1]
		fallthrough
	case 1:
		resp.Data = data[0]
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// ResponseError ...
func (c *ApiController) ResponseError(error string, data ...interface{}) {
	resp := Response{Status: "error", Msg: error}
	switch len(data) {
	case 2:
		resp.Data2 = data[1]
		fallthrough
	case 1:
		resp.Data = data[0]
	}
	c.Data["json"] = resp
	c.ServeJSON()
}

// RequireSignedIn ...
func (c *ApiController) RequireSignedIn() (string, bool) {
	userId := c.GetSessionUsername()
	if userId == "" {
		c.ResponseError("Please sign in first")
		return "", false
	}
	return userId, true
}
