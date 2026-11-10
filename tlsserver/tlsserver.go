// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

// Package tlsserver serves Casdoor on the extra port of the enterprise edition for certificate (UKey) sign-in: a TLS
// port asking browsers for a client certificate
package tlsserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

const ListenerCertLogin = "Certificate login"

// ListenerStatus tells a global admin whether an extra port is running, and why not
type ListenerStatus struct {
	Name         string   `json:"name"`
	Port         int      `json:"port"`
	Running      bool     `json:"running"`
	Certificates []string `json:"certificates"`
	Error        string   `json:"error,omitempty"`
	Code         string   `json:"code,omitempty"`
}

var (
	statusMu sync.RWMutex
	statuses []*ListenerStatus
)

type connContextKey struct{}

// GetStatuses returns the extra ports configured in app.conf
func GetStatuses() []*ListenerStatus {
	statusMu.RLock()
	defer statusMu.RUnlock()
	res := []*ListenerStatus{}
	for _, status := range statuses {
		item := *status
		res = append(res, &item)
	}
	return res
}

func setStatus(status *ListenerStatus) {
	statusMu.Lock()
	defer statusMu.Unlock()
	for i, item := range statuses {
		if item.Name == status.Name {
			statuses[i] = status
			return
		}
	}
	statuses = append(statuses, status)
}

func getPort(key string) int {
	port, err := conf.GetConfigInt64(key)
	if err != nil || port <= 0 {
		return 0
	}
	return int(port)
}

// readKeyPair reads the certificate and private key that a config item points to: the ID of a Casdoor cert like
// "admin/cert-login", or the path of a PEM certificate file with the key file in the config item keyKey
func readKeyPair(certKey string, keyKey string) ([]byte, []byte, string, error) {
	value := strings.TrimSpace(conf.GetConfigString(certKey))
	if value == "" {
		return nil, nil, "", fmt.Errorf("%s is not set", certKey)
	}

	if info, err := os.Stat(value); err == nil && !info.IsDir() {
		keyFile := strings.TrimSpace(conf.GetConfigString(keyKey))
		if keyFile == "" {
			return nil, nil, "", fmt.Errorf("%s is a file, set %s to the file of its private key", certKey, keyKey)
		}
		certPem, err := os.ReadFile(value)
		if err != nil {
			return nil, nil, "", err
		}
		keyPem, err := os.ReadFile(keyFile)
		if err != nil {
			return nil, nil, "", err
		}
		return certPem, keyPem, value, nil
	}

	cert, err := object.GetCert(value)
	if err != nil {
		return nil, nil, "", err
	}
	if cert == nil {
		return nil, nil, "", fmt.Errorf("the cert: %s set in %s does not exist", value, certKey)
	}
	if cert.Certificate == "" || cert.PrivateKey == "" {
		return nil, nil, "", fmt.Errorf("the cert: %s has no certificate or private key", value)
	}
	return []byte(cert.Certificate), []byte(cert.PrivateKey), value, nil
}

func getCertLoginTlsConfig() (*tls.Config, []string, error) {
	certPem, keyPem, name, err := readKeyPair("certLoginCert", "certLoginKey")
	if err != nil {
		return nil, nil, err
	}
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return nil, nil, fmt.Errorf("the server certificate: %s is invalid: %w", name, err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// the chain is verified against the trusted CAs of the user's organization, which are only known after
		// the request names its application
		ClientAuth: tls.RequestClientCert,
		MinVersion: tls.VersionTLS12,
	}
	return config, []string{name}, nil
}

func newServer(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 30 * time.Second,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, connContextKey{}, c)
		},
	}
}

// waitForMainPort waits until beego listens on its port, as the routes are only ready to serve then
func waitForMainPort(port int) {
	for i := 0; i < 600; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func serve(status *ListenerStatus, listener net.Listener, mainPort int, middleware web.MiddleWare) {
	util.SafeGoroutine(func() {
		waitForMainPort(mainPort)
		var handler http.Handler = web.BeeApp.Handlers
		if middleware != nil {
			handler = middleware(handler)
		}
		logs.Info("%s port: listening on %d", status.Name, status.Port)
		err := newServer(handler).Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logs.Error("%s port: %v", status.Name, err)
			failed := *status
			failed.Running = false
			failed.Error = err.Error()
			setStatus(&failed)
		}
	})
}

func startListener(name string, portKey string, feature string, mainPort int, middleware web.MiddleWare, wrap func(net.Listener) (net.Listener, []string, error)) {
	port := getPort(portKey)
	if port == 0 {
		return
	}
	status := &ListenerStatus{Name: name, Port: port, Certificates: []string{}}

	if err := edition.Require(edition.Enterprise, feature); err != nil {
		status.Error = err.Error()
		setStatus(status)
		logs.Error("%s port %d is not started: %s", name, port, err.Error())
		return
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		status.Error = err.Error()
		setStatus(status)
		logs.Error("%s port %d is not started: %s", name, port, err.Error())
		return
	}
	wrapped, certificates, err := wrap(listener)
	if err != nil {
		_ = listener.Close()
		status.Error = err.Error()
		setStatus(status)
		logs.Error("%s port %d is not started: %s", name, port, err.Error())
		return
	}

	status.Running = true
	status.Certificates = certificates
	setStatus(status)
	serve(status, wrapped, mainPort, middleware)
}

// Start opens the certificate sign-in port (certLoginPort) when it is configured, serving the same routes as the
// main port. Without an enterprise edition license it is not opened, which is logged and shown to global admins on
// the system info page
func Start(mainPort int, middleware web.MiddleWare) {
	startListener(ListenerCertLogin, "certLoginPort", edition.FeatureCertLogin, mainPort, middleware, func(listener net.Listener) (net.Listener, []string, error) {
		config, certificates, err := getCertLoginTlsConfig()
		if err != nil {
			return nil, nil, err
		}
		return tls.NewListener(listener, config), certificates, nil
	})
}

// GetPeerCertificates returns the DER client certificates of a request on the certificate sign-in port, the client's
// own one first, and whether the request came through that port at all
func GetPeerCertificates(r *http.Request) ([][]byte, bool) {
	conn := r.Context().Value(connContextKey{})
	switch c := conn.(type) {
	case *tls.Conn:
		res := [][]byte{}
		for _, certificate := range c.ConnectionState().PeerCertificates {
			res = append(res, certificate.Raw)
		}
		return res, true
	}
	return nil, false
}

func isTrustedCertProxy(r *http.Request) bool {
	trustedProxies := strings.TrimSpace(conf.GetConfigString("certLoginTrustedProxies"))
	if trustedProxies == "" {
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return false
	}
	for _, item := range strings.Split(trustedProxies, ",") {
		item = strings.TrimSpace(item)
		if _, ipNet, err := net.ParseCIDR(item); err == nil {
			if ipNet.Contains(ip) {
				return true
			}
		} else if proxyIp := net.ParseIP(item); proxyIp != nil && proxyIp.Equal(ip) {
			return true
		}
	}
	return false
}

// GetProxiedCertificate returns the client certificate that a reverse proxy terminating TLS put in a header, nginx's
// $ssl_client_escaped_cert in X-SSL-Client-Cert by default (config certLoginHeader). The header is only read from the
// proxies in the config certLoginTrustedProxies
func GetProxiedCertificate(r *http.Request) string {
	if !isTrustedCertProxy(r) {
		return ""
	}
	header := strings.TrimSpace(conf.GetConfigString("certLoginHeader"))
	if header == "" {
		header = "X-SSL-Client-Cert"
	}
	value := strings.TrimSpace(r.Header.Get(header))
	if value == "" || value == "(null)" {
		return ""
	}
	if strings.Contains(value, "%") {
		if unescaped, err := url.QueryUnescape(value); err == nil {
			value = unescaped
		}
	} else if strings.Contains(value, "-----BEGIN") && !strings.Contains(value, "\n") {
		// a PEM put in a header with its line breaks replaced by spaces or tabs
		value = strings.NewReplacer("\t", "\n").Replace(value)
		begin := strings.Index(value, "-----BEGIN CERTIFICATE-----")
		end := strings.Index(value, "-----END CERTIFICATE-----")
		if begin >= 0 && end > begin {
			body := strings.ReplaceAll(value[begin+len("-----BEGIN CERTIFICATE-----"):end], " ", "\n")
			value = "-----BEGIN CERTIFICATE-----" + body + "\n-----END CERTIFICATE-----\n"
		}
	}
	return value
}
