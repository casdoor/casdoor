// Copyright 2023 The casbin Authors. All Rights Reserved.
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

package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/rule"
	"github.com/casdoor/casdoor/util"
)

func forwardHandler(targetUrl string, writer http.ResponseWriter, request *http.Request) {
	target, err := url.Parse(targetUrl)

	if nil != err {
		panic(err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Director = func(r *http.Request) {
		r.URL = target

		if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" && xff != clientIP {
				newXff := fmt.Sprintf("%s, %s", xff, clientIP)
				// r.Header.Set("X-Forwarded-For", newXff)
				r.Header.Set("X-Real-Ip", newXff)
			} else {
				// r.Header.Set("X-Forwarded-For", clientIP)
				r.Header.Set("X-Real-Ip", clientIP)
			}
		}
	}

	proxy.ModifyResponse = func(resp *http.Response) error {
		// Add Secure flag to all Set-Cookie headers in HTTPS responses
		if request.TLS != nil {
			// Add HSTS header for HTTPS responses if not already set by backend
			if resp.Header.Get("Strict-Transport-Security") == "" {
				resp.Header.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			cookies := resp.Header["Set-Cookie"]
			if len(cookies) > 0 {
				// Clear existing Set-Cookie headers
				resp.Header.Del("Set-Cookie")
				// Add them back with Secure flag if not already present
				for _, cookie := range cookies {
					// Check if Secure attribute is already present (case-insensitive)
					cookieLower := strings.ToLower(cookie)
					hasSecure := strings.Contains(cookieLower, ";secure;") ||
						strings.Contains(cookieLower, "; secure;") ||
						strings.HasSuffix(cookieLower, ";secure") ||
						strings.HasSuffix(cookieLower, "; secure")
					if !hasSecure {
						cookie = cookie + "; Secure"
					}
					resp.Header.Add("Set-Cookie", cookie)
				}
			}
		}

		// Fix CORS issue: Remove CORS header combinations that allow credential theft from any origin
		allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
		allowCredentials := resp.Header.Get("Access-Control-Allow-Credentials")

		// Remove CORS headers when the combination is present:
		// 1. Access-Control-Allow-Credentials: true with Access-Control-Allow-Origin: *
		//    This is actually blocked by browsers but we sanitize it anyway
		// 2. Access-Control-Allow-Credentials: true with any origin
		//    Without a configured allowlist, we cannot safely validate if the origin
		//    is trusted or if it's being reflected from the request, so we remove all
		//    CORS headers for credential-bearing responses to prevent theft
		if strings.EqualFold(allowCredentials, "true") && allowOrigin != "" {
			// Remove CORS headers to prevent credential theft
			resp.Header.Del("Access-Control-Allow-Origin")
			resp.Header.Del("Access-Control-Allow-Credentials")
			resp.Header.Del("Access-Control-Allow-Methods")
			resp.Header.Del("Access-Control-Allow-Headers")
			resp.Header.Del("Access-Control-Expose-Headers")
			resp.Header.Del("Access-Control-Max-Age")
		}

		return nil
	}

	proxy.ServeHTTP(writer, request)
}

func getHostNonWww(host string) string {
	res := ""
	tokens := strings.Split(host, ".")
	if len(tokens) > 2 && tokens[0] == "www" {
		res = strings.Join(tokens[1:], ".")
	}
	return res
}

func logRequest(clientIp string, r *http.Request) {
	if !strings.Contains(r.UserAgent(), "Uptime-Kuma") {
		fmt.Printf("handleRequest: %s\t%s\t%s\t%s\t%s\t%s\n", clientIp, r.Method, r.Host, r.RequestURI, r.UserAgent(), r.RemoteAddr)
		record := object.Record{
			Owner:       "admin",
			CreatedTime: util.GetCurrentTime(),
			Method:      r.Method,
			RequestUri:  r.RequestURI,
			ClientIp:    clientIp,
		}
		object.AddRecord(&record)
	}
}

func redirectToHttps(w http.ResponseWriter, r *http.Request) {
	targetUrl := fmt.Sprintf("https://%s", joinPath(r.Host, r.RequestURI))
	http.Redirect(w, r, targetUrl, http.StatusMovedPermanently)
}

func redirectToHost(w http.ResponseWriter, r *http.Request, host string) {
	protocol := "https"
	if r.TLS == nil {
		protocol = "http"
	}

	targetUrl := fmt.Sprintf("%s://%s", protocol, joinPath(host, r.RequestURI))
	http.Redirect(w, r, targetUrl, http.StatusMovedPermanently)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	clientIp := util.GetClientIp(r)
	logRequest(clientIp, r)

	site := getSiteByDomainWithWww(r.Host)
	if site == nil {
		if isHostIp(r.Host) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(r.Host, ".casdoor.com") && r.RequestURI == "/health-ping" {
			w.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintf(w, "OK")
			if err != nil {
				panic(err)
			}
			return
		}

		responseError(w, "CasWAF error: site not found for host: %s", r.Host)
		return
	}

	hostNonWww := getHostNonWww(r.Host)
	if hostNonWww != "" {
		redirectToHost(w, r, hostNonWww)
		return
	}

	if site.Domain != r.Host && site.NeedRedirect {
		redirectToHost(w, r, site.Domain)
		return
	}

	if strings.HasPrefix(r.RequestURI, "/.well-known/acme-challenge/") {
		challengeMap := site.GetChallengeMap()
		for token, keyAuth := range challengeMap {
			if r.RequestURI == fmt.Sprintf("/.well-known/acme-challenge/%s", token) {
				responseOk(w, "%s", keyAuth)
				return
			}
		}

		responseError(w, "CasWAF error: ACME HTTP-01 challenge failed, requestUri cannot match with challengeMap, requestUri = %s, challengeMap = %v", r.RequestURI, challengeMap)
		return
	}

	if strings.HasPrefix(r.RequestURI, "/MP_verify_") {
		challengeMap := site.GetChallengeMap()
		for path, value := range challengeMap {
			if r.RequestURI == fmt.Sprintf("/%s", path) {
				responseOk(w, "%s", value)
				return
			}
		}
	}

	if site.SslMode == "HTTPS Only" {
		// This domain only supports https but receive http request, redirect to https
		if r.TLS == nil {
			redirectToHttps(w, r)
			return
		}
	}

	// oAuth proxy
	if site.CasdoorApplication != "" {
		// handle oAuth proxy
		cookie, err := r.Cookie("casdoor_access_token")
		if err != nil && err.Error() != "http: named cookie not present" {
			panic(err)
		}

		casdoorClient, err := getCasdoorClientFromSite(site)
		if err != nil {
			responseError(w, "CasWAF error: getCasdoorClientFromSite() error: %s", err.Error())
			return
		}

		if cookie == nil {
			// not logged in
			redirectToCasdoor(casdoorClient, w, r)
			return
		} else {
			_, err = casdoorClient.ParseJwtToken(cookie.Value)
			if err != nil {
				responseError(w, "CasWAF error: casdoorClient.ParseJwtToken() error: %s", err.Error())
				return
			}
		}
	}

	host := site.GetHost()
	if host == "" {
		responseError(w, "CasWAF error: targetUrl should not be empty for host: %s, site = %v", r.Host, site)
		return
	}

	if len(site.Rules) == 0 {
		nextHandle(w, r)
		return
	}

	result, err := rule.CheckRules(site.Rules, r)
	if err != nil {
		responseError(w, "Internal Server Error: %v", err)
		return
	}

	reason := result.Reason
	if reason != "" && site.DisableVerbose {
		reason = "the rule has been hit"
	}

	switch result.Action {
	case "", "Allow":
		// Do not write header for Allow action, let the proxy handle it
	case "Block":
		w.WriteHeader(result.StatusCode)
		responseErrorWithoutCode(w, "Blocked by CasWAF: %s", reason)
		return
	case "Drop":
		w.WriteHeader(result.StatusCode)
		responseErrorWithoutCode(w, "Dropped by CasWAF: %s", reason)
		return
	default:
		responseError(w, "Error in CasWAF: %s", reason)
	}
	nextHandle(w, r)
}

func nextHandle(w http.ResponseWriter, r *http.Request) {
	site := getSiteByDomainWithWww(r.Host)
	host := site.GetHost()
	if site.SslMode == "Static Folder" {
		var path string
		if r.RequestURI != "/" {
			path = filepath.Join(host, r.RequestURI)
		} else {
			path = filepath.Join(host, "/index.htm")
			if !util.FileExist(path) {
				path = filepath.Join(host, "/index.html")
				if !util.FileExist(path) {
					path = filepath.Join(host, r.RequestURI)
				}
			}
		}
		http.ServeFile(w, r, path)
	} else {
		targetUrl := joinPath(site.GetHost(), r.RequestURI)
		forwardHandler(targetUrl, w, r)
	}
}

func Start() {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", handleRequest)
	serverMux.HandleFunc("/caswaf-handler", handleAuthCallback)

	gatewayHttpPort, err := conf.GetConfigInt64("gatewayHttpPort")
	if err != nil {
		gatewayHttpPort = 80
	}

	gatewayHttpsPort, err := conf.GetConfigInt64("gatewayHttpsPort")
	if err != nil {
		gatewayHttpsPort = 443
	}

	go func() {
		fmt.Printf("CasWAF gateway running on: http://127.0.0.1:%d\n", gatewayHttpPort)
		err := http.ListenAndServe(fmt.Sprintf(":%d", gatewayHttpPort), serverMux)
		if err != nil {
			logs.Error(err)
		}
	}()

	go func() {
		fmt.Printf("CasWAF gateway running on: https://127.0.0.1:%d\n", gatewayHttpsPort)
		server := &http.Server{
			Handler: serverMux,
			Addr:    fmt.Sprintf(":%d", gatewayHttpsPort),
			TLSConfig: &tls.Config{
				// Minimum TLS version 1.2, TLS 1.3 is automatically supported
				MinVersion: tls.VersionTLS12,
				// Secure cipher suites for TLS 1.2 (excluding 3DES to prevent Sweet32 attack)
				// TLS 1.3 cipher suites are automatically configured by Go
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				},
				// Prefer strong elliptic curves
				CurvePreferences: []tls.CurveID{
					tls.X25519,
					tls.CurveP256,
					tls.CurveP384,
				},
			},
		}

		// start https server and set how to get certificate
		server.TLSConfig.GetCertificate = func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			domain := info.ServerName
			cert, err := getX509CertByDomain(domain)
			if err != nil {
				return nil, err
			}

			return cert, nil
		}

		err := server.ListenAndServeTLS("", "")
		if err != nil {
			logs.Error(err)
		}
	}()
}
