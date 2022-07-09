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

export function getRoles(owner, page = '', pageSize = '', field = '', value = '', sortField = '', sortOrder = '') {
  return fetch(`${Setting.ServerUrl}/api/get-roles?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function getRole(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-role?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'GET',
    credentials: 'include'
  }).then(res => res.json());
}

export function updateRole(owner, name, role) {
  let newRole = Setting.deepCopy(role);
  return fetch(`${Setting.ServerUrl}/api/update-role?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newRole),
  }).then(res => res.json());
}

export function addRole(role) {
  let newRole = Setting.deepCopy(role);
  return fetch(`${Setting.ServerUrl}/api/add-role`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newRole),
  }).then(res => res.json());
}

export function deleteRole(role) {
  let newRole = Setting.deepCopy(role);
  return fetch(`${Setting.ServerUrl}/api/delete-role`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newRole),
  }).then(res => res.json());
}
