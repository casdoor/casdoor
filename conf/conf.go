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

package conf

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/server/web"
	"golang.org/x/text/language"
)

//go:embed waf.conf
var WafConf string

func init() {
	// this array contains the beego configuration items that may be modified via env
	presetConfigItems := []string{"httpport", "appname"}
	for _, key := range presetConfigItems {
		if value, ok := os.LookupEnv(key); ok {
			err := web.AppConfig.Set(key, value)
			if err != nil {
				panic(err)
			}
		}
	}
}

func GetConfigString(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	res, _ := web.AppConfig.String(key)
	if res == "" {
		if key == "staticBaseUrl" {
			res = "https://cdn.casbin.org"
		} else if key == "logConfig" {
			appname, _ := web.AppConfig.String("appname")
			res = fmt.Sprintf("{\"filename\": \"logs/%s.log\", \"maxdays\":99999, \"perm\":\"0770\"}", appname)
		}
	}

	return res
}

func GetConfigBool(key string) bool {
	value := GetConfigString(key)
	if value == "true" {
		return true
	} else {
		return false
	}
}

func GetConfigInt64(key string) (int64, error) {
	value := GetConfigString(key)
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("GetConfigInt64(%s) error, %s", key, err.Error())
	}

	return num, nil
}

// GetVerificationCodeTimeout falls back to 10 minutes when the config item is missing or invalid
func GetVerificationCodeTimeout() int64 {
	timeoutInMinutes, err := GetConfigInt64("verificationCodeTimeout")
	if err != nil || timeoutInMinutes <= 0 {
		return 10
	}

	return timeoutInMinutes
}

func GetConfigDataSourceName() string {
	dataSourceName := GetConfigString("dataSourceName")
	return ReplaceDataSourceNameByDocker(dataSourceName)
}

func ReplaceDataSourceNameByDocker(dataSourceName string) string {
	runningInDocker := os.Getenv("RUNNING_IN_DOCKER")
	if runningInDocker == "true" {
		// https://stackoverflow.com/questions/48546124/what-is-linux-equivalent-of-host-docker-internal
		if runtime.GOOS == "linux" {
			dataSourceName = strings.ReplaceAll(dataSourceName, "localhost", "172.17.0.1")
		} else {
			dataSourceName = strings.ReplaceAll(dataSourceName, "localhost", "host.docker.internal")
		}
	}
	return dataSourceName
}

func NormalizeLanguage(value string) string {
	tag, err := language.Parse(strings.ReplaceAll(strings.TrimSpace(value), "_", "-"))
	if err != nil {
		return ""
	}
	base, _, _ := tag.Raw()
	code := base.String()
	if code == "zh" {
		script, _ := tag.Script()
		if script.String() == "Hant" {
			return "zh-TW"
		}
	}
	if len(code) != 2 || code == "nu" {
		return ""
	}
	return code
}

func GetAcceptLanguage(acceptLanguage string) string {
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err == nil {
		for _, tag := range tags {
			if code := NormalizeLanguage(tag.String()); code != "" {
				return code
			}
		}
	}
	return "en"
}

func GetLanguage(language string) string {
	if code := NormalizeLanguage(language); code != "" {
		return code
	}
	return "en"
}

func IsDemoMode() bool {
	return strings.ToLower(GetConfigString("isDemoMode")) == "true"
}

func IsDemoDatabase() bool {
	return GetConfigString("dbName") == "casdoor_demo"
}

func GetConfigBatchSize() int {
	res, err := strconv.Atoi(GetConfigString("batchSize"))
	if err != nil {
		res = 100
	}
	return res
}

// GetBasePath returns the path prefix Casdoor is served under behind a reverse proxy, like "/casdoor", or "" at the root
func GetBasePath() string {
	res := strings.Trim(strings.TrimSpace(GetConfigString("basePath")), "/")
	if res == "" {
		return ""
	}
	return "/" + res
}
