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
import {Button, Col, Row, Select, Table, Tooltip} from "antd";
import {EmailMfaType, SmsMfaType, TotpMfaType} from "../auth/MfaSetupPage";
import * as Setting from "../Setting";
import i18next from "i18next";

const {Option} = Select;

const MfaItems = [
  {name: "Phone", value: SmsMfaType},
  {name: "Email", value: EmailMfaType},
  {name: "App", value: TotpMfaType},
];

const RuleItems = [
  {value: "Optional", label: i18next.t("organization:Optional")},
  {value: "Prompt", label: i18next.t("organization:Prompt")},
  {value: "Required", label: i18next.t("organization:Required")},
];

class MfaTable extends React.Component {
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
    const row = {name: Setting.getNewRowNameForTable(table, "Please select a MFA method"), rule: "Optional"};
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
          return (
            <Select virtual={false} style={{width: "100%"}}
              value={text}
              onChange={value => {
                this.updateField(table, index, "name", value);
              }} >
              {
                Setting.getDeduplicatedArray(MfaItems, table, "name").map((item, index) => <Option key={index} value={item.value}>{item.name}</Option>)
              }
            </Select>
          );
        },
      },
      {
        title: i18next.t("application:Rule"),
        dataIndex: "rule",
        key: "rule",
        width: "100px",
        render: (text, record, index) => {
          return (
            <Select virtual={false} style={{width: "100%"}}
              value={text}
              defaultValue="Optional"
              options={RuleItems.map((item) =>
                Setting.getOption(item.label, item.value))
              }
              onChange={value => {
                let requiredCount = 0;
                table.forEach((item) => {
                  if (item.rule === "Required") {
                    requiredCount++;
                  }
                });
                // eslint-disable-next-line no-console
                console.log(requiredCount);
                if (value === "Required" && requiredCount >= 1) {
                  Setting.showMessage("error", "Only 1 MFA methods can be required");
                  return;
                }
                this.updateField(table, index, "rule", value);
              }} >
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
            <Button disabled={table.length >= MfaItems.length} style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
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

export default MfaTable;
