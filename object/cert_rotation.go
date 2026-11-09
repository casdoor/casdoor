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

package object

import (
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
	"github.com/go-jose/go-jose/v4"
	"github.com/xorm-io/core"
)

const defaultCertExpireAlertDays = 30

// GetKeyId is the "kid" of the tokens the cert signs: the cert name until the key is rotated
func (p *Cert) GetKeyId() string {
	if p.KeyId != "" {
		return p.KeyId
	}
	return p.Name
}

// getCertificateByKeyId picks the public key that verifies a token: the key before the last rotation for the
// tokens it signed, otherwise the current one
func (p *Cert) getCertificateByKeyId(kid string) string {
	if kid != "" && p.PreviousCertificate != "" && kid == p.PreviousKeyId && kid != p.GetKeyId() {
		return p.PreviousCertificate
	}
	return p.Certificate
}

func getPreviousJsonWebKey(cert *Cert) *jose.JSONWebKey {
	if cert.Type != "x509" || cert.PreviousCertificate == "" || cert.PreviousKeyId == "" || cert.PreviousKeyId == cert.GetKeyId() {
		return nil
	}
	return newJsonWebKey(cert, cert.PreviousCertificate, cert.PreviousKeyId)
}

func getCertExpireTime(cert *Cert) (time.Time, bool) {
	expireTime := cert.ExpireTime
	if cert.Certificate != "" {
		if t, err := util.GetCertExpireTime(cert.Certificate); err == nil {
			expireTime = t
		}
	}
	if cert.Type == "SSL" && cert.DomainExpireTime != "" && expireTime == "" {
		expireTime = cert.DomainExpireTime
	}
	return parseAnyTime(expireTime)
}

// RotateCert generates a new key pair for a JWT signing cert. The tokens get a new "kid", and the previous key
// stays in the JWKS and verifies the tokens it signed until the next rotation.
func RotateCert(cert *Cert) error {
	if cert.Type != "x509" {
		return fmt.Errorf("only the x509 certs that sign tokens can be rotated, the cert: %s is %s", cert.GetId(), cert.Type)
	}
	if err := edition.Require(edition.Commercial, edition.FeatureKeyRotation); err != nil {
		return err
	}

	rotated := *cert
	rotated.Certificate = ""
	rotated.PrivateKey = ""
	if rotated.ExpireInYears <= 0 {
		rotated.ExpireInYears = 1
	}
	err := rotated.populateContent()
	if err != nil {
		return err
	}

	now := time.Now()
	rotated.PreviousKeyId = cert.GetKeyId()
	rotated.PreviousCertificate = cert.Certificate
	rotated.KeyId = fmt.Sprintf("%s-%s", cert.Name, now.Format("20060102150405"))
	rotated.RotatedTime = util.GetCurrentTime()
	if expireTime, err := util.GetCertExpireTime(rotated.Certificate); err == nil {
		rotated.ExpireTime = expireTime
	}

	_, err = ormer.Engine.ID(core.PK{cert.Owner, cert.Name}).
		Cols("certificate", "private_key", "key_id", "previous_key_id", "previous_certificate", "rotated_time", "expire_time").Update(&rotated)
	if err != nil {
		return err
	}

	*cert = rotated
	return nil
}

func getCertOrganization(cert *Cert) string {
	if cert.Owner == "admin" {
		return "built-in"
	}
	return cert.Owner
}

// CheckCertExpiration rotates the JWT signing certs and alerts on the certs that expire within the organization's
// CertExpireAlertDays
func CheckCertExpiration() error {
	certs := []*Cert{}
	err := ormer.Engine.Find(&certs)
	if err != nil {
		return err
	}

	organizations := map[string]*Organization{}
	expiring := map[string][]string{}
	for _, cert := range certs {
		orgName := getCertOrganization(cert)
		organization, ok := organizations[orgName]
		if !ok {
			organization, err = getOrganization("admin", orgName)
			if err != nil {
				return err
			}
			organizations[orgName] = organization
		}

		policy := GetSecurityPolicy(organization)
		if policy == nil {
			continue
		}
		alertDays := policy.CertExpireAlertDays
		if alertDays <= 0 {
			if !policy.AutoRotateCerts {
				continue
			}
			alertDays = defaultCertExpireAlertDays
		}

		expireTime, ok := getCertExpireTime(cert)
		if !ok || time.Until(expireTime) > time.Duration(alertDays)*24*time.Hour {
			continue
		}

		if policy.AutoRotateCerts && cert.Type == "x509" && cert.PrivateKey != "" {
			err = RotateCert(cert)
			if err != nil {
				logs.Warn("the cert: %s is not rotated: %s", cert.GetId(), err.Error())
			} else {
				logs.Info("rotated the cert: %s, the new key id is %s", cert.GetId(), cert.KeyId)
				expiring[orgName] = append(expiring[orgName], fmt.Sprintf("%s (expires at %s, rotated, new key id %s)", cert.GetId(), expireTime.Format(time.RFC3339), cert.KeyId))
				continue
			}
		}

		if policy.CertExpireAlertDays > 0 {
			expiring[orgName] = append(expiring[orgName], fmt.Sprintf("%s (%s, expires at %s)", cert.GetId(), cert.Type, expireTime.Format(time.RFC3339)))
		}
	}

	for orgName, items := range expiring {
		SendSecurityAlert(organizations[orgName], fmt.Sprintf("[%s] %d certs expire soon", orgName, len(items)), strings.Join(items, "\n"))
	}
	return nil
}
