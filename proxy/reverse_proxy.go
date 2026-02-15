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
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
)

// Application represents a simplified application structure for reverse proxy
type Application struct {
	Owner        string
	Name         string
	UpstreamHost string
}

// ApplicationLookupFunc is a function type for looking up applications by domain
type ApplicationLookupFunc func(domain string) *Application

var applicationLookup ApplicationLookupFunc

// SetApplicationLookup sets the function to use for looking up applications by domain
func SetApplicationLookup(lookupFunc ApplicationLookupFunc) {
	applicationLookup = lookupFunc
}

// getDomainWithoutPort removes the port from a domain string
func getDomainWithoutPort(domain string) string {
	if !strings.Contains(domain, ":") {
		return domain
	}

	tokens := strings.SplitN(domain, ":", 2)
	if len(tokens) > 1 {
		return tokens[0]
	}
	return domain
}

// forwardHandler creates and configures a reverse proxy for the given target URL
func forwardHandler(targetUrl string, writer http.ResponseWriter, request *http.Request) {
	target, err := url.Parse(targetUrl)
	if err != nil {
		logs.Error("Failed to parse target URL %s: %v", targetUrl, err)
		http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Configure the Director to set proper headers
	proxy.Director = func(r *http.Request) {
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.Host = target.Host

		// Set X-Real-IP and X-Forwarded-For headers
		if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" && xff != clientIP {
				r.Header.Set("X-Forwarded-For", fmt.Sprintf("%s, %s", xff, clientIP))
			} else {
				r.Header.Set("X-Forwarded-For", clientIP)
			}
			r.Header.Set("X-Real-IP", clientIP)
		}

		// Set X-Forwarded-Proto header
		if r.TLS != nil {
			r.Header.Set("X-Forwarded-Proto", "https")
		} else {
			r.Header.Set("X-Forwarded-Proto", "http")
		}

		// Set X-Forwarded-Host header
		r.Header.Set("X-Forwarded-Host", request.Host)
	}

	// Handle ModifyResponse for security enhancements
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

		return nil
	}

	proxy.ServeHTTP(writer, request)
}

// HandleReverseProxy handles incoming requests and forwards them to the appropriate upstream
func HandleReverseProxy(w http.ResponseWriter, r *http.Request) {
	domain := getDomainWithoutPort(r.Host)
	
	if applicationLookup == nil {
		logs.Error("Application lookup function not set")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Lookup the application by domain
	app := applicationLookup(domain)
	if app == nil {
		logs.Info("No application found for domain: %s", domain)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Check if the application has an upstream host configured
	if app.UpstreamHost == "" {
		logs.Warn("Application %s/%s has no upstream host configured", app.Owner, app.Name)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Build the target URL
	targetUrl := app.UpstreamHost
	if !strings.HasPrefix(targetUrl, "http://") && !strings.HasPrefix(targetUrl, "https://") {
		targetUrl = "http://" + targetUrl
	}

	// Append the request URI to the target URL
	if !strings.HasSuffix(targetUrl, "/") && strings.HasPrefix(r.RequestURI, "/") {
		targetUrl = targetUrl + r.RequestURI
	} else if strings.HasSuffix(targetUrl, "/") && strings.HasPrefix(r.RequestURI, "/") {
		targetUrl = targetUrl + r.RequestURI[1:]
	} else if !strings.HasSuffix(targetUrl, "/") && !strings.HasPrefix(r.RequestURI, "/") {
		targetUrl = targetUrl + "/" + r.RequestURI
	} else {
		targetUrl = targetUrl + r.RequestURI
	}

	logs.Debug("Forwarding request from %s to %s", r.Host+r.RequestURI, targetUrl)
	forwardHandler(targetUrl, w, r)
}

// StartProxyServer starts the HTTP and HTTPS proxy servers based on configuration
func StartProxyServer() {
	proxyHttpPort := conf.GetConfigString("proxyHttpPort")
	proxyHttpsPort := conf.GetConfigString("proxyHttpsPort")

	if proxyHttpPort == "" && proxyHttpsPort == "" {
		logs.Info("Reverse proxy not enabled (proxyHttpPort and proxyHttpsPort are empty)")
		return
	}

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", HandleReverseProxy)

	// Start HTTP proxy if configured
	if proxyHttpPort != "" {
		go func() {
			addr := fmt.Sprintf(":%s", proxyHttpPort)
			logs.Info("Starting reverse proxy HTTP server on %s", addr)
			err := http.ListenAndServe(addr, serverMux)
			if err != nil {
				logs.Error("Failed to start HTTP proxy server: %v", err)
			}
		}()
	}

	// Start HTTPS proxy if configured
	if proxyHttpsPort != "" {
		go func() {
			addr := fmt.Sprintf(":%s", proxyHttpsPort)
			
			// For now, HTTPS will need certificate configuration
			// This can be enhanced later to use Application's SslCert field
			logs.Info("HTTPS proxy server on %s requires certificate configuration - not implemented yet", addr)
			
			// When implemented, use code like:
			// server := &http.Server{
			// 	Handler: serverMux,
			// 	Addr:    addr,
			// 	TLSConfig: &tls.Config{
			// 		MinVersion:               tls.VersionTLS12,
			// 		PreferServerCipherSuites: true,
			// 		CipherSuites: []uint16{
			// 			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			// 			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			// 			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			// 			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			// 			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			// 			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			// 		},
			// 		CurvePreferences: []tls.CurveID{
			// 			tls.X25519,
			// 			tls.CurveP256,
			// 			tls.CurveP384,
			// 		},
			// 	},
			// }
			// err := server.ListenAndServeTLS("", "")
			// if err != nil {
			// 	logs.Error("Failed to start HTTPS proxy server: %v", err)
			// }
		}()
	}
}
