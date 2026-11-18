// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

import * as Setting from "@/lib/setting";

function get(path: string) {
  return fetch(`${Setting.ServerUrl}${path}`, {
    method: "GET",
    credentials: "include",
    headers: {"Accept-Language": Setting.getAcceptLanguage()},
  }).then(res => res.json());
}

function post(path: string, body: any = {}) {
  return fetch(`${Setting.ServerUrl}${path}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify(body),
    headers: {"Content-Type": "application/json", "Accept-Language": Setting.getAcceptLanguage()},
  }).then(res => res.json());
}

const reviewId = (owner: string, name: string) => `${encodeURIComponent(owner)}/${encodeURIComponent(name)}`;

export function getAccessReviews(owner: string, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return get(`/api/get-access-reviews?owner=${encodeURIComponent(owner)}&p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value)}&sortField=${sortField}&sortOrder=${sortOrder}`);
}

export function getAccessReview(owner: string, name: string) {
  return get(`/api/get-access-review?id=${reviewId(owner, name)}`);
}

export function addAccessReview(review: any) {
  return post("/api/add-access-review", review);
}

export function updateAccessReview(owner: string, name: string, review: any) {
  return post(`/api/update-access-review?id=${reviewId(owner, name)}`, review);
}

export function deleteAccessReview(review: any) {
  return post("/api/delete-access-review", {owner: review.owner, name: review.name});
}

export function startAccessReview(owner: string, name: string) {
  return post(`/api/start-access-review?id=${reviewId(owner, name)}`);
}

export function closeAccessReview(owner: string, name: string) {
  return post(`/api/close-access-review?id=${reviewId(owner, name)}`);
}

export function applyAccessReview(owner: string, name: string) {
  return post(`/api/apply-access-review?id=${reviewId(owner, name)}`);
}

export function remindAccessReview(owner: string, name: string) {
  return post(`/api/remind-access-review?id=${reviewId(owner, name)}`);
}

export function getAccessReviewItems(owner: string, review: string, round: number | string = "") {
  return get(`/api/get-access-review-items?owner=${encodeURIComponent(owner)}&review=${encodeURIComponent(review)}&round=${round}`);
}

export function getExportUrl(owner: string, name: string, round: number | string = "") {
  return `${Setting.ServerUrl}/api/export-access-review?id=${reviewId(owner, name)}&round=${round}`;
}

export function getMyAccessReviewItems(pendingOnly: boolean) {
  return get(`/api/get-my-access-review-items?pending=${pendingOnly ? "1" : ""}`);
}

export function decideAccessReviewItems(items: string[], decision: string, comment: string) {
  return post("/api/decide-access-review-items", {items, decision, comment});
}
