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

/** The Casdoor device agent listens on the loopback interface of the computer, only for the sites it enrolled with. */
export const DeviceAgentUrl = "http://127.0.0.1:47630";

export interface DeviceAgentAccount {
  server: string;
  organization: string;
  device: string;
  user: string;
  displayName: string;
  status: string;
  passwordless: boolean;
}

let enabled = false;
let probePromise: Promise<DeviceAgentAccount[]> | null = null;

/** Turned on by the sign-in page of an organization that uses device trust, so that other sites never call the agent. */
export function enableDeviceAgent(value: boolean) {
  enabled = value;
}

function fetchWithTimeout(url: string, init: RequestInit, timeout: number) {
  const controller = new AbortController();
  const timer = window.setTimeout(() => controller.abort(), timeout);
  return fetch(url, {...init, signal: controller.signal}).finally(() => window.clearTimeout(timer));
}

/** The accounts that the agent of this computer enrolled for this site, none when there is no agent. */
export function probeDeviceAgent(): Promise<DeviceAgentAccount[]> {
  if (!enabled) {
    return Promise.resolve([]);
  }
  if (probePromise === null) {
    probePromise = fetchWithTimeout(`${DeviceAgentUrl}/v1/probe`, {method: "GET"}, 1500)
      .then((res) => res.json())
      .then((data) => (data?.accounts ?? []) as DeviceAgentAccount[])
      .then((accounts) => accounts.filter((account) => sameOrigin(account.server, window.location.origin)))
      .catch(() => []);
  }
  return probePromise;
}

function sameOrigin(a: string, b: string) {
  try {
    return new URL(a).origin === new URL(b).origin;
  } catch {
    return false;
  }
}

/** Asks the agent to sign a fresh challenge of the server, for the header X-Casdoor-Device-Assertion. */
export async function getDeviceAssertion(organization?: string): Promise<string> {
  const accounts = await probeDeviceAgent();
  const account = accounts.find((item) => !organization || item.organization === organization) ?? accounts[0];
  if (!account) {
    return "";
  }
  try {
    const challenge = await fetch(`${Setting.ServerUrl}/api/device/challenge`, {credentials: "include"}).then((res) => res.json());
    if (challenge.status !== "ok") {
      return "";
    }
    const res = await fetchWithTimeout(`${DeviceAgentUrl}/v1/assert`, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({device: account.device, nonce: challenge.data}),
    }, 5000);
    const data = await res.json();
    return data?.assertion ?? "";
  } catch {
    return "";
  }
}
