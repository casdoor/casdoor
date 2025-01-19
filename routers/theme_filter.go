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

func appendThemeCookie(ctx *context.Context, urlPath string) error {
	if urlPath == "/login" {
		application, err := object.GetDefaultApplication(fmt.Sprintf("admin/built-in"))
		if err != nil {
			return err
		}
		if application.ThemeData != nil {
			return setThemeDataCookie(ctx, application.ThemeData, application.Logo, application.FooterHtml)
		}
		organization := application.OrganizationObj
		if organization == nil {
			organization, err = object.GetOrganization(fmt.Sprintf("admin/built-in"))
			if err != nil {
				return err
			}
		}
		if organization != nil {
			return setThemeDataCookie(ctx, organization.ThemeData, organization.Logo, application.FooterHtml)
		}
	} else if strings.HasPrefix(urlPath, "/login/oauth/authorize") {
		clientId := ctx.Input.Query("client_id")
		if clientId == "" {
			return nil
		}
		application, err := object.GetApplicationByClientId(clientId)
		if err != nil {
			return err
		}
		if application != nil {
			organization, err := object.GetOrganization(fmt.Sprintf("admin/%s", application.Organization))
			if err != nil {
				return err
			}
			if application.ThemeData != nil {
				return setThemeDataCookie(ctx, application.ThemeData, application.Logo, application.FooterHtml)
			}
			if organization != nil {
				return setThemeDataCookie(ctx, organization.ThemeData, organization.Logo, application.FooterHtml)
			}
		}
	} else if strings.HasPrefix(urlPath, "/login/") {
		owner := strings.Replace(urlPath, "/login/", "", -1)
		if owner != "undefined" && owner != "oauth/undefined" {
			application, err := object.GetDefaultApplication(fmt.Sprintf("admin/%s", owner))
			if err != nil {
				return err
			}
			if application.ThemeData != nil {
				return setThemeDataCookie(ctx, application.ThemeData, application.Logo, application.FooterHtml)
			}
			organization := application.OrganizationObj
			if organization == nil {
				organization, err = object.GetOrganization(fmt.Sprintf("admin/%s", owner))
				if err != nil {
					return err
				}
			}
			if organization != nil {
				return setThemeDataCookie(ctx, organization.ThemeData, organization.Logo, application.FooterHtml)
			}
		}
	}

	return nil
}

func setThemeDataCookie(ctx *context.Context, themeData *object.ThemeData, logoUrl string, footerHtml string) error {
	themeDataString, err := json.Marshal(themeData)
	if err != nil {
		return err
	}
	ctx.SetCookie("organizationTheme", string(themeDataString))
	ctx.SetCookie("organizationLogo", logoUrl)
	ctx.SetCookie("organizationFootHtml", footerHtml)
	return nil
}
