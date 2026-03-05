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
	"fmt"
	"time"

	"github.com/casdoor/casdoor/certificate"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
	"golang.org/x/net/publicsuffix"
)

type Cert struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`

	DisplayName     string `xorm:"varchar(100)" json:"displayName"`
	Scope           string `xorm:"varchar(100)" json:"scope"`
	Type            string `xorm:"varchar(100)" json:"type"`
	CryptoAlgorithm string `xorm:"varchar(100)" json:"cryptoAlgorithm"`
	BitSize         int    `json:"bitSize"`
	ExpireInYears   int    `json:"expireInYears"`

	ExpireTime       string `xorm:"varchar(100)" json:"expireTime"`
	DomainExpireTime string `xorm:"varchar(100)" json:"domainExpireTime"`
	Provider         string `xorm:"varchar(100)" json:"provider"`
	Account          string `xorm:"varchar(100)" json:"account"`
	AccessKey        string `xorm:"varchar(100)" json:"accessKey"`
	AccessSecret     string `xorm:"varchar(100)" json:"accessSecret"`

	Certificate string `xorm:"mediumtext" json:"certificate"`
	PrivateKey  string `xorm:"mediumtext" json:"privateKey"`
}

func GetMaskedCert(cert *Cert) *Cert {
	if cert == nil {
		return nil
	}

	return cert
}

func GetMaskedCerts(certs []*Cert, err error) ([]*Cert, error) {
	if err != nil {
		return nil, err
	}

	for _, cert := range certs {
		cert = GetMaskedCert(cert)
	}
	return certs, nil
}

func GetCertCount(owner, field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Where("owner = ? or owner = ? ", "admin", owner).Count(&Cert{})
}

func GetCerts(owner string) ([]*Cert, error) {
	certs := []*Cert{}
	db := ormer.Engine.NewSession()
	if owner != "" {
		db = db.Where("owner = ? or owner = ? ", "admin", owner)
	}
	err := db.Desc("created_time").Find(&certs, &Cert{})
	if err != nil {
		return certs, err
	}

	return certs, nil
}

func GetPaginationCerts(owner string, offset, limit int, field, value, sortField, sortOrder string) ([]*Cert, error) {
	certs := []*Cert{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Where("owner = ? or owner = ? ", "admin", owner).Find(&certs)
	if err != nil {
		return certs, err
	}

	return certs, nil
}

func GetGlobalCertsCount(field, value string) (int64, error) {
	session := GetSession("", -1, -1, field, value, "", "")
	return session.Count(&Cert{})
}

func GetGlobalCerts() ([]*Cert, error) {
	certs := []*Cert{}
	err := ormer.Engine.Desc("created_time").Find(&certs)
	if err != nil {
		return certs, err
	}

	return certs, nil
}

func GetPaginationGlobalCerts(offset, limit int, field, value, sortField, sortOrder string) ([]*Cert, error) {
	certs := []*Cert{}
	session := GetSession("", offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&certs)
	if err != nil {
		return certs, err
	}

	return certs, nil
}

func getCert(owner string, name string) (*Cert, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	cert := Cert{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&cert)
	if err != nil {
		return &cert, err
	}

	if existed {
		return &cert, nil
	} else {
		return nil, nil
	}
}

func getCertByName(name string) (*Cert, error) {
	if name == "" {
		return nil, nil
	}

	cert := Cert{Name: name}
	existed, err := ormer.Engine.Get(&cert)
	if err != nil {
		return &cert, nil
	}

	if existed {
		return &cert, nil
	} else {
		return nil, nil
	}
}

func GetCert(id string) (*Cert, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	cert, err := getCert(owner, name)
	if cert == nil && owner != "admin" {
		return getCert("admin", name)
	} else {
		return cert, err
	}
}

func UpdateCert(id string, cert *Cert) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	if c, err := getCert(owner, name); err != nil {
		return false, err
	} else if c == nil {
		return false, nil
	}

	if name != cert.Name {
		err = certChangeTrigger(name, cert.Name)
		if err != nil {
			return false, err
		}
	}

	err = cert.populateContent()
	if err != nil {
		return false, err
	}

	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(cert)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func AddCert(cert *Cert) (bool, error) {
	err := cert.populateContent()
	if err != nil {
		return false, err
	}

	affected, err := ormer.Engine.Insert(cert)
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func DeleteCert(cert *Cert) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{cert.Owner, cert.Name}).Delete(&Cert{})
	if err != nil {
		return false, err
	}

	return affected != 0, nil
}

func (p *Cert) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}

func (p *Cert) populateContent() error {
	if p.Certificate != "" && p.PrivateKey != "" {
		return nil
	}

	if p.Type == "SSL" {
		if p.Certificate != "" {
			expireTime, err := util.GetCertExpireTime(p.Certificate)
			if err != nil {
				return err
			}

			p.ExpireTime = expireTime
		} else {
			p.ExpireTime = ""
		}
		return nil
	}

	if len(p.CryptoAlgorithm) < 3 {
		err := fmt.Errorf("populateContent() error, unsupported crypto algorithm: %s", p.CryptoAlgorithm)
		return err
	}

	if p.CryptoAlgorithm == "RSA" {
		p.CryptoAlgorithm = "RS256"
	}

	sigAlgorithm := p.CryptoAlgorithm[:2]
	shaSize, err := util.ParseIntWithError(p.CryptoAlgorithm[2:])
	if err != nil {
		return err
	}

	var certificate, privateKey string
	if sigAlgorithm == "RS" {
		certificate, privateKey, err = generateRsaKeys(p.BitSize, shaSize, p.ExpireInYears, p.Name, p.Owner)
	} else if sigAlgorithm == "ES" {
		certificate, privateKey, err = generateEsKeys(shaSize, p.ExpireInYears, p.Name, p.Owner)
	} else if sigAlgorithm == "PS" {
		certificate, privateKey, err = generateRsaPssKeys(p.BitSize, shaSize, p.ExpireInYears, p.Name, p.Owner)
	} else {
		err = fmt.Errorf("populateContent() error, unsupported signature algorithm: %s", sigAlgorithm)
	}
	if err != nil {
		return err
	}

	p.Certificate = certificate
	p.PrivateKey = privateKey
	return nil
}

func RenewCert(cert *Cert) (bool, error) {
	useProxy := false
	if cert.Provider == "GoDaddy" {
		useProxy = true
	}

	client, err := GetAcmeClient(useProxy)
	if err != nil {
		return false, err
	}

	var certStr, privateKey string
	if cert.Provider == "Aliyun" {
		certStr, privateKey, err = certificate.ObtainCertificateAli(client, cert.Name, cert.AccessKey, cert.AccessSecret)
	} else if cert.Provider == "GoDaddy" {
		certStr, privateKey, err = certificate.ObtainCertificateGoDaddy(client, cert.Name, cert.AccessKey, cert.AccessSecret)
	} else {
		return false, fmt.Errorf("unknown provider: %s", cert.Provider)
	}

	if err != nil {
		return false, err
	}

	expireTime, err := util.GetCertExpireTime(certStr)
	if err != nil {
		return false, err
	}

	cert.ExpireTime = expireTime
	cert.Certificate = certStr
	cert.PrivateKey = privateKey

	return UpdateCert(cert.GetId(), cert)
}

func getCertByApplication(application *Application) (*Cert, error) {
	if application.Cert != "" {
		return getCertByName(application.Cert)
	} else {
		return GetDefaultCert()
	}
}

func GetDefaultCert() (*Cert, error) {
	return getCert("admin", "cert-built-in")
}

func certChangeTrigger(oldName string, newName string) error {
	session := ormer.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	application := new(Application)
	application.Cert = newName
	_, err = session.Where("cert=?", oldName).Update(application)
	if err != nil {
		return err
	}

	return session.Commit()
}

func getBaseDomain(domain string) (string, error) {
	// abc.com -> abc.com
	// abc.com.it -> abc.com.it
	// subdomain.abc.io -> abc.io
	// subdomain.abc.org.us -> abc.org.us
	return publicsuffix.EffectiveTLDPlusOne(domain)
}

func GetCertByDomain(domain string) (*Cert, error) {
	if domain == "" {
		return nil, fmt.Errorf("GetCertByDomain() error: domain should not be empty")
	}

	cert, ok := certMap[domain]
	if ok {
		return cert, nil
	}

	baseDomain, err := getBaseDomain(domain)
	if err != nil {
		return nil, err
	}

	cert, ok = certMap[baseDomain]
	if ok {
		return cert, nil
	}

	return nil, nil
}

func getCertMap() (map[string]*Cert, error) {
	certs, err := GetGlobalCerts()
	if err != nil {
		return nil, err
	}

	res := map[string]*Cert{}
	for _, cert := range certs {
		res[cert.Name] = cert
	}
	return res, nil
}

func (p *Cert) isCertNearExpire() (bool, error) {
	if p.ExpireTime == "" {
		return true, nil
	}

	expireTime, err := time.Parse(time.RFC3339, p.ExpireTime)
	if err != nil {
		return false, err
	}

	now := time.Now()
	duration := expireTime.Sub(now)
	res := duration <= 7*24*time.Hour

	return res, nil
}
