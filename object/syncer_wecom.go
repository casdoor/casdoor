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
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

// WecomSyncerProvider implements SyncerProvider for WeCom (WeChat Work) API-based syncers
type WecomSyncerProvider struct {
	Syncer *Syncer
	// the last error met while reading department names, so that GetOriginalGroups() can tell
	// why some groups fall back to the department ID as their display name
	deptNameError error
}

// InitAdapter initializes the WeCom syncer (no database adapter needed)
func (p *WecomSyncerProvider) InitAdapter() error {
	// WeCom syncer doesn't need database adapter
	return nil
}

// GetOriginalUsers retrieves all users from WeCom API
func (p *WecomSyncerProvider) GetOriginalUsers() ([]*OriginalUser, error) {
	return p.getWecomUsers()
}

// AddUser adds a new user to WeCom (not supported for read-only API)
func (p *WecomSyncerProvider) AddUser(user *OriginalUser) (bool, error) {
	// WeCom syncer is typically read-only
	return false, fmt.Errorf("adding users to WeCom is not supported")
}

// UpdateUser updates an existing user in WeCom (not supported for read-only API)
func (p *WecomSyncerProvider) UpdateUser(user *OriginalUser) (bool, error) {
	// WeCom syncer is typically read-only
	return false, fmt.Errorf("updating users in WeCom is not supported")
}

// TestConnection tests the WeCom API connection
func (p *WecomSyncerProvider) TestConnection() error {
	_, err := p.getWecomAccessToken()
	return err
}

// Close closes any open connections (no-op for WeCom API-based syncer)
func (p *WecomSyncerProvider) Close() error {
	// WeCom syncer doesn't maintain persistent connections
	return nil
}

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

type WecomUserGetResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	*WecomUser
}

type WecomDeptUser struct {
	UserId     string `json:"userid"`
	Department int    `json:"department"`
}

type WecomUserListIdResp struct {
	Errcode    int              `json:"errcode"`
	Errmsg     string           `json:"errmsg"`
	NextCursor string           `json:"next_cursor"`
	DeptUser   []*WecomDeptUser `json:"dept_user"`
}

type WecomDepartment struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	NameEn   string `json:"name_en"`
	ParentId int    `json:"parentid"`
	Order    int    `json:"order"`
}

type WecomDeptListResp struct {
	Errcode    int                `json:"errcode"`
	Errmsg     string             `json:"errmsg"`
	Department []*WecomDepartment `json:"department"`
}

type WecomDeptSimpleListResp struct {
	Errcode      int                `json:"errcode"`
	Errmsg       string             `json:"errmsg"`
	DepartmentId []*WecomDepartment `json:"department_id"`
}

type WecomDeptGetResp struct {
	Errcode    int              `json:"errcode"`
	Errmsg     string           `json:"errmsg"`
	Department *WecomDepartment `json:"department"`
}

// getWecomApi sends a GET request to the WeCom API and returns the response body
func (p *WecomSyncerProvider) getWecomApi(apiUrl string) ([]byte, error) {
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

	return io.ReadAll(resp.Body)
}

// postWecomApi sends a POST request with a JSON body to the WeCom API and returns the response body
func (p *WecomSyncerProvider) postWecomApi(apiUrl string, data map[string]interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", apiUrl, bytes.NewReader(jsonData))
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

	return io.ReadAll(resp.Body)
}

// getWecomAccessToken gets access token from WeCom API
func (p *WecomSyncerProvider) getWecomAccessToken() (string, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s",
		url.QueryEscape(p.Syncer.User), url.QueryEscape(p.Syncer.Password))

	data, err := p.getWecomApi(apiUrl)
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

// getWecomDepartments gets all departments from WeCom API
func (p *WecomSyncerProvider) getWecomDepartments(accessToken string) ([]*WecomDepartment, error) {
	depts, err := p.getWecomDepartmentsByList(accessToken)
	if err == nil {
		p.fillWecomDepartmentNames(accessToken, depts)
		return depts, nil
	}

	// department/list is not available to apps created after 2022-08-15, fall back to
	// department/simplelist + department/get, which every app can call
	depts, err2 := p.getWecomDepartmentsBySimpleList(accessToken)
	if err2 != nil {
		return nil, err
	}

	return depts, nil
}

// getWecomDepartmentsByList gets all departments with their details in a single call
func (p *WecomSyncerProvider) getWecomDepartmentsByList(accessToken string) ([]*WecomDepartment, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/department/list?access_token=%s",
		url.QueryEscape(accessToken))

	data, err := p.getWecomApi(apiUrl)
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

	return deptResp.Department, nil
}

// getWecomDepartmentsBySimpleList gets all department IDs and then the details of each department
func (p *WecomSyncerProvider) getWecomDepartmentsBySimpleList(accessToken string) ([]*WecomDepartment, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/department/simplelist?access_token=%s",
		url.QueryEscape(accessToken))

	data, err := p.getWecomApi(apiUrl)
	if err != nil {
		return nil, err
	}

	var deptResp WecomDeptSimpleListResp
	err = json.Unmarshal(data, &deptResp)
	if err != nil {
		return nil, err
	}

	if deptResp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get departments: errcode=%d, errmsg=%s",
			deptResp.Errcode, deptResp.Errmsg)
	}

	depts := []*WecomDepartment{}
	for _, simpleDept := range deptResp.DepartmentId {
		dept, err := p.getWecomDepartmentDetails(accessToken, simpleDept.Id)
		if err != nil {
			// Keep the name-less department so that the department tree stays complete
			fmt.Printf("Warning: failed to get details for department %d: %v\n", simpleDept.Id, err)
			p.deptNameError = err
			depts = append(depts, simpleDept)
			continue
		}

		depts = append(depts, dept)
	}

	return depts, nil
}

// fillWecomDepartmentNames reads the names that department/list did not return. Apps without
// the contact permission get departments with an empty name, while department/get may still
// return the name of the departments inside the app's visible scope.
func (p *WecomSyncerProvider) fillWecomDepartmentNames(accessToken string, depts []*WecomDepartment) {
	for _, dept := range depts {
		if dept.Name != "" || dept.NameEn != "" {
			continue
		}

		detail, err := p.getWecomDepartmentDetails(accessToken, dept.Id)
		if err != nil {
			fmt.Printf("Warning: failed to get details for department %d: %v\n", dept.Id, err)
			p.deptNameError = err
			continue
		}

		dept.Name = detail.Name
		dept.NameEn = detail.NameEn
	}
}

// reportNamelessWecomDepts records why some departments have no name. Their groups are still
// synced, using the department ID as display name, so the sync looks successful and the real
// reason would otherwise never reach the user.
func (p *WecomSyncerProvider) reportNamelessWecomDepts(depts []*WecomDepartment) {
	namelessDepts := []string{}
	for _, dept := range depts {
		if dept.Name == "" && dept.NameEn == "" {
			namelessDepts = append(namelessDepts, getWecomGroupName(dept.Id))
		}
	}

	if len(namelessDepts) == 0 {
		return
	}

	reason := "WeCom returned no department name"
	if p.deptNameError != nil {
		reason = p.deptNameError.Error()
	}

	line := fmt.Sprintf("[%s] failed to get the names of the WeCom departments [%s], their groups use the department ID as display name, please make sure the departments are inside the app's visible scope and the app has the contact permission: %s\n",
		util.GetCurrentTime(), strings.Join(namelessDepts, ", "), reason)
	_, err := updateSyncerErrorText(p.Syncer, line)
	if err != nil {
		fmt.Printf("reportNamelessWecomDepts() error: %s\n", err.Error())
	}
}

// getWecomDepartmentDetails gets detailed department information
func (p *WecomSyncerProvider) getWecomDepartmentDetails(accessToken string, deptId int) (*WecomDepartment, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/department/get?access_token=%s&id=%d",
		url.QueryEscape(accessToken), deptId)

	data, err := p.getWecomApi(apiUrl)
	if err != nil {
		return nil, err
	}

	var deptResp WecomDeptGetResp
	err = json.Unmarshal(data, &deptResp)
	if err != nil {
		return nil, err
	}

	if deptResp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get department details for %d: errcode=%d, errmsg=%s",
			deptId, deptResp.Errcode, deptResp.Errmsg)
	}

	if deptResp.Department == nil {
		return nil, fmt.Errorf("failed to get department details for %d: the response is empty", deptId)
	}

	return deptResp.Department, nil
}

// getWecomUsersFromDept gets users from a specific department
func (p *WecomSyncerProvider) getWecomUsersFromDept(accessToken string, deptId int) ([]*WecomUser, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/list?access_token=%s&department_id=%d",
		url.QueryEscape(accessToken), deptId)

	data, err := p.getWecomApi(apiUrl)
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

// getWecomUsersFromDepts gets the users of all departments, deduplicated by userid
func (p *WecomSyncerProvider) getWecomUsersFromDepts(accessToken string, depts []*WecomDepartment) (map[string]*WecomUser, error) {
	userMap := map[string]*WecomUser{}
	for _, dept := range depts {
		users, err := p.getWecomUsersFromDept(accessToken, dept.Id)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			if _, exists := userMap[user.UserId]; !exists {
				userMap[user.UserId] = user
			}
		}
	}

	return userMap, nil
}

// getWecomUsersByListId gets all user IDs and then the details of each user
func (p *WecomSyncerProvider) getWecomUsersByListId(accessToken string) (map[string]*WecomUser, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/list_id?access_token=%s",
		url.QueryEscape(accessToken))

	userMap := map[string]*WecomUser{}
	visited := map[string]bool{}
	cursor := ""
	var lastErr error

	for {
		data, err := p.postWecomApi(apiUrl, map[string]interface{}{"cursor": cursor, "limit": 10000})
		if err != nil {
			return nil, err
		}

		var userResp WecomUserListIdResp
		err = json.Unmarshal(data, &userResp)
		if err != nil {
			return nil, err
		}

		if userResp.Errcode != 0 {
			return nil, fmt.Errorf("failed to get user IDs: errcode=%d, errmsg=%s",
				userResp.Errcode, userResp.Errmsg)
		}

		// A user belonging to several departments is returned once per department
		for _, deptUser := range userResp.DeptUser {
			if visited[deptUser.UserId] {
				continue
			}
			visited[deptUser.UserId] = true

			user, err := p.getWecomUserDetails(accessToken, deptUser.UserId)
			if err != nil {
				fmt.Printf("Warning: failed to get details for user %s: %v\n", deptUser.UserId, err)
				lastErr = err
				continue
			}

			userMap[user.UserId] = user
		}

		if userResp.NextCursor == "" {
			break
		}
		cursor = userResp.NextCursor
	}

	if len(userMap) == 0 && lastErr != nil {
		return nil, lastErr
	}

	return userMap, nil
}

// getWecomUserDetails gets detailed user information
func (p *WecomSyncerProvider) getWecomUserDetails(accessToken string, userId string) (*WecomUser, error) {
	apiUrl := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/user/get?access_token=%s&userid=%s",
		url.QueryEscape(accessToken), url.QueryEscape(userId))

	data, err := p.getWecomApi(apiUrl)
	if err != nil {
		return nil, err
	}

	var userResp WecomUserGetResp
	err = json.Unmarshal(data, &userResp)
	if err != nil {
		return nil, err
	}

	if userResp.Errcode != 0 {
		return nil, fmt.Errorf("failed to get user details for %s: errcode=%d, errmsg=%s",
			userId, userResp.Errcode, userResp.Errmsg)
	}

	if userResp.WecomUser == nil {
		return nil, fmt.Errorf("failed to get user details for %s: the response is empty", userId)
	}

	return userResp.WecomUser, nil
}

// getWecomUsers gets all users from WeCom API
func (p *WecomSyncerProvider) getWecomUsers() ([]*OriginalUser, error) {
	// Get access token
	accessToken, err := p.getWecomAccessToken()
	if err != nil {
		return nil, err
	}

	// Get users from all departments (deduplicate by userid)
	userMap := map[string]*WecomUser{}

	depts, err := p.getWecomDepartments(accessToken)
	if err == nil {
		userMap, err = p.getWecomUsersFromDepts(accessToken, depts)
	}

	if err != nil {
		// user/list is not available to apps created after 2022-08-15, fall back to
		// user/list_id + user/get, which every app can call
		fallbackUserMap, err2 := p.getWecomUsersByListId(accessToken)
		if err2 != nil {
			return nil, err
		}

		userMap = fallbackUserMap
	}

	// Convert WeCom users to Casdoor OriginalUser
	originalUsers := []*OriginalUser{}
	for _, wecomUser := range userMap {
		originalUser := p.wecomUserToOriginalUser(wecomUser)
		originalUsers = append(originalUsers, originalUser)
	}

	return originalUsers, nil
}

// wecomUserToOriginalUser converts WeCom user to Casdoor OriginalUser
func (p *WecomSyncerProvider) wecomUserToOriginalUser(wecomUser *WecomUser) *OriginalUser {
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
		Wecom:       wecomUser.UserId, // Link WeCom provider account
	}

	// Add department IDs to Groups field
	for _, deptId := range wecomUser.Department {
		user.Groups = append(user.Groups, p.getWecomGroupId(deptId))
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

// getWecomGroupName returns the Casdoor group name of a WeCom department. The department
// ID is used as the name because WeCom department names are not unique.
func getWecomGroupName(deptId int) string {
	return strconv.Itoa(deptId)
}

// getWecomGroupId returns the Casdoor group ID ("organization/name") of a WeCom department.
// User.Groups holds full group IDs, so a bare department ID would not match any group and
// the membership would be silently dropped by the group and permission APIs.
func (p *WecomSyncerProvider) getWecomGroupId(deptId int) string {
	return util.GetId(p.Syncer.Organization, getWecomGroupName(deptId))
}

// wecomDepartmentToOriginalGroup converts WeCom department to Casdoor OriginalGroup
func (p *WecomSyncerProvider) wecomDepartmentToOriginalGroup(dept *WecomDepartment, hasParent bool) *OriginalGroup {
	displayName := dept.Name
	if displayName == "" {
		displayName = dept.NameEn
	}
	if displayName == "" {
		displayName = getWecomGroupName(dept.Id)
	}

	parentId := ""
	if hasParent {
		parentId = getWecomGroupName(dept.ParentId)
	}

	return &OriginalGroup{
		Id:          p.getWecomGroupId(dept.Id),
		Name:        getWecomGroupName(dept.Id), // Use ID as name for uniqueness
		DisplayName: displayName,
		Description: "",           // WeCom doesn't provide description
		Type:        "department", // Mark as department type
		ParentId:    parentId,
		Manager:     "", // WeCom doesn't provide manager in dept details
		Email:       "", // WeCom doesn't provide email for departments
	}
}

// GetOriginalGroups retrieves all groups (departments) from WeCom
func (p *WecomSyncerProvider) GetOriginalGroups() ([]*OriginalGroup, error) {
	accessToken, err := p.getWecomAccessToken()
	if err != nil {
		return nil, err
	}

	depts, err := p.getWecomDepartments(accessToken)
	if err != nil {
		return nil, err
	}

	deptIds := map[int]bool{}
	for _, dept := range depts {
		deptIds[dept.Id] = true
	}

	originalGroups := []*OriginalGroup{}
	for _, dept := range depts {
		// A department whose parent is out of the app's scope becomes a top group
		originalGroups = append(originalGroups, p.wecomDepartmentToOriginalGroup(dept, deptIds[dept.ParentId]))
	}

	p.reportNamelessWecomDepts(depts)

	return originalGroups, nil
}

// GetOriginalUserGroups retrieves the group (department) IDs that a user belongs to
func (p *WecomSyncerProvider) GetOriginalUserGroups(userId string) ([]string, error) {
	accessToken, err := p.getWecomAccessToken()
	if err != nil {
		return nil, err
	}

	user, err := p.getWecomUserDetails(accessToken, userId)
	if err != nil {
		return nil, err
	}

	groupIds := []string{}
	for _, deptId := range user.Department {
		groupIds = append(groupIds, p.getWecomGroupId(deptId))
	}

	return groupIds, nil
}
