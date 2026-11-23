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


const pageQuery = (page: any, pageSize: any, field: any, value: any, sortField: any, sortOrder: any) =>
  `p=${page}&pageSize=${pageSize}&field=${field}&value=${encodeURIComponent(value ?? "")}&sortField=${sortField}&sortOrder=${sortOrder}`;

export function getSigninRiskEvents(owner: string, page: any = "", pageSize: any = "", field: any = "", value: any = "", sortField: any = "", sortOrder: any = "") {
  return get(`/api/get-signin-risk-events?owner=${encodeURIComponent(owner)}&${pageQuery(page, pageSize, field, value, sortField, sortOrder)}`);
}

export function getSigninRiskSummary(owner: string, days = 7) {
  return get(`/api/get-signin-risk-summary?owner=${encodeURIComponent(owner)}&days=${days}`);
}
