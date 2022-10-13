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
import {Input, Select} from "antd";
import i18next from "i18next";

class CaptchaWaySelect extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      way: props.captchaWay,
      onChange: props.onChange,
    };

    // ensure backward compatibility
    if (!props.captchaWay) {
      this.handleWayChange("never");
    }
  }

  handleWayChange(way) {
    let captchaWay;
    if (way === "never" || way === "everytime") {
      this.setState({way: way});
      captchaWay = way;
    } else {
      this.setState({way: "never"});
      captchaWay = "never";
    }
    this.state.onChange(captchaWay);
  }

  render() {
    return (
      <Input.Group compact>
        <Select style={{width: "70%"}} value={this.state.way} onChange={value => this.handleWayChange(value)}>
          <Select.Option key="never" value="never">{i18next.t("general:Never")}</Select.Option>
          <Select.Option key="everytime" value="everytime">{i18next.t("general:Everytime")}</Select.Option>
        </Select>
      </Input.Group>
    );
  }
}

export default CaptchaWaySelect;
