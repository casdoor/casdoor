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

import moment from "moment";
import React from "react";
import {Button, Card, Col, DatePicker, Input, Row, Select} from "antd";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as PricingBackend from "./backend/PricingBackend";
import * as PlanBackend from "./backend/PlanBackend";
import * as SubscriptionBackend from "./backend/SubscriptionBackend";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import dayjs from "dayjs";

class SubscriptionEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      subscriptionName: props.match.params.subscriptionName,
      subscription: null,
      organizations: [],
      users: [],
      pricings: [],
      plans: [],
      providers: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getSubscription();
    this.getOrganizations();
  }

  getSubscription() {
    SubscriptionBackend.getSubscription(this.state.organizationName, this.state.subscriptionName)
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
          subscription: res.data,
        });

        this.getUsers(this.state.organizationName);
        this.getPricings(this.state.organizationName);
        this.getPlans(this.state.organizationName);
      });
  }

  getPricings(organizationName) {
    PricingBackend.getPricings(organizationName)
      .then((res) => {
        this.setState({
          pricings: res.data,
        });
      });
  }

  getPlans(organizationName) {
    PlanBackend.getPlans(organizationName)
      .then((res) => {
        this.setState({
          plans: res.data,
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

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: res.data || [],
        });
      });
  }

  parseSubscriptionField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateSubscriptionField(key, value) {
    value = this.parseSubscriptionField(key, value);

    const subscription = this.state.subscription;
    subscription[key] = value;
    this.setState({
      subscription: subscription,
    });
  }

  renderSubscription() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("subscription:New Subscription") : i18next.t("subscription:Edit Subscription")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitSubscriptionEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitSubscriptionEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteSubscription()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.subscription.owner} onChange={(owner => {
              this.updateSubscriptionField("owner", owner);
              this.getUsers(owner);
              this.getPlans(owner);
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
            <Input value={this.state.subscription.name} onChange={e => {
              this.updateSubscriptionField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.subscription.displayName} onChange={e => {
              this.updateSubscriptionField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("subscription:Start time"), i18next.t("subscription:Start time - Tooltip"))}
          </Col>
          <Col span={22} >
            <DatePicker value={dayjs(this.state.subscription.startTime)} onChange={value => {
              this.updateSubscriptionField("startTime", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("subscription:End time"), i18next.t("subscription:End time - Tooltip"))}
          </Col>
          <Col span={22} >
            <DatePicker value={dayjs(this.state.subscription.endTime)} onChange={value => {
              this.updateSubscriptionField("endTime", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("plan:Period"), i18next.t("plan:Period - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select
              defaultValue={this.state.subscription.period === "" ? "Monthly" : this.state.subscription.period}
              onChange={value => {
                this.updateSubscriptionField("period", value);
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
            {Setting.getLabel(i18next.t("general:User"), i18next.t("general:User - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select style={{width: "100%"}} value={this.state.subscription.user}
              onChange={(value => {this.updateSubscriptionField("user", value);})}
              options={this.state.users.map((user) => Setting.getOption(user.name, user.name))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Pricing"), i18next.t("general:Pricing - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.subscription.pricing}
              onChange={(value => {this.updateSubscriptionField("pricing", value);})}
              options={this.state.pricings.map((pricing) => Setting.getOption(pricing.name, pricing.name))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Plan"), i18next.t("general:Plan - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.subscription.plan}
              onChange={(value => {this.updateSubscriptionField("plan", value);})}
              options={this.state.plans.map((plan) => Setting.getOption(plan.name, plan.name))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Payment"), i18next.t("general:Payment - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.subscription.payment} disabled={true} onChange={e => {
              this.updateSubscriptionField("payment", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.subscription.description} onChange={e => {
              this.updateSubscriptionField("description", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:State"), i18next.t("general:State - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} disabled={!Setting.isLocalAdminUser(this.props.account)} style={{width: "100%"}} value={this.state.subscription.state} onChange={(value => {
              if (this.state.subscription.state !== value) {
                if (value === "Approved") {
                  this.updateSubscriptionField("approver", this.props.account.name);
                  this.updateSubscriptionField("approveTime", moment().format());
                } else {
                  this.updateSubscriptionField("approver", "");
                  this.updateSubscriptionField("approveTime", "");
                }
              }

              this.updateSubscriptionField("state", value);
            })}
            options={[
              {value: "Pending", name: i18next.t("permission:Pending")},
              {value: "Active", name: i18next.t("permission:Active")},
              {value: "Upcoming", name: i18next.t("permission:Upcoming")},
              {value: "Expired", name: i18next.t("permission:Expired")},
              {value: "Error", name: i18next.t("permission:Error")},
              {value: "Suspended", name: i18next.t("permission:Suspended")},
            ].map((item) => Setting.getOption(item.name, item.value))}
            />
          </Col>
        </Row>
      </Card>
    );
  }

  submitSubscriptionEdit(willExist) {
    const subscription = Setting.deepCopy(this.state.subscription);
    SubscriptionBackend.updateSubscription(this.state.organizationName, this.state.subscriptionName, subscription)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            subscriptionName: this.state.subscription.name,
          });

          if (willExist) {
            this.props.history.push("/subscriptions");
          } else {
            this.props.history.push(`/subscriptions/${this.state.subscription.owner}/${this.state.subscription.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateSubscriptionField("name", this.state.subscriptionName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteSubscription() {
    SubscriptionBackend.deleteSubscription(this.state.subscription)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/subscriptions");
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
          this.state.subscription !== null ? this.renderSubscription() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitSubscriptionEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitSubscriptionEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteSubscription()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default SubscriptionEditPage;
