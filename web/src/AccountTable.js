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
import {Button, Col, Row, Select, Switch, Table, Tooltip} from "antd";
import * as Setting from "./Setting";
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
    let row = {name: Setting.getNewRowNameForTable(table, "Please select an account item"), visible: true};
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
        title: i18next.t("provider:Name"),
        dataIndex: "name",
        key: "name",
        render: (text, record, index) => {
          const items = [
            {name: "Organization", displayName: i18next.t("general:Organization")},
            {name: "ID", displayName: i18next.t("general:ID")},
            {name: "Name", displayName: i18next.t("general:Name")},
            {name: "Display name", displayName: i18next.t("general:Display name")},
            {name: "Avatar", displayName: i18next.t("general:Avatar")},
            {name: "User type", displayName: i18next.t("general:User type")},
            {name: "Password", displayName: i18next.t("general:Password")},
            {name: "Email", displayName: i18next.t("general:Email")},
            {name: "Phone", displayName: i18next.t("general:Phone")},
            {name: "Country/Region", displayName: i18next.t("user:Country/Region")},
            {name: "Location", displayName: i18next.t("user:Location")},
            {name: "Affiliation", displayName: i18next.t("user:Affiliation")},
            {name: "Title", displayName: i18next.t("user:Title")},
            {name: "Homepage", displayName: i18next.t("user:Homepage")},
            {name: "Bio", displayName: i18next.t("user:Bio")},
            {name: "Tag", displayName: i18next.t("user:Tag")},
            {name: "Signup application", displayName: i18next.t("general:Signup application")},
            {name: "Roles", displayName: i18next.t("general:Roles")},
            {name: "Permissions", displayName: i18next.t("general:Permissions")},
            {name: "3rd-party logins", displayName: i18next.t("user:3rd-party logins")},
            {name: "Properties", displayName: i18next.t("user:Properties")},
            {name: "Is admin", displayName: i18next.t("user:Is admin")},
            {name: "Is global admin", displayName: i18next.t("user:Is global admin")},
            {name: "Is forbidden", displayName: i18next.t("user:Is forbidden")},
            {name: "Is deleted", displayName: i18next.t("user:Is deleted")},
            {name: "WebAuthn credentials", displayName: i18next.t("user:WebAuthn credentials")},
          ];

          const getItemDisplayName = (text) => {
            const item = items.filter(item => item.name === text);
            if (item.length === 0) {
              return "";
            }
            return item[0].displayName;
          };

          return (
            <Select virtual={false} style={{width: "100%"}}
              value={getItemDisplayName(text)}
              onChange={value => {
                this.updateField(table, index, "name", value);
              }} >
              {
                Setting.getDeduplicatedArray(items, table, "name").map((item, index) => <Option key={index} value={item.name}>{item.displayName}</Option>)
              }
            </Select>
          );
        }
      },
      {
        title: i18next.t("provider:visible"),
        dataIndex: "visible",
        key: "visible",
        width: "120px",
        render: (text, record, index) => {
          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "visible", checked);
            }} />
          );
        }
      },
      {
        title: i18next.t("organization:viewRule"),
        dataIndex: "viewRule",
        key: "viewRule",
        width: "155px",
        render: (text, record, index) => {
          if (!record.visible) {
            return null;
          }

          let options = [
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
        }
      },
      {
        title: i18next.t("organization:modifyRule"),
        dataIndex: "modifyRule",
        key: "modifyRule",
        width: "155px",
        render: (text, record, index) => {
          if (!record.visible) {
            return null;
          }

          let options;
          if (record.viewRule === "Admin") {
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
        }
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
        }
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
