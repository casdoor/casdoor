import * as Setting from "../Setting";

export function getGlobalUsers() {
  return fetch(`${Setting.ServerUrl}/api/get-global-users`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getUsers(owner) {
  return fetch(`${Setting.ServerUrl}/api/get-users?owner=${owner}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getUser(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-user?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function updateUser(owner, name, user) {
  let newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/update-user?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newUser),
  }).then(res => res.json());
}

export function addUser(user) {
  let newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/add-user`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newUser),
  }).then(res => res.json());
}

export function deleteUser(user) {
  let newUser = Setting.deepCopy(user);
  return fetch(`${Setting.ServerUrl}/api/delete-user`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newUser),
  }).then(res => res.json());
}
