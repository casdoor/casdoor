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
import {createFromIconfontCN} from "@ant-design/icons";
import "./App.less";

const IconFont = createFromIconfontCN({
  scriptUrl: "//at.alicdn.com/t/font_2680620_ffij16fkwdg.js",
});

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
        <Menu.Item key="en" icon={<IconFont type="icon-en" />}>English</Menu.Item>
        <Menu.Item key="zh" icon={<IconFont type="icon-zh" />}>简体中文</Menu.Item>
        <Menu.Item key="fr" icon={<IconFont type="icon-fr" />}>Français</Menu.Item>
        <Menu.Item key="de" icon={<IconFont type="icon-de" />}>Deutsch</Menu.Item>
        <Menu.Item key="ja" icon={<IconFont type="icon-ja" />}>日本語</Menu.Item>
        <Menu.Item key="ko" icon={<IconFont type="icon-ko" />}>한국어</Menu.Item>
        <Menu.Item key="ru" icon={<IconFont type="icon-ru" />}>Русский</Menu.Item>
      </Menu>
    );

    return (
      <Dropdown overlay={menu} >
        <div className="language_box" />
      </Dropdown>
    );
  }
}

export default SelectLanguageBox;
