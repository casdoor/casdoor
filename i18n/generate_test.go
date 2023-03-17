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
	enData := parseEnData("frontend")

	applyToOtherLanguage("frontend", "en", enData)
	applyToOtherLanguage("frontend", "zh", enData)
	applyToOtherLanguage("frontend", "es", enData)
	applyToOtherLanguage("frontend", "fr", enData)
	applyToOtherLanguage("frontend", "de", enData)
	applyToOtherLanguage("frontend", "ja", enData)
	applyToOtherLanguage("frontend", "ko", enData)
	applyToOtherLanguage("frontend", "ru", enData)
	applyToOtherLanguage("frontend", "vi", enData)
}

func TestGenerateI18nBackend(t *testing.T) {
	enData := parseEnData("backend")

	applyToOtherLanguage("backend", "en", enData)
	applyToOtherLanguage("backend", "zh", enData)
	applyToOtherLanguage("backend", "es", enData)
	applyToOtherLanguage("backend", "fr", enData)
	applyToOtherLanguage("backend", "de", enData)
	applyToOtherLanguage("backend", "ja", enData)
	applyToOtherLanguage("backend", "ko", enData)
	applyToOtherLanguage("backend", "ru", enData)
	applyToOtherLanguage("backend", "vi", enData)
}
