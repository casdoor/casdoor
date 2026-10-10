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

package service

import (
	"net/http"
	"net/url"
	"strings"
)

func getForwardedRequest(r *http.Request) (string, string, string) {
	if originalUrl := r.Header.Get("X-Original-URL"); originalUrl != "" {
		u, err := url.Parse(originalUrl)
		if err == nil && u.Host != "" {
			scheme := "http"
			if strings.EqualFold(u.Scheme, "https") {
				scheme = "https"
			}
			return scheme, u.Host, u.RequestURI()
		}
	}

	scheme := "http"
	if strings.EqualFold(strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Proto"), ",")[0]), "https") {
		scheme = "https"
	}

	host := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0])

	uri := r.Header.Get("X-Forwarded-Uri")
	if !strings.HasPrefix(uri, "/") {
		uri = "/"
	}
	return scheme, host, uri
}

// HandleForwardAuth answers the auth subrequest of a reverse proxy (Traefik forwardAuth, Caddy forward_auth,
// Nginx auth_request) for the site whose domain is the forwarded host. Nginx sends X-Original-URL and gets
// 401 with a Location header instead of a redirect.
func HandleForwardAuth(w http.ResponseWriter, r *http.Request) {
	scheme, host, uri := getForwardedRequest(r)
	if host == "" {
		w.WriteHeader(http.StatusBadRequest)
		responseErrorWithoutCode(w, "Casdoor forward auth error: the X-Forwarded-Host or X-Original-URL header is missing")
		return
	}

	site := getSiteByDomainWithWww(host)
	if site == nil || site.CasdoorApplication == "" {
		responseForbidden(w, "Casdoor forward auth error: no site with a Casdoor application is found for host: %s", host)
		return
	}

	forwardedUrl, err := url.Parse(uri)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		responseErrorWithoutCode(w, "Casdoor forward auth error: invalid URI: %s", uri)
		return
	}
	if forwardedUrl.Path == "/caswaf-handler" {
		handleSiteCallback(w, r, site, scheme, forwardedUrl.Query())
		return
	}

	identity, result, msg := authenticateSiteRequest(site, r)
	switch result {
	case siteAuthNeedLogin:
		casdoorClient, err := getCasdoorClientFromSite(site)
		if err != nil {
			responseError(w, "Casdoor forward auth error: getCasdoorClientFromSite() error: %s", err.Error())
			return
		}
		if _, err = r.Cookie("casdoor_access_token"); err == nil {
			clearAccessTokenCookie(w)
		}

		signinUrl := getSiteSigninUrl(casdoorClient, scheme, host, uri)
		if r.Header.Get("X-Original-URL") != "" {
			w.Header().Set("Location", signinUrl)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, signinUrl, http.StatusFound)
		return
	case siteAuthForbidden:
		responseForbidden(w, "Forbidden: %s", msg)
		return
	case siteAuthError:
		responseError(w, "Casdoor forward auth error: %s", msg)
		return
	}

	for header, value := range getUserHeaderValues(identity) {
		w.Header().Set(header, value)
	}
	w.WriteHeader(http.StatusOK)
}
