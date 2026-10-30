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
	"html"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/i18n"
	"golang.org/x/text/language"
)

var (
	forceLanguage   = conf.GetConfigString("forceLanguage")
	defaultLanguage = conf.GetConfigString("defaultLanguage")
)

// supportedLanguages mirrors the list used by the frontend in web/src/i18n.ts.
// It is used to normalize the language written into the <html lang="..."> attribute
// of the served index.html, so browsers do not mis-detect the page language and
// offer to translate a page that already matches the user's language.
var supportedLanguages = map[string]bool{
	"en": true, "zh": true, "zh-TW": true, "es": true, "fr": true, "de": true, "id": true,
	"ja": true, "ko": true, "ru": true, "vi": true, "pt": true, "it": true,
	"ms": true, "tr": true, "ar": true, "he": true, "nl": true, "pl": true,
	"fi": true, "sv": true, "uk": true, "kk": true, "fa": true, "cs": true,
	"sk": true, "az": true,
}

// normalizeLanguage returns the canonical supported locale, otherwise "".
func normalizeLanguage(value string) string {
	code := conf.NormalizeLanguage(value)
	if supportedLanguages[code] {
		return code
	}
	return ""
}

// parseAcceptLanguage picks the first supported language from an Accept-Language
// header value such as "zh-CN,zh;q=0.9,en;q=0.8".
func parseAcceptLanguage(header string) string {
	tags, _, err := language.ParseAcceptLanguage(header)
	if err == nil {
		for _, tag := range tags {
			if lang := normalizeLanguage(tag.String()); lang != "" {
				return lang
			}
		}
	}
	return ""
}

// getIndexHtmlLanguage determines the language for the <html lang="..."> attribute
// of index.html, mirroring the frontend precedence in web/src/i18n.ts as closely as
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

// These first-party messages must work even when the React bundle cannot load.
// Keep the resolver ES5-only and its payload separate from executable script.
const nativeTextResolver = `(function() {
  var meta = document.querySelector('meta[name="casdoor-native-messages"]');
  var payload;
  try { payload = JSON.parse(meta ? meta.getAttribute("content") : "{}"); } catch (e) { payload = {}; }
  var catalogs = payload.catalogs || {};
  var selected = "";
  try { selected = window.localStorage.getItem("language") || ""; } catch (e) {}
  selected = (selected || payload.language || "en").replace(/_/g, "-").toLowerCase();
  if (/^(zh|cmn)(-|$)/.test(selected)) {
    selected = /-hans(-|$)/.test(selected) ? "zh" : /-(hant|tw|hk|mo)(-|$)/.test(selected) ? "zh-TW" : "zh";
  } else {
    selected = selected.split("-")[0];
  }
  if (!Object.prototype.hasOwnProperty.call(catalogs, selected)) { selected = payload.language || "en"; }
  var messages = catalogs[selected] || catalogs.en || {};
  window.CasdoorNativeText = function(key) {
    return Object.prototype.hasOwnProperty.call(messages, key) && typeof messages[key] === "string" ? messages[key] : key;
  };
  document.documentElement.lang = selected;
  function localize() {
    var nodes = document.querySelectorAll("[data-casdoor-native-text]");
    var initial = catalogs[payload.language] || {};
    for (var i = 0; i < nodes.length; i++) {
      var key = nodes[i].getAttribute("data-casdoor-native-text");
      if (nodes[i].textContent === key || nodes[i].textContent === initial[key]) {
        nodes[i].textContent = window.CasdoorNativeText(key);
      }
    }
    var loading = document.getElementById("casdoor-boot-fallback");
    if (loading) { loading.setAttribute("aria-label", window.CasdoorNativeText("Loading")); }
  }
  if (document.readyState === "loading") { document.addEventListener("DOMContentLoaded", localize); } else { localize(); }
})();`

func nativeMessages(language string) map[string]string {
	return map[string]string{
		"Redirecting...":                   i18n.Translate(language, "native:Redirecting..."),
		"Signing in...":                    i18n.Translate(language, "native:Signing in..."),
		"Failed to load callback handler.": i18n.Translate(language, "native:Failed to load callback handler."),
		"Missing callback state.":          i18n.Translate(language, "native:Missing callback state."),
		"Unauthorized":                     i18n.Translate(language, "native:Unauthorized"),
		"Failed to sign in.":               i18n.Translate(language, "native:Failed to sign in."),
		"Logged in successfully. Now you can visit apps protected by Casdoor.": i18n.Translate(language, "native:Logged in successfully. Now you can visit apps protected by Casdoor."),
		"Unsupported SAML callback response.":                                  i18n.Translate(language, "native:Unsupported SAML callback response."),
		"Failed to complete callback.":                                         i18n.Translate(language, "native:Failed to complete callback."),
		"Casdoor failed to load":                                               i18n.Translate(language, "native:Casdoor failed to load"),
		"Could not load: ":                                                     i18n.Translate(language, "native:Could not load: "),
		"Casdoor is taking too long to load":                                   i18n.Translate(language, "native:Casdoor is taking too long to load"),
		"You need to enable JavaScript to run this app.":                       i18n.Translate(language, "native:You need to enable JavaScript to run this app."),
		"Loading": i18n.Translate(language, "native:Loading"),
		"Your browser may be too old to run Casdoor. Please open this page in a recent version of Chrome, Edge, Firefox or Safari.": i18n.Translate(language, "native:Your browser may be too old to run Casdoor. Please open this page in a recent version of Chrome, Edge, Firefox or Safari."),
		"If this page was opened inside an app's built-in login window, sign in with your system browser instead.":                  i18n.Translate(language, "native:If this page was opened inside an app's built-in login window, sign in with your system browser instead."),
	}
}

func localizeNativeHtml(content string, r *http.Request) string {
	language := getIndexHtmlLanguage(r)
	catalogs := map[string]map[string]string{"en": nativeMessages("en"), "zh-TW": nativeMessages("zh-TW")}
	catalogs[language] = nativeMessages(language)
	payload, _ := json.Marshal(struct {
		Language string                       `json:"language"`
		Catalogs map[string]map[string]string `json:"catalogs"`
	}{Language: language, Catalogs: catalogs})
	meta := fmt.Sprintf(`<meta name="casdoor-native-messages" content="%s"><script>%s</script>`, html.EscapeString(string(payload)), nativeTextResolver)
	content = strings.Replace(content, "<head>", "<head>\n"+meta, 1)
	content = strings.ReplaceAll(content, `<html lang="en">`, fmt.Sprintf(`<html lang="%s">`, language))
	for source, translated := range catalogs[language] {
		content = strings.ReplaceAll(content, ">"+source+"<", ">"+html.EscapeString(translated)+"<")
		content = strings.ReplaceAll(content, "\n            "+source+"\n", "\n            "+html.EscapeString(translated)+"\n")
	}
	content = strings.ReplaceAll(content, `aria-label="Loading"`, `aria-label="`+html.EscapeString(catalogs[language]["Loading"])+`"`)
	return content
}
