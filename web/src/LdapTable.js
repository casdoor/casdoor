// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import {Button, Col, Popconfirm, Row, Table} from 'antd';
import * as Setting from "./Setting";
import i18next from "i18next";
import * as LdapBackend from "./backend/LdapBackend";
import {Link} from "react-router-dom";

class LdapTable extends React.Component {
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

  newLdap() {
    return {
      id: "",
      owner: this.props.organizationName,
      createdTime: "",
      serverName: "Example LDAP Server",
      host: "example.com",
      port: 389,
      admin: "cn=admin,dc=example,dc=com",
      passwd: "123",
      baseDn: "ou=People,dc=example,dc=com",
      autosync: 0,
      lastSync: ""
    }
  }

  addRow(table) {
    const newLdap = this.newLdap();
    LdapBackend.addLdap(newLdap)
      .then((res) => {
          if (res.status === "ok") {
            Setting.showMessage("success", `Add LDAP server success`);
            if (table === undefined) {
              table = [];
            }
            table = Setting.addRow(table, res.data2);
            this.updateTable(table);
          } else {
            Setting.showMessage("error", res.msg);
          }
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Add LDAP server failed: ${error}`);
      });
  }

  deleteRow(table, i) {
    LdapBackend.deleteLdap(table[i])
      .then((res) => {
          if (res.status === "ok") {
            Setting.showMessage("success", `Delete LDAP server success`);
            table = Setting.deleteRow(table, i);
            this.updateTable(table);
          } else {
            Setting.showMessage("error", res.msg);
          }
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Delete LDAP server failed: ${error}`);
      });
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("ldap:Server Name"),
        dataIndex: "serverName",
        key: "serverName",
        width: "160px",
        sorter: (a, b) => a.serverName.localeCompare(b.serverName),
        render: (text, record, index) => {
          return (
            <Link to={`/ldaps/${record.id}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("ldap:Server"),
        dataIndex: "host",
        key: "host",
        ellipsis: true,
        sorter: (a, b) => a.host.localeCompare(b.host),
        render: (text, record, index) => {
          return `${text}:${record.port}`
        }
      },
      {
        title: i18next.t("ldap:Base DN"),
        dataIndex: "baseDn",
        key: "baseDn",
        ellipsis: true,
        sorter: (a, b) => a.baseDn.localeCompare(b.baseDn),
      },
      {
        title: i18next.t("ldap:Auto Sync"),
        dataIndex: "autoSync",
        key: "autoSync",
        width: "120px",
        sorter: (a, b) => a.autoSync.localeCompare(b.autoSync),
        render: (text, record, index) => {
          return text === 0 ? (<span style={{color: "#faad14"}}>Disable</span>) : (
            <span style={{color: "#52c41a"}}>{text + " mins"}</span>)
        }
      },
      {
        title: i18next.t("ldap:Last Sync"),
        dataIndex: "lastSync",
        key: "lastSync",
        ellipsis: true,
        sorter: (a, b) => a.lastSync.localeCompare(b.lastSync),
        render: (text, record, index) => {
          return text
        }
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "240px",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}}
                      type="primary"
                      onClick={() => Setting.goToLink(`/ldap/sync/${record.id}`)}>{i18next.t("ldap:Sync")}</Button>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}}
                      onClick={() => Setting.goToLink(`/ldap/${record.id}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete LDAP Config: ${record.serverName} ?`}
                onConfirm={() => this.deleteRow(table, index)}
              >
                <Button style={{marginBottom: "10px"}}
                        type="danger">{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
          )
        }
      },
    ];

    return (
      <Table scroll={{x: 'max-content'}} rowKey="id" columns={columns} dataSource={table} size="small" bordered pagination={false}
        title={() => (
          <div>
            {this.props.title}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button style={{marginRight: "5px"}} type="primary" size="small"
                    onClick={() => this.addRow(table)}>{i18next.t("general:Add")}</Button>
          </div>
        )}
      />
    );
  }

  render() {
    return (
      <div>
        <Row>
          <Col span={24}>
            {
              this.renderTable(this.props.table)
            }
          </Col>
        </Row>
      </div>
    )
  }
}

export default LdapTable;
