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

import * as Setting from "../Setting";

export function getOAuthApps(userId) {
  return fetch(`${Setting.ServerUrl}/api/oauth2/get-oauth-apps?userId=${userId}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getOAuthApp(userId, name) {
  return fetch(`${Setting.ServerUrl}/api/oauth2/get-oauth-app?userId=${userId}&&name=${name}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function registerOAuthApp(app) {
  let newApp = Setting.deepCopy(app);
  return fetch(`${Setting.ServerUrl}/api/oauth2/register-oauth-app`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newApp),
  }).then(res => res.json());
}

export function updateOAuthApp(clientId, app) {
  let newApp = Setting.deepCopy(app);
  return fetch(`${Setting.ServerUrl}/api/oauth2/update-oauth-app?clientId=${clientId}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newApp),
  }).then(res => res.json());
}

export function deleteOAuthApp(app) {
  let newApp = Setting.deepCopy(app);
  return fetch(`${Setting.ServerUrl}/api/oauth2/delete-oauth-app`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newApp),
  }).then(res => res.json());
}