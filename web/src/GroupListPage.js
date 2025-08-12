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
import {Button, Table, Tooltip, Upload} from "antd";
import {UploadOutlined} from "@ant-design/icons";
import moment from "moment";
import * as Setting from "./Setting";
import * as GroupBackend from "./backend/GroupBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class GroupListPage extends BaseListPage {
  constructor(props) {
    super(props);
    this.state = {
      ...this.state,
      owner: Setting.isAdminUser(this.props.account) ? "" : this.props.account.owner,
      groups: [],
    };
  }
  UNSAFE_componentWillMount() {
    super.UNSAFE_componentWillMount();
  }

  newGroup() {
    const randomName = Setting.getRandomName();
    const owner = Setting.getRequestOrganization(this.props.account);
    return {
      owner: owner,
      name: `group_${randomName}`,
      createdTime: moment().format(),
      updatedTime: moment().format(),
      displayName: `New Group - ${randomName}`,
      type: "Virtual",
      parentId: this.props.account.owner,
      isTopGroup: true,
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

  uploadFile(info) {
    const {status, response: res} = info.file;
    if (status === "done") {
      if (res.status === "ok") {
        Setting.showMessage("success", "Groups uploaded successfully, refreshing the page");
        const {pagination} = this.state;
        this.fetch({pagination});
      } else {
        Setting.showMessage("error", `Groups failed to upload: ${res.msg}`);
      }
    } else if (status === "error") {
      Setting.showMessage("error", "File failed to upload");
    }
  }

  renderUpload() {
    const props = {
      name: "file",
      accept: ".xlsx",
      method: "post",
      action: `${Setting.ServerUrl}/api/upload-groups`,
      withCredentials: true,
      onChange: (info) => {
        this.uploadFile(info);
      },
    };

    return (
      <Upload {...props}>
        <Button icon={<UploadOutlined />} id="upload-button" type="primary" size="small">
          {i18next.t("group:Upload (.xlsx)")}
        </Button>
      </Upload>
    );
  }

  renderTable(data) {
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
            <Link to={`/groups/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "140px",
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
        title: i18next.t("general:Updated time"),
        dataIndex: "updatedTime",
        key: "updatedTime",
        width: "180px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        // width: "200px",
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("general:Type"),
        dataIndex: "type",
        key: "type",
        width: "140px",
        sorter: true,
        filterMultiple: false,
        filters: [
          {text: i18next.t("group:Virtual"), value: "Virtual"},
          {text: i18next.t("group:Physical"), value: "Physical"},
        ],
        render: (text, record, index) => {
          return i18next.t("group:" + text);
        },
      },
      {
        title: i18next.t("group:Parent group"),
        dataIndex: "parentId",
        key: "parentId",
        width: "220px",
        sorter: true,
        ...this.getColumnSearchProps("parentId"),
        render: (text, record, index) => {
          if (record.isTopGroup) {
            return <Link to={`/organizations/${record.parentId}`}>
              {record.parentId}
            </Link>;
          }
          return <Link to={`/groups/${record.owner}/${record.parentId}`}>
            {record?.parentName}
          </Link>;
        },
      },
      {
        title: i18next.t("general:Users"),
        dataIndex: "users",
        key: "users",
        // width: "200px",
        sorter: true,
        ...this.getColumnSearchProps("users"),
        render: (text, record, index) => {
          return Setting.getTags(text, "users");
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "180px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/groups/${record.owner}/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              {
                record.haveChildren ? <Tooltip placement="topLeft" title={i18next.t("group:You need to delete all subgroups first. You can view the subgroups in the left group tree of the [Organizations] -> [Groups] page")}>
                  <Button disabled type="primary" danger>{i18next.t("general:Delete")}</Button>
                </Tooltip> :
                  <PopconfirmModal
                    title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                    onConfirm={() => this.deleteGroup(index)}
                  >
                  </PopconfirmModal>
              }
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={data} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Groups")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={this.addGroup.bind(this)}>{i18next.t("general:Add")}</Button>
              {
                this.renderUpload()
              }
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
    GroupBackend.getGroups(Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account), false, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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
