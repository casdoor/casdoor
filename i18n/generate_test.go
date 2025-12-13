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

//go:build !skipCi

package i18n

import "testing"

func TestGenerateI18nFrontend(t *testing.T) {
	data := parseAllWords("frontend")

	applyToOtherLanguage("frontend", "en", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "zh", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "es", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "fr", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "de", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "id", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "ja", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "ko", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "ru", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "vi", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "pt", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "it", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "ms", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "tr", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "ar", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "he", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "nl", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "pl", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "fi", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "sv", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "uk", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "kk", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "fa", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "cs", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "sk", deepCopyI18nData(data))
	applyToOtherLanguage("frontend", "az", deepCopyI18nData(data))
}

func TestGenerateI18nBackend(t *testing.T) {
	data := parseAllWords("backend")

	applyToOtherLanguage("backend", "en", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "zh", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "es", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "fr", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "de", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "id", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "ja", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "ko", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "ru", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "vi", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "pt", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "it", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "ms", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "tr", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "ar", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "he", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "nl", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "pl", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "fi", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "sv", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "uk", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "kk", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "fa", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "cs", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "sk", deepCopyI18nData(data))
	applyToOtherLanguage("backend", "az", deepCopyI18nData(data))
}
