// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/object"
)

// mergePaidIntegrations adds the paid integrations of the Marketplace server to the index from the CDN. With a
// license, their versions carry the checksums of the files written for it.
func mergePaidIntegrations(index *object.MarketplaceIndex) {
	var err error
	for _, serverUrl := range object.GetMarketplaceServerUrls() {
		var data []byte
		data, err = object.FetchMarketplaceFileWithLicense(serverUrl+"/paid-index.json", edition.GetLicenseText() != "")
		if err != nil {
			continue
		}

		paid := &object.MarketplaceIndex{}
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
			item.Logo = object.JoinMarketplaceUrl(serverUrl, item.Logo)
			for i, screenshot := range item.Screenshots {
				item.Screenshots[i] = object.JoinMarketplaceUrl(serverUrl, screenshot)
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
	item, err := object.GetMarketplaceIntegration(id)
	if err != nil {
		return nil, err
	}
	if item.Edition != "" {
		if err = edition.Require(item.Edition, edition.FeaturePaidIntegrations); err != nil {
			return nil, fmt.Errorf("%s: %w", item.Id, err)
		}
	}

	bundle, _, checkedVersion, err := object.GetIntegrationBundle(item, "", true)
	if err == nil {
		err = fmt.Errorf("the integration: %s has no download %s", id, path)
		for _, download := range bundle.Manifest.Downloads {
			if download == path {
				var data []byte
				data, err = bundle.Fetch(path)
				if err == nil {
					if checkedVersion != "" {
						object.ReadMarketplaceVersionMetaStep(item.Id, checkedVersion, item.Edition != "", object.MarketplaceStepInstalled, "")
					}
					return data, nil
				}
				break
			}
		}
	}
	if checkedVersion != "" {
		stage := object.InstallStageDownload
		if errors.As(err, &object.IntegrationRevokedError{}) {
			stage = object.InstallStageRevoked
		}
		object.ReadMarketplaceVersionMetaStep(item.Id, checkedVersion, item.Edition != "", object.MarketplaceStepFailed, stage)
	}
	return nil, err
}
