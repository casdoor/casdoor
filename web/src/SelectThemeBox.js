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

import React from "react";
import * as Setting from "./Setting";
import {Dropdown} from "antd";
import "./App.less";

function flagIcon(themeKey) {
  if (themeKey === "Dark") {
    return <img width={24} alt={themeKey} src={`${Setting.StaticBaseUrl}/img/dark.svg`} />;
  }
  if (themeKey === "Default") {
    return <img width={24} alt={themeKey} src={`${Setting.StaticBaseUrl}/img/light.svg`} />;
  }
  if (themeKey === "Compact") {
    return <img width={24} alt={themeKey} src={`${Setting.StaticBaseUrl}/img/light.svg`} />;
  }
}

class SelectLanguageBox extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      themes: props.theme ?? ["Default", "Dark", "Compact"],
      icon: null,
    };
  }

  items = Setting.Themes.map((theme) => Setting.getItem(theme.label, theme.key, flagIcon(theme.key)));

  componentDidMount() {
    if (localStorage.getItem("theme") === null) {
      this.setState({"icon": `${Setting.StaticBaseUrl}/img/light.svg`});
    } else {
      this.setState({"icon": Setting.Themes.find(t => t.key === localStorage.getItem("theme"))["selectThemeLogo"]});
    }
    addEventListener("themeChange", (e) => {
      this.setState({"icon": Setting.Themes.find(t => t.key === localStorage.getItem("theme"))["selectThemeLogo"]});
    });
  }

  getOrganizationThemes(themes) {
    const select = [];
    for (const theme of themes) {
      this.items.map((item, index) => item.key === theme ? select.push(item) : null);
    }
    return select;
  }

  render() {
    const themeItems = this.getOrganizationThemes(this.state.themes);
    const onClick = (e) => {
      Setting.setTheme(e.key);
    };

    return (
      <Dropdown menu={{items: themeItems, onClick}} >
        <div className="theme-box" style={{display: themeItems.length === 0 ? "none" : null, background: `url(${this.state.icon})`, ...this.props.style}} />
      </Dropdown>
    );
  }
}

export default SelectLanguageBox;
