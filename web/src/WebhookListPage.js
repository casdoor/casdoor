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
import {Button, Popconfirm, Switch, Table} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as WebhookBackend from "./backend/WebhookBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";

class WebhookListPage extends BaseListPage {
  newWebhook() {
    const randomName = Setting.getRandomName();
    return {
      owner: "admin", // this.props.account.webhookname,
      name: `webhook_${randomName}`,
      createdTime: moment().format(),
      organization: "built-in",
      url: "https://example.com/callback",
      method: "POST",
      contentType: "application/json",
      headers: [],
      events: ["signup", "login", "logout", "update-user"],
      isEnabled: true,
    };
  }

  addWebhook() {
    const newWebhook = this.newWebhook();
    WebhookBackend.addWebhook(newWebhook)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/webhooks/${newWebhook.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteWebhook(i) {
    WebhookBackend.deleteWebhook(this.state.data[i])
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.setState({
            data: Setting.deleteRow(this.state.data, i),
            pagination: {total: this.state.pagination.total - 1},
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderTable(webhooks) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "150px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/webhooks/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: "110px",
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
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "180px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("webhook:URL"),
        dataIndex: "url",
        key: "url",
        width: "300px",
        sorter: true,
        ...this.getColumnSearchProps("url"),
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              {
                Setting.getShortText(text)
              }
            </a>
          );
        },
      },
      {
        title: i18next.t("webhook:Method"),
        dataIndex: "method",
        key: "method",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("method"),
      },
      {
        title: i18next.t("webhook:Content type"),
        dataIndex: "contentType",
        key: "contentType",
        width: "200px",
        sorter: true,
        filterMultiple: false,
        filters: [
          {text: "application/json", value: "application/json"},
          {text: "application/x-www-form-urlencoded", value: "application/x-www-form-urlencoded"},
        ],
      },
      {
        title: i18next.t("webhook:Events"),
        dataIndex: "events",
        key: "events",
        // width: '100px',
        sorter: true,
        ...this.getColumnSearchProps("events"),
        render: (text, record, index) => {
          return Setting.getTags(text);
        },
      },
      {
        title: i18next.t("webhook:Is user extended"),
        dataIndex: "isUserExtended",
        key: "isUserExtended",
        width: "160px",
        sorter: true,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          );
        },
      },
      {
        title: i18next.t("general:Is enabled"),
        dataIndex: "isEnabled",
        key: "isEnabled",
        width: "120px",
        sorter: true,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "170px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/webhooks/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete webhook: ${record.name} ?`}
                onConfirm={() => this.deleteWebhook(index)}
              >
                <Button style={{marginBottom: "10px"}} type="primary" danger>{i18next.t("general:Delete")}</Button>
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={webhooks} rowKey="name" size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Webhooks")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addWebhook.bind(this)}>{i18next.t("general:Add")}</Button>
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
    if (params.contentType !== undefined && params.contentType !== null) {
      field = "contentType";
      value = params.contentType;
    }
    this.setState({loading: true});
    WebhookBackend.getWebhooks("admin", params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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
        } else {
          if (res.msg.includes("Unauthorized")) {
            this.setState({
              loading: false,
              isAuthorized: false,
            });
          }
        }
      });
  };
}

export default WebhookListPage;
