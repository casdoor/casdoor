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

function fetchDashboardApi(path, owner) {
  const apiPath = `/api/${path}`;
  const url = `${Setting.ServerUrl}${apiPath}?owner=${owner}`;
  return fetch(url, {
    method: "GET",
    credentials: "include",
    headers: {
      "Accept-Language": Setting.getAcceptLanguage(),
    },
  }).then(async res => {
    if (!res.ok) {
      const statusText = `${res.status} ${res.statusText}`;
      let bodyMsg = "";
      try {
        const contentType = res.headers.get("content-type") || "";
        if (contentType.includes("application/json")) {
          const json = await res.json();
          bodyMsg = json.msg || json.message || "";
        }
      } catch (e) {
        bodyMsg = "";
      }
      const detail = bodyMsg ? `${statusText} - ${bodyMsg}` : statusText;
      return {status: "error", msg: `${apiPath}: ${detail}`};
    }
    return res.json();
  });
}

export function getDashboard(owner) {
  return fetchDashboardApi("get-dashboard", owner);
}

export function getDashboardProviders(owner) {
  return fetchDashboardApi("get-dashboard-providers", owner);
}

export function getDashboardMfa(owner) {
  return fetchDashboardApi("get-dashboard-mfa", owner);
}

export function getDashboardHeatmap(owner) {
  return fetchDashboardApi("get-dashboard-heatmap", owner);
}
