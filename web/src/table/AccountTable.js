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
import {DeleteOutlined, DownOutlined, UpOutlined} from "@ant-design/icons";
import {Button, Col, Input, Row, Select, Switch, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

const {Option} = Select;

class AccountTable extends React.Component {
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
    const row = {name: Setting.getNewRowNameForTable(table, "Please select an account item"), visible: true, viewRule: "Public", modifyRule: "Self", tab: ""};
    if (table === undefined) {
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
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        render: (text, record, index) => {
          const items = Setting.GetTranslatedUserItems();
          return (
            <Select virtual={false} style={{width: "100%"}}
              options={Setting.getDeduplicatedArray(items, table, "name").map(item => Setting.getOption(item.label, item.name))}
              value={text}
              onChange={value => {
                this.updateField(table, index, "name", value);
              }} >
            </Select>
          );
        },
      },
      {
        title: i18next.t("organization:Visible"),
        dataIndex: "visible",
        key: "visible",
        width: "120px",
        render: (text, record, index) => {
          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "visible", checked);
            }} />
          );
        },
      },
      {
        title: i18next.t("general:Tab"),
        dataIndex: "tab",
        key: "tab",
        width: "150px",
        render: (text, record, index) => {
          return (
            <Input value={text} placeholder={i18next.t("general:Default")} onChange={e => {
              this.updateField(table, index, "tab", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("signup:Regex"),
        dataIndex: "regex",
        key: "regex",
        width: "200px",
        render: (text, record, index) => {
          const regexIncludeList = ["Display name", "Password", "Email", "Phone", "Location",
            "Title", "Homepage", "Bio", "Gender", "Birthday", "Education", "ID card",
            "ID card type"];
          if (!regexIncludeList.includes(record.name)) {
            return null;
          }

          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "regex", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("organization:View rule"),
        dataIndex: "viewRule",
        key: "viewRule",
        width: "155px",
        render: (text, record, index) => {
          if (!record.visible) {
            return null;
          }

          const options = [
            {id: "Public", name: "Public"},
            {id: "Self", name: "Self"},
            {id: "Admin", name: "Admin"},
          ];

          return (
            <Select virtual={false} style={{width: "100%"}} value={text} onChange={(value => {
              this.updateField(table, index, "viewRule", value);
            })}>
              {
                options.map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          );
        },
      },
      {
        title: i18next.t("organization:Modify rule"),
        dataIndex: "modifyRule",
        key: "modifyRule",
        width: "155px",
        render: (text, record, index) => {
          if (!record.visible) {
            return null;
          }

          let options;
          if (record.viewRule === "Admin" || record.name === "Is admin") {
            options = [
              {id: "Admin", name: "Admin"},
              {id: "Immutable", name: "Immutable"},
            ];
          } else {
            options = [
              {id: "Self", name: "Self"},
              {id: "Admin", name: "Admin"},
              {id: "Immutable", name: "Immutable"},
            ];
          }

          return (
            <Select virtual={false} style={{width: "100%"}} value={text} onChange={(value => {
              this.updateField(table, index, "modifyRule", value);
            })}>
              {
                options.map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: "100px",
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
      <Table scroll={{x: "max-content"}} rowKey="name" columns={columns} dataSource={table} size="middle" bordered pagination={false}
        title={() => (
          <div>
            {this.props.title}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
          </div>
        )}
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

export default AccountTable;
