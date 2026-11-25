// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/casdoor/casdoor/edition"
)

const cloudSigninTicketType = "signin"

// a ticket is good for a minute, longer ones are refused whatever they say
const maxCloudSigninTicketLifetime = 2 * time.Minute

// CloudSigninTicket is the console of Casdoor Cloud signing the owner of a hosted instance in to it, signed by the
// license key. Only a hosted instance (a SaaS license) takes one, and only one made for its license.
type CloudSigninTicket struct {
	Type       string `json:"type"`
	Id         string `json:"id"`
	License    string `json:"license"`
	User       string `json:"user"`
	Operator   string `json:"operator,omitempty"`
	IssuedTime string `json:"issuedTime"`
	ExpireTime string `json:"expireTime"`
}

var (
	usedCloudTicketMu sync.Mutex
	usedCloudTickets  = map[string]time.Time{}
)

// ParseCloudSigninTicket verifies a sign-in ticket, checks that it is meant for this hosted instance and that it is
// used once
func ParseCloudSigninTicket(text string) (*CloudSigninTicket, error) {
	payload, err := edition.VerifySigned(text)
	if err != nil {
		return nil, fmt.Errorf("the sign-in ticket is invalid: %s", err.Error())
	}

	ticket := &CloudSigninTicket{}
	if err = json.Unmarshal(payload, ticket); err != nil || ticket.Type != cloudSigninTicketType || ticket.Id == "" || ticket.User == "" {
		return nil, fmt.Errorf("the sign-in ticket is malformed")
	}

	issuedTime, err := time.Parse(time.RFC3339, ticket.IssuedTime)
	if err != nil {
		return nil, fmt.Errorf("the sign-in ticket is malformed")
	}
	expireTime, err := time.Parse(time.RFC3339, ticket.ExpireTime)
	if err != nil || expireTime.Sub(issuedTime) > maxCloudSigninTicketLifetime {
		return nil, fmt.Errorf("the sign-in ticket is malformed")
	}
	now := time.Now()
	if now.After(expireTime) || issuedTime.After(now.Add(time.Minute)) {
		return nil, fmt.Errorf("the sign-in ticket has expired, open the instance from the console again")
	}

	licenseId, kind := edition.GetLicenseIdentity()
	if kind != edition.KindSaaS {
		return nil, fmt.Errorf("this Casdoor is not hosted by Casdoor Cloud, it takes no sign-in ticket")
	}
	if licenseId != ticket.License {
		return nil, fmt.Errorf("the sign-in ticket is for another Casdoor instance")
	}

	usedCloudTicketMu.Lock()
	defer usedCloudTicketMu.Unlock()
	for id, t := range usedCloudTickets {
		if now.After(t) {
			delete(usedCloudTickets, id)
		}
	}
	if _, ok := usedCloudTickets[ticket.Id]; ok {
		return nil, fmt.Errorf("the sign-in ticket has been used, open the instance from the console again")
	}
	usedCloudTickets[ticket.Id] = expireTime
	return ticket, nil
}
