// Copyright 2023 The casbin Authors. All Rights Reserved.
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
import {Button, Col, Row, Select, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

const {Option} = Select;

class RuleTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
    if (this.props.rules === null) {
      // rerender
      this.props.onUpdateRules([]);
    }
  }

  updateTable(table) {
    const rules = [];
    for (let i = 0; i < table.length; i++) {
      rules.push(table[i].owner + "/" + table[i].name);
    }
    this.props.onUpdateRules(rules);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {owner: this.props.account.name, name: ""};
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
        width: "180px",
        render: (text, record, index) => (
          <Select value={text} virtual={false} style={{width: "100%"}} onChange={value => {
            this.updateField(table, index, "name", value);
          }}>
            {
              Setting.getDeduplicatedArray(this.props.sources, table, "name").map((record, index) => {
                return <Option key={record.name} value={record.name}>{record.name}</Option>;
              })
            }
          </Select>
        ),
      },
      {
        title: "Action",
        key: "action",
        width: "100px",
        render: (text, record, index) => (
          <div>
            <Tooltip placement="bottomLeft" title={"Up"}>
              <Button style={{marginRight: "5px"}} disabled={index === 0} icon={<UpOutlined />} size="small" onClick={() => this.upRow(table, index)} />
            </Tooltip>
            <Tooltip placement="topLeft" title={"Down"}>
              <Button style={{marginRight: "5px"}} disabled={index === table.length - 1} icon={<DownOutlined />} size="small" onClick={() => this.downRow(table, index)} />
            </Tooltip>
            <Tooltip placement="topLeft" title={"Delete"}>
              <Button icon={<DeleteOutlined />} size="small" onClick={() => this.deleteRow(table, index)} />
            </Tooltip>
          </div>
        ),
      },
    ];
    return (
      <Table rowKey="index" columns={columns} dataSource={table} size="middle" bordered pagination={false}
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
              this.props.rules === null ? null : this.renderTable(this.props.rules.map((item, index) => {
                const values = item.split("/");
                return {owner: values[0], name: values[1]};
              }))
            }
          </Col>
        </Row>
      </div>
    );
  }
}

export default RuleTable;
