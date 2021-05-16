// Copyright 2021 The casbin Authors. All Rights Reserved.
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

function oAuthParamsToQuery(oAuthParams) {
  return `?clientId=${oAuthParams.clientId}&responseType=${oAuthParams.responseType}&redirectUri=${oAuthParams.redirectUri}&scope=${oAuthParams.scope}&state=${oAuthParams.state}`;
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
