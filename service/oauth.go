// Copyright 2023 The casbin Authors. All Rights Reserved.
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
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

func getSigninUrl(casdoorClient *casdoorsdk.Client, callbackUrl string, originalPath string) string {
	scope := "read"
	return fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&response_type=code&redirect_uri=%s&scope=%s&state=%s",
		casdoorClient.Endpoint, casdoorClient.ClientId, url.QueryEscape(callbackUrl), scope, url.QueryEscape(originalPath))
}

func redirectToCasdoor(casdoorClient *casdoorsdk.Client, w http.ResponseWriter, r *http.Request) {
	scheme := getScheme(r)

	callbackUrl := fmt.Sprintf("%s://%s/caswaf-handler", scheme, r.Host)
	originalPath := r.RequestURI
	signinUrl := getSigninUrl(casdoorClient, callbackUrl, originalPath)
	http.Redirect(w, r, signinUrl, http.StatusFound)
}

func checkSiteAccessToken(casdoorClient *casdoorsdk.Client, accessToken string) error {
	claims, err := casdoorClient.ParseJwtToken(accessToken)
	if err != nil {
		return err
	}

	if claims.IsRefreshToken() || claims.TokenType == "refresh-token" {
		return fmt.Errorf("a refresh token cannot be used as the access token")
	}
	if !slices.Contains(claims.Audience, casdoorClient.ClientId) {
		return fmt.Errorf("the access token is not issued to the application: %s", casdoorClient.ApplicationName)
	}
	return nil
}

func getSafeRedirectPath(state string) string {
	if !strings.HasPrefix(state, "/") || strings.HasPrefix(state, "//") || strings.HasPrefix(state, "/\\") {
		return "/"
	}
	return state
}

func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	site := getSiteByDomainWithWww(r.Host)
	if site == nil {
		responseError(w, "CasWAF error: site not found for host: %s", r.Host)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if code == "" {
		responseError(w, "CasWAF error: the code should not be empty")
		return
	} else if state == "" {
		responseError(w, "CasWAF error: the state should not be empty")
		return
	}

	application, err := object.GetApplication(util.GetId(site.Owner, site.CasdoorApplication))
	if err != nil {
		responseError(w, "CasWAF error: casdoorClient.GetOAuthToken() error: %s", err.Error())
		return
	}

	//casdoorClient, err := getCasdoorClientFromSite(site)
	//if err != nil {
	//	responseError(w, "CasWAF error: getCasdoorClientFromSite() error: %s", err.Error())
	//	return
	//}

	token, tokenError, err := object.GetAuthorizationCodeToken(application, application.ClientSecret, code, "", "", conf.GetAcceptLanguage(r.Header.Get("Accept-Language")))
	if tokenError != nil {
		responseError(w, "CasWAF error: casdoorClient.GetOAuthToken() error: %s", tokenError.Error)
		return
	}
	if err != nil {
		responseError(w, "CasWAF error: casdoorClient.GetOAuthToken() error: %s", err.Error())
		return
	}

	cookie := &http.Cookie{
		Name:  "casdoor_access_token",
		Value: token.AccessToken,
		Path:  "/",
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, getSafeRedirectPath(state), http.StatusFound)
}
