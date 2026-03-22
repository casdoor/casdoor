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
import {Button, Card, Col, DatePicker, Input, Row, Select} from "antd";
import * as KeyBackend from "./backend/KeyBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import moment from "moment";

const {Option} = Select;

class KeyEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.match.params.organizationName,
      keyName: props.match.params.keyName,
      key: null,
      organizations: [],
      applications: [],
      users: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getKey();
    this.getOrganizations();
  }

  getKey() {
    KeyBackend.getKey(this.state.organizationName, this.state.keyName)
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

        this.getApplicationsByOrganization(res.data.organization || this.state.organizationName);
        this.getUsersByOrganization(res.data.organization || this.state.organizationName);
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: res.data || [],
        });
      });
  }

  getApplicationsByOrganization(organizationName) {
    ApplicationBackend.getApplicationsByOrganization("admin", organizationName)
      .then((res) => {
        this.setState({
          applications: res.data || [],
        });
      });
  }

  getUsersByOrganization(organizationName) {
    UserBackend.getUsers(organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            users: res.data || [],
          });
        }
      });
  }

  parseKeyField(key, value) {
    return value;
  }

  updateKeyField(key, value) {
    value = this.parseKeyField(key, value);

    const keyObj = this.state.key;
    keyObj[key] = value;
    this.setState({
      key: keyObj,
    });
  }

  renderKey() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("key:New Key") : i18next.t("key:Edit Key")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitKeyEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitKeyEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteKey()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.key.owner} onChange={(value => {
              this.updateKeyField("owner", value);
              this.updateKeyField("organization", value);
              this.getApplicationsByOrganization(value);
              this.getUsersByOrganization(value);
            })}>
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
            <Input value={this.state.key.name} onChange={e => {
              this.updateKeyField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.key.displayName} onChange={e => {
              this.updateKeyField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Type"), i18next.t("general:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.key.type} onChange={(value => {
              this.updateKeyField("type", value);
            })}>
              <Option value="Organization">{i18next.t("general:Organization")}</Option>
              <Option value="Application">{i18next.t("general:Application")}</Option>
              <Option value="User">{i18next.t("general:User")}</Option>
              <Option value="General">{i18next.t("general:General")}</Option>
            </Select>
          </Col>
        </Row>
        {
          this.state.key.type === "Application" ? (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Select virtual={false} style={{width: "100%"}} value={this.state.key.application} onChange={(value => {
                  this.updateKeyField("application", value);
                })}>
                  {
                    this.state.applications.map((application, index) => <Option key={index} value={application.name}>{application.name}</Option>)
                  }
                </Select>
              </Col>
            </Row>
          ) : null
        }
        {
          this.state.key.type === "User" ? (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("general:User"), i18next.t("general:User - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Select virtual={false} style={{width: "100%"}} value={this.state.key.user} onChange={(value => {
                  this.updateKeyField("user", value);
                })}>
                  {
                    this.state.users.map((user, index) => <Option key={index} value={user.name}>{user.name}</Option>)
                  }
                </Select>
              </Col>
            </Row>
          ) : null
        }
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("key:Access key"), i18next.t("key:Access key - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.key.accessKey} readOnly={true} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("key:Access secret"), i18next.t("key:Access secret - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input.Password value={this.state.key.accessSecret} readOnly={true} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Expire time"), i18next.t("general:Expire time - Tooltip"))} :
          </Col>
          <Col span={22} >
            <DatePicker
              showTime
              value={this.state.key.expireTime ? moment(this.state.key.expireTime) : null}
              onChange={(value, dateString) => {
                this.updateKeyField("expireTime", dateString);
              }}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:State"), i18next.t("general:State - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.key.state} onChange={(value => {
              this.updateKeyField("state", value);
            })}>
              <Option value="Active">{i18next.t("subscription:Active")}</Option>
              <Option value="Inactive">{i18next.t("key:Inactive")}</Option>
            </Select>
          </Col>
        </Row>
      </Card>
    );
  }

  submitKeyEdit(exitAfterSave) {
    const key = Setting.deepCopy(this.state.key);
    KeyBackend.updateKey(this.state.organizationName, this.state.keyName, key)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            organizationName: this.state.key.owner,
            keyName: this.state.key.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/keys");
          } else {
            this.props.history.push(`/keys/${this.state.key.owner}/${this.state.key.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateKeyField("owner", this.state.organizationName);
          this.updateKeyField("name", this.state.keyName);
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
        {
          this.state.key !== null ? this.renderKey() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitKeyEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitKeyEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </div>
      </div>
    );
  }
}

export default KeyEditPage;
