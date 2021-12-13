// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import {Button, Popconfirm, Switch, Table} from 'antd';
import moment from "moment";
import * as Setting from "./Setting";
import * as UserBackend from "./backend/UserBackend";
import i18next from "i18next";

class UserListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      users: null,
      organizationName: props.match.params.organizationName,
      total: 0,
    };
  }

  UNSAFE_componentWillMount() {
    this.getUsers(1, 10);
  }

  getUsers(page, pageSize) {
    if (this.state.organizationName === undefined) {
      UserBackend.getGlobalUsers(page, pageSize)
          .then((res) => {
            if (res.status === "ok") {
              this.setState({
                users: res.data,
                total: res.data2
              });
            }
          });
    } else {
      UserBackend.getUsers(this.state.organizationName, page, pageSize)
          .then((res) => {
            if (res.status === "ok") {
              this.setState({
                users: res.data,
                total: res.data2
              });
            }
          });
    }
  }

  newUser() {
    const randomName = Setting.getRandomName();
    return {
      owner: "built-in", // this.props.account.username,
      name: `user_${randomName}`,
      createdTime: moment().format(),
      type: "normal-user",
      password: "123",
      passwordSalt: "",
      displayName: `New User - ${randomName}`,
      avatar: "https://casbin.org/img/casbin.svg",
      email: "user@example.com",
      phone: "12345678",
      address: [],
      affiliation: "Example Inc.",
      tag: "staff",
      region: "",
      isAdmin: false,
      isGlobalAdmin: false,
      IsForbidden: false,
      isDeleted: false,
      properties: {},
      signupApplication: "app-built-in",
    }
  }

  addUser() {
    const newUser = this.newUser();
    UserBackend.addUser(newUser)
      .then((res) => {
          Setting.showMessage("success", `User added successfully`);
          this.setState({
            users: Setting.prependRow(this.state.users, newUser),
            total: this.state.total + 1
          });
          this.props.history.push(`/users/${newUser.owner}/${newUser.name}`);
        }
      )
      .catch(error => {
        Setting.showMessage("error", `User failed to add: ${error}`);
      });
  }

  deleteUser(i) {
    UserBackend.deleteUser(this.state.users[i])
      .then((res) => {
          Setting.showMessage("success", `User deleted successfully`);
          this.setState({
            users: Setting.deleteRow(this.state.users, i),
            total: this.state.total - 1
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `User failed to delete: ${error}`);
      });
  }

  renderTable(users) {
    // transfer country code to name based on selected language
    var countries = require("i18n-iso-countries");
    countries.registerLocale(require("i18n-iso-countries/langs/" + i18next.language + ".json"));
    for (var index in users) {
      users[index].region = countries.getName(users[index].region, i18next.language, {select: "official"})
    }

    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: 'owner',
        key: 'owner',
        width: (Setting.isMobile()) ? "100px" : "120px",
        fixed: 'left',
        sorter: (a, b) => a.owner.localeCompare(b.owner),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: 'signupApplication',
        key: 'signupApplication',
        width: (Setting.isMobile()) ? "100px" : "120px",
        fixed: 'left',
        sorter: (a, b) => a.owner.localeCompare(b.owner),
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: (Setting.isMobile()) ? "80px" : "100px",
        fixed: 'left',
        sorter: (a, b) => a.name.localeCompare(b.name),
        render: (text, record, index) => {
          return (
            <Link to={`/users/${record.owner}/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: 'createdTime',
        key: 'createdTime',
        width: '160px',
        sorter: (a, b) => a.createdTime.localeCompare(b.createdTime),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: 'displayName',
        key: 'displayName',
        width: '100px',
        sorter: (a, b) => a.displayName.localeCompare(b.displayName),
      },
      {
        title: i18next.t("general:Avatar"),
        dataIndex: 'avatar',
        key: 'avatar',
        width: '80px',
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              <img src={text} alt={text} width={50} />
            </a>
          )
        }
      },
      {
        title: i18next.t("general:Email"),
        dataIndex: 'email',
        key: 'email',
        width: '160px',
        sorter: (a, b) => a.email.localeCompare(b.email),
        render: (text, record, index) => {
          return (
            <a href={`mailto:${text}`}>
              {text}
            </a>
          )
        }
      },
      {
        title: i18next.t("general:Phone"),
        dataIndex: 'phone',
        key: 'phone',
        width: '120px',
        sorter: (a, b) => a.phone.localeCompare(b.phone),
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
        dataIndex: 'affiliation',
        key: 'affiliation',
        width: '120px',
        sorter: (a, b) => a.affiliation.localeCompare(b.affiliation),
      },
      {
        title: i18next.t("user:Country/Region"),
        dataIndex: 'region',
        key: 'region',
        width: '120px',
        sorter: (a, b) => a.region.localeCompare(b.region),
      },
      {
        title: i18next.t("user:Tag"),
        dataIndex: 'tag',
        key: 'tag',
        width: '100px',
        sorter: (a, b) => a.tag.localeCompare(b.tag),
      },
      {
        title: i18next.t("user:Is admin"),
        dataIndex: 'isAdmin',
        key: 'isAdmin',
        width: '110px',
        sorter: (a, b) => a.isAdmin - b.isAdmin,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
        }
      },
      {
        title: i18next.t("user:Is global admin"),
        dataIndex: 'isGlobalAdmin',
        key: 'isGlobalAdmin',
        width: '110px',
        sorter: (a, b) => a.isGlobalAdmin - b.isGlobalAdmin,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
        }
      },
      {
        title: i18next.t("user:Is forbidden"),
        dataIndex: 'isForbidden',
        key: 'isForbidden',
        width: '110px',
        sorter: (a, b) => a.isForbidden - b.isForbidden,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
        }
      },
      {
        title: i18next.t("user:Is deleted"),
        dataIndex: 'isDeleted',
        key: 'isDeleted',
        width: '110px',
        sorter: (a, b) => a.isDeleted - b.isDeleted,
        render: (text, record, index) => {
          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
        }
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: '',
        key: 'op',
        width: '190px',
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/users/${record.owner}/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete user: ${record.name} ?`}
                onConfirm={() => this.deleteUser(index)}
              >
                <Button style={{marginBottom: '10px'}} type="danger">{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
          )
        }
      },
    ];

    const paginationProps = {
      total: this.state.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.total),
      onChange: (page, pageSize) => this.getUsers(page, pageSize),
      onShowSizeChange: (current, size) => this.getUsers(current, size),
    };

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={users} rowKey="name" size="middle" bordered pagination={paginationProps}
               title={() => (
                 <div>
                  {i18next.t("general:Users")}&nbsp;&nbsp;&nbsp;&nbsp;
                  <Button type="primary" size="small" onClick={this.addUser.bind(this)}>{i18next.t("general:Add")}</Button>
                 </div>
               )}
               loading={users === null}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.state.users)
        }
      </div>
    );
  }
}

export default UserListPage;
