// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
)

type CasServiceResponse struct {
	XMLName      xml.Name `xml:"cas:serviceResponse" json:"-"`
	Xmlns        string   `xml:"xmlns:cas,attr"`
	Failure      *CasAuthenticationFailure
	Success      *CasAuthenticationSuccess
	ProxySuccess *CasProxySuccess
	ProxyFailure *CasProxyFailure
}

type CasAuthenticationFailure struct {
	XMLName xml.Name `xml:"cas:authenticationFailure" json:"-"`
	Code    string   `xml:"code,attr"`
	Message string   `xml:",innerxml"`
}

type CasAuthenticationSuccess struct {
	XMLName             xml.Name           `xml:"cas:authenticationSuccess" json:"-"`
	User                string             `xml:"cas:user"`
	ProxyGrantingTicket string             `xml:"cas:proxyGrantingTicket,omitempty"`
	Proxies             *CasProxies        `xml:"cas:proxies"`
	Attributes          *CasAttributes     `xml:"cas:attributes"`
	ExtraAttributes     []*CasAnyAttribute `xml:",any"`
}

type CasProxies struct {
	XMLName xml.Name `xml:"cas:proxies" json:"-"`
	Proxies []string `xml:"cas:proxy"`
}

type CasAttributes struct {
	XMLName                                xml.Name  `xml:"cas:attributes" json:"-"`
	AuthenticationDate                     time.Time `xml:"cas:authenticationDate"`
	LongTermAuthenticationRequestTokenUsed bool      `xml:"cas:longTermAuthenticationRequestTokenUsed"`
	IsFromNewLogin                         bool      `xml:"cas:isFromNewLogin"`
	MemberOf                               []string  `xml:"cas:memberOf"`
	UserAttributes                         *CasUserAttributes
	ExtraAttributes                        []*CasAnyAttribute `xml:",any"`
}

type CasUserAttributes struct {
	XMLName       xml.Name             `xml:"cas:userAttributes" json:"-"`
	Attributes    []*CasNamedAttribute `xml:"cas:attribute"`
	AnyAttributes []*CasAnyAttribute   `xml:",any"`
}

type CasNamedAttribute struct {
	XMLName xml.Name `xml:"cas:attribute" json:"-"`
	Name    string   `xml:"name,attr,omitempty"`
	Value   string   `xml:",innerxml"`
}

type CasAnyAttribute struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type CasAuthenticationSuccessWrapper struct {
	AuthenticationSuccess *CasAuthenticationSuccess // the token we issued
	Service               string                    //to which service this token is issued
}

type CasProxySuccess struct {
	XMLName     xml.Name `xml:"cas:proxySuccess" json:"-"`
	ProxyTicket string   `xml:"cas:proxyTicket"`
}
type CasProxyFailure struct {
	XMLName xml.Name `xml:"cas:proxyFailure" json:"-"`
	Code    string   `xml:"code,attr"`
	Message string   `xml:",innerxml"`
}

//st is short for service ticket
var stToServiceResponse sync.Map

//pgt is short for proxy granting ticket
var pgtToServiceResponse sync.Map

func StoreCasTokenForPgt(token *CasAuthenticationSuccess, service string) string {
	pgt := fmt.Sprintf("PGT-%s", util.GenerateId())
	pgtToServiceResponse.Store(pgt, &CasAuthenticationSuccessWrapper{
		AuthenticationSuccess: token,
		Service:               service,
	})
	return pgt
}

func GenerateId() {
	panic("unimplemented")
}

func GetCasTokenByPgt(pgt string) (bool, *CasAuthenticationSuccess, string) {
	if responseWrapperType, ok := pgtToServiceResponse.LoadAndDelete(pgt); ok {
		responseWrapperTypeCast := responseWrapperType.(*CasAuthenticationSuccessWrapper)
		return true, responseWrapperTypeCast.AuthenticationSuccess, responseWrapperTypeCast.Service
	}
	return false, nil, ""
}

func GetCasTokenByTicket(ticket string) (bool, *CasAuthenticationSuccess, string) {
	if responseWrapperType, ok := stToServiceResponse.LoadAndDelete(ticket); ok {
		responseWrapperTypeCast := responseWrapperType.(*CasAuthenticationSuccessWrapper)
		return true, responseWrapperTypeCast.AuthenticationSuccess, responseWrapperTypeCast.Service
	}
	return false, nil, ""
}

func StoreCasTokenForProxyTicket(token *CasAuthenticationSuccess, targetService string) string {
	proxyTicket := fmt.Sprintf("PT-%s", util.GenerateId())
	stToServiceResponse.Store(proxyTicket, &CasAuthenticationSuccessWrapper{
		AuthenticationSuccess: token,
		Service:               targetService,
	})
	return proxyTicket
}

func GenerateCasToken(userId string, service string) (string, error) {

	if user := GetUser(userId); user != nil {
		authenticationSuccess := CasAuthenticationSuccess{
			User: user.Name,
			Attributes: &CasAttributes{
				AuthenticationDate: time.Now(),
				UserAttributes:     &CasUserAttributes{},
			},
			ProxyGrantingTicket: fmt.Sprintf("PGTIOU-%s", util.GenerateId()),
		}
		data, _ := json.Marshal(user)
		tmp := map[string]string{}
		json.Unmarshal(data, &tmp)
		for k, v := range tmp {
			if v != "" {
				authenticationSuccess.Attributes.UserAttributes.Attributes = append(authenticationSuccess.Attributes.UserAttributes.Attributes, &CasNamedAttribute{
					Name:  k,
					Value: v,
				})
			}
		}
		st := fmt.Sprintf("ST-%d", rand.Int())
		stToServiceResponse.Store(st, &CasAuthenticationSuccessWrapper{
			AuthenticationSuccess: &authenticationSuccess,
			Service:               service,
		})
		return st, nil
	} else {
		return "", fmt.Errorf("invalid user Id")
	}

}

func (c *CasAuthenticationSuccess) DeepCopy() CasAuthenticationSuccess {
	res := *c
	//copy proxy
	if c.Proxies != nil {
		tmp := c.Proxies.DeepCopy()
		res.Proxies = &tmp
	}
	if c.Attributes != nil {
		tmp := c.Attributes.DeepCopy()
		res.Attributes = &tmp
	}
	res.ExtraAttributes = make([]*CasAnyAttribute, len(c.ExtraAttributes))
	for i, e := range c.ExtraAttributes {
		tmp := *e
		res.ExtraAttributes[i] = &tmp
	}
	return res
}

func (c *CasProxies) DeepCopy() CasProxies {
	res := CasProxies{
		Proxies: make([]string, len(c.Proxies)),
	}
	copy(res.Proxies, c.Proxies)
	return res
}

func (c *CasAttributes) DeepCopy() CasAttributes {
	res := *c
	if c.MemberOf != nil {
		res.MemberOf = make([]string, len(c.MemberOf))
		copy(res.MemberOf, c.MemberOf)
	}
	tmp := c.UserAttributes.DeepCopy()
	res.UserAttributes = &tmp

	res.ExtraAttributes = make([]*CasAnyAttribute, len(c.ExtraAttributes))
	for i, e := range c.ExtraAttributes {
		tmp := *e
		res.ExtraAttributes[i] = &tmp
	}
	return res

}

func (c *CasUserAttributes) DeepCopy() CasUserAttributes {
	res := CasUserAttributes{
		AnyAttributes: make([]*CasAnyAttribute, len(c.AnyAttributes)),
		Attributes:    make([]*CasNamedAttribute, len(c.Attributes)),
	}
	for i, a := range c.AnyAttributes {
		var tmp = *a
		res.AnyAttributes[i] = &tmp
	}
	for i, a := range c.Attributes {
		var tmp = *a
		res.Attributes[i] = &tmp
	}
	return res
}
