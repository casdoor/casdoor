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

import BaseListPage from "./BaseListPage";
import * as Setting from "./Setting";
import moment from "moment/moment";
import * as VerificationBackend from "./backend/VerificationBackend";
import i18next from "i18next";
import {Link} from "react-router-dom";
import React from "react";
import {Switch, Table} from "antd";

class VerificationListPage extends BaseListPage {
  newVerification() {
    const randomName = Setting.getRandomName();

    return {
      owner: "admin",
      name: `Verification_${randomName}`,
      createdTime: moment().format(),
    };
  }

  renderTable(verifications) {
    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("owner"),
        render: (text, record, index) => {
          if (text === "admin") {
            return `(${i18next.t("general:empty")})`;
          }

          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "260px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "160px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("provider:Type"),
        dataIndex: "type",
        key: "type",
        width: "90px",
        sorter: true,
        ...this.getColumnSearchProps("type"),
      },
      {
        title: i18next.t("general:User"),
        dataIndex: "user",
        key: "user",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("user"),
        render: (text, record, index) => {
          return (
            <Link to={`/users/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Provider"),
        dataIndex: "provider",
        key: "provider",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("provider"),
        render: (text, record, index) => {
          return (
            <Link to={`/providers/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Client IP"),
        dataIndex: "remoteAddr",
        key: "remoteAddr",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("remoteAddr"),
        render: (text, record, index) => {
          let clientIp = text;
          if (clientIp.endsWith(": ")) {
            clientIp = clientIp.slice(0, -2);
          }

          return (
            <a target="_blank" rel="noreferrer" href={`https://db-ip.com/${clientIp}`}>
              {clientIp}
            </a>
          );
        },
      },
      {
        title: i18next.t("verification:Receiver"),
        dataIndex: "receiver",
        key: "receiver",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("receiver"),
      },
      {
        title: i18next.t("login:Verification code"),
        dataIndex: "code",
        key: "code",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("code"),
      },
      {
        title: i18next.t("verification:Is used"),
        dataIndex: "isUsed",
        key: "isUsed",
        width: "90px",
        sorter: true,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          );
        },
      },
    ];

    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <div>
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={verifications} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Verifications")}&nbsp;&nbsp;&nbsp;&nbsp;
            </div>
          )}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    let field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (params.type !== undefined && params.type !== null) {
      field = "type";
      value = params.type;
    }
    this.setState({loading: true});
    VerificationBackend.getVerifications("", Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        this.setState({
          loading: false,
        });
        if (res.status === "ok") {
          this.setState({
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
          if (Setting.isResponseDenied(res)) {
            this.setState({
              isAuthorized: false,
            });
          } else {
            Setting.showMessage("error", res.msg);
          }
        }
      });
  };
}

export default VerificationListPage;
