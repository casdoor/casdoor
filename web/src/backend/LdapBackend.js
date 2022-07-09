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

export function getLdaps(owner) {
  return fetch(`${Setting.ServerUrl}/api/get-ldaps?owner=${owner}`, {
    method: 'POST',
    credentials: 'include',
  }).then(res => res.json());
}

export function getLdap(id) {
  return fetch(`${Setting.ServerUrl}/api/get-ldap?id=${id}`, {
    method: 'POST',
    credentials: 'include',
  }).then(res => res.json());
}

export function addLdap(body) {
  return fetch(`${Setting.ServerUrl}/api/add-ldap`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(body),
  }).then(res => res.json());
}

export function deleteLdap(body) {
  return fetch(`${Setting.ServerUrl}/api/delete-ldap`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(body),
  }).then(res => res.json());
}

export function updateLdap(body) {
  return fetch(`${Setting.ServerUrl}/api/update-ldap`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(body),
  }).then(res => res.json());
}

export function getLdapUser(body) {
  return fetch(`${Setting.ServerUrl}/api/get-ldap-user`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(body),
  }).then(res => res.json());
}

export function syncUsers(owner, ldapId, body) {
  return fetch(`${Setting.ServerUrl}/api/sync-ldap-users?owner=${owner}&ldapId=${ldapId}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(body),
  }).then(res => res.json());
}

export function checkLdapUsersExist(owner, body) {
  return fetch(`${Setting.ServerUrl}/api/check-ldap-users-exist?owner=${owner}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(body),
  }).then(res => res.json());
}
