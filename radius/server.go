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
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
)

var StateMap map[string]AccessStateContent

const StateExpiredTime = time.Second * 120

type AccessStateContent struct {
	ExpiredAt time.Time
}

func StartRadiusServer() {
	secret := conf.GetConfigString("radiusSecret")
	server := radius.PacketServer{
		Addr:         "0.0.0.0:" + conf.GetConfigString("radiusServerPort"),
		Handler:      radius.HandlerFunc(handlerRadius),
		SecretSource: radius.StaticSecretSource([]byte(secret)),
	}
	log.Printf("Starting Radius server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("StartRadiusServer() failed, err = %v", err)
	}
}

func handlerRadius(w radius.ResponseWriter, r *radius.Request) {
	switch r.Code {
	case radius.CodeAccessRequest:
		handleAccessRequest(w, r)
	case radius.CodeAccountingRequest:
		handleAccountingRequest(w, r)
	default:
		log.Printf("radius message, code = %d", r.Code)
	}
}

func handleAccessRequest(w radius.ResponseWriter, r *radius.Request) {
	username := rfc2865.UserName_GetString(r.Packet)
	password := rfc2865.UserPassword_GetString(r.Packet)
	organization := rfc2865.Class_GetString(r.Packet)
	state := rfc2865.State_GetString(r.Packet)
	log.Printf("handleAccessRequest() username=%v, org=%v, password=%v", username, organization, password)

	if organization == "" {
		organization = conf.GetConfigString("radiusDefaultOrganization")
		if organization == "" {
			organization = "built-in"
		}
	}

	var user *object.User
	var err error

	if state == "" {
		user, err = object.CheckUserPassword(organization, username, password, "en")
	} else {
		user, err = object.GetUser(fmt.Sprintf("%s/%s", organization, username))
	}

	if err != nil {
		w.Write(r.Response(radius.CodeAccessReject))
		return
	}

	if user.IsMfaEnabled() {
		mfaProp := user.GetMfaProps(object.TotpType, false)
		if mfaProp == nil {
			w.Write(r.Response(radius.CodeAccessReject))
			return
		}

		if StateMap == nil {
			StateMap = map[string]AccessStateContent{}
		}

		if state != "" {
			stateContent, ok := StateMap[state]
			if !ok {
				w.Write(r.Response(radius.CodeAccessReject))
				return
			}

			delete(StateMap, state)
			if stateContent.ExpiredAt.Before(time.Now()) {
				w.Write(r.Response(radius.CodeAccessReject))
				return
			}

			mfaUtil := object.GetMfaUtil(mfaProp.MfaType, mfaProp)
			if mfaUtil.Verify(password) != nil {
				w.Write(r.Response(radius.CodeAccessReject))
				return
			}

			w.Write(r.Response(radius.CodeAccessAccept))
			return
		}

		responseState := util.GenerateId()
		StateMap[responseState] = AccessStateContent{
			time.Now().Add(StateExpiredTime),
		}

		err = rfc2865.State_Set(r.Packet, []byte(responseState))
		if err != nil {
			w.Write(r.Response(radius.CodeAccessReject))
			return
		}

		err = rfc2865.ReplyMessage_Set(r.Packet, []byte("please enter OTP"))
		if err != nil {
			w.Write(r.Response(radius.CodeAccessReject))
			return
		}

		r.Packet.Code = radius.CodeAccessChallenge
		w.Write(r.Packet)
	}

	w.Write(r.Response(radius.CodeAccessAccept))
}

func handleAccountingRequest(w radius.ResponseWriter, r *radius.Request) {
	statusType := rfc2866.AcctStatusType_Get(r.Packet)
	username := rfc2865.UserName_GetString(r.Packet)
	organization := rfc2865.Class_GetString(r.Packet)

	if strings.Contains(username, "/") {
		organization, username = util.GetOwnerAndNameFromId(username)
	}

	log.Printf("handleAccountingRequest() username=%v, org=%v, statusType=%v", username, organization, statusType)
	w.Write(r.Response(radius.CodeAccountingResponse))
	var err error
	defer func() {
		if err != nil {
			log.Printf("handleAccountingRequest() failed, err = %v", err)
		}
	}()
	switch statusType {
	case rfc2866.AcctStatusType_Value_Start:
		// Start an accounting session
		ra := GetAccountingFromRequest(r)
		err = object.AddRadiusAccounting(ra)
	case rfc2866.AcctStatusType_Value_InterimUpdate, rfc2866.AcctStatusType_Value_Stop:
		// Interim update to an accounting session | Stop an accounting session
		var (
			newRa = GetAccountingFromRequest(r)
			oldRa *object.RadiusAccounting
		)
		oldRa, err = object.GetRadiusAccountingBySessionId(newRa.AcctSessionId)
		if err != nil {
			return
		}
		if oldRa == nil {
			if err = object.AddRadiusAccounting(newRa); err != nil {
				return
			}
		}
		stop := statusType == rfc2866.AcctStatusType_Value_Stop
		err = object.InterimUpdateRadiusAccounting(oldRa, newRa, stop)
	case rfc2866.AcctStatusType_Value_AccountingOn, rfc2866.AcctStatusType_Value_AccountingOff:
		// By default, no Accounting-On or Accounting-Off messages are sent (no acct-on-off).
	default:
		err = fmt.Errorf("unsupport statusType = %v", statusType)
	}
}
