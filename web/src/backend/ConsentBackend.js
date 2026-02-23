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

import * as Setting from "../Setting";

export function grantConsent(consent, oAuthParams) {
  const request = {
    ...consent,
    clientId: oAuthParams.clientId,
    provider: "",
    signinMethod: "",
    responseType: oAuthParams.responseType || "code",
    redirectUri: oAuthParams.redirectUri,
    scope: oAuthParams.scope,
    state: oAuthParams.state,
    nonce: oAuthParams.nonce || "",
    challenge: oAuthParams.codeChallenge || "",
    resource: "",
  };
  return fetch(`${Setting.ServerUrl}/api/grant-consent`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(request),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function revokeConsent(consent) {
  return fetch(`${Setting.ServerUrl}/api/revoke-consent`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(consent),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
