// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
import {Button, Card, Col, Input, Row, Select} from "antd";
import * as AdapterBackend from "./backend/AdapterBackend";
import * as EnforcerBackend from "./backend/EnforcerBackend";
import * as ModelBackend from "./backend/ModelBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import PolicyTable from "./table/PolicyTable";
import * as Setting from "./Setting";
import i18next from "i18next";

class EnforcerEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      enforcerName: props.match.params.enforcerName,
      enforcer: null,
      organizations: [],
      models: [],
      adapters: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getEnforcer();
    this.getOrganizations();
  }

  getEnforcer() {
    EnforcerBackend.getEnforcer(this.state.organizationName, this.state.enforcerName, true)
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
          enforcer: res.data,
        });

        this.getModels(this.state.organizationName);
        this.getAdapters(this.state.organizationName);
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

  getModels(organizationName) {
    ModelBackend.getModels(organizationName)
      .then((res) => {
        this.setState({
          models: res.data || [],
        });
      });
  }

  getAdapters(organizationName) {
    AdapterBackend.getAdapters(organizationName)
      .then((res) => {
        this.setState({
          adapters: res.data || [],
        });
      });
  }

  parseEnforcerField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateEnforcerField(key, value) {
    value = this.parseEnforcerField(key, value);

    const enforcer = this.state.enforcer;
    enforcer[key] = value;
    this.setState({
      enforcer: enforcer,
    });
  }

  renderEnforcer() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("enforcer:New Enforcer") : i18next.t("enforcer:Edit Enforcer")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitEnforcerEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitEnforcerEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteEnforcer()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account) || Setting.builtInObject(this.state.enforcer)} value={this.state.enforcer.owner} onChange={(owner => {
              this.updateEnforcerField("owner", owner);
              this.getModels(owner);
              this.getAdapters(owner);
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
            <Input disabled={Setting.builtInObject(this.state.enforcer)} value={this.state.enforcer.name} onChange={e => {
              this.updateEnforcerField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.enforcer.displayName} onChange={e => {
              this.updateEnforcerField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.enforcer.description} onChange={e => {
              this.updateEnforcerField("description", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Model"), i18next.t("general:Model - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} disabled={Setting.builtInObject(this.state.enforcer)} style={{width: "100%"}} value={this.state.enforcer.model} onChange={(model => {
              this.updateEnforcerField("model", model);
            })}
            options={this.state.models.map((model) => Setting.getOption(`${model.owner}/${model.name}`, `${model.owner}/${model.name}`))
            } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Adapter"), i18next.t("general:Adapter - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} disabled={Setting.builtInObject(this.state.enforcer)} style={{width: "100%"}} value={this.state.enforcer.adapter} onChange={(adapter => {
              this.updateEnforcerField("adapter", adapter);
            })}
            options={this.state.adapters.map((adapter) => Setting.getOption(`${adapter.owner}/${adapter.name}`, `${adapter.owner}/${adapter.name}`))
            } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("adapter:Policies"), i18next.t("adapter:Policies - Tooltip"))} :
          </Col>
          <Col span={22}>
            <PolicyTable enforcer={this.state.enforcer} modelCfg={this.state.enforcer?.modelCfg} mode={this.state.mode} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitEnforcerEdit(exitAfterSave) {
    const enforcer = Setting.deepCopy(this.state.enforcer);
    EnforcerBackend.updateEnforcer(this.state.organizationName, this.state.enforcerName, enforcer)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            enforcerName: this.state.enforcer.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/enforcers");
          } else {
            this.props.history.push(`/enforcers/${this.state.enforcer.owner}/${this.state.enforcer.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateEnforcerField("name", this.state.enforcerName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteEnforcer() {
    EnforcerBackend.deleteEnforcer(this.state.enforcer)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/enforcers");
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
          this.state.enforcer !== null ? this.renderEnforcer() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitEnforcerEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitEnforcerEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteEnforcer()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default EnforcerEditPage;
