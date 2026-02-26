// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package util

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"golang.org/x/net/publicsuffix"
)

func GetCertExpireTime(s string) (string, error) {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return "", errors.New("getCertExpireTime() error, block should not be nil")
	} else if block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("getCertExpireTime() error, block.Type should be \"CERTIFICATE\" instead of %s", block.Type)
	}

	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}

	t := certificate.NotAfter
	return t.Local().Format(time.RFC3339), nil
}

func GetBaseDomain(domain string) (string, error) {
	// abc.com -> abc.com
	// abc.com.it -> abc.com.it
	// subdomain.abc.io -> abc.io
	// subdomain.abc.org.us -> abc.org.us
	return publicsuffix.EffectiveTLDPlusOne(domain)
}
