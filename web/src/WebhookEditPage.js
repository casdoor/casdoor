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
import {Button, Card, Col, Input, Row, Select, Switch} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as WebhookBackend from "./backend/WebhookBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import WebhookHeaderTable from "./table/WebhookHeaderTable";

import {Controlled as CodeMirror} from "react-codemirror2";
import "codemirror/lib/codemirror.css";
require("codemirror/theme/material-darker.css");
require("codemirror/mode/javascript/javascript");

const {Option} = Select;

const applicationTemplate = {
  owner: "admin", // this.props.account.applicationName,
  name: "application_123",
  organization: "built-in",
  createdTime: "2022-01-01T01:03:42+08:00",
  displayName: "New Application - 123",
  logo: `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`,
  enablePassword: true,
  enableSignUp: true,
  enableSigninSession: false,
  enableCodeSignin: false,
  enableSamlCompress: false,
};

const previewTemplate = {
  "id": 9078,
  "owner": "built-in",
  "name": "68f55b28-7380-46b1-9bde-64fe1576e3b3",
  "createdTime": "2022-01-01T01:03:42+08:00",
  "organization": "built-in",
  "clientIp": "159.89.126.192",
  "user": "admin",
  "method": "POST",
  "requestUri": "/api/add-application",
  "action": "login",
  "isTriggered": false,
  "object": JSON.stringify(applicationTemplate),
};

const userTemplate = {
  "owner": "built-in",
  "name": "admin",
  "createdTime": "2020-07-16T21:46:52+08:00",
  "updatedTime": "",
  "id": "9eb20f79-3bb5-4e74-99ac-39e3b9a171e8",
  "type": "normal-user",
  "password": "***",
  "passwordSalt": "",
  "displayName": "Admin",
  "avatar": "https://cdn.casbin.com/usercontent/admin/avatar/1596241359.png",
  "permanentAvatar": "https://cdn.casbin.com/casdoor/avatar/casbin/admin.png",
  "email": "admin@example.com",
  "phone": "",
  "location": "",
  "address": null,
  "affiliation": "",
  "title": "",
  "score": 10000,
  "ranking": 10,
  "isOnline": false,
  "isAdmin": true,
  "isGlobalAdmin": false,
  "isForbidden": false,
  "isDeleted": false,
  "signupApplication": "app-casnode",
  "properties": {
    "bio": "",
    "checkinDate": "20200801",
    "editorType": "",
    "emailVerifiedTime": "2020-07-16T21:46:52+08:00",
    "fileQuota": "50",
    "location": "",
    "no": "22",
    "oauth_QQ_displayName": "",
    "oauth_QQ_verifiedTime": "",
    "oauth_WeChat_displayName": "",
    "oauth_WeChat_verifiedTime": "",
    "onlineStatus": "false",
    "phoneVerifiedTime": "",
    "renameQuota": "3",
    "tagline": "",
    "website": "",
  },
};

class WebhookEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      webhookName: props.match.params.webhookName,
      webhook: null,
      organizations: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
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

    const webhook = this.state.webhook;
    webhook[key] = value;
    this.setState({
      webhook: webhook,
    });
  }

  renderWebhook() {
    const preview = Setting.deepCopy(previewTemplate);
    if (this.state.webhook.isUserExtended) {
      preview["extendedUser"] = userTemplate;
    }
    const previewText = JSON.stringify(preview, null, 2);

    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("webhook:New Webhook") : i18next.t("webhook:Edit Webhook")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitWebhookEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitWebhookEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteWebhook()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.webhook.organization} onChange={(value => {this.updateWebhookField("organization", value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.webhook.name} onChange={e => {
              this.updateWebhookField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.webhook.url} onChange={e => {
              this.updateWebhookField("url", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Method"), i18next.t("webhook:Method - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.webhook.method} onChange={(value => {this.updateWebhookField("method", value);})}>
              {
                [
                  {id: "POST", name: "POST"},
                  {id: "GET", name: "GET"},
                  {id: "PUT", name: "PUT"},
                  {id: "DELETE", name: "DELETE"},
                ].map((method, index) => <Option key={index} value={method.id}>{method.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("webhook:Content type"), i18next.t("webhook:Content type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.webhook.contentType} onChange={(value => {this.updateWebhookField("contentType", value);})}>
              {
                [
                  {id: "application/json", name: "application/json"},
                  {id: "application/x-www-form-urlencoded", name: "application/x-www-form-urlencoded"},
                ].map((contentType, index) => <Option key={index} value={contentType.id}>{contentType.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("webhook:Headers"), i18next.t("webhook:Headers - Tooltip"))} :
          </Col>
          <Col span={22} >
            <WebhookHeaderTable
              title={i18next.t("webhook:Headers")}
              table={this.state.webhook.headers}
              onUpdateTable={(value) => {this.updateWebhookField("headers", value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("webhook:Events"), i18next.t("webhook:Events - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: "100%"}}
              value={this.state.webhook.events}
              onChange={value => {
                this.updateWebhookField("events", value);
              }} >
              {
                (
                  ["signup", "login", "logout", "add-user", "update-user", "delete-user", "add-organization", "update-organization", "delete-organization", "add-application", "update-application", "delete-application", "add-provider", "update-provider", "delete-provider"].map((option, index) => {
                    return (
                      <Option key={option} value={option}>{option}</Option>
                    );
                  })
                )
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("webhook:Is user extended"), i18next.t("webhook:Is user extended - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.webhook.isUserExtended} onChange={checked => {
              this.updateWebhookField("isUserExtended", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
          </Col>
          <Col span={22} >
            <div style={{width: "900px", height: "300px"}} >
              <CodeMirror
                value={previewText}
                options={{mode: "javascript", theme: "material-darker"}}
                onBeforeChange={(editor, data, value) => {}}
              />
            </div>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("general:Is enabled"), i18next.t("general:Is enabled - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.webhook.isEnabled} onChange={checked => {
              this.updateWebhookField("isEnabled", checked);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitWebhookEdit(willExist) {
    const webhook = Setting.deepCopy(this.state.webhook);
    WebhookBackend.updateWebhook(this.state.webhook.owner, this.state.webhookName, webhook)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            webhookName: this.state.webhook.name,
          });

          if (willExist) {
            this.props.history.push("/webhooks");
          } else {
            this.props.history.push(`/webhooks/${this.state.webhook.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateWebhookField("name", this.state.webhookName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteWebhook() {
    WebhookBackend.deleteWebhook(this.state.webhook)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/webhooks");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.webhook !== null ? this.renderWebhook() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitWebhookEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitWebhookEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteWebhook()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default WebhookEditPage;
