package routers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beego/beego/context"
	"github.com/casdoor/casdoor/object"
)

func appendThemeCookie(ctx *context.Context, urlPath string) {
	if urlPath == "/login" {
		organization, err := object.GetOrganization(fmt.Sprintf("admin/built-in"))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if organization != nil {
			setThemeDataCookie(ctx, organization.ThemeData)
		}
	} else if strings.HasPrefix(urlPath, "/login/oauth/authorize") {
		clientId := ctx.Input.Query("client_id")
		if clientId == "" {
			return
		}
		application, err := object.GetApplicationByClientId(clientId)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if application != nil {
			organization, err := object.GetOrganization(fmt.Sprintf("admin/%s", application.Organization))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if organization != nil {
				setThemeDataCookie(ctx, organization.ThemeData)
			}
		}
	} else if strings.HasPrefix(urlPath, "/login/") {
		owner := strings.Replace(urlPath, "/login/", "", -1)
		if owner != "undefined" && owner != "oauth/undefined" {
			organization, err := object.GetOrganization(fmt.Sprintf("admin/%s", owner))
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if organization != nil {
				setThemeDataCookie(ctx, organization.ThemeData)
			}
		}
	}
}

func setThemeDataCookie(ctx *context.Context, themeData *object.ThemeData) {
	themeDataString, err := json.Marshal(themeData)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ctx.SetCookie("organizationTheme", string(themeDataString))
}
