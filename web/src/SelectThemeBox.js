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

function flagIcon(country, alt) {
  return (
    <img width={24} alt={alt} src={`${Setting.StaticBaseUrl}/flag-icons/${country}.svg`} />
  );
}

class SelectLanguageBox extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      themes: props.theme ?? ["casbin", "calm", "lava"],
    };
  }

  items = Setting.Themes.map((theme) => Setting.getItem(theme.label, theme.key, flagIcon(theme.country, theme.alt)));

  getOrganizationThemes(languages) {
    const select = [];
    for (const language of languages) {
      this.items.map((item, index) => item.key === language ? select.push(item) : null);
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
        <div className="language-box" style={{display: themeItems.length === 0 ? "none" : null, ...this.props.style}} />
      </Dropdown>
    );
  }
}

export default SelectLanguageBox;
