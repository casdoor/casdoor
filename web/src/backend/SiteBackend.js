// Copyright 2023 The casbin Authors. All Rights Reserved.
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

export function getGlobalSites() {
  return fetch(`${Setting.ServerUrl}/api/get-global-sites`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function getSites(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-sites?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function getSite(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-site?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function updateSite(owner, name, site) {
  const newSite = Setting.deepCopy(site);
  return fetch(`${Setting.ServerUrl}/api/update-site?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newSite),
  }).then(res => res.json());
}

export function addSite(site) {
  const newSite = Setting.deepCopy(site);
  return fetch(`${Setting.ServerUrl}/api/add-site`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newSite),
  }).then(res => res.json());
}

export function deleteSite(site) {
  const newSite = Setting.deepCopy(site);
  return fetch(`${Setting.ServerUrl}/api/delete-site`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newSite),
  }).then(res => res.json());
}
