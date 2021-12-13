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
import {Button, Col, List, Popconfirm, Row, Table, Tooltip} from 'antd';
import {EditOutlined} from "@ant-design/icons";
import moment from "moment";
import * as Setting from "./Setting";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import i18next from "i18next";

class ApplicationListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applications: null,
      total: 0,
    };
  }

  UNSAFE_componentWillMount() {
    this.getApplications(1, 10);
  }

  getApplications(page, pageSize) {
    ApplicationBackend.getApplications("admin", page, pageSize)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            applications: res.data,
            total: res.data2
          });
        }
      });
  }

  newApplication() {
    const randomName = Setting.getRandomName();
    return {
      owner: "admin", // this.props.account.applicationname,
      name: `application_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Application - ${randomName}`,
      logo: "https://cdn.casbin.com/logo/logo_1024x256.png",
      enablePassword: true,
      enableSignUp: true,
      EnableCodeSignin: false,
      providers: [],
      signupItems: [
        {name: "ID", visible: false, required: true, rule: "Random"},
        {name: "Username", visible: true, required: true, rule: "None"},
        {name: "Display name", visible: true, required: true, rule: "None"},
        {name: "Password", visible: true, required: true, rule: "None"},
        {name: "Confirm password", visible: true, required: true, rule: "None"},
        {name: "Email", visible: true, required: true, rule: "None"},
        {name: "Phone", visible: true, required: true, rule: "None"},
        {name: "Agreement", visible: true, required: true, rule: "None"},
      ],
      redirectUris: ["http://localhost:9000/callback"],
      expireInHours: 24 * 7,
    }
  }

  addApplication() {
    const newApplication = this.newApplication();
    ApplicationBackend.addApplication(newApplication)
      .then((res) => {
          Setting.showMessage("success", `Application added successfully`);
          this.setState({
            applications: Setting.prependRow(this.state.applications, newApplication),
            total: this.state.total + 1
          });
          this.props.history.push(`/applications/${newApplication.name}`);
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
            total: this.state.total - 1
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
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '150px',
        fixed: 'left',
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
        title: 'Logo',
        dataIndex: 'logo',
        key: 'logo',
        width: '200px',
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              <img src={text} alt={text} width={150} />
            </a>
          )
        }
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: 'organization',
        key: 'organization',
        width: '150px',
        sorter: (a, b) => a.organization.localeCompare(b.organization),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Providers"),
        dataIndex: 'providers',
        key: 'providers',
        // width: '600px',
        render: (text, record, index) => {
          const providers = text;
          if (providers.length === 0) {
            return "(empty)";
          }

          const half = Math.floor((providers.length + 1) / 2);

          const getList = (providers) => {
            return (
              <List
                size="small"
                locale={{emptyText: " "}}
                dataSource={providers}
                renderItem={(providerItem, i) => {
                  return (
                    <List.Item>
                      <div style={{display: "inline"}}>
                        <Tooltip placement="topLeft" title="Edit">
                          <Button style={{marginRight: "5px"}} icon={<EditOutlined />} size="small" onClick={() => Setting.goToLinkSoft(this, `/providers/${providerItem.name}`)} />
                        </Tooltip>
                        <Link to={`/providers/${providerItem.name}`}>
                          {providerItem.name}
                        </Link>
                      </div>
                    </List.Item>
                  )
                }}
              />
            )
          }

          return (
            <div>
              <Row>
                <Col span={12}>
                  {
                    getList(providers.slice(0, half))
                  }
                </Col>
                <Col span={12}>
                  {
                    getList(providers.slice(half))
                  }
                </Col>
              </Row>
            </div>
          )
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: '',
        key: 'op',
        width: '170px',
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/applications/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete application: ${record.name} ?`}
                onConfirm={() => this.deleteApplication(index)}
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
      onChange: (page, pageSize) => this.getApplications(page, pageSize),
      onShowSizeChange: (current, size) => this.getApplications(current, size),
    };

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={applications} rowKey="name" size="middle" bordered pagination={paginationProps}
               title={() => (
                 <div>
                  {i18next.t("general:Applications")}&nbsp;&nbsp;&nbsp;&nbsp;
                  <Button type="primary" size="small" onClick={this.addApplication.bind(this)}>{i18next.t("general:Add")}</Button>
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
        {
          this.renderTable(this.state.applications)
        }
      </div>
    );
  }
}

export default ApplicationListPage;
