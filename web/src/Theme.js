// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

import * as Setting from "./Setting";
import less from "less";

const themes = {
  light: {
    "@primary-color": "rgb(45,120,213)",
    "@layout-body-background": "#ffffff",
    "@layout-header-background": "#ffffff",
    "@body-background": "#ffffff",
    "@component-background": "#fff",
    "@heading-color": "rgba(0, 0, 0, 0.85)",
    "@text-color": "rgba(0, 0, 0, 0.85)",
    "@text-color-inverse": "#fff",
    "@text-color-secondary": "rgba(0, 0, 0, 0.45)",
    "@shadow-color": "rgba(0, 0, 0, 0.15)",
    "@border-color-split": "#f0f0f0",
    "@background-color-light": "#fafafa",
    "@background-color-base": "#ffffff",
    "@table-selected-row-bg": "#fafafa",
    "@table-expanded-row-bg": "#fbfbfb",
    "@checkbox-check-color": "#fff",
    "@disabled-color": "rgba(0, 0, 0, 0.25)",
    "@menu-dark-color": "rgba(1, 1, 1, 0.85)",
    "@menu-dark-highlight-color": "#fefefe",
    "@menu-dark-arrow-color": "#fefefe",
    "@btn-primary-color": "#fff",
  },
  dark: {
    "@primary-color": "#9b9b9b",
    "@layout-body-background": "#000",
    "@layout-header-background": "#141414",
    "@body-background": "#141414",
    "@component-background": "#141414",
    "@heading-color": "rgba(255, 255, 255, 0.85)",
    "@text-color": "rgba(255, 255, 255, 0.85)",
    "@text-color-inverse": "#141414",
    "@text-color-secondary": "rgba(255, 255, 255, 0.45)",
    "@shadow-color": "rgba(255, 255, 255, 0.15)",
    "@border-color-split": "#303030",
    "@background-color-light": "#2a2a2a",
    "@background-color-base": "#2a2a2a",
    "@table-selected-row-bg": "#ffffff",
    "@table-expanded-row-bg": "#ffffff0b",
    "@checkbox-check-color": "#141414",
    "@disabled-color": "rgba(255, 255, 255, 0.25)",
    "@menu-dark-color": "rgba(254, 254, 254, 0.65)",
    "@menu-dark-highlight-color": "#fefefe",
    "@menu-dark-arrow-color": "#fefefe",
    "@btn-primary-color": "#141414",
  },
};

export const setThemeColor = (theme) => {
  less.modifyVars(themes[theme]).catch((error) => {
    Setting.showMessage("error", `Failed to switch: ${error}`);
  });
  setTheme(theme);
};

export function getTheme() {
  const theme = localStorage.getItem("theme");
  if (theme === null) {
    setTheme("light");
    return "light";
  }
  return theme;
}

export function setTheme(theme) {
  localStorage.setItem("theme", theme);
}
