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
	"fmt"
	"time"

	"github.com/casbin/lego/v4/certificate"
	"github.com/casbin/lego/v4/challenge/dns01"
	"github.com/casbin/lego/v4/cmd"
	"github.com/casbin/lego/v4/lego"
	"github.com/casbin/lego/v4/providers/dns/alidns"
	"github.com/casbin/lego/v4/providers/dns/godaddy"
)

type AliConf struct {
	Domains       []string // The domain names for which you want to apply for a certificate
	AccessKey     string   // Aliyun account's AccessKey, if this is not empty, Secret is required.
	Secret        string
	RAMRole       string // Use Ramrole to control aliyun account
	SecurityToken string // Optional
	Path          string // The path to store cert file
	Timeout       int    // Maximum waiting time for certificate application, in minutes
}

type GodaddyConf struct {
	Domains   []string // The domain names for which you want to apply for a certificate
	APIKey    string   // GoDaddy account's API Key
	APISecret string
	Path      string // The path to store cert file
	Timeout   int    // Maximum waiting time for certificate application, in minutes
}

// getCert Verify domain ownership, then obtain a certificate, and finally store it locally.
// Need to pass in an AliConf struct, some parameters are required, other parameters can be left blank
func getAliCert(client *lego.Client, conf AliConf) (string, string) {
	if conf.Timeout <= 0 {
		conf.Timeout = 3
	}

	config := alidns.NewDefaultConfig()
	config.PropagationTimeout = time.Duration(conf.Timeout) * time.Minute
	config.APIKey = conf.AccessKey
	config.SecretKey = conf.Secret
	config.RAMRole = conf.RAMRole
	config.SecurityToken = conf.SecurityToken

	dnsProvider, err := alidns.NewDNSProvider(config)
	if err != nil {
		panic(err)
	}

	// Choose a local DNS service provider to increase the authentication speed
	servers := []string{"223.5.5.5:53"}
	err = client.Challenge.SetDNS01Provider(dnsProvider, dns01.CondOption(len(servers) > 0, dns01.AddRecursiveNameservers(dns01.ParseNameservers(servers))), dns01.DisableCompletePropagationRequirement())
	if err != nil {
		panic(err)
	}

	// Obtain the certificate
	request := certificate.ObtainRequest{
		Domains: conf.Domains,
		Bundle:  true,
	}
	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		panic(err)
	}

	return string(cert.Certificate), string(cert.PrivateKey)
}

func getGoDaddyCert(client *lego.Client, conf GodaddyConf) (string, string) {
	if conf.Timeout <= 0 {
		conf.Timeout = 3
	}

	config := godaddy.NewDefaultConfig()
	config.PropagationTimeout = time.Duration(conf.Timeout) * time.Minute
	config.PollingInterval = time.Duration(conf.Timeout) * time.Minute / 9
	config.APIKey = conf.APIKey
	config.APISecret = conf.APISecret

	dnsProvider, err := godaddy.NewDNSProvider(config)
	if err != nil {
		panic(err)
	}

	// Choose a local DNS service provider to increase the authentication speed
	servers := []string{"223.5.5.5:53"}
	err = client.Challenge.SetDNS01Provider(dnsProvider, dns01.CondOption(len(servers) > 0, dns01.AddRecursiveNameservers(dns01.ParseNameservers(servers))), dns01.DisableCompletePropagationRequirement())
	if err != nil {
		panic(err)
	}

	// Obtain the certificate
	request := certificate.ObtainRequest{
		Domains: conf.Domains,
		Bundle:  true,
	}
	cert, err := client.Certificate.Obtain(request)
	if err != nil {
		panic(err)
	}

	return string(cert.Certificate), string(cert.PrivateKey)
}

func ObtainCertificateAli(client *lego.Client, domain string, accessKey string, accessSecret string) (string, string) {
	conf := AliConf{
		Domains:       []string{fmt.Sprintf("*.%s", domain), domain},
		AccessKey:     accessKey,
		Secret:        accessSecret,
		RAMRole:       "",
		SecurityToken: "",
		Path:          "",
		Timeout:       3,
	}
	return getAliCert(client, conf)
}

func ObtainCertificateGoDaddy(client *lego.Client, domain string, accessKey string, accessSecret string) (string, string) {
	conf := GodaddyConf{
		Domains:   []string{fmt.Sprintf("*.%s", domain), domain},
		APIKey:    accessKey,
		APISecret: accessSecret,
		Path:      "",
		Timeout:   3,
	}
	return getGoDaddyCert(client, conf)
}

func SaveCert(path, filename string, cert *certificate.Resource) {
	// Store the certificate file locally
	certsStorage := cmd.NewCertificatesStorageLib(path, filename, true)
	certsStorage.CreateRootFolder()
	certsStorage.SaveResource(cert)
}
