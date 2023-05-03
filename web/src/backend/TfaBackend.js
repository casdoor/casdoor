// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

export function twoFactorSetupInitiate(values) {
  const formData = new FormData();
  formData.append("userId", values.userId);
  formData.append("type", values.type);
  return fetch(`${Setting.ServerUrl}/api/two-factor/setup/initiate`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function twoFactorSetupVerity(values) {
  const formData = new FormData();
  formData.append("type", values.type);
  formData.append("passcode", values.passcode);
  return fetch(`${Setting.ServerUrl}/api/two-factor/setup/verity`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function twoFactorSetupEnable(values) {
  const formData = new FormData();
  formData.append("type", values.type);
  formData.append("userId", values.userId);
  return fetch(`${Setting.ServerUrl}/api/two-factor/setup/enable`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function twoFactorRemoveTotp(values) {
  const formData = new FormData();
  formData.append("type", values.type);
  formData.append("userId", values.userId);
  return fetch(`${Setting.ServerUrl}/api/two-factor`, {
    method: "DELETE",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function twoFactorAuthTotp(values) {
  const formData = new FormData();
  formData.append("passcode", values.passcode);
  return fetch(`${Setting.ServerUrl}/api/two-factor/auth`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function twoFactorRecoverTotp(values) {
  const formData = new FormData();
  formData.append("recoveryCode", values.recoveryCode);
  return fetch(`${Setting.ServerUrl}/api/two-factor/auth/totp/recover`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}
