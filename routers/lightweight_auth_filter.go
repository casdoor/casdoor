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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/util"
)

const (
	providerHintRedirectScriptName = "ProviderHintRedirect.js"
	authCallbackHandlerScriptName  = "AuthCallbackHandler.js"
)

func getLightweightAuthScriptPath(scriptName string) string {
	candidates := []string{
		filepath.Join(getWebBuildFolder(), scriptName),
	}

	if frontendBaseDir != "" {
		candidates = append(candidates,
			filepath.Join(frontendBaseDir, "public", scriptName),
			filepath.Join(filepath.Dir(frontendBaseDir), "public", scriptName),
		)
	}

	candidates = append(candidates, filepath.Join("web", "public", scriptName))

	for _, candidate := range candidates {
		if util.FileExist(candidate) {
			return candidate
		}
	}

	return ""
}

func serveLightweightAuthScript(ctx *context.Context, requestPath string, scriptName string) bool {
	if ctx.Request.URL.Path != requestPath {
		return false
	}

	scriptPath := getLightweightAuthScriptPath(scriptName)
	if scriptPath == "" {
		ctx.ResponseWriter.WriteHeader(http.StatusNotFound)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, scriptName, time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}

	f, err := os.Open(filepath.Clean(scriptPath))
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, scriptName, time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		http.ServeContent(ctx.ResponseWriter, ctx.Request, scriptName, time.Now(), strings.NewReader("window.location.replace('/');"))
		return true
	}

	ctx.Output.Header("Content-Type", "application/javascript; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, fileInfo.Name(), fileInfo.ModTime(), f)
	return true
}

func serveProviderHintRedirectScript(ctx *context.Context) bool {
	return serveLightweightAuthScript(ctx, "/"+providerHintRedirectScriptName, providerHintRedirectScriptName)
}

func serveAuthCallbackHandlerScript(ctx *context.Context) bool {
	return serveLightweightAuthScript(ctx, "/"+authCallbackHandlerScriptName, authCallbackHandlerScriptName)
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
	<script src="/ProviderHintRedirect.js"></script>
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

	err := util.AppendWebConfigCookie(ctx)
	if err != nil {
		logs.Error("AppendWebConfigCookie failed in serveProviderHintRedirectPage, error: %s", err)
	}

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
	<script src="/AuthCallbackHandler.js"></script>
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

	err := util.AppendWebConfigCookie(ctx)
	if err != nil {
		logs.Error("AppendWebConfigCookie failed in serveAuthCallbackPage, error: %s", err)
	}

	ctx.Output.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Output.Header("Cache-Control", "no-store")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, "auth-callback.html", time.Now(), strings.NewReader(authCallbackHtml))
	return true
}
