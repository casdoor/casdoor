// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
import {Col, InputNumber, Modal, Row, Select} from "antd";
import * as Setting from "../../Setting";
import * as OrganizationBackend from "../../backend/OrganizationBackend";
import * as ApplicationBackend from "../../backend/ApplicationBackend";
import i18next from "i18next";

const {Option} = Select;

class RechargeModal extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      tag: "Organization",
      amount: null,
      currency: "USD",
      application: "",
      organization: props.currentOrganization || "",
      organizations: [],
      applications: [],
    };
  }

  componentDidMount() {
    this.loadOrganizations();
    this.loadApplications();
  }

  isGlobalAdmin() {
    return Setting.isAdminUser(this.props.account);
  }

  loadOrganizations() {
    if (this.isGlobalAdmin()) {
      // Global admin can see all organizations
      OrganizationBackend.getOrganizations("admin")
        .then((res) => {
          if (res.status === "ok") {
            this.setState({
              organizations: res.data || [],
            });
          }
        });
    } else {
      // Organization admin can only see their own organization
      const currentOrg = this.props.currentOrganization;
      if (currentOrg) {
        OrganizationBackend.getOrganization("admin", currentOrg)
          .then((res) => {
            if (res.status === "ok" && res.data) {
              this.setState({
                organizations: [res.data],
              });
            }
          });
      }
    }
  }

  loadApplications() {
    if (this.isGlobalAdmin()) {
      // Global admin can see all applications
      ApplicationBackend.getApplications("admin")
        .then((res) => {
          if (res.status === "ok") {
            this.setState({
              applications: res.data || [],
            });
          }
        });
    } else {
      // Organization admin can see their organization's applications
      const currentOrg = this.props.currentOrganization;
      if (currentOrg) {
        ApplicationBackend.getApplicationsByOrganization("admin", currentOrg)
          .then((res) => {
            if (res.status === "ok") {
              this.setState({
                applications: res.data || [],
              });
            }
          });
      }
    }
  }

  handleOk = () => {
    const {tag, amount, currency, application, organization} = this.state;

    // Validation
    if (!tag || tag.trim() === "") {
      Setting.showMessage("error", i18next.t("general:Please input") + " " + i18next.t("user:Tag"));
      return;
    }
    if (!amount || amount <= 0) {
      Setting.showMessage("error", i18next.t("general:Please input") + " " + i18next.t("transaction:Amount"));
      return;
    }

    this.props.onOk({
      tag,
      amount,
      currency,
      application,
      organization,
    });
  };

  handleCancel = () => {
    this.props.onCancel();
  };

  render() {
    return (
      <Modal
        title={i18next.t("transaction:Recharge")}
        open={this.props.visible}
        onOk={this.handleOk}
        onCancel={this.handleCancel}
        okText={i18next.t("general:OK")}
        cancelText={i18next.t("general:Cancel")}
        width={600}
      >
        <Row style={{marginTop: "20px"}}>
          <Col span={6}>
            {Setting.getLabel(i18next.t("user:Tag"), i18next.t("transaction:Tag - Tooltip"))} :
          </Col>
          <Col span={18}>
            <Select
              virtual={false}
              style={{width: "100%"}}
              value={this.state.tag}
              onChange={(value) => this.setState({tag: value})}
              mode="tags"
              maxCount={1}
            >
              <Option value="Organization">{i18next.t("general:Organization")}</Option>
              <Option value="User">{i18next.t("general:User")}</Option>
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col span={6}>
            {Setting.getLabel(i18next.t("transaction:Amount"), i18next.t("transaction:Amount - Tooltip"))} :
          </Col>
          <Col span={18}>
            <InputNumber
              style={{width: "100%"}}
              value={this.state.amount}
              onChange={(value) => this.setState({amount: value})}
              min={0}
              step={0.01}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col span={6}>
            {Setting.getLabel(i18next.t("payment:Currency"), i18next.t("currency:Currency - Tooltip"))} :
          </Col>
          <Col span={18}>
            <Select
              virtual={false}
              style={{width: "100%"}}
              value={this.state.currency}
              onChange={(value) => this.setState({currency: value})}
              showSearch
            >
              {Setting.CurrencyOptions.map((item, index) => (
                <Option key={index} value={item.id}>
                  {Setting.getCurrencyWithFlag(item.id)}
                </Option>
              ))}
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col span={6}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={18}>
            <Select
              virtual={false}
              style={{width: "100%"}}
              value={this.state.organization}
              onChange={(value) => this.setState({organization: value})}
              allowClear
              showSearch
            >
              {this.state.organizations.map((org, index) => (
                <Option key={index} value={org.name}>
                  {org.name}
                </Option>
              ))}
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col span={6}>
            {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
          </Col>
          <Col span={18}>
            <Select
              virtual={false}
              style={{width: "100%"}}
              value={this.state.application}
              onChange={(value) => this.setState({application: value})}
              allowClear
              showSearch
            >
              {this.state.applications.map((app, index) => (
                <Option key={index} value={app.name}>
                  {app.name}
                </Option>
              ))}
            </Select>
          </Col>
        </Row>
      </Modal>
    );
  }
}

export default RechargeModal;
