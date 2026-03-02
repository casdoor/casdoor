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
import {Button, Col, Input, Row, Select, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

const {Option} = Select;

class IpRuleTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      options: [],
      defaultRules: [
        {
          name: "loopback",
          operator: "is in",
          value: "127.0.0.1",
        },
        {
          name: "lan cidr",
          operator: "is in",
          value: "10.0.0.0/8,192.168.0.0/16",
        },
      ],
    };
    if (this.props.table.length === 0) {
      this.restore();
    }
    for (let i = 0; i < this.props.table.length; i++) {
      const values = this.props.table[i].value.split(",");
      const options = [];
      for (let j = 0; j < values.length; j++) {
        options[j] = {value: values[j], label: values[j]};
      }
      this.state.options.push(options);
    }
  }

  updateTable(table) {
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    if (key === "value") {
      let v = "";
      for (let i = 0; i < value.length; i++) {
        v += value[i].trim() + ",";
      }
      table[index][key] = v.slice(0, -1);
    } else {
      table[index][key] = value;
    }
    this.updateTable(table);
  }

  addRow(table) {
    const row = {name: `New IP Rule - ${table.length}`, operator: "is in", value: "127.0.0.1"};
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
    Setting.swapRow(this.state.options, i - 1, i);
    this.updateTable(table);
  }

  downRow(table, i) {
    table = Setting.swapRow(table, i, i + 1);
    Setting.swapRow(this.state.options, i, i + 1);
    this.updateTable(table);
  }

  restore() {
    this.updateTable(this.state.defaultRules);
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "180px",
        render: (text, record, index) => (
          <Input value={text} onChange={e => {
            this.updateField(table, index, "name", e.target.value);
          }} />
        ),
      },
      {
        title: i18next.t("rule:Operator"),
        dataIndex: "operator",
        key: "operator",
        width: "180px",
        render: (text, record, index) => (
          <Select value={text} virtual={false} style={{width: "100%"}} onChange={value => {
            this.updateField(table, index, "operator", value);
          }}>
            {
              [
                {value: "is in", text: i18next.t("rule:is in")},
                {value: "is not in", text: i18next.t("rule:is not in")},
                {value: "is abroad", text: i18next.t("rule:is abroad")},
              ].map((item, index) => <Option key={index} value={item.value}>{item.text}</Option>)
            }
          </Select>
        ),
      },
      {
        title: i18next.t("rule:IP List"),
        dataIndex: "value",
        key: "value",
        render: (text, record, index) => (
          <Select
            mode="tags"
            style={{width: "100%"}}
            placeholder="Input IP Addresses"
            value={record.value ? record.value.split(",") : []}
            onChange={value => this.updateField(table, index, "value", value)}
            options={this.state.options[index]}
          />
        ),
      },
      {
        title: i18next.t("general:Action"),
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
            <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.restore()}>{i18next.t("general:Restore")}</Button>
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

export default IpRuleTable;
