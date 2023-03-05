// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import {Select} from "antd";
import * as Setting from "../Setting";
import React from "react";

export const PhoneNumberInput = (props) => {
  const {onChange, style, disabled, value} = props;
  const countryCodes = props.countryCodes ?? [];

  const handleOnChange = (value) => {
    onChange?.(value);
  };

  return (
    <Select
      virtual={false}
      showSearch
      style={style}
      disabled={disabled}
      value={value}
      dropdownMatchSelectWidth={false}
      optionLabelProp={"label"}
      onChange={handleOnChange}
      filterOption={(input, option) => (option?.text ?? "").toLowerCase().includes(input.toLowerCase())}
    >
      {
        Setting.getCountryCodeData(countryCodes).map((country) => Setting.getCountryCodeOption(country))
      }
    </Select>
  );
};
