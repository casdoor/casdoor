// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

import * as Setting from "@/lib/setting";

export function getMcpChanges(owner: string, page = 1, pageSize = 50, filters: Record<string, string> = {}) {
  const query = Object.entries(filters).filter(([, value]) => value).map(([key, value]) => `&${key}=${encodeURIComponent(value)}`).join("");
  return fetch(`${Setting.ServerUrl}/api/get-mcp-changes?owner=${encodeURIComponent(owner)}&p=${page}&pageSize=${pageSize}${query}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

/** Rolls back a change, or with wholeCall all the changes of its tool call. */
export function rollbackMcpChange(owner: string, name: string, wholeCall = false) {
  return fetch(`${Setting.ServerUrl}/api/rollback-mcp-change?id=${encodeURIComponent(`${owner}/${name}`)}&wholeCall=${wholeCall}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
