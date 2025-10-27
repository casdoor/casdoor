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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/casdoor/casdoor/util"
)

type WecomAccessTokenResp struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type WecomUser struct {
	UserId     string `json:"userid"`
	Name       string `json:"name"`
	Department []int  `json:"department"`
	Position   string `json:"position"`
	Mobile     string `json:"mobile"`
	Gender     string `json:"gender"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar"`
	Status     int    `json:"status"`
	Enable     int    `json:"enable"`
}

type WecomUserListResp struct {
	Errcode  int          `json:"errcode"`
	Errmsg   string       `json:"errmsg"`
	Userlist []*WecomUser `json:"userlist"`
}

type WecomDeptListResp struct {
	Errcode    int    `json:"errcode"`
	Errmsg     string `json:"errmsg"`
	Department []struct {
		Id int `json:"id"`
	} `json:"department"`
}

// getWecomAccessToken gets access token from WeCom API
func (syncer *Syncer) getWecomAccessToken() (string, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		url.QueryEscape(syncer.User), url.QueryEscape(syncer.Password))

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

	var tokenResp WecomAccessTokenResp
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

// getWecomDepartments gets all department IDs from WeCom API
func (syncer *Syncer) getWecomDepartments(accessToken string) ([]int, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/department/list?access_token=%s",
		url.QueryEscape(accessToken))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

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

	var deptResp WecomDeptListResp
	err = json.Unmarshal(data, &deptResp)
	if err != nil {
		return nil, err
	}

	if deptResp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get departments: errcode=%d, errmsg=%s",
			deptResp.Errcode, deptResp.Errmsg)
	}

	deptIds := []int{}
	for _, dept := range deptResp.Department {
		deptIds = append(deptIds, dept.Id)
	}

	return deptIds, nil
}

// getWecomUsersFromDept gets users from a specific department
func (syncer *Syncer) getWecomUsersFromDept(accessToken string, deptId int) ([]*WecomUser, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/list?access_token=%s&department_id=%d",
		url.QueryEscape(accessToken), deptId)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, err
	}

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

	var userResp WecomUserListResp
	err = json.Unmarshal(data, &userResp)
	if err != nil {
		return nil, err
	}

	if userResp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get users from dept %d: errcode=%d, errmsg=%s",
			deptId, userResp.Errcode, userResp.Errmsg)
	}

	return userResp.Userlist, nil
}

// getWecomUsers gets all users from WeCom API
func (syncer *Syncer) getWecomUsers() ([]*OriginalUser, error) {
	// Get access token
	accessToken, err := syncer.getWecomAccessToken()
	if err != nil {
		return nil, err
	}

	// Get all departments
	deptIds, err := syncer.getWecomDepartments(accessToken)
	if err != nil {
		return nil, err
	}

	// Get users from all departments (deduplicate by userid)
	userMap := make(map[string]*WecomUser)
	for _, deptId := range deptIds {
		users, err := syncer.getWecomUsersFromDept(accessToken, deptId)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			// Deduplicate users by userid
			if _, exists := userMap[user.UserId]; !exists {
				userMap[user.UserId] = user
			}
		}
	}

	// Convert WeCom users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, wecomUser := range userMap {
		originalUser := syncer.wecomUserToOriginalUser(wecomUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// wecomUserToOriginalUser converts WeCom user to Casdoor OriginalUser
func (syncer *Syncer) wecomUserToOriginalUser(wecomUser *WecomUser) *OriginalUser {
	user := &OriginalUser{
		Id:          wecomUser.UserId,
		Name:        wecomUser.UserId,
		DisplayName: wecomUser.Name,
		Email:       wecomUser.Email,
		Phone:       wecomUser.Mobile,
		Avatar:      wecomUser.Avatar,
		Title:       wecomUser.Position,
		Address:     []string{},
		Properties:  map[string]string{},
		Groups:      []string{},
	}

	// Set gender
	switch wecomUser.Gender {
	case "1":
		user.Gender = "Male"
	case "2":
		user.Gender = "Female"
	default:
		user.Gender = ""
	}

	// Set IsForbidden based on status
	// status: 1=activated, 2=disabled, 4=not activated, 5=quit
	// enable: 1=enabled, 0=disabled
	if wecomUser.Status == 2 || wecomUser.Status == 4 || wecomUser.Status == 5 || wecomUser.Enable == 0 {
		user.IsForbidden = true
	} else {
		user.IsForbidden = false
	}

	// Set CreatedTime to current time if not set
	if user.CreatedTime == "" {
		user.CreatedTime = util.GetCurrentTime()
	}

	return user
}

// getWecomOriginalUsers is the main entry point for WeCom syncer
func (syncer *Syncer) getWecomOriginalUsers() ([]*OriginalUser, error) {
	return syncer.getWecomUsers()
}
