import * as Setting from "../Setting";

export function getOrganizations(owner) {
  return fetch(`${Setting.ServerUrl}/api/get-organizations?owner=${owner}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getOrganization(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-organization?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function updateOrganization(owner, name, organization) {
  let newOrganization = Setting.deepCopy(organization);
  return fetch(`${Setting.ServerUrl}/api/update-organization?id=${owner}/${encodeURIComponent(name)}`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newOrganization),
  }).then(res => res.json());
}

export function addOrganization(organization) {
  let newOrganization = Setting.deepCopy(organization);
  return fetch(`${Setting.ServerUrl}/api/add-organization`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newOrganization),
  }).then(res => res.json());
}

export function deleteOrganization(organization) {
  let newOrganization = Setting.deepCopy(organization);
  return fetch(`${Setting.ServerUrl}/api/delete-organization`, {
    method: 'POST',
    credentials: 'include',
    body: JSON.stringify(newOrganization),
  }).then(res => res.json());
}
