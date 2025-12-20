// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
import i18next from "i18next";
import {Link} from "react-router-dom";
import {Popconfirm, Table, Tag} from "antd";
import React from "react";
import * as SessionBackend from "./backend/SessionBackend";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class SessionListPage extends BaseListPage {
  handleTagClose = (rowIndex, sessionId, e) => {
    e.preventDefault();
    e.stopPropagation();

    this.setState({
      confirmTagKey: `${rowIndex}-${sessionId}`,
    });
  };

  deleteSession(i, sessionId = "") {
    // Pass the optional sessionId to the backend. If sessionId is empty, the backend will delete the whole session record.
    SessionBackend.deleteSession(this.state.data[i], sessionId)
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

  renderTable(sessions) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "150px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "110px",
        sorter: true,
        ...this.getColumnSearchProps("owner"),
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
        title: i18next.t("general:Session ID"),
        dataIndex: "sessionId",
        key: "sessionId",
        width: "180px",
        sorter: true,
        render: (text, record, index) => {
          return text.map((item, idx) => {
            const tagKey = `${index}-${item}`;
            const confirmTitle = i18next.t("general:Sure to delete");
            const confirmContent = `${i18next.t("general:Session ID")}: ${item}`;
            const isActive = this.state.confirmTagKey === tagKey;
            return (
              <Popconfirm
                key={`${index}-${idx}`}
                title={confirmTitle}
                description={confirmContent}
                open={isActive}
                onConfirm={() => {this.deleteSession(index, item); this.setState({confirmTagKey: null});}}
                onCancel={() => this.setState({confirmTagKey: null})}
                onOpenChange={(visible) => {if (!visible && isActive) {this.setState({confirmTagKey: null});}}}
                okText={i18next.t("general:OK")}
                cancelText={i18next.t("general:Cancel")}
              >
                <Tag closable onClose={(e) => this.handleTagClose(index, item, e)}>{item}</Tag>
              </Popconfirm>
            );
          });
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "70px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <PopconfirmModal
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteSession(index)}
              >
              </PopconfirmModal>
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={sessions} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
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
    SessionBackend.getSessions(Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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

export default SessionListPage;
