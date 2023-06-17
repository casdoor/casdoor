// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"net/http"
	"strings"
)

func getFaviconFileBuffer(client *http.Client, email string) (*bytes.Buffer, string, error) {
	tokens := strings.Split(email, "@")
	domain := tokens[1]
	if domain == "gmail.com" || domain == "163.com" || domain == "qq.com" {
		return nil, "", nil
	}

	//htmlUrl := fmt.Sprintf("https://%s", domain)
	//buffer, fileExtension, err := downloadImage(client, htmlUrl)

	faviconUrl := fmt.Sprintf("https://%s/favicon.ico", domain)
	return downloadImage(client, faviconUrl)
}
