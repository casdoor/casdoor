// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
import {Button, Col, Input, Row, Select, Switch, Table, Tooltip} from "antd";
import {DeleteOutlined, DownOutlined, UpOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";
import i18next from "i18next";

class FormItemTable extends React.Component {
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
    const row = {name: "", label: "", visible: false};
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

  defaultTable() {
    let rows = this.getItems();
    if (!Array.isArray(rows)) {
      rows = [rows];
    }
    this.updateTable(rows);
  }

  getItems() {
    const formType = this.props.formType;
    return Setting.getFormTypeItems(formType);
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "200px",
        render: (text, record, index) => {
          const items = this.getItems();
          const options = Setting.getDeduplicatedArray(items, table, "name").map(item => ({label: i18next.t(item.label), value: item.name}));
          const selectedLabel = items.find(item => item.name === text)?.label || text;
          return (
            <Select
              virtual={false}
              style={{width: "100%"}}
              options={options}
              value={i18next.t(selectedLabel)}
              onChange={value => {
                this.updateField(table, index, "name", value);
              }}
              optionLabelProp="label"
            />
          );
        },
      },
      {
        title: i18next.t("signup:Label"),
        dataIndex: "label",
        key: "label",
        width: "200px",
        render: (text, record, index) => {
          const items = this.getItems();
          const selectedItem = items.find(item => item.name === text);
          const currentLabel = selectedItem?.label || text;
          return (
            <Input
              value={i18next.t(currentLabel)}
              onChange={e => {
                const newLabel = e.target.value;
                this.updateField(this.props.table, index, "label", newLabel);
                if (selectedItem) {
                  selectedItem.label = newLabel;
                }
              }}
            />
          );
        },
      },
      {
        title: i18next.t("organization:Visible"),
        dataIndex: "visible",
        key: "visible",
        width: "200px",
        render: (text, record, index) => {
          return (
            <Switch checked={text} onChange={checked => {
              this.updateField(table, index, "visible", checked);
            }} />
          );
        },
      },
      {
        title: i18next.t("form:Width"),
        dataIndex: "width",
        key: "width",
        render: (text, record, index) => {
          return (
            <Input value={text} onChange={e => {
              this.updateField(table, index, "width", e.target.value);
            }} />
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
                <Button style={{marginRight: "5px"}} disabled={index === 0} icon={<UpOutlined />}
                  size="small" onClick={() => this.upRow(table, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Down")}>
                <Button style={{marginRight: "5px"}} disabled={index === table.length - 1}
                  icon={<DownOutlined />} size="small" onClick={() => this.downRow(table, index)} />
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
      <Table scroll={{x: "max-content"}} rowKey="name" columns={columns} dataSource={table} size="middle" bordered
        pagination={false}
        title={() => (
          <div>
            {this.props.title}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button style={{marginRight: "10px"}} size="small" onClick={() => this.defaultTable()}>{i18next.t("general:Reset to Default")}</Button>
            <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
          </div>
        )}
      />
    );
  }

  render() {
    return (
      <div>
        <Row style={{marginTop: "20px"}}>
          <Col span={24}>{this.renderTable(this.props.table)}</Col>
        </Row>
      </div>
    );
  }
}

export default FormItemTable;
