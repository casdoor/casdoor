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
import {Button, Popconfirm, Table} from "antd";
import * as Setting from "./Setting";
import * as LdapBackend from "./backend/LdapBackend";
import i18next from "i18next";
import {Link} from "react-router-dom";

class LdapSyncPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      ldapId: props.match.params.ldapId,
      organizationName: props.match.params.organizationName,
      ldap: null,
      users: [],
      existUuids: [],
      selectedUsers: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getLdap();
  }

  syncUsers() {
    const selectedUsers = this.state.selectedUsers;
    if (selectedUsers === null || selectedUsers.length === 0) {
      Setting.showMessage("error", "Please select al least 1 user first");
      return;
    }

    LdapBackend.syncUsers(this.state.ldap.owner, this.state.ldap.id, selectedUsers)
      .then((res => {
        if (res.status === "ok") {
          const exist = res.data.exist;
          const failed = res.data.failed;
          const existUser = [];
          const failedUser = [];

          if ((!exist || exist.length === 0) && (!failed || failed.length === 0)) {
            Setting.goToLink(`/organizations/${this.state.ldap.owner}/users`);
          } else {
            if (exist && exist.length > 0) {
              exist.forEach(elem => {
                existUser.push(elem.cn);
              });
              Setting.showMessage("error", `User [${existUser}] is already exist`);
            }

            if (failed && failed.length > 0) {
              failed.forEach(elem => {
                failedUser.push(elem.cn);
              });
              Setting.showMessage("error", `Sync [${failedUser}] failed`);
            }
          }
        } else {
          Setting.showMessage("error", res.msg);
        }
      }));
  }

  getLdap() {
    LdapBackend.getLdap(this.state.organizationName, this.state.ldapId)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            ldap: res.data,
          });
          this.getLdapUser();
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  getLdapUser() {
    LdapBackend.getLdapUser(this.state.organizationName, this.state.ldapId)
      .then((res) => {
        if (res.status === "ok") {
          this.setState((prevState) => {
            prevState.users = res.data.users;
            prevState.existUuids = res.data2?.length > 0 ? res.data2 : [];
            return prevState;
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  buildValArray(data, key) {
    const valTypesArray = [];

    if (data !== null && data.length > 0) {
      data.forEach(elem => {
        const val = elem[key];
        if (!valTypesArray.includes(val)) {
          valTypesArray.push(val);
        }
      });
    }
    return valTypesArray;
  }

  buildFilter(data, key) {
    const filterArray = [];

    if (data !== null && data.length > 0) {
      const valArray = this.buildValArray(data, key);
      valArray.forEach(elem => {
        filterArray.push({
          text: elem,
          value: elem,
        });
      });
    }
    return filterArray;
  }

  renderTable(users) {
    const columns = [
      {
        title: i18next.t("ldap:CN"),
        dataIndex: "cn",
        key: "cn",
        sorter: (a, b) => a.cn.localeCompare(b.cn),
        render: (text, record, index) => {
          return (<div style={{display: "flex", justifyContent: "space-between"}}>
            <div>
              {text}
            </div>
            {this.state.existUuids.includes(record.uuid) ?
              Setting.getTag("green", i18next.t("ldap:synced")) :
              Setting.getTag("red", i18next.t("ldap:unsynced"))
            }
          </div>);
        },
      },
      {
        title: "Uid",
        dataIndex: "uid",
        key: "uid",
        sorter: (a, b) => a.uid.localeCompare(b.uid),
        render: (text, record, index) => {
          return (
            this.state.existUuids.includes(record.uuid) ?
              <Link to={`/users/${this.state.organizationName}/${text}`}>
                {text}
              </Link> :
              text
          );
        },
      },
      {
        title: "UidNumber",
        dataIndex: "uidNumber",
        key: "uidNumber",
        sorter: (a, b) => a.uidNumber.localeCompare(b.uidNumber),
        render: (text, record, index) => {
          return text;
        },
      },
      {
        title: i18next.t("ldap:Group ID"),
        dataIndex: "groupId",
        key: "groupId",
        sorter: (a, b) => a.groupId.localeCompare(b.groupId),
        filters: this.buildFilter(this.state.users, "groupId"),
        onFilter: (value, record) => record.groupId.indexOf(value) === 0,
      },
      {
        title: i18next.t("general:Email"),
        dataIndex: "email",
        key: "email",
        sorter: (a, b) => a.email.localeCompare(b.email),
      },
      {
        title: i18next.t("general:Phone"),
        dataIndex: "phone",
        key: "phone",
        sorter: (a, b) => a.phone.localeCompare(b.phone),
      },
      {
        title: i18next.t("user:Address"),
        dataIndex: "address",
        key: "address",
        sorter: (a, b) => a.address.localeCompare(b.address),
      },
    ];

    const rowSelection = {
      onChange: (selectedRowKeys, selectedRows) => {
        this.setState({
          selectedUsers: selectedRows,
        });
      },
      getCheckboxProps: record => ({
        disabled: this.state.existUuids.indexOf(record.uuid) !== -1,
      }),
    };

    return (
      <Table rowSelection={rowSelection} columns={columns} dataSource={users} rowKey="uuid" bordered size="small"
        pagination={{defaultPageSize: 10, showQuickJumper: true, showSizeChanger: true}}
        title={() => (
          <div>
            {this.state.ldap?.serverName}
            <Popconfirm placement={"right"} disabled={this.state.selectedUsers.length === 0}
              title={"Please confirm to sync selected users"}
              onConfirm={() => this.syncUsers()}
            >
              <Button type="primary" style={{marginLeft: "10px"}} disabled={this.state.selectedUsers.length === 0}>
                {i18next.t("general:Sync")}
              </Button>
            </Popconfirm>
            <Button style={{marginLeft: "20px"}}
              onClick={() => Setting.goToLink(`/ldap/${this.state.organizationName}/${this.state.ldapId}`)}>
              {i18next.t("general:Edit")} LDAP
            </Button>
          </div>
        )}
        loading={users === null}
      />
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.state.users)
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => {
            this.props.history.push(`/organizations/${this.state.organizationName}`);
          }}>
            {i18next.t("general:Save & Exit")}
          </Button>
        </div>
      </div>
    );
  }
}

export default LdapSyncPage;
