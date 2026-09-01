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
import {Col, Row} from "antd";
import * as Setting from "../Setting";

class SamlWidget extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      addressOptions: [],
      affiliationOptions: [],
    };
  }

  renderIdp(user, application, providerItem) {
    const provider = providerItem.provider;
    const name = user.name;

    return (
      <Row key={provider.name} style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={this.props.labelSpan}>
          {
            Setting.getProviderLogo(provider)
          }
          <span style={{marginLeft: "5px"}}>
            {
              `${provider.type}:`
            }
          </span>
        </Col>
        <Col span={24 - this.props.labelSpan} style={{marginTop: "5px"}}>
          <span style={{
            width: this.props.labelSpan === 3 ? "300px" : "130px",
            display: (Setting.isMobile()) ? "inline" : "inline-block"}}>{name}</span>
        </Col>
      </Row>
    );
  }

  render() {
    return this.renderIdp(this.props.user, this.props.application, this.props.providerItem);
  }
}

export default SamlWidget;
