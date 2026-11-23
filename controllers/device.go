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
	"fmt"
	"io"

	"github.com/beego/beego/v2/server/web/pagination"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

// readDeviceBody reads the body of a request of the device agent and authenticates the device that signed it
func (c *ApiController) readDeviceBody() (*object.Device, []byte, bool) {
	body := c.Ctx.Input.RequestBody
	if len(body) == 0 && c.Ctx.Request.Body != nil {
		body, _ = io.ReadAll(io.LimitReader(c.Ctx.Request.Body, 1<<20))
	}
	device, err := object.AuthenticateDeviceRequest(c.Ctx.Request, body)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, nil, false
	}
	return device, body, true
}

func getDevicePolicyInfo(device *object.Device) map[string]interface{} {
	policy := object.GetDeviceTrustPolicy(device.Owner)
	if policy == nil {
		return map[string]interface{}{"deviceTrust": ""}
	}
	return map[string]interface{}{
		"deviceTrust":      policy.Mode,
		"requireEncrypted": policy.RequireEncrypted,
		"requireFirewall":  policy.RequireFirewall,
		"requireLock":      policy.RequireLock,
		"passwordless":     policy.Passwordless,
	}
}

// EnrollDevice
// @Title EnrollDevice
// @Tag Device API
// @Description start the enrollment of a device by the device agent, which shows the returned code for a signed-in user to approve
// @Param   body    body   object.DeviceEnrollRequest  true        "The device and its public key"
// @Success 200 {object} controllers.Response The Response object
// @router /device/enroll [post]
func (c *ApiController) EnrollDevice() {
	var request object.DeviceEnrollRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &request); err != nil {
		c.ResponseError(err.Error())
		return
	}
	device, err := object.StartDeviceEnrollment(&request, util.GetClientIpFromRequest(c.Ctx.Request))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(map[string]interface{}{
		"device":          device.GetId(),
		"enrollCode":      device.EnrollCode,
		"expireTime":      device.EnrollExpireTime,
		"verificationUri": "/device-enroll?code=" + device.EnrollCode + "&organization=" + device.Owner,
	})
}

// GetDeviceStatus
// @Title GetDeviceStatus
// @Tag Device API
// @Description tell the device agent whether its device is approved and what the organization requires, signed by the device
// @Success 200 {object} controllers.Response The Response object
// @router /device/status [post]
func (c *ApiController) GetDeviceStatus() {
	device, _, ok := c.readDeviceBody()
	if !ok {
		return
	}
	c.ResponseOk(map[string]interface{}{
		"device":      device.GetId(),
		"status":      device.Status,
		"user":        device.User,
		"isCompliant": device.IsCompliant,
		"problems":    device.Problems,
	}, getDevicePolicyInfo(device))
}

type deviceReportForm struct {
	Posture      *object.DevicePosture `json:"posture"`
	AgentVersion string                `json:"agentVersion"`
}

// ReportDevice
// @Title ReportDevice
// @Tag Device API
// @Description report the posture of an enrolled device, signed by the device
// @Param   body    body   controllers.deviceReportForm  true        "The posture"
// @Success 200 {object} controllers.Response The Response object
// @router /device/report [post]
func (c *ApiController) ReportDevice() {
	device, body, ok := c.readDeviceBody()
	if !ok {
		return
	}
	var form deviceReportForm
	if err := json.Unmarshal(body, &form); err != nil {
		c.ResponseError(err.Error())
		return
	}
	if err := object.ReportDevicePosture(device, form.Posture, form.AgentVersion, util.GetClientIpFromRequest(c.Ctx.Request)); err != nil {
		c.ResponseError(err.Error())
		return
	}
	device, err := object.GetDevice(device.GetId())
	if err != nil || device == nil {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return
	}
	c.ResponseOk(map[string]interface{}{"status": device.Status, "user": device.User, "isCompliant": device.IsCompliant, "problems": device.Problems}, getDevicePolicyInfo(device))
}

// GetDeviceChallenge
// @Title GetDeviceChallenge
// @Tag Device API
// @Description get a one-time nonce for the device agent to sign on the sign-in page
// @Success 200 {object} controllers.Response The Response object
// @router /device/challenge [get]
func (c *ApiController) GetDeviceChallenge() {
	c.ResponseOk(object.NewDeviceChallenge())
}

// GetDeviceEnrollment
// @Title GetDeviceEnrollment
// @Tag Device API
// @Description get the device that waits for the enrollment code, for the signed-in user to see what they approve
// @Param   code     query    string  true        "The enrollment code"
// @Success 200 {object} object.Device The Response object
// @router /get-device-enrollment [get]
func (c *ApiController) GetDeviceEnrollment() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}
	device, err := object.GetDeviceEnrollment(user.Owner, c.Ctx.Input.Query("code"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	if device == nil {
		c.ResponseError(c.T("enterprise:The enrollment code is wrong or has expired"))
		return
	}
	c.ResponseOk(device)
}

// ApproveDeviceEnrollment
// @Title ApproveDeviceEnrollment
// @Tag Device API
// @Description bind the device that waits for the enrollment code to the signed-in user
// @Param   code     query    string  true        "The enrollment code"
// @Success 200 {object} object.Device The Response object
// @router /approve-device-enrollment [post]
func (c *ApiController) ApproveDeviceEnrollment() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}
	device, err := object.ApproveDeviceEnrollment(user, c.Ctx.Input.Query("code"))
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(device)
}

// GetDevices
// @Title GetDevices
// @Tag Device API
// @Description list the devices of an organization, or of a user of it
// @Param   owner     query    string  true        "The organization"
// @Param   user      query    string  false       "The name of the user"
// @Success 200 {array} object.Device The Response object
// @router /get-devices [get]
func (c *ApiController) GetDevices() {
	owner, ok := c.requireOrganizationAdmin()
	if !ok {
		return
	}
	limit, field, value, sortField, sortOrder := c.getListPageParams()
	user := c.Ctx.Input.Query("user")
	count, err := object.GetDeviceCount(owner, user, field, value)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	paginator := pagination.SetPaginator(c.Ctx, limit, count)
	devices, err := object.GetPaginationDevices(owner, user, paginator.Offset(), limit, field, value, sortField, sortOrder)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(devices, paginator.Nums())
}

// GetMyDevices
// @Title GetMyDevices
// @Tag Device API
// @Description list the devices of the signed-in user
// @Success 200 {array} object.Device The Response object
// @router /get-my-devices [get]
func (c *ApiController) GetMyDevices() {
	user, ok := c.RequireSignedInUser()
	if !ok {
		return
	}
	devices, err := object.GetPaginationDevices(user.Owner, user.Name, 0, 100, "", "", "", "")
	if err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.ResponseOk(devices)
}

// requireDevice returns the device in the query parameter "id" when the current user administers its organization,
// or owns it and may only remove it
func (c *ApiController) requireDevice(allowOwner bool) (*object.Device, bool) {
	id := c.Ctx.Input.Query("id")
	device, err := object.GetDevice(id)
	if err != nil {
		c.ResponseError(err.Error())
		return nil, false
	}
	if device == nil {
		c.ResponseError(fmt.Sprintf(c.T("enterprise:The device: %s doesn't exist"), id))
		return nil, false
	}
	user := c.GetCurrentUser()
	isOwner := allowOwner && user != nil && user.Owner == device.Owner && user.Name == device.User
	if !isOwner && !c.IsAdminOfOrganization(device.Owner) {
		c.ResponseError(c.T("auth:Unauthorized operation"))
		return nil, false
	}
	return device, true
}

// UpdateDevice
// @Title UpdateDevice
// @Tag Device API
// @Description trust, block or rename a device
// @Param   id      query  string     true        "The id ( owner/name ) of the device"
// @Param   body    body   object.Device  true        "The device"
// @Success 200 {object} controllers.Response The Response object
// @router /update-device [post]
func (c *ApiController) UpdateDevice() {
	device, ok := c.requireDevice(false)
	if !ok {
		return
	}
	var form object.Device
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &form); err != nil {
		c.ResponseError(err.Error())
		return
	}
	c.Data["json"] = WrapActionResponse(object.UpdateDevice(device.GetId(), &form))
	c.ServeJSON()
}

// DeleteDevice
// @Title DeleteDevice
// @Tag Device API
// @Description remove a device, by an admin or the user it is bound to
// @Param   id     query    string  true        "The id ( owner/name ) of the device"
// @Success 200 {object} controllers.Response The Response object
// @router /delete-device [post]
func (c *ApiController) DeleteDevice() {
	device, ok := c.requireDevice(true)
	if !ok {
		return
	}
	c.Data["json"] = WrapActionResponse(object.DeleteDevice(device))
	c.ServeJSON()
}
