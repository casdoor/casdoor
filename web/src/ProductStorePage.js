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

const {Meta} = Card;
const {Text, Title} = Typography;

class ProductStorePage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      products: [],
      loading: true,
    };
  }

  UNSAFE_componentWillMount() {
    this.getProducts();
  }

  getProducts() {
    this.setState({loading: true});
    ProductBackend.getProducts("", 1, 100, "state", "Published", "", "")
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
          cover={
            <img
              alt={product.displayName}
              src={product.image}
              style={{height: "200px", objectFit: "cover"}}
            />
          }
          actions={[
            <Button
              key="buy"
              type="primary"
              onClick={() => this.handleBuyProduct(product)}
              style={{width: "90%"}}
            >
              {i18next.t("product:Buy")}
            </Button>,
          ]}
        >
          <Meta
            title={<Title level={5} style={{marginBottom: 8}}>{Setting.getLanguageText(product.displayName)}</Title>}
            description={
              <div>
                <Text style={{display: "block", marginBottom: 8}} ellipsis={{rows: 2}}>
                  {Setting.getLanguageText(product.detail)}
                </Text>
                {product.tag && (
                  <Tag color="blue" style={{marginBottom: 8}}>{product.tag}</Tag>
                )}
                <div style={{marginBottom: 8}}>
                  <Text strong style={{fontSize: "18px", color: "#ff4d4f"}}>
                    {Setting.getCurrencySymbol(product.currency)}{product.price}
                  </Text>
                  <Text type="secondary" style={{marginLeft: 8}}>
                    {Setting.getCurrencyText(product)}
                  </Text>
                </div>
                <div>
                  <Text type="secondary" style={{fontSize: "12px"}}>
                    {i18next.t("product:Quantity")}: {product.quantity} | {i18next.t("product:Sold")}: {product.sold}
                  </Text>
                </div>
              </div>
            }
          />
        </Card>
      </Col>
    );
  }

  render() {
    return (
      <div>
        <Title level={2} style={{marginBottom: 24}}>
          {i18next.t("general:Product Store")}
        </Title>
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
