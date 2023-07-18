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

import * as Setting from "../Setting";
import i18next from "i18next";

export function getGlobalUsers(page, pageSize, field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-global-users?p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getUsers(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "", groupName = "") {
  return fetch(`${Setting.ServerUrl}/api/get-users?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}&groupName=${groupName}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getUser(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-user?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addUserKeys(user) {
  return fetch(`${Setting.ServerUrl}/api/add-user-keys`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(user),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateUser(owner, name, user) {
  const newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/update-user?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newUser),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addUser(user) {
  const newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/add-user`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newUser),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteUser(user) {
  const newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/delete-user`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newUser),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
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

export function setPassword(userOwner, userName, oldPassword, newPassword, code = "") {
  const formData = new FormData();
  if (userOwner) {
    formData.append("userOwner", userOwner);
  }
  if (userName) {
    formData.append("userName", userName);
  }
  formData.append("oldPassword", oldPassword);
  formData.append("newPassword", newPassword);
  if (code) {
    formData.append("code", code);
  }
  return fetch(`${Setting.ServerUrl}/api/set-password`, {
    method: "POST",
    credentials: "include",
    body: formData,
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function sendCode(captchaType, captchaToken, clientSecret, method, countryCode = "", dest, type, applicationId, checkUser = "") {
  const formData = new FormData();
  formData.append("captchaType", captchaType);
  formData.append("captchaToken", captchaToken);
  formData.append("clientSecret", clientSecret);
  formData.append("method", method);
  formData.append("countryCode", countryCode);
  formData.append("dest", dest);
  formData.append("type", type);
  formData.append("applicationId", applicationId);
  formData.append("checkUser", checkUser);
  return fetch(`${Setting.ServerUrl}/api/send-verification-code`, {
    method: "POST",
    credentials: "include",
    body: formData,
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json()).then(res => {
    if (res.status === "ok") {
      Setting.showMessage("success", i18next.t("user:Verification code sent"));
      return true;
    } else {
      Setting.showMessage("error", i18next.t("user:" + res.msg));
      return false;
    }
  });
}

export function verifyCaptcha(captchaType, captchaToken, clientSecret) {
  const formData = new FormData();
  formData.append("captchaType", captchaType);
  formData.append("captchaToken", captchaToken);
  formData.append("clientSecret", clientSecret);
  return fetch(`${Setting.ServerUrl}/api/verify-captcha`, {
    method: "POST",
    credentials: "include",
    body: formData,
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json()).then(res => {
    if (res.status === "ok") {
      if (res.data) {
        Setting.showMessage("success", i18next.t("user:Captcha Verify Success"));
      } else {
        Setting.showMessage("error", i18next.t("user:Captcha Verify Failed"));
      }
      return true;
    } else {
      Setting.showMessage("error", i18next.t("user:" + res.msg));
      return false;
    }
  });
}

export function resetEmailOrPhone(dest, type, code) {
  const formData = new FormData();
  formData.append("dest", dest);
  formData.append("type", type);
  formData.append("code", code);
  return fetch(`${Setting.ServerUrl}/api/reset-email-or-phone`, {
    method: "POST",
    credentials: "include",
    body: formData,
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getCaptcha(owner, name, isCurrentProvider) {
  return fetch(`${Setting.ServerUrl}/api/get-captcha?applicationId=${owner}/${encodeURIComponent(name)}&isCurrentProvider=${isCurrentProvider}`, {
    method: "GET",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json()).then(res => res.data);
}

export function verifyCode(values) {
  return fetch(`${Setting.ServerUrl}/api/verify-code`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(values),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function checkUserPassword(values) {
  return fetch(`${Setting.ServerUrl}/api/check-user-password`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(values),
  }).then(res => res.json());
}

export function removeUserFromGroup({owner, name, groupName}) {
  const formData = new FormData();
  formData.append("owner", owner);
  formData.append("name", name);
  formData.append("groupName", groupName);
  return fetch(`${Setting.ServerUrl}/api/remove-user-from-group`, {
    method: "POST",
    credentials: "include",
    body: formData,
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
