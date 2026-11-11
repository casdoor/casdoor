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

package object

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
)

const (
	scimUserSchema           = "urn:ietf:params:scim:schemas:core:2.0:User"
	scimEnterpriseUserSchema = "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
	scimPatchOpSchema        = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
	scimContentType          = "application/scim+json"
)

type scimProvisioningTarget struct {
	application *Application
	config      *ProvisioningConfig
	client      *http.Client
}

func newScimProvisioningTarget(application *Application) *scimProvisioningTarget {
	client := util.NewInternetOnlyHttpClient(30 * time.Second)
	// the systems behind the applications of a private deployment are usually on its intranet
	if application.Organization == "built-in" || conf.GetConfigBool("provisioningAllowIntranet") {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	return &scimProvisioningTarget{application: application, config: application.Provisioning, client: client}
}

type scimError struct {
	StatusCode int
	Body       string
}

func (e *scimError) Error() string {
	return fmt.Sprintf("the SCIM server answered %d: %s", e.StatusCode, truncateString(e.Body, 500))
}

func (t *scimProvisioningTarget) request(method string, path string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, strings.TrimSuffix(t.config.Url, "/")+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", scimContentType)
	if body != nil {
		req.Header.Set("Content-Type", scimContentType)
	}
	if t.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.config.Token)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return data, &scimError{StatusCode: resp.StatusCode, Body: string(data)}
	}
	return data, nil
}

func isScimStatus(err error, statusCodes ...int) bool {
	scimErr, ok := err.(*scimError)
	if !ok {
		return false
	}
	for _, statusCode := range statusCodes {
		if scimErr.StatusCode == statusCode {
			return true
		}
	}
	return false
}

func (t *scimProvisioningTarget) getUserName(user *User) string {
	switch t.config.UserNameField {
	case "Email":
		if user.Email != "" {
			return user.Email
		}
	case "Phone":
		if user.Phone != "" {
			return user.Phone
		}
	case "Id":
		return user.Id
	}
	return user.Name
}

func (t *scimProvisioningTarget) payload(user *User) (interface{}, error) {
	displayName := user.DisplayName
	if displayName == "" {
		displayName = user.Name
	}

	res := map[string]interface{}{
		"schemas":     []string{scimUserSchema, scimEnterpriseUserSchema},
		"externalId":  user.Id,
		"userName":    t.getUserName(user),
		"displayName": displayName,
		"active":      true,
	}

	name := map[string]string{"formatted": displayName}
	if user.FirstName != "" {
		name["givenName"] = user.FirstName
	}
	if user.LastName != "" {
		name["familyName"] = user.LastName
	}
	res["name"] = name

	if user.Email != "" {
		res["emails"] = []map[string]interface{}{{"value": user.Email, "type": "work", "primary": true}}
	}
	if user.Phone != "" {
		phone := user.Phone
		if e164, ok := util.GetE164Number(user.Phone, user.CountryCode); ok {
			phone = e164
		}
		res["phoneNumbers"] = []map[string]interface{}{{"value": phone, "type": "work", "primary": true}}
	}
	if user.Title != "" {
		res["title"] = user.Title
	}
	if user.Language != "" {
		res["preferredLanguage"] = user.Language
	}

	enterprise := map[string]interface{}{"organization": user.Owner}
	if len(user.Groups) > 0 {
		_, department := util.GetOwnerAndNameFromIdNoCheck(user.Groups[0])
		enterprise["department"] = department
	}
	if user.Affiliation != "" {
		enterprise["division"] = user.Affiliation
	}
	res[scimEnterpriseUserSchema] = enterprise
	return res, nil
}

type scimListResponse struct {
	Resources []struct {
		Id string `json:"id"`
	} `json:"Resources"`
}

// findRemoteId looks the user up by userName, for an account that the system already has
func (t *scimProvisioningTarget) findRemoteId(userName string) (string, error) {
	filter := url.QueryEscape(fmt.Sprintf("userName eq \"%s\"", strings.ReplaceAll(userName, "\"", "\\\"")))
	data, err := t.request(http.MethodGet, "/Users?filter="+filter, nil)
	if err != nil {
		return "", err
	}
	list := scimListResponse{}
	if err = json.Unmarshal(data, &list); err != nil {
		return "", err
	}
	if len(list.Resources) == 0 {
		return "", nil
	}
	return list.Resources[0].Id, nil
}

func getScimId(data []byte) (string, error) {
	resource := struct {
		Id string `json:"id"`
	}{}
	if err := json.Unmarshal(data, &resource); err != nil {
		return "", err
	}
	if resource.Id == "" {
		return "", fmt.Errorf("the SCIM server returned no id")
	}
	return resource.Id, nil
}

func (t *scimProvisioningTarget) replace(remoteId string, body map[string]interface{}) error {
	body["id"] = remoteId
	_, err := t.request(http.MethodPut, "/Users/"+url.PathEscape(remoteId), body)
	return err
}

func (t *scimProvisioningTarget) upsert(user *User, account *ProvisionedAccount) (string, error) {
	payload, err := t.payload(user)
	if err != nil {
		return "", err
	}
	body := payload.(map[string]interface{})

	if account.RemoteId != "" {
		err = t.replace(account.RemoteId, body)
		if err == nil {
			return account.RemoteId, nil
		}
		if !isScimStatus(err, http.StatusNotFound) {
			return "", err
		}
		delete(body, "id")
	}

	data, err := t.request(http.MethodPost, "/Users", body)
	if err == nil {
		return getScimId(data)
	}
	if !isScimStatus(err, http.StatusConflict) {
		return "", err
	}

	// the system already has the account, e.g. made before the provisioning was turned on
	remoteId, findErr := t.findRemoteId(body["userName"].(string))
	if findErr != nil || remoteId == "" {
		return "", err
	}
	delete(body, "id")
	return remoteId, t.replace(remoteId, body)
}

func (t *scimProvisioningTarget) disable(user *User, account *ProvisionedAccount) error {
	if account.RemoteId == "" {
		return nil
	}

	patch := map[string]interface{}{
		"schemas":    []string{scimPatchOpSchema},
		"Operations": []map[string]interface{}{{"op": "replace", "path": "active", "value": false}},
	}
	_, err := t.request(http.MethodPatch, "/Users/"+url.PathEscape(account.RemoteId), patch)
	if err == nil || isScimStatus(err, http.StatusNotFound) {
		return nil
	}

	// a server without PATCH gets the whole account with "active": false
	if user != nil && isScimStatus(err, http.StatusMethodNotAllowed, http.StatusNotImplemented, http.StatusBadRequest) {
		payload, payloadErr := t.payload(user)
		if payloadErr != nil {
			return payloadErr
		}
		body := payload.(map[string]interface{})
		body["active"] = false
		return t.replace(account.RemoteId, body)
	}
	return err
}

func (t *scimProvisioningTarget) remove(account *ProvisionedAccount) error {
	if account.RemoteId == "" {
		return nil
	}
	_, err := t.request(http.MethodDelete, "/Users/"+url.PathEscape(account.RemoteId), nil)
	if err != nil && !isScimStatus(err, http.StatusNotFound) {
		return err
	}
	return nil
}

func (t *scimProvisioningTarget) test() error {
	_, err := t.request(http.MethodGet, "/ServiceProviderConfig", nil)
	if err == nil {
		return nil
	}
	_, err = t.request(http.MethodGet, "/Users?count=1", nil)
	return err
}
