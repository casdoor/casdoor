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

func (mfa *RadiusMfa) Initiate(userId string) (*MfaProps, error) {
	if mfa.Secret == "" {
		return nil, errors.New("RADIUS provider not configured")
	}

	mfaProps := MfaProps{
		MfaType: mfa.MfaType,
		Secret:  mfa.Secret,
	}
	return &mfaProps, nil
}

func (mfa *RadiusMfa) SetupVerify(passcode string) error {
	if mfa.provider == nil {
		return errors.New("RADIUS provider not configured")
	}

	// For setup verification, we need the username from the config
	username := mfa.CountryCode // We'll use CountryCode field to pass the username temporarily
	if username == "" {
		return errors.New("RADIUS username not provided")
	}

	return mfa.verifyWithRadius(username, passcode)
}

func (mfa *RadiusMfa) Enable(user *User) error {
	if mfa.provider == nil {
		return errors.New("RADIUS provider not configured")
	}

	columns := []string{"recovery_codes", "preferred_mfa_type", "radius_secret", "radius_provider", "radius_username"}

	user.RecoveryCodes = append(user.RecoveryCodes, mfa.RecoveryCodes...)
	if user.PreferredMfaType == "" {
		user.PreferredMfaType = mfa.MfaType
	}
	user.RadiusSecret = mfa.provider.ClientSecret
	user.RadiusProvider = mfa.provider.GetId()
	user.RadiusUsername = mfa.CountryCode // Username passed via CountryCode field during setup

	_, err := updateUser(user.GetId(), user, columns)
	if err != nil {
		return err
	}

	return nil
}

func (mfa *RadiusMfa) Verify(passcode string) error {
	if mfa.provider == nil {
		return errors.New("RADIUS provider not configured")
	}

	username := mfa.CountryCode
	if username == "" {
		return errors.New("RADIUS username not configured")
	}

	return mfa.verifyWithRadius(username, passcode)
}

func (mfa *RadiusMfa) verifyWithRadius(username, passcode string) error {
	if mfa.provider == nil {
		return errors.New("RADIUS provider not configured")
	}

	// Build RADIUS Access-Request packet
	packet := radius.New(radius.CodeAccessRequest, []byte(mfa.provider.ClientSecret))
	if err := rfc2865.UserName_SetString(packet, username); err != nil {
		return fmt.Errorf("failed to set RADIUS username: %v", err)
	}
	if err := rfc2865.UserPassword_SetString(packet, passcode); err != nil {
		return fmt.Errorf("failed to set RADIUS password: %v", err)
	}

	// Construct RADIUS server address
	radiusAddr := fmt.Sprintf("%s:%d", mfa.provider.Host, mfa.provider.Port)

	// Send RADIUS request with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, err := radius.Exchange(ctx, packet, radiusAddr)
	if err != nil {
		return fmt.Errorf("RADIUS server communication failed: %v", err)
	}

	if response.Code == radius.CodeAccessAccept {
		return nil
	} else if response.Code == radius.CodeAccessReject {
		return errors.New("RADIUS authentication failed: invalid passcode")
	} else {
		return fmt.Errorf("unexpected RADIUS response code: %d", response.Code)
	}
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

	// Load provider if Secret contains provider ID (format: owner/name)
	if config.Secret != "" {
		// Check if the secret is in the correct format (owner/name)
		// Only try to load if it contains a slash
		if len(config.Secret) > 0 && containsSlash(config.Secret) {
			provider, err := GetProvider(config.Secret)
			if err == nil && provider != nil {
				radiusMfa.provider = provider
			}
		}
	}

	return radiusMfa
}

func containsSlash(s string) bool {
	for _, c := range s {
		if c == '/' {
			return true
		}
	}
	return false
}

// GetRadiusMfaProvider returns a RADIUS provider for MFA
func GetRadiusMfaProvider(providerId string) (*Provider, error) {
	provider, err := GetProvider(providerId)
	if err != nil {
		return nil, err
	}

	if provider == nil {
		return nil, fmt.Errorf("provider %s not found", providerId)
	}

	if provider.Category != "MFA" || provider.Type != "RADIUS" {
		return nil, fmt.Errorf("provider %s is not a RADIUS MFA provider", providerId)
	}

	return provider, nil
}

// Helper function to get user's RADIUS MFA configuration
func (user *User) GetRadiusMfaConfig() (*MfaProps, error) {
	if user.RadiusProvider == "" || user.RadiusSecret == "" {
		return nil, errors.New("RADIUS MFA not configured for user")
	}

	provider, err := GetProvider(user.RadiusProvider)
	if err != nil {
		return nil, err
	}

	if provider == nil {
		return nil, errors.New("RADIUS provider not found")
	}

	return &MfaProps{
		MfaType:     RadiusType,
		Enabled:     true,
		Secret:      user.RadiusProvider,
		CountryCode: user.RadiusUsername,
	}, nil
}
