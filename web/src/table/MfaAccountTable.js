// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
import {Button, Col, Image, Input, Popover, Row, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import {CasdoorAppQrCode, CasdoorAppUrl} from "../common/CasdoorAppConnector";

class MfaAccountTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      icon: this.props.icon,
      mfaAccounts: this.props.table !== null ? this.props.table.map((item, index) => {
        item.key = index;
        return item;
      }) : [],
    };
  }

  count = this.props.table?.length ?? 0;

  updateTable(table) {
    this.setState({
      mfaAccounts: table,
    });

    this.props.onUpdateTable([...table].map((item) => {
      const newItem = Setting.deepCopy(item);
      delete newItem.key;
      return newItem;
    }));
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {key: this.count, accountName: "", issuer: "", secretKey: ""};
    if (table === undefined || table === null) {
      table = [];
    }

    this.count += 1;
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
        title: i18next.t("forget:Account"),
        dataIndex: "accountName",
        key: "accountName",
        width: "400px",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "accountName", e.target.value);
            }} />
          );
        },
      },
      {
        title: "Issuer",
        dataIndex: "issuer",
        key: "issuer",
        width: "300px",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "issuer", e.target.value);
            }} />
          );
        },
      },
      {
        title: "Origin",
        dataIndex: "origin",
        key: "origin",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "origin", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("provider:Secret key"),
        dataIndex: "secretKey",
        key: "secretKey",
        render: (text, record, index) => {
          return (
            <Input.Password value={text} onChange={e => {
              this.updateField(table, index, "secretKey", e.target.value);
            }} />
          );
        },
      },
      {
        title: i18next.t("general:Logo"),
        dataIndex: "issuer",
        key: "logo",
        width: "60px",
        render: (text, record, index) => (
          <Tooltip>
            {text ? (
              <Image width={36} height={36} preview={false} src={`${Setting.StaticBaseUrl}/img/social_${text.toLowerCase()}.png`}
                fallback={`${Setting.StaticBaseUrl}/img/social_default.png`} alt={text} />
            ) : (
              <Image width={36} height={36} preview={false} src={`${Setting.StaticBaseUrl}/img/social_default.png`} alt="default" />
            )}
          </Tooltip>
        ),
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
      <Table scroll={{x: "max-content"}} rowKey="key" columns={columns} dataSource={table} size="middle" bordered pagination={false}
        title={() => (
          <div>
            {this.props.title}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button style={{marginRight: "10px"}} type="primary" size="small" onClick={() => this.addRow(table)}>
              {i18next.t("general:Add")}
            </Button>
            <Popover
              trigger="focus"
              overlayInnerStyle={{padding: 0}}
              content={<CasdoorAppQrCode accessToken={this.props.accessToken} icon={this.state.icon} />}
            >
              <Button style={{marginRight: "10px"}} type="primary" size="small">
                {i18next.t("general:QR Code")}
              </Button>
            </Popover>
            <Popover
              trigger="click"
              content={<CasdoorAppUrl accessToken={this.props.accessToken} />}
            >
              <Button type="primary" size="small">
                {i18next.t("general:URL")}
              </Button>
            </Popover>
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
              this.renderTable(this.state.mfaAccounts)
            }
          </Col>
        </Row>
      </div>
    );
  }
}

export default MfaAccountTable;
