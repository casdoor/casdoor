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

	"github.com/casbin/lego/v4/certificate"
)

type HttpProvider struct {
	siteId string
}

func (p *HttpProvider) Present(domain string, token string, keyAuth string) error {
	site, err := GetSite(p.siteId)
	if err != nil {
		return err
	}

	site.Challenges = []string{fmt.Sprintf("%s:%s", token, keyAuth)}
	_, err = UpdateSiteNoRefresh(site.GetId(), site)
	if err != nil {
		return err
	}

	err = refreshSiteMap()
	if err != nil {
		return err
	}

	return nil
}

func (p *HttpProvider) CleanUp(domain string, token string, keyAuth string) error {
	site, err := GetSite(p.siteId)
	if err != nil {
		return err
	}

	site.Challenges = []string{}
	_, err = UpdateSiteNoRefresh(site.GetId(), site)
	if err != nil {
		return err
	}

	err = refreshSiteMap()
	if err != nil {
		return err
	}

	return nil
}

func getHttp01Cert(siteId string, domain string) (string, string, error) {
	client, err := GetAcmeClient(false)
	if err != nil {
		return "", "", err
	}

	provider := HttpProvider{siteId: siteId}
	err = client.Challenge.SetHTTP01Provider(&provider)
	if err != nil {
		return "", "", err
	}

	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}

	resource, err := client.Certificate.Obtain(request)
	if err != nil {
		return "", "", err
	}

	return string(resource.Certificate), string(resource.PrivateKey), nil
}
