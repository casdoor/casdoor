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
import {Button, Col, Input, Row, Select, Switch, Table, Tooltip} from "antd";
import {CountryCodeSelect} from "../common/select/CountryCodeSelect";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as Provider from "../auth/Provider";

const {Option} = Select;

class ProviderTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  getUserOrganization() {
    return this.props.application?.organizationObj;
  }

  updateTable(table) {
    this.props.onUpdateTable(table);
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {name: Setting.getNewRowNameForTable(table, "Please select a provider"), canSignUp: true, canSignIn: true, canUnlink: true, prompted: false, signupGroup: "", rule: "None"};
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
    let columns = [
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
                const provider = Setting.getArrayItem(this.props.providers, "name", value);
                this.updateField(table, index, "provider", provider);

                // If the provider is email or SMS, set the rule to "all" instead of the default "None"
                if (provider.category === "Email" || provider.category === "SMS") {
                  this.updateField(table, index, "rule", "All");
                }
              }} >
              {
                Setting.getDeduplicatedArray(this.props.providers, table, "name").filter(provider => provider.category !== "Captcha" || !table.some(tableItem => {
                  const existingProvider = Setting.getArrayItem(this.props.providers, "name", tableItem.name);
                  return existingProvider && existingProvider.category === "Captcha";
                })).map((provider, index) => <Option key={index} value={provider.name}>{provider.name}</Option>)
              }
            </Select>
          );
        },
      },
      {
        title: i18next.t("provider:Category"),
        dataIndex: "category",
        key: "category",
        width: "100px",
        render: (text, record, index) => {
          const provider = Setting.getArrayItem(this.props.providers, "name", record.name);
          return provider?.category;
        },
      },
      {
        title: i18next.t("provider:Type"),
        dataIndex: "type",
        key: "type",
        width: "80px",
        render: (text, record, index) => {
          const provider = Setting.getArrayItem(this.props.providers, "name", record.name);
          return Provider.getProviderLogoWidget(provider);
        },
      },
      {
        title: i18next.t("user:Country/Region"),
        dataIndex: "countryCodes",
        key: "countryCodes",
        width: "140px",
        render: (text, record, index) => {
          if (record.provider?.category !== "SMS") {
            return null;
          }

          return (
            <CountryCodeSelect
              style={{width: "100%"}}
              hasDefault={true}
              mode={"multiple"}
              initValue={text ? text : ["All"]}
              onChange={(value) => {
                this.updateField(table, index, "countryCodes", value);
              }}
              countryCodes={this.getUserOrganization()?.countryCodes}
            />
          );
        },
      },
      {
        title: i18next.t("provider:Can signup"),
        dataIndex: "canSignUp",
        key: "canSignUp",
        width: "120px",
        render: (text, record, index) => {
          if (!["OAuth", "Web3"].includes(record.provider?.category)) {
            return null;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "canSignUp", checked);
            }} />
          );
        },
      },
      {
        title: i18next.t("provider:Can signin"),
        dataIndex: "canSignIn",
        key: "canSignIn",
        width: "120px",
        render: (text, record, index) => {
          if (!["OAuth", "Web3"].includes(record.provider?.category)) {
            return null;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "canSignIn", checked);
            }} />
          );
        },
      },
      {
        title: i18next.t("provider:Can unlink"),
        dataIndex: "canUnlink",
        key: "canUnlink",
        width: "120px",
        render: (text, record, index) => {
          if (!["OAuth", "Web3"].includes(record.provider?.category)) {
            return null;
          }

          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "canUnlink", checked);
            }} />
          );
        },
      },
      {
        title: i18next.t("provider:Prompted"),
        dataIndex: "prompted",
        key: "prompted",
        width: "120px",
        render: (text, record, index) => {
          if (!["OAuth", "Web3"].includes(record.provider?.category)) {
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
        title: i18next.t("provider:Signup group"),
        dataIndex: "signupGroup",
        key: "signupGroup",
        width: "120px",
        render: (text, record, index) => {
          if (!["OAuth", "Web3"].includes(record.provider?.category)) {
            return null;
          }

          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "signupGroup", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("application:Rule"),
        dataIndex: "rule",
        key: "rule",
        width: "160px",
        render: (text, record, index) => {
          if (record.provider?.type === "Google") {
            if (text === "None") {
              text = "Default";
            }
            return (
              <Select virtual={false} style={{width: "100%"}}
                value={text}
                defaultValue="Default"
                onChange={value => {
                  this.updateField(table, index, "rule", value);
                }} >
                <Option key="Default" value="Default">{i18next.t("general:Default")}</Option>
                <Option key="OneTap" value="OneTap">{"One Tap"}</Option>
              </Select>
            );
          } else if (record.provider?.category === "Captcha") {
            return (
              <Select virtual={false} style={{width: "100%"}}
                value={text}
                defaultValue="None"
                onChange={value => {
                  this.updateField(table, index, "rule", value);
                }} >
                <Option key="None" value="None">{i18next.t("general:None")}</Option>
                <Option key="Dynamic" value="Dynamic">{i18next.t("application:Dynamic")}</Option>
                <Option key="Always" value="Always">{i18next.t("application:Always")}</Option>
                <Option key="Internet-Only" value="Internet-Only">{i18next.t("application:Internet-Only")}</Option>
              </Select>
            );
          } else if (record.provider?.category === "SMS" || record.provider?.category === "Email") {
            if (text === "None") {
              text = "All";
            }
            return (
              <Select virtual={false} style={{width: "100%"}}
                value={text}
                defaultValue="All"
                onChange={value => {
                  this.updateField(table, index, "rule", value);
                }}>
                <Option key="all" value="all">{"All"}</Option>
                <Option key="signup" value="signup">{"Signup"}</Option>
                <Option key="login" value="login">{"Login"}</Option>
                <Option key="forget" value="forget">{"Forget Password"}</Option>
                <Option key="reset" value="reset">{"Reset Password"}</Option>
                <Option key="mfaSetup" value="mfaSetup">{"Set MFA"}</Option>
                <Option key="mfaAuth" value="mfaAuth">{"MFA Auth"}</Option>
              </Select>
            );
          } else {
            return null;
          }
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

    if (!this.props.application.enableSignUp) {
      columns = columns.filter(column => column.key !== "canSignUp");
    }

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

export default ProviderTable;
