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

export function getFlows(owner: string, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return get(`/api/get-flows?owner=${encodeURIComponent(owner)}&${pageQuery(page, pageSize, field, value, sortField, sortOrder)}`);
}

export function getFlow(owner: string, name: string) {
  return get(`/api/get-flow?id=${objectId(owner, name)}`);
}

export function addFlow(flow: any) {
  return post("/api/add-flow", flow);
}

export function updateFlow(owner: string, name: string, flow: any) {
  return post(`/api/update-flow?id=${objectId(owner, name)}`, flow);
}

export function deleteFlow(flow: any) {
  return post(`/api/delete-flow?id=${objectId(flow.owner, flow.name)}`);
}

export function getFlowRuns(owner: string, flow = "", page: any = 1, pageSize: any = 20, field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return get(`/api/get-flow-runs?owner=${encodeURIComponent(owner)}&flow=${encodeURIComponent(flow)}&${pageQuery(page, pageSize, field, value, sortField, sortOrder)}`);
}

export function testFlow(form: any) {
  return post("/api/test-flow", form);
}
