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
	"fmt"

	"github.com/casdoor/casdoor/util"
)

// getActorClaim says who the application acts as: its agent when it is one, otherwise the application itself
func getActorClaim(application *Application) (*ActClaim, *TokenError, error) {
	agent, err := GetAgentByApplication(application.Name)
	if err != nil {
		return nil, nil, err
	}
	actor := &ActClaim{Sub: GetAppUserId(application), ClientId: application.ClientId}
	if agent != nil {
		if agent.IsForbidden {
			return nil, &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("the agent: %s is disabled", agent.GetId())}, nil
		}
		actor.Agent = agent.GetId()
	}
	return actor, nil, nil
}

// getTokenExchangeActor finds who acts for the user in an RFC 8693 token exchange: the client of the actor token when
// one is given, otherwise the requesting client when it is an agent. Whoever acted in the subject token is nested in
// the new actor, so the token tells the whole chain of delegation.
func getTokenExchangeActor(application *Application, user *User, subjectToken string, actorToken string, actorTokenType string) (*ActClaim, *TokenError, error) {
	subjectRecord, err := GetTokenByAccessToken(subjectToken)
	if err != nil {
		return nil, nil, err
	}
	if subjectRecord == nil {
		subjectRecord, err = GetTokenByIdToken(subjectToken)
		if err != nil {
			return nil, nil, err
		}
	}
	subjectActor := subjectRecord.GetActor()

	var actor *ActClaim
	var tokenError *TokenError
	if actorToken != "" {
		if actorTokenType != "" && actorTokenType != "urn:ietf:params:oauth:token-type:access_token" && actorTokenType != "urn:ietf:params:oauth:token-type:jwt" {
			return nil, &TokenError{Error: InvalidRequest, ErrorDescription: fmt.Sprintf("unsupported actor_token_type: %s", actorTokenType)}, nil
		}

		actorRecord, err := GetTokenByAccessToken(actorToken)
		if err != nil {
			return nil, nil, err
		}
		if actorRecord == nil || actorRecord.ExpiresIn <= 0 || actorRecord.TokenType == "ID-JAG" {
			return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "actor_token is revoked or unknown"}, nil
		}
		if isExpired, _ := util.IsTokenExpired(actorRecord.CreatedTime, actorRecord.ExpiresIn); isExpired {
			return nil, &TokenError{Error: InvalidGrant, ErrorDescription: "actor_token has expired"}, nil
		}

		actorApplication, err := getApplication(actorRecord.Owner, actorRecord.Application)
		if err != nil {
			return nil, nil, err
		}
		if actorApplication == nil {
			return nil, &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("the application of actor_token does not exist: %s", actorRecord.Application)}, nil
		}
		if actorApplication.Organization != user.Owner && actorApplication.Organization != application.Organization {
			return nil, &TokenError{Error: InvalidGrant, ErrorDescription: fmt.Sprintf("the application of actor_token: %s does not serve the organization of the user: %s", actorApplication.Name, user.GetId())}, nil
		}

		actor, tokenError, err = getActorClaim(actorApplication)
		if err != nil || tokenError != nil {
			return nil, tokenError, err
		}
	} else {
		agent, err := GetAgentByApplication(application.Name)
		if err != nil {
			return nil, nil, err
		}
		if agent == nil {
			return subjectActor, nil, nil
		}

		actor, tokenError, err = getActorClaim(application)
		if err != nil || tokenError != nil {
			return nil, tokenError, err
		}
	}

	if subjectActor != nil && subjectActor.Sub != actor.Sub {
		actor.Act = subjectActor
	} else if subjectActor != nil {
		actor.Act = subjectActor.Act
	}
	return actor, nil, nil
}
