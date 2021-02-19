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
import {Button, Card, Col, Input, Row, Select, Switch} from 'antd';
import {LinkOutlined} from "@ant-design/icons";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as Setting from "./Setting";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import Face from "./auth/Face";
import i18next from "i18next";

const { Option } = Select;

class ApplicationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applicationName: props.match.params.applicationName,
      application: null,
      organizations: [],
      providers: [],
    };
  }

  componentWillMount() {
    this.getApplication();
    this.getOrganizations();
    this.getProviders();
  }

  getApplication() {
    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: res,
        });
      });
  }

  getProviders() {
    ProviderBackend.getProviders("admin")
      .then((res) => {
        this.setState({
          providers: res,
        });
      });
  }

  parseApplicationField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateApplicationField(key, value) {
    value = this.parseApplicationField(key, value);

    let application = this.state.application;
    application[key] = value;
    this.setState({
      application: application,
    });
  }

  renderApplication() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("application:Edit Application")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitApplicationEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.application.name} onChange={e => {
              this.updateApplicationField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Display Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.application.displayName} onChange={e => {
              this.updateApplicationField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            Logo:
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={1}>
                URL:
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined/>} value={this.state.application.logo} onChange={e => {
                  this.updateApplicationField('logo', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={1}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" href={this.state.application.logo}>
                  <img src={this.state.application.logo} alt={this.state.application.logo} height={90} style={{marginBottom: '20px'}}/>
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Organization")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.application.organization} onChange={(value => {this.updateApplicationField('organization', value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("application:Enable Password")}:
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enablePassword} onChange={checked => {
              this.updateApplicationField('enablePassword', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Providers")}:
          </Col>
          <Col span={22} >
            <Select mode="tags" style={{width: '100%'}}
                    value={this.state.application.providers}
                    onChange={value => {
                      this.updateApplicationField('providers', value);
                    }} >
              {
                this.state.providers.map((provider, index) => <Option key={index} value={provider.name}>{provider.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("application:Face Preview")}:
          </Col>
          <Col span={22} >
            <a style={{marginBottom: '10px'}} target="_blank" href={`/doors/${this.state.application.name}`}>
              {
                `${window.location.host}/doors/${this.state.application.name}`
              }
            </a>
            <br/>
            <br/>
            <div style={{width: "500px", height: "600px", border: "1px solid rgb(217,217,217)"}}>
              <Face application={this.state.application} />
            </div>
          </Col>
        </Row>
      </Card>
    )
  }

  submitApplicationEdit() {
    let application = Setting.deepCopy(this.state.application);
    ApplicationBackend.updateApplication(this.state.application.owner, this.state.applicationName, application)
      .then((res) => {
        if (res) {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            applicationName: this.state.application.name,
          });
          this.props.history.push(`/applications/${this.state.application.name}`);
        } else {
          Setting.showMessage("error", `failed to save: server side failure`);
          this.updateApplicationField('name', this.state.applicationName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `failed to save: ${error}`);
      });
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.state.application !== null ? this.renderApplication() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitApplicationEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default ApplicationEditPage;
