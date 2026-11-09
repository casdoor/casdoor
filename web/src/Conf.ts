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

export const ThemeDefault = {
  themeType: "default",
  colorPrimary: "#262626",
  borderRadius: 10,
  isCompact: false,
};

// the commercial and enterprise editions each ship a default theme of their own,
// which an organization or application theme still overrides
const EditionThemeDefaults: Record<string, typeof ThemeDefault> = {
  Commercial: {
    themeType: "default",
    colorPrimary: "#4338ca",
    borderRadius: 12,
    isCompact: false,
  },
  Enterprise: {
    themeType: "default",
    colorPrimary: "#16294f",
    borderRadius: 8,
    isCompact: false,
  },
};

// "Community", "Commercial" or "Enterprise", from the license the backend runs with
export let Edition = "Community";

export function getThemeDefault() {
  return EditionThemeDefaults[Edition] ?? ThemeDefault;
}

export const CustomFooter = null;

// Maximum number of navbar items before switching from flat to grouped menu
export let MaxItemsForFlatMenu = 7;

// The demo sites rotate the Marketplace themes on their pages and let visitors pick one
export let ThemeShowcase = false;
export let MarketplaceUrl = "https://cdn.jsdelivr.net/gh/casdoor/casdoor-integrations@dist";

// the links to Casdoor Cloud in the console, which a global admin can turn off in the global settings
export let ShowCloudPromotion = true;

// setConfig updates the frontend configuration from backend
export function setConfig(config: Record<string, any>) {
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
  if (config.defaultApplication !== undefined) {
    DefaultApplication = config.defaultApplication;
  }
  if (config.maxItemsForFlatMenu !== undefined) {
    MaxItemsForFlatMenu = config.maxItemsForFlatMenu;
  }
  if (config.themeShowcase !== undefined) {
    ThemeShowcase = config.themeShowcase;
  }
  if (config.marketplaceUrl) {
    MarketplaceUrl = config.marketplaceUrl;
  }
  if (config.showCloudPromotion !== undefined) {
    ShowCloudPromotion = config.showCloudPromotion;
  }
  if (config.edition) {
    setEdition(config.edition);
  }
}

export const Editions = ["Community", "Commercial", "Enterprise"];

function setEdition(edition: string) {
  Edition = edition;
  // index.css styles the console and the default sign-in page per edition
  document.documentElement.dataset.edition = Edition.toLowerCase();
}

const DEMO_EDITION_KEY = "casdoorDemoEdition";

function findEdition(value: string | null | undefined): string | undefined {
  return Editions.find((edition) => edition.toLowerCase() === value?.toLowerCase());
}

/**
 * A demo site lets the visitor try on the look of each edition: `?edition=enterprise` or the
 * switchers of EditionPreview pick one, kept in the visitor's own browser. Only the look changes,
 * the backend keeps running as the edition its license gives.
 */
function applyDemoEdition() {
  if (!ThemeShowcase) {
    return;
  }

  const fromUrl = findEdition(new URLSearchParams(window.location.search).get("edition"));
  let edition = fromUrl;
  try {
    // a page embedded with ?edition= must not change the site's own choice
    if (fromUrl && window.self === window.top) {
      localStorage.setItem(DEMO_EDITION_KEY, fromUrl);
    } else if (!fromUrl) {
      edition = findEdition(localStorage.getItem(DEMO_EDITION_KEY));
    }
  } catch {
    // the choice then only holds for this page
  }

  if (edition) {
    setEdition(edition);
  }
}

export function setDemoEdition(edition: string) {
  try {
    localStorage.setItem(DEMO_EDITION_KEY, edition);
    // the default theme is the one that shows the edition's look, so the theme carousel stops on it
    sessionStorage.setItem("themeShowcase", JSON.stringify({selected: "default", playing: false, expanded: false}));
  } catch {
    // fall through to the address, which carries the choice for this page
  }

  const url = new URL(window.location.href);
  url.searchParams.set("edition", edition.toLowerCase());
  url.searchParams.delete("theme");
  window.location.replace(url.toString());
}

export function initConfigFromCookie() {
  if (typeof document === "undefined") {
    return;
  }

  try {
    const curCookie = Cookie.parse(document.cookie);
    // a page embedded in another site gets no cookie, the server also puts the config into the page
    const raw = curCookie["jsonWebConfig"] ?? document.querySelector<HTMLMetaElement>("meta[name='casdoor-web-config']")?.content;
    if (!raw || raw === "null") {
      return;
    }

    const config = JSON.parse(raw);
    setConfig(config);
  } catch {
    // Ignore malformed cookie and keep compile-time defaults.
  }

  applyDemoEdition();
}
