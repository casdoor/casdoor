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

package util

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/beego/beego/context"
	"github.com/beego/beego/logs"
)

func getIpInfo(clientIp string) string {
	if clientIp == "" {
		return ""
	}

	ips := strings.Split(clientIp, ",")
	res := strings.TrimSpace(ips[0])
	//res := ""
	//for i := range ips {
	//	ip := strings.TrimSpace(ips[i])
	//	ipstr := fmt.Sprintf("%s: %s", ip, "")
	//	if i != len(ips)-1 {
	//		res += ipstr + " -> "
	//	} else {
	//		res += ipstr
	//	}
	//}

	return res
}

func GetClientIpFromRequest(req *http.Request) string {
	clientIp := req.Header.Get("x-forwarded-for")
	if clientIp == "" {
		ipPort := strings.Split(req.RemoteAddr, ":")
		if len(ipPort) >= 1 && len(ipPort) <= 2 {
			clientIp = ipPort[0]
		} else if len(ipPort) > 2 {
			idx := strings.LastIndex(req.RemoteAddr, ":")
			clientIp = req.RemoteAddr[0:idx]
			clientIp = strings.TrimLeft(clientIp, "[")
			clientIp = strings.TrimRight(clientIp, "]")
		}
	}

	return getIpInfo(clientIp)
}

func LogInfo(ctx *context.Context, f string, v ...interface{}) {
	ipString := fmt.Sprintf("(%s) ", GetClientIpFromRequest(ctx.Request))
	logs.Info(ipString+f, v...)
}

func LogWarning(ctx *context.Context, f string, v ...interface{}) {
	ipString := fmt.Sprintf("(%s) ", GetClientIpFromRequest(ctx.Request))
	logs.Warning(ipString+f, v...)
}
