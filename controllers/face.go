// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

// Casdoor will expose its providers as services to SDK
// We are going to implement those services as APIs here

package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/casdoor/casdoor/faceId"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type faceIdDetectImageRequest struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Application string `json:"application"`
	Image       string `json:"image"`
}

// FaceIDSigninBegin
// @Title FaceIDSigninBegin
// @Tag Login API
// @Description FaceId Login Flow 1st stage
// @Param   owner     query    string  true        "owner"
// @Param   name     query    string  true        "name"
// @Success 200 {object} controllers.Response The Response object
// @router /faceid-signin-begin [get]
func (c *ApiController) FaceIDSigninBegin() {
	userOwner := c.Ctx.Input.Query("owner")
	userName := c.Ctx.Input.Query("name")

	user, err := object.GetUserByFields(userOwner, userName)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if user == nil {
		c.ResponseError(fmt.Sprintf(c.T("general:The user: %s doesn't exist"), util.GetId(userOwner, userName)))
		return
	}

	if len(user.FaceIds) == 0 {
		c.ResponseError(c.T("check:Face data does not exist, cannot log in"))
		return
	}

	c.ResponseOk()
}

// DetectFaceIdImage
// @Title DetectFaceIdImage
// @Tag Login API
// @Description Detect whether a captured Face ID image is valid by using the configured Local UniFace provider
// @Param   body     body    controllers.faceIdDetectImageRequest  true        "Face image detect request"
// @Success 200 {object} controllers.Response The Response object
// @router /detect-faceid-image [post]
func (c *ApiController) DetectFaceIdImage() {
	var request faceIdDetectImageRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		c.ResponseError(err.Error())
		return
	}

	if request.Image == "" {
		c.ResponseError(c.T("check:Image cannot be empty"))
		return
	}

	applicationName := request.Application
	if applicationName == "" {
		applicationName = "app-built-in"
	}
	applicationId := applicationName
	if !strings.Contains(applicationId, "/") {
		applicationId = fmt.Sprintf("admin/%s", applicationName)
	}

	application, err := object.GetApplication(applicationId)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if application == nil {
		c.ResponseError(fmt.Sprintf(c.T("auth:The application: %s does not exist"), applicationName))
		return
	}

	provider, err := object.GetFaceIdProviderByApplication(util.GetId(application.Owner, application.Name), "false", c.GetAcceptLanguage())
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if provider == nil || provider.Type != "Local UniFace" {
		c.ResponseError("Local UniFace Face ID provider is not configured")
		return
	}

	localProvider := faceId.NewLocalUniFaceProvider(provider.Endpoint, provider.ClientSecret)
	faces, err := localProvider.Detect(request.Image)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if len(faces) == 0 {
		c.ResponseError(c.T("check:Please ensure sufficient lighting and align your face in the center of the recognition box"))
		return
	}
	if len(faces) > 1 {
		c.ResponseError(c.T("check:Please keep only one face in the recognition box"))
		return
	}

	c.ResponseOk(map[string]interface{}{"faces": faces})
}
