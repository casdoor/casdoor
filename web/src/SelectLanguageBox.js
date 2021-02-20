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
        <div class="box" style={{width: "600px"}}>
          <a href="javascript:void(0)" onClick={() => Setting.changeLanguage("en")} class="lang-selector">English</a>/
          <a href="javascript:void(0)" onClick={() => Setting.changeLanguage("zh")} class="lang-selector">简体中文</a>
        </div>
      </div>
    )
  }
}

export default SelectLanguageBox;
