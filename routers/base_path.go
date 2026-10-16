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
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/conf"
)

var (
	basePath          = conf.GetBasePath()
	reRootRelativeUrl = regexp.MustCompile(`\b(src|href|data-casdoor-local)="/([^/"])`)
)

// BasePathMiddleware lets Casdoor run under a path like "https://example.com/casdoor": the prefix is removed before
// routing, so it works both when the reverse proxy keeps the prefix and when it strips it
func BasePathMiddleware(next http.Handler) http.Handler {
	if basePath == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == basePath || strings.HasPrefix(r.URL.Path, basePath+"/") {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, basePath)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
			if r.URL.RawPath != "" {
				r.URL.RawPath = strings.TrimPrefix(r.URL.RawPath, basePath)
				if r.URL.RawPath == "" {
					r.URL.RawPath = "/"
				}
			}
			r.RequestURI = r.URL.RequestURI()
		}
		next.ServeHTTP(w, r)
	})
}

// addBasePathToHtml prefixes the root-relative script, stylesheet and link URLs of a page with the base path, and
// tells the frontend the base path, which it uses for its routes and API calls
func addBasePathToHtml(content string) string {
	if basePath == "" {
		return content
	}

	content = reRootRelativeUrl.ReplaceAllStringFunc(content, func(s string) string {
		i := strings.Index(s, `="/`)
		return s[:i+2] + basePath + s[i+2:]
	})
	meta := fmt.Sprintf(`<meta name="casdoor-base-path" content="%s" />`, html.EscapeString(basePath))
	return strings.Replace(content, "<head>", "<head>\n    "+meta, 1)
}
