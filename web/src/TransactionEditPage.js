// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
import * as TransactionBackend from "./backend/TransactionBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";
import {Button, Card, Col, Input, InputNumber, Row, Select} from "antd";
import i18next from "i18next";

const {Option} = Select;

class TransactionEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      transactionName: props.match.params.transactionName,
      transaction: null,
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
      organizations: [],
      applications: [],
      users: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getTransaction();
    if (this.state.mode === "recharge") {
      this.getOrganizations();
      this.getApplications(this.state.organizationName);
      this.getUsers(this.state.organizationName);
    }
  }

  getTransaction() {
    TransactionBackend.getTransaction(this.state.organizationName, this.state.transactionName)
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
          transaction: res.data,
        });

        Setting.scrollToDiv("invoice-area");
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  getOrganizations() {
    const isGlobalAdmin = Setting.isAdminUser(this.props.account);
    const owner = isGlobalAdmin ? "admin" : this.state.organizationName;

    OrganizationBackend.getOrganizations(owner)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            organizations: res.data || [],
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  getApplications(organizationName) {
    const targetOrganizationName = organizationName || this.state.organizationName;
    ApplicationBackend.getApplicationsByOrganization("admin", targetOrganizationName)
      .then((res) => {
        this.setState({
          applications: res.data || [],
        });
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  getUsers(organizationName) {
    const targetOrganizationName = organizationName || this.state.organizationName;
    UserBackend.getUsers(targetOrganizationName)
      .then((res) => {
        this.setState({
          users: res.data || [],
        });
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  submitTransactionEdit(exitAfterSave) {
    if (this.state.transaction === null) {
      return;
    }
    const transaction = Setting.deepCopy(this.state.transaction);
    TransactionBackend.updateTransaction(this.state.transaction.owner, this.state.transactionName, transaction)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            transactionName: this.state.transaction.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/transactions");
          } else {
            this.props.history.push(`/transactions/${this.state.organizationName}/${this.state.transaction.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateTransactionField("name", this.state.transactionName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteTransaction() {
    if (this.state.transaction === null) {
      return;
    }
    TransactionBackend.deleteTransaction(this.state.transaction)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/transactions");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  parseTransactionField(key, value) {
    if (["amount"].includes(key)) {
      value = parseFloat(value);
      if (isNaN(value)) {
        value = 0;
      }
    }
    return value;
  }

  updateTransactionField(key, value) {
    value = this.parseTransactionField(key, value);

    const transaction = this.state.transaction;
    transaction[key] = value;
    this.setState({
      transaction: transaction,
    });
  }

  renderTransaction() {
    const isRechargeMode = this.state.mode === "recharge";
    const title = isRechargeMode ? i18next.t("transaction:Recharge") : (this.state.mode === "add" ? i18next.t("transaction:New Transaction") : i18next.t("transaction:Edit Transaction"));

    return (
      <Card size="small" title={
        <div>
          {title}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitTransactionEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitTransactionEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {(this.state.mode === "add" || isRechargeMode) ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteTransaction()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        {isRechargeMode ? (
          <Row style={{marginTop: "10px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Select virtual={false} style={{width: "100%"}} value={this.state.transaction.owner}
                onChange={(value) => {
                  this.updateTransactionField("owner", value);
                  this.updateTransactionField("application", "");
                  this.getApplications(value);
                  this.getUsers(value);
                }}>
                {
                  this.state.organizations.map((org, index) => <Option key={index} value={org.name}>{org.name}</Option>)
                }
              </Select>
            </Col>
          </Row>
        ) : (
          <Row style={{marginTop: "10px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input disabled={true} value={this.state.transaction.owner} onChange={e => {
                // this.updatePaymentField('organization', e.target.value);
              }} />
            </Col>
          </Row>
        )}
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.transaction.name} onChange={e => {
              // this.updatePaymentField('name', e.target.value);
            }} />
          </Col>
        </Row>
        {isRechargeMode ? (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Select virtual={false} style={{width: "100%"}} value={this.state.transaction.application}
                allowClear
                onChange={(value) => {
                  this.updateTransactionField("application", value || "");
                }}>
                {
                  this.state.applications.map((app, index) => <Option key={index} value={app.name}>{app.name}</Option>)
                }
              </Select>
            </Col>
          </Row>
        ) : (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input disabled={true} value={this.state.transaction.application} onChange={e => {
              }} />
            </Col>
          </Row>
        )}
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Domain"), i18next.t("provider:Domain - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.transaction.domain} onChange={e => {
              // this.updatePaymentField('domain', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Category"), i18next.t("provider:Category - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.transaction.category} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Type"), i18next.t("payment:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.transaction.type} onChange={e => {
              // this.updatePaymentField('type', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Subtype"), i18next.t("provider:Subtype - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.transaction.subtype} onChange={e => {
              // this.updatePaymentField('subtype', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Provider"), i18next.t("general:Provider - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.transaction.provider} onChange={e => {
              // this.updatePaymentField('provider', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Tag"), i18next.t("transaction:Tag - Tooltip"))} :
          </Col>
          <Col span={22} >
            {isRechargeMode ? (
              <Select virtual={false} style={{width: "100%"}}
                value={this.state.transaction.tag}
                onChange={(value) => {
                  this.updateTransactionField("tag", value);
                  if (value === "Organization") {
                    this.updateTransactionField("user", "");
                  }
                }}>
                <Option value="User">User</Option>
                <Option value="Organization">Organization</Option>
              </Select>
            ) : (
              <Input disabled={true} value={this.state.transaction.tag} onChange={e => {
              }} />
            )}
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:User"), i18next.t("general:User - Tooltip"))} :
          </Col>
          <Col span={22} >
            {isRechargeMode ? (
              <Select virtual={false} style={{width: "100%"}}
                value={this.state.transaction.user}
                disabled={this.state.transaction.tag === "Organization"}
                allowClear
                onChange={(value) => {
                  this.updateTransactionField("user", value || "");
                }}>
                {
                  this.state.users.map((user, index) => <Option key={index} value={user.name}>{user.name}</Option>)
                }
              </Select>
            ) : (
              <Input disabled={true} value={this.state.transaction.user} onChange={e => {
              }} />
            )}
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("transaction:Amount"), i18next.t("transaction:Amount - Tooltip"))} :
          </Col>
          <Col span={4} >
            <InputNumber disabled={!isRechargeMode} value={this.state.transaction.amount ?? 0} onChange={value => {
              this.updateTransactionField("amount", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Currency"), i18next.t("payment:Currency - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.transaction.currency} disabled={!isRechargeMode} onChange={(value => {
              this.updateTransactionField("currency", value);
            })}>
              {
                Setting.CurrencyOptions.map((item, index) => <Option key={index} value={item.id}>{Setting.getCurrencyWithFlag(item.id)}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Payment"), i18next.t("general:Payment - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.transaction.payment} onChange={e => {
              // this.updatePaymentField('amount', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:State"), i18next.t("general:State - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={true} value={this.state.transaction.state} onChange={e => {
              // this.updatePaymentField('state', e.target.value);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  render() {
    return (
      <div>
        {
          this.state.transaction !== null ? (
            <>
              {this.renderTransaction()}
              <div style={{marginTop: "20px", marginLeft: "40px"}}>
                <Button size="large" onClick={() => this.submitTransactionEdit(false)}>{i18next.t("general:Save")}</Button>
                <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitTransactionEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
                {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteTransaction()}>{i18next.t("general:Cancel")}</Button> : null}
              </div>
            </>
          ) : null
        }
      </div>
    );
  }
}

export default TransactionEditPage;
