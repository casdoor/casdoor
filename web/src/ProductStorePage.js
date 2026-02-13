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
import {Button, Card, Col, Row, Tag, Typography} from "antd";
import * as Setting from "./Setting";
import * as ProductBackend from "./backend/ProductBackend";
import * as UserBackend from "./backend/UserBackend";
import i18next from "i18next";
import {FloatingCartButton, QuantityStepper} from "./common/product/CartControls";

const {Text, Title} = Typography;

class ProductStorePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      products: [],
      loading: true,
      addingToCartProducts: [],
      productQuantities: {},
      cartItemCount: 0,
    };
  }

  componentDidMount() {
    if (!this.props.account) {
      return;
    }

    this.getProducts();
    this.getCartItemCount();
  }

  componentDidUpdate(prevProps) {
    if (!prevProps.account && this.props.account) {
      this.getProducts();
      this.getCartItemCount();
    }
  }

  getCartItemCount() {
    if (!this.props.account) {
      return;
    }

    const userOwner = this.props.account.owner;
    const userName = this.props.account.name;
    UserBackend.getUser(userOwner, userName).then((res) => {
      if (res.status === "ok" && res.data.cart) {
        this.setState({
          cartItemCount: res.data.cart.length,
        });
      }
    });
  }

  updateProductQuantity(productName, value) {
    this.setState(prevState => ({
      productQuantities: {
        ...prevState.productQuantities,
        [productName]: value,
      },
    }));
  }

  getProducts() {
    if (!this.props.account) {
      return;
    }

    const pageSize = 100; // Max products to display in the store
    const owner = Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account);
    this.setState({loading: true});
    ProductBackend.getProducts(owner, 1, pageSize, "state", "Published", "", "")
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            products: res.data,
            loading: false,
          });
        } else {
          Setting.showMessage("error", res.msg);
          this.setState({loading: false});
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        this.setState({loading: false});
      });
  }

  addToCart(product) {
    if (this.state.addingToCartProducts.includes(product.name)) {
      return;
    }

    this.setState(prevState => ({addingToCartProducts: [...prevState.addingToCartProducts, product.name]}));

    const userOwner = this.props.account.owner;
    const userName = this.props.account.name;

    UserBackend.getUser(userOwner, userName)
      .then((res) => {
        if (res.status === "ok") {
          const user = res.data;
          const cart = user.cart || [];

          if (cart.length > 0) {
            const firstItem = cart[0];
            if (firstItem.currency && product.currency && firstItem.currency !== product.currency) {
              Setting.showMessage("error", i18next.t("product:The currency of the product you are adding is different from the currency of the items in the cart"));
              this.setState(prevState => ({addingToCartProducts: prevState.addingToCartProducts.filter(name => name !== product.name)}));
              return;
            }
          }

          if (product.isRecharge) {
            Setting.showMessage("error", i18next.t("product:Recharge products need to go to the product detail page to set custom amount"));
            this.setState(prevState => ({addingToCartProducts: prevState.addingToCartProducts.filter(name => name !== product.name)}));
            return;
          }

          const existingItemIndex = cart.findIndex(item => item.name === product.name);
          const quantityToAdd = this.state.productQuantities[product.name] || 1;

          if (existingItemIndex !== -1) {
            cart[existingItemIndex].quantity = (cart[existingItemIndex].quantity ?? 1) + quantityToAdd;
          } else {
            const newCartProductInfo = {
              name: product.name,
              currency: product.currency,
              pricingName: "",
              planName: "",
              quantity: quantityToAdd,
              addTime: new Date().toISOString(),
            };
            cart.push(newCartProductInfo);
          }

          user.cart = cart;
          UserBackend.updateUser(user.owner, user.name, user)
            .then((res) => {
              if (res.status === "ok") {
                Setting.showMessage("success", i18next.t("general:Successfully added"));
                this.setState({
                  cartItemCount: cart.length,
                });
              } else {
                Setting.showMessage("error", res.msg);
              }
            })
            .catch(error => {
              Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
            })
            .finally(() => {
              this.setState(prevState => ({addingToCartProducts: prevState.addingToCartProducts.filter(name => name !== product.name)}));
            });
        } else {
          Setting.showMessage("error", res.msg);
          this.setState(prevState => ({addingToCartProducts: prevState.addingToCartProducts.filter(name => name !== product.name)}));
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        this.setState(prevState => ({addingToCartProducts: prevState.addingToCartProducts.filter(name => name !== product.name)}));
      });
  }

  handleBuyProduct(product) {
    const quantity = this.state.productQuantities[product.name] || 1;
    this.props.history.push(`/products/${product.owner}/${product.name}/buy?quantity=${quantity}`);
  }

  renderProductCard(product) {
    const isAdding = this.state.addingToCartProducts.includes(product.name);
    const quantity = this.state.productQuantities[product.name] || 1;

    return (
      <Col xs={24} sm={12} md={8} lg={6} key={`${product.owner}/${product.name}`} style={{marginBottom: "20px"}}>
        <Card
          hoverable
          onClick={() => this.handleBuyProduct(product)}
          style={{cursor: "pointer", height: "100%", display: "flex", flexDirection: "column"}}
          cover={
            <div style={{height: "200px", overflow: "hidden", display: "flex", alignItems: "center", justifyContent: "center", backgroundColor: "#f0f0f0"}}>
              <img
                alt={product.displayName}
                src={product.image}
                style={{width: "100%", height: "100%", objectFit: "contain"}}
              />
            </div>
          }
          actions={[
            <div key="actions" style={{display: "flex", justifyContent: "center", gap: "10px", width: "100%", padding: "0 10px"}} onClick={(e) => e.stopPropagation()}>
              {!product.isRecharge && (
                <>
                  <QuantityStepper
                    value={quantity}
                    min={1}
                    onIncrease={() => this.updateProductQuantity(product.name, quantity + 1)}
                    onDecrease={() => this.updateProductQuantity(product.name, Math.max(1, quantity - 1))}
                    onChange={(val) => this.updateProductQuantity(product.name, val || 1)}
                    disabled={isAdding}
                    style={{
                      height: "45px",
                      fontSize: "16px",
                      width: "120px",
                    }}
                  />
                  <Button
                    key="add"
                    type="default"
                    onClick={(e) => {
                      e.stopPropagation();
                      this.addToCart(product);
                    }}
                    style={{
                      width: "150px",
                      height: "45px",
                      fontSize: "16px",
                    }}
                    disabled={isAdding}
                    loading={isAdding}
                  >
                    {i18next.t("product:Add to cart")}
                  </Button>
                </>
              )}
              <Button
                key="buy"
                type="primary"
                onClick={(e) => {
                  e.stopPropagation();
                  this.handleBuyProduct(product);
                }}
                style={{
                  width: "150px",
                  height: "45px",
                  fontSize: "16px",
                }}
              >
                {i18next.t("product:Buy")}
              </Button>
            </div>,
          ]}
          bodyStyle={{flex: 1, display: "flex", flexDirection: "column"}}
        >
          <div style={{flex: 1, display: "flex", flexDirection: "column"}}>
            <Title level={5} ellipsis={{rows: 2}} style={{margin: "0 0 12px 0", minHeight: "44px", fontWeight: 600}}>
              {Setting.getLanguageText(product.displayName)}
            </Title>
            {product.detail && (
              <Text type="secondary" style={{display: "block", marginBottom: 12, fontSize: "13px", lineHeight: "1.5"}} ellipsis={{rows: 2}}>
                {Setting.getLanguageText(product.detail)}
              </Text>
            )}
            {product.tag && (
              <div style={{marginBottom: 12}}>
                <Tag color="blue" style={{display: "inline-block", width: "fit-content", fontSize: "12px"}}>{product.tag}</Tag>
              </div>
            )}
            <div style={{marginTop: "auto", paddingTop: 8}}>
              {product.isRecharge ? (
                <>
                  {product.rechargeOptions && product.rechargeOptions.length > 0 && (
                    <div style={{marginBottom: 8}}>
                      <Text type="secondary" style={{fontSize: "13px", display: "block", marginBottom: 4}}>
                        {i18next.t("product:Recharge options")}:
                      </Text>
                      <div style={{display: "flex", flexWrap: "wrap", gap: "4px", alignItems: "center"}}>
                        {product.rechargeOptions.map((amount, index) => (
                          <Tag key={amount} color="blue" style={{fontSize: "14px", fontWeight: 600, margin: 0}}>
                            {Setting.getCurrencySymbol(product.currency)}{amount}
                          </Tag>
                        ))}
                        <Text type="secondary" style={{fontSize: "13px", marginLeft: 8}}>
                          {Setting.getCurrencyWithFlag(product.currency)}
                        </Text>
                      </div>
                    </div>
                  )}
                  {product.disableCustomRecharge !== true && (
                    <div style={{marginBottom: 8}}>
                      <Text strong style={{fontSize: "16px", color: "#1890ff"}}>
                        {i18next.t("product:Custom amount available")}
                      </Text>
                      {(!product.rechargeOptions || product.rechargeOptions.length === 0) && (
                        <Text type="secondary" style={{fontSize: "13px", marginLeft: 8}}>
                          {Setting.getCurrencyWithFlag(product.currency)}
                        </Text>
                      )}
                    </div>
                  )}
                  {(!product.rechargeOptions || product.rechargeOptions.length === 0) && product.disableCustomRecharge === true && (
                    <div style={{marginBottom: 8}}>
                      <Text type="secondary" style={{fontSize: "13px", display: "block", marginBottom: 4}}>
                        {i18next.t("product:No recharge options available")}
                      </Text>
                      <Text type="secondary" style={{fontSize: "13px"}}>
                        {Setting.getCurrencyWithFlag(product.currency)}
                      </Text>
                    </div>
                  )}
                </>
              ) : (
                <>
                  <div style={{marginBottom: 8}}>
                    <Text strong style={{fontSize: "28px", color: "#ff4d4f", fontWeight: 600}}>
                      {Setting.getCurrencySymbol(product.currency)}{product.price}
                    </Text>
                    <Text type="secondary" style={{fontSize: "13px", marginLeft: 8}}>
                      {Setting.getCurrencyWithFlag(product.currency)}
                    </Text>
                  </div>
                  <div>
                    <Text type="secondary" style={{fontSize: "13px"}}>
                      {i18next.t("product:Sold")}: {product.sold}
                    </Text>
                  </div>
                </>
              )}
            </div>
          </div>
        </Card>
      </Col>
    );
  }

  render() {
    return (
      <div>
        <FloatingCartButton
          itemCount={this.state.cartItemCount}
          onClick={() => this.props.history.push("/cart")}
        />
        <Row gutter={[16, 16]}>
          {this.state.loading ? (
            <Col span={24}>
              <Card loading={true} />
            </Col>
          ) : this.state.products.length === 0 ? (
            <Col span={24}>
              <Card>
                <Text type="secondary">{i18next.t("general:No products available")}</Text>
              </Card>
            </Col>
          ) : (
            this.state.products.map(product => this.renderProductCard(product))
          )}
        </Row>
      </div>
    );
  }
}

export default ProductStorePage;
