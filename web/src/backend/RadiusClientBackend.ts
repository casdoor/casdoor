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

export function getRadiusClients(page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return fetch(`${Setting.ServerUrl}/api/get-radius-clients?owner=admin&p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value)}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function getRadiusClient(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-radius-client?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function addRadiusClient(radiusClient) {
  return fetch(`${Setting.ServerUrl}/api/add-radius-client`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(radiusClient),
  }).then(res => res.json());
}

export function updateRadiusClient(owner, name, radiusClient) {
  return fetch(`${Setting.ServerUrl}/api/update-radius-client?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(radiusClient),
  }).then(res => res.json());
}

export function deleteRadiusClient(radiusClient) {
  return fetch(`${Setting.ServerUrl}/api/delete-radius-client`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(radiusClient),
  }).then(res => res.json());
}
