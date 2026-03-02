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

package object

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

func (site *Site) populateCert() error {
	if site.Domain == "" {
		return nil
	}

	cert, err := GetCertByDomain(site.Domain)
	if err != nil {
		return err
	}
	if cert == nil {
		return nil
	}

	site.SslCert = cert.Name
	return nil
}

func checkUrlToken(url string, keyAuth string) (bool, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(string(body)) == keyAuth {
		return true, nil
	}

	return false, fmt.Errorf("checkUrlToken() error, response mismatch: expected %q, got %q", keyAuth, body)
}

func (site *Site) preCheckCertForDomain(domain string) (bool, error) {
	token, keyAuth, err := util.GenerateTwoUniqueRandomStrings()
	if err != nil {
		return false, err
	}

	site.Challenges = []string{fmt.Sprintf("%s:%s", token, keyAuth)}

	_, err = UpdateSiteNoRefresh(site.GetId(), site)
	if err != nil {
		return false, err
	}

	err = refreshSiteMap()
	if err != nil {
		return false, err
	}

	url := fmt.Sprintf("http://%s/.well-known/acme-challenge/%s", domain, token)
	var ok bool
	for i := 0; i < 10; i++ {
		fmt.Printf("checkUrlToken(): try time: %d\n", i+1)
		ok, err = checkUrlToken(url, keyAuth)
		if err != nil {
			fmt.Printf("preCheckCertForDomain() error: %v\n", err)
			time.Sleep(time.Second)
		}
		if ok {
			fmt.Printf("checkUrlToken(): try time: %d, succeed!\n", i+1)
			break
		}
	}

	site.Challenges = []string{}
	_, err = UpdateSiteNoRefresh(site.GetId(), site)
	if err != nil {
		return false, err
	}

	err = refreshSiteMap()
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (site *Site) updateCertForDomain(domain string) error {
	ok, err := site.preCheckCertForDomain(domain)
	if err != nil {
		return err
	}
	if !ok {
		fmt.Printf("preCheckCertForDomain(): not ok for domain: %s\n", domain)
		return nil
	}

	certificate, privateKey, err := getHttp01Cert(site.GetId(), domain)
	if err != nil {
		return err
	}

	expireTime, err := util.GetCertExpireTime(certificate)
	if err != nil {
		fmt.Printf("getCertExpireTime() error: %v\n", err)
	}

	domainExpireTime, err := getDomainExpireTime(domain)
	if err != nil {
		fmt.Printf("getDomainExpireTime() error: %v\n", err)
	}

	cert := Cert{
		Owner:            site.Owner,
		Name:             domain,
		CreatedTime:      util.GetCurrentTime(),
		DisplayName:      domain,
		Type:             "SSL",
		CryptoAlgorithm:  "RSA",
		ExpireTime:       expireTime,
		DomainExpireTime: domainExpireTime,
		Provider:         "",
		Account:          "",
		AccessKey:        "",
		AccessSecret:     "",
		Certificate:      certificate,
		PrivateKey:       privateKey,
	}

	_, err = DeleteCert(&cert)
	if err != nil {
		return err
	}

	_, err = AddCert(&cert)
	if err != nil {
		return err
	}

	err = refreshSiteMap()
	if err != nil {
		return err
	}

	return nil
}

func (site *Site) checkCerts() error {
	domains := []string{}
	if site.Domain != "" {
		domains = append(domains, site.Domain)
	}

	for _, domain := range site.OtherDomains {
		domains = append(domains, domain)
	}

	for _, domain := range domains {
		if site.Owner == "admin" || strings.HasSuffix(domain, ".casdoor.com") {
			continue
		}

		cert, err := GetCertByDomain(domain)
		if err != nil {
			return err
		}

		if cert != nil {
			var nearExpire bool
			nearExpire, err = cert.isCertNearExpire()
			if err != nil {
				return err
			}

			if !nearExpire {
				continue
			}
		}

		err = site.updateCertForDomain(domain)
		if err != nil {
			return err
		}
	}

	return nil
}
