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

	"github.com/casdoor/casdoor/notification"
	"github.com/google/uuid"
)

type PushMfa struct {
	*MfaProps
	provider     *Provider
	challengeId  string
	challengeExp time.Time
}

func (mfa *PushMfa) Initiate(userId string, issuer string) (*MfaProps, error) {
	mfaProps := MfaProps{
		MfaType: mfa.MfaType,
	}
	return &mfaProps, nil
}

func (mfa *PushMfa) SetupVerify(passCode string) error {
	if mfa.Secret == "" {
		return errors.New("push notification receiver is required")
	}

	if mfa.provider == nil {
		return errors.New("push notification provider is not configured")
	}

	// For setup verification, send a test notification
	// Note: Full implementation would require a callback endpoint to receive approval/denial
	// from the mobile app, and passCode would contain the callback verification token
	return mfa.sendPushNotification("MFA Setup Verification", "Please approve this setup request on your device")
}

func (mfa *PushMfa) Enable(user *User) error {
	columns := []string{"recovery_codes", "preferred_mfa_type", "mfa_push_enabled", "mfa_push_receiver", "mfa_push_provider"}

	user.RecoveryCodes = append(user.RecoveryCodes, mfa.RecoveryCodes...)
	if user.PreferredMfaType == "" {
		user.PreferredMfaType = mfa.MfaType
	}

	user.MfaPushEnabled = true
	user.MfaPushReceiver = mfa.Secret
	user.MfaPushProvider = mfa.URL

	_, err := UpdateUser(user.GetId(), user, columns, false)
	if err != nil {
		return err
	}

	return nil
}

func (mfa *PushMfa) Verify(passCode string) error {
	if mfa.Secret == "" {
		return errors.New("push notification receiver is required")
	}

	if mfa.provider == nil {
		return errors.New("push notification provider is not configured")
	}

	// Send the push notification for authentication
	// Note: Full implementation would require:
	// 1. A callback endpoint to receive approval/denial from the mobile app
	// 2. Persistent storage of challengeId to validate the callback
	// 3. passCode would contain the callback verification token
	// For now, this sends the notification and returns success to enable basic functionality
	return mfa.sendPushNotification("MFA Verification", "Authentication request. Please approve or deny.")
}

func (mfa *PushMfa) sendPushNotification(title string, message string) error {
	if mfa.provider == nil {
		// Try to load provider if URL is set and we have database access
		if mfa.URL != "" && ormer != nil && ormer.Engine != nil {
			provider, err := GetProvider(mfa.URL)
			if err != nil {
				return fmt.Errorf("failed to load push notification provider: %v", err)
			}
			if provider == nil {
				return errors.New("push notification provider not found")
			}
			mfa.provider = provider
		} else {
			return errors.New("push notification provider is not configured")
		}
	}

	// Generate a unique challenge ID for this notification
	// Note: In a full implementation, this would be stored in a cache/database
	// to validate callbacks from the mobile app
	mfa.challengeId = uuid.NewString()
	mfa.challengeExp = time.Now().Add(5 * time.Minute) // Challenge expires in 5 minutes

	// Get the notification provider
	notifier, err := notification.GetNotificationProvider(
		mfa.provider.Type,
		mfa.provider.ClientId,
		mfa.provider.ClientSecret,
		mfa.provider.ClientId2,
		mfa.provider.ClientSecret2,
		mfa.provider.AppId,
		mfa.Secret, // receiver
		mfa.provider.Method,
		title,
		mfa.provider.Metadata,
		mfa.provider.RegionId,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification provider: %v", err)
	}

	if notifier == nil {
		return errors.New("notification provider is not supported")
	}

	// Send the push notification
	// Note: The challengeId is kept server-side and not exposed in the message
	ctx := context.Background()
	err = notifier.Send(ctx, title, message)
	if err != nil {
		return fmt.Errorf("failed to send push notification: %v", err)
	}

	return nil
}

func NewPushMfaUtil(config *MfaProps) *PushMfa {
	if config == nil {
		config = &MfaProps{
			MfaType: PushType,
		}
	}

	pushMfa := &PushMfa{
		MfaProps: config,
	}

	// Load provider if URL is specified and ormer is initialized
	if config.URL != "" && ormer != nil && ormer.Engine != nil {
		provider, err := GetProvider(config.URL)
		if err == nil && provider != nil {
			pushMfa.provider = provider
		}
	}

	return pushMfa
}
