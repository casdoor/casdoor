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
	storageProvider := storage.GetStorageProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.RegionId, provider.Bucket, provider.Endpoint)
	if storageProvider == nil {
		panic(fmt.Sprintf("the provider type: %s is not supported", provider.Type))
	}

	uploadFolder(storageProvider, "js")
	uploadFolder(storageProvider, "css")
	updateHtml(provider.Domain)
}

func uploadFolder(storageProvider oss.StorageInterface, folder string) {
	path := fmt.Sprintf("../web/build/static/%s/", folder)
	filenames := util.ListFiles(path)

	for _, filename := range filenames {
		if !strings.HasSuffix(filename, folder) {
			continue
		}

		file, err := os.Open(filepath.Clean(path + filename))
		if err != nil {
			panic(err)
		}

		objectKey := fmt.Sprintf("static/%s/%s", folder, filename)
		_, err = storageProvider.Put(objectKey, file)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Uploaded [%s] to [%s]\n", path, objectKey)
	}
}

func updateHtml(domainPath string) {
	htmlPath := "../web/build/index.html"
	html := util.ReadStringFromPath(htmlPath)
	html = strings.Replace(html, "\"/static/", fmt.Sprintf("\"%s", domainPath), -1)
	util.WriteStringToPath(html, htmlPath)

	fmt.Printf("Updated HTML to [%s]\n", html)
}
