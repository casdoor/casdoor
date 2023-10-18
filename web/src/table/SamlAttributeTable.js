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

import React from "react";
import {DeleteOutlined, DownOutlined, UpOutlined} from "@ant-design/icons";
import {Button, Col, Input, Row, Select, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

const {Option} = Select;

class SamlAttributeTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  updateTable(table) {
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {Name: "", nameformat: "", value: ""};
    if (table === undefined || table === null) {
      table = [];
    }
    table = Setting.addRow(table, row);
    this.updateTable(table);
  }

  deleteRow(table, i) {
    table = Setting.deleteRow(table, i);
    this.updateTable(table);
  }

  upRow(table, i) {
    table = Setting.swapRow(table, i - 1, i);
    this.updateTable(table);
  }

  downRow(table, i) {
    table = Setting.swapRow(table, i, i + 1);
    this.updateTable(table);
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("user:Name"),
        dataIndex: "name",
        key: "name",
        width: "200px",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "name", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("user:Name format"),
        dataIndex: "nameformat",
        key: "nameformat",
        width: "200px",
        render: (text, record, index) => {
          return (
            <Select virtual={false} style={{width: "100%"}}
              value={text}
              defaultValue="urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified"
              onChange={value => {
                this.updateField(table, index, "nameformat", value);
              }} >
              <Option key="Unspecified" value="urn:oasis:names:tc:SAML:2.0:attrname-format:unspecified">{i18next.t("general:Unspecified")}</Option>
              <Option key="Basic" value="urn:oasis:names:tc:SAML:2.0:attrname-format:basic">{i18next.t("application:Basic")}</Option>
              <Option key="UriReference" value="urn:oasis:names:tc:SAML:2.0:attrname-format:uri">{i18next.t("application:UriReference")}</Option>
              <Option key="x500AttributeName" value="urn:oasis:names:tc:SAML:2.0:attrname-format:X500">{i18next.t("application:x500AttributeName")}</Option>
            </Select>
          );
        },
      },
      {
        title: i18next.t("user:Value"),
        dataIndex: "value",
        key: "value",
        width: "200px",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "value", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "action",
        key: "action",
        width: "20px",
        render: (text, record, index) => {
          return (
            <div>
              <Tooltip placement="bottomLeft" title={i18next.t("general:Up")}>
                <Button style={{marginRight: "5px"}} disabled={index === 0} icon={<UpOutlined />} size="small" onClick={() => this.upRow(table, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Down")}>
                <Button style={{marginRight: "5px"}} disabled={index === table.length - 1} icon={<DownOutlined />} size="small" onClick={() => this.downRow(table, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Delete")}>
                <Button icon={<DeleteOutlined />} size="small" onClick={() => this.deleteRow(table, index)} />
              </Tooltip>
            </div>
          );
        },
      },
    ];

    return (
      <Table title={() => (
        <div>
          <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
        </div>
      )}
      columns={columns} dataSource={table} rowKey="key" size="middle" bordered
      />
    );
  }

  render() {
    return (
      <div>
        <Row style={{marginTop: "20px"}} >
          <Col span={24}>
            {
              this.renderTable(this.props.table)
            }
          </Col>
        </Row>
      </div>
    );
  }
}

export default SamlAttributeTable;
