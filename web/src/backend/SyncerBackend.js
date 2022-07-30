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

export function getSyncers(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-syncers?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getSyncer(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-syncer?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function updateSyncer(owner, name, syncer) {
  let newSyncer = Setting.deepCopy(syncer);
  return fetch(`${Setting.ServerUrl}/api/update-syncer?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newSyncer),
  }).then(res => res.json());
}

export function addSyncer(syncer) {
  let newSyncer = Setting.deepCopy(syncer);
  return fetch(`${Setting.ServerUrl}/api/add-syncer`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newSyncer),
  }).then(res => res.json());
}

export function deleteSyncer(syncer) {
  let newSyncer = Setting.deepCopy(syncer);
  return fetch(`${Setting.ServerUrl}/api/delete-syncer`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newSyncer),
  }).then(res => res.json());
}

export function runSyncer(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/run-syncer?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}
