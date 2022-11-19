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
import {Dropdown, Menu} from "antd";
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
      languages: props.languages ?? ["en", "zh", "es", "fr", "de", "ja", "ko", "ru"],
    };
  }

  items = [
    this.getItem("English", "en", flagIcon("US", "English")),
    this.getItem("简体中文", "zh", flagIcon("CN", "简体中文")),
    this.getItem("Español", "es", flagIcon("ES", "Español")),
    this.getItem("Français", "fr", flagIcon("FR", "Français")),
    this.getItem("Deutsch", "de", flagIcon("DE", "Deutsch")),
    this.getItem("日本語", "ja", flagIcon("JP", "日本語")),
    this.getItem("한국어", "ko", flagIcon("KR", "한국어")),
    this.getItem("Русский", "ru", flagIcon("RU", "Русский")),
  ];

  getOrganizationLanguages(languages) {
    const select = [];
    for (const language of languages) {
      this.items.map((item, index) => item.key === language ? select.push(item) : null);
    }
    return select;
  }

  getItem(label, key, icon) {
    return {key, icon, label};
  }

  render() {
    const languageItems = this.getOrganizationLanguages(this.state.languages);
    const menu = (
      <Menu items={languageItems} onClick={(e) => {
        Setting.setLanguage(e.key);
      }}>
      </Menu>
    );

    return (
      <Dropdown overlay={menu} >
        <div className="language-box" style={{display: languageItems.length === 0 ? "none" : null, ...this.props.style}} />
      </Dropdown>
    );
  }
}

export default SelectLanguageBox;
