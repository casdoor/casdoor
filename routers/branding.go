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
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
)

const webManifestPath = "/manifest.webmanifest"

var (
	reDescriptionMeta = regexp.MustCompile(`<meta name="description" content="[^"]*"\s*/?>`)
	reManifestLink    = regexp.MustCompile(`<link rel="manifest" href="[^"]*"\s*/?>`)
)

// addBrandingToHtml puts the application's name, description, icon and manifest into index.html, so link previews
// and the home-screen / PWA surface, which read the raw HTML without running the app, show them instead of Casdoor's
func addBrandingToHtml(content string, organizationThemeCookie *OrganizationThemeCookie) string {
	title := html.EscapeString(organizationThemeCookie.DisplayName)
	description := html.EscapeString(organizationThemeCookie.Description)
	favicon := html.EscapeString(organizationThemeCookie.Favicon)

	if favicon != "" {
		content = strings.ReplaceAll(content, "https://cdn.casbin.org/img/favicon.png", favicon)
	}
	if title != "" {
		content = strings.ReplaceAll(content, "<title>Casdoor</title>", fmt.Sprintf("<title>%s</title>", title))
	}
	if description != "" {
		content = reDescriptionMeta.ReplaceAllLiteralString(content, fmt.Sprintf(`<meta name="description" content="%s" />`, description))
	}
	if organizationThemeCookie.Application != "" {
		manifestUrl := fmt.Sprintf("%s?application=%s", webManifestPath, url.QueryEscape(organizationThemeCookie.Application))
		content = reManifestLink.ReplaceAllLiteralString(content, fmt.Sprintf(`<link rel="manifest" href="%s" />`, html.EscapeString(manifestUrl)))
	}

	image := organizationThemeCookie.LogoUrl
	if image == "" {
		image = organizationThemeCookie.Favicon
	}
	metas := []string{`<meta property="og:type" content="website" />`}
	if title != "" {
		metas = append(metas, fmt.Sprintf(`<meta property="og:title" content="%s" />`, title))
	}
	if description != "" {
		metas = append(metas, fmt.Sprintf(`<meta property="og:description" content="%s" />`, description))
	}
	if image != "" {
		metas = append(metas, fmt.Sprintf(`<meta property="og:image" content="%s" />`, html.EscapeString(image)))
	}
	return strings.Replace(content, "</head>", strings.Join(metas, "")+"</head>", 1)
}

type webManifestIcon struct {
	Src   string `json:"src"`
	Sizes string `json:"sizes"`
}

type webManifest struct {
	ShortName       string             `json:"short_name"`
	Name            string             `json:"name"`
	Description     string             `json:"description,omitempty"`
	Icons           []*webManifestIcon `json:"icons,omitempty"`
	StartUrl        string             `json:"start_url"`
	Display         string             `json:"display"`
	ThemeColor      string             `json:"theme_color"`
	BackgroundColor string             `json:"background_color"`
}

// serveWebManifest serves the manifest that index.html links to on a branded page, built from the application and
// its organization
func serveWebManifest(ctx *context.Context) bool {
	if ctx.Request.URL.Path != webManifestPath {
		return false
	}

	application, err := object.GetApplication(ctx.Input.Query("application"))
	if err != nil {
		responseErrorFrom(ctx, err)
		return true
	}
	if application == nil {
		http.NotFound(ctx.ResponseWriter, ctx.Request)
		return true
	}

	organizationThemeCookie, err := getOrganizationThemeCookieFromApplication(application)
	if err != nil {
		responseErrorFrom(ctx, err)
		return true
	}

	themeColor := "#000000"
	if application.ThemeData != nil && application.ThemeData.IsEnabled && application.ThemeData.ColorPrimary != "" {
		themeColor = application.ThemeData.ColorPrimary
	}

	manifest := &webManifest{
		ShortName:       organizationThemeCookie.DisplayName,
		Name:            organizationThemeCookie.DisplayName,
		Description:     organizationThemeCookie.Description,
		StartUrl:        ".",
		Display:         "standalone",
		ThemeColor:      themeColor,
		BackgroundColor: "#ffffff",
	}
	if manifest.Name == "" {
		manifest.ShortName = application.Name
		manifest.Name = application.Name
	}
	if organizationThemeCookie.Favicon != "" {
		manifest.Icons = []*webManifestIcon{{Src: organizationThemeCookie.Favicon, Sizes: "any"}}
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		responseErrorFrom(ctx, err)
		return true
	}

	ctx.ResponseWriter.Header().Set("Content-Type", "application/manifest+json")
	ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(ctx.ResponseWriter, ctx.Request, "manifest.webmanifest", time.Time{}, strings.NewReader(string(data)))
	return true
}
