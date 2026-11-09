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
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/casdoor/casdoor/action"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/notification"
	"github.com/casdoor/casdoor/util"
)

const (
	PushChallengePending  = "Pending"
	PushChallengeApproved = "Approved"
	PushChallengeDenied   = "Denied"
	PushChallengeExpired  = "Expired"

	pushChallengeLifetime = 5 * time.Minute
)

// PushChallenge is a sign-in waiting for the user to confirm it on the phone
type PushChallenge struct {
	Id          string    `json:"id"`
	UserId      string    `json:"userId"`
	Provider    string    `json:"-"`
	Receiver    string    `json:"-"`
	Application string    `json:"application"`
	Ip          string    `json:"ip"`
	Location    string    `json:"location"`
	UserAgent   string    `json:"userAgent"`
	Status      string    `json:"status"`
	CreatedTime time.Time `json:"createdTime"`
}

var pushChallenges sync.Map

func getPushChallenge(id string) *PushChallenge {
	value, ok := pushChallenges.Load(id)
	if !ok {
		return nil
	}
	challenge := value.(*PushChallenge)
	if challenge.Status == PushChallengePending && time.Since(challenge.CreatedTime) > pushChallengeLifetime {
		challenge.Status = PushChallengeExpired
	}
	if time.Since(challenge.CreatedTime) > 2*pushChallengeLifetime {
		pushChallenges.Delete(id)
		return nil
	}
	return challenge
}

// GetPushChallenge returns the challenge for the confirmation page and the sign-in page waiting for it
func GetPushChallenge(id string) *PushChallenge {
	if id == "" {
		return nil
	}
	return getPushChallenge(id)
}

// AnswerPushChallenge records the user's answer on the confirmation page
func AnswerPushChallenge(id string, approve bool) (*PushChallenge, error) {
	challenge := GetPushChallenge(id)
	if challenge == nil {
		return nil, errors.New("the sign-in request doesn't exist or has expired")
	}
	if challenge.Status != PushChallengePending {
		return challenge, fmt.Errorf("the sign-in request has been %s", challenge.Status)
	}
	if approve {
		challenge.Status = PushChallengeApproved
	} else {
		challenge.Status = PushChallengeDenied
	}
	return challenge, nil
}

// consumePushChallenge is the passcode check of the push MFA: the passcode is the id of a challenge that the
// user has approved for the same provider and receiver, and it works once
func consumePushChallenge(id string, provider string, receiver string) error {
	challenge := GetPushChallenge(id)
	if challenge == nil || challenge.Provider != provider || challenge.Receiver != receiver {
		return errors.New("the sign-in request doesn't exist or has expired, please send it again")
	}
	switch challenge.Status {
	case PushChallengeApproved:
		pushChallenges.Delete(id)
		return nil
	case PushChallengeDenied:
		pushChallenges.Delete(id)
		return errors.New("the sign-in request was denied on the phone")
	case PushChallengeExpired:
		return errors.New("the sign-in request has expired, please send it again")
	default:
		return errors.New("the sign-in request hasn't been confirmed on the phone yet")
	}
}

// SendPushChallenge sends a card with a confirmation link to the receiver through the provider: the enterprise
// app of a WeCom, DingTalk or Lark login provider, or a notification provider
func SendPushChallenge(userId string, providerId string, receiver string, application string, clientIp string, userAgent string, origin string) (*PushChallenge, error) {
	if err := edition.Require(edition.Commercial, "Push notification MFA"); err != nil {
		return nil, err
	}
	if providerId == "" || receiver == "" {
		return nil, errors.New("push notification provider and receiver are required")
	}

	provider, err := GetProvider(providerId)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("the provider: %s does not exist", providerId)
	}

	challenge := &PushChallenge{
		Id:          util.GenerateId() + util.GenerateId(),
		UserId:      userId,
		Provider:    providerId,
		Receiver:    receiver,
		Application: application,
		Ip:          clientIp,
		Location:    getIpLocation(clientIp),
		UserAgent:   userAgent,
		Status:      PushChallengePending,
		CreatedTime: time.Now(),
	}

	confirmUrl := fmt.Sprintf("%s/mfa/push-confirm/%s", origin, challenge.Id)
	title := "Confirm your sign-in"
	description := fmt.Sprintf("Account: %s\nApplication: %s\nIP address: %s %s\nTime: %s\nOpen the link and approve it if it is you. It expires in 5 minutes.",
		userId, application, clientIp, challenge.Location, challenge.CreatedTime.Format("2006-01-02 15:04:05"))

	switch {
	case provider.Category == "Notification" && provider.Type == ProviderTypePlugin:
		// an action from the Marketplace, such as an app message of WeCom, DingTalk or Lark
		var m *action.Module
		m, _, err = getInstalledAction(provider.Owner, provider.SubType, "push")
		if err == nil {
			err = action.SendPush(m, getPushActionConfig(provider), receiver, title, description, confirmUrl)
		}
	case provider.Category == "Notification":
		err = sendNotificationCard(provider, receiver, title, description+"\n"+confirmUrl)
	default:
		err = fmt.Errorf("the provider: %s can't send push notifications, use a notification provider", providerId)
	}
	if err != nil {
		return nil, err
	}

	pushChallenges.Store(challenge.Id, challenge)
	return challenge, nil
}

// GetFrontendOrigin is the origin of the pages, for the links sent out
func GetFrontendOrigin(host string) string {
	origin, _ := getOriginFromHost(host)
	return origin
}

func getPushActionConfig(provider *Provider) *action.Config {
	return &action.Config{
		ClientId:      provider.ClientId,
		ClientSecret:  provider.ClientSecret,
		ClientId2:     provider.AppId,
		ClientSecret2: provider.ClientSecret2,
		Host:          provider.Host,
		Metadata:      provider.Metadata,
	}
}

// GetPushReceiverField returns the user field that holds the receiver of a push provider of type Plugin, so that
// the receiver is filled in when the user has linked that account, e.g. "wecom"
func GetPushReceiverField(provider *Provider) string {
	if provider == nil || provider.Type != ProviderTypePlugin {
		return ""
	}
	_, file, err := getInstalledAction(provider.Owner, provider.SubType, "push")
	if err != nil {
		return ""
	}
	return file.ReceiverField
}

// GetUserFieldValue returns the value of one of the account fields that a push receiver may be taken from
func GetUserFieldValue(user *User, field string) string {
	switch field {
	case "wecom":
		return user.Wecom
	case "dingtalk":
		return user.DingTalk
	case "lark":
		return user.Lark
	case "email":
		return user.Email
	case "phone":
		return user.Phone
	}
	return ""
}

func sendNotificationCard(provider *Provider, receiver string, title string, content string) error {
	notifier, err := notification.GetNotificationProvider(provider.Type, provider.ClientId, provider.ClientSecret, provider.ClientId2, provider.ClientSecret2,
		provider.AppId, receiver, provider.Method, title, provider.Metadata, provider.RegionId)
	if err != nil {
		return err
	}
	if notifier == nil {
		return errors.New("notification provider is not supported")
	}
	restrictNotificationClient(notifier, util.NewInternetOnlyHttpClient(30*time.Second))
	return notifier.Send(context.Background(), title, content)
}
