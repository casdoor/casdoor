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
import i18next from "i18next";

export function getGlobalUsers(page, pageSize, field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-global-users?p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getUsers(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-users?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
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

export function getAddressOptions(url) {
  return fetch(url, {
    method: "GET",
  }).then(res => res.json());
}

export function getAffiliationOptions(url, code) {
  return fetch(`${url}/${code}`, {
    method: "GET",
  }).then(res => res.json());
}

export function setPassword(userOwner, userName, oldPassword, newPassword) {
  let formData = new FormData();
  formData.append("userOwner", userOwner);
  formData.append("userName", userName);
  formData.append("oldPassword", oldPassword);
  formData.append("newPassword", newPassword);
  return fetch(`${Setting.ServerUrl}/api/set-password`, {
    method: "POST",
    credentials: "include",
    body: formData
  }).then(res => res.json());
}

export function sendCode(checkType, checkId, checkKey, dest, type, orgId, checkUser) {
  let formData = new FormData();
  formData.append("checkType", checkType);
  formData.append("checkId", checkId);
  formData.append("checkKey", checkKey);
  formData.append("dest", dest);
  formData.append("type", type);
  formData.append("organizationId", orgId);
  formData.append("checkUser", checkUser);
  return fetch(`${Setting.ServerUrl}/api/send-verification-code`, {
    method: "POST",
    credentials: "include",
    body: formData
  }).then(res => res.json()).then(res => {
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("user:Code Sent"));
      return true;
    } else {
      Setting.showMessage("error", i18next.t("user:" + res.msg));
      return false;
    }
  });
}

export function resetEmailOrPhone(dest, type, code) {
  let formData = new FormData();
  formData.append("dest", dest);
  formData.append("type", type);
  formData.append("code", code);
  return fetch(`${Setting.ServerUrl}/api/reset-email-or-phone`, {
    method: "POST",
    credentials: "include",
    body: formData
  }).then(res => res.json());
}

export function getHumanCheck() {
  return fetch(`${Setting.ServerUrl}/api/get-human-check`, {
    method: "GET"
  }).then(res => res.json());
}

export function checkUserPassword(values) {
  return fetch(`${Setting.ServerUrl}/api/check-user-password`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(values)
  }).then(res => res.json());
}

export function initTOTP() {
  return fetch(`${Setting.ServerUrl}/api/totp`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function setTOTP(secret, code) {
  return fetch(`${Setting.ServerUrl}/api/totp?secret=${secret}&code=${code}`, {
    method: 'POST',
    credentials: 'include'
  }).then(res => res.json());
}

export function deleteTOTP(recoveryCode) {
  return fetch(`${Setting.ServerUrl}/api/delete-totp?recoveryCode=${recoveryCode}`, {
    method: 'POST',
    credentials: 'include'
  }).then(res => res.json());
}