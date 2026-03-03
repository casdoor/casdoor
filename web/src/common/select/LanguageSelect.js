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
      classes: props,
      languages: props.languages ?? Setting.Countries.map(item => item.key),
      onClick: props.onClick,
    };

    Setting.Countries.forEach((country) => {
      new Image().src = `${Setting.StaticBaseUrl}/flag-icons/${country.country}.svg`;
    });
  }

  items = Setting.Countries.map((country) => Setting.getItem(country.label, country.key, flagIcon(country.country, country.alt)));

  getOrganizationLanguages(languages) {
    const select = [];
    for (const language of languages) {
      this.items.map((item, index) => item.key === language ? select.push(item) : null);
    }
    return select;
  }

  renderSelect(languageItems) {
    const currentLanguage = Setting.getLanguage();
    const validKeys = languageItems.map(item => item.key);
    const selectedValue = validKeys.includes(currentLanguage) ? currentLanguage : validKeys[0];

    const options = languageItems.map(item => ({
      value: item.key,
      label: (
        <span style={{display: "flex", alignItems: "center", gap: "8px"}}>
          {item.icon}
          {item.label}
        </span>
      ),
    }));

    return (
      <Select
        virtual={false}
        style={{width: "140px", ...this.props.style}}
        value={selectedValue}
        onChange={(value) => {
          if (typeof this.state.onClick === "function") {
            this.state.onClick(value);
          }
          Setting.setLanguage(value);
        }}
        options={options}
      />
    );
  }

  render() {
    const languageItems = this.getOrganizationLanguages(this.state.languages);

    if (this.props.type === "Select") {
      if (languageItems.length === 0) {
        return null;
      }
      return this.renderSelect(languageItems);
    }

    const onClick = (e) => {
      if (typeof this.state.onClick === "function") {
        this.state.onClick(e.key);
      }
      Setting.setLanguage(e.key);
    };

    return (
      <Dropdown menu={{items: languageItems, onClick}} >
        <div className="select-box" style={{display: languageItems.length === 0 ? "none" : null, ...this.props.style}} >
          <GlobalOutlined style={{fontSize: "24px"}} />
        </div>
      </Dropdown>
    );
  }
}

export default LanguageSelect;
