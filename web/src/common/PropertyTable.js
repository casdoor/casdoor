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
import {DeleteOutlined} from "@ant-design/icons";
import {Button, Input, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

class PolicyTable extends React.Component {
  constructor(props) {
    super(props);
    const keys = Object.keys(this.props.properties);
    const properties = [];
    for (let i = 0; i < keys.length; i++) {
      const property = new Object();
      property.index = crypto.randomUUID();
      property.key = keys[i];
      property.value = this.props.properties[keys[i]];
      properties[i] = property;
    }
    this.state = {
      properties: properties,
      loading: false,
    };
  }

  updateTable(table) {
    this.setState({properties: table});
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {index: crypto.randomUUID()};
    table = Setting.addRow(table, row, "top");
    if (table === undefined) {
      table = [];
    }
    this.updateTable(table);
  }

  deleteRow(table, i) {
    table = Setting.deleteRow(table, i);
    this.updateTable(table);
  }

  renderTable(table) {
    const columns = [
      {
        title: "Keys",
        dataIndex: "key",
        width: "100px",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "key", e.target.value);
            }} />
          );
        },
      },
      {
        title: "Values",
        dataIndex: "value",
        width: "100px",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "value", e.target.value);
            }} />
          );
        },
      },
      {
        title: "Option",
        key: "option",
        width: "30px",
        render: (text, record, index) => {
          return (
            <span>
              <Tooltip placement="topLeft" title={i18next.t("general:Delete")}>
                <Button icon={<DeleteOutlined />} size="small" onClick={() => this.deleteRow(table, index)} />
              </Tooltip>
            </span>
          );
        },
      },
    ];

    return (
      <Table
        columns={columns} dataSource={table} rowKey="index" size="middle" bordered pagination={false}
        loading={this.state.loading}
        title={() => (
          <div>
            <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
          </div>
        )}
      />
    );
  }

  render() {
    return (<>
      {
        this.renderTable(this.state.properties)
      }
    </>
    );
  }
}

export default PolicyTable;
