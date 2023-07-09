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
import {Button, Descriptions, Modal, Spin} from "antd";
import {CheckCircleTwoTone} from "@ant-design/icons";
import i18next from "i18next";
import * as ProductBackend from "./backend/ProductBackend";
import * as Setting from "./Setting";

class ProductBuyPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      productName: props.match?.params.productName,
      product: null,
      isPlacingOrder: false,
      qrCodeModalProvider: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getProduct();
  }

  getProduct() {
    if (this.state.productName === undefined) {
      return;
    }

    ProductBackend.getProduct(this.props.account.owner, this.state.productName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          product: res,
        });
      });
  }

  getProductObj() {
    if (this.props.product !== undefined) {
      return this.props.product;
    } else {
      return this.state.product;
    }
  }

  getCurrencySymbol(product) {
    if (product?.currency === "USD") {
      return "$";
    } else if (product?.currency === "CNY") {
      return "￥";
    } else {
      return "(Unknown currency)";
    }
  }

  getCurrencyText(product) {
    if (product?.currency === "USD") {
      return i18next.t("product:USD");
    } else if (product?.currency === "CNY") {
      return i18next.t("product:CNY");
    } else {
      return "(Unknown currency)";
    }
  }

  getPrice(product) {
    return `${this.getCurrencySymbol(product)}${product?.price} (${this.getCurrencyText(product)})`;
  }

  buyProduct(product, provider) {
    if (provider.clientId.startsWith("http")) {
      this.setState({
        qrCodeModalProvider: provider,
      });
      return;
    }

    this.setState({
      isPlacingOrder: true,
    });

    ProductBackend.buyProduct(this.state.product.owner, this.state.productName, provider.name)
      .then((res) => {
        if (res.status === "ok") {
          const payUrl = res.data;
          Setting.goToLink(payUrl);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);

          this.setState({
            isPlacingOrder: false,
          });
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderQrCodeModal() {
    if (this.state.qrCodeModalProvider === undefined || this.state.qrCodeModalProvider === null) {
      return null;
    }

    return (
      <Modal title={
        <div>
          <CheckCircleTwoTone twoToneColor="rgb(45,120,213)" />
          {" " + i18next.t("product:Please scan the QR code to pay")}
        </div>
      }
      open={this.state.qrCodeModalProvider !== undefined && this.state.qrCodeModalProvider !== null}
      onOk={() => {
        Setting.goToLink(this.state.product.returnUrl);
      }}
      onCancel={() => {
        this.setState({
          qrCodeModalProvider: null,
        });
      }}
      okText={i18next.t("product:I have completed the payment")}
      cancelText={i18next.t("general:Cancel")}>
        <p key={this.state.qrCodeModalProvider?.name}>
          {
            i18next.t("product:Please provide your username in the remark")
          }
          :&nbsp;&nbsp;
          {
            Setting.getTag("default", this.props.account.name)
          }
          <br />
          <br />
          <img src={this.state.qrCodeModalProvider?.clientId} alt={this.state.qrCodeModalProvider?.name} width={"472px"} style={{marginBottom: "20px"}} />
        </p>
      </Modal>
    );
  }

  getPayButton(provider) {
    let text = provider.type;
    if (provider.type === "Dummy") {
      text = i18next.t("product:Dummy");
    } else if (provider.type === "Alipay") {
      text = i18next.t("product:Alipay");
    } else if (provider.type === "WeChat Pay") {
      text = i18next.t("product:WeChat Pay");
    } else if (provider.type === "PayPal") {
      text = i18next.t("product:PayPal");
    }

    return (
      <Button style={{height: "50px", borderWidth: "2px"}} shape="round" icon={
        <img style={{marginRight: "10px"}} width={36} height={36} src={Setting.getProviderLogoURL(provider)} alt={provider.displayName} />
      } size={"large"} >
        {
          text
        }
      </Button>
    );
  }

  renderProviderButton(provider, product) {
    return (
      <span key={provider.name} style={{width: "200px", marginRight: "20px", marginBottom: "10px"}}>
        <span style={{width: "200px", cursor: "pointer"}} onClick={() => this.buyProduct(product, provider)}>
          {
            this.getPayButton(provider)
          }
        </span>
      </span>
    );
  }

  renderPay(product) {
    if (product === undefined || product === null) {
      return null;
    }

    if (product.state !== "Published") {
      return i18next.t("product:This product is currently not in sale.");
    }
    if (product.providerObjs.length === 0) {
      return i18next.t("product:There is no payment channel for this product.");
    }

    return product.providerObjs.map(provider => {
      return this.renderProviderButton(provider, product);
    });
  }

  render() {
    const product = this.getProductObj();

    if (product === null) {
      return null;
    }

    return (
      <div>
        <Spin spinning={this.state.isPlacingOrder} size="large" tip={i18next.t("product:Placing order...")} style={{paddingTop: "10%"}} >
          <Descriptions title={i18next.t("product:Buy Product")} bordered>
            <Descriptions.Item label={i18next.t("general:Name")} span={3}>
              <span style={{fontSize: 28}}>
                {Setting.getLanguageText(product?.displayName)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:Detail")}><span style={{fontSize: 16}}>{Setting.getLanguageText(product?.detail)}</span></Descriptions.Item>
            <Descriptions.Item label={i18next.t("user:Tag")}><span style={{fontSize: 16}}>{product?.tag}</span></Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:SKU")}><span style={{fontSize: 16}}>{product?.name}</span></Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:Image")} span={3}>
              <img src={product?.image} alt={product?.name} height={90} style={{marginBottom: "20px"}} />
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:Price")}>
              <span style={{fontSize: 28, color: "red", fontWeight: "bold"}}>
                {
                  this.getPrice(product)
                }
              </span>
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:Quantity")}><span style={{fontSize: 16}}>{product?.quantity}</span></Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:Sold")}><span style={{fontSize: 16}}>{product?.sold}</span></Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:Pay")} span={3}>
              {
                this.renderPay(product)
              }
            </Descriptions.Item>
          </Descriptions>
        </Spin>
        {
          this.renderQrCodeModal()
        }
      </div>
    );
  }
}

export default ProductBuyPage;
