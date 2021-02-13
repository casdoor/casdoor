import * as Setting from "../Setting";

export function getProviders(owner) {
  return fetch(`${Setting.ServerUrl}/api/get-providers?owner=${owner}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getProvider(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-provider?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function updateProvider(owner, name, provider) {
  let newProvider = Setting.deepCopy(provider);
  return fetch(`${Setting.ServerUrl}/api/update-provider?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newProvider),
  }).then(res => res.json());
}

export function addProvider(provider) {
  let newProvider = Setting.deepCopy(provider);
  return fetch(`${Setting.ServerUrl}/api/add-provider`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newProvider),
  }).then(res => res.json());
}

export function deleteProvider(provider) {
  let newProvider = Setting.deepCopy(provider);
  return fetch(`${Setting.ServerUrl}/api/delete-provider`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newProvider),
  }).then(res => res.json());
}
