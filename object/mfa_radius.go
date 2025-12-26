// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"context"
	"errors"
	"fmt"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type RadiusMfa struct {
	*MfaProps
	provider *Provider
}

func (mfa *RadiusMfa) Initiate(userId string, issuer string) (*MfaProps, error) {
	mfaProps := MfaProps{
		MfaType: mfa.MfaType,
	}
	return &mfaProps, nil
}

func (mfa *RadiusMfa) SetupVerify(passCode string) error {
	if mfa.Secret == "" {
		return errors.New("RADIUS username is required")
	}

	if mfa.provider == nil {
		return errors.New("RADIUS provider is not configured")
	}

	return mfa.authenticateWithRadius(mfa.Secret, passCode)
}

func (mfa *RadiusMfa) Enable(user *User) error {
	columns := []string{"recovery_codes", "preferred_mfa_type", "mfa_radius_enabled", "mfa_radius_username", "mfa_radius_provider"}

	user.RecoveryCodes = append(user.RecoveryCodes, mfa.RecoveryCodes...)
	if user.PreferredMfaType == "" {
		user.PreferredMfaType = mfa.MfaType
	}

	user.MfaRadiusEnabled = true
	user.MfaRadiusUsername = mfa.Secret
	user.MfaRadiusProvider = mfa.URL

	_, err := UpdateUser(user.GetId(), user, columns, false)
	if err != nil {
		return err
	}

	return nil
}

func (mfa *RadiusMfa) Verify(passCode string) error {
	if mfa.Secret == "" {
		return errors.New("RADIUS username is required")
	}

	if mfa.provider == nil {
		return errors.New("RADIUS provider is not configured")
	}

	return mfa.authenticateWithRadius(mfa.Secret, passCode)
}

func (mfa *RadiusMfa) authenticateWithRadius(username, password string) error {
	if mfa.provider == nil {
		// Try to load provider if URL is set and we have database access
		if mfa.URL != "" && ormer != nil && ormer.Engine != nil {
			provider, err := GetProvider(mfa.URL)
			if err != nil {
				return fmt.Errorf("failed to load RADIUS provider: %v", err)
			}
			if provider == nil {
				return errors.New("RADIUS provider not found")
			}
			mfa.provider = provider
		} else {
			return errors.New("RADIUS provider is not configured")
		}
	}

	// Create RADIUS packet
	packet := radius.New(radius.CodeAccessRequest, []byte(mfa.provider.ClientSecret))
	if err := rfc2865.UserName_SetString(packet, username); err != nil {
		return fmt.Errorf("failed to set RADIUS username: %v", err)
	}
	if err := rfc2865.UserPassword_SetString(packet, password); err != nil {
		return fmt.Errorf("failed to set RADIUS password: %v", err)
	}

	// Send request to RADIUS server
	address := fmt.Sprintf("%s:%d", mfa.provider.Host, mfa.provider.Port)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := radius.Exchange(ctx, packet, address)
	if err != nil {
		return fmt.Errorf("RADIUS authentication failed: %v", err)
	}

	if response.Code == radius.CodeAccessAccept {
		return nil
	}

	return errors.New("RADIUS authentication rejected")
}

func NewRadiusMfaUtil(config *MfaProps) *RadiusMfa {
	if config == nil {
		config = &MfaProps{
			MfaType: RadiusType,
		}
	}

	radiusMfa := &RadiusMfa{
		MfaProps: config,
	}

	// Load provider if URL is specified and ormer is initialized
	if config.URL != "" && ormer != nil && ormer.Engine != nil {
		provider, err := GetProvider(config.URL)
		if err == nil && provider != nil {
			radiusMfa.provider = provider
		}
	}

	return radiusMfa
}
