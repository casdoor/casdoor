// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
import {Button, Col, Input, Popover, Row, Select, Switch, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import Editor from "../common/Editor";

const EmailCss = ".signup-email{}\n.signup-email-input{}\n.signup-email-code{}\n.signup-email-code-input{}\n";
const PhoneCss = ".signup-phone{}\n.signup-phone-input{}\n.phone-code{}\n.signup-phone-code-input{}";

export const SignupTableDefaultCssMap = {
  "Username": ".signup-username {}\n.signup-username-input {}",
  "Display name": ".signup-first-name {}\n.signup-first-name-input{}\n.signup-last-name{}\n.signup-last-name-input{}\n.signup-name{}\n.signup-name-input{}",
  "Affiliation": ".signup-affiliation{}\n.signup-affiliation-input{}",
  "Country/Region": ".signup-country-region{}\n.signup-region-select{}",
  "ID card": ".signup-idcard{}\n.signup-idcard-input{}",
  "Password": ".signup-password{}\n.signup-password-input{}",
  "Confirm password": ".signup-confirm{}",
  "Email": EmailCss,
  "Phone": PhoneCss,
  "Email or Phone": EmailCss + PhoneCss,
  "Phone or Email": EmailCss + PhoneCss,
  "Invitation code": ".signup-invitation-code{}\n.signup-invitation-code-input{}",
  "Agreement": ".login-agreement{}",
  "Signup button": ".signup-button{}\n.signup-link{}",
  "Providers": ".provider-img {\n width: 30px;\n margin: 5px;\n }\n .provider-big-img {\n margin-bottom: 10px;\n }\n ",
};

const {Option} = Select;

class SignupTable extends React.Component {
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
    const row = {name: Setting.getNewRowNameForTable(table, "Please select a signup item"), visible: true, required: true, options: [], rule: "None", customCss: ""};
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
          const items = [
            {name: "Username", displayName: i18next.t("signup:Username")},
            {name: "ID", displayName: i18next.t("general:ID")},
            {name: "Display name", displayName: i18next.t("general:Display name")},
            {name: "Affiliation", displayName: i18next.t("user:Affiliation")},
            {name: "Gender", displayName: i18next.t("user:Gender")},
            {name: "Bio", displayName: i18next.t("user:Bio")},
            {name: "Tag", displayName: i18next.t("user:Tag")},
            {name: "Education", displayName: i18next.t("user:Education")},
            {name: "Country/Region", displayName: i18next.t("user:Country/Region")},
            {name: "ID card", displayName: i18next.t("user:ID card")},
            {name: "Password", displayName: i18next.t("general:Password")},
            {name: "Confirm password", displayName: i18next.t("signup:Confirm")},
            {name: "Email", displayName: i18next.t("general:Email")},
            {name: "Phone", displayName: i18next.t("general:Phone")},
            {name: "Email or Phone", displayName: i18next.t("general:Email or Phone")},
            {name: "Phone or Email", displayName: i18next.t("general:Phone or Email")},
            {name: "Invitation code", displayName: i18next.t("application:Invitation code")},
            {name: "Agreement", displayName: i18next.t("signup:Agreement")},
            {name: "Signup button", displayName: i18next.t("signup:Signup button")},
            {name: "Providers", displayName: i18next.t("general:Providers")},
            {name: "Text 1", displayName: i18next.t("signup:Text 1")},
            {name: "Text 2", displayName: i18next.t("signup:Text 2")},
            {name: "Text 3", displayName: i18next.t("signup:Text 3")},
            {name: "Text 4", displayName: i18next.t("signup:Text 4")},
            {name: "Text 5", displayName: i18next.t("signup:Text 5")},
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
                this.updateField(table, index, "customCss", SignupTableDefaultCssMap[value]);
              }} >
              {
                Setting.getDeduplicatedArray(items, table, "name").map((item, index) => <Option key={index} value={item.name}>{item.displayName}</Option>)
              }
            </Select>
          );
        },
      },
      {
        title: i18next.t("organization:Visible"),
        dataIndex: "visible",
        key: "visible",
        width: "80px",
        render: (text, record, index) => {
          if (record.name === "ID") {
            return null;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "visible", checked);
              if (!checked) {
                this.updateField(table, index, "required", false);
              } else {
                this.updateField(table, index, "required", true);
              }
            }} />
          );
        },
      },
      {
        title: i18next.t("provider:Required"),
        dataIndex: "required",
        key: "required",
        width: "80px",
        render: (text, record, index) => {
          if (!record.visible || ["Signup button", "Providers"].includes(record.name)) {
            return null;
          }

          return (
            <Switch checked={text} disabled={record.name === "Password"} onChange={checked => {
              this.updateField(table, index, "required", checked);
            }} />
          );
        },
      },
      {
        title: i18next.t("provider:Prompted"),
        dataIndex: "prompted",
        key: "prompted",
        width: "80px",
        render: (text, record, index) => {
          if (["ID", "Signup button", "Providers"].includes(record.name)) {
            return null;
          }

          if (record.visible && record.name !== "Country/Region") {
            return null;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "prompted", checked);
            }} />
          );
        },
      },
      {
        title: i18next.t("provider:Type"),
        dataIndex: "type",
        key: "type",
        width: "160px",
        render: (text, record, index) => {
          const options = [
            {id: "Input", name: i18next.t("application:Input")},
            {id: "Single Choice", name: i18next.t("application:Single Choice")},
            {id: "Multiple Choices", name: i18next.t("application:Multiple Choices")},
          ];

          return (
            <Select virtual={false} style={{width: "100%"}} value={text} onChange={(value => {
              this.updateField(table, index, "type", value);
            })} options={options.map(item => Setting.getOption(item.name, item.id))} />
          );
        },
      },
      {
        title: i18next.t("signup:Label"),
        dataIndex: "label",
        key: "label",
        width: "150px",
        render: (text, record, index) => {
          if (record.name.startsWith("Text ")) {
            return (
              <Popover placement="right" content={
                <div style={{width: "900px", height: "300px"}} >
                  <Editor value={text} lang="html" fillHeight dark onChange={value => {
                    this.updateField(table, index, "label", value);
                  }} />
                </div>
              } title={i18next.t("signup:Label HTML")} trigger="click">
                <Input value={text} style={{marginBottom: "10px"}} onChange={e => {
                  this.updateField(table, index, "label", e.target.value);
                }} />
              </Popover>
            );
          }

          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "label", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("application:Custom CSS"),
        dataIndex: "customCss",
        key: "customCss",
        width: "180px",
        render: (text, record, index) => {
          return (
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}}>
                <Editor
                  value={text ? text : SignupTableDefaultCssMap[record.name]}
                  lang="css"
                  fillHeight
                  dark
                  onChange={value => {
                    this.updateField(table, index, "customCss", value ? value : SignupTableDefaultCssMap[record.name]);
                  }}
                />
              </div>
            } title={i18next.t("application:CSS style")} trigger="click">
              <Input value={text ? text : SignupTableDefaultCssMap[record.name]} onChange={e => {
                this.updateField(table, index, "customCss", e.target.value ? e.target.value : SignupTableDefaultCssMap[record.name]);
              }} />
            </Popover>
          );
        },
      },
      {
        title: i18next.t("signup:Placeholder"),
        dataIndex: "placeholder",
        key: "placeholder",
        width: "110px",
        render: (text, record, index) => {
          if (record.name.startsWith("Text ")) {
            return null;
          }

          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "placeholder", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("signup:Options"),
        dataIndex: "options",
        key: "options",
        width: "180px",
        render: (text, record, index) => {
          if (record.type !== "Single Choice" && record.type !== "Multiple Choices") {
            return null;
          }

          return (
            <Select virtual={false} mode="tags" style={{width: "100%"}} value={text}
              onChange={(value => {
                this.updateField(table, index, "options", value);
              })}
              options={text?.map((option) => Setting.getOption(option, option))}
            />
          );
        },
      },
      {
        title: i18next.t("signup:Regex"),
        dataIndex: "regex",
        key: "regex",
        width: "180px",
        render: (text, record, index) => {
          if (record.name.startsWith("Text ") || ["Password", "Confirm password", "Signup button", "Provider"].includes(record.name)) {
            return null;
          }

          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "regex", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("application:Rule"),
        dataIndex: "rule",
        key: "rule",
        width: "155px",
        render: (text, record, index) => {
          let options = [];
          if (record.name === "ID") {
            options = [
              {id: "Random", name: i18next.t("application:Random")},
              {id: "Incremental", name: i18next.t("application:Incremental")},
            ];
          } else if (record.name === "Display name") {
            options = [
              {id: "None", name: i18next.t("general:None")},
              {id: "Real name", name: i18next.t("application:Real name")},
              {id: "First, last", name: i18next.t("application:First, last")},
            ];
          } else if (record.name === "Email") {
            options = [
              {id: "Normal", name: i18next.t("application:Normal")},
              {id: "No verification", name: i18next.t("application:No verification")},
            ];
          } else if (record.name === "Phone") {
            options = [
              {id: "Normal", name: i18next.t("application:Normal")},
              {id: "No verification", name: i18next.t("application:No verification")},
            ];
          } else if (record.name === "Agreement") {
            options = [
              {id: "None", name: i18next.t("application:Only signup")},
              {id: "Signin", name: i18next.t("application:Signin")},
              {id: "Signin (Default True)", name: i18next.t("application:Signin (Default True)")},
            ];
          } else if (record.name === "Providers") {
            options = [
              {id: "big", name: i18next.t("application:Big icon")},
              {id: "small", name: i18next.t("application:Small icon")},
            ];
          }

          if (options.length === 0) {
            return null;
          }

          return (
            <Select virtual={false} style={{width: "100%"}} value={text} onChange={(value => {
              this.updateField(table, index, "rule", value);
            })} options={options.map(item => Setting.getOption(item.name, item.id))} />
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
                <Button disabled={record.name === "Signup button"} icon={<DeleteOutlined />} size="small" onClick={() => this.deleteRow(table, index)} />
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

export default SignupTable;
