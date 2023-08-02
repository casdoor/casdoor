// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type MetaMaskIdProvider struct {
	Client *http.Client
}

func NewMetaMaskIdProvider() *MetaMaskIdProvider {
	idp := &MetaMaskIdProvider{}
	return idp
}

func (idp *MetaMaskIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *MetaMaskIdProvider) GetToken(code string) (*oauth2.Token, error) {
	web3AuthToken := Web3AuthToken{}
	if err := json.Unmarshal([]byte(code), &web3AuthToken); err != nil {
		return nil, err
	}
	token := &oauth2.Token{
		AccessToken: web3AuthToken.Signature,
		TokenType:   "Bearer",
		Expiry:      time.Now().AddDate(0, 1, 0),
	}

	token = token.WithExtra(map[string]interface{}{
		Web3AuthTokenKey: web3AuthToken,
	})
	return token, nil
}

func (idp *MetaMaskIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// TODO use "github.com/ethereum/go-ethereum" to check address's eth balance or transaction
	web3AuthToken, ok := token.Extra(Web3AuthTokenKey).(Web3AuthToken)
	if !ok {
		return nil, errors.New("invalid web3AuthToken")
	}
	userInfo := &UserInfo{
		Id:          web3AuthToken.Address,
		Username:    web3AuthToken.Address,
		DisplayName: web3AuthToken.Address,
		AvatarUrl:   fmt.Sprintf("metamask:%v", web3AuthToken.Address),
	}
	return userInfo, nil
}
