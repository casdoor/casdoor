// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

package captcha

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

const AliyunCaptchaVerifyUrl = "http://afs.aliyuncs.com"

type AliyunCaptchaProvider struct{}

func NewAliyunCaptchaProvider() *AliyunCaptchaProvider {
	captcha := &AliyunCaptchaProvider{}
	return captcha
}

func contentEscape(str string) string {
	str = strings.Replace(str, " ", "%20", -1)
	str = url.QueryEscape(str)
	return str
}

func (captcha *AliyunCaptchaProvider) VerifyCaptcha(token, clientSecret string) (bool, error) {
	pathData, err := url.ParseQuery(token)
	if err != nil {
		return false, err
	}

	pathData["Action"] = []string{"AuthenticateSig"}
	pathData["Format"] = []string{"json"}
	pathData["SignatureMethod"] = []string{"HMAC-SHA1"}
	pathData["SignatureNonce"] = []string{strconv.FormatInt(time.Now().UnixNano(), 10)}
	pathData["SignatureVersion"] = []string{"1.0"}
	pathData["Timestamp"] = []string{time.Now().UTC().Format("2006-01-02T15:04:05Z")}
	pathData["Version"] = []string{"2018-01-12"}

	var keys []string
	for k := range pathData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sortQuery := ""
	for _, k := range keys {
		sortQuery += k + "=" + contentEscape(pathData[k][0]) + "&"
	}
	sortQuery = strings.TrimSuffix(sortQuery, "&")

	stringToSign := fmt.Sprintf("GET&%s&%s", url.QueryEscape("/"), url.QueryEscape(sortQuery))

	signature := util.GetHmacSha1(clientSecret+"&", stringToSign)

	resp, err := http.Get(fmt.Sprintf("%s?%s&Signature=%s", AliyunCaptchaVerifyUrl, sortQuery, url.QueryEscape(signature)))
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	type captchaResponse struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	captchaResp := &captchaResponse{}

	err = json.Unmarshal(body, captchaResp)
	if err != nil {
		return false, err
	}

	if captchaResp.Code != "100" {
		return false, errors.New(captchaResp.Message)
	}

	return true, nil
}
