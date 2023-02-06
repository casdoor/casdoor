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
import {Select} from "antd";

const {Option} = Select;

class SelectPhonePrefixBox extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      value: "",
    };
  }

  onChange(e) {
    this.props.onChange(e);
    this.setState({value: e});
  }

  render() {
    const isOrgnazition = this.props.isOrgnazition;
    const wid = isOrgnazition ? "100%" : 100;
    let phoneArray = [];
    if (isOrgnazition) {
      phoneArray = Setting.CountryPrefiexphone;
    } else {
      for (let i = 0; i < this.props.obj.phonePrefix.length; i++) {
        phoneArray.push(Setting.PhonePrefiexMap.get(this.props.obj.phonePrefix[i]));
      }
    }

    return (
      <Select virtual={false}
        mode={isOrgnazition ? "multiple" : ""}
        allowClear
        showSearch={true}
        optionFilterProp="label"
        style={{width: wid}}
        defaultValue={this.props.defaultPhone ? this.props.defaultPhone : undefined}
        placeholder= "please select phone prefix"
        dropdownMatchSelectWidth = {false}
        optionLabelProp={isOrgnazition ? "label" : "value"}
        onChange={(value => {
          this.onChange(value);
        })}
        filterOption={(input, option) =>
          (option?.label ?? "").toLowerCase().includes(input.toLowerCase())
        }
        filterSort={(optionA, optionB) =>
          (optionA?.label ?? "").toLowerCase().localeCompare((optionB?.label ?? "").toLowerCase())
        }
      >
        {
          phoneArray.map((item, index) => (
            <Option key={index} value={item.phone} label={`${item.name} (${item.code}) (${item.phone})`} >
              <img src={`${Setting.StaticBaseUrl}/flag-icons/${item.code}.svg`} alt={item.name} height={20} style={{marginRight: 10}} />
              {`${item.name} (${item.code}) +${item.phone}`}
            </Option>
          ))
        }
      </Select>
    );
  }
}

export default SelectPhonePrefixBox;
