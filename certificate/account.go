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

package certificate

import (
	"crypto"

	"github.com/casbin/lego/v4/acme"
	"github.com/casbin/lego/v4/certcrypto"
	"github.com/casbin/lego/v4/lego"
	"github.com/casbin/lego/v4/registration"
	"github.com/casdoor/casdoor/proxy"
)

type Account struct {
	Email        string
	Registration *registration.Resource
	Key          crypto.PrivateKey
}

/** Implementation of the registration.User interface **/

// GetEmail returns the email address for the account.
func (a *Account) GetEmail() string {
	return a.Email
}

// GetPrivateKey returns the private RSA account key.
func (a *Account) GetPrivateKey() crypto.PrivateKey {
	return a.Key
}

// GetRegistration returns the server registration.
func (a *Account) GetRegistration() *registration.Resource {
	return a.Registration
}

func getLegoClientAndAccount(email string, privateKey string, devMode bool) (*lego.Client, *Account, error) {
	key, err := decodeEccKey(privateKey)
	if err != nil {
		return nil, nil, err
	}
	account := &Account{
		Email: email,
		Key:   key,
	}

	config := lego.NewConfig(account)
	if devMode {
		config.CADirURL = lego.LEDirectoryStaging
	} else {
		config.CADirURL = lego.LEDirectoryProduction
	}

	config.Certificate.KeyType = certcrypto.RSA2048
	config.HTTPClient = proxy.ProxyHttpClient

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, nil, err
	}

	return client, account, err
}

// GetAcmeClient Incoming an email ,a privatekey and a Boolean value that controls the opening of the test environment
// When this function is started for the first time, it will initialize the account-related configuration,
// After initializing the configuration, It will try to obtain an account based on the private key,
// if it fails, it will create an account based on the private key.
// This account will be used during the running of the program
func GetAcmeClient(email string, privateKey string, devMode bool) (*lego.Client, error) {
	// Create a user. New accounts need an email and private key to start.
	client, account, err := getLegoClientAndAccount(email, privateKey, devMode)

	// try to obtain an account based on the private key
	account.Registration, err = client.Registration.ResolveAccountByKey()
	if err != nil {
		acmeError, ok := err.(*acme.ProblemDetails)
		if !ok {
			return nil, err
		}

		if acmeError.Type != "urn:ietf:params:acme:error:accountDoesNotExist" {
			return nil, acmeError
		}

		// Failed to get account, so create an account based on the private key.
		account.Registration, err = client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}
