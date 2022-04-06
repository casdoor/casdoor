// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

import {authConfig} from "./Auth";

export function getAccount(query) {
  return fetch(`${authConfig.serverUrl}/api/get-account${query}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function signup(values) {
  return fetch(`${authConfig.serverUrl}/api/signup`, {
    method: 'POST',
    credentials: "include",
    body: JSON.stringify(values),
  }).then(res => res.json());
}

export function getEmailAndPhone(values) {
  return fetch(`${authConfig.serverUrl}/api/get-email-and-phone`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(values),
  }).then((res) => res.json());
}

function oAuthParamsToQuery(oAuthParams) {
  // login
  if (oAuthParams === null) {
    return "";
  }

  // code
  return `?clientId=${oAuthParams.clientId}&responseType=${oAuthParams.responseType}&redirectUri=${oAuthParams.redirectUri}&scope=${oAuthParams.scope}&state=${oAuthParams.state}&nonce=${oAuthParams.nonce}&code_challenge_method=${oAuthParams.challengeMethod}&code_challenge=${oAuthParams.codeChallenge}`;
}

export function getApplicationLogin(oAuthParams) {
  return fetch(`${authConfig.serverUrl}/api/get-app-login${oAuthParamsToQuery(oAuthParams)}`, {
    method: 'GET',
    credentials: 'include',
  }).then(res => res.json());
}

export function login(values, oAuthParams) {
  return fetch(`${authConfig.serverUrl}/api/login${oAuthParamsToQuery(oAuthParams)}`, {
    method: 'POST',
    credentials: "include",
    body: JSON.stringify(values),
  }).then(res => res.json());
}

export function loginCas(values, params) {
  return fetch(`${authConfig.serverUrl}/api/login?service=${params.service}`, {
    method: 'POST',
    credentials: "include",
    body: JSON.stringify(values),
  }).then(res => res.json());
}

export function logout() {
  return fetch(`${authConfig.serverUrl}/api/logout`, {
    method: 'POST',
    credentials: "include",
  }).then(res => res.json());
}

export function unlink(values) {
  return fetch(`${authConfig.serverUrl}/api/unlink`, {
    method: 'POST',
    credentials: "include",
    body: JSON.stringify(values),
  }).then(res => res.json());
}

export function getSamlLogin(providerId, relayState) {
  return fetch(`${authConfig.serverUrl}/api/get-saml-login?id=${providerId}&relayState=${relayState}`, {
    method: 'GET',
    credentials: 'include',
  }).then(res => res.json());
}

export function loginWithSaml(values, param) {
  return fetch(`${authConfig.serverUrl}/api/login${param}`, {
    method: 'POST',
    credentials: "include",
    body: JSON.stringify(values),
  }).then(res => res.json());
}