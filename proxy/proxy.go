// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
)

var (
	DefaultHttpClient *http.Client
	ProxyHttpClient   *http.Client
)

func InitHttpClient() {
	// not use proxy
	DefaultHttpClient = http.DefaultClient

	// use proxy
	ProxyHttpClient = getProxyHttpClient()
}

func isAddressOpen(address string) bool {
	timeout := time.Millisecond * 100
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		// cannot connect to address, proxy is not active
		return false
	}

	if conn != nil {
		defer conn.Close()
		fmt.Printf("Socks5 proxy enabled: %s\n", address)
		return true
	}

	return false
}

func getProxyHttpClient() *http.Client {
	socks5Proxy := conf.GetConfigString("socks5Proxy")
	if socks5Proxy == "" {
		return &http.Client{}
	}

	proxyURL, err := url.Parse(socks5Proxy)
	if err != nil {
		fmt.Println("Socks5 parse error using default HTTP client", err)
		return &http.Client{}
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL), 
	}
	return &http.Client{Transport: transport}
}

func GetHttpClient(url string) *http.Client {
	if strings.Contains(url, "githubusercontent.com") || strings.Contains(url, "googleusercontent.com") {
		return ProxyHttpClient
	} else {
		return DefaultHttpClient
	}
}
