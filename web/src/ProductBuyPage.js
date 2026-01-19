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
import {Button, Descriptions, Divider, InputNumber, Radio, Space, Spin, Typography} from "antd";
import i18next from "i18next";
import * as ProductBackend from "./backend/ProductBackend";
import * as PlanBackend from "./backend/PlanBackend";
import * as PricingBackend from "./backend/PricingBackend";
import * as OrderBackend from "./backend/OrderBackend";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";

class ProductBuyPage extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(window.location.search);
    this.state = {
      classes: props,
      owner: props?.organizationName ?? props?.match?.params?.organizationName ?? props?.match?.params?.owner ?? null,
      productName: props?.productName ?? props?.match?.params?.productName ?? null,
      pricingName: props?.pricingName ?? props?.match?.params?.pricingName ?? null,
      planName: params.get("plan"),
      userName: params.get("user"),
      paymentEnv: "",
      product: null,
      pricing: props?.pricing ?? null,
      plan: null,
      isPlacingOrder: false,
      isAddingToCart: false,
      customPrice: 0,
    };
  }

  getPaymentEnv() {
    let env = "";
    const ua = navigator.userAgent.toLocaleLowerCase();
    // Only support Wechat Pay in Wechat Browser for mobile devices
    if (ua.indexOf("micromessenger") !== -1 && ua.indexOf("mobile") !== -1) {
      env = "WechatBrowser";
    }
    this.setState({
      paymentEnv: env,
    });
  }

  UNSAFE_componentWillMount() {
    this.getProduct();
    this.getPaymentEnv();
  }

  setStateAsync(state) {
    return new Promise((resolve, reject) => {
      this.setState(state, () => {
        resolve();
      });
    });
  }

  onUpdatePricing(pricing) {
    this.props.onUpdatePricing(pricing);
  }

  async getProduct() {
    if (!this.state.owner || (!this.state.productName && !this.state.pricingName)) {
      return ;
    }
    try {
      // load pricing & plan
      if (this.state.pricingName) {
        if (!this.state.planName || !this.state.userName) {
          return ;
        }
        let res = await PricingBackend.getPricing(this.state.owner, this.state.pricingName);
        if (res.status !== "ok") {
          throw new Error(res.msg);
        }
        const pricing = res.data;
        res = await PlanBackend.getPlan(this.state.owner, this.state.planName);
        if (res.status !== "ok") {
          throw new Error(res.msg);
        }
        const plan = res.data;
        await this.setStateAsync({
          pricing: pricing,
          plan: plan,
          productName: plan.product,
        });
        this.onUpdatePricing(pricing);
      }
      // load product
      const res = await ProductBackend.getProduct(this.state.owner, this.state.productName);
      if (res.status !== "ok") {
        throw new Error(res.msg);
      }
      this.setState({
        product: res.data,
      });

      if (res.data.isRecharge && res.data.rechargeOptions?.length > 0) {
        this.setState({
          customPrice: res.data.rechargeOptions[0],
        });
      }
    } catch (err) {
      Setting.showMessage("error", err.message);
      return;
    }
  }

  getProductObj() {
    if (this.props.product !== undefined) {
      return this.props.product;
    } else {
      return this.state.product;
    }
  }

  getPrice(product) {
    return `${Setting.getCurrencySymbol(product?.currency)}${product?.price} (${Setting.getCurrencyText(product)})`;
  }

  addToCart(product) {
    if (this.state.isAddingToCart) {
      return;
    }

    this.setState({isAddingToCart: true});

    const userOwner = this.props.account.owner;
    const userName = this.props.account.name;

    UserBackend.getUser(userOwner, userName)
      .then((res) => {
        if (res.status === "ok") {
          const user = res.data;
          const cart = user.cart || [];

          let actualPrice = product.price;
          if (product.isRecharge) {
            actualPrice = this.state.customPrice;
            if (actualPrice <= 0) {
              Setting.showMessage("error", i18next.t("product:Custom price should be greater than zero"));
              this.setState({isAddingToCart: false});
              return;
            }
          }

          if (cart.length > 0) {
            const firstItem = cart[0];
            if (firstItem.currency && product.currency && firstItem.currency !== product.currency) {
              Setting.showMessage("error", i18next.t("product:The currency of the product you are adding is different from the currency of the items in the cart"));
              this.setState({isAddingToCart: false});
              return;
            }
          }

          const existingItemIndex = cart.findIndex(item => item.name === product.name && item.price === actualPrice);

          if (existingItemIndex !== -1) {
            cart[existingItemIndex].quantity += 1;
          } else {
            const newProductInfo = {
              name: product.name,
              displayName: product.displayName,
              image: product.image,
              detail: product.detail,
              price: actualPrice,
              currency: product.currency,
              quantity: 1,
              isRecharge: product.isRecharge,
            };
            cart.push(newProductInfo);
          }

          user.cart = cart;
          UserBackend.updateUser(user.owner, user.name, user)
            .then((res) => {
              if (res.status === "ok") {
                Setting.showMessage("success", i18next.t("general:Successfully added"));
              } else {
                Setting.showMessage("error", res.msg);
              }
            })
            .catch((error) => {
              Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
            })
            .finally(() => {
              this.setState({isAddingToCart: false});
            });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${res.msg}`);
          this.setState({isAddingToCart: false});
        }
      })
      .catch((error) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        this.setState({isAddingToCart: false});
      });
  }

  placeOrder(product) {
    this.setState({
      isPlacingOrder: true,
    });

    const pricingName = this.state.pricingName || "";
    const planName = this.state.planName || "";
    const customPrice = this.state.customPrice || 0;

    const productInfos = [{
      name: product.name,
      price: product.isRecharge ? customPrice : product.price,
    }];

    OrderBackend.placeOrder(product.owner, productInfos, pricingName, planName, this.state.userName ?? "")
      .then((res) => {
        if (res.status === "ok") {
          const order = res.data;
          Setting.showMessage("success", i18next.t("product:Order created successfully"));
          // Redirect to order pay page
          Setting.goToLink(`/orders/${order.owner}/${order.name}/pay`);
        } else {
          Setting.showMessage("error", `${i18next.t("product:Failed to create order")}: ${res.msg}`);
          this.setState({
            isPlacingOrder: false,
          });
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        this.setState({
          isPlacingOrder: false,
        });
      });
  }

  renderRechargeInput(product) {
    const hasOptions = product.rechargeOptions && product.rechargeOptions.length > 0;
    const disableCustom = product.disableCustomRecharge;

    if (!hasOptions && disableCustom) {
      return (
        <Typography.Text type="danger">
          {i18next.t("product:This product is currently not purchasable (No options available)")}
        </Typography.Text>
      );
    }

    return (
      <Space direction="vertical" style={{width: "100%"}}>
        {hasOptions && (
          <>
            <div>
              <span style={{marginRight: "10px", fontSize: 16}}>
                {i18next.t("product:Select amount")}:
              </span>
              <Radio.Group
                value={this.state.customPrice}
                onChange={(e) => {this.setState({customPrice: e.target.value});}}
              >
                <Space wrap>
                  {product.rechargeOptions.map((amount, index) => (
                    <Radio.Button key={index} value={amount}>
                      {Setting.getCurrencySymbol(product.currency)}{amount}
                    </Radio.Button>
                  ))}
                </Space>
              </Radio.Group>
            </div>
            {!disableCustom && <Divider style={{margin: "10px 0"}}>{i18next.t("general:Or")}</Divider>}
          </>
        )}
        <Space>
          <span style={{fontSize: 16}}>
            {i18next.t("product:Amount")}:
          </span>
          <InputNumber
            min={0}
            value={this.state.customPrice}
            onChange={(e) => {this.setState({customPrice: e});}}
            disabled={disableCustom}
          />
          <span style={{fontSize: 16}}>{Setting.getCurrencyText(product)}</span>
        </Space>
      </Space>
    );
  }

  renderPlaceOrderButton(product) {
    if (product === undefined || product === null) {
      return null;
    }

    if (product.state !== "Published") {
      return i18next.t("product:This product is currently not in sale.");
    }

    const hasOptions = product.rechargeOptions && product.rechargeOptions.length > 0;
    const disableCustom = product.disableCustomRecharge;
    const isRechargeUnpurchasable = product.isRecharge && !hasOptions && disableCustom;
    const isSubscription = product.tag === "Subscription";

    return (
      <div style={{display: "flex", justifyContent: "center", alignItems: "center", gap: "20px"}}>
        <Button
          type="primary"
          size="large"
          style={{
            height: "50px",
            fontSize: "18px",
            borderRadius: "30px",
            paddingLeft: "60px",
            paddingRight: "60px",
          }}
          onClick={() => this.placeOrder(product)}
          disabled={this.state.isPlacingOrder || isRechargeUnpurchasable}
          loading={this.state.isPlacingOrder}
        >
          {i18next.t("order:Place Order")}
        </Button>
        {!isSubscription && (
          <Button
            type="primary"
            size="large"
            style={{
              height: "50px",
              fontSize: "18px",
              borderRadius: "30px",
              paddingLeft: "30px",
              paddingRight: "30px",
            }}
            onClick={() => this.addToCart(product)}
            disabled={isRechargeUnpurchasable || this.state.isAddingToCart}
            loading={this.state.isAddingToCart}
          >
            {i18next.t("product:Add to cart")}
          </Button>
        )}
      </div>
    );
  }

  render() {
    const product = this.getProductObj();
    const placeOrderButton = this.renderPlaceOrderButton(product);

    if (product === null) {
      return null;
    }

    return (
      <div className="login-content">
        <Spin spinning={this.state.isPlacingOrder} size="large" tip={i18next.t("product:Placing order...")} style={{paddingTop: "10%"}} >
          <Descriptions title={<span style={Setting.isMobile() ? {fontSize: 20} : {fontSize: 28}}>{i18next.t("product:Buy Product")}</span>} bordered>
            <Descriptions.Item label={i18next.t("general:Name")} span={3}>
              <span style={{fontSize: 25}}>
                {Setting.getLanguageText(product?.displayName)}
              </span>
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:Detail")}><span style={{fontSize: 16}}>{Setting.getLanguageText(product?.detail)}</span></Descriptions.Item>
            <Descriptions.Item label={i18next.t("user:Tag")}><span style={{fontSize: 16}}>{product?.tag}</span></Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:SKU")}><span style={{fontSize: 16}}>{product?.name}</span></Descriptions.Item>
            <Descriptions.Item label={i18next.t("product:Image")} span={3}>
              <img src={product?.image} alt={product?.name} height={90} style={{marginBottom: "20px"}} />
            </Descriptions.Item>
            {
              product.isRecharge ? (
                <Descriptions.Item span={3} label={i18next.t("product:Price")}>
                  {this.renderRechargeInput(product)}
                </Descriptions.Item>
              ) : (
                <React.Fragment>
                  <Descriptions.Item label={i18next.t("product:Price")}>
                    <span style={{fontSize: 28, color: "red", fontWeight: "bold"}}>
                      {
                        this.getPrice(product)
                      }
                    </span>
                  </Descriptions.Item>
                  <Descriptions.Item label={i18next.t("product:Quantity")}><span style={{fontSize: 16}}>{product?.quantity}</span></Descriptions.Item>
                  <Descriptions.Item label={i18next.t("product:Sold")}><span style={{fontSize: 16}}>{product?.sold}</span></Descriptions.Item>
                </React.Fragment>
              )
            }
            <Descriptions.Item label={i18next.t("order:Place Order")} span={3}>
              <div style={{display: "flex", justifyContent: "center", alignItems: "center", minHeight: "80px"}}>
                {placeOrderButton}
              </div>
            </Descriptions.Item>
          </Descriptions>
        </Spin>
      </div>
    );
  }
}

export default ProductBuyPage;
