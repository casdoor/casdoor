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
import {Select} from "antd";
import "../../App.less";

const {Option} = Select;

function flagIcon(country, alt) {
  return (
    <img width={20} alt={alt} src={`${Setting.StaticBaseUrl}/flag-icons/${country}.svg`} style={{marginRight: 8}} />
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

  getCountryByLanguage(languageKey) {
    return Setting.Countries.find(country => country.key === languageKey);
  }

  render() {
    const currentLanguage = Setting.getLanguage();

    const onChange = (value) => {
      if (typeof this.state.onClick === "function") {
        this.state.onClick(value);
      }
      Setting.setLanguage(value);
    };

    return (
      <Select
        value={currentLanguage}
        onChange={onChange}
        style={{width: 150, display: this.state.languages.length === 0 ? "none" : null, ...this.props.style}}
      >
        {this.state.languages.map((langKey) => {
          const country = this.getCountryByLanguage(langKey);
          if (!country) {return null;}
          return (
            <Option key={langKey} value={langKey}>
              {flagIcon(country.country, country.alt)}
              {country.label}
            </Option>
          );
        })}
      </Select>
    );
  }
}

export default LanguageSelect;
