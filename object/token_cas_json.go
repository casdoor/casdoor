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
	"html"
	"reflect"
)

type CasJsonResponse struct {
	ServiceResponse CasJsonServiceResponse `json:"serviceResponse"`
}

type CasJsonServiceResponse struct {
	AuthenticationSuccess *CasJsonAuthenticationSuccess `json:"authenticationSuccess,omitempty"`
	AuthenticationFailure *CasJsonFailure               `json:"authenticationFailure,omitempty"`
	ProxySuccess          *CasJsonProxySuccess          `json:"proxySuccess,omitempty"`
	ProxyFailure          *CasJsonFailure               `json:"proxyFailure,omitempty"`
}

type CasJsonAuthenticationSuccess struct {
	User                string                 `json:"user"`
	ProxyGrantingTicket string                 `json:"proxyGrantingTicket,omitempty"`
	Proxies             []string               `json:"proxies,omitempty"`
	Attributes          map[string]interface{} `json:"attributes,omitempty"`
}

type CasJsonFailure struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type CasJsonProxySuccess struct {
	ProxyTicket string `json:"proxyTicket"`
}

func (r *CasServiceResponse) ToJson() *CasJsonResponse {
	res := &CasJsonResponse{}
	if r.Success != nil {
		res.ServiceResponse.AuthenticationSuccess = r.Success.toJson()
	}
	if r.Failure != nil {
		res.ServiceResponse.AuthenticationFailure = &CasJsonFailure{
			Code:        r.Failure.Code,
			Description: html.UnescapeString(r.Failure.Message),
		}
	}
	if r.ProxySuccess != nil {
		res.ServiceResponse.ProxySuccess = &CasJsonProxySuccess{
			ProxyTicket: r.ProxySuccess.ProxyTicket,
		}
	}
	if r.ProxyFailure != nil {
		res.ServiceResponse.ProxyFailure = &CasJsonFailure{
			Code:        r.ProxyFailure.Code,
			Description: html.UnescapeString(r.ProxyFailure.Message),
		}
	}
	return res
}

func (s *CasAuthenticationSuccess) toJson() *CasJsonAuthenticationSuccess {
	res := &CasJsonAuthenticationSuccess{
		User:                s.User,
		ProxyGrantingTicket: s.ProxyGrantingTicket,
	}
	if s.Proxies != nil && len(s.Proxies.Proxies) > 0 {
		res.Proxies = append([]string{}, s.Proxies.Proxies...)
	}

	attributes := map[string]interface{}{}
	if a := s.Attributes; a != nil {
		attributes["authenticationDate"] = a.AuthenticationDate
		attributes["longTermAuthenticationRequestTokenUsed"] = a.LongTermAuthenticationRequestTokenUsed
		attributes["isFromNewLogin"] = a.IsFromNewLogin
		if len(a.MemberOf) > 0 {
			attributes["memberOf"] = append([]string{}, a.MemberOf...)
		}

		for name, value := range map[string]string{
			"firstName":   a.FirstName,
			"lastName":    a.LastName,
			"title":       a.Title,
			"email":       a.Email,
			"affiliation": a.Affiliation,
			"avatar":      a.Avatar,
			"phone":       a.Phone,
			"displayName": a.DisplayName,
		} {
			if value != "" {
				addCasJsonAttribute(attributes, name, html.UnescapeString(value))
			}
		}

		if a.UserAttributes != nil {
			for _, attribute := range a.UserAttributes.Attributes {
				if attribute != nil && attribute.Name != "" {
					addCasJsonAttribute(attributes, attribute.Name, html.UnescapeString(attribute.Value))
				}
			}
			addCasJsonAnyAttributes(attributes, a.UserAttributes.AnyAttributes)
		}
		addCasJsonAnyAttributes(attributes, a.ExtraAttributes)
	}
	addCasJsonAnyAttributes(attributes, s.ExtraAttributes)

	if len(attributes) > 0 {
		res.Attributes = attributes
	}
	return res
}

func addCasJsonAnyAttributes(attributes map[string]interface{}, anyAttributes []*CasAnyAttribute) {
	for _, attribute := range anyAttributes {
		if attribute != nil && attribute.XMLName.Local != "" {
			addCasJsonAttribute(attributes, attribute.XMLName.Local, attribute.Value)
		}
	}
}

func addCasJsonAttribute(attributes map[string]interface{}, name string, value interface{}) {
	existing, ok := attributes[name]
	if !ok {
		attributes[name] = value
		return
	}

	values, isList := existing.([]interface{})
	if !isList {
		if reflect.DeepEqual(existing, value) {
			return
		}
		values = []interface{}{existing}
	}
	for _, v := range values {
		if reflect.DeepEqual(v, value) {
			return
		}
	}
	attributes[name] = append(values, value)
}
