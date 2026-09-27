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
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const dataSourceFieldChars = "?#/()'\"\\=;& \t\r\n"

func checkDataSourceFields(fields map[string]string) error {
	for name, value := range fields {
		if strings.ContainsAny(value, dataSourceFieldChars) {
			return fmt.Errorf("the %s: %s of the database connection contains characters that are not allowed", name, value)
		}
	}
	return nil
}

func quotePostgresDataSourceValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `\'`)
	return fmt.Sprintf("'%s'", value)
}

func getMssqlDataSourceName(user string, password string, host string, port int, database string) string {
	dataSourceUrl := url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, strconv.Itoa(port)),
		RawQuery: url.Values{"database": {database}}.Encode(),
	}
	return dataSourceUrl.String()
}
