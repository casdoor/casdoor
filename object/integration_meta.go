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
	"net/url"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
)

// the steps after the check before installing at which Casdoor reads the meta of the version again
const (
	marketplaceStepInstalled   = "installed"
	marketplaceStepFailed      = "failed"
	marketplaceStepUninstalled = "uninstalled"
)

// the stages an install fails at
const (
	installStageRevoked   = "revoked"
	installStageDownload  = "download"
	installStageVariables = "variables"
	installStageCreate    = "create"
	installStageSave      = "save"
)

// integrationRevokedError is an install the Marketplace server refused because the version has been withdrawn
type integrationRevokedError struct {
	error
}

// readMarketplaceVersionMetaStep reads the meta of a version again when its install has ended or it is uninstalled,
// so the Marketplace knows how installs end. As with the check before installing, the address has only the
// integration, the version and the step. It runs in the background and nothing waits for it.
func readMarketplaceVersionMetaStep(id string, version string, withLicense bool, step string, stage string) {
	query := url.Values{"step": {step}}
	if stage != "" {
		query.Set("stage", stage)
	}
	path := fmt.Sprintf("integrations/%s/%s/meta.json?%s", url.PathEscape(id), url.PathEscape(version), query.Encode())
	go func() {
		for _, serverUrl := range getMarketplaceServerUrls() {
			_, err := fetchMarketplaceFileWithLicense(joinMarketplaceUrl(serverUrl, path), withLicense)
			if err == nil {
				return
			}
		}
	}()
}

// MarketplaceVersionMeta is what the Marketplace server says about a version right now. The catalog and bundles come
// from the CDN, whose copy may be days old, so before installing, Casdoor checks with the server that the version
// has not been withdrawn and that the bundle has the checksum the server has for it.
type MarketplaceVersionMeta struct {
	Id      string `json:"id"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Sha256  string `json:"sha256"`
}

// isOfficialMarketplace tells whether the catalog is the official one, which the Marketplace server knows about,
// rather than a catalog of one's own set in marketplaceUrl
func isOfficialMarketplace(item *MarketplaceIntegration) bool {
	return item.Edition != "" || conf.GetConfigString("marketplaceUrl") == ""
}

// getMarketplaceVersionMeta returns the meta of a version and the Marketplace server it came from, or nil when no
// server could be reached (e.g. an intranet Casdoor), in which case the CDN catalog is all there is to check against
func getMarketplaceVersionMeta(item *MarketplaceIntegration, version *MarketplaceVersion) (*MarketplaceVersionMeta, string) {
	var err error
	for _, serverUrl := range getMarketplaceServerUrls() {
		var data []byte
		data, err = fetchMarketplaceFileWithLicense(joinMarketplaceUrl(serverUrl, metaPath(version.Url)), item.Edition != "")
		if err != nil {
			continue
		}
		meta := &MarketplaceVersionMeta{}
		err = json.Unmarshal(data, meta)
		if err != nil {
			continue
		}
		if meta.Id != item.Id || meta.Version != version.Version {
			err = fmt.Errorf("the Marketplace answered for %s %s", meta.Id, meta.Version)
			continue
		}
		return meta, serverUrl
	}
	logs.Warn("the Marketplace server could not check %s %s, using the CDN catalog only: %v", item.Id, version.Version, err)
	return nil, ""
}

// metaPath turns the path of a bundle, integrations/<id>/<version>/bundle.json, into that of its meta.json
func metaPath(bundlePath string) string {
	for i := len(bundlePath) - 1; i >= 0; i-- {
		if bundlePath[i] == '/' {
			return bundlePath[:i+1] + "meta.json"
		}
	}
	return "meta.json"
}
