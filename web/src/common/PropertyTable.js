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
import {DeleteOutlined, EditOutlined} from "@ant-design/icons";
import {Button, Input, Popconfirm, Table, Tooltip} from "antd";
import * as Setting from "../Setting";
import * as UserBackend from "../backend/UserBackend";
import i18next from "i18next";

class PolicyTable extends React.Component {
  constructor(props) {
    super(props);
    const keys = Object.keys(this.props.properties);
    const properties = [];
    for (let i = 0; i < keys.length; i++) {
      const property = new Object();
      property.index = crypto.randomUUID();
      property.key = keys[i];
      property.value = this.props.properties[keys[i]];
      properties[i] = property;
    }
    this.state = {
      properties: properties,
      loading: false,
      editingIndex: "",
      oldProperty: "",
      add: false,
      page: 1,
    };
  }

  getTrueIndex = (index) => {
    const pageSize = 10;
    index = (this.state.page - 1) * pageSize + index;
    return index;
  };

  isEditing = (index) => {
    return index === this.state.editingIndex;
  };

  edit = (record, index) => {
    this.setState({editingIndex: index, oldProperty: Setting.deepCopy(record)});
  };

  cancel = (table, index) => {
    index = this.getTrueIndex(index);
    Object.keys(table[index]).forEach((key) => {
      table[index][key] = this.state.oldProperty[key];
    });
    this.updateTable(table);
    this.setState({editingIndex: "", oldProperty: ""});
    if (this.state.add) {
      this.deleteRow(this.state.properties, index);
      this.setState({add: false});
    }
  };

  updateTable(table) {
    this.setState({properties: table});
  }

  updateField(table, index, key, value) {
    index = this.getTrueIndex(index);
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {index: crypto.randomUUID()};
    if (table === undefined) {
      table = [];
    }
    this.setState({page: 1});
    table = Setting.addRow(table, row, "top");
    this.updateTable(table);
    this.edit(row, 0);
    this.setState({add: true});
  }

  deleteRow(table, i) {
    table = Setting.deleteRow(table, i);
    this.updateTable(table);
  }

  save(table, i) {
    i = this.getTrueIndex(i);
    this.state.add ? this.addPropety(table, i) : this.updateProperty(table, i);
  }

  updateProperty(table, i) {
    this.updateUser(table, "update", i);
  }

  addPropety(table, i) {
    this.updateUser(table, "update", i);
  }

  deleteProperty(table, i) {
    i = this.getTrueIndex(i);
    this.updateUser(table, "delete", i);
    table = Setting.deleteRow(table, i);
    this.updateTable(table);
  }

  updateUser(table, method, i) {
    const user = this.props.user;
    if (method === "delete") {
      delete user.properties[table[i].key];
    }
    if (method === "update") {
      user.properties[table[i].key] = table[i].value;
    }
    UserBackend.updateUser(user.owner, user.name, user).then(res => {
      if (res.status === "ok") {
        this.setState({editingIndex: "", oldPropertyKey: "", oldPropertyValue: "", add: false});
        if (res.data !== "Affected") {
          Setting.showMessage("info", "Repeated property");
        } else {
          Setting.showMessage("success", "Success");
        }
      } else {
        Setting.showMessage("error", "Failed to add: ${res.msg}");
      }
    });
  }

  renderTable(table) {
    const columns = [
      {
        title: "Keys",
        dataIndex: "key",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing && this.state.add ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "key", e.target.value);
              }} />
              : text
          );
        },
      },
      {
        title: "Values",
        dataIndex: "value",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "value", e.target.value);
              }} />
              : text
          );
        },
      },
      {
        title: "Option",
        key: "option",
        width: "100px",
        render: (text, record, index) => {
          const editable = this.isEditing(index);
          return editable ? (
            <span>
              <Button style={{marginRight: 8}} onClick={() => this.save(table, index)}>
                {i18next.t("general:Save")}
              </Button>
              <Popconfirm title={i18next.t("general:Cancel")} onConfirm={() => this.cancel(table, index)}>
                <a>{i18next.t("general:Cancel")}</a>
              </Popconfirm>
            </span>
          ) : (
            <div>
              <Tooltip placement="topLeft" title={i18next.t("general:Edit")}>
                <Button disabled={this.state.editingIndex !== ""} style={{marginRight: "5px"}} icon={<EditOutlined />} size="small" onClick={() => this.edit(record, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title={i18next.t("general:Delete")}>
                <Button disabled={this.state.editingIndex !== ""} style={{marginRight: "5px"}} icon={<DeleteOutlined />} size="small" onClick={() => this.deleteProperty(table, index)} />
              </Tooltip>
            </div>
          );
        },
      },
    ];

    return (
      <Table
        pagination={{
          defaultPageSize: 10,
          onChange: (page) => {
            this.setState({page: page});
          },
          current: this.state.page,
        }}
        columns={columns} dataSource={table} rowKey="index" size="middle" bordered
        loading={this.state.loading}
        title={() => (
          <div>
            <Button disabled={this.state.editingIndex !== ""} style={{marginRight: "5px"}} type="primary" size="small" onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
          </div>
        )}
      />
    );
  }

  render() {
    return (<>
      {
        this.renderTable(this.state.properties)
      }
    </>
    );
  }
}

export default PolicyTable;
