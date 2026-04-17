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

export function getCoupons(owner, page = "", pageSize = "", field = "", value = "", sortField = "", sortOrder = "") {
  return fetch(`${Setting.ServerUrl}/api/get-coupons?owner=${owner}&p=${page}&pageSize=${pageSize}&field=${field}&value=${value}&sortField=${sortField}&sortOrder=${sortOrder}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function getCoupon(owner, name) {
  return fetch(`${Setting.ServerUrl}/api/get-coupon?id=${owner}/${encodeURIComponent(name)}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function updateCoupon(owner, name, coupon) {
  const newCoupon = Setting.deepCopy(coupon);
  return fetch(`${Setting.ServerUrl}/api/update-coupon?id=${owner}/${encodeURIComponent(name)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newCoupon),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function addCoupon(coupon) {
  const newCoupon = Setting.deepCopy(coupon);
  return fetch(`${Setting.ServerUrl}/api/add-coupon`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newCoupon),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function deleteCoupon(coupon) {
  const newCoupon = Setting.deepCopy(coupon);
  return fetch(`${Setting.ServerUrl}/api/delete-coupon`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(newCoupon),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function validateCoupon(owner, couponCode, products, amount, currency) {
  return fetch(`${Setting.ServerUrl}/api/validate-coupon`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify({owner, couponCode, products, amount, currency}),
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
