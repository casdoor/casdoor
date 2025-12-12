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
import {Button, Descriptions, Spin} from "antd";
import i18next from "i18next";
import * as OrderBackend from "./backend/OrderBackend";
import * as ProductBackend from "./backend/ProductBackend";
import * as Setting from "./Setting";

class OrderPayPage extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(window.location.search);
    this.state = {
      owner: props?.match?.params?.organizationName ?? props?.match?.params?.owner ?? null,
      orderName: props?.match?.params?.orderName ?? null,
      order: null,
      product: null,
      paymentEnv: "",
      isProcessingPayment: false,
      isViewMode: params.get("view") === "true",
    };
  }

  getPaymentEnv() {
    let env = "";
    const ua = navigator.userAgent.toLowerCase();
    // Only support WeChat Pay in WeChat Browser for mobile devices
    if (ua.indexOf("micromessenger") !== -1 && ua.indexOf("mobile") !== -1) {
      env = "WechatBrowser";
    }
    this.setState({
      paymentEnv: env,
    });
  }

  componentDidMount() {
    this.getOrder();
    this.getPaymentEnv();
  }

  async getOrder() {
    if (!this.state.owner || !this.state.orderName) {
      return;
    }
    const res = await OrderBackend.getOrder(this.state.owner, this.state.orderName);
    if (res.status === "ok") {
      this.setState({
        order: res.data,
      }, () => {
        this.getProduct();
      });
    } else {
      Setting.showMessage("error", res.msg);
    }
  }

  async getProduct() {
    if (!this.state.order || !this.state.order.productName) {
      return;
    }
    const res = await ProductBackend.getProduct(this.state.order.owner, this.state.order.productName);
    if (res.status === "ok") {
      this.setState({
        product: res.data,
      });
    } else {
      Setting.showMessage("error", res.msg);
    }
  }

  getPrice(order) {
    return `${Setting.getCurrencySymbol(order?.currency)}${order?.price} (${Setting.getCurrencyText(order)})`;
  }

  getProductPrice(product) {
    return `${Setting.getCurrencySymbol(product?.currency)}${product?.price} (${Setting.getCurrencyText(product)})`;
  }

  // Call Wechat Pay via jsapi
  onBridgeReady(attachInfo) {
    const {WeixinJSBridge} = window;
    this.setState({
      isProcessingPayment: false,
    });
    WeixinJSBridge.invoke(
      "getBrandWCPayRequest", {
        "appId": attachInfo.appId,
        "timeStamp": attachInfo.timeStamp,
        "nonceStr": attachInfo.nonceStr,
        "package": attachInfo.package,
        "signType": attachInfo.signType,
        "paySign": attachInfo.paySign,
      },
      function(res) {
        if (res.err_msg === "get_brand_wcpay_request:ok") {
          Setting.goToLink(attachInfo.payment.successUrl);
          return;
        }
        if (res.err_msg === "get_brand_wcpay_request:cancel") {
          Setting.showMessage("error", i18next.t("product:Payment cancelled"));
        } else {
          Setting.showMessage("error", i18next.t("product:Payment failed"));
        }
      }
    );
  }

  // In WeChat browser, call this function to pay via jsapi
  callWechatPay(attachInfo) {
    const {WeixinJSBridge} = window;
    if (typeof WeixinJSBridge === "undefined") {
      document.addEventListener("WeixinJSBridgeReady", () => this.onBridgeReady(attachInfo), false);
    } else {
      this.onBridgeReady(attachInfo);
    }
  }

  payOrder(provider) {
    const {product, order} = this.state;
    if (!product || !order) {
      return;
    }

    this.setState({
      isProcessingPayment: true,
    });

    OrderBackend.payOrder(order.owner, order.name, provider.name, this.state.paymentEnv)
      .then((res) => {
        if (res.status === "ok") {
          const payment = res.data;
          const attachInfo = res.data2;

          let payUrl = payment.payUrl;
          if (provider.type === "WeChat Pay") {
            if (this.state.paymentEnv === "WechatBrowser") {
              attachInfo.payment = payment;
              this.callWechatPay(attachInfo);
              return;
            }
            payUrl = `/qrcode/${payment.owner}/${payment.name}?providerName=${provider.name}&payUrl=${encodeURIComponent(payment.payUrl)}&successUrl=${encodeURIComponent(payment.successUrl)}`;
          }
          Setting.goToLink(payUrl);
        } else {
          Setting.showMessage("error", `${i18next.t("product:Payment failed")}: ${res.msg}`);
          this.setState({
            isProcessingPayment: false,
          });
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        this.setState({
          isProcessingPayment: false,
        });
      });
  }

  getPayButton(provider, onClick) {
    const providerTypeMap = {
      "Dummy": i18next.t("product:Dummy"),
      "Alipay": i18next.t("product:Alipay"),
      "WeChat Pay": i18next.t("product:WeChat Pay"),
      "PayPal": i18next.t("product:PayPal"),
      "Stripe": i18next.t("product:Stripe"),
      "AirWallex": i18next.t("product:AirWallex"),
    };
    const text = providerTypeMap[provider.type] || provider.type;

    return (
      <Button style={{height: "50px", borderWidth: "2px"}} shape="round" icon={
        <img style={{marginRight: "10px"}} width={36} height={36} src={Setting.getProviderLogoURL(provider)} alt={provider.displayName} />
      } size={"large"} onClick={onClick}>
        {text}
      </Button>
    );
  }

  renderProviderButton(provider) {
    return (
      <span key={provider.name} style={{width: "200px", marginRight: "20px", marginBottom: "10px"}}>
        {this.getPayButton(provider, () => this.payOrder(provider))}
      </span>
    );
  }

  renderPaymentMethods() {
    const {product} = this.state;
    if (!product || !product.providerObjs || product.providerObjs.length === 0) {
      return <div>{i18next.t("product:There is no payment channel for this product.")}</div>;
    }

    return product.providerObjs.map(provider => {
      return this.renderProviderButton(provider);
    });
  }

  render() {
    const {order, product} = this.state;

    if (!order || !product) {
      return null;
    }

    const isSubscriptionOrder = order.pricingName && order.planName;

    return (
      <div className="login-content">
        <Spin spinning={this.state.isProcessingPayment} size="large" tip={i18next.t("product:Processing payment...")} style={{paddingTop: "10%"}} >
          <div style={{marginBottom: "20px"}}>
            <Descriptions title={<span style={Setting.isMobile() ? {fontSize: 18} : {fontSize: 24}}>{i18next.t("order:Order Information")}</span>} bordered column={3}>
              <Descriptions.Item label={i18next.t("order:Order ID")} span={3}>
                <span style={{fontSize: 16}}>
                  {order.name}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label={i18next.t("order:Order Status")}>
                <span style={{fontSize: 16}}>
                  {order.state}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label={i18next.t("general:Created time")}>
                <span style={{fontSize: 16}}>
                  {Setting.getFormattedDate(order.createdTime)}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label={i18next.t("general:User")}>
                <span style={{fontSize: 16}}>
                  {order.user}
                </span>
              </Descriptions.Item>
            </Descriptions>
          </div>

          <div style={{marginBottom: "20px"}}>
            <Descriptions title={<span style={Setting.isMobile() ? {fontSize: 18} : {fontSize: 24}}>{i18next.t("product:Product Information")}</span>} bordered column={3}>
              <Descriptions.Item label={i18next.t("general:Name")} span={3}>
                <span style={{fontSize: 20}}>
                  {Setting.getLanguageText(product?.displayName)}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label={i18next.t("product:Image")} span={3}>
                <img src={product?.image} alt={Setting.getLanguageText(product?.displayName)} height={90} style={{marginBottom: "20px"}} />
              </Descriptions.Item>
              <Descriptions.Item label={i18next.t("product:Price")} span={3}>
                <span style={{fontSize: 18, fontWeight: "bold"}}>
                  {this.getProductPrice(product)}
                </span>
              </Descriptions.Item>
              <Descriptions.Item label={i18next.t("product:Detail")} span={3}>
                <span style={{fontSize: 16}}>{Setting.getLanguageText(product?.detail)}</span>
              </Descriptions.Item>
            </Descriptions>
          </div>

          {isSubscriptionOrder && (
            <div style={{marginBottom: "20px"}}>
              <Descriptions title={<span style={Setting.isMobile() ? {fontSize: 18} : {fontSize: 24}}>{i18next.t("subscription:Subscription Information")}</span>} bordered column={3}>
                <Descriptions.Item label={i18next.t("general:Plan")} span={3}>
                  <span style={{fontSize: 16}}>{order.planName}</span>
                </Descriptions.Item>
                <Descriptions.Item label={i18next.t("general:Pricing")} span={3}>
                  <span style={{fontSize: 16}}>{order.pricingName}</span>
                </Descriptions.Item>
              </Descriptions>
            </div>
          )}

          <div>
            <Descriptions title={<span style={Setting.isMobile() ? {fontSize: 18} : {fontSize: 24}}>{i18next.t("payment:Payment Information")}</span>} bordered column={3}>
              <Descriptions.Item label={i18next.t("product:Price")} span={3}>
                <span style={{fontSize: 28, color: "red", fontWeight: "bold"}}>
                  {this.getPrice(order)}
                </span>
              </Descriptions.Item>
              {!this.state.isViewMode && (
                <Descriptions.Item label={i18next.t("product:Pay")} span={3}>
                  {this.renderPaymentMethods()}
                </Descriptions.Item>
              )}
            </Descriptions>
          </div>
        </Spin>
      </div>
    );
  }
}

export default OrderPayPage;
