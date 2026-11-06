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

export function getApplicationGrants(organization: string, application: string) {
  return fetch(`${Setting.ServerUrl}/api/get-application-grants?owner=${encodeURIComponent(organization)}&application=${encodeURIComponent(application)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addApplicationGrant(grant: any) {
  return fetch(`${Setting.ServerUrl}/api/add-application-grant`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(grant),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteApplicationGrant(grant: any) {
  return fetch(`${Setting.ServerUrl}/api/delete-application-grant`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify({owner: grant.owner, name: grant.name, application: grant.application}),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getApplicationGrantPrincipals(organization: string, application: string, principalType: string, principalOrganization: string, query: string) {
  const params = new URLSearchParams({
    owner: organization,
    application,
    principalType,
    organization: principalOrganization,
    query,
  });
  return fetch(`${Setting.ServerUrl}/api/get-application-grant-principals?${params.toString()}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateApplicationGrant(grant: any) {
  return fetch(`${Setting.ServerUrl}/api/update-application-grant`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify({
      owner: grant.owner,
      name: grant.name,
      application: grant.application,
      roles: grant.roles ?? [],
      expireTime: grant.expireTime ?? "",
      description: grant.description ?? "",
    }),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateApplicationRoles(organization: string, application: string, appRoles: any[]) {
  return fetch(`${Setting.ServerUrl}/api/update-application-roles`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify({owner: organization, name: application, application, appRoles}),
    headers: {
      "Content-Type": "application/json",
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getApplicationAccessPreview(organization: string, application: string, user: string, roles: string[] = [], draft = false) {
  const params = new URLSearchParams({owner: organization, application, user, roles: roles.join(","), draft: String(draft)});
  return fetch(`${Setting.ServerUrl}/api/get-application-access-preview?${params.toString()}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
