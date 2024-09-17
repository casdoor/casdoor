// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
import {Select} from "antd";
import i18next from "i18next";

const {Option} = Select;

class CustomItemSelect extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      value: props.defaultValue || [],
    };
  }

  onChange(value) {
    this.props.onChange(value);
    this.setState({value});
  }

  render() {
    const {options, isMultiple, isSingle, placeholder} = this.props;

    return (
      <Select
        virtual={false}
        mode={isMultiple ? "multiple" : (isSingle ? "single" : undefined)}
        showSearch
        allowClear={true}
        optionFilterProp="label"
        style={{width: "100%"}}
        placeholder={isMultiple ? "" : placeholder || i18next.t("Please select an item")}
        onChange={(value) => this.onChange(value)}
        filterOption={(input, option) =>
          (option?.label ?? "").toLowerCase().includes(input.toLowerCase())
        }
        filterSort={(optionA, optionB) =>
          (optionA?.label ?? "").toLowerCase().localeCompare((optionB?.label ?? "").toLowerCase())
        }
        value={Array.isArray(this.state.value) && this.state.value.length ? this.state.value : undefined}
      >
        {options.map((item) => (
          <Option key={item.value} value={item.value} label={item.label}>
            {item.label}
          </Option>
        ))}
      </Select>
    );
  }
}

export default CustomItemSelect;
