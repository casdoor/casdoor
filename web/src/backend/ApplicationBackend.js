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

import * as Setting from '../Setting';

export function getApplications(owner, page = '', pageSize = '', field = '', value = '', sortField = '', sortOrder = '') {
  return fetch(`${Setting.ServerUrl}/api/get-applications?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function getApplicationsByOrganization(owner, organization) {
  return fetch(`${Setting.ServerUrl}/api/get-applications?owner=${owner}&organization=${organization}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function getApplication(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-application?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function getUserApplication(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-user-application?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function updateApplication(owner, name, application) {
  let newApplication = Setting.deepCopy(application);
  return fetch(`${Setting.ServerUrl}/api/update-application?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newApplication),
  }).then(res => res.json());
}

export function addApplication(application) {
  let newApplication = Setting.deepCopy(application);
  newApplication.organization = 'built-in';
  return fetch(`${Setting.ServerUrl}/api/add-application`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newApplication),
  }).then(res => res.json());
}

export function deleteApplication(application) {
  let newApplication = Setting.deepCopy(application);
  return fetch(`${Setting.ServerUrl}/api/delete-application`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newApplication),
  }).then(res => res.json());
}

export function getSamlMetadata(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/saml/metadata?application=${owner}/${encodeURIComponent(name)}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.text());
}
