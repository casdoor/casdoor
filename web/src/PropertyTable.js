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
import * as Setting from "./Setting";
import * as UserBackend from "./backend/UserBackend";
import i18next from "i18next";

class PropertyTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      propertyKeys: Object.keys(this.props.properties),
      propertyValues: Object.values(this.props.properties),
      loading: false,
      editingIndex: "",
      oldPropertyKey: "",
      oldPropertyValue: "",
      add: false,
    };
  }

  isEditing = (index) => {
    return index === this.state.editingIndex;
  };

  edit = (record, index) => {
    this.setState({editingIndex: index, oldPropertyKey: this.state.propertyKeys[index], oldPropertyValue: this.state.propertyValues[index]});
  };

  cancel = (table, index) => {
    this.state.propertyKeys[index]=this.state.oldPropertyKey
    this.state.propertyValues[index]=this.state.oldPropertyValue
    this.updateTable(this.state.propertyKeys,this.state.propertyValues)
    this.setState({editingIndex: "", oldPropertyKey: "", oldPropertyValue: ""});
    if(this.state.add){
      this.deleteRow(this.state.propertyKeys,this.state.propertyValues,index)
      this.setState({add: false})
    }
  };

  updateTable(keys,values) {
    this.setState({propertyKeys: keys,propertyValues: values});
  }

  updateField(table, index, key, value) {
    if(key=="Keys"){
      this.state.propertyKeys[index]=value;
    }
    if(key=="Values"){
      this.state.propertyValues[index]=value;
    }
    this.updateTable(this.state.propertyKeys,this.state.propertyValues)
  }

  addRow(table) {
    const row = {};
    if (table === undefined) {
      table = [];
    }
    table = Setting.addRow(table, row, "top");
    this.state.propertyValues = Setting.addRow(this.state.propertyValues, row, "top");
    this.updateTable(table, this.state.propertyValues);
    this.edit(row, 0); 
    this.setState({add: true});
  }

  deleteRow(keys, values, i) {
    keys = Setting.deleteRow(keys, i);
    values = Setting.deleteRow(values, i);
    this.updateTable(keys,values);
  }

  save(table, i) {
    this.state.add ? this.addProperty(table, i) : this.updateProperty(table, i);
  }

  updateProperty(table, i) {
    var user = this.props.user;
    var key = this.state.propertyKeys[i];
    var value = this.state.propertyValues[i];
    user.properties[key]=value
    this.updateUser(user)
  }

  addProperty(table, i) {
    var user = this.props.user;
    var key = this.state.propertyKeys[i];
    var value = this.state.propertyValues[i];
    user.properties[key]=value
    this.updateUser(user)
  }

  deleteProperty(table, i) {
    var user = this.props.user;
    var key = this.state.propertyKeys;
    delete user.properties[key[i]]
    this.updateUser(user)
    this.updateTable(Object.keys(user.properties),Object.values(user.properties))
  }
  
  updateUser(user){
    UserBackend.updateUser(user.owner, user.name, user).then(res=>{
      if(res.status === "ok"){
        this.setState({editingIndex: "", oldPropertyKey: "", oldPropertyValue: "" ,add: false});
        if(res.data !== "Affected"){
          Setting.showMessage("info", "Repeated property");
        }else{
          Setting.showMessage("success", "Success");
        } 
      }else{
        Setting.showMessage("error", "Failed to add: ${res.msg}")
      }
    });
  }

  renderTable(table) {
    const columns = [
      {
        title: "Key",
        dataIndex: "key",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing&&this.state.add ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "Keys", e.target.value);
              }} />
            : table[index]
          );
        },
      },
      {
        title: "Value",
        dataIndex: "value",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "Values", e.target.value);
              }} />
              : this.state.propertyValues[index]
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
              Save
              </Button>
              <Popconfirm title="Sure to cancel?" onConfirm={() => this.cancel(table, index)}>
                <a>Cancel</a>
              </Popconfirm>
            </span>
          ) : (
            <div>
              <Tooltip placement="topLeft" title="Edit">
                <Button disabled={this.state.editingIndex !== ""} style={{marginRight: "5px"}} icon={<EditOutlined />} size="small" onClick={() => this.edit(record, index)} />
              </Tooltip>
              <Tooltip placement="topLeft" title="Delete">
                <Button disabled={this.state.editingIndex !== ""} style={{marginRight: "5px"}} icon={<DeleteOutlined />} size="small" onClick={() => this.deleteProperty(table, index)} />
              </Tooltip>
            </div>
          );
        },
      }];

    return (
      <Table
        pagination={{
          defaultPageSize: 10,
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
        this.renderTable(this.state.propertyKeys)
      }
    </>
    );
  }
}

export default PropertyTable;
