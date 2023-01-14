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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/util"
)

type I18nData map[string]map[string]string

var (
	reI18nFrontend          *regexp.Regexp
	reI18nBackendObject     *regexp.Regexp
	reI18nBackendController *regexp.Regexp
)

func init() {
	reI18nFrontend, _ = regexp.Compile("i18next.t\\(\"(.*?)\"\\)")
	reI18nBackendObject, _ = regexp.Compile("i18n.Translate\\((.*?)\"\\)")
	reI18nBackendController, _ = regexp.Compile("c.T\\((.*?)\"\\)")
}

func getAllI18nStringsFrontend(fileContent string) []string {
	res := []string{}

	matches := reI18nFrontend.FindAllStringSubmatch(fileContent, -1)
	if matches == nil {
		return res
	}

	for _, match := range matches {
		res = append(res, match[1])
	}
	return res
}

func getAllI18nStringsBackend(fileContent string, isObjectPackage bool) []string {
	res := []string{}
	if isObjectPackage {
		matches := reI18nBackendObject.FindAllStringSubmatch(fileContent, -1)
		if matches == nil {
			return res
		}
		for _, match := range matches {
			match := strings.SplitN(match[1], ",", 2)
			res = append(res, match[1][2:])
		}
	} else {
		matches := reI18nBackendController.FindAllStringSubmatch(fileContent, -1)
		if matches == nil {
			return res
		}
		for _, match := range matches {
			res = append(res, match[1][1:])
		}
	}

	return res
}

func getAllFilePathsInFolder(folder string, fileSuffix string) []string {
	res := []string{}
	err := filepath.Walk(folder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(info.Name(), fileSuffix) {
				return nil
			}

			res = append(res, path)
			fmt.Println(path, info.Name())
			return nil
		})
	if err != nil {
		panic(err)
	}

	return res
}

func parseEnData(category string) *I18nData {
	var paths []string
	if category == "backend" {
		paths = getAllFilePathsInFolder("../", ".go")
	} else {
		paths = getAllFilePathsInFolder("../web/src", ".js")
	}

	allWords := []string{}
	for _, path := range paths {
		fileContent := util.ReadStringFromPath(path)

		var words []string
		if category == "backend" {
			isObjectPackage := strings.Contains(path, "object")
			words = getAllI18nStringsBackend(fileContent, isObjectPackage)
		} else {
			words = getAllI18nStringsFrontend(fileContent)
		}
		allWords = append(allWords, words...)
	}
	fmt.Printf("%v\n", allWords)

	data := I18nData{}
	for _, word := range allWords {
		tokens := strings.SplitN(word, ":", 2)
		namespace := tokens[0]
		key := tokens[1]

		if _, ok := data[namespace]; !ok {
			data[namespace] = map[string]string{}
		}
		data[namespace][key] = key
	}

	return &data
}
