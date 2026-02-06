// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
import * as OrderBackend from "./backend/OrderBackend";
import * as ProductBackend from "./backend/ProductBackend";
import * as UserBackend from "./backend/UserBackend";
import * as PaymentBackend from "./backend/PaymentBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const {Option} = Select;

class OrderEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      orderName: props.match.params.orderName,
      order: null,
      products: [],
      users: [],
      payments: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrder();
    this.getProducts();
    this.getUsers();
    this.getPayments();
  }

  getOrder() {
    OrderBackend.getOrder(this.state.organizationName, this.state.orderName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        this.setState({
          order: res.data,
        });
      });
  }

  getProducts() {
    ProductBackend.getProducts(this.state.organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            products: res.data,
          });
        } else {
          Setting.showMessage("error", `Failed to get products: ${res.msg}`);
        }
      });
  }

  getUsers() {
    UserBackend.getUsers(this.state.organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            users: res.data,
          });
        } else {
          Setting.showMessage("error", `Failed to get users: ${res.msg}`);
        }
      });
  }

  getPayments() {
    PaymentBackend.getPayments(this.state.organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            payments: res.data,
          });
        } else {
          Setting.showMessage("error", `Failed to get payments: ${res.msg}`);
        }
      });
  }

  parseOrderField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateOrderField(key, value) {
    value = this.parseOrderField(key, value);

    const order = this.state.order;
    order[key] = value;
    this.setState({
      order: order,
    });
  }

  renderOrder() {
    const isViewMode = this.state.mode === "view";
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("order:New Order") : (isViewMode ? i18next.t("order:View Order") : i18next.t("order:Edit Order"))}&nbsp;&nbsp;&nbsp;&nbsp;
          {!isViewMode && (<>
            <Button onClick={() => this.submitOrderEdit(false)}>{i18next.t("general:Save")}</Button>
            <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitOrderEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
            {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteOrder()}>{i18next.t("general:Cancel")}</Button> : null}
          </>)}
        </div>
      } style={{marginLeft: "5px"}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Organization")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.order.owner} disabled />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.order.name} disabled={isViewMode} onChange={e => {
              this.updateOrderField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Display name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.order.displayName} disabled={isViewMode} onChange={e => {
              this.updateOrderField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Products")}:
          </Col>
          <Col span={22} >
            <Select
              mode="multiple"
              style={{width: "100%"}}
              value={this.state.order?.products || []}
              disabled={isViewMode}
              allowClear
              options={(this.state.products || [])
                .map((p) => ({
                  label: Setting.getLanguageText(p?.displayName) || p?.name,
                  value: p?.name,
                }))
                .filter((o) => o.value)}
              onChange={(value) => {
                this.updateOrderField("products", value);
              }}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:User")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.order.user} disabled={isViewMode} onChange={(value) => {
              this.updateOrderField("user", value);
            }}>
              {
                this.state.users?.map((user, index) => <Option key={index} value={user.name}>{user.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Payment")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.order.payment} disabled={isViewMode} onChange={(value) => {
              this.updateOrderField("payment", value);
            }}>
              <Option value="">{"(empty)"}</Option>
              {
                this.state.payments?.map((payment, index) => <Option key={index} value={payment.name}>{payment.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:State")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.order.state} disabled={isViewMode} onChange={(value) => {
              this.updateOrderField("state", value);
            }}>
              {
                [
                  {id: "Created", name: "Created"},
                  {id: "Paid", name: "Paid"},
                  {id: "Delivered", name: "Delivered"},
                  {id: "Completed", name: "Completed"},
                  {id: "Canceled", name: "Canceled"},
                  {id: "Expired", name: "Expired"},
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("payment:Message")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.order.message} onChange={e => {
              this.updateOrderField("message", e.target.value);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitOrderEdit(exitAfterSave) {
    const order = Setting.deepCopy(this.state.order);
    OrderBackend.updateOrder(this.state.organizationName, this.state.orderName, order)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            orderName: this.state.order.name,
          });
          if (exitAfterSave) {
            this.props.history.push("/orders");
          } else {
            this.props.history.push(`/orders/${this.state.order.owner}/${this.state.order.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteOrder() {
    OrderBackend.deleteOrder(this.state.order)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/orders");
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
          this.state.order !== null ? this.renderOrder() : null
        }
        {this.state.mode !== "view" && (
          <div style={{marginTop: "20px", marginLeft: "40px"}}>
            <Button size="large" onClick={() => this.submitOrderEdit(false)}>{i18next.t("general:Save")}</Button>
            <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitOrderEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
            {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteOrder()}>{i18next.t("general:Cancel")}</Button> : null}
          </div>
        )}
      </div>
    );
  }
}

export default OrderEditPage;
