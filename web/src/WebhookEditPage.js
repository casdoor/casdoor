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
import * as WebhookBackend from "./backend/WebhookBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import WebhookHeaderTable from "./WebhookHeaderTable";

const { Option } = Select;

class WebhookEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      webhookName: props.match.params.webhookName,
      webhook: null,
      organizations: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getWebhook();
    this.getOrganizations();
  }

  getWebhook() {
    WebhookBackend.getWebhook("admin", this.state.webhookName)
      .then((webhook) => {
        this.setState({
          webhook: webhook,
        });
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: (res.msg === undefined) ? res : [],
        });
      });
  }

  parseWebhookField(key, value) {
    if (["port"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateWebhookField(key, value) {
    value = this.parseWebhookField(key, value);

    let webhook = this.state.webhook;
    webhook[key] = value;
    this.setState({
      webhook: webhook,
    });
  }

  renderWebhook() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("webhook:Edit Webhook")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitWebhookEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" onClick={() => this.submitWebhookEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </div>
      } style={(Setting.isMobile())? {margin: '5px'}:{}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.webhook.organization} onChange={(value => {this.updateWebhookField('organization', value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.webhook.name} onChange={e => {
              this.updateWebhookField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("webhook:URL"), i18next.t("webhook:URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.webhook.url} onChange={e => {
              this.updateWebhookField('url', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("webhook:Method"), i18next.t("webhook:Method - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.webhook.method} onChange={(value => {this.updateWebhookField('method', value);})}>
              {
                [
                  {id: 'POST', name: 'POST'},
                  {id: 'GET', name: 'GET'},
                  {id: 'PUT', name: 'PUT'},
                  {id: 'DELETE', name: 'DELETE'},
                ].map((method, index) => <Option key={index} value={method.id}>{method.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("webhook:Content type"), i18next.t("webhook:Content type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.webhook.contentType} onChange={(value => {this.updateWebhookField('contentType', value);})}>
              {
                [
                  {id: 'application/json', name: 'application/json'},
                  {id: 'application/x-www-form-urlencoded', name: 'application/x-www-form-urlencoded'},
                ].map((contentType, index) => <Option key={index} value={contentType.id}>{contentType.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("webhook:Headers"), i18next.t("webhook:Headers - Tooltip"))} :
          </Col>
          <Col span={22} >
            <WebhookHeaderTable
              title={i18next.t("webhook:Headers")}
              table={this.state.webhook.headers}
              onUpdateTable={(value) => { this.updateWebhookField('headers', value)}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("webhook:Events"), i18next.t("webhook:Events - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: '100%'}}
                    value={this.state.webhook.events}
                    onChange={value => {
                      this.updateWebhookField('events', value);
                    }} >
              {
                (
                  ["signup", "login", "logout", "update-user"].map((option, index) => {
                    return (
                      <Option key={option} value={option}>{option}</Option>
                    )
                  })
                )
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("general:Is enabled"), i18next.t("general:Is enabled - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.webhook.isEnabled} onChange={checked => {
              this.updateWebhookField('isEnabled', checked);
            }} />
          </Col>
        </Row>
      </Card>
    )
  }

  submitWebhookEdit(willExist) {
    let webhook = Setting.deepCopy(this.state.webhook);
    WebhookBackend.updateWebhook(this.state.webhook.owner, this.state.webhookName, webhook)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            webhookName: this.state.webhook.name,
          });

          if (willExist) {
            this.props.history.push(`/webhooks`);
          } else {
            this.props.history.push(`/webhooks/${this.state.webhook.name}`);
          }
        } else {
          Setting.showMessage("error", res.msg);
          this.updateWebhookField('name', this.state.webhookName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.webhook !== null ? this.renderWebhook() : null
        }
        <div style={{marginTop: '20px', marginLeft: '40px'}}>
          <Button size="large" onClick={() => this.submitWebhookEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" size="large" onClick={() => this.submitWebhookEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </div>
      </div>
    );
  }
}

export default WebhookEditPage;
