// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

export function getOrders(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-orders?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function payOrder(owner, name, providerName, paymentEnv = "") {
  return fetch(`${Setting.ServerUrl}/api/pay-order?id=${owner}/${encodeURIComponent(name)}&providerName=${encodeURIComponent(providerName)}&paymentEnv=${paymentEnv}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getUserOrders(owner, user) {
  return fetch(`${Setting.ServerUrl}/api/get-user-orders?owner=${owner}&user=${user}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getOrder(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-order?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateOrder(owner, name, order) {
  const newOrder = Setting.deepCopy(order);
  return fetch(`${Setting.ServerUrl}/api/update-order?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newOrder),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addOrder(order) {
  const newOrder = Setting.deepCopy(order);
  return fetch(`${Setting.ServerUrl}/api/add-order`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newOrder),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteOrder(order) {
  const newOrder = Setting.deepCopy(order);
  return fetch(`${Setting.ServerUrl}/api/delete-order`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newOrder),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function placeOrder(owner, productInfos, pricingName = "", planName = "", userName = "") {
  return fetch(`${Setting.ServerUrl}/api/place-order?owner=${encodeURIComponent(owner)}&pricingName=${encodeURIComponent(pricingName)}&planName=${encodeURIComponent(planName)}&userName=${encodeURIComponent(userName)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify({productInfos}),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function cancelOrder(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/cancel-order?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
