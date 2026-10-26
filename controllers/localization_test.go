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

package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/pp"
	"github.com/casdoor/casdoor/util"
)

const localizationFixtureArgument = `<&%:mfa:provider>`

func localizationFixtureError() error {
	return util.MessageErrorf("mfa:RADIUS provider not found: %s", localizationFixtureArgument)
}

// Test-only methods exercise dispatch on the actual controller types, without a DB or server listener.
func (c *ApiController) LocalizationFixture() {
	err := localizationFixtureError()
	c.Ctx.Output.Header("X-Localization-Fixture", "unchanged")
	switch c.Ctx.Input.Query("mode") {
	case "from":
		c.ResponseErrorFromWithCode("fixture-code", err, "mfa:opaque data", map[string]string{"id": "zh-TW"})
	case "opaque":
		c.ResponseErrorFrom(errors.New("mfa:RADIUS provider not found: opaque"))
	case "outer":
		c.ResponseErrorFrom(fmt.Errorf("outer context: %w", err))
	case "signin":
		c.responseSigninError(err, object.SigninReasonMfaFailed)
	case "idp-child":
		c.ResponseErrorFrom(util.MessageErrorf("auth:Failed to login in: %s", util.NewMessageError("provider:id_token has expired")))
	case "invoice":
		c.responseInvoiceError(pp.ErrInvoicePending)
	case "protocol":
		c.Data["json"] = map[string]string{"error": "invalid_grant", "error_description": "mfa:opaque protocol prose"}
		_ = c.ServeJSON()
	default:
		response := wrapErrorResponse(err)
		if c.Ctx.Input.Query("mode") == "action" {
			response = wrapActionResponse(false, err)
		}
		response.Code, response.Sub, response.Name = "fixture-code", "subject-id", "fixture-name"
		response.Data, response.Data2, response.Data3 = "mfa:opaque data", 2, "raw-third-value"
		c.Data["json"] = response
		_ = c.ServeJSON()
	}
}

func (c *RootController) LocalizationFixture() { c.ApiController.LocalizationFixture() }

func localizationController(t *testing.T, language string) (*ApiController, *httptest.ResponseRecorder) {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, "http://localhost/__localization", nil)
	request.Header.Set("Accept-Language", language)
	recorder := httptest.NewRecorder()
	ctx := context.NewContext()
	ctx.Reset(recorder, request)
	controller := &ApiController{}
	controller.Init(ctx, "ApiController", "LocalizationFixture", controller)
	return controller, recorder
}

func TestLocalizationRoutedControllers(t *testing.T) {
	t.Setenv("enableErrorMask2", "false")
	t.Setenv("forceLanguage", "")
	t.Setenv("defaultLanguage", "en")
	config := *web.BConfig
	config.RunMode = web.PROD
	server := web.NewHttpServerWithCfg(&config)
	api, root := &ApiController{}, &RootController{}
	server.Handlers.Add("/__localization/api", api, web.WithRouterMethods(api, "get:LocalizationFixture"))
	server.Handlers.Add("/__localization/root", root, web.WithRouterMethods(root, "get:LocalizationFixture"))
	for _, path := range []string{"/__localization/api", "/__localization/root"} {
		for _, language := range []string{"en", "zh-TW"} {
			for _, mode := range []string{"from", "wrap", "action", "opaque", "outer", "signin", "idp-child", "invoice", "protocol"} {
				t.Run(path+"/"+language+"/"+mode, func(t *testing.T) {
					request := httptest.NewRequest(http.MethodGet, path+"?mode="+mode, nil)
					request.Header.Set("Accept-Language", language)
					recorder := httptest.NewRecorder()
					server.Handlers.ServeHTTP(recorder, request)
					if recorder.Code != http.StatusOK || recorder.Header().Get("X-Localization-Fixture") != "unchanged" {
						t.Fatalf("dispatch/status/header changed: %d %s", recorder.Code, recorder.Body.String())
					}
					if mode == "protocol" {
						var value map[string]string
						if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
							t.Fatal(err)
						}
						if !reflect.DeepEqual(value, map[string]string{"error": "invalid_grant", "error_description": "mfa:opaque protocol prose"}) {
							t.Fatalf("protocol object changed: %v", value)
						}
						return
					}
					var value Response
					if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
						t.Fatalf("invalid JSON: %v: %s", err, recorder.Body.String())
					}
					want := "RADIUS provider not found: " + localizationFixtureArgument
					if language == "zh-TW" {
						want = "找不到 RADIUS 服務供應商：" + localizationFixtureArgument
					}
					if mode == "opaque" {
						want = "mfa:RADIUS provider not found: opaque"
					}
					if mode == "outer" {
						want = "outer context: RADIUS provider not found: " + localizationFixtureArgument
					}
					if mode == "idp-child" {
						want = "Failed to login in: id_token has expired"
						if language == "zh-TW" {
							want = "無法登入：id_token 已到期"
						}
					}
					if mode == "invoice" {
						want = "Request accepted. The invoice is being issued"
						if language == "zh-TW" {
							want = "申請成功，發票開立中"
						}
						if value.Code != "invoice-pending" {
							t.Fatalf("invoice state was classified using prose: %#v", value)
						}
					}
					if value.Status != "error" || value.Msg != want {
						t.Fatalf("wrong presentation: %#v, want %q", value, want)
					}
					if mode == "from" || mode == "wrap" || mode == "action" {
						if value.Code != "fixture-code" || value.Data != "mfa:opaque data" {
							t.Fatalf("wire fields changed: %#v", value)
						}
					}
					if mode == "wrap" || mode == "action" {
						if value.Sub != "subject-id" || value.Name != "fixture-name" || value.Data2 != float64(2) || value.Data3 != "raw-third-value" {
							t.Fatalf("metadata changed: %#v", value)
						}
					}
					if mode == "signin" && value.Code != object.SigninReasonMfaFailed {
						t.Fatalf("reason code changed: %q", value.Code)
					}
				})
			}
		}
	}
}

func TestLocalizationOriginalResponseAndRecord(t *testing.T) {
	t.Setenv("enableErrorMask2", "false")
	controller, recorder := localizationController(t, "zh-TW")
	original := wrapErrorResponse(localizationFixtureError())
	controller.Data["json"] = original
	if err := controller.ServeJSON(); err != nil {
		t.Fatal(err)
	}
	if controller.Data["json"] != original || original.Msg != localizationFixtureError().Error() {
		t.Fatal("diagnostic original was replaced or modified")
	}
	record, err := object.NewRecord(controller.Ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(record.Response, original.Msg) || strings.Contains(record.Response, "找不到") {
		t.Fatalf("record diagnostic changed: %s", record.Response)
	}
	var displayed Response
	if err = json.Unmarshal(recorder.Body.Bytes(), &displayed); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(displayed.Msg, "找不到") {
		t.Fatalf("presentation was not localized: %q", displayed.Msg)
	}

	broken, _ := localizationController(t, "zh-TW")
	bad := wrapErrorResponse(localizationFixtureError())
	bad.Data = make(chan int)
	broken.Data["json"] = bad
	if err = broken.ServeJSON(); err == nil {
		t.Fatal("marshal failure was swallowed")
	}
	if broken.Data["json"] != bad || bad.Msg != localizationFixtureError().Error() {
		t.Fatal("marshal failure did not restore the diagnostic original")
	}
}

func TestLocalizationEncodingAndNonErrorPassThrough(t *testing.T) {
	t.Setenv("enableErrorMask2", "false")
	for _, encoding := range []bool{false, true} {
		controller, recorder := localizationController(t, "en")
		baseline, expected := localizationController(t, "en")
		controller.Data["json"] = wrapErrorResponse(localizationFixtureError())
		baseline.Data["json"] = wrapErrorResponse(localizationFixtureError())
		if err := controller.ServeJSON(encoding); err != nil {
			t.Fatal(err)
		}
		if err := baseline.Controller.ServeJSON(encoding); err != nil {
			t.Fatal(err)
		}
		if recorder.Body.String() != expected.Body.String() {
			t.Fatalf("encoding option changed: %t", encoding)
		}
	}
	for _, value := range []interface{}{nil, Response{Status: "error", Msg: "opaque value response"}, &Response{Status: "ok", Msg: "opaque success", cause: localizationFixtureError()}, map[string]string{"error": "invalid_grant"}} {
		controller, recorder := localizationController(t, "zh-TW")
		baseline, expected := localizationController(t, "zh-TW")
		controller.Data["json"], baseline.Data["json"] = value, value
		if err := controller.ServeJSON(); err != nil {
			t.Fatal(err)
		}
		if err := baseline.Controller.ServeJSON(); err != nil {
			t.Fatal(err)
		}
		if recorder.Body.String() != expected.Body.String() {
			t.Fatalf("noneligible response changed: %T", value)
		}
	}
}

func TestLocalizationMaskAndSigninRecordReason(t *testing.T) {
	for _, masked := range []bool{false, true} {
		t.Run(fmt.Sprint(masked), func(t *testing.T) {
			t.Setenv("enableErrorMask2", fmt.Sprint(masked))
			controller, recorder := localizationController(t, "zh-TW")
			controller.responseSigninError(localizationFixtureError(), object.SigninReasonMfaFailed)
			if reason := controller.Ctx.Input.Param("recordFailureReason"); reason != object.SigninReasonMfaFailed {
				t.Fatalf("production recording reason lost: %q", reason)
			}
			var value Response
			if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
				t.Fatal(err)
			}
			if masked {
				if value.Code != "" || strings.Contains(value.Msg, localizationFixtureArgument) || strings.Contains(value.Msg, "RADIUS") {
					t.Fatalf("mask was undone: %#v", value)
				}
			} else if value.Code != object.SigninReasonMfaFailed || !strings.Contains(value.Msg, localizationFixtureArgument) {
				t.Fatalf("unmasked presentation changed: %#v", value)
			}

			wrapped, output := localizationController(t, "zh-TW")
			response := wrapActionResponse(false, localizationFixtureError())
			response.Code = "fixture-code"
			wrapped.Data["json"] = response
			if err := wrapped.ServeJSON(); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(output.Body.Bytes(), &value); err != nil {
				t.Fatal(err)
			}
			if masked && (value.Code != "" || strings.Contains(value.Msg, localizationFixtureArgument)) {
				t.Fatalf("free-wrapper mask was undone: %#v", value)
			}
			if wrapped.Data["json"] != response || response.Code != "fixture-code" {
				t.Fatal("mask modified the original")
			}
		})
	}
}
