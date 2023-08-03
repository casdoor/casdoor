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
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const Web3AuthTokenKey = "web3AuthToken"

type Web3AuthToken struct {
	Address    string `json:"address"`
	Nonce      string `json:"nonce"`
	CreateAt   uint64 `json:"createAt"`
	TypedData  string `json:"typedData"`  // typed data use for application
	Signature  string `json:"signature"`  // signature for typed data
	WalletType string `json:"walletType"` // e.g."MetaMask", "Coinbase"
}

type Web3OnboardIdProvider struct {
	Client *http.Client
}

func NewWeb3OnboardIdProvider() *Web3OnboardIdProvider {
	idp := &Web3OnboardIdProvider{}
	return idp
}

func (idp *Web3OnboardIdProvider) SetHttpClient(client *http.Client) {
	idp.Client = client
}

func (idp *Web3OnboardIdProvider) GetToken(code string) (*oauth2.Token, error) {
	web3AuthToken := Web3AuthToken{}
	if err := json.Unmarshal([]byte(code), &web3AuthToken); err != nil {
		return nil, err
	}
	token := &oauth2.Token{
		AccessToken: fmt.Sprintf("%v:%v", Web3AuthTokenKey, web3AuthToken.Address),
		TokenType:   "Bearer",
		Expiry:      time.Now().AddDate(0, 1, 0),
	}

	token = token.WithExtra(map[string]interface{}{
		Web3AuthTokenKey: web3AuthToken,
	})
	return token, nil
}

func (idp *Web3OnboardIdProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	web3AuthToken, ok := token.Extra(Web3AuthTokenKey).(Web3AuthToken)
	if !ok {
		return nil, errors.New("invalid web3AuthToken")
	}

	fmtAddress := fmt.Sprintf("%v_%v",
		strings.ReplaceAll(strings.TrimSpace(web3AuthToken.WalletType), " ", "_"),
		web3AuthToken.Address,
	)
	userInfo := &UserInfo{
		Id:          fmtAddress,
		Username:    fmtAddress,
		DisplayName: fmtAddress,
		AvatarUrl:   fmt.Sprintf("metamask:%v", forceEthereumAddress(web3AuthToken.Address)),
	}
	return userInfo, nil
}

func forceEthereumAddress(address string) string {
	// The required address to general MetaMask avatar is a string of length 42 that represents an Ethereum address.
	// This function is used to force any address as an Ethereum address
	address = strings.TrimSpace(address)
	var builder strings.Builder
	for _, ch := range address {
		builder.WriteRune(ch)
	}
	for len(builder.String()) < 42 {
		builder.WriteString("0")
	}
	if len(builder.String()) > 42 {
		return builder.String()[:42]
	}
	return builder.String()
}
