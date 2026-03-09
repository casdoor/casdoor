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

package routers

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

var (
	oldStaticBaseUrl = "https://cdn.casbin.org"
	newStaticBaseUrl = conf.GetConfigString("staticBaseUrl")
	enableGzip       = conf.GetConfigBool("enableGzip")
	frontendBaseDir  = conf.GetConfigString("frontendBaseDir")
)

func getWebBuildFolder() string {
	path := "web/build"
	if util.FileExist(filepath.Join(path, "index.html")) || frontendBaseDir == "" {
		return path
	}

	if util.FileExist(filepath.Join(frontendBaseDir, "index.html")) {
		return frontendBaseDir
	}

	path = filepath.Join(frontendBaseDir, "web/build")
	return path
}

func fastAutoSignin(ctx *context.Context) (string, error) {
	userId := getSessionUser(ctx)
	if userId == "" {
		return "", nil
	}

	clientId := ctx.Input.Query("client_id")
	responseType := ctx.Input.Query("response_type")
	redirectUri := ctx.Input.Query("redirect_uri")
	scope := ctx.Input.Query("scope")
	state := ctx.Input.Query("state")
	nonce := ctx.Input.Query("nonce")
	codeChallenge := ctx.Input.Query("code_challenge")
	if clientId == "" || responseType != "code" || redirectUri == "" {
		return "", nil
	}

	application, err := object.GetApplicationByClientId(clientId)
	if err != nil {
		return "", err
	}
	if application == nil {
		return "", nil
	}

	if !application.EnableAutoSignin {
		return "", nil
	}

	isAllowed, err := object.CheckLoginPermission(userId, application)
	if err != nil {
		return "", err
	}

	if !isAllowed {
		return "", nil
	}

	user, err := object.GetUser(userId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", nil
	}

	consentRequired, err := object.CheckConsentRequired(user, application, scope)
	if err != nil {
		return "", err
	}

	if consentRequired {
		return "", nil
	}

	code, err := object.GetOAuthCode(userId, clientId, "", "autoSignin", responseType, redirectUri, scope, state, nonce, codeChallenge, "", ctx.Request.Host, getAcceptLanguage(ctx))
	if err != nil {
		return "", err
	} else if code.Message != "" {
		return "", fmt.Errorf(code.Message)
	}

	sep := "?"
	if strings.Contains(redirectUri, "?") {
		sep = "&"
	}
	res := fmt.Sprintf("%s%scode=%s&state=%s", redirectUri, sep, code.Code, state)
	return res, nil
}

func getProviderHintRedirectScriptPath() string {
	candidates := []string{
		filepath.Join(getWebBuildFolder(), "provider-hint-redirect.js"),
	}

	if frontendBaseDir != "" {
		candidates = append(candidates,
			filepath.Join(frontendBaseDir, "public", "provider-hint-redirect.js"),
			filepath.Join(filepath.Dir(frontendBaseDir), "public", "provider-hint-redirect.js"),
		)
	}

	candidates = append(candidates, filepath.Join("web", "public", "provider-hint-redirect.js"))

	for _, candidate := range candidates {
		if util.FileExist(candidate) {
			return candidate
		}
	}

	return ""
}

func serveProviderHintRedirectScript(ctx *context.Context) bool {
	if ctx.Request.URL.Path != "/provider-hint-redirect.js" {
		return false
	}

	scriptPath := getProviderHintRedirectScriptPath()
	if scriptPath == "" {
		ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "provider-hint-redirect.js", time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}

	f, err := os.Open(filepath.Clean(scriptPath))
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "provider-hint-redirect.js", time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "provider-hint-redirect.js", time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}

	ctx.Output.Header("Content-Type", "application/javascript; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, fileInfo.Name(), fileInfo.ModTime(), f)
	return true
}

func getAuthCallbackHandlerScriptPath() string {
	candidates := []string{
		filepath.Join(getWebBuildFolder(), "auth-callback-handler.js"),
	}

	if frontendBaseDir != "" {
		candidates = append(candidates,
			filepath.Join(frontendBaseDir, "public", "auth-callback-handler.js"),
			filepath.Join(filepath.Dir(frontendBaseDir), "public", "auth-callback-handler.js"),
		)
	}

	candidates = append(candidates, filepath.Join("web", "public", "auth-callback-handler.js"))

	for _, candidate := range candidates {
		if util.FileExist(candidate) {
			return candidate
		}
	}

	return ""
}

func serveAuthCallbackHandlerScript(ctx *context.Context) bool {
	if ctx.Request.URL.Path != "/auth-callback-handler.js" {
		return false
	}

	scriptPath := getAuthCallbackHandlerScriptPath()
	if scriptPath == "" {
		ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "auth-callback-handler.js", time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}

	f, err := os.Open(filepath.Clean(scriptPath))
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "auth-callback-handler.js", time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "auth-callback-handler.js", time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}

	ctx.Output.Header("Content-Type", "application/javascript; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, fileInfo.Name(), fileInfo.ModTime(), f)
	return true
}

func serveProviderHintRedirectPage(ctx *context.Context) bool {
	if ctx.Request.URL.Path != "/login/oauth/authorize" {
		return false
	}

	providerHint := ctx.Input.Query("provider_hint")
	if providerHint == "" {
		return false
	}

	const providerHintRedirectHtml = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Redirecting...</title>
	<style>
		html, body {
			width: 100%;
			height: 100%;
			margin: 0;
			background: #ffffff;
			color: #1f2937;
			font-family: sans-serif;
		}

		body {
			display: flex;
			align-items: center;
			justify-content: center;
		}

		.redirecting {
			font-size: 14px;
			opacity: 0.72;
		}
	</style>
</head>
<body>
	<div class="redirecting">Redirecting...</div>
	<script src="/provider-hint-redirect.js"></script>
	<script>
		(function() {
			function redirectToFallback() {
				var url = new URL(window.location.href);
				url.searchParams.delete("provider_hint");
				window.location.replace(url.pathname + url.search + url.hash);
			}

			if (!window.CasdoorProviderHintRedirect || typeof window.CasdoorProviderHintRedirect.run !== "function") {
				redirectToFallback();
				return;
			}

			window.CasdoorProviderHintRedirect.run();
		})();
	</script>
</body>
</html>
`

	ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, "provider-hint-redirect.html", time.Now(), strings.NewReader(providerHintRedirectHtml))
	return true
}

func serveAuthCallbackPage(ctx *context.Context) bool {
	if ctx.Request.URL.Path != "/callback" {
		return false
	}

	if ctx.Input.Query("__casdoor_callback_react") == "1" {
		return false
	}

	if ctx.Input.Query("state") == "" {
		return false
	}

	const authCallbackHtml = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>Signing in...</title>
	<style>
		html, body {
			width: 100%;
			height: 100%;
			margin: 0;
			background: #ffffff;
			color: #1f2937;
			font-family: sans-serif;
		}

		body {
			display: flex;
			align-items: center;
			justify-content: center;
		}

		.callback-status {
			font-size: 14px;
			opacity: 0.82;
			padding: 0 24px;
			text-align: center;
		}
	</style>
</head>
<body>
	<div id="callback-status" class="callback-status">Signing in...</div>
	<script src="/auth-callback-handler.js"></script>
	<script>
		(function() {
			if (!window.CasdoorAuthCallback || typeof window.CasdoorAuthCallback.run !== "function") {
				document.getElementById("callback-status").textContent = "Failed to load callback handler.";
				return;
			}

			window.CasdoorAuthCallback.run();
		})();
	</script>
</body>
</html>
`

	ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, "auth-callback.html", time.Now(), strings.NewReader(authCallbackHtml))
	return true
}

func StaticFilter(ctx *context.Context) {
	urlPath := ctx.Request.URL.Path

	if urlPath == "/.well-known/acme-challenge/filename" {
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "acme-challenge", time.Now(), strings.NewReader("content"))
	}

	if strings.HasPrefix(urlPath, "/api/") || strings.HasPrefix(urlPath, "/.well-known/") {
		return
	}
	if serveAuthCallbackHandlerScript(ctx) {
		return
	}
	if serveProviderHintRedirectScript(ctx) {
		return
	}
	if strings.HasPrefix(urlPath, "/cas") && (strings.HasSuffix(urlPath, "/serviceValidate") || strings.HasSuffix(urlPath, "/proxy") || strings.HasSuffix(urlPath, "/proxyValidate") || strings.HasSuffix(urlPath, "/validate") || strings.HasSuffix(urlPath, "/p3/serviceValidate") || strings.HasSuffix(urlPath, "/p3/proxyValidate") || strings.HasSuffix(urlPath, "/samlValidate")) {
		return
	}
	if strings.HasPrefix(urlPath, "/scim") {
		return
	}

	if urlPath == "/login/oauth/authorize" {
		redirectUrl, err := fastAutoSignin(ctx)
		if err != nil {
			responseError(ctx, err.Error())
			return
		}

		if redirectUrl != "" {
			http.Redirect(ctx.ResponseWriter, ctx.Request, redirectUrl, http.StatusFound)
			return
		}

		if serveProviderHintRedirectPage(ctx) {
			return
		}
	}

	if serveAuthCallbackPage(ctx) {
		return
	}

	webBuildFolder := getWebBuildFolder()
	path := webBuildFolder
	if urlPath == "/" {
		path += "/index.html"
	} else {
		path += urlPath
	}

	// Preventing synchronization problems from concurrency
	ctx.Input.CruSession = nil

	organizationThemeCookie, err := appendThemeCookie(ctx, urlPath)
	if err != nil {
		fmt.Println(err)
	}

	if strings.Contains(path, "/../") || !util.FileExist(path) {
		path = webBuildFolder + "/index.html"
	}
	if !util.FileExist(path) {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		dir = strings.ReplaceAll(dir, "\\", "/")
		ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		errorText := fmt.Sprintf("The Casdoor frontend HTML file: \"index.html\" was not found, it should be placed at: \"%s/web/build/index.html\". For more information, see: https://casdoor.org/docs/basic/server-installation/#frontend-1", dir)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, "Casdoor frontend has encountered error...", time.Now(), strings.NewReader(errorText))
		return
	}

	if oldStaticBaseUrl == newStaticBaseUrl {
		makeGzipResponse(ctx.ResponseWriter, ctx.Request, path, organizationThemeCookie)
	} else {
		serveFileWithReplace(ctx.ResponseWriter, ctx.Request, path, organizationThemeCookie)
	}
}

func serveFileWithReplace(w http.ResponseWriter, r *http.Request, name string, organizationThemeCookie *OrganizationThemeCookie) {
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		panic(err)
	}

	oldContent := util.ReadStringFromPath(name)
	newContent := oldContent
	if organizationThemeCookie != nil {
		newContent = strings.ReplaceAll(newContent, "https://cdn.casbin.org/img/favicon.png", organizationThemeCookie.Favicon)
		newContent = strings.ReplaceAll(newContent, "<title>Casdoor</title>", fmt.Sprintf("<title>%s</title>", organizationThemeCookie.DisplayName))
	}

	newContent = strings.ReplaceAll(newContent, oldStaticBaseUrl, newStaticBaseUrl)

	http.ServeContent(w, r, d.Name(), d.ModTime(), strings.NewReader(newContent))
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func makeGzipResponse(w http.ResponseWriter, r *http.Request, path string, organizationThemeCookie *OrganizationThemeCookie) {
	if !enableGzip || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		serveFileWithReplace(w, r, path, organizationThemeCookie)
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzip.NewWriter(w)
	defer gz.Close()
	gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}
	serveFileWithReplace(gzw, r, path, organizationThemeCookie)
}
