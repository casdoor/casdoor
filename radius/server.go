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

package radius

import (
	"log"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// https://support.huawei.com/enterprise/zh/doc/EDOC1000178159/35071f9a#tab_3
func StartRadiusServer() {
	server := radius.PacketServer{
		Addr:         "0.0.0.0:" + conf.GetConfigString("radiusServerPort"),
		Handler:      radius.HandlerFunc(handlerRadius),
		SecretSource: radius.StaticSecretSource([]byte(`secret`)),
	}
	log.Printf("Starting Radius server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("StartRadiusServer() failed, err = %v", err)
	}
}

func handlerRadius(w radius.ResponseWriter, r *radius.Request) {
	switch r.Code {
	case radius.CodeAccessRequest:
		handleAccessRequest(w, r)
	default:
		log.Printf("radius message, code = %d", r.Code)
	}
}

func handleAccessRequest(w radius.ResponseWriter, r *radius.Request) {
	username := rfc2865.UserName_GetString(r.Packet)
	password := rfc2865.UserPassword_GetString(r.Packet)
	organization := parseOrganization(r.Packet)
	code := radius.CodeAccessAccept

	log.Printf("username=%v, password=%v, code=%v, org=%v", username, password, code, organization)
	if organization == "" {
		code = radius.CodeAccessReject
		w.Write(r.Response(code))
		return
	}
	_, msg := object.CheckUserPassword(organization, username, password, "en")
	if msg != "" {
		code = radius.CodeAccessReject
		w.Write(r.Response(code))
		return
	}
	w.Write(r.Response(code))
}
