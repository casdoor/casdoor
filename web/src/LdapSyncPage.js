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
import {Button, Col, Popconfirm, Row, Table} from "antd";
import * as Setting from "./Setting";
import * as LdapBackend from "./backend/LdapBackend";
import i18next from "i18next";

class LdapSyncPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      ldapId: props.match.params.ldapId,
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
    let selectedUsers = this.state.selectedUsers;
    if (selectedUsers === null || selectedUsers.length === 0) {
      Setting.showMessage("error", "Please select al least 1 user first");
      return;
    }

    LdapBackend.syncUsers(this.state.ldap.owner, this.state.ldap.id, selectedUsers)
      .then((res => {
        if (res.status === "ok") {
          let exist = res.data.exist;
          let failed = res.data.failed;
          let existUser = [];
          let failedUser = [];

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
    LdapBackend.getLdap(this.state.ldapId)
      .then((res) => {
        if (res.status === "ok") {
          this.setState((prevState) => {
            prevState.ldap = res.data;
            return prevState;
          });
          this.getLdapUser(res.data);
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }


  getLdapUser(ldap) {
    LdapBackend.getLdapUser(ldap)
      .then((res) => {
        if (res.status === "ok") {
          this.setState((prevState) => {
            prevState.users = res.data.users;
            return prevState;
          });
          this.getExistUsers(ldap.owner, res.data.users);
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  getExistUsers(owner, users) {
    let uuidArray = [];
    users.forEach(elem => {
      uuidArray.push(elem.uuid);
    });
    LdapBackend.checkLdapUsersExist(owner, uuidArray)
      .then((res) => {
        if (res.status === "ok") {
          this.setState(prevState => {
            prevState.existUuids = res.data?.length > 0 ? res.data : [];
            return prevState;
          });
        }
      });
  }

  buildValArray(data, key) {
    let valTypesArray = [];

    if (data !== null && data.length > 0) {
      data.forEach(elem => {
        let val = elem[key];
        if (!valTypesArray.includes(val)) {
          valTypesArray.push(val);
        }
      });
    }
    return valTypesArray;
  }

  buildFilter(data, key) {
    let filterArray = [];

    if (data !== null && data.length > 0) {
      let valArray = this.buildValArray(data, key);
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
      },
      {
        title: i18next.t("ldap:UidNumber / Uid"),
        dataIndex: "uidNumber",
        key: "uidNumber",
        width: "200px",
        sorter: (a, b) => a.uidNumber.localeCompare(b.uidNumber),
        render: (text, record, index) => {
          return `${text} / ${record.uid}`;
        },
      },
      {
        title: i18next.t("ldap:Group Id"),
        dataIndex: "groupId",
        key: "groupId",
        width: "140px",
        sorter: (a, b) => a.groupId.localeCompare(b.groupId),
        filters: this.buildFilter(this.state.users, "groupId"),
        onFilter: (value, record) => record.groupId.indexOf(value) === 0,
      },
      {
        title: i18next.t("ldap:Email"),
        dataIndex: "email",
        key: "email",
        width: "240px",
        sorter: (a, b) => a.email.localeCompare(b.email),
      },
      {
        title: i18next.t("ldap:Phone"),
        dataIndex: "phone",
        key: "phone",
        width: "160px",
        sorter: (a, b) => a.phone.localeCompare(b.phone),
      },
      {
        title: i18next.t("ldap:Address"),
        dataIndex: "address",
        key: "address",
        sorter: (a, b) => a.address.localeCompare(b.address),
      },
    ];

    const rowSelection = {
      onChange: (selectedRowKeys, selectedRows) => {
        this.setState(prevState => {
          prevState.selectedUsers = selectedRows;
          return prevState;
        });
      },
      getCheckboxProps: record => ({
        disabled: this.state.existUuids.indexOf(record.uuid) !== -1,
      }),
    };

    return (
      <div>
        <Table rowSelection={rowSelection} columns={columns} dataSource={users} rowKey="uuid" bordered
          pagination={{defaultPageSize: 10, showQuickJumper: true, showSizeChanger: true}}
          title={() => (
            <div>
              <span>{this.state.ldap?.serverName}</span>
              <Popconfirm placement={"right"}
                title={"Please confirm to sync selected users"}
                onConfirm={() => this.syncUsers()}
              >
                <Button type="primary" size="small"
                  style={{marginLeft: "10px"}}>{i18next.t("ldap:Sync")}</Button>
              </Popconfirm>
            </div>
          )}
          loading={users === null}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.renderTable(this.state.users)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default LdapSyncPage;
