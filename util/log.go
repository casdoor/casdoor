package util

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/adapter/context"
	"github.com/beego/beego/v2/adapter/logs"
)

func GetIPInfo(clientIP string) string {
	if clientIP == "" {
		return ""
	}

	ips := strings.Split(clientIP, ",")
	res := ""
	for i := range ips {
		ip := strings.TrimSpace(ips[i])
		//desc := GetDescFromIP(ip)
		ipstr := fmt.Sprintf("%s: %s", ip, "")
		if i != len(ips)-1 {
			res += ipstr + " -> "
		} else {
			res += ipstr
		}
	}

	return res
}

func getIPFromRequest(req *http.Request) string {
	clientIP := req.Header.Get("x-forwarded-for")
	if clientIP == "" {
		ipPort := strings.Split(req.RemoteAddr, ":")
		if len(ipPort) >= 1 && len(ipPort) <= 2 {
			clientIP = ipPort[0]
		} else if len(ipPort) > 2 {
			idx := strings.LastIndex(req.RemoteAddr, ":")
			clientIP = req.RemoteAddr[0:idx]
			clientIP = strings.TrimLeft(clientIP, "[")
			clientIP = strings.TrimRight(clientIP, "]")
		}
	}

	return GetIPInfo(clientIP)
}

func LogInfo(ctx *context.Context, f string, v ...interface{}) {
	ipString := fmt.Sprintf("(%s) ", getIPFromRequest(ctx.Request))
	logs.Info(ipString+f, v...)
}

func LogWarning(ctx *context.Context, f string, v ...interface{}) {
	ipString := fmt.Sprintf("(%s) ", getIPFromRequest(ctx.Request))
	logs.Warning(ipString+f, v...)
}
