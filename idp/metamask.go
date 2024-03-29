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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"golang.org/x/oauth2"
)

type EIP712Message struct {
	Domain struct {
		ChainId string `json:"chainId"`
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"domain"`
	Message struct {
		Prompt   string `json:"prompt"`
		Nonce    string `json:"nonce"`
		CreateAt string `json:"createAt"`
	} `json:"message"`
	PrimaryType string `json:"primaryType"`
	Types       struct {
		EIP712Domain []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"EIP712Domain"`
		AuthRequest []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"AuthRequest"`
	} `json:"types"`
}

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
	valid, err := VerifySignature(web3AuthToken.Address, web3AuthToken.TypedData, web3AuthToken.Signature)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("invalid signature")
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

func VerifySignature(userAddress string, originalMessage string, signatureHex string) (bool, error) {
	var eip712Mes EIP712Message
	err := json.Unmarshal([]byte(originalMessage), &eip712Mes)
	if err != nil {
		return false, fmt.Errorf("invalid signature (Error parsing JSON)")
	}

	createAtTime, err := time.Parse("2006/1/2 15:04:05", eip712Mes.Message.CreateAt)
	currentTime := time.Now()
	if createAtTime.Before(currentTime.Add(-1*time.Minute)) && createAtTime.After(currentTime) {
		return false, fmt.Errorf("invalid signature (signature does not meet time requirements)")
	}

	if !strings.HasPrefix(signatureHex, "0x") {
		signatureHex = "0x" + signatureHex
	}

	signatureBytes, err := hex.DecodeString(signatureHex[2:])
	if err != nil {
		return false, err
	}

	if signatureBytes[64] != 27 && signatureBytes[64] != 28 {
		return false, fmt.Errorf("invalid signature (incorrect recovery id)")
	}
	signatureBytes[64] -= 27

	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len([]byte(originalMessage)), []byte(originalMessage))
	hash := crypto.Keccak256Hash([]byte(msg))

	pubKey, err := crypto.SigToPub(hash.Bytes(), signatureBytes)
	if err != nil {
		return false, err
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	return strings.EqualFold(recoveredAddr.Hex(), userAddress), nil
}
