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
import * as PermissionBackend from "./backend/PermissionBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import * as RoleBackend from "./backend/RoleBackend";
import * as ModelBackend from "./backend/ModelBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import moment from "moment/moment";

class PermissionEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      permissionName: props.match.params.permissionName,
      permission: null,
      organizations: [],
      model: null,
      users: [],
      roles: [],
      models: [],
      resources: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getPermission();
    this.getOrganizations();
  }

  getPermission() {
    PermissionBackend.getPermission(this.state.organizationName, this.state.permissionName)
      .then((permission) => {
        this.setState({
          permission: permission,
        });

        this.getUsers(permission.owner);
        this.getRoles(permission.owner);
        this.getModels(permission.owner);
        this.getResources(permission.owner);
        this.getModel(permission.owner, permission.model);
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

  getUsers(organizationName) {
    UserBackend.getUsers(organizationName)
      .then((res) => {
        this.setState({
          users: res,
        });
      });
  }

  getRoles(organizationName) {
    RoleBackend.getRoles(organizationName)
      .then((res) => {
        this.setState({
          roles: res,
        });
      });
  }

  getModels(organizationName) {
    ModelBackend.getModels(organizationName)
      .then((res) => {
        this.setState({
          models: res,
        });
      });
  }

  getModel(organizationName, modelName) {
    ModelBackend.getModel(organizationName, modelName)
      .then((res) => {
        this.setState({
          model: res,
        });
      });
  }

  getResources(organizationName) {
    ApplicationBackend.getApplicationsByOrganization("admin", organizationName)
      .then((res) => {
        this.setState({
          resources: (res.msg === undefined) ? res : [],
        });
      });
  }

  parsePermissionField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updatePermissionField(key, value) {
    if (key === "model") {
      this.getModel(this.state.permission.owner, value);
    }

    value = this.parsePermissionField(key, value);

    const permission = this.state.permission;
    permission[key] = value;
    this.setState({
      permission: permission,
    });
  }

  hasRoleDefinition(model) {
    if (model !== null) {
      return model.modelText.includes("role_definition");
    }
    return false;
  }

  renderPermission() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("permission:New Permission") : i18next.t("permission:Edit Permission")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitPermissionEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitPermissionEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deletePermission()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.permission.owner} onChange={(owner => {
              this.updatePermissionField("owner", owner);
              this.getUsers(owner);
              this.getRoles(owner);
              this.getModels(owner);
              this.getResources(owner);
            })}
            options={this.state.organizations.map((organization) => Setting.getOption(organization.name, organization.name))
            } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.permission.name} onChange={e => {
              this.updatePermissionField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.permission.displayName} onChange={e => {
              this.updatePermissionField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.permission.description} onChange={e => {
              this.updatePermissionField("description", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Model"), i18next.t("general:Model - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.permission.model} onChange={(model => {
              this.updatePermissionField("model", model);
            })}
            options={this.state.models.map((model) => Setting.getOption(model.name, model.name))
            } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Adapter"), i18next.t("general:Adapter - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.permission.adapter} onChange={e => {
              this.updatePermissionField("adapter", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("role:Sub users"), i18next.t("role:Sub users - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}} value={this.state.permission.users}
              onChange={(value => {this.updatePermissionField("users", value);})}
              options={this.state.users.map((user) => Setting.getOption(`${user.owner}/${user.name}`, `${user.owner}/${user.name}`))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("role:Sub roles"), i18next.t("role:Sub roles - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select disabled={!this.hasRoleDefinition(this.state.model)} virtual={false} mode="multiple" style={{width: "100%"}} value={this.state.permission.roles}
              onChange={(value => {this.updatePermissionField("roles", value);})}
              options={this.state.roles.filter(roles => (roles.owner !== this.state.roles.owner || roles.name !== this.state.roles.name)).map((permission) => Setting.getOption(`${permission.owner}/${permission.name}`, `${permission.owner}/${permission.name}`))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("role:Sub domains"), i18next.t("role:Sub domains - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: "100%"}} value={this.state.permission.domains}
              onChange={(value => {
                this.updatePermissionField("domains", value);
              })}
              options={this.state.permission.domains.map((domain) => Setting.getOption(domain, domain))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("permission:Resource type"), i18next.t("permission:Resource type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.permission.resourceType} onChange={(value => {
              this.updatePermissionField("resourceType", value);
            })}
            options={[
              {value: "Application", name: i18next.t("general:Application")},
              {value: "TreeNode", name: i18next.t("permission:TreeNode")},
            ].map((item) => Setting.getOption(item.name, item.value))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Resources"), i18next.t("permission:Resources - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}} value={this.state.permission.resources}
              onChange={(value => {this.updatePermissionField("resources", value);})}
              options={this.state.resources.map((resource) => Setting.getOption(`${resource.name}`, `${resource.name}`))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("permission:Actions"), i18next.t("permission:Actions - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}} value={this.state.permission.actions} onChange={(value => {
              this.updatePermissionField("actions", value);
            })}
            options={[
              {value: "Read", name: i18next.t("permission:Read")},
              {value: "Write", name: i18next.t("permission:Write")},
              {value: "Admin", name: i18next.t("permission:Admin")},
            ].map((item) => Setting.getOption(item.name, item.value))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("permission:Effect"), i18next.t("permission:Effect - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.permission.effect} onChange={(value => {
              this.updatePermissionField("effect", value);
            })}
            options={[
              {value: "Allow", name: i18next.t("permission:Allow")},
              {value: "Deny", name: i18next.t("permission:Deny")},
            ].map((item) => Setting.getOption(item.name, item.value))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("general:Is enabled"), i18next.t("general:Is enabled - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.permission.isEnabled} onChange={checked => {
              this.updatePermissionField("isEnabled", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("permission:Submitter"), i18next.t("permission:Submitter - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.permission.submitter} onChange={e => {
              this.updatePermissionField("submitter", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("permission:Approver"), i18next.t("permission:Approver - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.permission.approver} onChange={e => {
              this.updatePermissionField("approver", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("permission:Approve time"), i18next.t("permission:Approve time - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={Setting.getFormattedDate(this.state.permission.approveTime)} onChange={e => {
              this.updatePermissionField("approveTime", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:State"), i18next.t("general:State - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} disabled={!Setting.isLocalAdminUser(this.props.account)} style={{width: "100%"}} value={this.state.permission.state} onChange={(value => {
              if (this.state.permission.state !== value) {
                if (value === "Approved") {
                  this.updatePermissionField("approver", this.props.account.name);
                  this.updatePermissionField("approveTime", moment().format());
                } else {
                  this.updatePermissionField("approver", "");
                  this.updatePermissionField("approveTime", "");
                }
              }

              this.updatePermissionField("state", value);
            })}
            options={[
              {value: "Approved", name: i18next.t("permission:Approved")},
              {value: "Pending", name: i18next.t("permission:Pending")},
            ].map((item) => Setting.getOption(item.name, item.value))}
            />
          </Col>
        </Row>
      </Card>
    );
  }

  submitPermissionEdit(willExist) {
    if (this.state.permission.users.length === 0 && this.state.permission.roles.length === 0) {
      Setting.showMessage("error", "The users and roles cannot be empty at the same time");
      return;
    }
    // if (this.state.permission.domains.length === 0) {
    //   Setting.showMessage("error", "The domains cannot be empty");
    //   return;
    // }
    if (this.state.permission.resources.length === 0) {
      Setting.showMessage("error", "The resources cannot be empty");
      return;
    }
    if (this.state.permission.actions.length === 0) {
      Setting.showMessage("error", "The actions cannot be empty");
      return;
    }
    if (!Setting.isLocalAdminUser(this.props.account) && this.state.permission.submitter !== this.props.account.name) {
      Setting.showMessage("error", "A normal user can only modify the permission submitted by itself");
      return;
    }

    const permission = Setting.deepCopy(this.state.permission);
    PermissionBackend.updatePermission(this.state.organizationName, this.state.permissionName, permission)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            permissionName: this.state.permission.name,
          });

          if (willExist) {
            this.props.history.push("/permissions");
          } else {
            this.props.history.push(`/permissions/${this.state.permission.owner}/${this.state.permission.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updatePermissionField("name", this.state.permissionName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deletePermission() {
    PermissionBackend.deletePermission(this.state.permission)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/permissions");
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
          this.state.permission !== null ? this.renderPermission() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitPermissionEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitPermissionEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deletePermission()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default PermissionEditPage;
