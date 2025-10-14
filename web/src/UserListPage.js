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
import {Button, Space, Switch, Table, Upload} from "antd";
import {UploadOutlined} from "@ant-design/icons";
import moment from "moment";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import * as UserBackend from "./backend/UserBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";
import AccountAvatar from "./account/AccountAvatar";

class UserListPage extends BaseListPage {
  constructor(props) {
    super(props);
    this.state = {
      ...this.state,
      organization: null,
    };
  }

  UNSAFE_componentWillMount() {
    super.UNSAFE_componentWillMount();
    this.getOrganization(this.state.organizationName);
  }

  componentDidUpdate(prevProps, prevState) {
    if (this.props.match.path !== prevProps.match.path || this.props.organizationName !== prevProps.organizationName) {
      this.setState({
        organizationName: this.props.organizationName ?? this.props.match?.params.organizationName,
      });
    }

    if (this.state.organizationName !== prevState.organizationName) {
      this.getOrganization(this.state.organizationName);
    }

    if (prevProps.groupName !== this.props.groupName || this.state.organizationName !== prevState.organizationName) {
      this.fetch({
        pagination: this.state.pagination,
        searchText: this.state.searchText,
        searchedColumn: this.state.searchedColumn,
      });
    }
  }

  newUser() {
    const randomName = Setting.getRandomName();
    const owner = (Setting.isDefaultOrganizationSelected(this.props.account) || this.props.groupName) ? this.state.organizationName : Setting.getRequestOrganization(this.props.account);
    return {
      owner: owner,
      name: `user_${randomName}`,
      createdTime: moment().format(),
      type: "normal-user",
      password: "123",
      passwordSalt: "",
      displayName: `New User - ${randomName}`,
      avatar: this.state.organization.defaultAvatar ?? `${Setting.StaticBaseUrl}/img/casbin.svg`,
      email: `${randomName}@example.com`,
      phone: Setting.getRandomNumber(),
      countryCode: this.state.organization.countryCodes?.length > 0 ? this.state.organization.countryCodes[0] : "",
      address: [],
      groups: this.props.groupName ? [`${owner}/${this.props.groupName}`] : [],
      affiliation: "Example Inc.",
      tag: "staff",
      region: "",
      isAdmin: (owner === "built-in"),
      IsForbidden: false,
      score: this.state.organization.initScore,
      isDeleted: false,
      properties: {},
      signupApplication: this.state.organization.defaultApplication,
      registerType: "Add User",
      registerSource: `${this.props.account.owner}/${this.props.account.name}`,
    };
  }

  addUser() {
    const newUser = this.newUser();
    UserBackend.addUser(newUser)
      .then((res) => {
        if (res.status === "ok") {
          sessionStorage.setItem("userListUrl", window.location.pathname);
          this.props.history.push({pathname: `/users/${newUser.owner}/${newUser.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteUser(i) {
    UserBackend.deleteUser(this.state.data[i])
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

  removeUserFromGroup(i) {
    const user = this.state.data[i];
    const group = this.props.groupName;
    UserBackend.removeUserFromGroup({groupName: group, owner: user.owner, name: user.name})
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully removed"));
          this.setState({
            data: Setting.deleteRow(this.state.data, i),
            pagination: {total: this.state.pagination.total - 1},
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to remove")}: ${res.msg}`);
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
        Setting.showMessage("success", "Users uploaded successfully, refreshing the page");

        const {pagination} = this.state;
        this.fetch({pagination});
      } else {
        Setting.showMessage("error", `Users failed to upload: ${res.msg}`);
      }
    } else if (status === "error") {
      Setting.showMessage("error", "File failed to upload");
    }
  }

  getOrganization(organizationName) {
    OrganizationBackend.getOrganization("admin", organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            organization: res.data,
          });
        } else {
          Setting.showMessage("error", `Failed to get organization: ${res.msg}`);
        }
      });
  }

  renderUpload() {
    const props = {
      name: "file",
      accept: ".xlsx",
      method: "post",
      action: `${Setting.ServerUrl}/api/upload-users`,
      withCredentials: true,
      onChange: (info) => {
        this.uploadFile(info);
      },
    };

    return (
      <Upload {...props}>
        <Button icon={<UploadOutlined />} id="upload-button" size="small">
          {i18next.t("user:Upload (.xlsx)")}
        </Button>
      </Upload>
    );
  }

  renderTable(users) {
    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: (Setting.isMobile()) ? "100px" : "120px",
        fixed: "left",
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
        title: i18next.t("general:Application"),
        dataIndex: "signupApplication",
        key: "signupApplication",
        width: (Setting.isMobile()) ? "100px" : "120px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("signupApplication"),
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: (Setting.isMobile()) ? "80px" : "110px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/users/${record.owner}/${text}`}>
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
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        // width: '100px',
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("general:Avatar"),
        dataIndex: "avatar",
        key: "avatar",
        width: "80px",
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              <AccountAvatar referrerPolicy="no-referrer" src={text} alt={text} size={50} />
            </a>
          );
        },
      },
      {
        title: i18next.t("general:Email"),
        dataIndex: "email",
        key: "email",
        width: "160px",
        sorter: true,
        ...this.getColumnSearchProps("email"),
        render: (text, record, index) => {
          return (
            <a href={`mailto:${text}`}>
              {text}
            </a>
          );
        },
      },
      {
        title: i18next.t("general:Phone"),
        dataIndex: "phone",
        key: "phone",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("phone"),
      },
      {
        title: i18next.t("user:Affiliation"),
        dataIndex: "affiliation",
        key: "affiliation",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("affiliation"),
      },
      {
        title: i18next.t("user:Country/Region"),
        dataIndex: "region",
        key: "region",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("region"),
        render: (text, record, index) => {
          return Setting.initCountries().getName(record.region, Setting.getLanguage(), {select: "official"});
        },
      },
      {
        title: i18next.t("general:User type"),
        dataIndex: "type",
        key: "type",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("type"),
      },
      {
        title: i18next.t("user:Tag"),
        dataIndex: "tag",
        key: "tag",
        width: "110px",
        sorter: true,
        ...this.getColumnSearchProps("tag"),
        render: (text, record, index) => {
          if (this.state.organization?.tags?.length === 0) {
            return text;
          }

          const tagMap = {};
          this.state.organization?.tags?.map((tag, index) => {
            const tokens = tag.split("|");
            const displayValue = Setting.getLanguage() !== "zh" ? tokens[0] : tokens[1];
            tagMap[tokens[0]] = displayValue;
          });
          return tagMap[text];
        },
      },
      {
        title: i18next.t("user:Register type"),
        dataIndex: "registerType",
        key: "registerType",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("registerType"),
      },
      {
        title: i18next.t("user:Register source"),
        dataIndex: "registerSource",
        key: "registerSource",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("registerSource"),
      },
      {
        title: i18next.t("user:Is admin"),
        dataIndex: "isAdmin",
        key: "isAdmin",
        width: "120px",
        sorter: true,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          );
        },
      },
      {
        title: i18next.t("user:Is forbidden"),
        dataIndex: "isForbidden",
        key: "isForbidden",
        width: "110px",
        sorter: true,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          );
        },
      },
      {
        title: i18next.t("user:Is deleted"),
        dataIndex: "isDeleted",
        key: "isDeleted",
        width: "110px",
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
        width: "190px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          const isTreePage = this.props.groupName !== undefined;
          const disabled = (record.owner === this.props.account.owner && record.name === this.props.account.name) || (record.owner === "built-in" && record.name === "admin");
          return (
            <Space>
              <Button size={isTreePage ? "small" : "middle"} type="primary" onClick={() => {
                sessionStorage.setItem("userListUrl", window.location.pathname);
                this.props.history.push(`/users/${record.owner}/${record.name}`);
              }}>{i18next.t("general:Edit")}
              </Button>
              {isTreePage ?
                <PopconfirmModal
                  text={i18next.t("general:remove")}
                  title={i18next.t("general:Sure to remove") + `: ${record.name} ?`}
                  onConfirm={() => this.removeUserFromGroup(index)}
                  disabled={disabled}
                  size="small"
                /> : null}
              <PopconfirmModal
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteUser(index)}
                disabled={disabled}
                size={isTreePage ? "small" : "default"}
              />
            </Space>
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
      <div>
        <Table scroll={{x: "max-content"}} columns={filteredColumns} dataSource={users} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Users")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button style={{marginRight: "15px"}} type="primary" size="small" onClick={this.addUser.bind(this)}>{i18next.t("general:Add")} </Button>
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
    const field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    this.setState({loading: true});
    if (this.props.match?.path === "/users") {
      (Setting.isDefaultOrganizationSelected(this.props.account) ? UserBackend.getGlobalUsers(params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder) : UserBackend.getUsers(Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder))
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
    } else {
      (this.props.groupName ?
        UserBackend.getUsers(this.state.organizationName, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder, this.props.groupName) :
        UserBackend.getUsers(this.state.organizationName, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder))
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
    }
  };
}

export default UserListPage;
