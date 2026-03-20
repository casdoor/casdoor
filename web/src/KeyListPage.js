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
import {Link} from "react-router-dom";
import {Button, Table, Tag} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as KeyBackend from "./backend/KeyBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class KeyListPage extends BaseListPage {
  getQueryFilters() {
    const params = new URLSearchParams(this.props.location.search);
    return {
      keyType: params.get("type") || "",
      organization: params.get("organization") || (Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account)),
      application: params.get("application") || "",
      user: params.get("user") || "",
    };
  }

  newKey() {
    const {keyType, organization, application, user} = this.getQueryFilters();
    return {
      owner: "admin",
      name: `key_${Setting.getRandomName()}`,
      createdTime: moment().format(),
      updatedTime: moment().format(),
      displayName: "",
      description: "",
      type: keyType || "general",
      organization: organization,
      application: application || "app-built-in",
      user: user,
      scopes: ["read"],
      isEnabled: true,
      expiresTime: "",
      lastUsedTime: "",
      secretPreview: "",
    };
  }

  addKey() {
    const newKey = this.newKey();
    this.props.history.push({
      pathname: "/keys/new",
      draftKey: newKey,
      mode: "add",
    });
  }

  deleteKey(i) {
    KeyBackend.deleteKey(this.state.data[i])
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.fetch({
            pagination: {
              ...this.state.pagination,
              current: this.state.pagination.current > 1 && this.state.data.length === 1 ? this.state.pagination.current - 1 : this.state.pagination.current,
            },
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderTable(keys) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: Setting.isMobile() ? "120px" : "220px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text) => (
          <Link to={`/keys/${text}`}>
            {text}
          </Link>
        ),
      },
      {
        title: i18next.t("general:Type"),
        dataIndex: "type",
        key: "type",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("type"),
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "180px",
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "160px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
        render: (text, record) => record.organization ? (
          <Link to={`/applications/${record.organization}/${text}`}>
            {text}
          </Link>
        ) : text,
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("organization"),
        render: (text) => text ? (
          <Link to={`/organizations/${text}`}>
            {text}
          </Link>
        ) : "",
      },
      {
        title: i18next.t("general:User"),
        dataIndex: "user",
        key: "user",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("user"),
        render: (text, record) => text ? (
          <Link to={`/users/${record.organization}/${text}`}>
            {text}
          </Link>
        ) : "",
      },
      {
        title: i18next.t("provider:Scope"),
        dataIndex: "scopes",
        key: "scopes",
        width: "180px",
        render: (text, record) => (record.scopes || []).map(scope => <Tag key={scope}>{scope}</Tag>),
      },
      {
        title: i18next.t("general:Enabled"),
        dataIndex: "isEnabled",
        key: "isEnabled",
        width: "110px",
        sorter: true,
        render: (text) => text ? i18next.t("general:Yes") : i18next.t("general:No"),
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "160px",
        sorter: true,
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: "Last used time",
        dataIndex: "lastUsedTime",
        key: "lastUsedTime",
        width: "170px",
        sorter: true,
        render: (text) => text ? Setting.getFormattedDate(text) : "",
      },
      {
        title: i18next.t("general:API key"),
        dataIndex: "secretPreview",
        key: "secretPreview",
        width: "160px",
        ...this.getColumnSearchProps("secretPreview"),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "170px",
        fixed: Setting.isMobile() ? "false" : "right",
        render: (text, record, index) => (
          <div>
            <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/keys/${record.name}`)}>{i18next.t("general:Edit")}</Button>
            <PopconfirmModal
              title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
              onConfirm={() => this.deleteKey(index)}
            />
          </div>
        ),
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
        <Table
          scroll={{x: "100%"}}
          columns={columns}
          dataSource={keys}
          rowKey={(record) => `${record.owner}/${record.name}`}
          size="middle"
          bordered
          pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Keys")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addKey.bind(this)}>{i18next.t("general:Add")}</Button>
            </div>
          )}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    const field = params.searchedColumn;
    const value = params.searchText;
    const sortField = params.sortField;
    const sortOrder = params.sortOrder;
    const {keyType, organization, application, user} = this.getQueryFilters();

    this.setState({loading: true});
    KeyBackend.getKeys("admin", keyType, organization, application, user, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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
        } else if (Setting.isResponseDenied(res)) {
          this.setState({
            isAuthorized: false,
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  };
}

export default KeyListPage;
