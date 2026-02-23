// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
import {Button, Popconfirm, Table, Tag} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as ConsentBackend from "../backend/ConsentBackend";

class ConsentTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  deleteScope(record, scopeToDelete) {
    ConsentBackend.revokeConsent({
      application: record.application,
      grantedScopes: scopeToDelete ? [scopeToDelete] : record.grantedScopes,
    })
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully revoked"));
          this.props.onUpdateTable();
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderTable(table) {
    const columns = [
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "200px",
        render: (text) => {
          return text;
        },
      },
      {
        title: i18next.t("consent:Granted scopes"),
        dataIndex: "grantedScopes",
        key: "grantedScopes",
        render: (text, record) => {
          return (
            <div style={{display: "flex", flexWrap: "wrap", gap: "4px"}}>
              {
                (Array.isArray(text) ? text : []).map((scope, index) => {
                  return (
                    <Popconfirm
                      key={index}
                      title={`${i18next.t("consent:Are you sure you want to revoke scope")}: ${scope}?`}
                      onConfirm={() => this.deleteScope(record, scope)}
                      okText={i18next.t("general:OK")}
                      cancelText={i18next.t("general:Cancel")}
                    >
                      <Tag
                        color="blue"
                        style={{cursor: "pointer"}}
                      >
                        {scope}
                      </Tag>
                    </Popconfirm>
                  );
                })
              }
            </div>
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: "100px",
        render: (_, record, __) => {
          return (
            <Popconfirm
              title={i18next.t("consent:Are you sure you want to revoke this consent?")}
              onConfirm={() => this.deleteScope(record)}
              okText={i18next.t("general:OK")}
              cancelText={i18next.t("general:Cancel")}
            >
              <Button type="primary" danger size="small">
                {i18next.t("consent:Delete")}
              </Button>
            </Popconfirm>
          );
        },
      },
    ];

    return (
      <Table scroll={{x: "max-content"}} rowKey="application" columns={columns} dataSource={table} size="middle" bordered pagination={false}
        title={() => (
          <div>
            {this.props.title}
          </div>
        )}
      />
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.props.table)
        }
      </div>
    );
  }
}

export default ConsentTable;
