// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"net/url"
	"path/filepath"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"github.com/casdoor/casdoor/storage"
	"github.com/casdoor/casdoor/util"
	"github.com/casdoor/oss"
)

var isCloudIntranet bool

const (
	ProviderTypeGoogleCloudStorage = "Google Cloud Storage"
	ProviderTypeTencentCloudCOS    = "Tencent Cloud COS"
	ProviderTypeAzureBlob          = "Azure Blob"
	ProviderTypeLocalFileSystem    = "Local File System"
)

func init() {
	isCloudIntranet = conf.GetConfigBool("isCloudIntranet")
}

func getProviderEndpoint(provider *Provider) string {
	endpoint := provider.Endpoint
	if provider.IntranetEndpoint != "" && isCloudIntranet {
		endpoint = provider.IntranetEndpoint
	}
	return endpoint
}

func escapePath(path string) string {
	tokens := strings.Split(path, "/")
	if len(tokens) > 0 {
		tokens[len(tokens)-1] = url.QueryEscape(tokens[len(tokens)-1])
	}

	res := strings.Join(tokens, "/")
	return res
}

func GetTruncatedPath(provider *Provider, fullFilePath string, limit int) string {
	pathPrefix := util.UrlJoin(util.GetUrlPath(provider.Domain), provider.PathPrefix)

	dir, file := filepath.Split(fullFilePath)
	ext := filepath.Ext(file)
	fileName := strings.TrimSuffix(file, ext)
	for {
		escapedString := escapePath(escapePath(fullFilePath))
		if len(escapedString) < limit-len(pathPrefix) {
			break
		}
		rs := []rune(fileName)
		fileName = string(rs[0 : len(rs)-1])
		fullFilePath = dir + fileName + ext
	}

	return fullFilePath
}

func GetUploadFileUrl(provider *Provider, fullFilePath string, hasTimestamp bool) (string, string) {
	if provider.Domain != "" && !strings.HasPrefix(provider.Domain, "http://") && !strings.HasPrefix(provider.Domain, "https://") {
		provider.Domain = fmt.Sprintf("https://%s", provider.Domain)
	}

	escapedPath := util.UrlJoin(provider.PathPrefix, fullFilePath)
	objectKey := util.UrlJoin(util.GetUrlPath(provider.Domain), escapedPath)

	host := ""
	if provider.Type != ProviderTypeLocalFileSystem {
		// provider.Domain = "https://cdn.casbin.com/casdoor/"
		host = util.GetUrlHost(provider.Domain)
	} else {
		// provider.Domain = "http://localhost:8000" or "https://door.casdoor.com"
		host = util.UrlJoin(provider.Domain, "/files")
	}
	if provider.Type == ProviderTypeAzureBlob || provider.Type == ProviderTypeGoogleCloudStorage {
		host = util.UrlJoin(host, provider.Bucket)
	}

	fileUrl := ""
	if host != "" {
		// fileUrl = util.UrlJoin(host, escapePath(objectKey))
		fileUrl = util.UrlJoin(host, objectKey)
	}

	// if fileUrl != "" && hasTimestamp {
	//	fileUrl = fmt.Sprintf("%s?t=%s", fileUrl, util.GetCurrentUnixTime())
	// }

	if provider.Type == ProviderTypeTencentCloudCOS {
		objectKey = escapePath(objectKey)
	}

	return fileUrl, objectKey
}

func getStorageProvider(provider *Provider, lang string) (oss.StorageInterface, error) {
	endpoint := getProviderEndpoint(provider)
	certificate := ""
	if provider.Category == "Storage" && provider.Type == "Casdoor" {
		cert, err := GetCert(util.GetId(provider.Owner, provider.Cert))
		if err != nil {
			return nil, err
		}
		if cert == nil {
			return nil, fmt.Errorf("no cert for %s", provider.Cert)
		}
		certificate = cert.Certificate
	}
	storageProvider, err := storage.GetStorageProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.RegionId, provider.Bucket, endpoint, certificate, provider.Content)
	if err != nil {
		return nil, err
	}
	if storageProvider == nil {
		return nil, fmt.Errorf(i18n.Translate(lang, "storage:The provider type: %s is not supported"), provider.Type)
	}

	if provider.Domain == "" {
		provider.Domain = storageProvider.GetEndpoint()
		_, err = UpdateProvider(provider.GetId(), provider)
		if err != nil {
			return nil, err
		}
	}

	return storageProvider, nil
}

func uploadFile(provider *Provider, fullFilePath string, fileBuffer *bytes.Buffer, lang string) (string, string, error) {
	storageProvider, err := getStorageProvider(provider, lang)
	if err != nil {
		return "", "", err
	}

	fileUrl, objectKey := GetUploadFileUrl(provider, fullFilePath, true)
	objectKeyRefined := refineObjectKey(provider, objectKey)

	object, err := storageProvider.Put(objectKeyRefined, fileBuffer)
	if err != nil {
		return "", "", err
	}

	if provider.Type == "Casdoor" {
		fileUrl = object.Path
	}

	return fileUrl, objectKey, nil
}

func UploadFileSafe(provider *Provider, fullFilePath string, fileBuffer *bytes.Buffer, lang string) (string, string, error) {
	// check fullFilePath is there security issue
	if strings.Contains(fullFilePath, "..") {
		return "", "", fmt.Errorf("the fullFilePath: %s is not allowed", fullFilePath)
	}

	var fileUrl string
	var objectKey string
	var err error
	times := 0
	for {
		fileUrl, objectKey, err = uploadFile(provider, fullFilePath, fileBuffer, lang)
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

func DeleteFile(provider *Provider, objectKey string, lang string) error {
	// check fullFilePath is there security issue
	if strings.Contains(objectKey, "..") {
		return fmt.Errorf(i18n.Translate(lang, "storage:The objectKey: %s is not allowed"), objectKey)
	}

	storageProvider, err := getStorageProvider(provider, lang)
	if err != nil {
		return err
	}

	objectKeyRefined := refineObjectKey(provider, objectKey)
	return storageProvider.Delete(objectKeyRefined)
}

func refineObjectKey(provider *Provider, objectKey string) string {
	if provider.Type == ProviderTypeGoogleCloudStorage {
		return strings.TrimPrefix(objectKey, "/")
	}
	return objectKey
}
