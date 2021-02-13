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
import * as ApplicationBackend from "./backend/ApplicationBackend";

class ApplicationListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applications: null,
    };
  }

  componentWillMount() {
    this.getApplications();
  }

  getApplications() {
    ApplicationBackend.getApplications("admin")
      .then((res) => {
        this.setState({
          applications: res,
        });
      });
  }

  newApplication() {
    return {
      owner: "admin", // this.props.account.applicationname,
      name: `application_${this.state.applications.length}`,
      createdTime: moment().format(),
      displayName: `New Application - ${this.state.applications.length}`,
      logo: "https://cdn.jsdelivr.net/gh/casbin/static/img/logo@2x.png",
      providers: [],
    }
  }

  addApplication() {
    const newApplication = this.newApplication();
    ApplicationBackend.addApplication(newApplication)
      .then((res) => {
          Setting.showMessage("success", `Application added successfully`);
          this.setState({
            applications: Setting.prependRow(this.state.applications, newApplication),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Application failed to add: ${error}`);
      });
  }

  deleteApplication(i) {
    ApplicationBackend.deleteApplication(this.state.applications[i])
      .then((res) => {
          Setting.showMessage("success", `Application deleted successfully`);
          this.setState({
            applications: Setting.deleteRow(this.state.applications, i),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Application failed to delete: ${error}`);
      });
  }

  renderTable(applications) {
    const columns = [
      {
        title: 'Name',
        dataIndex: 'name',
        key: 'name',
        width: '150px',
        sorter: (a, b) => a.name.localeCompare(b.name),
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: 'Created Time',
        dataIndex: 'createdTime',
        key: 'createdTime',
        width: '160px',
        sorter: (a, b) => a.createdTime.localeCompare(b.createdTime),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: 'Display Name',
        dataIndex: 'displayName',
        key: 'displayName',
        // width: '100px',
        sorter: (a, b) => a.displayName.localeCompare(b.displayName),
      },
      {
        title: 'Logo',
        dataIndex: 'logo',
        key: 'logo',
        width: '250px',
        render: (text, record, index) => {
          return (
            <a target="_blank" href={text}>
              <img src={text} alt={text} width={150} />
            </a>
          )
        }
      },
      {
        title: 'Organization',
        dataIndex: 'organization',
        key: 'organization',
        width: '200px',
        sorter: (a, b) => a.organization.localeCompare(b.organization),
        render: (text, record, index) => {
          return (
            <a href={`/organizations/${text}`}>
              {text}
            </a>
          )
        }
      },
      {
        title: 'Providers',
        dataIndex: 'providers',
        key: 'providers',
        width: '200px',
        sorter: (a, b) => a.providers.localeCompare(b.providers),
        render: (text, record, index) => {
          return (
            <a href={`/providers/${text}`}>
              {text}
            </a>
          )
        }
      },
      {
        title: 'Action',
        dataIndex: '',
        key: 'op',
        width: '170px',
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/applications/${record.name}`)}>Edit</Button>
              <Popconfirm
                title={`Sure to delete application: ${record.name} ?`}
                onConfirm={() => this.deleteApplication(index)}
              >
                <Button style={{marginBottom: '10px'}} type="danger">Delete</Button>
              </Popconfirm>
            </div>
          )
        }
      },
    ];

    return (
      <div>
        <Table columns={columns} dataSource={applications} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   Applications&nbsp;&nbsp;&nbsp;&nbsp;
                   <Button type="primary" size="small" onClick={this.addApplication.bind(this)}>Add</Button>
                 </div>
               )}
               loading={applications === null}
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
              this.renderTable(this.state.applications)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default ApplicationListPage;
