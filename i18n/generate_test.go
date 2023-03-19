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
}
