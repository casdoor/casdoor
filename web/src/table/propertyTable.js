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
import {Button, Input, Table} from "antd";
import i18next from "i18next";
import {DeleteOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";

class PropertyTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      properties: [],
    };

    // transfer the Object to object[]
    if (this.props.properties !== null) {
      Object.entries(this.props.properties).map((item, index) => {
        this.state.properties.push({key: index, name: item[0], value: item[1]});
      });
    }
  }

  page = 1;
  pageSize = 10;
  count = this.props.properties !== null ? Object.entries(this.props.properties).length : 0;

  updateTable(table) {
    this.setState({properties: table});
    const properties = {};
    table.map((item) => {
      properties[item.name] = item.value;
    });
    this.props.onUpdateTable(properties);
  }

  addRow(table) {
    const row = {key: this.count, name: "", value: ""};
    if (table === undefined) {
      table = [];
    }
    table = Setting.addRow(table, row);
    this.count = this.count + 1;
    this.updateTable(table);
  }

  deleteRow(table, index) {
    table = Setting.deleteRow(table, this.getIndex(index));
    this.updateTable(table);
  }

  getIndex(index) {
    // Need to be used in all place when modify table. Parameter is the row index in table, need to calculate the index in dataSource.
    return index + (this.page - 1) * this.pageSize;
  }

  updateField(table, index, key, value) {
    table[this.getIndex(index)][key] = value;
    this.updateTable(table);
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("user:Keys"),
        dataIndex: "name",
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
        title: i18next.t("user:Values"),
        dataIndex: "value",
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
        dataIndex: "operation",
        width: "20px",
        render: (text, record, index) => {
          return (
            <Button icon={<DeleteOutlined />} size="small" onClick={() => this.deleteRow(table, index)} />
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
      pagination={{
        defaultPageSize: this.pageSize,
        onChange: page => {this.page = page;},
      }}
      columns={columns} dataSource={table} rowKey="key" size="middle" bordered
      />
    );
  }

  render() {
    return (
      <React.Fragment>
        {
          this.renderTable(this.state.properties)
        }
      </React.Fragment>
    );
  }
}

export default PropertyTable;
