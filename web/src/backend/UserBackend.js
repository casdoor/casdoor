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
import * as AuthBackend from "../auth/AuthBackend";

export function getGlobalUsers() {
  return fetch(`${Setting.ServerUrl}/api/get-global-users`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getUsers(owner) {
  return fetch(`${Setting.ServerUrl}/api/get-users?owner=${owner}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getUser(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-user?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function updateUser(owner, name, user) {
  let newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/update-user?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newUser),
  }).then(res => res.json());
}

export function addUser(user) {
  let newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/add-user`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newUser),
  }).then(res => res.json());
}

export function deleteUser(user) {
  let newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/delete-user`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newUser),
  }).then(res => res.json());
}

export function uploadAvatar(avatar) {
  let account;
  AuthBackend.getAccount(null).then((res) => {
    account = res.data;
    let formData = new FormData();
    formData.append("avatarfile", avatar);
    formData.append("password", account.password);
    fetch(`${Setting.ServerUrl}/api/upload-avatar`, {
      body: formData,
      method: 'POST',
      credentials: 'include',
    }).then((res) => {
      window.location.href = "/account";
    });
  });
}
