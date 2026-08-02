// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package routers

import (
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/conf"
)

var (
	forceLanguage   = conf.GetConfigString("forceLanguage")
	defaultLanguage = conf.GetConfigString("defaultLanguage")
)

// supportedLanguages mirrors the list used by the frontend in web/src/i18n.js.
// It is used to normalize the language written into the <html lang="..."> attribute
// of the served index.html, so browsers do not mis-detect the page language and
// offer to translate a page that already matches the user's language.
var supportedLanguages = map[string]bool{
	"en": true, "zh": true, "es": true, "fr": true, "de": true, "id": true,
	"ja": true, "ko": true, "ru": true, "vi": true, "pt": true, "it": true,
	"ms": true, "tr": true, "ar": true, "he": true, "nl": true, "pl": true,
	"fi": true, "sv": true, "uk": true, "kk": true, "fa": true, "cs": true,
	"sk": true, "az": true,
}

// normalizeLanguage returns the 2-letter code if it is supported, otherwise "".
func normalizeLanguage(language string) string {
	language = strings.ToLower(strings.TrimSpace(language))
	if supportedLanguages[language] {
		return language
	}
	return ""
}

// parseAcceptLanguage picks the first supported language from an Accept-Language
// header value such as "zh-CN,zh;q=0.9,en;q=0.8".
func parseAcceptLanguage(header string) string {
	for _, part := range strings.Split(header, ",") {
		tag := strings.TrimSpace(part)
		if i := strings.Index(tag, ";"); i != -1 {
			tag = tag[:i]
		}
		base := strings.SplitN(tag, "-", 2)[0]
		if lang := normalizeLanguage(base); lang != "" {
			return lang
		}
	}
	return ""
}

// getIndexHtmlLanguage determines the language for the <html lang="..."> attribute
// of index.html, mirroring the frontend precedence in web/src/i18n.js as closely as
// the server can (localStorage is not visible to the server): forceLanguage first,
// then the browser's Accept-Language, then defaultLanguage, falling back to "en".
func getIndexHtmlLanguage(r *http.Request) string {
	if lang := normalizeLanguage(forceLanguage); lang != "" {
		return lang
	}
	if lang := parseAcceptLanguage(r.Header.Get("Accept-Language")); lang != "" {
		return lang
	}
	if lang := normalizeLanguage(defaultLanguage); lang != "" {
		return lang
	}
	return "en"
}
