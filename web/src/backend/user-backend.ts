// Copyright 2020 The casbin Authors. All Rights Reserved.
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

import { httpClient } from './backend';

export interface User {
  email: string;
  name: string;
  owner: string;
  passwordType: string;
  password: string;
  displayName: string;
  phone: string;
}

export function getUsers(owner: string) {
  return httpClient.get(`/get-users?owner=${owner}`);
}

export function getUser(owner: string, name: string) {
  return httpClient.get(`/get-user?id=${owner}/${name}`);
}

export function updateUser(owner: string, name: string, user: any) {
  return httpClient.post(`/update-user?id=${owner}/${name}`, user);
}

export function addUser(user: any) {
  return httpClient.post(`/add-user`, user);
}

export function deleteUser(user: any) {
  return httpClient.post(`/delete-user`, user);
}
