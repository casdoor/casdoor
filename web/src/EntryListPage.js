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
import {Button, Popover, Table} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as EntryBackend from "./backend/EntryBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";
import Editor from "./common/Editor";

class EntryListPage extends BaseListPage {
  newEntry() {
    const randomHex = Math.random().toString(16).slice(2, 18);
    const owner = Setting.getRequestOrganization(this.props.account);
    return {
      owner: owner,
      name: randomHex,
      createdTime: moment().format(),
      displayName: randomHex,
      provider: "",
      url: "",
      token: "",
      application: "",
      type: "",
      message: "",
    };
  }

  addEntry() {
    const newEntry = this.newEntry();
    EntryBackend.addEntry(newEntry)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/entries/${newEntry.owner}/${newEntry.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteEntry(i) {
    EntryBackend.deleteEntry(this.state.data[i])
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

  fetch = (params = {}) => {
    const field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (!params.pagination) {
      params.pagination = {current: 1, pageSize: 10};
    }

    this.setState({loading: true});
    EntryBackend.getEntries(Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        this.setState({loading: false});
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
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      });
  };

  renderTable(entries) {
    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "130px",
        sorter: true,
        ...this.getColumnSearchProps("owner"),
        render: (text) => {
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
        width: "160px",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record) => {
          return (
            <Link to={`/entries/${record.owner}/${text}`}>
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
        render: (text) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:Provider"),
        dataIndex: "provider",
        key: "provider",
        width: "160px",
        sorter: true,
        ...this.getColumnSearchProps("provider"),
        render: (text, record) => {
          if (!text) {
            return null;
          }
          return (
            <Link to={`/providers/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
        render: (text, record) => {
          if (!text) {
            return null;
          }
          return (
            <Link to={`/applications/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Listening URL"),
        dataIndex: "url",
        key: "url",
        sorter: true,
        ...this.getColumnSearchProps("url"),
        render: (text) => {
          if (!text) {
            return null;
          }
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              {Setting.getShortText(text, 40)}
            </a>
          );
        },
      },
      {
        title: i18next.t("payment:Message"),
        dataIndex: "message",
        key: "message",
        sorter: true,
        ...this.getColumnSearchProps("message"),
        render: (text) => {
          if (!text) {
            return null;
          }
          return (
            <Popover placement="topRight" content={() => (
              <Editor value={text} readOnly={true} />
            )} title="" trigger="hover">
              {Setting.getShortText(text, 60)}
            </Popover>
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "op",
        key: "op",
        width: "180px",
        fixed: (Setting.isMobile()) ? false : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/entries/${record.owner}/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <PopconfirmModal title={i18next.t("general:Sure to delete") + `: ${record.name} ?`} onConfirm={() => this.deleteEntry(index)}>
              </PopconfirmModal>
            </div>
          );
        },
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
        dataSource={entries}
        columns={filteredColumns}
        rowKey={record => `${record.owner}/${record.name}`}
        pagination={{...this.state.pagination, ...paginationProps}}
        loading={this.state.loading}
        onChange={this.handleTableChange}
        size="middle"
        bordered
        title={() => (
          <div>
            {i18next.t("general:Entries")}&nbsp;&nbsp;&nbsp;&nbsp;
            <Button type="primary" size="small" onClick={() => this.addEntry()}>{i18next.t("general:Add")}</Button>
          </div>
        )}
      />
    );
  }
}

export default EntryListPage;
