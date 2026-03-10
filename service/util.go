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
	"strings"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
)

func joinPath(a string, b string) string {
	if strings.HasSuffix(a, "/") && strings.HasPrefix(b, "/") {
		b = b[1:]
	} else if !strings.HasSuffix(a, "/") && !strings.HasPrefix(b, "/") {
		b = "/" + b
	}
	res := a + b
	return res
}

func isHostIp(host string) bool {
	hostWithoutPort := strings.Split(host, ":")[0]
	ip := net.ParseIP(hostWithoutPort)
	return ip != nil
}

func responseOk(w http.ResponseWriter, format string, a ...interface{}) {
	w.WriteHeader(http.StatusOK)

	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
	_, err := fmt.Fprint(w, msg)
	if err != nil {
		panic(err)
	}
}

func responseError(w http.ResponseWriter, format string, a ...interface{}) {
	w.WriteHeader(http.StatusInternalServerError)

	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
	_, err := fmt.Fprint(w, msg)
	if err != nil {
		panic(err)
	}
}

func responseErrorWithoutCode(w http.ResponseWriter, format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	fmt.Println(msg)
	_, err := fmt.Fprint(w, msg)
	if err != nil {
		panic(err)
	}
}

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

func getSiteByDomainWithWww(domain string) *object.Site {
	hostNonWww := getHostNonWww(domain)
	if hostNonWww != "" {
		domain = hostNonWww
	}

	domainWithoutPort := getDomainWithoutPort(domain)

	site := object.GetSiteByDomain(domainWithoutPort)
	return site
}

func getX509CertByDomain(domain string) (*tls.Certificate, error) {
	cert, err := object.GetCertByDomain(domain)
	if err != nil {
		return nil, fmt.Errorf("getX509CertByDomain() error: %v, domain: [%s]", err, domain)
	}
	if cert == nil {
		return nil, fmt.Errorf("getX509CertByDomain() error: cert not found for domain: [%s]", domain)
	}

	tlsCert, certErr := tls.X509KeyPair([]byte(cert.Certificate), []byte(cert.PrivateKey))

	return &tlsCert, certErr
}

func getCasdoorClientFromSite(site *object.Site) (*casdoorsdk.Client, error) {
	if site.ApplicationObj == nil {
		return nil, fmt.Errorf("site.ApplicationObj is empty")
	}

	casdoorEndpoint := conf.GetConfigString("origin")
	if casdoorEndpoint == "" {
		casdoorEndpoint = "http://localhost:8000"
	}

	clientId := site.ApplicationObj.ClientId
	clientSecret := site.ApplicationObj.ClientSecret

	certificate := ""
	if site.ApplicationObj.CertObj != nil {
		certificate = site.ApplicationObj.CertObj.Certificate
	}

	res := casdoorsdk.NewClient(casdoorEndpoint, clientId, clientSecret, certificate, site.ApplicationObj.Organization, site.CasdoorApplication)
	return res, nil
}

func getScheme(r *http.Request) string {
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}
	return scheme
}
