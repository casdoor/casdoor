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
import {Button, Col, List, Row, Table, Tooltip} from "antd";
import {EditOutlined} from "@ant-design/icons";
import moment from "moment";
import * as Setting from "./Setting";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";
import {SignupTableDefaultCssMap} from "./table/SignupTable";

class ApplicationListPage extends BaseListPage {
  constructor(props) {
    super(props);
  }

  newApplication() {
    const randomName = Setting.getRandomName();
    const organizationName = Setting.getRequestOrganization(this.props.account);
    return {
      owner: "admin", // this.props.account.applicationName,
      name: `application_${randomName}`,
      organization: organizationName,
      createdTime: moment().format(),
      displayName: `New Application - ${randomName}`,
      logo: `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`,
      enablePassword: true,
      enableSignUp: true,
      disableSignin: false,
      enableSigninSession: false,
      enableCodeSignin: false,
      enableSamlCompress: false,
      providers: [
        {name: "provider_captcha_default", canSignUp: false, canSignIn: false, canUnlink: false, prompted: false, signupGroup: "", rule: ""},
      ],
      SigninMethods: [
        {name: "Password", displayName: "Password", rule: "All"},
        {name: "Verification code", displayName: "Verification code", rule: "All"},
        {name: "WebAuthn", displayName: "WebAuthn", rule: "None"},
        {name: "Face ID", displayName: "Face ID", rule: "None"},
      ],
      signupItems: [
        {name: "ID", visible: false, required: true, rule: "Random"},
        {name: "Username", visible: true, required: true, rule: "None"},
        {name: "Display name", visible: true, required: true, rule: "None"},
        {name: "Password", visible: true, required: true, rule: "None"},
        {name: "Confirm password", visible: true, required: true, rule: "None"},
        {name: "Email", visible: true, required: true, rule: "Normal"},
        {name: "Phone", visible: true, required: true, rule: "None"},
        {name: "Agreement", visible: true, required: true, rule: "None"},
        {name: "Signup button", visible: true, required: true, rule: "None"},
        {name: "Providers", visible: true, required: true, rule: "None", customCss: SignupTableDefaultCssMap["Providers"]},
      ],
      grantTypes: ["authorization_code", "password", "client_credentials", "token", "id_token", "refresh_token"],
      cert: "cert-built-in",
      redirectUris: ["http://localhost:9000/callback"],
      tokenFormat: "JWT",
      tokenFields: [],
      expireInHours: 24 * 7,
      refreshExpireInHours: 24 * 7,
      formOffset: 2,
    };
  }

  addApplication() {
    const newApplication = this.newApplication();
    ApplicationBackend.addApplication(newApplication)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/applications/${newApplication.organization}/${newApplication.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteApplication(i) {
    ApplicationBackend.deleteApplication(this.state.data[i])
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

  renderTable(applications) {
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
            <Link to={`/applications/${record.organization}/${text}`}>
              {Setting.getApplicationDisplayName(record)}
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
        title: "Logo",
        dataIndex: "logo",
        key: "logo",
        width: "200px",
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              <img src={text} alt={text} width={150} />
            </a>
          );
        },
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("organization"),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Providers"),
        dataIndex: "providers",
        key: "providers",
        ...this.getColumnSearchProps("providers"),
        // width: '600px',
        render: (text, record, index) => {
          const providers = text;
          if (providers === null || providers.length === 0) {
            return `(${i18next.t("general:empty")})`;
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
                          <Button style={{marginRight: "5px"}} icon={<EditOutlined />} size="small" onClick={() => Setting.goToLinkSoft(this, `/providers/${record.organization}/${providerItem.name}`)} />
                        </Tooltip>
                        <Link to={`/providers/${record.organization}/${providerItem.name}`}>
                          {providerItem.name}
                        </Link>
                      </div>
                    </List.Item>
                  );
                }}
              />
            );
          };

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
          );
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
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/applications/${record.organization}/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <PopconfirmModal
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteApplication(index)}
                disabled={record.name === "app-built-in"}
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={applications} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Applications")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addApplication.bind(this)}>{i18next.t("general:Add")}</Button>
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
    (Setting.isDefaultOrganizationSelected(this.props.account) ? ApplicationBackend.getApplications("admin", params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder) :
      ApplicationBackend.getApplicationsByOrganization("admin", Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder))
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

export default ApplicationListPage;
