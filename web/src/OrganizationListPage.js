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
import * as OrganizationBackend from "./backend/OrganizationBackend";
import i18next from "i18next";

class OrganizationListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizations: null,
      total: 0
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrganizations(1, 10);
  }

  getOrganizations(page, pageSize) {
    OrganizationBackend.getOrganizations("admin", page, pageSize)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            organizations: res.data,
            total: res.data2
          });
        }
      });
  }

  newOrganization() {
    var randomName = Math.random().toString(36).slice(-6)
    return {
      owner: "admin", // this.props.account.organizationname,
      name: `organization_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Organization - ${randomName}`,
      websiteUrl: "https://door.casbin.com",
      favicon: "https://cdn.casbin.com/static/favicon.ico",
      passwordType: "plain",
      PasswordSalt: "",
      phonePrefix: "86",
      defaultAvatar: "https://casbin.org/img/casbin.svg",
      masterPassword: "",
      enableSoftDeletion: false,
    }
  }

  addOrganization() {
    const newOrganization = this.newOrganization();
    OrganizationBackend.addOrganization(newOrganization)
      .then((res) => {
          Setting.showMessage("success", `Organization added successfully`);
          this.setState({
            organizations: Setting.prependRow(this.state.organizations, newOrganization),
            total: this.state.total + 1
          });
          this.props.history.push(`/organizations/${newOrganization.name}`);
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Organization failed to add: ${error}`);
      });
  }

  deleteOrganization(i) {
    OrganizationBackend.deleteOrganization(this.state.organizations[i])
      .then((res) => {
          Setting.showMessage("success", `Organization deleted successfully`);
          this.setState({
            organizations: Setting.deleteRow(this.state.organizations, i),
            total: this.state.total - 1
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Organization failed to delete: ${error}`);
      });
  }

  renderTable(organizations) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '120px',
        fixed: 'left',
        sorter: (a, b) => a.name.localeCompare(b.name),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
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
        // width: '100px',
        sorter: (a, b) => a.displayName.localeCompare(b.displayName),
      },
      {
        title: i18next.t("organization:Favicon"),
        dataIndex: 'favicon',
        key: 'favicon',
        width: '50px',
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              <img src={text} alt={text} width={40} />
            </a>
          )
        }
      },
      {
        title: i18next.t("organization:Website URL"),
        dataIndex: 'websiteUrl',
        key: 'websiteUrl',
        width: '300px',
        sorter: (a, b) => a.websiteUrl.localeCompare(b.websiteUrl),
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              {text}
            </a>
          )
        }
      },
      {
        title: i18next.t("general:Password type"),
        dataIndex: 'passwordType',
        key: 'passwordType',
        width: '150px',
        sorter: (a, b) => a.passwordType.localeCompare(b.passwordType),
      },
      {
        title: i18next.t("general:Password salt"),
        dataIndex: 'passwordSalt',
        key: 'passwordSalt',
        width: '150px',
        sorter: (a, b) => a.passwordSalt.localeCompare(b.passwordSalt),
      },
      {
        title: i18next.t("organization:Default avatar"),
        dataIndex: 'defaultAvatar',
        key: 'defaultAvatar',
        width: '120px',
        render: (text, record, index) => {
          return (
              <a target="_blank" rel="noreferrer" href={text}>
                <img src={text} alt={text} width={40} />
              </a>
          )
        }
      },
      {
        title: i18next.t("organization:Soft deletion"),
        dataIndex: 'enableSoftDeletion',
        key: 'enableSoftDeletion',
        width: '140px',
        sorter: (a, b) => a.enableSoftDeletion - b.enableSoftDeletion,
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
        width: '240px',
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/organizations/${record.name}/users`)}>{i18next.t("general:Users")}</Button>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} onClick={() => this.props.history.push(`/organizations/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete organization: ${record.name} ?`}
                onConfirm={() => this.deleteOrganization(index)}
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
      onChange: (page, pageSize) => this.getOrganizations(page, pageSize),
      onShowSizeChange: (current, size) => this.getOrganizations(current, size),
    };

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={organizations} rowKey="name" size="middle" bordered pagination={paginationProps}
               title={() => (
                 <div>
                  {i18next.t("general:Organizations")}&nbsp;&nbsp;&nbsp;&nbsp;
                  <Button type="primary" size="small" onClick={this.addOrganization.bind(this)}>{i18next.t("general:Add")}</Button>
                 </div>
               )}
               loading={organizations === null}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.state.organizations)
        }
      </div>
    );
  }
}

export default OrganizationListPage;
