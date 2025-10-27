// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// WeComSyncer provides methods to sync users from WeCom
type WeComSyncer struct {
	Client       *http.Client
	CorpId       string
	CorpSecret   string
	AccessToken  string
	DepartmentId string
}

// WeComUserListResponse represents the response from WeCom user list API
// API: https://developer.work.weixin.qq.com/document/path/96021
type WeComUserListResponse struct {
	Errcode    int      `json:"errcode"`
	Errmsg     string   `json:"errmsg"`
	UserIdList []string `json:"userid_list"`
}

// WeComUserDetailResponse represents the response from WeCom user detail API
// API: https://developer.work.weixin.qq.com/document/path/90332
type WeComUserDetailResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	UserId  string `json:"userid"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Avatar  string `json:"avatar"`
	Mobile  string `json:"mobile"`
	Gender  string `json:"gender"`
	Status  int    `json:"status"`
}

// NewWeComSyncer creates a new WeCom syncer
func NewWeComSyncer(corpId, corpSecret string, departmentId string) *WeComSyncer {
	return &WeComSyncer{
		Client:       &http.Client{},
		CorpId:       corpId,
		CorpSecret:   corpSecret,
		DepartmentId: departmentId,
	}
}

// SetHttpClient sets the HTTP client for the syncer
func (s *WeComSyncer) SetHttpClient(client *http.Client) {
	s.Client = client
}

// getAccessToken retrieves access token from WeCom API
func (s *WeComSyncer) getAccessToken() error {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", s.CorpId, s.CorpSecret)
	resp, err := s.Client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var tokenResp WecomInterToken
	err = json.Unmarshal(data, &tokenResp)
	if err != nil {
		return err
	}

	if tokenResp.Errcode != 0 {
		return fmt.Errorf("failed to get access token: errcode=%d, errmsg=%s", tokenResp.Errcode, tokenResp.Errmsg)
	}

	s.AccessToken = tokenResp.AccessToken
	return nil
}

// GetUserIdList fetches the list of user IDs from WeCom
// API: https://developer.work.weixin.qq.com/document/path/96021
func (s *WeComSyncer) GetUserIdList() ([]string, error) {
	if s.AccessToken == "" {
		if err := s.getAccessToken(); err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/list_id?access_token=%s", s.AccessToken)

	// Create request body
	reqBody := map[string]interface{}{
		"cursor": "",
		"limit":  10000,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := s.Client.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userListResp WeComUserListResponse
	err = json.Unmarshal(data, &userListResp)
	if err != nil {
		return nil, err
	}

	if userListResp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get user list: errcode=%d, errmsg=%s", userListResp.Errcode, userListResp.Errmsg)
	}

	return userListResp.UserIdList, nil
}

// GetUserDetail fetches detailed user information from WeCom
// API: https://developer.work.weixin.qq.com/document/path/90332
func (s *WeComSyncer) GetUserDetail(userId string) (*UserInfo, error) {
	if s.AccessToken == "" {
		if err := s.getAccessToken(); err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token=%s&userid=%s", s.AccessToken, userId)
	resp, err := s.Client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userDetailResp WeComUserDetailResponse
	err = json.Unmarshal(data, &userDetailResp)
	if err != nil {
		return nil, err
	}

	if userDetailResp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get user detail: errcode=%d, errmsg=%s", userDetailResp.Errcode, userDetailResp.Errmsg)
	}

	// Map WeCom user to UserInfo
	userInfo := &UserInfo{
		Id:          userDetailResp.UserId,
		Username:    userDetailResp.Name,
		DisplayName: userDetailResp.Name,
		Email:       userDetailResp.Email,
		Phone:       userDetailResp.Mobile,
		AvatarUrl:   userDetailResp.Avatar,
	}

	return userInfo, nil
}

// GetAllUsers fetches all users from WeCom
func (s *WeComSyncer) GetAllUsers() ([]*UserInfo, error) {
	// Get list of user IDs
	userIds, err := s.GetUserIdList()
	if err != nil {
		return nil, err
	}

	// Fetch details for each user
	users := make([]*UserInfo, 0, len(userIds))
	for _, userId := range userIds {
		userInfo, err := s.GetUserDetail(userId)
		if err != nil {
			// Log error but continue with other users
			fmt.Printf("Failed to get user detail for %s: %v\n", userId, err)
			continue
		}
		users = append(users, userInfo)
	}

	return users, nil
}
