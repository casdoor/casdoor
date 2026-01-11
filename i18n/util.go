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
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
)

var enableErrorMask = false

//go:embed locales/*/data.json
var f embed.FS

var langMap = make(map[string]map[string]map[string]string) // for example : langMap[en][account][Invalid information] = Invalid information

func init() {
	enableErrorMask = conf.GetConfigBool("enableErrorMask")
}

func getI18nFilePath(category string, language string) string {
	if category == "backend" {
		return fmt.Sprintf("../i18n/locales/%s/data.json", language)
	} else {
		return fmt.Sprintf("../web/src/locales/%s/data.json", language)
	}
}

func readI18nFile(category string, language string) *I18nData {
	s := util.ReadStringFromPath(getI18nFilePath(category, language))

	data := &I18nData{}
	err := util.JsonToStruct(s, data)
	if err != nil {
		panic(err)
	}
	return data
}

func writeI18nFile(category string, language string, data *I18nData) {
	s := util.StructToJsonFormatted(data)
	s = strings.ReplaceAll(s, "\\u0026", "&")
	s += "\n"
	println(s)

	util.WriteStringToPath(s, getI18nFilePath(category, language))
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

func Translate(language string, errorText string) string {
	modified := false
	if enableErrorMask {
		if errorText == "general:The user: %s doesn't exist" ||
			errorText == "check:password or code is incorrect, you have %s remaining chances" {
			modified = true
			errorText = "check:password or code is incorrect"
		}
	}

	tokens := strings.SplitN(errorText, ":", 2)
	if !strings.Contains(errorText, ":") || len(tokens) != 2 {
		return fmt.Sprintf("Translate error: the error text doesn't contain \":\", errorText = %s", errorText)
	}

	if langMap[language] == nil {
		file, err := f.ReadFile(fmt.Sprintf("locales/%s/data.json", language))
		if err != nil {
			return fmt.Sprintf("Translate error: the language \"%s\" is not supported, err = %s", language, err.Error())
		}

		data := I18nData{}
		err = util.JsonToStruct(string(file), &data)
		if err != nil {
			panic(err)
		}
		langMap[language] = data
	}

	res := langMap[language][tokens[0]][tokens[1]]
	if res == "" {
		res = tokens[1]
	}

	if modified {
		res += "%.s"
	}
	return res
}

// TranslatableError is an error type that can be translated when needed
// It stores the translation key and parameters, allowing translation to happen
// at the point where the error is displayed, not when it's created
type TranslatableError struct {
	Key    string
	Params []interface{}
}

// NewTranslatableError creates a new TranslatableError with the given key and parameters
// The key should be in the format "namespace:message" (e.g., "general:User not found")
// Parameters can be used for formatting the translated message
func NewTranslatableError(key string, params ...interface{}) *TranslatableError {
	return &TranslatableError{
		Key:    key,
		Params: params,
	}
}

// Error implements the error interface, returning the untranslated key
// This is useful for logging or when translation context is not available
func (te *TranslatableError) Error() string {
	firstColonIndex := strings.Index(te.Key, ":")
	if firstColonIndex == -1 {
		// If no ":" is found, format with params if any
		if len(te.Params) > 0 {
			return fmt.Sprintf(te.Key, te.Params...)
		}
		return te.Key
	}
	// Extract the part after the first ":"
	keyWithoutCategory := te.Key[firstColonIndex+1:]
	if len(te.Params) > 0 {
		return fmt.Sprintf(keyWithoutCategory, te.Params...)
	}
	return keyWithoutCategory
}

func (te *TranslatableError) TranslateWithLanguage(language string) string {
	translatedFormat := Translate(language, te.Key)
	if len(te.Params) > 0 {
		return fmt.Sprintf(translatedFormat, te.Params...)
	}
	return translatedFormat
}

// T is a convenience function that creates a TranslatableError
// Usage: return i18n.T("general:User not found", userId)
func T(key string, params ...interface{}) *TranslatableError {
	return NewTranslatableError(key, params...)
}
