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

export function getInvitations(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-invitations?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getInvitation(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-invitation?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getInvitationCodeInfo(code, applicationName) {
  return fetch(`${Setting.ServerUrl}/api/get-invitation-info?code=${code}&applicationId=${encodeURIComponent(applicationName)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateInvitation(owner, name, invitation) {
  const newInvitation = Setting.deepCopy(invitation);
  return fetch(`${Setting.ServerUrl}/api/update-invitation?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newInvitation),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addInvitation(invitation) {
  const newInvitation = Setting.deepCopy(invitation);
  return fetch(`${Setting.ServerUrl}/api/add-invitation`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newInvitation),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteInvitation(invitation) {
  const newInvitation = Setting.deepCopy(invitation);
  return fetch(`${Setting.ServerUrl}/api/delete-invitation`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newInvitation),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function verifyInvitation(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/verify-invitation?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
