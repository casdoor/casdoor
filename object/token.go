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

package object

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

type Token struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	Application  string `xorm:"varchar(100)" json:"application"`
	Organization string `xorm:"varchar(100)" json:"organization"`
	User         string `xorm:"varchar(100)" json:"user"`

	Code             string `xorm:"varchar(100) index" json:"code"`
	AccessToken      string `xorm:"mediumtext" json:"accessToken"`
	RefreshToken     string `xorm:"mediumtext" json:"refreshToken"`
	AccessTokenHash  string `xorm:"varchar(100) index" json:"accessTokenHash"`
	RefreshTokenHash string `xorm:"varchar(100) index" json:"refreshTokenHash"`
	ExpiresIn        int    `json:"expiresIn"`
	Scope            string `xorm:"varchar(100)" json:"scope"`
	TokenType        string `xorm:"varchar(100)" json:"tokenType"`
	CodeChallenge    string `xorm:"varchar(100)" json:"codeChallenge"`
	CodeIsUsed       bool   `json:"codeIsUsed"`
	CodeExpireIn     int64  `json:"codeExpireIn"`
}

func GetTokenCount(owner, organization, field, value string) (int64, error) {
	session := GetSession(owner, -1, -1, field, value, "", "")
	return session.Count(&Token{Organization: organization})
}

func GetTokens(owner string, organization string) ([]*Token, error) {
	tokens := []*Token{}
	err := ormer.Engine.Desc("created_time").Find(&tokens, &Token{Owner: owner, Organization: organization})
	return tokens, err
}

func GetPaginationTokens(owner, organization string, offset, limit int, field, value, sortField, sortOrder string) ([]*Token, error) {
	tokens := []*Token{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&tokens, &Token{Organization: organization})
	return tokens, err
}

func getToken(owner string, name string) (*Token, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	token := Token{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if existed {
		return &token, nil
	}

	return nil, nil
}

func getTokenByCode(code string) (*Token, error) {
	token := Token{Code: code}
	existed, err := ormer.Engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if existed {
		return &token, nil
	}

	return nil, nil
}

func GetTokenByAccessToken(accessToken string) (*Token, error) {
	token := Token{AccessTokenHash: getTokenHash(accessToken)}
	existed, err := ormer.Engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if !existed {
		token = Token{AccessToken: accessToken}
		existed, err = ormer.Engine.Get(&token)
		if err != nil {
			return nil, err
		}
	}

	if !existed {
		return nil, nil
	}
	return &token, nil
}

func GetTokenByRefreshToken(refreshToken string) (*Token, error) {
	token := Token{RefreshTokenHash: getTokenHash(refreshToken)}
	existed, err := ormer.Engine.Get(&token)
	if err != nil {
		return nil, err
	}

	if !existed {
		token = Token{RefreshToken: refreshToken}
		existed, err = ormer.Engine.Get(&token)
		if err != nil {
			return nil, err
		}
	}

	if !existed {
		return nil, nil
	}
	return &token, nil
}

func GetTokenByTokenValue(tokenValue string) (*Token, error) {
	token, err := GetTokenByAccessToken(tokenValue)
	if err != nil {
		return nil, err
	}
	if token != nil {
		return token, nil
	}

	token, err = GetTokenByRefreshToken(tokenValue)
	if err != nil {
		return nil, err
	}
	if token != nil {
		return token, nil
	}

	return nil, nil
}

func updateUsedByCode(token *Token) bool {
	affected, err := ormer.Engine.Where("code=?", token.Code).Cols("code_is_used").Update(token)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func GetToken(id string) (*Token, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getToken(owner, name)
}

func (token *Token) GetId() string {
	return fmt.Sprintf("%s/%s", token.Owner, token.Name)
}

func getTokenHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	res := hex.EncodeToString(hash[:])
	if len(res) > 64 {
		return res[:64]
	}
	return res
}

func (token *Token) popularHashes() {
	if token.AccessTokenHash == "" && token.AccessToken != "" {
		token.AccessTokenHash = getTokenHash(token.AccessToken)
	}
	if token.RefreshTokenHash == "" && token.RefreshToken != "" {
		token.RefreshTokenHash = getTokenHash(token.RefreshToken)
	}
}

func UpdateToken(id string, token *Token) (bool, error) {
	owner, name := util.GetOwnerAndNameFromId(id)
	if t, err := getToken(owner, name); err != nil {
		return false, err
	} else if t == nil {
		return false, nil
	}

	token.popularHashes()

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(token)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddToken(token *Token) (bool, error) {
	token.popularHashes()

	affected, err := ormer.Engine.Insert(token)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteToken(token *Token) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{token.Owner, token.Name}).Delete(&Token{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

// ivan TODO logout expire TOKEN (or not)
//func NotifySsoWebApi(token string) error {
//	// ivan 240226, api call to notify logout event
//	var ssoweb string
//	var casKey string
//	var casId string
//	ssoweb = os.Getenv("SSO_API")
//	if ssoweb == "" {
//		ssoweb = conf.GetConfigString("SSO_API")
//	}
//	casKey = os.Getenv("CASDOOR_KEY")
//	if casKey == "" {
//		casKey = conf.GetConfigString("CASDOOR_KEY")
//	}
//	casId = os.Getenv("CASDOOR_ID")
//	if casId == "" {
//		casId = conf.GetConfigString("CASDOOR_ID")
//	}
//
//	url := ssoweb + "/pb.hooks.v1.HookService/ExpireToken"
//	payloadBytes, err := json.Marshal(struct {
//		Token string `json:"token"`
//	}{
//		Token: token,
//	})
//	if err != nil {
//		return err
//	}
//	payloadBody := bytes.NewReader(payloadBytes)
//	req, err := http.NewRequest("POST", url, payloadBody)
//	req.Header.Add("ApiKey", casKey)
//	req.Header.Add("AppId", casId)
//	req.Header.Add("Content-Type", "application/json")
//	client := &http.Client{
//		Timeout: time.Second * 30,
//	}
//	resp, err := client.Do(req)
//	if err != nil {
//		return err
//	}
//	if resp.StatusCode != http.StatusOK {
//		bodyBytes, err := io.ReadAll(resp.Body)
//		if err != nil {
//			return err
//		}
//		bodyString := string(bodyBytes)
//		log.Println("Response Status:", resp.StatusCode)
//		log.Println("Response Body:", bodyString)
//		return err
//	}
//	return nil
//}
//
//func ExpireTokenByAccessToken(accessToken string) (bool, *Application, *Token, error) {
//	token, err := GetTokenByAccessToken(accessToken)
//	if err != nil {
//		return false, nil, nil, err
//	}
//	if token == nil {
//		return false, nil, nil, nil
//	}
//
//	token.ExpiresIn = 0
//	affected, err := ormer.Engine.ID(core.PK{token.Owner, token.Name}).Cols("expires_in").Update(token)
//	if err != nil {
//		return false, nil, nil, err
//	}
//
//	err = NotifySsoWebApi(token.AccessToken)
//	if err != nil {
//		log.Println("Send logout notify failed:", err)
//	}
//
//	// case-sensitive user, failed to locate records
//	//affected, err := ormer.Engine.Where(" user = ? ", token.User).Cols("expires_in").Update(&Token{ExpiresIn: 0})
//	//if err != nil {
//	//	return false, nil, nil, err
//	//}
//
//	//// ivan 20240216
//	//// instead of just this specific token, update all tokens of that user
//	//sql := `UPDATE "token" SET "expires_in" = ? WHERE "user" = ? AND (CAST("created_time" AS timestamptz) + ("expires_in" * INTERVAL '1 second')) > CURRENT_TIMESTAMP`
//	//result, err := ormer.Engine.Exec(sql, 0, token.User)
//	//if err != nil {
//	//	return false, nil, nil, err
//	//}
//	//affected, err := result.RowsAffected()
//	//if err != nil {
//	//	return false, nil, nil, err
//	//}
//
//	application, err := getApplication(token.Owner, token.Application)
//	if err != nil {
//		return false, nil, nil, err
//	}
//
//	return affected != 0, application, token, nil
//}
