// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
import * as ProductBackend from "./backend/ProductBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import {LinkOutlined} from "@ant-design/icons";
import * as ProviderBackend from "./backend/ProviderBackend";
import ProductBuyPage from "./ProductBuyPage";
import * as OrganizationBackend from "./backend/OrganizationBackend";

const {Option} = Select;

class ProductEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      productName: props.match.params.productName,
      product: null,
      providers: [],
      organizations: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getProduct();
    this.getOrganizations();
    this.getPaymentProviders(this.state.organizationName);
  }

  getProduct() {
    ProductBackend.getProduct(this.state.organizationName, this.state.productName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        this.setState({
          product: res.data,
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

  getPaymentProviders(organizationName) {
    ProviderBackend.getProviders(organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            providers: res.data.filter(provider => provider.category === "Payment"),
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  parseProductField(key, value) {
    if ([""].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateProductField(key, value) {
    value = this.parseProductField(key, value);

    const product = this.state.product;
    product[key] = value;
    this.setState({
      product: product,
    });
  }

  renderProduct() {
    const isCreatedByPlan = this.state.product.tag === "auto_created_product_for_plan";
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("product:New Product") : i18next.t("product:Edit Product")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitProductEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitProductEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteProduct()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account) || isCreatedByPlan} value={this.state.product.owner} onChange={(value => {this.updateProductField("owner", value);})}>
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
            <Input value={this.state.product.name} disabled={isCreatedByPlan} onChange={e => {
              this.updateProductField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.product.displayName} onChange={e => {
              this.updateProductField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Image"), i18next.t("product:Image - Tooltip"))} :
          </Col>
          <Col span={22} style={(Setting.isMobile()) ? {maxWidth: "100%"} : {}}>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined />} value={this.state.product.image} onChange={e => {
                  this.updateProductField("image", e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" rel="noreferrer" href={this.state.product.image}>
                  <img src={this.state.product.image} alt={this.state.product.image} height={90} style={{marginBottom: "20px"}} />
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("user:Tag"), i18next.t("product:Tag - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.product.tag} disabled={isCreatedByPlan} onChange={e => {
              this.updateProductField("tag", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Detail"), i18next.t("product:Detail - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.product.detail} onChange={e => {
              this.updateProductField("detail", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.product.description} onChange={e => {
              this.updateProductField("description", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Currency"), i18next.t("payment:Currency - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.product.currency} disabled={isCreatedByPlan} onChange={(value => {
              this.updateProductField("currency", value);
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
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Is recharge"), i18next.t("product:Is recharge - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Switch checked={this.state.product.isRecharge} onChange={value => {
              this.updateProductField("isRecharge", value);
            }} />
          </Col>
        </Row>
        {
          this.state.product.isRecharge ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("product:Price"), i18next.t("product:Price - Tooltip"))} :
              </Col>
              <Col span={22} >
                <InputNumber value={this.state.product.price} disabled={isCreatedByPlan} onChange={value => {
                  this.updateProductField("price", value);
                }} />
              </Col>
            </Row>
          )}
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Quantity"), i18next.t("product:Quantity - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.product.quantity} disabled={isCreatedByPlan} onChange={value => {
              this.updateProductField("quantity", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Sold"), i18next.t("product:Sold - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.product.sold} disabled={isCreatedByPlan} onChange={value => {
              this.updateProductField("sold", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Payment providers"), i18next.t("product:Payment providers - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}} disabled={isCreatedByPlan} value={this.state.product.providers} onChange={(value => {this.updateProductField("providers", value);})}>
              {
                this.state.providers.map((provider, index) => <Option key={index} value={provider.name}>{provider.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Return URL"), i18next.t("product:Return URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.product.returnUrl} onChange={e => {
              this.updateProductField("returnUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Return Type"), i18next.t("product:Return Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.product.returnType} onChange={(value => {
              this.updateProductField("returnType", value);
            })}>
              {
                [
                  {id: "manualRedirect", name: i18next.t("product:Return Type - manualRedirect")},
                  {id: "directRedirect", name: i18next.t("product:Return Type - directRedirect")},
                  {id: "paidAutoRedirect", name: i18next.t("product:Return Type - paidAutoRedirect")},
                  {id: "paidAutoClose", name: i18next.t("product:Return Type - paidAutoClose")},
                  {id: "autoClose", name: i18next.t("product:Return Type - autoClose")},
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:State"), i18next.t("general:State - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.product.state} onChange={(value => {
              this.updateProductField("state", value);
            })}>
              {
                [
                  {id: "Published", name: "Published"},
                  {id: "Draft", name: "Draft"},
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
          </Col>
          {
            this.renderPreview()
          }
        </Row>
      </Card>
    );
  }

  renderPreview() {
    const buyUrl = `/products/${this.state.product.owner}/${this.state.product.name}/buy`;
    return (
      <Col span={22} style={{display: "flex", flexDirection: "column"}}>
        <a style={{marginBottom: "10px", display: "flex"}} target="_blank" rel="noreferrer" href={buyUrl}>
          <Button type="primary">{i18next.t("product:Test buy page..")}</Button>
        </a>
        <br />
        <br />
        <div style={{width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", alignItems: "center", overflow: "auto", flexDirection: "column", flex: "auto"}}>
          <ProductBuyPage product={this.state.product} />
        </div>
      </Col>
    );
  }

  submitProductEdit(exitAfterSave) {
    const product = Setting.deepCopy(this.state.product);
    ProductBackend.updateProduct(this.state.organizationName, this.state.productName, product)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            productName: this.state.product.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/products");
          } else {
            this.props.history.push(`/products/${this.state.product.owner}/${this.state.product.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateProductField("name", this.state.productName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteProduct() {
    ProductBackend.deleteProduct(this.state.product)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/products");
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
          this.state.product !== null ? this.renderProduct() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitProductEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitProductEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteProduct()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default ProductEditPage;
