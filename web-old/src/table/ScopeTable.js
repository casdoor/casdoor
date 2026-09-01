// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
import {Button, Input, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

class ScopeTable extends React.Component {
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
    const row = {name: "", displayName: "", description: ""};
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
    if (table === null) {
      return null;
    }

    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "25%",
        render: (text, record, index) => {
          return (
            <Input
              value={text}
              placeholder="e.g., files:read"
              onChange={e => {
                this.updateField(table, index, "name", e.target.value);
              }}
            />
          );
        },
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "25%",
        render: (text, record, index) => {
          return (
            <Input
              value={text}
              placeholder="e.g., Read Files"
              onChange={e => {
                this.updateField(table, index, "displayName", e.target.value);
              }}
            />
          );
        },
      },
      {
        title: i18next.t("general:Description"),
        dataIndex: "description",
        key: "description",
        width: "40%",
        render: (text, record, index) => {
          return (
            <Input
              value={text}
              placeholder="e.g., Allow reading your files and documents"
              onChange={e => {
                this.updateField(table, index, "description", e.target.value);
              }}
            />
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: "10%",
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
      <div>
        <Table scroll={{x: "max-content"}} rowKey={(record, index) => index} columns={columns} dataSource={table} size="middle" bordered pagination={false}
          title={() => (
            <div>
              {this.props.title}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
            </div>
          )}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.props.table)
        }
      </div>
    );
  }
}

export default ScopeTable;
