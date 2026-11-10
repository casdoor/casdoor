// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

import * as Setting from "@/lib/setting";

function get(path: string) {
  return fetch(`${Setting.ServerUrl}${path}`, {
    method: "GET",
    credentials: "include",
    headers: {"Accept-Language": Setting.getAcceptLanguage()},
  }).then(res => res.json());
}

export function getComplianceReport(owner: string) {
  return get(`/api/get-compliance-report?owner=${encodeURIComponent(owner)}`);
}

export function getSecurityOverview(owner: string, days: number, month = "") {
  return get(`/api/get-security-overview?owner=${encodeURIComponent(owner)}&days=${days}&month=${month}`);
}

export function getOrphanAccounts(owner: string, days: number) {
  return get(`/api/get-orphan-accounts?owner=${encodeURIComponent(owner)}&days=${days}`);
}

export function getSodConflicts(owner: string) {
  return get(`/api/get-sod-conflicts?owner=${encodeURIComponent(owner)}`);
}

export function getConsentLogs(owner: string, page: number, pageSize: number) {
  return get(`/api/get-consent-logs?owner=${encodeURIComponent(owner)}&p=${page}&pageSize=${pageSize}`);
}

export function verifyRecords(owner: string) {
  return get(`/api/verify-records?owner=${encodeURIComponent(owner)}`);
}

export function importOtpDevices(owner: string, tokens: string, assignments: string) {
  return fetch(`${Setting.ServerUrl}/api/import-otp-devices?owner=${encodeURIComponent(owner)}`, {
    method: "POST",
    credentials: "include",
    body: JSON.stringify({owner, tokens, assignments}),
    headers: {"Accept-Language": Setting.getAcceptLanguage()},
  }).then(res => res.json());
}
