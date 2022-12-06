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
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/util"
)

var (
	reI18nBackendObject    *regexp.Regexp
	re18nBackendController *regexp.Regexp
)

func init() {
	reI18nBackendObject, _ = regexp.Compile("i18n.Translate\\((.*?)\"\\)")
	re18nBackendController, _ = regexp.Compile("c.T\\((.*?)\"\\)")
}

func GetAllI18nStrings(fileContent string, path string) []string {
	res := []string{}
	if strings.Contains(path, "object") {
		matches := reI18nBackendObject.FindAllStringSubmatch(fileContent, -1)
		if matches == nil {
			return res
		}
		for _, match := range matches {
			match := strings.SplitN(match[1], ",", 2)
			res = append(res, match[1][2:])
		}
	} else {
		matches := re18nBackendController.FindAllStringSubmatch(fileContent, -1)
		if matches == nil {
			return res
		}
		for _, match := range matches {
			res = append(res, match[1][1:])
		}
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

func getErrName(paths []string) map[string]string {
	ErrName := make(map[string]string)
	for i := 0; i < len(paths); i++ {
		content := util.ReadStringFromPath(paths[i])
		words := GetAllI18nStrings(content, paths[i])
		for j := 0; j < len(words); j++ {
			ErrName[words[j]] = paths[i]
		}
	}
	return ErrName
}

func getI18nJSONData(errName map[string]string) *I18nData {
	data := I18nData{}
	for k, v := range errName {
		index := strings.LastIndex(v, "\\")
		namespace := v[index+1 : len(v)-3]
		key := k[len(namespace)+1:]
		if _, ok := data[namespace]; !ok {
			data[namespace] = map[string]string{}
		}
		data[namespace][key] = key
	}
	return &data
}
