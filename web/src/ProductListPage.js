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
import {Link} from "react-router-dom";
import {Button, Col, List, Popconfirm, Row, Table, Tooltip} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as ProductBackend from "./backend/ProductBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import {EditOutlined} from "@ant-design/icons";

class ProductListPage extends BaseListPage {
  newProduct() {
    const randomName = Setting.getRandomName();
    return {
      owner: "admin",
      name: `product_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Product - ${randomName}`,
      image: "https://cdn.casdoor.com/logo/casdoor-logo_1185x256.png",
      tag: "Casdoor Summit 2022",
      currency: "USD",
      price: 300,
      quantity: 99,
      sold: 10,
      providers: [],
      state: "Published",
    };
  }

  addProduct() {
    const newProduct = this.newProduct();
    ProductBackend.addProduct(newProduct)
      .then((res) => {
        this.props.history.push({pathname: `/products/${newProduct.name}`, mode: "add"});
      }
      )
      .catch(error => {
        Setting.showMessage("error", `Product failed to add: ${error}`);
      });
  }

  deleteProduct(i) {
    ProductBackend.deleteProduct(this.state.data[i])
      .then((res) => {
        Setting.showMessage("success", "Product deleted successfully");
        this.setState({
          data: Setting.deleteRow(this.state.data, i),
          pagination: {total: this.state.pagination.total - 1},
        });
      }
      )
      .catch(error => {
        Setting.showMessage("error", `Product failed to delete: ${error}`);
      });
  }

  renderTable(products) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "140px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/products/${text}`}>
              {text}
            </Link>
          );
        }
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "160px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "170px",
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("product:Image"),
        dataIndex: "image",
        key: "image",
        width: "170px",
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              <img src={text} alt={text} width={150} />
            </a>
          );
        }
      },
      {
        title: i18next.t("product:Tag"),
        dataIndex: "tag",
        key: "tag",
        width: "160px",
        sorter: true,
        ...this.getColumnSearchProps("tag"),
      },
      {
        title: i18next.t("product:Currency"),
        dataIndex: "currency",
        key: "currency",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("currency"),
      },
      {
        title: i18next.t("product:Price"),
        dataIndex: "price",
        key: "price",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("price"),
      },
      {
        title: i18next.t("product:Quantity"),
        dataIndex: "quantity",
        key: "quantity",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("quantity"),
      },
      {
        title: i18next.t("product:Sold"),
        dataIndex: "sold",
        key: "sold",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("sold"),
      },
      {
        title: i18next.t("general:State"),
        dataIndex: "state",
        key: "state",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("state"),
      },
      {
        title: i18next.t("product:Payment providers"),
        dataIndex: "providers",
        key: "providers",
        width: "500px",
        ...this.getColumnSearchProps("providers"),
        render: (text, record, index) => {
          const providers = text;
          if (providers.length === 0) {
            return "(empty)";
          }

          const half = Math.floor((providers.length + 1) / 2);

          const getList = (providers) => {
            return (
              <List
                size="small"
                locale={{emptyText: " "}}
                dataSource={providers}
                renderItem={(providerName, i) => {
                  return (
                    <List.Item>
                      <div style={{display: "inline"}}>
                        <Tooltip placement="topLeft" title="Edit">
                          <Button style={{marginRight: "5px"}} icon={<EditOutlined />} size="small" onClick={() => Setting.goToLinkSoft(this, `/providers/${providerName}`)} />
                        </Tooltip>
                        <Link to={`/providers/${providerName}`}>
                          {providerName}
                        </Link>
                      </div>
                    </List.Item>
                  );
                }}
              />
            );
          };

          return (
            <div>
              <Row>
                <Col span={12}>
                  {
                    getList(providers.slice(0, half))
                  }
                </Col>
                <Col span={12}>
                  {
                    getList(providers.slice(half))
                  }
                </Col>
              </Row>
            </div>
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "230px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} onClick={() => this.props.history.push(`/products/${record.name}/buy`)}>{i18next.t("product:Buy")}</Button>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/products/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete product: ${record.name} ?`}
                onConfirm={() => this.deleteProduct(index)}
              >
                <Button style={{marginBottom: "10px"}} type="danger">{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
          );
        }
      },
    ];

    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <div>
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={products} rowKey="name" size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Products")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addProduct.bind(this)}>{i18next.t("general:Add")}</Button>
            </div>
          )}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    let field = params.searchedColumn, value = params.searchText;
    let sortField = params.sortField, sortOrder = params.sortOrder;
    if (params.type !== undefined && params.type !== null) {
      field = "type";
      value = params.type;
    }
    this.setState({ loading: true });
    ProductBackend.getProducts("", params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            loading: false,
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        }
      });
  };
}

export default ProductListPage;
