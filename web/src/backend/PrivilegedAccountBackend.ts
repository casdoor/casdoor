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
    headers: {"Content-Type": "application/json", "Accept-Language": Setting.getAcceptLanguage()},
    body: JSON.stringify(body),
  }).then(res => res.json());
}

export function getPrivilegedAccounts(owner: string, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return get(`/api/get-privileged-accounts?owner=${encodeURIComponent(owner)}&p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value)}&sortField=${sortField}&sortOrder=${sortOrder}`);
}

export function getPrivilegedAccount(owner: string, name: string) {
  return get(`/api/get-privileged-account?id=${owner}/${encodeURIComponent(name)}`);
}

export function addPrivilegedAccount(account: any) {
  return post("/api/add-privileged-account", account);
}

export function updatePrivilegedAccount(owner: string, name: string, account: any) {
  return post(`/api/update-privileged-account?id=${owner}/${encodeURIComponent(name)}`, account);
}

export function deletePrivilegedAccount(account: any) {
  return post("/api/delete-privileged-account", account);
}

export function rotatePrivilegedAccount(owner: string, name: string) {
  return post(`/api/rotate-privileged-account?id=${owner}/${encodeURIComponent(name)}`);
}

export function getPrivilegedRequests(owner: string, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return get(`/api/get-privileged-requests?owner=${encodeURIComponent(owner)}&p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value)}&sortField=${sortField}&sortOrder=${sortOrder}`);
}

export function getPrivilegedAccess() {
  return get("/api/get-privileged-access");
}

export function requestPrivilegedAccess(account: string, reason: string, durationMinutes: number) {
  return post("/api/request-privileged-access", {account, reason, durationMinutes});
}

function requestAction(action: string, request: any, body: any = {}) {
  return post(`/api/${action}?id=${request.owner}/${encodeURIComponent(request.name)}`, body);
}

export function approvePrivilegedRequest(request: any, comment: string) {
  return requestAction("approve-privileged-request", request, {comment});
}

export function rejectPrivilegedRequest(request: any, comment: string) {
  return requestAction("reject-privileged-request", request, {comment});
}

export function cancelPrivilegedRequest(request: any) {
  return requestAction("cancel-privileged-request", request);
}

export function returnPrivilegedRequest(request: any) {
  return requestAction("return-privileged-request", request);
}

export function revealPrivilegedPassword(request: any) {
  return requestAction("reveal-privileged-password", request);
}
