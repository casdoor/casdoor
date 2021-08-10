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

package util

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func FileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func UrlJoin(base string, path string) string {
	res := fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(path, "/"))
	return res
}

func GetUrlPath(urlString string) string {
	u, _ := url.Parse(urlString)
	return u.Path
}

func GetUrlHost(urlString string) string {
	u, _ := url.Parse(urlString)
	return fmt.Sprintf("%s://%s", u.Scheme, u.Host)
}

func GetUrlParams(urlString string) map[string][]string {
	u, _ := url.Parse(urlString)
	q, _ := url.ParseQuery(u.RawQuery)
	return q
}
