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
import {AutoComplete, Button, Col, Input, Row, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

const DefaultScopes = [
  {scope: "openid", displayName: "OpenID", description: "Authenticate the user and obtain an ID token"},
  {scope: "profile", displayName: "Profile", description: "Read all user profile data"},
  {scope: "email", displayName: "Email", description: "Access user email addresses (read-only)"},
  {scope: "address", displayName: "Address", description: "Access the user's address information"},
  {scope: "phone", displayName: "Phone", description: "Access the user's phone number information"},
  {scope: "offline_access", displayName: "Offline Access", description: "Obtain refresh tokens for offline access"},
];

class CustomScopeTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  normalizeScope(scope) {
    return (scope || "").trim().toLowerCase();
  }

  getAvailableDefaultScopes(table) {
    const existingScopes = new Set((table || []).map(item => this.normalizeScope(item?.scope)).filter(Boolean));
    return DefaultScopes.filter(item => !existingScopes.has(this.normalizeScope(item.scope)));
  }

  updateTable(table) {
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  isScopeMissing(row) {
    if (!row) {
      return true;
    }
    const scope = (row.scope || "").trim();
    return scope === "";
  }

  addRow(table) {
    const row = {scope: "", displayName: "", description: ""};
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
    table = table || [];

    const columns = [
      {
        title: (
          <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
            <span className="ant-form-item-required">{i18next.t("general:Name")}</span>
            <div style={{color: "red"}}>*</div>
          </div>
        ),
        dataIndex: "scope",
        key: "scope",
        width: "260px",
        render: (text, record, index) => {
          const availableDefaultScopes = this.getAvailableDefaultScopes(table);
          const autoCompleteOptions = availableDefaultScopes.map(item => ({
            label: `${item.scope}`,
            value: item.scope,
          }));

          return (
            <AutoComplete
              status={this.isScopeMissing(record) ? "error" : ""}
              value={text}
              options={autoCompleteOptions}
              placeholder="Select or input scope"
              onSelect={(value) => {
                this.updateField(table, index, "scope", value);
                const selectedScope = availableDefaultScopes.find(item => item.scope === value);
                if (selectedScope) {
                  this.updateField(table, index, "displayName", selectedScope.displayName);
                  this.updateField(table, index, "description", selectedScope.description);
                }
              }}
              onChange={(value) => {
                this.updateField(table, index, "scope", value);
              }}
            >
              <Input />
            </AutoComplete>
          );
        },
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "200px",
        render: (text, _, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "displayName", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("general:Description"),
        dataIndex: "description",
        key: "description",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "description", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "action",
        key: "action",
        width: "110px",
        // eslint-disable-next-line
        render: (_, __, index) => {
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
        <div style={{display: "flex", justifyContent: "space-between"}}>
          <div style={{marginTop: "5px"}}>{this.props.title}</div>
          <Button type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
        </div>
      )}
      columns={columns} dataSource={table} rowKey={(record, index) => record.scope?.trim() || `temp_${index}`} size="middle" bordered pagination={false}
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

export default CustomScopeTable;
