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
import i18next from "i18next";

const {Text, Title} = Typography;

class ProductStorePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      products: [],
      loading: true,
    };
  }

  componentDidMount() {
    this.getProducts();
  }

  getProducts() {
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

  handleBuyProduct(product) {
    this.props.history.push(`/products/${product.owner}/${product.name}/buy`);
  }

  renderProductCard(product) {
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
            <Button
              key="buy"
              type="primary"
              onClick={(e) => {
                e.stopPropagation();
                this.handleBuyProduct(product);
              }}
            >
              {i18next.t("product:Buy")}
            </Button>,
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
                  {product.rechargeOptions && product.rechargeOptions.length > 0 ? (
                    <div style={{marginBottom: 8}}>
                      <Text type="secondary" style={{fontSize: "13px", display: "block", marginBottom: 4}}>
                        {i18next.t("product:Recharge options")}:
                      </Text>
                      <div style={{display: "flex", flexWrap: "wrap", gap: "4px"}}>
                        {product.rechargeOptions.slice(0, 3).map((amount, index) => (
                          <Tag key={index} color="blue" style={{fontSize: "14px", fontWeight: 600, margin: 0}}>
                            {Setting.getCurrencySymbol(product.currency)}{amount}
                          </Tag>
                        ))}
                        {product.rechargeOptions.length > 3 && (
                          <Tag color="blue" style={{fontSize: "14px", fontWeight: 600, margin: 0}}>
                            +{product.rechargeOptions.length - 3}
                          </Tag>
                        )}
                      </div>
                    </div>
                  ) : null}
                  {!product.disableCustomRecharge && (
                    <div style={{marginBottom: 8}}>
                      <Text strong style={{fontSize: "16px", color: "#1890ff"}}>
                        {i18next.t("product:Custom amount available")}
                      </Text>
                    </div>
                  )}
                  <div>
                    <Text type="secondary" style={{fontSize: "13px"}}>
                      {Setting.getCurrencyWithFlag(product.currency)}
                    </Text>
                  </div>
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
