// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"bytes"
	"fmt"
	"strings"

	"github.com/casbin/casdoor/storage"
	"github.com/casbin/casdoor/util"
)

func UploadFile(provider *Provider, fullFilePath string, fileBuffer *bytes.Buffer) (string, error) {
	storageProvider := storage.GetStorageProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.RegionId, provider.Bucket, provider.Endpoint)
	if storageProvider == nil {
		return "", fmt.Errorf("the provider type: %s is not supported", provider.Type)
	}

	if provider.Domain == "" {
		provider.Domain = storageProvider.GetEndpoint()
		UpdateProvider(provider.GetId(), provider)
	}

	path := util.UrlJoin(util.GetUrlPath(provider.Domain), fullFilePath)
	_, err := storageProvider.Put(path, fileBuffer)
	if err != nil {
		return "", err
	}

	host := ""
	if provider.Type != "Local File System" {
		// provider.Domain = "https://cdn.casbin.com/casdoor/"
		host = util.GetUrlHost(provider.Domain)
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = fmt.Sprintf("https://%s", host)
		}
	} else {
		// provider.Domain = "http://localhost:8000" or "https://door.casbin.com"
		host = util.UrlJoin(provider.Domain, "/files")
	}

	fileUrl := fmt.Sprintf("%s?time=%s", util.UrlJoin(host, path), util.GetCurrentUnixTime())
	return fileUrl, nil
}
