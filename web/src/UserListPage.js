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
import {Button, Popconfirm, Switch, Table, Upload} from "antd";
import {UploadOutlined} from "@ant-design/icons";
import moment from "moment";
import * as Setting from "./Setting";
import * as UserBackend from "./backend/UserBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";

class UserListPage extends BaseListPage {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.match.params.organizationName,
      data: [],
      pagination: {
        current: 1,
        pageSize: 10,
      },
      loading: false,
      searchText: "",
      searchedColumn: "",
    };
  }

  newUser() {
    const randomName = Setting.getRandomName();
    const owner = (this.state.organizationName !== undefined) ? this.state.organizationName : this.props.account.owner;
    return {
      owner: owner,
      name: `user-${randomName}`,
      createdTime: moment().format(),
      type: "normal-user",
      password: "123",
      passwordSalt: "",
      displayName: `New User - ${randomName}`,
      avatar: `${Setting.StaticBaseUrl}/img/casbin.svg`,
      email: `${randomName}@example.com`,
      phone: Setting.getRandomNumber(),
      address: [],
      affiliation: "Example Inc.",
      tag: "staff",
      region: "",
      isAdmin: (owner === "built-in"),
      isGlobalAdmin: (owner === "built-in"),
      IsForbidden: false,
      isDeleted: false,
      properties: {},
      signupApplication: "app-built-in",
    };
  }

  addUser() {
    const newUser = this.newUser();
    UserBackend.addUser(newUser)
      .then((res) => {
        this.props.history.push({pathname: `/users/${newUser.owner}/${newUser.name}`, mode: "add"});
      }
      )
      .catch(error => {
        Setting.showMessage("error", `User failed to add: ${error}`);
      });
  }

  deleteUser(i) {
    UserBackend.deleteUser(this.state.data[i])
      .then((res) => {
        Setting.showMessage("success", "User deleted successfully");
        this.setState({
          data: Setting.deleteRow(this.state.data, i),
          pagination: {total: this.state.pagination.total - 1},
        });
      }
      )
      .catch(error => {
        Setting.showMessage("error", `User failed to delete: ${error}`);
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
        <Button type="primary" size="small">
          <UploadOutlined /> {i18next.t("user:Upload (.xlsx)")}
        </Button>
      </Upload>
    );
  }

  renderTable(users) {
    // transfer country code to name based on selected language
    const countries = require("i18n-iso-countries");
    countries.registerLocale(require("i18n-iso-countries/langs/" + i18next.language + ".json"));
    for (const index in users) {
      users[index].region = countries.getName(users[index].region, i18next.language, {select: "official"});
    }

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
            <Link to={`/applications/${text}`}>
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
              <img src={text} alt={text} width={50} />
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
      // {
      //   title: 'Phone',
      //   dataIndex: 'phone',
      //   key: 'phone',
      //   width: '120px',
      //   sorter: (a, b) => a.phone.localeCompare(b.phone),
      // },
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
      },
      {
        title: i18next.t("user:Tag"),
        dataIndex: "tag",
        key: "tag",
        width: "110px",
        sorter: true,
        ...this.getColumnSearchProps("tag"),
      },
      {
        title: i18next.t("user:Is admin"),
        dataIndex: "isAdmin",
        key: "isAdmin",
        width: "110px",
        sorter: true,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          );
        },
      },
      {
        title: i18next.t("user:Is global admin"),
        dataIndex: "isGlobalAdmin",
        key: "isGlobalAdmin",
        width: "140px",
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
          const disabled = (record.owner === this.props.account.owner && record.name === this.props.account.name);
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/users/${record.owner}/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete user: ${record.name} ?`}
                onConfirm={() => this.deleteUser(index)}
              >
                <Button disabled={disabled} style={{marginBottom: "10px"}} type="danger">{i18next.t("general:Delete")}</Button>
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={users} rowKey="name" size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Users")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button style={{marginRight: "5px"}} type="primary" size="small" onClick={this.addUser.bind(this)}>{i18next.t("general:Add")}</Button>
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
    if (this.state.organizationName === undefined) {
      (Setting.isAdminUser(this.props.account) ? UserBackend.getGlobalUsers(params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder) : UserBackend.getUsers(this.props.account.owner, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder))
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
    } else {
      UserBackend.getUsers(this.state.organizationName, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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
    }
  };
}

export default UserListPage;
