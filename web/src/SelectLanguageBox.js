// Copyright 2021 The casbin Authors. All Rights Reserved.
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

class SelectLanguageBox extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  render() {
    return (
      <div align="center">
        <div className="box" style={{width: "600px"}}>
          {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
          <a onClick={() => Setting.changeLanguage("en")} className="lang-selector">
            English
          </a>
          /
          {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
          <a onClick={() => Setting.changeLanguage("zh")} className="lang-selector">
            简体中文
          </a>
          /
          {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
          <a onClick={() => Setting.changeLanguage("fr")} className="lang-selector">
            Français
          </a>
          /
          {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
          <a onClick={() => Setting.changeLanguage("de")} className="lang-selector">
            Deutsch
          </a>
          /
          {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
          <a onClick={() => Setting.changeLanguage("ja")} className="lang-selector">
            日本語
          </a>
          /
          {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
          <a onClick={() => Setting.changeLanguage("ko")} className="lang-selector">
            한국어
          </a>
          /
          {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
          <a onClick={() => Setting.changeLanguage("ru")} className="lang-selector">
            Русский
          </a>
        </div>
      </div>
    )
  }
}

export default SelectLanguageBox;
