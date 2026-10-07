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
	"strconv"
	"strings"

	"github.com/casdoor/casdoor/util"
)

type I18nData map[string]map[string]string

var (
	reI18nFrontend          *regexp.Regexp
	reI18nFrontendNamespace *regexp.Regexp
	reI18nFrontendDynamic   *regexp.Regexp
	reI18nBackendObject     *regexp.Regexp
	reI18nBackendController *regexp.Regexp
)

func init() {
	// the console passes most keys around as plain strings (labelKey="general:Name"),
	// so every "namespace:key" literal counts, not only the argument of i18next.t()
	reI18nFrontend = regexp.MustCompile("\"(\\w+:(?:[^\"\\\\\\n]|\\\\.)+)\"")
	reI18nFrontendNamespace = regexp.MustCompile("i18next\\.t\\([\"`](\\w+):")
	// keys built at runtime: i18next.t("webhook:" + state), i18next.t(`rule:${item}`)
	reI18nFrontendDynamic = regexp.MustCompile("\"(\\w+):\"\\s*\\+|`(\\w+):\\$\\{")
	reI18nBackendObject, _ = regexp.Compile("i18n.Translate\\((.*?)\"\\)")
	reI18nBackendController, _ = regexp.Compile("c.T\\((.*?)\"\\)")
}

func getAllI18nStringsFrontend(fileContent string, namespaces map[string]bool) []string {
	res := []string{}

	matches := reI18nFrontend.FindAllStringSubmatchIndex(fileContent, -1)
	if matches == nil {
		return res
	}

	for _, match := range matches {
		// an OAuth scope like "user:email" is not an i18n key
		if strings.HasSuffix(strings.TrimRight(fileContent[:match[0]], " "), "scope:") {
			continue
		}

		raw := fileContent[match[2]:match[3]]
		target, err := strconv.Unquote("\"" + raw + "\"")
		if err != nil {
			target = raw
		}

		if !namespaces[strings.SplitN(target, ":", 2)[0]] {
			continue
		}
		res = append(res, target)
	}
	return res
}

func getI18nNamespacesFrontend(fileContent string) []string {
	res := []string{}
	for _, match := range reI18nFrontendNamespace.FindAllStringSubmatch(fileContent, -1) {
		res = append(res, match[1])
	}
	return res
}

func getDynamicI18nNamespacesFrontend(fileContent string) []string {
	res := []string{}
	for _, match := range reI18nFrontendDynamic.FindAllStringSubmatch(fileContent, -1) {
		res = append(res, match[1]+match[2])
	}
	return res
}

// keepDerivedWords keeps the existing keys that the code never spells out in full:
// "X - Tooltip" (FormRow derives it from labelKey), "New X" / "View X" (getModeTitleKey
// derives them from "Edit X") and every key of a namespace whose keys are built at runtime.
func keepDerivedWords(data *I18nData, oldData *I18nData, dynamicNamespaces map[string]bool) {
	for namespace, oldPairs := range *oldData {
		pairs, ok := (*data)[namespace]
		if !ok {
			if !dynamicNamespaces[namespace] {
				continue
			}
			pairs = map[string]string{}
			(*data)[namespace] = pairs
		}

		for key := range oldPairs {
			keep := dynamicNamespaces[namespace]
			if base, found := strings.CutSuffix(key, " - Tooltip"); found {
				if _, ok := pairs[base]; ok {
					keep = true
				}
			}
			for _, prefix := range []string{"New ", "View "} {
				if base, found := strings.CutPrefix(key, prefix); found {
					if _, ok := pairs["Edit "+base]; ok {
						keep = true
					}
				}
			}

			if _, ok := pairs[key]; !ok && keep {
				pairs[key] = key
			}
		}
	}
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
			target, err := strconv.Unquote("\"" + match[1][2:] + "\"")
			if err != nil {
				target = match[1][2:]
			}

			res = append(res, target)
		}
	} else {
		matches := reI18nBackendController.FindAllStringSubmatch(fileContent, -1)
		if matches == nil {
			return res
		}
		for _, match := range matches {
			target, err := strconv.Unquote("\"" + match[1][1:] + "\"")
			if err != nil {
				target = match[1][1:]
			}
			res = append(res, target)
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

			// hidden folders hold other checkouts of this repo (.claude/worktrees, .git)
			if info.IsDir() && path != folder && (info.Name() == "node_modules" || strings.HasPrefix(info.Name(), ".")) {
				return filepath.SkipDir
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

func parseAllWords(category string) *I18nData {
	var paths []string
	if category == "backend" {
		paths = getAllFilePathsInFolder("../", ".go")
	} else {
		// the console is TypeScript, so ".js" alone would walk right past it
		paths = getAllFilePathsInFolder("../web/src", ".tsx")
		paths = append(paths, getAllFilePathsInFolder("../web/src", ".ts")...)
	}

	fileContents := []string{}
	for _, path := range paths {
		fileContents = append(fileContents, util.ReadStringFromPath(path))
	}

	var oldData *I18nData
	namespaces := map[string]bool{}
	dynamicNamespaces := map[string]bool{}
	if category != "backend" {
		oldData = readI18nFile(category, "en")
		for namespace := range *oldData {
			namespaces[namespace] = true
		}
		for _, fileContent := range fileContents {
			for _, namespace := range getI18nNamespacesFrontend(fileContent) {
				namespaces[namespace] = true
			}
			for _, namespace := range getDynamicI18nNamespacesFrontend(fileContent) {
				dynamicNamespaces[namespace] = true
			}
		}
	}

	allWords := []string{}
	for i, path := range paths {
		fileContent := fileContents[i]

		var words []string
		if category == "backend" {
			isObjectPackage := strings.Contains(path, "object")
			words = getAllI18nStringsBackend(fileContent, isObjectPackage)
		} else {
			words = getAllI18nStringsFrontend(fileContent, namespaces)
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

	if category != "backend" {
		keepDerivedWords(&data, oldData, dynamicNamespaces)
	}

	return &data
}

// copyI18nData creates a deep copy of an I18nData structure to prevent shared reference issues
// between language translations. This ensures each language starts with fresh English defaults
// rather than inheriting values from previously processed languages.
func copyI18nData(src *I18nData) *I18nData {
	dst := I18nData{}
	for namespace, pairs := range *src {
		dst[namespace] = make(map[string]string)
		for key, value := range pairs {
			dst[namespace][key] = value
		}
	}
	return &dst
}

func applyToOtherLanguage(category string, language string, newData *I18nData) {
	oldData := readI18nFile(category, language)
	println(oldData)

	// Create a copy of newData to avoid modifying the shared data across languages
	dataCopy := copyI18nData(newData)
	applyData(dataCopy, oldData)
	writeI18nFile(category, language, dataCopy)
}
