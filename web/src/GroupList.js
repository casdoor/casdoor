// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
import {Button, Table} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as GroupBackend from "./backend/GroupBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class GroupListPage extends BaseListPage {
  newGroup() {
    const randomName = Setting.getRandomName();
    return {
      owner: this.props.account.owner,
      name: `group_${randomName}`,
      createdTime: moment().format(),
      updatedTime: moment().format(),
      displayName: `New Group - ${randomName}`,
      type: "virtual",
      parentGroupId: "",
      isEnabled: true,
    };
  }

  addGroup() {
    const newGroup = this.newGroup();
    GroupBackend.addGroup(newGroup)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/groups/${newGroup.owner}/${newGroup.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteGroup(i) {
    GroupBackend.deleteGroup(this.state.data[i])
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

  renderTable(groups) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "120px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/groups/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "120px",
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
        width: "150px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:Updated time"),
        dataIndex: "updatedTime",
        key: "updatedTime",
        width: "150px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("Group:Type"),
        dataIndex: "type",
        key: "type",
        width: "110px",
        sorter: true,
        filterMultiple: false,
        filters: [
          {text: i18next.t("Group:Virtual"), value: "Virtual"},
          {text: i18next.t("Group:Physical"), value: "Physical"},
        ],
        render: (text, record, index) => {
          return i18next.t(`group:${text}`);
        },
      },
      {
        title: i18next.t("Group:Superior group"),
        dataIndex: "parentGroupId",
        key: "parentGroupId",
        width: "110px",
        sorter: true,
        ...this.getColumnSearchProps("parentGroupId"),
        render: (text, record, index) => {
          return groups.filter((group) => group.id === text)[0]?.displayName;
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
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/groups/${record.owner}/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <PopconfirmModal
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteGroup(index)}
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={groups} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Groups")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addGroup.bind(this)}>{i18next.t("general:Add")}</Button>
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
    if (params.category !== undefined && params.category !== null) {
      field = "category";
      value = params.category;
    } else if (params.type !== undefined && params.type !== null) {
      field = "type";
      value = params.type;
    }
    this.setState({loading: true});
    GroupBackend.getGroups(Setting.isAdminUser(this.props.account) ? "" : this.props.account.owner, false, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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
          if (Setting.isResponseDenied(res)) {
            this.setState({
              loading: false,
              isAuthorized: false,
            });
          }
        }
      })
      .catch(error => {
        this.setState({
          loading: false,
        });
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };
}

export default GroupListPage;
