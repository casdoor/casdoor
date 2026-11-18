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

export function getSecurityScreen(owner: string) {
  return get(`/api/get-security-screen?owner=${encodeURIComponent(owner)}`);
}

/** Downloads the security report as a .docx, or returns the JSON error the server sent instead. */
export async function getSecurityReportDocx(owner: string, days: number, month = ""): Promise<{status: string; msg?: string; blob?: Blob; fileName?: string}> {
  const res = await fetch(`${Setting.ServerUrl}/api/get-security-report-docx?owner=${encodeURIComponent(owner)}&days=${days}&month=${month}`, {
    method: "GET",
    credentials: "include",
    headers: {"Accept-Language": Setting.getAcceptLanguage()},
  });
  if ((res.headers.get("Content-Type") ?? "").includes("json")) {
    return res.json();
  }
  const fileName = /filename="?([^";]+)"?/.exec(res.headers.get("Content-Disposition") ?? "")?.[1] ?? "security-report.docx";
  return {status: "ok", blob: await res.blob(), fileName};
}
