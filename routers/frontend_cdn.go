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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/casdoor/casdoor/conf"
)

// frontendCdnUrl is where the frontend build published as an npm package (casdoor-web) is served from, with "{name}"
// and "{version}" taken from the package.json in the build folder, e.g. "https://cdn.jsdelivr.net/npm/{name}@{version}".
// Only a build installed from that package has the package.json, any other build is served from this server.
var frontendCdnUrl = conf.GetConfigString("frontendCdnUrl")

const frontendCdnOffCookie = "casdoor_cdn_off"

var reFrontendAsset = regexp.MustCompile(`(<script type="module" crossorigin src=|<link rel="stylesheet" crossorigin href=|<link rel="modulepreload" crossorigin href=)"(/assets/[^"]+)"`)

// When a file can't be loaded from the CDN, it is loaded from this server instead, and when a lazily loaded chunk
// fails, the page is reloaded once without the CDN (the cookie makes the server keep the local paths).
// Keep it ES5-only like the boot fallback in index.html.
const frontendCdnScript = `<script>
      (function() {
        function disableCdn() {
          document.cookie = "` + frontendCdnOffCookie + `=1; path=/; SameSite=Lax";
        }

        window.casdoorCdnFallback = function(el) {
          disableCdn();
          el.onerror = null;
          var local = el.getAttribute("data-casdoor-local");
          if (el.tagName === "LINK") {
            el.href = local;
            return;
          }
          var script = document.createElement("script");
          script.type = "module";
          script.crossOrigin = "anonymous";
          script.src = local;
          document.head.appendChild(script);
        };

        window.addEventListener("vite:preloadError", function(event) {
          if (document.cookie.indexOf("` + frontendCdnOffCookie + `=1") !== -1) {
            return;
          }
          disableCdn();
          if (event && event.preventDefault) {
            event.preventDefault();
          }
          window.location.reload();
        });
      })();
    </script>
    `

type frontendPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func getFrontendCdnBase(htmlPath string) string {
	if frontendCdnUrl == "" {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(filepath.Dir(htmlPath), "package.json"))
	if err != nil {
		return ""
	}

	pkg := frontendPackage{}
	err = json.Unmarshal(data, &pkg)
	if err != nil || pkg.Name == "" || pkg.Version == "" {
		return ""
	}

	res := strings.ReplaceAll(frontendCdnUrl, "{name}", pkg.Name)
	res = strings.ReplaceAll(res, "{version}", pkg.Version)
	return strings.TrimSuffix(res, "/")
}

// useFrontendCdn points the script and stylesheets of index.html to the CDN, keeping the local paths as the fallback.
// The bool is true when the CDN is configured for this build, so the response depends on the cookie, not only on the file.
func useFrontendCdn(content string, htmlPath string, r *http.Request) (string, bool) {
	cdn := getFrontendCdnBase(htmlPath)
	if cdn == "" {
		return content, false
	}

	if cookie, err := r.Cookie(frontendCdnOffCookie); err == nil && cookie.Value == "1" {
		return content, true
	}

	first := reFrontendAsset.FindStringIndex(content)
	if first == nil {
		return content, true
	}

	assets := reFrontendAsset.ReplaceAllStringFunc(content[first[0]:], func(s string) string {
		m := reFrontendAsset.FindStringSubmatch(s)
		if strings.Contains(m[1], "modulepreload") {
			return fmt.Sprintf(`%s"%s%s"`, m[1], cdn, m[2])
		}
		return fmt.Sprintf(`%s"%s%s" data-casdoor-local="%s" onerror="casdoorCdnFallback(this)"`, m[1], cdn, m[2], m[2])
	})
	return content[:first[0]] + frontendCdnScript + assets, true
}
