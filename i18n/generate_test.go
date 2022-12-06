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
	"testing"
)

func applyToOtherLanguage(dataEn *I18nData, lang string) {
	dataOther := readI18nFile(lang)
	println(dataOther)

	applyData(dataEn, dataOther)
	writeI18nFile(lang, dataEn)
}

func TestGenerateI18nStringsForFrontend(t *testing.T) {
	dataEn := parseToData()
	writeI18nFile("en", dataEn)

	applyToOtherLanguage(dataEn, "de")
	applyToOtherLanguage(dataEn, "fr")
	applyToOtherLanguage(dataEn, "ja")
	applyToOtherLanguage(dataEn, "ko")
	applyToOtherLanguage(dataEn, "ru")
	applyToOtherLanguage(dataEn, "zh")
}

func TestGenerateI18nStringsForBackend(t *testing.T) {
	paths := getAllGoFilePaths()

	errName := getErrName(paths)

	dataEn := getI18nJSONData(errName)

	writeI18nFile("backend_en", dataEn)

	applyToOtherLanguage(dataEn, "backend_de")
	applyToOtherLanguage(dataEn, "backend_fr")
	applyToOtherLanguage(dataEn, "backend_ja")
	applyToOtherLanguage(dataEn, "backend_ko")
	applyToOtherLanguage(dataEn, "backend_ru")
	applyToOtherLanguage(dataEn, "backend_zh")

	fmt.Println("Total Err Words:", len(errName))

	for i := range errName {
		fmt.Println(i)
	}
}
