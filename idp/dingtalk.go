// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// A total of three steps are required:
//
// 1. Construct the link and get the temporary authorization code
//	tmp_auth_code through the code at the end of the url.
//
// 2. Use hmac256 to calculate the signature, and then submit it together with timestamp,
//	tmp_auth_code, accessKey to obtain unionid, userid, accessKey.
//
// 3. Get detailed information through userid.

type DingTalkIdProvider struct {
	Client *http.Client
	Config *oauth2.Config
}

// NewDingTalkIdProvider ...
func NewDingTalkIdProvider(clientId string, clientSecret string, redirectUrl string) *DingTalkIdProvider {
	idp := &DingTalkIdProvider{}

	config := idp.getConfig(clientId, clientSecret, redirectUrl)
	idp.Config = config

	return idp
}

// SetHttpClient ...
func (idp *DingTalkIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

// getConfig return a point of Config, which describes a typical 3-legged OAuth2 flow
func (idp *DingTalkIdProvider) getConfig(clientId string, clientSecret string, redirectUrl string) *oauth2.Config {
	var endpoint = oauth2.Endpoint{
		AuthURL:  "https://oapi.dingtalk.com/sns/getuserinfo_bycode",
		TokenURL: "https://oapi.dingtalk.com/gettoken",
	}

	var config = &oauth2.Config{
		// DingTalk not allow to set scopes,here it is just a placeholder,
		// convenient to use later
		Scopes: []string{"", ""},

		Endpoint:     endpoint,
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
	}

	return config
}

type DingTalkAccessToken struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"` // Interface call credentials
	ExpiresIn   int64  `json:"expires_in"`   // access_token interface call credential timeout time, unit (seconds)
}

type DingTalkIds struct {
	UserId  string `json:"user_id"`
	UnionId string `json:"union_id"`
}

type InfoResp struct {
	Errcode  int `json:"errcode"`
	UserInfo struct {
		Nick                 string `json:"nick"`
		Unionid              string `json:"unionid"`
		Openid               string `json:"openid"`
		MainOrgAuthHighLevel bool   `json:"main_org_auth_high_level"`
	} `json:"user_info"`
	Errmsg string `json:"errmsg"`
}

// GetToken use code get access_token (*operation of getting code ought to be done in front)
// get more detail via: https://developers.dingtalk.com/document/app/dingtalk-retrieve-user-information?spm=ding_open_doc.document.0.0.51b91a31wWV3tY#doc-api-dingtalk-GetUser
func (idp *DingTalkIdProvider) GetToken(code string) (*oauth2.Token, error) {
	timestamp := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	signature := EncodeSHA256(timestamp, idp.Config.ClientSecret)
	u := fmt.Sprintf(
		"%s?accessKey=%s&timestamp=%s&signature=%s", idp.Config.Endpoint.AuthURL,
		idp.Config.ClientID, timestamp, signature)

	tmpCode := struct {
		TmpAuthCode string `json:"tmp_auth_code"`
	}{code}
	bs, _ := json.Marshal(tmpCode)
	r := strings.NewReader(string(bs))
	resp, err := http.Post(u, "application/json;charset=UTF-8", r)
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	body, _ := ioutil.ReadAll(resp.Body)
	info := InfoResp{}
	_ = json.Unmarshal(body, &info)
	errCode := info.Errcode
	if errCode != 0 {
		return nil, fmt.Errorf("%d: %s", errCode, info.Errmsg)
	}

	u2 := fmt.Sprintf("%s?appkey=%s&appsecret=%s", idp.Config.Endpoint.TokenURL, idp.Config.ClientID, idp.Config.ClientSecret)
	resp, _ = http.Get(u2)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)
	body, _ = ioutil.ReadAll(resp.Body)
	tokenResp := DingTalkAccessToken{}
	_ = json.Unmarshal(body, &tokenResp)
	if tokenResp.ErrCode != 0 {
		return nil, fmt.Errorf("%d: %s", tokenResp.ErrCode, tokenResp.ErrMsg)
	}

	// use unionid to get userid
	unionid := info.UserInfo.Unionid
	userid, err := idp.GetUseridByUnionid(tokenResp.AccessToken, unionid)
	if err != nil {
		return nil, err
	}

	// Since DingTalk does not require scopes, put userid and unionid into
	// idp.config.scopes to facilitate GetUserInfo() to obtain these two parameters.
	idp.Config.Scopes = []string{unionid, userid}

	token := &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		Expiry:      time.Unix(time.Now().Unix()+tokenResp.ExpiresIn, 0),
	}

	return token, nil
}

type UnionIdResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Result  struct {
		ContactType string `json:"contact_type"`
		Userid      string `json:"userid"`
	} `json:"result"`
	RequestId string `json:"request_id"`
}

// GetUseridByUnionid ...
func (idp *DingTalkIdProvider) GetUseridByUnionid(accesstoken, unionid string) (userid string, err error) {
	u := fmt.Sprintf("https://oapi.dingtalk.com/topapi/user/getbyunionid?access_token=%s&unionid=%s",
		accesstoken, unionid)
	useridInfo, err := idp.GetUrlResp(u)
	if err != nil {
		return "", err
	}

	uresp := UnionIdResponse{}
	_ = json.Unmarshal([]byte(useridInfo), &uresp)
	errcode := uresp.Errcode
	if errcode != 0 {
		return "", fmt.Errorf("%d: %s", errcode, uresp.Errmsg)
	}
	return uresp.Result.Userid, nil
}

/*
{
	"errcode":0,
	"result":{
		"boss":false,
		"unionid":"5M6zgZBKQPCxdiPdANeJ6MgiEiE",
		"role_list":[
			{
				"group_name":"默认",
				"name":"主管理员",
				"id":2062489174
			}
		],
		"exclusive_account":false,
		"mobile":"15236176076",
		"active":true,
		"admin":true,
		"avatar":"https://static-legacy.dingtalk.com/media/lALPDeRETW9WAnnNAyDNAyA_800_800.png",
		"hide_mobile":false,
		"userid":"manager4713",
		"senior":false,
		"dept_order_list":[
			{
				"dept_id":1,
				"order":176294576350761512
			}
		],
		"real_authed":true,
		"name":"刘继坤",
		"dept_id_list":[
			1
		],
		"state_code":"86",
		"email":"",
		"leader_in_dept":[
			{
				"leader":false,
				"dept_id":1
			}
		]
	},
	"errmsg":"ok",
	"request_id":"3sug9d2exsla"
}
*/

type DingTalkUserResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Result  struct {
		Extension   string `json:"extension"`
		Unionid     string `json:"unionid"`
		Boss        bool   `json:"boss"`
		UnionEmpExt struct {
			CorpId          string `json:"corpId"`
			Userid          string `json:"userid"`
			UnionEmpMapList []struct {
				CorpId string `json:"corpId"`
				Userid string `json:"userid"`
			} `json:"unionEmpMapList"`
		} `json:"unionEmpExt"`
		RoleList []struct {
			GroupName string `json:"group_name"`
			Id        int    `json:"id"`
			Name      string `json:"name"`
		} `json:"role_list"`
		Admin         bool   `json:"admin"`
		Remark        string `json:"remark"`
		Title         string `json:"title"`
		HiredDate     int64  `json:"hired_date"`
		Userid        string `json:"userid"`
		WorkPlace     string `json:"work_place"`
		DeptOrderList []struct {
			DeptId int   `json:"dept_id"`
			Order  int64 `json:"order"`
		} `json:"dept_order_list"`
		RealAuthed   bool   `json:"real_authed"`
		DeptIdList   []int  `json:"dept_id_list"`
		JobNumber    string `json:"job_number"`
		Email        string `json:"email"`
		LeaderInDept []struct {
			DeptId int  `json:"dept_id"`
			Leader bool `json:"leader"`
		} `json:"leader_in_dept"`
		ManagerUserid string `json:"manager_userid"`
		Mobile        string `json:"mobile"`
		Active        bool   `json:"active"`
		Telephone     string `json:"telephone"`
		Avatar        string `json:"avatar"`
		HideMobile    bool   `json:"hide_mobile"`
		Senior        bool   `json:"senior"`
		Name          string `json:"name"`
		StateCode     string `json:"state_code"`
	} `json:"result"`
	RequestId string `json:"request_id"`
}

// GetUserInfo Use userid and access_token to get UserInfo
// get more detail via: https://developers.dingtalk.com/document/app/query-user-details
func (idp *DingTalkIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	var dtUserInfo DingTalkUserResponse
	accessToken := token.AccessToken

	u := fmt.Sprintf("https://oapi.dingtalk.com/topapi/v2/user/get?access_token=%s&userid=%s",
		accessToken, idp.Config.Scopes[1])

	userinfoResp, err := idp.GetUrlResp(u)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal([]byte(userinfoResp), &dtUserInfo); err != nil {
		return nil, err
	}

	userInfo := UserInfo{
		Id:          strconv.Itoa(dtUserInfo.Result.RoleList[0].Id),
		Username:    dtUserInfo.Result.RoleList[0].Name,
		DisplayName: dtUserInfo.Result.Name,
		Email:       dtUserInfo.Result.Email,
		AvatarUrl:   dtUserInfo.Result.Avatar,
	}

	return &userInfo, nil
}

func (idp *DingTalkIdProvider) GetUrlResp(url string) (string, error) {
	resp, err := idp.Client.Get(url)
	if err != nil {
		return "", err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// EncodeSHA256 Use the HmacSHA256 algorithm to sign, the signature data is the current timestamp,
// and the key is the appSecret corresponding to the appId. Use this key to calculate the timestamp signature value.
// get more detail via: https://developers.dingtalk.com/document/app/signature-calculation-for-logon-free-scenarios-1?spm=ding_open_doc.document.0.0.63262ea7l6iEm1#topic-2021698
func EncodeSHA256(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	sum := h.Sum(nil)
	msg1 := base64.StdEncoding.EncodeToString(sum)

	uv := url.Values{}
	uv.Add("0", msg1)
	msg2 := uv.Encode()[2:]
	return msg2
}
