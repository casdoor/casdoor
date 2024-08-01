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
// +build !skipCi

package i18n

import "testing"

func TestGenerateI18nFrontend(t *testing.T) {
	data := parseAllWords("frontend")

	applyToOtherLanguage("frontend", "en", data)
	applyToOtherLanguage("frontend", "zh", data)
	applyToOtherLanguage("frontend", "es", data)
	applyToOtherLanguage("frontend", "fr", data)
	applyToOtherLanguage("frontend", "de", data)
	applyToOtherLanguage("frontend", "id", data)
	applyToOtherLanguage("frontend", "ja", data)
	applyToOtherLanguage("frontend", "ko", data)
	applyToOtherLanguage("frontend", "ru", data)
	applyToOtherLanguage("frontend", "vi", data)
	applyToOtherLanguage("frontend", "pt", data)
	applyToOtherLanguage("frontend", "it", data)
	applyToOtherLanguage("frontend", "ms", data)
	applyToOtherLanguage("frontend", "tr", data)
	applyToOtherLanguage("frontend", "ar", data)
	applyToOtherLanguage("frontend", "he", data)
	applyToOtherLanguage("frontend", "nl", data)
	applyToOtherLanguage("frontend", "pl", data)
	applyToOtherLanguage("frontend", "fi", data)
	applyToOtherLanguage("frontend", "sv", data)
	applyToOtherLanguage("frontend", "uk", data)
	applyToOtherLanguage("frontend", "kk", data)
	applyToOtherLanguage("frontend", "fa", data)
	applyToOtherLanguage("frontend", "cs", data)
	applyToOtherLanguage("frontend", "sk", data)
}

func TestGenerateI18nBackend(t *testing.T) {
	data := parseAllWords("backend")

	applyToOtherLanguage("backend", "en", data)
	applyToOtherLanguage("backend", "zh", data)
	applyToOtherLanguage("backend", "es", data)
	applyToOtherLanguage("backend", "fr", data)
	applyToOtherLanguage("backend", "de", data)
	applyToOtherLanguage("backend", "id", data)
	applyToOtherLanguage("backend", "ja", data)
	applyToOtherLanguage("backend", "ko", data)
	applyToOtherLanguage("backend", "ru", data)
	applyToOtherLanguage("backend", "vi", data)
	applyToOtherLanguage("backend", "pt", data)
	applyToOtherLanguage("backend", "it", data)
	applyToOtherLanguage("backend", "ms", data)
	applyToOtherLanguage("backend", "tr", data)
	applyToOtherLanguage("backend", "ar", data)
	applyToOtherLanguage("backend", "he", data)
	applyToOtherLanguage("backend", "nl", data)
	applyToOtherLanguage("backend", "pl", data)
	applyToOtherLanguage("backend", "fi", data)
	applyToOtherLanguage("backend", "sv", data)
	applyToOtherLanguage("backend", "uk", data)
	applyToOtherLanguage("backend", "kk", data)
	applyToOtherLanguage("backend", "fa", data)
	applyToOtherLanguage("backend", "cs", data)
	applyToOtherLanguage("backend", "sk", data)
}
