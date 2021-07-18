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
import {Button, Col, Input, Row, Table} from "antd";
import i18next from "i18next";

export default class ProviderPropertiesTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      properties: this.props.properties,
    }
  }

  convertPropertiesToAntdDataSource(properties) {
    let ret = [];
    for (let k in properties) {
      ret.push({
        key: k,
        value: properties[k]
      })
    }
    return ret;
  }

  updateProperty(key, value) {
    let properties = this.state.properties;
    properties[key] = value;
    this.setProperties(properties);
  }

  removeProperty(key) {
    let properties = this.state.properties;
    delete properties[key];
    this.setProperties(properties);
  }

  setProperties(properties) {
    this.props.onPropertyChange(properties);
    this.setState({ properties: properties })
  }

  render() {
    if (this.state.properties === undefined) return null;

    if (this.state.properties === null) {
      this.setProperties({})
      return null;
    }

    return (<Row style={{marginTop: '20px'}} >
      <Col style={{marginTop: '5px'}} span={2}>
        {`${i18next.t("provider:Provider Properties")}: `}
      </Col>
      <Col span={22} >
        <Button
          type={"primary"}
          size={"small"}
          onClick={() => this.updateProperty('', '')}
        >
          {i18next.t("provider:Add")}
        </Button>
        <Table
          dataSource={this.convertPropertiesToAntdDataSource(this.state.properties)}
          columns={[
            {
              title: i18next.t("provider:Property Key"),
              dataIndex: "key",
              key: "key",
              render: (text, record) => {
                return <Input
                  onChange={e => {
                    this.removeProperty(record.key);
                    this.updateProperty(e.target.value, record.value);
                  }}
                  placeholder={i18next.t("provider:Please input your key")}
                  value={text}
                />
              }
            },
            {
              title: i18next.t("provider:Property Value"),
              dataIndex: "value",
              key: "value",
              render: (text, record) => {
                return <Input
                  onChange={e => this.updateProperty(record.key, e.target.value)}
                  placeholder={i18next.t("provider:Please input your value")}
                  value={text}
                />
              }
            },
            {
              title: i18next.t("provider:Action"),
              dataIndex: 'prompted',
              key: 'prompted',
              render: (text, record) => {
                return <Button
                  type={"primary"}
                  size={"small"}
                  onClick={() => this.removeProperty(record.key)}
                >
                  {i18next.t("provider:Delete")}
                </Button>
              }
            }
          ]}
        />
      </Col>
    </Row>)
  }
}
