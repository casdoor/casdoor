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

export function getEntries(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-entries?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function getEntry(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-entry?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function getOpenClawSessionGraph(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-openclaw-session-graph?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function getOpenClawSessionTranscriptUrl(owner, name) {
  return `${Setting.ServerUrl}/api/get-openclaw-session-transcript?id=${owner}/${encodeURIComponent(name)}`;
}

export function getOpenClawSessionTranscript(owner, name) {
  return fetch(getOpenClawSessionTranscriptUrl(owner, name), {
    method: "GET",
    credentials: "include",
  }).then(res => res.json());
}

export function updateEntry(owner, name, entry) {
  const newEntry = Setting.deepCopy(entry);
  return fetch(`${Setting.ServerUrl}/api/update-entry?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newEntry),
  }).then(res => res.json());
}

export function addEntry(entry) {
  const newEntry = Setting.deepCopy(entry);
  return fetch(`${Setting.ServerUrl}/api/add-entry`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newEntry),
  }).then(res => res.json());
}

export function deleteEntry(entry) {
  const newEntry = Setting.deepCopy(entry);
  return fetch(`${Setting.ServerUrl}/api/delete-entry`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newEntry),
  }).then(res => res.json());
}
