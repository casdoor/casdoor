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
import {Button, Card, Col, Input, InputNumber, Row, Select, Switch} from "antd";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as RoleBackend from "./backend/RoleBackend";
import * as PlanBackend from "./backend/PlanBackend";
import * as UserBackend from "./backend/UserBackend";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const {Option} = Select;

class PlanEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props?.organizationName ?? props?.match?.params?.organizationName ?? null,
      planName: props?.match?.params?.planName ?? null,
      plan: null,
      organizations: [],
      users: [],
      roles: [],
      paymentProviders: [],
      mode: props?.location?.mode ?? "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getPlan();
    this.getOrganizations();
  }

  getPlan() {
    PlanBackend.getPlan(this.state.organizationName, this.state.planName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        this.setState({
          plan: res.data,
        });

        this.getUsers(this.state.organizationName);
        this.getRoles(this.state.organizationName);
        this.getPaymentProviders(this.state.organizationName);
      });
  }

  getRoles(organizationName) {
    RoleBackend.getRoles(organizationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          roles: res.data,
        });
      });
  }

  getUsers(organizationName) {
    UserBackend.getUsers(organizationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          users: res.data,
        });
      });
  }

  getPaymentProviders(organizationName) {
    ProviderBackend.getProviders(organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            paymentProviders: res.data.filter(provider => provider.category === "Payment"),
          });
          return;
        }

        Setting.showMessage("error", res.msg);
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

  parsePlanField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updatePlanField(key, value) {
    value = this.parsePlanField(key, value);

    const plan = this.state.plan;
    plan[key] = value;
    this.setState({
      plan: plan,
    });
  }

  renderPlan() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("plan:New Plan") : i18next.t("plan:Edit Plan")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitPlanEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitPlanEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deletePlan()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.plan.owner} onChange={(owner => {
              this.updatePlanField("owner", owner);
              this.getUsers(owner);
              this.getRoles(owner);
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
            <Input value={this.state.plan.name} onChange={e => {
              this.updatePlanField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.plan.displayName} onChange={e => {
              this.updatePlanField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Role"), i18next.t("general:Role - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.plan.role} onChange={(value => {this.updatePlanField("role", value);})}
              options={this.state.roles.map((role) => Setting.getOption(role.name, role.name))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.plan.description} onChange={e => {
              this.updatePlanField("description", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("plan:Price"), i18next.t("plan:Price - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.plan.price} onChange={value => {
              this.updatePlanField("price", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("plan:Period"), i18next.t("plan:Period - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.plan.period} onChange={value => {
              this.updatePlanField("period", value);
            }}
            options={[
              {value: "Monthly", label: "Monthly"},
              {value: "Yearly", label: "Yearly"},
            ]}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Currency"), i18next.t("payment:Currency - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.plan.currency} onChange={(value => {
              this.updatePlanField("currency", value);
            })}>
              {
                [
                  {id: "USD", name: "USD"},
                  {id: "CNY", name: "CNY"},
                  {id: "EUR", name: "EUR"},
                  {id: "JPY", name: "JPY"},
                  {id: "GBP", name: "GBP"},
                  {id: "AUD", name: "AUD"},
                  {id: "CAD", name: "CAD"},
                  {id: "CHF", name: "CHF"},
                  {id: "HKD", name: "HKD"},
                  {id: "SGD", name: "SGD"},
                  {id: "BRL", name: "BRL"},
                  {id: "PLN", name: "PLN"},
                  {id: "KRW", name: "KRW"},
                  {id: "INR", name: "INR"},
                  {id: "RUB", name: "RUB"},
                  {id: "MXN", name: "MXN"},
                  {id: "ZAR", name: "ZAR"},
                  {id: "TRY", name: "TRY"},
                  {id: "SEK", name: "SEK"},
                  {id: "NOK", name: "NOK"},
                  {id: "DKK", name: "DKK"},
                  {id: "THB", name: "THB"},
                  {id: "MYR", name: "MYR"},
                  {id: "TWD", name: "TWD"},
                  {id: "CZK", name: "CZK"},
                  {id: "HUF", name: "HUF"},
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Payment providers"), i18next.t("product:Payment providers - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}} value={this.state.plan.paymentProviders ?? []} onChange={(value => {this.updatePlanField("paymentProviders", value);})}>
              {
                this.state.paymentProviders.map((provider, index) => <Option key={index} value={provider.name}>{provider.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("general:Is enabled"), i18next.t("general:Is enabled - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.plan.isEnabled} onChange={checked => {
              this.updatePlanField("isEnabled", checked);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitPlanEdit(exitAfterSave) {
    const plan = Setting.deepCopy(this.state.plan);
    PlanBackend.updatePlan(this.state.organizationName, this.state.planName, plan)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            planName: this.state.plan.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/plans");
          } else {
            this.props.history.push(`/plans/${this.state.plan.owner}/${this.state.plan.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updatePlanField("name", this.state.planName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deletePlan() {
    PlanBackend.deletePlan(this.state.plan)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/plans");
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
          this.state.plan !== null ? this.renderPlan() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitPlanEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitPlanEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deletePlan()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default PlanEditPage;
