// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"strings"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/object"
)

type OrganizationThemeCookie struct {
	ThemeData   *object.ThemeData
	LogoUrl     string
	FooterHtml  string
	Favicon     string
	DisplayName string
}

func appendThemeCookie(ctx *context.Context, urlPath string) (*OrganizationThemeCookie, error) {
	organizationThemeCookie, err := getOrganizationThemeCookieFromUrlPath(ctx, urlPath)
	if err != nil {
		return nil, err
	}
	if organizationThemeCookie != nil {
		return organizationThemeCookie, setThemeDataCookie(ctx, organizationThemeCookie)
	}

	return nil, nil
}

func getOrganizationThemeCookieFromUrlPath(ctx *context.Context, urlPath string) (*OrganizationThemeCookie, error) {
	var application *object.Application
	var organization *object.Organization
	var err error
	if urlPath == "/login" || urlPath == "/signup" {
		application, err = object.GetDefaultApplication(fmt.Sprintf("admin/built-in"))
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(urlPath, "/oauth/authorize") {
		clientId := ctx.Input.Query("client_id")
		if clientId == "" {
			return nil, nil
		}
		application, err = object.GetApplicationByClientId(clientId)
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(urlPath, "/login/saml") {
		owner, _ := strings.CutPrefix(urlPath, "/login/saml/authorize/")
		application, err = object.GetApplication(owner)
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(urlPath, "/login/") {
		owner, _ := strings.CutPrefix(urlPath, "/login/")
		if owner == "undefined" || strings.Count(owner, "/") > 0 {
			return nil, nil
		}
		application, err = object.GetDefaultApplication(fmt.Sprintf("admin/%s", owner))
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(urlPath, "/signup/") {
		owner, _ := strings.CutPrefix(urlPath, "/signup/")
		if owner == "undefined" || strings.Count(owner, "/") > 0 {
			return nil, nil
		}
		application, err = object.GetDefaultApplication(fmt.Sprintf("admin/%s", owner))
		if err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(urlPath, "/cas/") && strings.HasSuffix(urlPath, "/login") {
		owner, _ := strings.CutPrefix(urlPath, "/cas/")
		owner, _ = strings.CutSuffix(owner, "/login")
		application, err = object.GetApplication(owner)
		if err != nil {
			return nil, err
		}
	}

	if application == nil {
		return nil, nil
	}
	organization = application.OrganizationObj
	if organization == nil {
		organization, err = object.GetOrganization(fmt.Sprintf("admin/%s", application.Organization))
		if err != nil {
			return nil, err
		}
	}

	organizationThemeCookie := &OrganizationThemeCookie{
		ThemeData:  application.ThemeData,
		LogoUrl:    application.Logo,
		FooterHtml: application.FooterHtml,
	}

	if organization != nil {
		organizationThemeCookie.Favicon = organization.Favicon
		organizationThemeCookie.DisplayName = organization.DisplayName
	}

	return organizationThemeCookie, nil
}

func setThemeDataCookie(ctx *context.Context, organizationThemeCookie *OrganizationThemeCookie) error {
	themeDataString, err := json.Marshal(organizationThemeCookie.ThemeData)
	if err != nil {
		return err
	}
	ctx.SetCookie("organizationTheme", string(themeDataString))
	ctx.SetCookie("organizationLogo", organizationThemeCookie.LogoUrl)
	ctx.SetCookie("organizationFootHtml", organizationThemeCookie.FooterHtml)
	return nil
}
