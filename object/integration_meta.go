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
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/edition"
)

// the steps after the check before installing at which Casdoor reads the meta of the version again
const (
	marketplaceStepInstalled   = "installed"
	marketplaceStepFailed      = "failed"
	marketplaceStepUninstalled = "uninstalled"
	// an admin opened the integration in the Marketplace page
	marketplaceStepView = "view"
)

// The Marketplace page reads the server's status of the catalog and of the integration opened with a short timeout,
// and after the server could not be reached, not again for a while, so a Casdoor without Internet access does not
// wait for it on every page.
const (
	marketplacePageTimeout       = 5 * time.Second
	marketplaceServerRetryPeriod = 10 * time.Minute
)

var (
	marketplaceServerMutex sync.Mutex
	marketplaceServerDown  time.Time
	// the versions the server says are withdrawn, as of the last time the Marketplace page was opened
	marketplaceServerRevoked []*MarketplaceRevoked
)

func isMarketplaceServerDown() bool {
	marketplaceServerMutex.Lock()
	defer marketplaceServerMutex.Unlock()
	return time.Since(marketplaceServerDown) < marketplaceServerRetryPeriod
}

// setMarketplaceServerDown remembers that the server could not be reached, when err, the last error, is that of the
// connection rather than an answer such as 404
func setMarketplaceServerDown(err error) {
	urlErr := &url.Error{}
	if !errors.As(err, &urlErr) {
		return
	}
	marketplaceServerMutex.Lock()
	defer marketplaceServerMutex.Unlock()
	marketplaceServerDown = time.Now()
}

func getMarketplaceServerRevoked() []*MarketplaceRevoked {
	marketplaceServerMutex.Lock()
	defer marketplaceServerMutex.Unlock()
	return marketplaceServerRevoked
}

// marketplaceStatus is what the Marketplace server says about the catalog right now
type marketplaceStatus struct {
	Revoked []*MarketplaceRevoked `json:"revoked"`
}

// refreshMarketplaceStatus reads status.json from the Marketplace server each time an admin opens the Marketplace
// page: the versions withdrawn right now, which the catalog from the CDN may not have yet. Like meta.json, the
// address carries nothing about this Casdoor.
func refreshMarketplaceStatus() {
	if conf.GetConfigString("marketplaceUrl") != "" || isMarketplaceServerDown() {
		return
	}
	var err error
	for _, serverUrl := range getMarketplaceServerUrls() {
		var data []byte
		data, err = fetchMarketplaceFileWithTimeout(serverUrl+"/status.json", false, marketplacePageTimeout)
		if err != nil {
			continue
		}
		status := &marketplaceStatus{}
		err = json.Unmarshal(data, status)
		if err != nil {
			continue
		}
		marketplaceServerMutex.Lock()
		marketplaceServerRevoked = status.Revoked
		marketplaceServerMutex.Unlock()
		return
	}
	setMarketplaceServerDown(err)
}

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
// server could be reached (e.g. an intranet Casdoor), in which case the CDN catalog is all there is to check against.
// For an admin who opens the integration in the Marketplace page, step is "view".
func getMarketplaceVersionMeta(item *MarketplaceIntegration, version *MarketplaceVersion, step string) (*MarketplaceVersionMeta, string) {
	path := metaPath(version.Url)
	timeout := 30 * time.Second
	if step != "" {
		if isMarketplaceServerDown() {
			return nil, ""
		}
		path += "?" + url.Values{"step": {step}}.Encode()
		timeout = marketplacePageTimeout
	}

	var err error
	for _, serverUrl := range getMarketplaceServerUrls() {
		var data []byte
		// a paid integration opened without a license gets the meta without its checksum
		data, err = fetchMarketplaceFileWithTimeout(joinMarketplaceUrl(serverUrl, path), item.Edition != "" && edition.GetLicenseText() != "", timeout)
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
	if step != "" {
		setMarketplaceServerDown(err)
		return nil, ""
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
