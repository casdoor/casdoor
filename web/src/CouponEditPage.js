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
import Loading from "./common/Loading";
import {Button, Card, Col, DatePicker, Input, InputNumber, Row, Select, Tag} from "antd";
import * as CouponBackend from "./backend/CouponBackend";
import * as ProductBackend from "./backend/ProductBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import moment from "moment";

const {Option} = Select;

class CouponEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props?.organizationName ?? props?.match?.params?.organizationName ?? null,
      couponName: props?.match?.params?.couponName ?? null,
      coupon: null,
      organizations: [],
      products: [],
      mode: props?.location?.mode ?? "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getCoupon();
    this.getOrganizations();
  }

  getCoupon() {
    CouponBackend.getCoupon(this.state.organizationName, this.state.couponName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        this.setState({
          coupon: res.data,
        });

        this.getProducts(this.state.organizationName);
      });
  }

  getProducts(organizationName) {
    ProductBackend.getProducts(organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            products: res.data || [],
          });
        }
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

  parseCouponField(key, value) {
    if (["discount", "maxDiscount", "minOrderAmount"].includes(key)) {
      value = parseFloat(value) || 0;
    }
    return value;
  }

  updateCouponField(key, value) {
    value = this.parseCouponField(key, value);

    const coupon = this.state.coupon;
    coupon[key] = value;
    this.setState({
      coupon: coupon,
    });
  }

  renderCoupon() {
    const isViewMode = this.state.mode === "view";
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("coupon:New Coupon") : i18next.t("coupon:Edit Coupon")}&nbsp;&nbsp;&nbsp;&nbsp;
          {!isViewMode ? <Button onClick={() => this.submitCouponEdit(false)}>{i18next.t("general:Save")}</Button> : null}
          {!isViewMode ? <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitCouponEdit(true)}>{i18next.t("general:Save & Exit")}</Button> : null}
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.props.history.goBack()}>{i18next.t("general:Cancel")}</Button> : null}
          {!isViewMode && this.state.mode !== "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteCoupon()}>{i18next.t("general:Delete")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={isViewMode} value={this.state.coupon.owner} onChange={(value => {this.updateCouponField("owner", value);})}>{
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
            <Input disabled={isViewMode} value={this.state.coupon.name} onChange={e => {this.updateCouponField("name", e.target.value);}} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={isViewMode} value={this.state.coupon.displayName} onChange={e => {this.updateCouponField("displayName", e.target.value);}} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={isViewMode} value={this.state.coupon.description} onChange={e => {this.updateCouponField("description", e.target.value);}} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("invitation:Code"), i18next.t("invitation:Code - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled={isViewMode} value={this.state.coupon.code} onChange={e => {this.updateCouponField("code", e.target.value);}} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("coupon:Discount type"), i18next.t("coupon:Discount type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={isViewMode} value={this.state.coupon.discountType} onChange={value => {this.updateCouponField("discountType", value);}}>
              <Option key="percentage" value="percentage">{i18next.t("coupon:Percentage")}</Option>
              <Option key="fixed" value="fixed">{i18next.t("coupon:Fixed")}</Option>
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("coupon:Discount"), i18next.t("coupon:Discount - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber disabled={isViewMode} value={this.state.coupon.discount} min={0} max={this.state.coupon.discountType === "percentage" ? 100 : undefined} onChange={value => {this.updateCouponField("discount", value);}} />
            {this.state.coupon.discountType === "percentage" ? <span style={{marginLeft: "10px"}}>%</span> : null}
          </Col>
        </Row>
        {this.state.coupon.discountType === "percentage" ? (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("coupon:Max discount"), i18next.t("coupon:Max discount - Tooltip"))} :
            </Col>
            <Col span={22} >
              <InputNumber disabled={isViewMode} value={this.state.coupon.maxDiscount} min={0} onChange={value => {this.updateCouponField("maxDiscount", value);}} />
              <span style={{marginLeft: "10px"}}>{i18next.t("coupon:0 means unlimited")}</span>
            </Col>
          </Row>
        ) : null}
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Scope"), i18next.t("provider:Scope - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={isViewMode} value={this.state.coupon.scope} onChange={value => {this.updateCouponField("scope", value);}}>
              <Option key="universal" value="universal"><Tag color="blue">{i18next.t("coupon:Universal")}</Tag></Option>
              <Option key="product" value="product"><Tag color="green">{i18next.t("coupon:Product specific")}</Tag></Option>
              <Option key="user" value="user"><Tag color="orange">{i18next.t("coupon:User specific")}</Tag></Option>
            </Select>
          </Col>
        </Row>
        {this.state.coupon.scope === "product" ? (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Products"), i18next.t("coupon:Products - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Select virtual={false} mode="multiple" style={{width: "100%"}} disabled={isViewMode} value={this.state.coupon.products} onChange={value => {this.updateCouponField("products", value);}}>
                {this.state.products.map((product, index) => <Option key={index} value={product.name}>{product.displayName || product.name}</Option>)}
              </Select>
            </Col>
          </Row>
        ) : null}
        {this.state.coupon.scope === "user" ? (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Users"), i18next.t("general:Users - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Select virtual={false} mode="tags" style={{width: "100%"}} disabled={isViewMode} value={this.state.coupon.users} onChange={value => {this.updateCouponField("users", value);}}>
              </Select>
            </Col>
          </Row>
        ) : null}
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("product:Quantity"), i18next.t("product:Quantity - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber disabled={isViewMode} value={this.state.coupon.quantity} min={0} precision={0} onChange={value => {this.updateCouponField("quantity", value);}} />
            <span style={{marginLeft: "10px"}}>{i18next.t("coupon:0 means unlimited")}</span>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("coupon:Max usage per user"), i18next.t("coupon:Max usage per user - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber disabled={isViewMode} value={this.state.coupon.maxUsagePerUser} min={0} precision={0} onChange={value => {this.updateCouponField("maxUsagePerUser", value);}} />
            <span style={{marginLeft: "10px"}}>{i18next.t("coupon:0 means unlimited")}</span>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("subscription:Start time"), i18next.t("subscription:Start time - Tooltip"))} :
          </Col>
          <Col span={22} >
            <DatePicker showTime disabled={isViewMode} value={this.state.coupon.startTime ? moment(this.state.coupon.startTime) : null} onChange={(value) => {this.updateCouponField("startTime", value ? value.format() : "");}} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Expire time"), i18next.t("general:Expire time - Tooltip"))} :
          </Col>
          <Col span={22} >
            <DatePicker showTime disabled={isViewMode} value={this.state.coupon.expireTime ? moment(this.state.coupon.expireTime) : null} onChange={(value) => {this.updateCouponField("expireTime", value ? value.format() : "");}} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("coupon:Min order amount"), i18next.t("coupon:Min order amount - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber disabled={isViewMode} value={this.state.coupon.minOrderAmount} min={0} onChange={value => {this.updateCouponField("minOrderAmount", value);}} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("payment:Currency"), i18next.t("payment:Currency - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={isViewMode} value={this.state.coupon.currency} onChange={value => {this.updateCouponField("currency", value);}}>
              {Setting.CurrencyOptions.map((item, index) => <Option key={index} value={item.id}>{Setting.getCurrencyWithFlag(item.id)}</Option>)}
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:State"), i18next.t("general:State - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={isViewMode} value={this.state.coupon.state} onChange={value => {this.updateCouponField("state", value);}}>
              <Option key="Active" value="Active">{i18next.t("subscription:Active")}</Option>
              <Option key="Inactive" value="Inactive">{i18next.t("key:Inactive")}</Option>
              <Option key="Expired" value="Expired">{i18next.t("subscription:Expired")}</Option>
            </Select>
          </Col>
        </Row>
      </Card>
    );
  }

  submitCouponEdit(willExist) {
    const coupon = Setting.deepCopy(this.state.coupon);
    CouponBackend.updateCoupon(this.state.coupon.owner, this.state.couponName, coupon)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            couponName: this.state.coupon.name,
          });
          if (willExist) {
            this.props.history.push("/coupons");
          } else {
            this.props.history.push(`/coupons/${this.state.coupon.owner}/${this.state.coupon.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteCoupon() {
    CouponBackend.deleteCoupon(this.state.coupon)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/coupons");
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
          this.state.coupon !== null ? this.renderCoupon() : <Loading />
        }
      </div>
    );
  }
}

export default CouponEditPage;
