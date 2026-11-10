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

import * as Setting from "@/lib/setting";

export function getCerts(owner, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return fetch(`${Setting.ServerUrl}/api/get-certs?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value)}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getGlobalCerts(page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return fetch(`${Setting.ServerUrl}/api/get-global-certs?&p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value)}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getCert(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-cert?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateCert(owner, name, cert) {
  const newCert = Setting.deepCopy(cert);
  return fetch(`${Setting.ServerUrl}/api/update-cert?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newCert),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addCert(cert) {
  const newCert = Setting.deepCopy(cert);
  return fetch(`${Setting.ServerUrl}/api/add-cert`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newCert),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteCert(cert) {
  const newCert = Setting.deepCopy(cert);
  return fetch(`${Setting.ServerUrl}/api/delete-cert`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newCert),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function refreshDomainExpire(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/update-cert-domain-expire?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
  }).then(res => res.json());
}

export function rotateCert(owner: string, name: string) {
  return fetch(`${Setting.ServerUrl}/api/rotate-cert?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getHsmStatus(connect: boolean) {
  return fetch(`${Setting.ServerUrl}/api/get-hsm-status?connect=${connect}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function generateHsmCert(cert) {
  return fetch(`${Setting.ServerUrl}/api/generate-hsm-cert`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(cert),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function testHsmCert(owner: string, name: string) {
  return fetch(`${Setting.ServerUrl}/api/test-hsm-cert?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
