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
	"html"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func nativeLanguageFixture(t *testing.T) {
	t.Helper()
	oldForce, oldDefault, oldBase := forceLanguage, defaultLanguage, basePath
	forceLanguage, defaultLanguage, basePath = "", "en", ""
	t.Cleanup(func() { forceLanguage, defaultLanguage, basePath = oldForce, oldDefault, oldBase })
}

func nativeRequest(language, target string) *http.Request {
	request := httptest.NewRequest(http.MethodGet, target, nil)
	request.Header.Set("Accept-Language", language)
	return request
}

func TestNativeLanguageSelection(t *testing.T) {
	nativeLanguageFixture(t)
	for _, value := range []string{"zh-TW", "zh-Hant", "zh-HK", "zh-MO", "zh_Hant_TW"} {
		if got := normalizeLanguage(value); got != "zh-TW" {
			t.Fatalf("Traditional alias was lost: %s: %q", value, got)
		}
	}
	for _, value := range []string{"zh-CN", "zh-Hans", "zh-Hans-HK"} {
		if got := normalizeLanguage(value); got != "zh" {
			t.Fatalf("Simplified alias changed: %s: %q", value, got)
		}
	}
	request := nativeRequest("zh-Hant-HK,zh;q=0.8,en;q=0.5", "/")
	if got := getIndexHtmlLanguage(request); got != "zh-TW" {
		t.Fatalf("browser locale lost: %q", got)
	}
	forceLanguage = "en"
	if got := getIndexHtmlLanguage(request); got != "en" {
		t.Fatalf("forced locale lost: %q", got)
	}
	forceLanguage, defaultLanguage = "", "zh-TW"
	if got := getIndexHtmlLanguage(nativeRequest("invalid-language-value", "/")); got != "zh-TW" {
		t.Fatalf("default locale lost: %q", got)
	}
}

type nativeCatalogPayload struct {
	Language string                       `json:"language"`
	Catalogs map[string]map[string]string `json:"catalogs"`
}

func nativePayload(t *testing.T, document string) nativeCatalogPayload {
	t.Helper()
	pattern := regexp.MustCompile(`<meta name="casdoor-native-messages" content="([^"]*)">`)
	matches := pattern.FindAllStringSubmatch(document, -1)
	if len(matches) != 1 {
		t.Fatalf("expected one native payload, got %d", len(matches))
	}
	var payload nativeCatalogPayload
	if err := json.Unmarshal([]byte(html.UnescapeString(matches[0][1])), &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestNativeIndexPresentationAndCache(t *testing.T) {
	nativeLanguageFixture(t)
	content, err := os.ReadFile("../web/index.html")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "index.html")
	if err = os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-24 * time.Hour)
	if err = os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	for _, language := range []string{"en", "zh-TW"} {
		request := nativeRequest(language, "/")
		request.Header.Set("If-Modified-Since", time.Now().Add(time.Hour).UTC().Format(http.TimeFormat))
		recorder := httptest.NewRecorder()
		serveFileWithReplace(recorder, request, path, nil)
		if recorder.Code != http.StatusOK {
			t.Fatalf("localized shell returned stale 304: %d", recorder.Code)
		}
		if recorder.Header().Get("Cache-Control") != "no-cache" || !strings.Contains(recorder.Header().Get("Vary"), "Accept-Language") || recorder.Header().Get("Last-Modified") != "" {
			t.Fatalf("wrong dynamic cache policy: %v", recorder.Header())
		}
		document := recorder.Body.String()
		payload := nativePayload(t, document)
		if payload.Language != language || len(payload.Catalogs) != 2 || len(payload.Catalogs["zh-TW"]) != len(nativeMessages("zh-TW")) {
			t.Fatalf("unexpected native payload: %s, %d catalogs", payload.Language, len(payload.Catalogs))
		}
		if strings.Count(document, nativeTextResolver) != 1 || strings.Contains(document, `content="{"`) {
			t.Fatal("native data is unescaped or executable resolver duplicated")
		}
		if !strings.Contains(document, `<html lang="`+language+`">`) {
			t.Fatal("initial document language changed")
		}
		if language == "zh-TW" {
			if !strings.Contains(document, `<noscript>請啟用 JavaScript 以使用此應用程式。</noscript>`) || !strings.Contains(document, `aria-label="載入中"`) || !strings.Contains(document, `>您的瀏覽器版本可能過舊`) {
				t.Fatal("no-script/boot guidance remains untranslated")
			}
		}
	}
	asset := filepath.Join(t.TempDir(), "asset.js")
	if err = os.WriteFile(asset, []byte("var fixture = 1;"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err = os.Chtimes(asset, old, old); err != nil {
		t.Fatal(err)
	}
	request := nativeRequest("zh-TW", "/asset.js")
	request.Header.Set("If-Modified-Since", time.Now().UTC().Format(http.TimeFormat))
	recorder := httptest.NewRecorder()
	serveFileWithReplace(recorder, request, asset, nil)
	if recorder.Code != http.StatusNotModified {
		t.Fatalf("ordinary asset validators changed: %d", recorder.Code)
	}
}

func TestNativeLightweightPages(t *testing.T) {
	nativeLanguageFixture(t)
	basePath = "/casdoor"
	for _, fixture := range []struct {
		target string
		serve  func(*context.Context) bool
		script string
		text   string
	}{
		{"/login/oauth/authorize?provider_hint=synthetic", serveProviderHintRedirectPage, "ProviderHintRedirect.js", "正在重新導向……"},
		{"/callback?state=synthetic", serveAuthCallbackPage, "AuthCallbackHandler.js", "正在登入……"},
	} {
		recorder := httptest.NewRecorder()
		ctx := context.NewContext()
		ctx.Reset(recorder, nativeRequest("zh-TW", fixture.target))
		if !fixture.serve(ctx) {
			t.Fatalf("lightweight route was not served: %s", fixture.target)
		}
		if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("lightweight policy changed: %d %v", recorder.Code, recorder.Header())
		}
		document := recorder.Body.String()
		if !strings.Contains(document, `src="/casdoor/`+fixture.script+`"`) || !strings.Contains(document, ">"+fixture.text+"<") || !strings.Contains(document, `<html lang="zh-TW">`) {
			t.Fatalf("language/base path lost: %s", fixture.target)
		}
		if strings.Count(document, nativeTextResolver) != 1 || nativePayload(t, document).Language != "zh-TW" {
			t.Fatal("resolver or payload changed")
		}
		if ctx.Request.URL.RawQuery != strings.SplitN(fixture.target, "?", 2)[1] {
			t.Fatal("presentation changed callback query/state")
		}
	}
}

func TestLocalizationRouterChallengeAndProtocolBoundary(t *testing.T) {
	value := util.NewMessageError("auth:Access token doesn't exist in database")
	for _, tokenType := range []string{"Bearer", "DPoP"} {
		recorder := httptest.NewRecorder()
		ctx := context.NewContext()
		ctx.Reset(recorder, nativeRequest("zh-TW", "/api/userinfo"))
		responseInvalidTokenFrom(ctx, &object.Token{TokenType: tokenType}, "invalid_token", value)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("challenge status changed: %d", recorder.Code)
		}
		challenge := recorder.Header().Get("WWW-Authenticate")
		if !strings.HasPrefix(challenge, tokenType+" error=\"invalid_token\"") || !strings.Contains(challenge, value.Error()) {
			t.Fatalf("original RFC challenge changed: %q", challenge)
		}
		for _, r := range challenge {
			if r > 127 {
				t.Fatalf("localized prose leaked into challenge: %q", challenge)
			}
		}
		var body Response
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Msg != "資料庫中沒有此存取權杖" {
			t.Fatalf("human body not localized: %#v", body)
		}
	}
	request := nativeRequest("zh-TW", "/api/mcp")
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, request)
	ctx.Input.RequestBody = []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/call"}`)
	responseErrorFrom(ctx, value)
	var envelope struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != -32001 || envelope.Error.Message != "Unauthorized" || envelope.Error.Data != value.Error() {
		t.Fatalf("MCP transport changed: %#v", envelope)
	}
}
