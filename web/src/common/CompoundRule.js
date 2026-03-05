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
import {getRules} from "../backend/RuleBackend";
import * as Setting from "../Setting";
import i18next from "i18next";

const {Option} = Select;

class CompoundRule extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      rules: [],
      defaultRules: [
        {
          name: "Start",
          operator: "begin",
          value: "rule1",
        },
        {
          name: "And",
          operator: "and",
          value: "rule2",
        },
      ],
    };
    if (this.props.table.length === 0) {
      this.restore();
    }
  }

  UNSAFE_componentWillMount() {
    this.getRules();
  }

  getRules() {
    getRules(this.props.owner).then((res) => {
      const rules = [];
      for (let i = 0; i < res.data.length; i++) {
        if (Setting.getItemId(res.data[i]) === this.props.owner + "/" + this.props.ruleName) {
          continue;
        }
        rules.push(Setting.getItemId(res.data[i]));
      }
      this.setState({
        rules: rules,
      });
    });
  }

  updateTable(table) {
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {name: `New Item - ${table.length}`, operator: "and", value: ""};
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

  restore() {
    this.updateTable(this.state.defaultRules);
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("rule:Logic"),
        dataIndex: "operator",
        key: "operator",
        width: "180px",
        render: (text, record, index) => {
          const options = [];
          if (index !== 0) {
            options.push({value: "and", text: i18next.t("rule:and")});
            options.push({value: "or", text: i18next.t("rule:or")});
          } else {
            options.push({value: "begin", text: i18next.t("rule:begin")});
          }
          return (
            <Select value={text} virtual={false} style={{width: "100%"}} onChange={value => {
              this.updateField(table, index, "operator", value);
            }}>
              {
                options.map((item, index) => <Option key={index} value={item.value}>{item.text}</Option>)
              }
            </Select>
          );
        },
      },
      {
        title: i18next.t("rule:Rule"),
        dataIndex: "value",
        key: "value",
        render: (text, record, index) => (
          <Select value={text} virtual={false} style={{width: "100%"}} onChange={value => {
            this.updateField(table, index, "value", value);
          }}>
            {
              this.state.rules.map((item, index) => <Option key={index} value={item}>{item}</Option>)
            }
          </Select>
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

export default CompoundRule;
