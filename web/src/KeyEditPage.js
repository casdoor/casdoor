// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
import * as KeyBackend from "./backend/KeyBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import copy from "copy-to-clipboard";

const {TextArea} = Input;
const {Option} = Select;

class KeyEditPage extends React.Component {
  constructor(props) {
    super(props);
    const mode = props.location.mode !== undefined ? props.location.mode : "edit";
    this.state = {
      classes: props,
      keyName: props.match.params.keyName,
      key: props.location.draftKey ?? (mode === "add" ? this.getDefaultKey() : null),
      mode: mode,
      latestApiKey: props.location.apiKey ?? "",
      organizations: [],
      applications: [],
    };
  }

  getDefaultKey() {
    const organizationName = Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account);
    return {
      owner: "admin",
      name: "",
      createdTime: "",
      updatedTime: "",
      displayName: "",
      description: "",
      type: "general",
      organization: organizationName,
      application: "app-built-in",
      user: "",
      scopes: ["read"],
      isEnabled: true,
      expiresTime: "",
      lastUsedTime: "",
      secretPreview: "",
    };
  }

  UNSAFE_componentWillMount() {
    if (this.state.mode !== "add") {
      this.getKey();
    }
    this.getOrganizations();
    this.getApplications();
  }

  getKey() {
    KeyBackend.getKey("admin", this.state.keyName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          key: res.data,
        });
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizationNames("admin")
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            organizations: res.data || [],
          });
        }
      });
  }

  getApplications() {
    ApplicationBackend.getApplications("admin")
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            applications: res.data || [],
          });
        }
      });
  }

  updateKeyField(keyField, value) {
    const key = this.state.key;
    key[keyField] = value;
    this.setState({
      key: key,
    });
  }

  updateKeyType(value) {
    const key = this.state.key;
    key.type = value;
    if (value === "general" || value === "application") {
      key.organization = "";
      key.user = "";
    } else if (value === "organization") {
      key.user = "";
    }

    this.setState({
      key: key,
    });
  }

  normalizeKey(key) {
    const normalizedKey = Setting.deepCopy(key);
    if (normalizedKey.type === "general" || normalizedKey.type === "application") {
      normalizedKey.organization = "";
      normalizedKey.user = "";
    } else if (normalizedKey.type === "organization") {
      normalizedKey.user = "";
    }
    return normalizedKey;
  }

  isUserKey() {
    return this.state.key?.type === "user";
  }

  renderLatestApiKey() {
    if (!this.state.latestApiKey) {
      return null;
    }

    return (
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
          {Setting.getLabel(i18next.t("general:API key"), i18next.t("general:API key - Tooltip"))} :
        </Col>
        <Col span={22}>
          <Button type="primary" style={{marginRight: "10px", marginBottom: "10px"}} onClick={() => {
            copy(this.state.latestApiKey);
            Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
          }}
          >
            {i18next.t("general:Copy")}
          </Button>
          <TextArea autoSize={{minRows: 3, maxRows: 10}} value={this.state.latestApiKey} readOnly />
        </Col>
      </Row>
    );
  }

  renderKey() {
    return (
      <Card size="small" title={(
        <div>
          {this.state.mode === "add" ? i18next.t("general:Keys") : i18next.t("general:Edit") + " " + i18next.t("general:Keys")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitKeyEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitKeyEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode !== "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.rotateKey()}>{i18next.t("general:Generate")}</Button> : null}
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.props.history.push("/keys")}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      )} style={Setting.isMobile() ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.key.name} onChange={e => this.updateKeyField("name", e.target.value)} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Type"), i18next.t("general:Type - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Select virtual={false} style={{width: "100%"}} value={this.state.key.type} onChange={(value) => this.updateKeyType(value)}>
              <Option value="organization">{i18next.t("general:Organization")}</Option>
              <Option value="application">{i18next.t("general:Application")}</Option>
              <Option value="user">{i18next.t("general:User")}</Option>
              <Option value="general">General</Option>
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.key.displayName} onChange={e => this.updateKeyField("displayName", e.target.value)} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.key.description} onChange={e => this.updateKeyField("description", e.target.value)} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Select virtual={false} showSearch optionFilterProp="children" style={{width: "100%"}} value={this.state.key.application} onChange={(value) => this.updateKeyField("application", value)}>
              {this.state.applications.map(application => <Option key={`${application.organization}/${application.name}`} value={application.name}>{application.displayName || application.name}</Option>)}
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Select virtual={false} showSearch optionFilterProp="children" style={{width: "100%"}} allowClear value={this.state.key.organization || undefined} onChange={(value) => this.updateKeyField("organization", value ?? "")} disabled={this.state.key.type === "general" || this.state.key.type === "application"}>
              {this.state.organizations.map(organization => <Option key={organization.name} value={organization.name}>{organization.displayName || organization.name}</Option>)}
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:User"), i18next.t("general:User - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.key.user} disabled={!this.isUserKey()} onChange={e => this.updateKeyField("user", e.target.value)} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Scope"), i18next.t("provider:Scope - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Select
              virtual={false}
              mode="tags"
              style={{width: "100%"}}
              value={this.state.key.scopes || []}
              onChange={(value) => this.updateKeyField("scopes", value)}
            >
              {(this.state.key.scopes || []).map(scope => <Option key={scope} value={scope}>{scope}</Option>)}
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Enabled"), i18next.t("general:Enabled - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Switch checked={this.state.key.isEnabled} onChange={(value) => this.updateKeyField("isEnabled", value)} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel("Expiration", "When set, the key can no longer be used after this time")} :
          </Col>
          <Col span={22}>
            <Input value={this.state.key.expiresTime} onChange={e => this.updateKeyField("expiresTime", e.target.value)} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Created time"), i18next.t("general:Created time - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.key.createdTime || ""} disabled />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel("Last used time", "The last time this key was used to access the server")} :
          </Col>
          <Col span={22}>
            <Input value={this.state.key.lastUsedTime || ""} disabled />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={Setting.isMobile() ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:API key"), i18next.t("general:API key - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input value={this.state.key.secretPreview || ""} disabled />
          </Col>
        </Row>
        {this.renderLatestApiKey()}
      </Card>
    );
  }

  submitKeyEdit(exitAfterSave) {
    const key = this.normalizeKey(this.state.key);
    const onSuccess = (savedKey, rawApiKey = "") => {
      Setting.showMessage("success", i18next.t("general:Successfully saved"));
      this.setState({
        keyName: savedKey.name,
        key: savedKey,
        latestApiKey: rawApiKey,
        mode: "edit",
      });

      if (exitAfterSave) {
        this.props.history.push("/keys");
      } else {
        this.props.history.push({
          pathname: `/keys/${savedKey.name}`,
          apiKey: rawApiKey,
        });
      }
    };

    if (this.state.mode === "add") {
      KeyBackend.addKey(key)
        .then((res) => {
          if (res.status === "ok") {
            onSuccess(res.data.key, res.data.apiKey);
          } else {
            Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          }
        })
        .catch(error => {
          Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        });
      return;
    }

    KeyBackend.updateKey(this.state.key.owner, this.state.keyName, key)
      .then((res) => {
        if (res.status === "ok") {
          onSuccess(key);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  rotateKey() {
    KeyBackend.rotateKey(this.state.key)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            key: res.data.key,
            latestApiKey: res.data.apiKey,
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteKey() {
    KeyBackend.deleteKey(this.state.key)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/keys");
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
        {this.state.key !== null ? this.renderKey() : null}
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitKeyEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitKeyEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode !== "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.rotateKey()}>{i18next.t("general:Generate")}</Button> : null}
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.props.history.push("/keys")}>{i18next.t("general:Cancel")}</Button> : null}
          {this.state.mode !== "add" ? <Button style={{marginLeft: "20px"}} danger size="large" onClick={() => this.deleteKey()}>{i18next.t("general:Delete")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default KeyEditPage;
