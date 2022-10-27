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

package i18n

import (
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/casdoor/casdoor/util"
	"gopkg.in/ini.v1"
)

//go:embed languages/*.ini
var f embed.FS

var (
	langMapConfig  = make(map[string]*ini.File)
	isNotFirstLoad = make(map[string]bool)
)

func getI18nFilePath(language string) string {
	return fmt.Sprintf("../web/src/locales/%s/data.json", language)
}

func readI18nFile(language string) *I18nData {
	s := util.ReadStringFromPath(getI18nFilePath(language))

	data := &I18nData{}
	err := util.JsonToStruct(s, data)
	if err != nil {
		panic(err)
	}
	return data
}

func writeI18nFile(language string, data *I18nData) {
	s := util.StructToJsonFormatted(data)
	s = strings.ReplaceAll(s, "\\u0026", "&")
	s += "\n"
	println(s)

	util.WriteStringToPath(s, getI18nFilePath(language))
}

func applyData(data1 *I18nData, data2 *I18nData) {
	for namespace, pairs2 := range *data2 {
		if _, ok := (*data1)[namespace]; !ok {
			continue
		}

		pairs1 := (*data1)[namespace]

		for key, value := range pairs2 {
			if _, ok := pairs1[key]; !ok {
				continue
			}

			pairs1[key] = value
		}
	}
}

func Translate(lang string, error string) string {
	parts := strings.Split(error, ".")
	if !strings.Contains(error, ".") || len(parts) != 2 {
		log.Println("Invalid Error Name")
		return ""
	}

	if isNotFirstLoad[lang] {
		return langMapConfig[lang].Section(parts[0]).Key(parts[1]).String()
	} else {
		file, _ := f.ReadFile("languages/locale_" + lang + ".ini")
		langMapConfig[lang], _ = ini.Load(file)
		isNotFirstLoad[lang] = true
		return langMapConfig[lang].Section(parts[0]).Key(parts[1]).String()
	}
}
