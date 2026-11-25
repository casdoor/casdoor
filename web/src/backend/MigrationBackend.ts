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

export function getMigrations(owner, organization, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return fetch(`${Setting.ServerUrl}/api/get-migrations?owner=${owner}&organization=${encodeURIComponent(organization)}&p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value)}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getMigration(owner, name, organization: any = owner) {
  return fetch(`${Setting.ServerUrl}/api/get-migration?id=${owner}/${encodeURIComponent(name)}&organization=${encodeURIComponent(organization)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateMigration(owner, name, migration) {
  const newMigration = Setting.deepCopy(migration);
  return fetch(`${Setting.ServerUrl}/api/update-migration?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newMigration),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addMigration(migration) {
  const newMigration = Setting.deepCopy(migration);
  return fetch(`${Setting.ServerUrl}/api/add-migration`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newMigration),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteMigration(migration) {
  const newMigration = Setting.deepCopy(migration);
  return fetch(`${Setting.ServerUrl}/api/delete-migration`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newMigration),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function runMigration(owner, name, organization, mode, files: any[] = []) {
  return fetch(`${Setting.ServerUrl}/api/run-migration?id=${owner}/${encodeURIComponent(name)}&organization=${encodeURIComponent(organization)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify({mode, files}),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
