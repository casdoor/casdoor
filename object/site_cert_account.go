// Copyright 2023 The casbin Authors. All Rights Reserved.
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
	"fmt"

	"github.com/casbin/lego/v4/acme"
	"github.com/casbin/lego/v4/certcrypto"
	"github.com/casbin/lego/v4/lego"
	"github.com/casbin/lego/v4/registration"
	"github.com/casdoor/casdoor/certificate"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/proxy"
)

func getLegoClientAndAccount(email string, privateKey string, devMode bool, useProxy bool) (*lego.Client, *certificate.Account, error) {
	eccKey, err := decodeEccKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	account := &certificate.Account{
		Email: email,
		Key:   eccKey,
	}

	config := lego.NewConfig(account)
	if devMode {
		config.CADirURL = lego.LEDirectoryStaging
	} else {
		config.CADirURL = lego.LEDirectoryProduction
	}

	config.Certificate.KeyType = certcrypto.RSA2048

	if useProxy {
		config.HTTPClient = proxy.ProxyHttpClient
	} else {
		config.HTTPClient = proxy.DefaultHttpClient
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, nil, err
	}

	return client, account, nil
}

func getAcmeClient(email string, privateKey string, devMode bool, useProxy bool) (*lego.Client, error) {
	// Create a user. New accounts need an email and private key to start.
	client, account, err := getLegoClientAndAccount(email, privateKey, devMode, useProxy)
	if err != nil {
		return nil, err
	}

	// try to obtain an account based on the private key
	account.Registration, err = client.Registration.ResolveAccountByKey()
	if err != nil {
		acmeError, ok := err.(*acme.ProblemDetails)
		if !ok {
			return nil, err
		}

		if acmeError.Type != "urn:ietf:params:acme:error:accountDoesNotExist" {
			return nil, err
		}

		// Failed to get account, so create an account based on the private key.
		account.Registration, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

func GetAcmeClient(useProxy bool) (*lego.Client, error) {
	acmeEmail := conf.GetConfigString("acmeEmail")
	acmePrivateKey := conf.GetConfigString("acmePrivateKey")
	if acmeEmail == "" {
		return nil, fmt.Errorf("acmeEmail should not be empty")
	}
	if acmePrivateKey == "" {
		return nil, fmt.Errorf("acmePrivateKey should not be empty")
	}

	return getAcmeClient(acmeEmail, acmePrivateKey, false, useProxy)
}
