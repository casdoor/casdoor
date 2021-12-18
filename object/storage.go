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

	"github.com/astaxie/beego"
	"github.com/casbin/casdoor/storage"
	"github.com/casbin/casdoor/util"
)

var isCloudIntranet bool

func init() {
	var err error
	isCloudIntranet, err = beego.AppConfig.Bool("isCloudIntranet")
	if err != nil {
		//panic(err)
	}
}

func getProviderEndpoint(provider *Provider) string {
	endpoint := provider.Endpoint
	if provider.IntranetEndpoint != "" && isCloudIntranet {
		endpoint = provider.IntranetEndpoint
	}
	return endpoint
}

func getUploadFileUrl(provider *Provider, fullFilePath string, hasTimestamp bool) (string, string) {
	objectKey := util.UrlJoin(util.GetUrlPath(provider.Domain), fullFilePath)

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

	fileUrl := util.UrlJoin(host, objectKey)
	if hasTimestamp {
		fileUrl = fmt.Sprintf("%s?t=%s", util.UrlJoin(host, objectKey), util.GetCurrentUnixTime())
	}

	return fileUrl, objectKey
}

func uploadFile(provider *Provider, fullFilePath string, fileBuffer *bytes.Buffer) (string, string, error) {
	endpoint := getProviderEndpoint(provider)
	storageProvider := storage.GetStorageProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.RegionId, provider.Bucket, endpoint)
	if storageProvider == nil {
		return "", "", fmt.Errorf("the provider type: %s is not supported", provider.Type)
	}

	if provider.Domain == "" {
		provider.Domain = storageProvider.GetEndpoint()
		UpdateProvider(provider.GetId(), provider)
	}

	fileUrl, objectKey := getUploadFileUrl(provider, fullFilePath, true)

	_, err := storageProvider.Put(objectKey, fileBuffer)
	if err != nil {
		return "", "", err
	}

	return fileUrl, objectKey, nil
}

func UploadFileSafe(provider *Provider, fullFilePath string, fileBuffer *bytes.Buffer) (string, string, error) {
	var fileUrl string
	var objectKey string
	var err error
	times := 0
	for {
		fileUrl, objectKey, err = uploadFile(provider, fullFilePath, fileBuffer)
		if err != nil {
			times += 1
			if times >= 5 {
				return "", "", err
			}
		} else {
			break
		}
	}
	return fileUrl, objectKey, nil
}

func DeleteFile(provider *Provider, objectKey string) error {
	endpoint := getProviderEndpoint(provider)
	storageProvider := storage.GetStorageProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.RegionId, provider.Bucket, endpoint)
	if storageProvider == nil {
		return fmt.Errorf("the provider type: %s is not supported", provider.Type)
	}

	if provider.Domain == "" {
		provider.Domain = storageProvider.GetEndpoint()
		UpdateProvider(provider.GetId(), provider)
	}

	return storageProvider.Delete(objectKey)
}
