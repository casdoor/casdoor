// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

import * as Cookie from "cookie";

export let DefaultApplication = "app-built-in";

export let ShowGithubCorner = false;
export let IsDemoMode = false;

export let ForceLanguage = "";
export let DefaultLanguage = "en";

export let StaticBaseUrl = "https://cdn.casbin.org";

export const InitThemeAlgorithm = true;
export const ThemeDefault = {
  themeType: "default",
  colorPrimary: "#5734d3",
  borderRadius: 6,
  isCompact: false,
};

export const CustomFooter = null;

// Blank or null to hide Ai Assistant button
export let AiAssistantUrl = "https://ai.casbin.com";

// Maximum number of navbar items before switching from flat to grouped menu
export let MaxItemsForFlatMenu = 7;

// setConfig updates the frontend configuration from backend
export function setConfig(config) {
  if (!config) {
    return;
  }
  if (config.showGithubCorner !== undefined) {
    ShowGithubCorner = config.showGithubCorner;
  }
  if (config.isDemoMode !== undefined) {
    IsDemoMode = config.isDemoMode;
  }
  if (config.forceLanguage !== undefined) {
    ForceLanguage = config.forceLanguage;
  }
  if (config.defaultLanguage !== undefined) {
    DefaultLanguage = config.defaultLanguage;
  }
  if (config.staticBaseUrl !== undefined) {
    StaticBaseUrl = config.staticBaseUrl;
  }
  if (config.aiAssistantUrl !== undefined) {
    AiAssistantUrl = config.aiAssistantUrl;
  }
  if (config.defaultApplication !== undefined) {
    DefaultApplication = config.defaultApplication;
  }
  if (config.maxItemsForFlatMenu !== undefined) {
    MaxItemsForFlatMenu = config.maxItemsForFlatMenu;
  }
}

export function initConfigFromCookie() {
  if (typeof document === "undefined") {
    return;
  }

  try {
    const curCookie = Cookie.parse(document.cookie);
    const raw = curCookie["jsonWebConfig"];
    if (!raw || raw === "null") {
      return;
    }

    const config = JSON.parse(raw);
    setConfig(config);
  } catch {
    // Ignore malformed cookie and keep compile-time defaults.
  }
}
