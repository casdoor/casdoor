// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

export function getAdapters(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-adapters?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getAdapter(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-adapter?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateAdapter(owner, name, Adapter) {
  const newAdapter = Setting.deepCopy(Adapter);
  return fetch(`${Setting.ServerUrl}/api/update-adapter?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newAdapter),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addAdapter(Adapter) {
  const newAdapter = Setting.deepCopy(Adapter);
  return fetch(`${Setting.ServerUrl}/api/add-adapter`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newAdapter),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteAdapter(Adapter) {
  const newAdapter = Setting.deepCopy(Adapter);
  return fetch(`${Setting.ServerUrl}/api/delete-adapter`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newAdapter),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function UpdatePolicy(owner, name, policy) {
  return fetch(`${Setting.ServerUrl}/api/update-policy?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(policy),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function AddPolicy(owner, name, policy) {
  return fetch(`${Setting.ServerUrl}/api/add-policy?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(policy),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function RemovePolicy(owner, name, policy) {
  return fetch(`${Setting.ServerUrl}/api/remove-policy?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(policy),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function syncPolicies(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/sync-policies?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
