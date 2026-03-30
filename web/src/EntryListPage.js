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
import {Input, Table, Tooltip} from "antd";
import {Link} from "react-router-dom";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import * as EntryBackend from "./backend/EntryBackend";
import * as Setting from "./Setting";

class EntryListPage extends BaseListPage {
  constructor(props) {
    super(props);
    this.state = {
      ...this.state,
      searchAgentName: "",
      searchedAgentName: "",
    };
  }

  handleAgentNameChange = (e) => {
    const value = e.target.value;
    this.setState({
      searchAgentName: value,
    });

    if (value === "" && this.state.searchedAgentName !== "") {
      this.fetch({
        pagination: {
          ...this.state.pagination,
          current: 1,
        },
        agentName: "",
      });
    }
  };

  handleAgentNameSearch = (value) => {
    this.fetch({
      pagination: {
        ...this.state.pagination,
        current: 1,
      },
      agentName: value.trim(),
    });
  };

  handleTableChange = (pagination) => {
    this.setState({
      pagination,
    });
  };

  renderTable(entries) {
    const entriesLabel = i18next.t("general:Entries", {defaultValue: "Entries"});
    const searchPlaceholder = i18next.language?.startsWith("zh") ? "按 Agent 名称搜索" : "Search entries by agent name";

    const columns = [
      {
        title: i18next.t("general:ID"),
        dataIndex: "id",
        key: "id",
        width: 90,
        sorter: (a, b) => (a.id || 0) - (b.id || 0),
      },
      {
        title: i18next.t("general:Owner"),
        dataIndex: "owner",
        key: "owner",
        width: 120,
        sorter: (a, b) => (a.owner || "").localeCompare(b.owner || ""),
      },
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: 220,
        sorter: (a, b) => (a.name || "").localeCompare(b.name || ""),
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: 180,
        sorter: (a, b) => (a.createdTime || "").localeCompare(b.createdTime || ""),
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: 140,
        sorter: (a, b) => (a.organization || "").localeCompare(b.organization || ""),
        render: (text) => (
          <Link to={`/organizations/${text}`}>
            {text}
          </Link>
        ),
      },
      {
        title: i18next.t("general:User"),
        dataIndex: "user",
        key: "user",
        width: 140,
        sorter: (a, b) => (a.user || "").localeCompare(b.user || ""),
        render: (text, record) => (
          <Link to={`/users/${record.organization}/${text}`}>
            {text}
          </Link>
        ),
      },
      {
        title: i18next.t("user:Language"),
        dataIndex: "language",
        key: "language",
        width: 100,
        sorter: (a, b) => (a.language || "").localeCompare(b.language || ""),
      },
      {
        title: "Time",
        dataIndex: "time",
        key: "time",
        width: 180,
        sorter: (a, b) => (a.time || "").localeCompare(b.time || ""),
      },
      {
        title: "Agent",
        dataIndex: "agent",
        key: "agent",
        width: 180,
        sorter: (a, b) => (a.agent || "").localeCompare(b.agent || ""),
      },
      {
        title: "Message",
        dataIndex: "message",
        key: "message",
        ellipsis: {
          showTitle: false,
        },
        sorter: (a, b) => (a.message || "").localeCompare(b.message || ""),
        render: (text) => (
          <Tooltip placement="topLeft" title={text}>
            {text}
          </Tooltip>
        ),
      },
    ];

    const filteredColumns = Setting.filterTableColumns(columns, this.props.formItems ?? this.state.formItems);
    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <Table
        scroll={{x: "max-content"}}
        columns={filteredColumns}
        dataSource={entries}
        rowKey="id"
        size="middle"
        bordered
        pagination={{...this.state.pagination, ...paginationProps}}
        loading={this.state.loading}
        onChange={this.handleTableChange}
        title={() => (
          <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12, flexWrap: "wrap"}}>
            <div>{entriesLabel}</div>
            <Input.Search
              allowClear
              enterButton={i18next.t("general:Search")}
              placeholder={searchPlaceholder}
              value={this.state.searchAgentName}
              onChange={this.handleAgentNameChange}
              onSearch={this.handleAgentNameSearch}
              style={{width: Setting.isMobile() ? "100%" : 320}}
            />
          </div>
        )}
      />
    );
  }

  fetch = (params = {}) => {
    const pagination = params.pagination ?? this.state.pagination;
    const agentName = params.agentName ?? this.state.searchedAgentName;
    const owner = Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account);

    this.setState({loading: true});
    const request = agentName === "" ? EntryBackend.getAllEntries(owner) : EntryBackend.getEntries(owner, agentName);

    request
      .then((res) => {
        this.setState({
          loading: false,
        });

        if (res.status === "ok") {
          const entries = res.data || [];
          const pageSize = pagination.pageSize || this.state.pagination.pageSize;
          const totalPages = Math.max(1, Math.ceil(entries.length / pageSize));
          const current = Math.min(pagination.current || 1, totalPages);

          this.setState({
            data: entries,
            pagination: {
              ...pagination,
              current,
              total: entries.length,
            },
            searchAgentName: agentName,
            searchedAgentName: agentName,
          });
        } else if (Setting.isResponseDenied(res)) {
          this.setState({
            isAuthorized: false,
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => {
        this.setState({
          loading: false,
        });
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };
}

export default EntryListPage;
