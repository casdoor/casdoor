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
import * as AdapterBackend from "../backend/AdapterBackend";
import i18next from "i18next";

class PolicyTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      policyLists: [],
      loading: false,
      editingIndex: "",
      oldPolicy: "",
      add: false,
    };
  }

  UNSAFE_componentWillMount() {
    if (this.props.mode === "edit") {
      this.synPolicies();
    }
  }

  isEditing = (index) => {
    return index === this.state.editingIndex;
  };

  edit = (record, index) => {
    this.setState({editingIndex: index, oldPolicy: Setting.deepCopy(record)});
  };

  cancel = (table, index) => {
    Object.keys(table[index]).forEach((key) => {
      table[index][key] = this.state.oldPolicy[key];
    });
    this.updateTable(table);
    this.setState({editingIndex: "", oldPolicy: ""});
    if (this.state.add) {
      this.deleteRow(this.state.policyLists, index);
      this.setState({add: false});
    }
  };

  updateTable(table) {
    this.setState({policyLists: table});
  }

  updateField(table, index, key, value) {
    table[index][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {Ptype: "p"};
    if (table === undefined) {
      table = [];
    }
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
    this.state.add ? this.addPolicy(table, i) : this.updatePolicy(table, i);
  }

  synPolicies() {
    this.setState({loading: true});
    AdapterBackend.syncPolicies(this.props.owner, this.props.name)
      .then((res) => {
        if (res.status !== "error") {
          this.setState({loading: false, policyLists: res});
        } else {
          this.setState({loading: false});
          Setting.showMessage("error", `Adapter failed to get policies, ${res.msg}`);
        }
      })
      .catch(error => {
        this.setState({loading: false});
        Setting.showMessage("error", `Adapter failed to get policies, ${error}`);
      });
  }

  updatePolicy(table, i) {
    AdapterBackend.UpdatePolicy(this.props.owner, this.props.name, [this.state.oldPolicy, table[i]]).then(res => {
      if (res.status === "ok") {
        this.setState({editingIndex: "", oldPolicy: ""});
        Setting.showMessage("success", i18next.t("adapter:Update policy successfully"));
      } else {
        Setting.showMessage("error", i18next.t(`adapter:Update policy failed, ${res.msg}`));
      }
    });
  }

  addPolicy(table, i) {
    AdapterBackend.AddPolicy(this.props.owner, this.props.name, table[i]).then(res => {
      if (res.status === "ok") {
        this.setState({editingIndex: "", oldPolicy: "", add: false});
        if (res.data !== "Affected") {
          Setting.showMessage("info", i18next.t("adapter:Repeated policy"));
        } else {
          Setting.showMessage("success", i18next.t("adapter:Add policy successfully"));
        }
      } else {
        Setting.showMessage("error", i18next.t(`adapter:Add policy failed, ${res.msg}`));
      }
    });
  }

  deletePolicy(table, i) {
    AdapterBackend.RemovePolicy(this.props.owner, this.props.name, table[i]).then(res => {
      if (res.status === "ok") {
        table = Setting.deleteRow(table, i);
        this.updateTable(table);
        Setting.showMessage("success", i18next.t("adapter:Delete policy successfully"));
      } else {
        Setting.showMessage("error", i18next.t(`adapter:Delete policy failed, ${res.msg}`));
      }
    });
  }

  renderTable(table) {
    const columns = [
      {
        title: "Rule Type",
        dataIndex: "Ptype",
        width: "100px",
        // render: (text, record, index) => {
        //   const editing = this.isEditing(index);
        //   return (
        //     editing ?
        //       <Input value={text} onChange={e => {
        //         this.updateField(table, index, "Ptype", e.target.value);
        //       }} />
        //       : text
        //   );
        // },
      },
      {
        title: "V0",
        dataIndex: "V0",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "V0", e.target.value);
              }} />
              : text
          );
        },
      },
      {
        title: "V1",
        dataIndex: "V1",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "V1", e.target.value);
              }} />
              : text
          );
        },
      },
      {
        title: "V2",
        dataIndex: "V2",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "V2", e.target.value);
              }} />
              : text
          );
        },
      },
      {
        title: "V3",
        dataIndex: "V3",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "V3", e.target.value);
              }} />
              : text
          );
        },
      },
      {
        title: "V4",
        dataIndex: "V4",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "V4", e.target.value);
              }} />
              : text
          );
        },
      },
      {
        title: "V5",
        dataIndex: "V5",
        width: "100px",
        render: (text, record, index) => {
          const editing = this.isEditing(index);
          return (
            editing ?
              <Input value={text} onChange={e => {
                this.updateField(table, index, "V5", e.target.value);
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
                <Button disabled={this.state.editingIndex !== ""} style={{marginRight: "5px"}} icon={<DeleteOutlined />} size="small" onClick={() => this.deletePolicy(table, index)} />
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
      <Button type="primary" onClick={() => {this.synPolicies();}}>
        {i18next.t("adapter:Sync")}
      </Button>
      {
        this.renderTable(this.state.policyLists)
      }
    </>
    );
  }
}

export default PolicyTable;
