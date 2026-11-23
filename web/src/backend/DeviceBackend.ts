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

function get(path: string) {
  return fetch(`${Setting.ServerUrl}${path}`, {
    method: "GET",
    credentials: "include",
    headers: {"Accept-Language": Setting.getAcceptLanguage()},
  }).then(res => res.json());
}

function post(path: string, body: any = {}) {
  return fetch(`${Setting.ServerUrl}${path}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(body),
    headers: {"Content-Type": "application/json", "Accept-Language": Setting.getAcceptLanguage()},
  }).then(res => res.json());
}

const pageQuery = (page: any, pageSize: any, field: any, value: any, sortField: any, sortOrder: any) =>
  `p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value ?? "")}&sortField=${sortField}&sortOrder=${sortOrder}`;
const objectId = (owner: string, name: string) => `${encodeURIComponent(owner)}/${encodeURIComponent(name)}`;

export function getDevices(owner: string, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return get(`/api/get-devices?owner=${encodeURIComponent(owner)}&${pageQuery(page, pageSize, field, value, sortField, sortOrder)}`);
}

export function getMyDevices() {
  return get("/api/get-my-devices");
}

export function updateDevice(device: any) {
  return post(`/api/update-device?id=${objectId(device.owner, device.name)}`, device);
}

export function deleteDevice(device: any) {
  return post(`/api/delete-device?id=${objectId(device.owner, device.name)}`);
}

export function getDeviceEnrollment(code: string) {
  return get(`/api/get-device-enrollment?code=${encodeURIComponent(code)}`);
}

export function approveDeviceEnrollment(code: string) {
  return post(`/api/approve-device-enrollment?code=${encodeURIComponent(code)}`);
}
