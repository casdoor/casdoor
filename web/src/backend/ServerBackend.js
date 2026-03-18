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

import * as Setting from "../Setting";

export function getServers(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-servers?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function getServer(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-server?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function updateServer(owner, name, server) {
  const newServer = Setting.deepCopy(server);
  return fetch(`${Setting.ServerUrl}/api/update-server?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newServer),
  }).then(res => res.json());
}

export function addServer(server) {
  const newServer = Setting.deepCopy(server);
  return fetch(`${Setting.ServerUrl}/api/add-server`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newServer),
  }).then(res => res.json());
}

export function deleteServer(server) {
  const newServer = Setting.deepCopy(server);
  return fetch(`${Setting.ServerUrl}/api/delete-server`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newServer),
  }).then(res => res.json());
}
