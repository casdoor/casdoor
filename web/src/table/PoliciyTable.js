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
      page: 1,
    };
  }

  count = 0;
  pageSize = 10;

  getIndex(index) {
    // Need to be used in all place when modify table. Parameter is the row index in table, need to calculate the index in dataSource.
    return index + (this.state.page - 1) * 10;
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
    Object.keys(table[this.getIndex(index)]).forEach((key) => {
      table[this.getIndex(index)][key] = this.state.oldPolicy[key];
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
    table[this.getIndex(index)][key] = value;
    this.updateTable(table);
  }

  addRow(table) {
    const row = {key: this.count, Ptype: "p"};
    if (table === undefined) {
      table = [];
    }
    table = Setting.addRow(table, row, "top");

    this.count = this.count + 1;
    this.updateTable(table);
    this.edit(row, 0);
    this.setState({
      page: 1,
      add: true,
    });
  }

  deleteRow(table, index) {
    table = Setting.deleteRow(table, this.getIndex(index));
    this.updateTable(table);
  }

  save(table, i) {
    this.state.add ? this.addPolicy(table, i) : this.updatePolicy(table, i);
  }

  synPolicies() {
    this.setState({loading: true});
    AdapterBackend.syncPolicies(this.props.owner, this.props.name)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("adapter:Sync policies successfully"));

          const policyList = res.data;
          policyList.map((policy, index) => {
            policy.key = index;
          });
          this.count = policyList.length;
          this.setState({policyLists: policyList});
        } else {
          Setting.showMessage("error", `${i18next.t("adapter:Failed to sync policies")}: ${res.msg}`);
        }
        this.setState({loading: false});
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  updatePolicy(table, i) {
    AdapterBackend.UpdatePolicy(this.props.owner, this.props.name, [this.state.oldPolicy, table[i]]).then(res => {
      if (res.status === "ok") {
        this.setState({editingIndex: "", oldPolicy: ""});
        Setting.showMessage("success", i18next.t("general:Successfully saved"));
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
      }
    });
  }

  addPolicy(table, i) {
    AdapterBackend.AddPolicy(this.props.owner, this.props.name, table[i]).then(res => {
      if (res.status === "ok") {
        this.setState({editingIndex: "", oldPolicy: "", add: false});
        if (res.data !== "Affected") {
          res.msg = i18next.t("adapter:Duplicated policy rules");
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        } else {
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        }
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
      }
    });
  }

  deletePolicy(table, index) {
    AdapterBackend.RemovePolicy(this.props.owner, this.props.name, table[this.getIndex(index)]).then(res => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Successfully deleted"));

        this.deleteRow(table, index);
      } else {
        Setting.showMessage("error", i18next.t("general:Failed to delete"));
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
          defaultPageSize: this.pageSize,
          onChange: (page) => this.setState({
            page: page,
          }),
          disabled: this.state.editingIndex !== "",
          current: this.state.page,
        }}
        columns={columns} dataSource={table} rowKey="key" size="middle" bordered
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
    return (
      <React.Fragment>
        <Button type="primary" disabled={this.state.editingIndex !== ""} onClick={() => {this.synPolicies();}}>
          {i18next.t("general:Sync")}
        </Button>
        {
          this.renderTable(this.state.policyLists)
        }
      </React.Fragment>
    );
  }
}

export default PolicyTable;
