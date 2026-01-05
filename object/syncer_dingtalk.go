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
	"net/url"
	"time"

	"github.com/casdoor/casdoor/util"
)

// DingtalkSyncerProvider implements SyncerProvider for DingTalk API-based syncers
type DingtalkSyncerProvider struct {
	Syncer *Syncer
}

// InitAdapter initializes the DingTalk syncer (no database adapter needed)
func (p *DingtalkSyncerProvider) InitAdapter() error {
	// DingTalk syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from DingTalk API
func (p *DingtalkSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getDingtalkUsers()
}

// AddUser adds a new user to DingTalk (not supported for read-only API)
func (p *DingtalkSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// DingTalk syncer is typically read-only
	return false, fmt.Errorf("adding users to DingTalk is not supported")
}

// UpdateUser updates an existing user in DingTalk (not supported for read-only API)
func (p *DingtalkSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// DingTalk syncer is typically read-only
	return false, fmt.Errorf("updating users in DingTalk is not supported")
}

// TestConnection tests the DingTalk API connection
func (p *DingtalkSyncerProvider) TestConnection() error {
	_, err := p.getDingtalkAccessToken()
	return err
}

// Close closes any open connections (no-op for DingTalk API-based syncer)
func (p *DingtalkSyncerProvider) Close() error {
	// DingTalk syncer doesn't maintain persistent connections
	return nil
}

type DingtalkAccessTokenResp struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type DingtalkUser struct {
	UserId     string `json:"userid"`
	Name       string `json:"name"`
	Department []int  `json:"dept_id_list"`
	Position   string `json:"title"`
	Mobile     string `json:"mobile"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	JobNumber  string `json:"job_number"`
	Active     bool   `json:"active"`
}

type DingtalkUserListResp struct {
	Errcode  int             `json:"errcode"`
	Errmsg   string          `json:"errmsg"`
	Result   *DingtalkResult `json:"result"`
	RequestId string         `json:"request_id"`
}

type DingtalkResult struct {
	List       []*DingtalkUser `json:"list"`
	HasMore    bool            `json:"has_more"`
	NextCursor int64           `json:"next_cursor"`
}

type DingtalkDeptListResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Result  []struct {
		DeptId int64 `json:"dept_id"`
	} `json:"result"`
	RequestId string `json:"request_id"`
}

// getDingtalkAccessToken gets access token from DingTalk API
func (p *DingtalkSyncerProvider) getDingtalkAccessToken() (string, error) {
	// syncer.User should be the appKey
	// syncer.Password should be the appSecret
	appKey := p.Syncer.User
	if appKey == "" {
		return "", fmt.Errorf("appKey (user field) is required for DingTalk syncer")
	}

	appSecret := p.Syncer.Password
	if appSecret == "" {
		return "", fmt.Errorf("appSecret (password field) is required for DingTalk syncer")
	}

	apiUrl := fmt.Sprintf("https://oapi.dingtalk.com/gettoken?appkey=%s&appsecret=%s",
		url.QueryEscape(appKey), url.QueryEscape(appSecret))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResp DingtalkAccessTokenResp
	err = json.Unmarshal(data, &tokenResp)
	if err != nil {
		return "", err
	}

	if tokenResp.Errcode != 0 {
		return "", fmt.Errorf("failed to get access token: errcode=%d, errmsg=%s",
			tokenResp.Errcode, tokenResp.Errmsg)
	}

	return tokenResp.AccessToken, nil
}

// getDingtalkDepartments gets all department IDs from DingTalk API
func (p *DingtalkSyncerProvider) getDingtalkDepartments(accessToken string) ([]int64, error) {
	apiUrl := fmt.Sprintf("https://oapi.dingtalk.com/topapi/v2/department/listsub?access_token=%s",
		url.QueryEscape(accessToken))

	// Get root department (dept_id=1)
	postData := map[string]interface{}{
		"dept_id": 1,
	}
	
	data, err := p.postJSON(apiUrl, postData)
	if err != nil {
		return nil, err
	}

	var deptResp DingtalkDeptListResp
	err = json.Unmarshal(data, &deptResp)
	if err != nil {
		return nil, err
	}

	if deptResp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get departments: errcode=%d, errmsg=%s",
			deptResp.Errcode, deptResp.Errmsg)
	}

	deptIds := []int64{1} // Include root department
	for _, dept := range deptResp.Result {
		deptIds = append(deptIds, dept.DeptId)
	}

	return deptIds, nil
}

// getDingtalkUsersFromDept gets users from a specific department
func (p *DingtalkSyncerProvider) getDingtalkUsersFromDept(accessToken string, deptId int64) ([]*DingtalkUser, error) {
	allUsers := []*DingtalkUser{}
	cursor := int64(0)
	
	for {
		apiUrl := fmt.Sprintf("https://oapi.dingtalk.com/topapi/user/listsimple?access_token=%s",
			url.QueryEscape(accessToken))

		postData := map[string]interface{}{
			"dept_id":    deptId,
			"cursor":     cursor,
			"size":       100,
		}

		data, err := p.postJSON(apiUrl, postData)
		if err != nil {
			return nil, err
		}

		var userResp DingtalkUserListResp
		err = json.Unmarshal(data, &userResp)
		if err != nil {
			return nil, err
		}

		if userResp.Errcode != 0 {
			return nil, fmt.Errorf("failed to get users from dept %d: errcode=%d, errmsg=%s",
				deptId, userResp.Errcode, userResp.Errmsg)
		}

		if userResp.Result != nil {
			allUsers = append(allUsers, userResp.Result.List...)
			
			if !userResp.Result.HasMore {
				break
			}
			cursor = userResp.Result.NextCursor
		} else {
			break
		}
	}

	return allUsers, nil
}

// getDingtalkUserDetails gets detailed user information
func (p *DingtalkSyncerProvider) getDingtalkUserDetails(accessToken string, userId string) (*DingtalkUser, error) {
	apiUrl := fmt.Sprintf("https://oapi.dingtalk.com/topapi/v2/user/get?access_token=%s",
		url.QueryEscape(accessToken))

	postData := map[string]interface{}{
		"userid": userId,
	}

	data, err := p.postJSON(apiUrl, postData)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
		Result  *DingtalkUser `json:"result"`
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get user details for %s: errcode=%d, errmsg=%s",
			userId, resp.Errcode, resp.Errmsg)
	}

	return resp.Result, nil
}

// postJSON sends a POST request with JSON body
func (p *DingtalkSyncerProvider) postJSON(url string, data map[string]interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, io.NopCloser(bytes.NewReader(jsonData)))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(jsonData))

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

// getDingtalkUsers gets all users from DingTalk API
func (p *DingtalkSyncerProvider) getDingtalkUsers() ([]*OriginalUser, error) {
	// Get access token
	accessToken, err := p.getDingtalkAccessToken()
	if err != nil {
		return nil, err
	}

	// Get all departments
	deptIds, err := p.getDingtalkDepartments(accessToken)
	if err != nil {
		return nil, err
	}

	// Get users from all departments (deduplicate by userid)
	userMap := make(map[string]*DingtalkUser)
	for _, deptId := range deptIds {
		users, err := p.getDingtalkUsersFromDept(accessToken, deptId)
		if err != nil {
			// Continue even if one department fails
			continue
		}

		for _, user := range users {
			// Deduplicate users by userid
			if _, exists := userMap[user.UserId]; !exists {
				// Get detailed user information
				detailedUser, err := p.getDingtalkUserDetails(accessToken, user.UserId)
				if err != nil {
					// Use basic user info if details fail
					userMap[user.UserId] = user
				} else {
					userMap[user.UserId] = detailedUser
				}
			}
		}
	}

	// Convert DingTalk users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, dingtalkUser := range userMap {
		originalUser := p.dingtalkUserToOriginalUser(dingtalkUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// dingtalkUserToOriginalUser converts DingTalk user to Casdoor OriginalUser
func (p *DingtalkSyncerProvider) dingtalkUserToOriginalUser(dingtalkUser *DingtalkUser) *OriginalUser {
	user := &OriginalUser{
		Id:          dingtalkUser.UserId,
		Name:        dingtalkUser.UserId,
		DisplayName: dingtalkUser.Name,
		Email:       dingtalkUser.Email,
		Phone:       dingtalkUser.Mobile,
		Avatar:      dingtalkUser.Avatar,
		Title:       dingtalkUser.Position,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Use job number as name if available
	if dingtalkUser.JobNumber != "" {
		user.Name = dingtalkUser.JobNumber
	}

	// Set IsForbidden based on active status
	user.IsForbidden = !dingtalkUser.Active

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}
