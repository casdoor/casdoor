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
import {Button, Descriptions, Drawer, Switch, Table, Tooltip} from "antd";
import * as Setting from "./Setting";
import * as RecordBackend from "./backend/RecordBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import Editor from "./common/Editor";

class RecordListPage extends BaseListPage {
  UNSAFE_componentWillMount() {
    this.state.pagination.pageSize = 20;
    const {pagination} = this.state;
    this.fetch({pagination});
  }

  renderTable(records) {
    let columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "320px",
        sorter: true,
        ...this.getColumnSearchProps("name"),
      },
      {
        title: i18next.t("general:ID"),
        dataIndex: "id",
        key: "id",
        width: "90px",
        sorter: true,
        ...this.getColumnSearchProps("id"),
      },
      {
        title: i18next.t("general:Client IP"),
        dataIndex: "clientIp",
        key: "clientIp",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("clientIp", (row, highlightContent) => (
          <a target="_blank" rel="noreferrer" href={`https://db-ip.com/${row.text}`}>
            {highlightContent}
          </a>
        )),
      },
      {
        title: i18next.t("general:Timestamp"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "150px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
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
        title: i18next.t("general:User"),
        dataIndex: "user",
        key: "user",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("user"),
        render: (text, record, index) => {
          return (
            <Link to={`/users/${record.organization}/${record.user}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Method"),
        dataIndex: "method",
        key: "method",
        width: "100px",
        sorter: true,
        filterMultiple: false,
        filters: [
          "GET", "HEAD", "POST", "PUT", "DELETE",
          "CONNECT", "OPTIONS", "TRACE", "PATCH",
        ].map(el => ({text: el, value: el})),
      },
      {
        title: i18next.t("general:Request URI"),
        dataIndex: "requestUri",
        key: "requestUri",
        width: "200px",
        sorter: true,
        ellipsis: {
          showTitle: false,
        },
        ...this.getColumnSearchProps("requestUri", (row, highlightContent) => (
          <Tooltip placement="topLeft" title={row.text}>
            {highlightContent}
          </Tooltip>
        )),
      },
      {
        title: i18next.t("user:Language"),
        dataIndex: "language",
        key: "language",
        width: "90px",
        sorter: true,
        ...this.getColumnSearchProps("language"),
      },
      {
        title: i18next.t("record:Status code"),
        dataIndex: "statusCode",
        key: "statusCode",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("statusCode"),
      },
      {
        title: i18next.t("record:Response"),
        dataIndex: "response",
        key: "response",
        width: "220px",
        sorter: true,
        ellipsis: {
          showTitle: false,
        },
        ...this.getColumnSearchProps("response", (row, highlightContent) => (
          <Tooltip placement="topLeft" title={row.text}>
            {highlightContent}
          </Tooltip>
        )),
      },
      {
        title: i18next.t("record:Object"),
        dataIndex: "object",
        key: "object",
        width: "200px",
        sorter: true,
        ellipsis: {
          showTitle: false,
        },
        ...this.getColumnSearchProps("object"),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "action",
        key: "action",
        width: "200px",
        sorter: true,
        ...this.getColumnSearchProps("action"),
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return text;
        },
      },
      {
        title: i18next.t("record:Is triggered"),
        dataIndex: "isTriggered",
        key: "isTriggered",
        width: "120px",
        sorter: true,
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          if (!["signup", "login", "logout", "update-user", "new-user"].includes(record.action)) {
            return null;
          }

          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "action",
        key: "action",
        width: "80px",
        sorter: true,
        fixed: "right",
        render: (text, record, index) => (
          <Button type="link" onClick={() => {
            this.setState({
              detailRecord: record,
              detailShow: true,
            });
          }}>
            {i18next.t("general:Detail")}
          </Button>
        ),
      },
    ];

    if (Setting.isLocalAdminUser(this.props.account)) {
      columns = columns.filter(column => column.key !== "name");
    }

    const paginationProps = {
      total: this.state.pagination.total,
      pageSize: this.state.pagination.pageSize,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <div>
        <Table scroll={{x: "100%"}} columns={columns} dataSource={records} rowKey="id" size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Records")}&nbsp;&nbsp;&nbsp;&nbsp;
            </div>
          )}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
        {/* TODO: Should be packaged as a component after confirm it run correctly.*/}
        <Drawer
          title={i18next.t("general:Detail")}
          width={Setting.isMobile() ? "100%" : 640}
          placement="right"
          destroyOnClose
          onClose={() => this.setState({detailShow: false})}
          open={this.state.detailShow}
        >
          <Descriptions bordered size="small" column={1} layout={Setting.isMobile() ? "vertical" : "horizontal"} style={{padding: "12px", height: "100%", overflowY: "auto"}}>
            <Descriptions.Item label={i18next.t("general:ID")}>{this.getDetailField("id")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Client IP")}>{this.getDetailField("clientIp")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Timestamp")}>{this.getDetailField("createdTime")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Organization")}>
              <Link to={`/organizations/${this.getDetailField("organization")}`}>
                {this.getDetailField("organization")}
              </Link>
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:User")}>
              <Link to={`/users/${this.getDetailField("organization")}/${this.getDetailField("user")}`}>
                {this.getDetailField("user")}
              </Link>
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Method")}>{this.getDetailField("method")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Request URI")}>{this.getDetailField("requestUri")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("user:Language")}>{this.getDetailField("language")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("record:Status code")}>{this.getDetailField("statusCode")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Action")}>{this.getDetailField("action")}</Descriptions.Item>
            <Descriptions.Item label={i18next.t("record:Response")}>
              <Editor
                value={this.getDetailField("response")}
                fillHeight
                fillWidth
                maxWidth={this.getEditorMaxWidth()}
                dark
                readOnly
              />
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("record:Object")}>
              <Editor
                value={this.jsonStrFormatter(this.getDetailField("object"))}
                lang="json"
                fillHeight
                fillWidth
                maxWidth={this.getEditorMaxWidth()}
                dark
                readOnly
              />
            </Descriptions.Item>
          </Descriptions>
        </Drawer>
      </div>
    );
  }

  getEditorMaxWidth = () => {
    return Setting.isMobile() ? window.innerWidth - 60 : 475;
  };

  jsonStrFormatter = str => {
    try {
      return JSON.stringify(JSON.parse(str), null, 2);
    } catch (e) {
      // eslint-disable-next-line no-console
      console.error(e);
      return str;
    }
  };

  getDetailField = dataIndex => {
    return this.state.detailRecord ? this.state.detailRecord?.[dataIndex] ?? "" : "";
  };

  fetch = (params = {}) => {
    let field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (params.method !== undefined && params.method !== null) {
      field = "method";
      value = params.method;
    }
    this.setState({loading: true});
    RecordBackend.getRecords(Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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
            detailShow: false,
            detailRecord: null,
          });
        } else {
          if (res.data.includes("Please login first")) {
            this.setState({
              loading: false,
              isAuthorized: false,
            });
          }
        }
      });
  };
}

export default RecordListPage;
