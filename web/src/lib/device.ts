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

// A device fingerprint for the new-device sign-in alerts: a hash of the browser's stable properties, so that it
// survives cleared storage. It identifies a kind of device rather than one machine, which is enough for alerts.

function hash(text: string): string {
  let h1 = 0xdeadbeef;
  let h2 = 0x41c6ce57;
  for (let i = 0; i < text.length; i++) {
    const ch = text.charCodeAt(i);
    h1 = Math.imul(h1 ^ ch, 2654435761);
    h2 = Math.imul(h2 ^ ch, 1597334677);
  }
  h1 = Math.imul(h1 ^ (h1 >>> 16), 2246822507) ^ Math.imul(h2 ^ (h2 >>> 13), 3266489909);
  h2 = Math.imul(h2 ^ (h2 >>> 16), 2246822507) ^ Math.imul(h1 ^ (h1 >>> 13), 3266489909);
  return (h2 >>> 0).toString(16).padStart(8, "0") + (h1 >>> 0).toString(16).padStart(8, "0");
}

function getCanvasSignature(): string {
  try {
    const canvas = document.createElement("canvas");
    canvas.width = 200;
    canvas.height = 40;
    const ctx = canvas.getContext("2d");
    if (!ctx) {
      return "";
    }
    ctx.textBaseline = "top";
    ctx.font = "16px Arial";
    ctx.fillStyle = "#f60";
    ctx.fillRect(100, 1, 60, 20);
    ctx.fillStyle = "#069";
    ctx.fillText("Casdoor, 登录 \u{1F512}", 2, 15);
    return canvas.toDataURL();
  } catch {
    return "";
  }
}

let deviceId: string | null = null;

export function getDeviceId(): string {
  if (deviceId !== null) {
    return deviceId;
  }

  const nav = navigator as any;
  const parts = [
    nav.userAgent,
    nav.platform,
    nav.language,
    (nav.languages ?? []).join(","),
    nav.hardwareConcurrency,
    nav.deviceMemory,
    nav.maxTouchPoints,
    `${screen.width}x${screen.height}x${screen.colorDepth}`,
    window.devicePixelRatio,
    Intl.DateTimeFormat().resolvedOptions().timeZone,
    getCanvasSignature(),
  ];
  deviceId = hash(parts.map((part) => String(part ?? "")).join("|"));
  return deviceId;
}
