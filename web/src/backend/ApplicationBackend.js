import * as Setting from "../Setting";

export function getApplications(owner) {
  return fetch(`${Setting.ServerUrl}/api/get-applications?owner=${owner}`, {
    method: "GET",
    credentials: "include"
  }).then(res => res.json());
}

export function getApplication(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-application?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include"
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
