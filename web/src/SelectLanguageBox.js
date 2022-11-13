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
    };
  }

  render() {
    const menu = (
      <Menu onClick={(e) => {
        Setting.changeLanguage(e.key);
      }}>
        <Menu.Item key="en" icon={flagIcon("US", "English")}>English</Menu.Item>
        <Menu.Item key="zh" icon={flagIcon("CN", "简体中文")}>简体中文</Menu.Item>
        <Menu.Item key="es" icon={flagIcon("ES", "Español")}>Español</Menu.Item>
        <Menu.Item key="fr" icon={flagIcon("FR", "Français")}>Français</Menu.Item>
        <Menu.Item key="de" icon={flagIcon("DE", "Deutsch")}>Deutsch</Menu.Item>
        <Menu.Item key="ja" icon={flagIcon("JP", "日本語")}>日本語</Menu.Item>
        <Menu.Item key="ko" icon={flagIcon("KR", "한국어")}>한국어</Menu.Item>
        <Menu.Item key="ru" icon={flagIcon("RU", "Русский")}>Русский</Menu.Item>
      </Menu>
    );

    return (
      <Dropdown overlay={menu} >
        <div></div>
      </Dropdown>
    );
  }
}

export default SelectLanguageBox;
