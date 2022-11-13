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
	"io"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/proxy"
)

var defaultStorageProvider *Provider = nil

func InitDefaultStorageProvider() {
	defaultStorageProviderStr := conf.GetConfigString("defaultStorageProvider")
	if defaultStorageProviderStr != "" {
		defaultStorageProvider = getProvider("admin", defaultStorageProviderStr)
	}
}

func downloadFile(url string) (*bytes.Buffer, error) {
	httpClient := proxy.GetHttpClient(url)

	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fileBuffer := bytes.NewBuffer(nil)
	_, err = io.Copy(fileBuffer, resp.Body)
	if err != nil {
		return nil, err
	}

	return fileBuffer, nil
}

func getPermanentAvatarUrl(organization string, username string, url string, upload bool) string {
	if url == "" {
		return ""
	}

	if defaultStorageProvider == nil {
		return ""
	}

	fullFilePath := fmt.Sprintf("/avatar/%s/%s.png", organization, username)
	uploadedFileUrl, _ := getUploadFileUrl(defaultStorageProvider, fullFilePath, false)

	if upload {
		DownloadAndUpload(url, fullFilePath)
	}

	return uploadedFileUrl
}

func DownloadAndUpload(url string, fullFilePath string) {
	fileBuffer, err := downloadFile(url)
	if err != nil {
		panic(err)
	}

	_, _, err = UploadFileSafe(defaultStorageProvider, fullFilePath, fileBuffer)
	if err != nil {
		panic(err)
	}
}
