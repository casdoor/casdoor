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
import i18next from "i18next";
import * as Setting from "../../Setting";
import React from "react";

const {Option} = Select;

export const CountryCodeSelect = (props) => {
  const {onChange, style, disabled, initValue, mode} = props;
  const countryCodes = props.countryCodes ?? [];
  const [value, setValue] = React.useState("");

  React.useEffect(() => {
    if (initValue !== undefined) {
      setValue(initValue);
    } else {
      const initValue = countryCodes.length > 0 ? countryCodes[0] : "";
      handleOnChange(initValue);
    }
  }, []);

  const handleOnChange = (value) => {
    setValue(value);
    onChange?.(value);
  };

  return (
    <Select
      virtual={false}
      showSearch
      style={style}
      disabled={disabled}
      value={value}
      mode={mode}
      dropdownMatchSelectWidth={false}
      optionLabelProp={"label"}
      onChange={handleOnChange}
      filterOption={(input, option) => (option?.text ?? "").toLowerCase().includes(input.toLowerCase())}
    >
      {
        props.hasDefault ? (<Option key={"All"} value={"All"} label={i18next.t("organization:All")} text={"organization:All"} >
          <div style={{display: "flex", justifyContent: "space-between", marginRight: "10px"}}>
            {i18next.t("organization:All")}
          </div>
        </Option>) : null
      }
      {
        Setting.getCountryCodeData(countryCodes).map((country) => Setting.getCountryCodeOption(country))
      }
    </Select>
  );
};
