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

package object

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/casdoor/casdoor/util"
)

// LarkSyncerProvider implements SyncerProvider for Lark API-based syncers
type LarkSyncerProvider struct {
	Syncer *Syncer
}

// InitAdapter initializes the Lark syncer (no database adapter needed)
func (p *LarkSyncerProvider) InitAdapter() error {
	// Lark syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from Lark API
func (p *LarkSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getLarkUsers()
}

// AddUser adds a new user to Lark (not supported for read-only API)
func (p *LarkSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// Lark syncer is typically read-only
	return false, fmt.Errorf("adding users to Lark is not supported")
}

// UpdateUser updates an existing user in Lark (not supported for read-only API)
func (p *LarkSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// Lark syncer is typically read-only
	return false, fmt.Errorf("updating users in Lark is not supported")
}

// TestConnection tests the Lark API connection
func (p *LarkSyncerProvider) TestConnection() error {
	_, err := p.getLarkAccessToken()
	return err
}

// Close closes any open connections (no-op for Lark API-based syncer)
func (p *LarkSyncerProvider) Close() error {
	// Lark syncer doesn't maintain persistent connections
	return nil
}

type LarkAccessTokenResp struct {
	Code              int    `json:"code"`
	Msg               string `json:"msg"`
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int    `json:"expire"`
}

type LarkUser struct {
	UserId        string      `json:"user_id"`
	UnionId       string      `json:"union_id"`
	OpenId        string      `json:"open_id"`
	Name          string      `json:"name"`
	EnName        string      `json:"en_name"`
	Email         string      `json:"email"`
	Mobile        string      `json:"mobile"`
	Gender        int         `json:"gender"`
	Avatar        *LarkAvatar `json:"avatar"`
	Status        *LarkStatus `json:"status"`
	DepartmentIds []string    `json:"department_ids"`
	JobTitle      string      `json:"job_title"`
}

type LarkAvatar struct {
	Avatar72     string `json:"avatar_72"`
	Avatar240    string `json:"avatar_240"`
	Avatar640    string `json:"avatar_640"`
	AvatarOrigin string `json:"avatar_origin"`
}

type LarkStatus struct {
	IsFrozen    bool `json:"is_frozen"`
	IsResigned  bool `json:"is_resigned"`
	IsActivated bool `json:"is_activated"`
	IsExited    bool `json:"is_exited"`
}

type LarkUserListResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Items     []*LarkUser `json:"items"`
		HasMore   bool        `json:"has_more"`
		PageToken string      `json:"page_token"`
	} `json:"data"`
}

type LarkDeptListResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Items []struct {
			DepartmentId string `json:"department_id"`
		} `json:"items"`
		HasMore   bool   `json:"has_more"`
		PageToken string `json:"page_token"`
	} `json:"data"`
}

// getLarkDomain returns the Lark API domain based on whether global endpoint is used
func (p *LarkSyncerProvider) getLarkDomain() string {
	// syncer.Host can be used to specify custom endpoint
	// If empty, default to global endpoint (larksuite.com)
	if p.Syncer.Host != "" {
		return p.Syncer.Host
	}
	return "https://open.larksuite.com"
}

// getLarkAccessToken gets access token from Lark API
func (p *LarkSyncerProvider) getLarkAccessToken() (string, error) {
	// syncer.User should be the app_id
	// syncer.Password should be the app_secret
	appId := p.Syncer.User
	if appId == "" {
		return "", fmt.Errorf("app_id (user field) is required for Lark syncer")
	}

	appSecret := p.Syncer.Password
	if appSecret == "" {
		return "", fmt.Errorf("app_secret (password field) is required for Lark syncer")
	}

	domain := p.getLarkDomain()
	apiUrl := fmt.Sprintf("%s/open-apis/auth/v3/tenant_access_token/internal", domain)

	postData := map[string]string{
		"app_id":     appId,
		"app_secret": appSecret,
	}

	data, err := p.postJSON(apiUrl, postData)
	if err != nil {
		return "", err
	}

	var tokenResp LarkAccessTokenResp
	err = json.Unmarshal(data, &tokenResp)
	if err != nil {
		return "", err
	}

	if tokenResp.Code != 0 {
		return "", fmt.Errorf("failed to get access token: code=%d, msg=%s",
			tokenResp.Code, tokenResp.Msg)
	}

	return tokenResp.TenantAccessToken, nil
}

// getLarkDepartments gets all department IDs from Lark API
func (p *LarkSyncerProvider) getLarkDepartments(accessToken string) ([]string, error) {
	domain := p.getLarkDomain()
	allDeptIds := []string{"0"} // Start with root department
	pageToken := ""

	for {
		apiUrl := fmt.Sprintf("%s/open-apis/contact/v3/departments?parent_department_id=0&fetch_child=true&page_size=50", domain)
		if pageToken != "" {
			apiUrl += fmt.Sprintf("&page_token=%s", pageToken)
		}

		data, err := p.getWithAuth(apiUrl, accessToken)
		if err != nil {
			return nil, err
		}

		var deptResp LarkDeptListResp
		err = json.Unmarshal(data, &deptResp)
		if err != nil {
			return nil, err
		}

		if deptResp.Code != 0 {
			return nil, fmt.Errorf("failed to get departments: code=%d, msg=%s",
				deptResp.Code, deptResp.Msg)
		}

		for _, dept := range deptResp.Data.Items {
			allDeptIds = append(allDeptIds, dept.DepartmentId)
		}

		if !deptResp.Data.HasMore {
			break
		}
		pageToken = deptResp.Data.PageToken
	}

	return allDeptIds, nil
}

// getLarkUsersFromDept gets users from a specific department
func (p *LarkSyncerProvider) getLarkUsersFromDept(accessToken string, deptId string) ([]*LarkUser, error) {
	domain := p.getLarkDomain()
	allUsers := []*LarkUser{}
	pageToken := ""

	for {
		apiUrl := fmt.Sprintf("%s/open-apis/contact/v3/users/find_by_department?department_id=%s&page_size=50", domain, deptId)
		if pageToken != "" {
			apiUrl += fmt.Sprintf("&page_token=%s", pageToken)
		}

		data, err := p.getWithAuth(apiUrl, accessToken)
		if err != nil {
			return nil, err
		}

		var userResp LarkUserListResp
		err = json.Unmarshal(data, &userResp)
		if err != nil {
			return nil, err
		}

		if userResp.Code != 0 {
			return nil, fmt.Errorf("failed to get users from dept %s: code=%d, msg=%s",
				deptId, userResp.Code, userResp.Msg)
		}

		allUsers = append(allUsers, userResp.Data.Items...)

		if !userResp.Data.HasMore {
			break
		}
		pageToken = userResp.Data.PageToken
	}

	return allUsers, nil
}

// postJSON sends a POST request with JSON body
func (p *LarkSyncerProvider) postJSON(url string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

// getWithAuth sends a GET request with authorization header
func (p *LarkSyncerProvider) getWithAuth(url string, accessToken string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// getLarkUsers gets all users from Lark API
func (p *LarkSyncerProvider) getLarkUsers() ([]*OriginalUser, error) {
	// Get access token
	accessToken, err := p.getLarkAccessToken()
	if err != nil {
		return nil, err
	}

	// Get all departments
	deptIds, err := p.getLarkDepartments(accessToken)
	if err != nil {
		return nil, err
	}

	// Get users from all departments (deduplicate by user_id)
	userMap := make(map[string]*LarkUser)
	for _, deptId := range deptIds {
		users, err := p.getLarkUsersFromDept(accessToken, deptId)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			// Deduplicate users by user_id
			if _, exists := userMap[user.UserId]; !exists {
				userMap[user.UserId] = user
			}
		}
	}

	// Convert Lark users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, larkUser := range userMap {
		originalUser := p.larkUserToOriginalUser(larkUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// larkUserToOriginalUser converts Lark user to Casdoor OriginalUser
func (p *LarkSyncerProvider) larkUserToOriginalUser(larkUser *LarkUser) *OriginalUser {
	// Use user_id as name, fallback to union_id or open_id
	userName := larkUser.UserId
	if userName == "" && larkUser.UnionId != "" {
		userName = larkUser.UnionId
	}
	if userName == "" && larkUser.OpenId != "" {
		userName = larkUser.OpenId
	}

	user := &OriginalUser{
		Id:          larkUser.UserId,
		Name:        userName,
		DisplayName: larkUser.Name,
		Email:       larkUser.Email,
		Phone:       larkUser.Mobile,
		Title:       larkUser.JobTitle,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Set avatar if available
	if larkUser.Avatar != nil {
		if larkUser.Avatar.Avatar240 != "" {
			user.Avatar = larkUser.Avatar.Avatar240
		} else if larkUser.Avatar.Avatar72 != "" {
			user.Avatar = larkUser.Avatar.Avatar72
		}
	}

	// Set gender
	switch larkUser.Gender {
	case 1:
		user.Gender = "Male"
	case 2:
		user.Gender = "Female"
	default:
		user.Gender = ""
	}

	// Set IsForbidden based on status
	// User is forbidden if frozen, resigned, not activated, or exited
	if larkUser.Status != nil {
		if larkUser.Status.IsFrozen || larkUser.Status.IsResigned || !larkUser.Status.IsActivated || larkUser.Status.IsExited {
			user.IsForbidden = true
		} else {
			user.IsForbidden = false
		}
	}

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}
