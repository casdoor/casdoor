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
import * as Setting from "../../Setting";
import {Dropdown, Select} from "antd";
import "../../App.less";
import {GlobalOutlined} from "@ant-design/icons";

function flagIcon(country, alt) {
  return (
    <img width={24} alt={alt} src={`${Setting.StaticBaseUrl}/flag-icons/${country}.svg`} />
  );
}

class LanguageSelect extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      selectedLanguage: Setting.getLanguage(),
    };

    Setting.Countries.forEach((country) => {
      new Image().src = `${Setting.StaticBaseUrl}/flag-icons/${country.country}.svg`;
    });
  }

  getLanguages() {
    return this.props.languages ?? Setting.Countries.map(item => item.key);
  }

  getLanguageItems(languages) {
    return Setting.Countries.filter((country) => languages.includes(country.key));
  }

  getSelectorMode() {
    return this.props.mode === "Label" ? "Label" : "Dropdown";
  }

  triggerLanguageChange(key) {
    if (typeof this.props.onClick === "function") {
      this.props.onClick(key);
    }
    Setting.setLanguage(key);
    this.setState({selectedLanguage: key});
  }

  renderDropdown(languageItems) {
    const items = languageItems.map((country) => Setting.getItem(country.label, country.key, flagIcon(country.country, country.alt)));
    const onClick = (e) => {
      this.triggerLanguageChange(e.key);
    };

    return (
      <Dropdown menu={{items, onClick}}>
        <div className="select-box" style={{display: items.length === 0 ? "none" : null, ...this.props.style}}>
          <GlobalOutlined style={{fontSize: "24px"}} />
        </div>
      </Dropdown>
    );
  }

  renderTextOnlySelect(languageItems) {
    if (languageItems.length === 0) {
      return null;
    }

    const options = languageItems.map((country) => Setting.getOption(country.label, country.key));
    let selectedLanguage = this.state.selectedLanguage || Setting.getLanguage();
    if (!options.some((item) => item.value === selectedLanguage)) {
      selectedLanguage = options[0].value;
    }

    return (
      <Select
        virtual={false}
        value={selectedLanguage}
        options={options}
        onChange={(value) => this.triggerLanguageChange(value)}
        style={{minWidth: "110px", ...this.props.style}}
      />
    );
  }

  render() {
    const languageItems = this.getLanguageItems(this.getLanguages());
    if (this.getSelectorMode() === "Label") {
      return this.renderTextOnlySelect(languageItems);
    }

    return this.renderDropdown(languageItems);
  }
}

export default LanguageSelect;
