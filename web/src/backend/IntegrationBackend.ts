// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

import * as Setting from "@/lib/setting";

export function getMarketplaceIndex(owner, refresh = false) {
  return fetch(`${Setting.ServerUrl}/api/get-marketplace-index?owner=${encodeURIComponent(owner)}&refresh=${refresh ? "1" : ""}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getMarketplaceBundle(owner, integrationId, version = "") {
  return fetch(`${Setting.ServerUrl}/api/get-marketplace-bundle?owner=${encodeURIComponent(owner)}&integrationId=${encodeURIComponent(integrationId)}&version=${encodeURIComponent(version)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getIntegrations(owner, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return fetch(`${Setting.ServerUrl}/api/get-integrations?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getIntegration(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-integration?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function installIntegration(request) {
  return fetch(`${Setting.ServerUrl}/api/install-integration`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(request),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteIntegration(integration) {
  const newIntegration = Setting.deepCopy(integration);
  return fetch(`${Setting.ServerUrl}/api/delete-integration`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newIntegration),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
