// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

package deployment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/storage"
	"github.com/casdoor/casdoor/util"
	"github.com/casdoor/oss"
)

func deployStaticFiles(provider *object.Provider) {
	certificate := ""
	if provider.Category == "Storage" && provider.Type == "Casdoor" {
		cert, err := object.GetCert(util.GetId(provider.Owner, provider.Cert))
		if err != nil {
			panic(err)
		}
		if cert == nil {
			panic(err)
		}
		certificate = cert.Certificate
	}
	storageProvider, err := storage.GetStorageProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.RegionId, provider.Bucket, provider.Endpoint, certificate, provider.Content)
	if err != nil {
		panic(err)
	}
	if storageProvider == nil {
		panic(fmt.Sprintf("the provider type: %s is not supported", provider.Type))
	}

	// Vite emits every hashed chunk and stylesheet into one "assets" directory,
	// where create-react-app used to split them into static/js and static/css.
	uploadFolder(storageProvider)
	updateHtml(provider.Domain)
}

const buildAssetsDir = "assets"

func uploadFolder(storageProvider oss.StorageInterface) {
	path := fmt.Sprintf("../web/build/%s/", buildAssetsDir)
	filenames := util.ListFiles(path)

	for _, filename := range filenames {
		file, err := os.Open(filepath.Clean(path + filename))
		if err != nil {
			panic(err)
		}

		objectKey := fmt.Sprintf("%s/%s", buildAssetsDir, filename)
		_, err = storageProvider.Put(objectKey, file)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Uploaded [%s] to [%s]\n", path+filename, objectKey)
	}
}

func updateHtml(domainPath string) {
	htmlPath := "../web/build/index.html"
	html := util.ReadStringFromPath(htmlPath)
	html = strings.Replace(html, fmt.Sprintf("\"/%s/", buildAssetsDir), fmt.Sprintf("\"%s%s/", domainPath, buildAssetsDir), -1)
	util.WriteStringToPath(html, htmlPath)

	fmt.Printf("Updated HTML to [%s]\n", html)
}
