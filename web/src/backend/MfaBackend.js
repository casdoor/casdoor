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

export function MfaSetupInitiate(values) {
  const formData = new FormData();
  formData.append("owner", values.owner);
  formData.append("name", values.name);
  formData.append("mfaType", values.mfaType);
  return fetch(`${Setting.ServerUrl}/api/mfa/setup/initiate`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function MfaSetupVerify(values) {
  const formData = new FormData();
  formData.append("owner", values.owner);
  formData.append("name", values.name);
  formData.append("mfaType", values.mfaType);
  formData.append("passcode", values.passcode);
  return fetch(`${Setting.ServerUrl}/api/mfa/setup/verify`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function MfaSetupEnable(values) {
  const formData = new FormData();
  formData.append("mfaType", values.mfaType);
  formData.append("owner", values.owner);
  formData.append("name", values.name);
  return fetch(`${Setting.ServerUrl}/api/mfa/setup/enable`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function DeleteMfa(values) {
  const formData = new FormData();
  formData.append("owner", values.owner);
  formData.append("name", values.name);
  return fetch(`${Setting.ServerUrl}/api/delete-mfa`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then(res => res.json());
}

export function SetPreferredMfa(values) {
  const formData = new FormData();
  formData.append("mfaType", values.mfaType);
  formData.append("owner", values.owner);
  formData.append("name", values.name);
  return fetch(`${Setting.ServerUrl}/api/set-preferred-mfa`, {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then((res) => res.json());
}
