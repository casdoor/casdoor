// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import * as Setting from "../Setting";

export function getMessages(owner, organization, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-messages?owner=${owner}&organization=${organization}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getChatMessages(chat) {
  return fetch(`${Setting.ServerUrl}/api/get-messages?chat=${chat}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getMessage(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-message?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getMessageAnswer(owner, name, onMessage, onError) {
  const eventSource = new EventSource(`${Setting.ServerUrl}/api/get-message-answer?id=${owner}/${encodeURIComponent(name)}`);

  eventSource.addEventListener("message", (e) => {
    onMessage(e.data);
  });

  eventSource.addEventListener("myerror", (e) => {
    onError(e.data);
    eventSource.close();
  });

  eventSource.addEventListener("end", (e) => {
    eventSource.close();
  });
}

export function updateMessage(owner, name, message) {
  const newMessage = Setting.deepCopy(message);
  return fetch(`${Setting.ServerUrl}/api/update-message?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newMessage),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addMessage(message) {
  const newMessage = Setting.deepCopy(message);
  return fetch(`${Setting.ServerUrl}/api/add-message`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newMessage),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteMessage(message) {
  const newMessage = Setting.deepCopy(message);
  return fetch(`${Setting.ServerUrl}/api/delete-message`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newMessage),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
