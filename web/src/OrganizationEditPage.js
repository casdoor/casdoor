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
import {Button, Card, Col, Input, Row, Select} from 'antd';
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import {LinkOutlined} from "@ant-design/icons";

const { Option } = Select;

class OrganizationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.match.params.organizationName,
      organization: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrganization();
  }

  getOrganization() {
    OrganizationBackend.getOrganization("admin", this.state.organizationName)
      .then((organization) => {
        this.setState({
          organization: organization,
        });
      });
  }

  parseOrganizationField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateOrganizationField(key, value) {
    value = this.parseOrganizationField(key, value);

    let organization = this.state.organization;
    organization[key] = value;
    this.setState({
      organization: organization,
    });
  }

  renderOrganization() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("organization:Edit Organization")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitOrganizationEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.name} onChange={e => {
              this.updateOrganizationField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Display name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.displayName} onChange={e => {
              this.updateOrganizationField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            Favicon:
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={1}>
                URL:
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined/>} value={this.state.organization.favicon} onChange={e => {
                  this.updateOrganizationField('favicon', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={1}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" rel="noreferrer" href={this.state.organization.favicon}>
                  <img src={this.state.organization.favicon} alt={this.state.organization.favicon} height={90} style={{marginBottom: '20px'}}/>
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("organization:Website URL")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.websiteUrl} onChange={e => {
              this.updateOrganizationField('websiteUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Password type")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.organization.passwordType} onChange={(value => {this.updateOrganizationField('passwordType', value);})}>
              {
                ['plain', 'salt']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Password salt")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.passwordSalt} onChange={e => {
              this.updateOrganizationField('passwordSalt', e.target.value);
            }} />
          </Col>
        </Row>
      </Card>
    )
  }

  submitOrganizationEdit() {
    let organization = Setting.deepCopy(this.state.organization);
    OrganizationBackend.updateOrganization(this.state.organization.owner, this.state.organizationName, organization)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            organizationName: this.state.organization.name,
          });
          this.props.history.push(`/organizations/${this.state.organization.name}`);
        } else {
          Setting.showMessage("error", res.msg);
          this.updateOrganizationField('name', this.state.organizationName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
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
              this.state.organization !== null ? this.renderOrganization() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitOrganizationEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default OrganizationEditPage;
