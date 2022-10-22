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

package i18n

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Unknwon/goconfig"
	"github.com/casdoor/casdoor/util"
)

var ReI18n *regexp.Regexp

func init() {
	ReI18n, _ = regexp.Compile("conf.Translate\\((.*?)\"\\)")
}

func GetAllI18nStrings(fileContent string) []string {
	res := []string{}

	matches := ReI18n.FindAllStringSubmatch(fileContent, -1)
	if matches == nil {
		return res
	}
	for _, match := range matches {
		match := strings.Split(match[1], ",")
		res = append(res, match[1][2:])
	}
	return res
}

func getAllGoFilePaths() []string {
	path := "../"

	res := []string{}
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(info.Name(), ".go") {
				return nil
			}

			res = append(res, path)
			// fmt.Println(path, info.Name())
			return nil
		})
	if err != nil {
		panic(err)
	}

	return res
}

func getErrName(paths []string) map[string]bool {
	ErrName := make(map[string]bool)
	for i := 0; i < len(paths); i++ {
		content := util.ReadStringFromPath(paths[i])
		words := GetAllI18nStrings(content)
		for i := 0; i < len(words); i++ {
			ErrName[words[i]] = true
		}
	}
	return ErrName
}

func writeToAllLanguageFiles(errName map[string]bool) {
	languages := "en,zh,es,fr,de,ja,ko,ru"
	languageArr := strings.Split(languages, ",")
	var c [10]*goconfig.ConfigFile
	for i := 0; i < len(languageArr); i++ {
		var err error
		c[i], err = goconfig.LoadConfigFile("../conf/languages/" + "locale_" + languageArr[i] + ".ini")
		if err != nil {
			log.Println(err.Error())
		}
		for j := range errName {
			parts := strings.Split(j, ".")

			_, err := c[i].GetValue(parts[0], parts[1])
			if err != nil {
				c[i].SetValue(parts[0], parts[1], parts[1])
			}
		}
		c[i].SetPrettyFormat(true)
		err = goconfig.SaveConfigFile(c[i], "../conf/languages/"+"locale_"+languageArr[i]+".ini")
		if err != nil {
			log.Println(err)
		}
	}
}
