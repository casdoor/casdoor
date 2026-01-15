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
import {Button, Card, Col, Input, InputNumber, Row, Select} from "antd";
import * as CartBackend from "./backend/CartBackend";
import * as ProductBackend from "./backend/ProductBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as UserBackend from "./backend/UserBackend";

const {Option} = Select;

class CartEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      cartName: props.match.params.cartName,
      cart: null,
      products: [],
      organizations: [],
      users: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getCart();
    this.getOrganizations();
    this.getProducts(this.state.organizationName);
    this.getUsers(this.state.organizationName);
  }

  getCart() {
    CartBackend.getCart(this.state.organizationName, this.state.cartName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        this.setState({
          cart: res.data,
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

  getProducts(organizationName) {
    ProductBackend.getProducts(organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            products: res.data || [],
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  getUsers(organizationName) {
    UserBackend.getUsers(organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            users: res.data || [],
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  parseCartField(key, value) {
    if (["quantity"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateCartField(key, value) {
    value = this.parseCartField(key, value);

    const cart = this.state.cart;
    cart[key] = value;
    this.setState({
      cart: cart,
    });
  }

  renderCart() {
    const isAdmin = Setting.isLocalAdminUser(this.props.account);
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("general:New Cart") : i18next.t("general:Edit Cart")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitCartEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitCartEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteCart()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!isAdmin} value={this.state.cart.owner} onChange={(value => {
              this.updateCartField("owner", value);
              this.getProducts(value);
              this.getUsers(value);
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
            <Input value={this.state.cart.name} onChange={e => {
              this.updateCartField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.cart.displayName} onChange={e => {
              this.updateCartField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:User"), i18next.t("general:User - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!isAdmin} value={this.state.cart.user} onChange={(value => {
              this.updateCartField("user", value);
            })}>
              {
                this.state.users.map((user, index) => <Option key={index} value={user.name}>{user.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Product"), i18next.t("product:Product - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.cart.productName} onChange={(value => {
              this.updateCartField("productName", value);
            })}>
              {
                this.state.products.map((product, index) => <Option key={index} value={product.name}>{product.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Quantity"), i18next.t("product:Quantity - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber min={1} value={this.state.cart.quantity} onChange={value => {
              this.updateCartField("quantity", value);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitCartEdit(exitAfterSave) {
    const cart = Setting.deepCopy(this.state.cart);
    CartBackend.updateCart(this.state.organizationName, this.state.cartName, cart)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            cartName: this.state.cart.name,
          });
          if (exitAfterSave) {
            this.props.history.push("/carts");
          } else {
            this.props.history.push(`/carts/${this.state.cart.owner}/${this.state.cart.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteCart() {
    CartBackend.deleteCart(this.state.cart)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/carts");
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
          this.state.cart !== null ? this.renderCart() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.props.history.push("/carts")}>{i18next.t("general:Cancel")}</Button>
        </div>
      </div>
    );
  }
}

export default CartEditPage;
