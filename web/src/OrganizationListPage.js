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
import {Button, Col, Popconfirm, Row, Table} from 'antd';
import moment from "moment";
import * as Setting from "./Setting";
import * as OrganizationBackend from "./backend/OrganizationBackend";

class OrganizationListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizations: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrganizations();
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: res,
        });
      });
  }

  newOrganization() {
    return {
      owner: "admin", // this.props.account.organizationname,
      name: `organization_${this.state.organizations.length}`,
      createdTime: moment().format(),
      displayName: `New Organization - ${this.state.organizations.length}`,
      websiteUrl: "https://door.casbin.com",
      favicon: "https://cdn.casbin.com/static/favicon.ico",
      passwordType: "plain",
      PasswordSalt: "",
      phonePrefix: "86",
      defaultAvatar: "https://casbin.org/img/casbin.svg",
    }
  }

  addOrganization() {
    const newOrganization = this.newOrganization();
    OrganizationBackend.addOrganization(newOrganization)
      .then((res) => {
          Setting.showMessage("success", `Organization added successfully`);
          this.setState({
            organizations: Setting.prependRow(this.state.organizations, newOrganization),
          });
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
        title: Setting.I18n("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '120px',
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
        title: Setting.I18n("general:Created time"),
        dataIndex: 'createdTime',
        key: 'createdTime',
        width: '160px',
        sorter: (a, b) => a.createdTime.localeCompare(b.createdTime),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: Setting.I18n("general:Display name"),
        dataIndex: 'displayName',
        key: 'displayName',
        // width: '100px',
        sorter: (a, b) => a.displayName.localeCompare(b.displayName),
      },
      {
        title: Setting.I18n("organization:Favicon"),
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
        title: Setting.I18n("organization:Website URL"),
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
        title: Setting.I18n("general:Password type"),
        dataIndex: 'passwordType',
        key: 'passwordType',
        width: '150px',
        sorter: (a, b) => a.passwordType.localeCompare(b.passwordType),
      },
      {
        title: Setting.I18n("general:Password salt"),
        dataIndex: 'passwordSalt',
        key: 'passwordSalt',
        width: '150px',
        sorter: (a, b) => a.passwordSalt.localeCompare(b.passwordSalt),
      },
      {
        title: Setting.I18n("organization:Default avatar"),
        dataIndex: 'defaultAvatar',
        key: 'defaultAvatar',
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
        title: Setting.I18n("general:Action"),
        dataIndex: '',
        key: 'op',
        width: '240px',
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/organizations/${record.name}/users`)}>{Setting.I18n("general:Users")}</Button>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} onClick={() => this.props.history.push(`/organizations/${record.name}`)}>{Setting.I18n("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete organization: ${record.name} ?`}
                onConfirm={() => this.deleteOrganization(index)}
              >
                <Button style={{marginBottom: '10px'}} type="danger">{Setting.I18n("general:Delete")}</Button>
              </Popconfirm>
            </div>
          )
        }
      },
    ];

    return (
      <div>
        <Table columns={columns} dataSource={organizations} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                  {Setting.I18n("general:Organizations")}&nbsp;&nbsp;&nbsp;&nbsp;
                  <Button type="primary" size="small" onClick={this.addOrganization.bind(this)}>{Setting.I18n("general:Add")}</Button>
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
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.renderTable(this.state.organizations)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default OrganizationListPage;
