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

package object

import (
	"fmt"

	"github.com/casbin/casdoor/util"
	"xorm.io/core"
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

	PublicKey  string `xorm:"mediumtext" json:"publicKey"`
	PrivateKey string `xorm:"mediumtext" json:"privateKey"`
}

func GetMaskedCert(cert *Cert) *Cert {
	if cert == nil {
		return nil
	}

	return cert
}

func GetMaskedCerts(certs []*Cert) []*Cert {
	for _, cert := range certs {
		cert = GetMaskedCert(cert)
	}
	return certs
}

func GetCertCount(owner, field, value string) int {
	session := adapter.Engine.Where("owner=?", owner)
	if field != "" && value != "" {
		session = session.And(fmt.Sprintf("%s like ?", util.SnakeString(field)), fmt.Sprintf("%%%s%%", value))
	}
	count, err := session.Count(&Cert{})
	if err != nil {
		panic(err)
	}

	return int(count)
}

func GetCerts(owner string) []*Cert {
	certs := []*Cert{}
	err := adapter.Engine.Desc("created_time").Find(&certs, &Cert{Owner: owner})
	if err != nil {
		panic(err)
	}

	return certs
}

func GetPaginationCerts(owner string, offset, limit int, field, value, sortField, sortOrder string) []*Cert {
	certs := []*Cert{}
	session := GetSession(owner, offset, limit, field, value, sortField, sortOrder)
	err := session.Find(&certs)
	if err != nil {
		panic(err)
	}

	return certs
}

func getCert(owner string, name string) *Cert {
	if owner == "" || name == "" {
		return nil
	}

	cert := Cert{Owner: owner, Name: name}
	existed, err := adapter.Engine.Get(&cert)
	if err != nil {
		panic(err)
	}

	if existed {
		return &cert
	} else {
		return nil
	}
}

func GetCert(id string) *Cert {
	owner, name := util.GetOwnerAndNameFromId(id)
	return getCert(owner, name)
}

func UpdateCert(id string, cert *Cert) bool {
	owner, name := util.GetOwnerAndNameFromId(id)
	if getCert(owner, name) == nil {
		return false
	}

	affected, err := adapter.Engine.ID(core.PK{owner, name}).AllCols().Update(cert)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func AddCert(cert *Cert) bool {
	affected, err := adapter.Engine.Insert(cert)
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func DeleteCert(cert *Cert) bool {
	affected, err := adapter.Engine.ID(core.PK{cert.Owner, cert.Name}).Delete(&Cert{})
	if err != nil {
		panic(err)
	}

	return affected != 0
}

func (p *Cert) GetId() string {
	return fmt.Sprintf("%s/%s", p.Owner, p.Name)
}
