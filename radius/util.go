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
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

func GetAccountingFromRequest(r *radius.Request) *object.RadiusAccounting {
	acctInputOctets := int(rfc2866.AcctInputOctets_Get(r.Packet))
	acctInputGigawords := int(rfc2869.AcctInputGigawords_Get(r.Packet))
	acctOutputOctets := int(rfc2866.AcctOutputOctets_Get(r.Packet))
	acctOutputGigawords := int(rfc2869.AcctOutputGigawords_Get(r.Packet))
	organization := rfc2865.Class_GetString(r.Packet)
	getAcctStartTime := func(sessionTime int) time.Time {
		m, _ := time.ParseDuration(fmt.Sprintf("-%ds", sessionTime))
		return time.Now().Add(m)
	}
	ra := &object.RadiusAccounting{
		Owner:       organization,
		Name:        "ra_" + util.GenerateId()[:6],
		CreatedTime: time.Now(),

		Username:    rfc2865.UserName_GetString(r.Packet),
		ServiceType: int64(rfc2865.ServiceType_Get(r.Packet)),

		NasId:       rfc2865.NASIdentifier_GetString(r.Packet),
		NasIpAddr:   rfc2865.NASIPAddress_Get(r.Packet).String(),
		NasPortId:   rfc2869.NASPortID_GetString(r.Packet),
		NasPortType: int64(rfc2865.NASPortType_Get(r.Packet)),
		NasPort:     int64(rfc2865.NASPort_Get(r.Packet)),

		FramedIpAddr:    rfc2865.FramedIPAddress_Get(r.Packet).String(),
		FramedIpNetmask: rfc2865.FramedIPNetmask_Get(r.Packet).String(),

		AcctSessionId:      rfc2866.AcctSessionID_GetString(r.Packet),
		AcctSessionTime:    int64(rfc2866.AcctSessionTime_Get(r.Packet)),
		AcctInputTotal:     int64(acctInputOctets) + int64(acctInputGigawords)*4*1024*1024*1024,
		AcctOutputTotal:    int64(acctOutputOctets) + int64(acctOutputGigawords)*4*1024*1024*1024,
		AcctInputPackets:   int64(rfc2866.AcctInputPackets_Get(r.Packet)),
		AcctOutputPackets:  int64(rfc2866.AcctInputPackets_Get(r.Packet)),
		AcctStartTime:      getAcctStartTime(int(rfc2866.AcctSessionTime_Get(r.Packet))),
		AcctTerminateCause: int64(rfc2866.AcctTerminateCause_Get(r.Packet)),
		LastUpdate:         time.Now(),
	}
	return ra
}
