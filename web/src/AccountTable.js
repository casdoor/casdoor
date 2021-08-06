// Copyright 2021 The casbin Authors. All Rights Reserved.
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
    let row = {name: "Please select a account item", visible: true, required: true, public: true};
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
        title: i18next.t("organization:Name"),
        dataIndex: "name",
        key: "name",
        render: (text, record, index) => {
          const items = [
            {id: "Organization", name: "Organization"},
            {id: "ID", name: "ID"},
            {id: "Name", name: "Name"},
            {id: "Display name", name: "Display name"},
            {id: "Avatar", name: "Avatar"},
            {id: "User type", name: "User type"},
            {id: "Password", name: "Password"},
            {id: "Email", name: "Email"},
            {id: "Phone", name: "Phone"},
            {id: "Affiliation", name: "Affiliation"},
            {id: "Tag", name: "Tag"},
            {id: "3rd-party logins", name: "3rd-party logins"},
          ];

          return (
            <Select virtual={false} style={{width: "100%"}}
                    value={text}
                    onChange={value => {
                      this.updateField(table, index, "name", value);
                    }}>
              {
                Setting.getDeduplicatedArray(items, table, "name").map((item, index) => <Option
                  key={index} value={item.name}>{item.name}</Option>)
              }
            </Select>
          )
        }
      },
      {
        title: i18next.t("organization:visible"),
        dataIndex: "visible",
        key: "visible",
        width: "120px",
        render: (text, record, index) => {
          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "visible", checked);
              if (!checked) {
                this.updateField(table, index, "required", false);
              } else {
                this.updateField(table, index, "required", true);
              }
            }}/>
          )
        }
      },
      {
        title: i18next.t("organization:required"),
        dataIndex: "required",
        key: "required",
        width: "120px",
        render: (text, record, index) => {
          if (!record.visible || !record.editable) {
            return null;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "required", checked);
            }}/>
          )
        }
      },
      {
        title: i18next.t("organization:editable"),
        dataIndex: "editable",
        key: "editable",
        width: "120px",
        render: (text, record, index) => {
          if (record.name === "ID" || record.name === "Organization" || record.name === "Name") {
            return null;
          }

          if (!record.visible) {
            return null;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "editable", checked);
              if (!checked) {
                this.updateField(table, index, "required", false);
              } else {
                this.updateField(table, index, "required", true);
              }
            }}/>
          )
        }
      },
      {
        title: i18next.t("organization:public"),
        dataIndex: "public",
        key: "public",
        width: "120px",
        render: (text, record, index) => {
          if (!record.visible) {
            return false;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "public", checked);
            }}/>
          )
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
                <Button style={{marginRight: "5px"}} disabled={index === 0} icon={<UpOutlined/>}
                        size="small" onClick={() => this.upRow(table, index)}/>
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Down")}>
                <Button style={{marginRight: "5px"}} disabled={index === table.length - 1}
                        icon={<DownOutlined/>} size="small" onClick={() => this.downRow(table, index)}/>
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Delete")}>
                <Button icon={<DeleteOutlined/>} size="small"
                        onClick={() => this.deleteRow(table, index)}/>
              </Tooltip>
            </div>
          );
        }
      },
    ];

    return (
      <Table scroll={{x: "max-content"}} rowKey="name" columns={columns} dataSource={table} size="middle" bordered
             pagination={false}
             title={() => (
               <div>
                 {this.props.title}&nbsp;&nbsp;&nbsp;&nbsp;
                 <Button style={{marginRight: "5px"}} type="primary" size="small"
                         onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
               </div>
             )}
      />
    );
  }

  render() {
    if (this.props.table === null) {
      this.updateTable([])
    }
    return (
      <div>
        <Row style={{marginTop: "20px"}}>
          <Col span={24}>
            {
              this.props.table === null ? null : this.renderTable(this.props.table)
            }
          </Col>
        </Row>
      </div>
    )
  }
}

export default AccountTable;
