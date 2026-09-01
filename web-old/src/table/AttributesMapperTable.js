// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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

class AttributesMapperTable extends React.Component {
  constructor(props) {
    super(props);

    // transfer the Object to object[]
    const customAttributes = this.props.customAttributes !== null
      ? Object.entries(this.props.customAttributes).map((item, index) => ({
        key: index,
        attributeName: item[0],
        userPropertyName: item[1],
      }))
      : [];
    this.state = {
      customAttributes: customAttributes,
      page: 1,
    };
  }

  pageSize = 10;
  count = this.props.customAttributes !== null ? Object.entries(this.props.customAttributes).length : 0;

  updateTable(table) {
    this.setState({customAttributes: table});
    const customAttributes = {};
    table.forEach((item) => {
      customAttributes[item.attributeName] = item.userPropertyName;
    });
    this.props.onUpdateTable(customAttributes);
  }

  addRow(table) {
    const row = {key: this.count, attributeName: "", userPropertyName: ""};
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
    return index + (this.state.page - 1) * this.pageSize;
  }

  updateField(table, index, attributeName, userPropertyName) {
    table[this.getIndex(index)][attributeName] = userPropertyName;
    this.updateTable(table);
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("ldap:LDAP attribute name"),
        dataIndex: "attributeName",
        width: "200px",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "attributeName", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("ldap:User property name"),
        dataIndex: "userPropertyName",
        width: "200px",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "userPropertyName", e.target.value);
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
        current: this.state.page,
        onChange: page => {this.setState({page});},
      }}
      columns={columns} dataSource={table} rowKey="key" size="middle" bordered
      />
    );
  }

  render() {
    return (
      <React.Fragment>
        {
          this.renderTable(this.state.customAttributes)
        }
      </React.Fragment>
    );
  }
}

export default AttributesMapperTable;
