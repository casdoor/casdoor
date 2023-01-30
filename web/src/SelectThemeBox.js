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

import React from "react";
import * as Setting from "./Setting";
import {Dropdown} from "antd";
import "./App.less";
import i18next from "i18next";

export const Themes = [
  {label: "Default", key: "default", icon: `${Setting.StaticBaseUrl}/img/light.svg`}, // i18next.t("theme:Default")
  {label: "Dark", key: "dark", icon: `${Setting.StaticBaseUrl}/img/dark.svg`}, // i18next.t("theme:Dark")
  {label: "Compact", key: "compact", icon: `${Setting.StaticBaseUrl}/img/compact.svg`}, // i18next.t("theme:Compact")
];

function themeIcon(themeKey, iconUrl) {
  return <img width={24} alt={themeKey} src={iconUrl} />;
}

function getIconURL(themeKey) {
  if (themeKey?.includes("dark")) {
    return Themes.find(t => t.key === "dark").icon;
  } else if (themeKey?.includes("default")) {
    return Themes.find(t => t.key === "default").icon;
  }
}

class SelectThemeBox extends React.Component {
  constructor(props) {
    super(props);
  }

  iconUrl = getIconURL(this.props.themeAlgorithm);
  items = this.getThemeItems();

  componentDidMount() {
    i18next.on("languageChanged", () => {
      this.items = this.getThemeItems();
    });
  }

  getThemeItems() {
    return Themes.map((theme) => Setting.getItem(i18next.t(`theme:${theme.label}`), theme.key, themeIcon(theme.key, theme.icon)));
  }

  render() {
    const onClick = (e) => {
      let nextTheme;
      if (e.key === "compact") {
        if (this.props.themeAlgorithm.includes("compact")) {
          nextTheme = this.props.themeAlgorithm.filter((theme) => theme !== "compact");
        } else {
          nextTheme = [...this.props.themeAlgorithm, "compact"];
        }
      } else {
        if (!this.props.themeAlgorithm.includes(e.key)) {
          if (e.key === "dark") {
            nextTheme = [...this.props.themeAlgorithm.filter((theme) => theme !== "default"), e.key];
          } else {
            nextTheme = [...this.props.themeAlgorithm.filter((theme) => theme !== "dark"), e.key];
          }
        } else {
          nextTheme = [...this.props.themeAlgorithm];
        }
      }

      this.iconUrl = getIconURL(nextTheme);
      this.props.onChange(nextTheme);
    };

    return (
      <Dropdown menu={{
        items: this.items,
        onClick,
        selectable: true,
        multiple: true,
        selectedKeys: [...this.props.themeAlgorithm]}} >
        <div className="theme-box" style={{background: `url(${this.iconUrl})`, ...this.props.style}} />
      </Dropdown>
    );
  }
}

export default SelectThemeBox;
