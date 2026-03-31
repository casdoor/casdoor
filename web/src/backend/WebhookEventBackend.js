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

import * as Setting from "../Setting";

export function getWebhookEvents(owner = "", organization = "", page = "", pageSize = "", webhookName = "", status = "", sortField = "", sortOrder = "") {
  const params = new URLSearchParams({
    owner,
    organization,
    pageSize,
    p: page,
    webhookName,
    status,
    sortField,
    sortOrder,
  });

  return fetch(`${Setting.ServerUrl}/api/get-webhook-events?${params.toString()}`, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}

export function replayWebhookEvent(eventId) {
  return fetch(`${Setting.ServerUrl}/api/replay-webhook-event?id=${encodeURIComponent(eventId)}`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(res => res.json());
}
