// Copyright 2020 The casbin Authors. All Rights Reserved.
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

export function getUser(username) {
  return fetch(`${Setting.ServerUrl}/api/get-user?username=${username}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function getAccount() {
  return fetch(`${Setting.ServerUrl}/api/get-account`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function register(values) {
  return fetch(`${Setting.ServerUrl}/api/register`, {
    method: 'POST',
    credentials: "include",
    body: JSON.stringify(values),
  }).then(res => res.json());
}

export function login(values) {
  return fetch(`${Setting.ServerUrl}/api/login`, {
    method: 'POST',
    credentials: "include",
    body: JSON.stringify(values),
  }).then(res => res.json());
}

export function logout() {
  return fetch(`${Setting.ServerUrl}/api/logout`, {
    method: 'POST',
    credentials: "include",
  }).then(res => res.json());
}

export function githubLogin(providerName, code, state, redirectUrl, addition) {
  console.log(redirectUrl)
  return fetch(`${Setting.ServerUrl}/api/auth/github?provider=${providerName}&code=${code}&state=${state}&redirect_url=${redirectUrl}&addition=${addition}`, {
    method: 'GET',
    credentials: 'include',
  }).then(res => res.json());
}
