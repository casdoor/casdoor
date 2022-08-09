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
import {Link} from "react-router-dom";
import {Button, Popconfirm, Table} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as TokenBackend from "./backend/TokenBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";

class TokenListPage extends BaseListPage {
  newToken() {
    const randomName = Setting.getRandomName();
    return {
      owner: "admin", // this.props.account.tokenname,
      name: `token_${randomName}`,
      createdTime: moment().format(),
      application: "app-built-in",
      organization: "built-in",
      user: "admin",
      accessToken: "",
      expiresIn: 7200,
      scope: "read",
      tokenType: "Bearer",
    };
  }

  addToken() {
    const newToken = this.newToken();
    TokenBackend.addToken(newToken)
      .then((res) => {
        this.props.history.push({pathname: `/tokens/${newToken.name}`, mode: "add"});
      }
      )
      .catch(error => {
        Setting.showMessage("error", `Token failed to add: ${error}`);
      });
  }

  deleteToken(i) {
    TokenBackend.deleteToken(this.state.data[i])
      .then((res) => {
        Setting.showMessage("success", "Token deleted successfully");
        this.setState({
          data: Setting.deleteRow(this.state.data, i),
          pagination: {total: this.state.pagination.total - 1},
        });
      }
      )
      .catch(error => {
        Setting.showMessage("error", `Token failed to delete: ${error}`);
      });
  }

  renderTable(tokens) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: (Setting.isMobile()) ? "100px" : "300px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/tokens/${text}`}>
              {text}
            </Link>
          );
        },
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
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("organization"),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          );
        },
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
            <Link to={`/users/${record.organization}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("token:Authorization code"),
        dataIndex: "code",
        key: "code",
        // width: '150px',
        sorter: true,
        ...this.getColumnSearchProps("code"),
        render: (text, record, index) => {
          return Setting.getClickable(text);
        },
      },
      {
        title: i18next.t("token:Access token"),
        dataIndex: "accessToken",
        key: "accessToken",
        // width: '150px',
        sorter: true,
        ellipsis: true,
        ...this.getColumnSearchProps("accessToken"),
        render: (text, record, index) => {
          return Setting.getClickable(text);
        },
      },
      {
        title: i18next.t("token:Expires in"),
        dataIndex: "expiresIn",
        key: "expiresIn",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("expiresIn"),
      },
      {
        title: i18next.t("token:Scope"),
        dataIndex: "scope",
        key: "scope",
        width: "110px",
        sorter: true,
        ...this.getColumnSearchProps("scope"),
      },
      // {
      //   title: i18next.t("token:Token type"),
      //   dataIndex: 'tokenType',
      //   key: 'tokenType',
      //   width: '130px',
      //   sorter: (a, b) => a.tokenType.localeCompare(b.tokenType),
      // },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "170px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/tokens/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete token: ${record.name} ?`}
                onConfirm={() => this.deleteToken(index)}
              >
                <Button style={{marginBottom: "10px"}} type="danger">{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={tokens} rowKey="name" size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Tokens")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addToken.bind(this)}>{i18next.t("general:Add")}</Button>
            </div>
          )}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    const field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    this.setState({loading: true});
    TokenBackend.getTokens("admin", params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            loading: false,
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        }
      });
  };
}

export default TokenListPage;
