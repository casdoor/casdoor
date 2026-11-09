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

package object

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/edition"
)

// The paid integrations aren't on the CDN: the Marketplace server lists them in paid-index.json and gives their
// files only to a Casdoor that sends a commercial or enterprise license, written into the files it gives.
var defaultMarketplaceServerUrls = []string{"https://marketplace.casdoor.com/data", "https://marketplace.casdoor.cn/data"}

const licenseHeader = "X-Casdoor-License"

func getMarketplaceServerUrls() []string {
	serverUrl := conf.GetConfigString("marketplaceServerUrl")
	if serverUrl == "" {
		return defaultMarketplaceServerUrls
	}
	return []string{strings.TrimSuffix(serverUrl, "/")}
}

func fetchMarketplaceFile(fileUrl string) ([]byte, error) {
	return fetchMarketplaceFileWithLicense(fileUrl, false)
}

func fetchMarketplaceFileWithLicense(fileUrl string, withLicense bool) ([]byte, error) {
	req, err := http.NewRequest("GET", fileUrl, nil)
	if err != nil {
		return nil, err
	}
	if withLicense {
		license := edition.GetLicenseText()
		if license == "" {
			return nil, edition.Require(edition.Commercial, edition.FeaturePaidIntegrations)
		}
		req.Header.Set(licenseHeader, license)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1000))
		if resp.StatusCode == http.StatusForbidden && len(body) > 0 {
			return nil, fmt.Errorf("%s", strings.TrimSpace(string(body)))
		}
		return nil, fmt.Errorf("failed to get %s, status code: %d", fileUrl, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 32<<20))
}

// mergePaidIntegrations adds the paid integrations of the Marketplace server to the index from the CDN. With a
// license, their versions carry the checksums of the files written for it.
func mergePaidIntegrations(index *MarketplaceIndex) {
	var err error
	for _, serverUrl := range getMarketplaceServerUrls() {
		var data []byte
		data, err = fetchMarketplaceFileWithLicense(serverUrl+"/paid-index.json", edition.GetLicenseText() != "")
		if err != nil {
			continue
		}

		paid := &MarketplaceIndex{}
		err = json.Unmarshal(data, paid)
		if err != nil {
			continue
		}

		existing := map[string]bool{}
		for _, item := range index.Integrations {
			existing[item.Id] = true
		}
		for _, item := range paid.Integrations {
			if existing[item.Id] || item.Edition == "" {
				continue
			}
			item.BaseUrl = serverUrl
			item.Logo = joinMarketplaceUrl(serverUrl, item.Logo)
			for i, screenshot := range item.Screenshots {
				item.Screenshots[i] = joinMarketplaceUrl(serverUrl, screenshot)
			}
			index.Integrations = append(index.Integrations, item)
		}
		index.Revoked = append(index.Revoked, paid.Revoked...)
		return
	}
	if err != nil {
		logs.Warn("the paid integrations of the Marketplace are not listed: %s", err.Error())
	}
}

// GetMarketplaceDownload downloads a file an integration lists in its downloads, such as the plugin of an API
// gateway, with the license when the integration is paid
func GetMarketplaceDownload(id string, path string) ([]byte, error) {
	item, err := GetMarketplaceIntegration(id)
	if err != nil {
		return nil, err
	}
	if item.Edition != "" {
		if err = edition.Require(item.Edition, edition.FeaturePaidIntegrations); err != nil {
			return nil, fmt.Errorf("%s: %w", item.Id, err)
		}
	}

	bundle, _, err := getIntegrationBundle(item, "")
	if err != nil {
		return nil, err
	}
	for _, download := range bundle.Manifest.Downloads {
		if download == path {
			return bundle.fetch(path)
		}
	}
	return nil, fmt.Errorf("the integration: %s has no download %s", id, path)
}
