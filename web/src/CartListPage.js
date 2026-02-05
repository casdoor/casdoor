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
import {Link} from "react-router-dom";
import {Button, Table} from "antd";
import * as Setting from "./Setting";
import * as UserBackend from "./backend/UserBackend";
import * as OrderBackend from "./backend/OrderBackend";
import * as ProductBackend from "./backend/ProductBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";
import {QuantityStepper} from "./common/product/CartControls";

class CartListPage extends BaseListPage {
  constructor(props) {
    super(props);
    this.state = {
      ...this.state,
      data: [],
      user: null,
      updatingCartItems: {},
      isPlacingOrder: false,
      loading: false,
      pagination: {
        current: 1,
        pageSize: 10,
        total: 0,
      },
      searchText: "",
      searchedColumn: "",
    };

    this.updatingCartItemsRef = {};
  }

  clearCart() {
    const user = Setting.deepCopy(this.state.user);
    if (user === undefined || user === null) {
      Setting.showMessage("error", i18next.t("general:Failed to delete"));
      return;
    }

    user.cart = [];
    UserBackend.updateUser(user.owner, user.name, user)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.fetch();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  placeOrder() {
    if (this.state.isPlacingOrder) {
      return;
    }

    const owner = this.state.user?.owner || this.props.account.owner;
    const carts = this.state.data || [];
    if (carts.length === 0) {
      Setting.showMessage("error", i18next.t("product:Product list cannot be empty"));
      return;
    }

    this.setState({isPlacingOrder: true});

    const productInfos = carts.map(item => ({
      name: item.name,
      price: item.price,
      quantity: item.quantity,
      pricingName: item.pricingName,
      planName: item.planName,
    }));

    OrderBackend.placeOrder(owner, productInfos, this.state.user?.name)
      .then((res) => {
        if (res.status === "ok") {
          const order = res.data;
          const user = Setting.deepCopy(this.state.user);
          user.cart = [];
          UserBackend.updateUser(user.owner, user.name, user);
          Setting.showMessage("success", i18next.t("product:Order created successfully"));
          Setting.goToLink(`/orders/${order.owner}/${order.name}/pay`);
        } else {
          Setting.showMessage("error", `${i18next.t("product:Failed to create order")}: ${res.msg}`);
          this.setState({isPlacingOrder: false});
        }
      })
      .catch((error) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        this.setState({isPlacingOrder: false});
      });
  }

  deleteCart(record) {
    const user = Setting.deepCopy(this.state.user);
    if (user === undefined || user === null || !Array.isArray(user.cart)) {
      Setting.showMessage("error", i18next.t("general:Failed to delete"));
      return;
    }

    const index = user.cart.findIndex(item => item.name === record.name && item.price === record.price && (item.pricingName || "") === (record.pricingName || "") && (item.planName || "") === (record.planName || ""));
    if (index === -1) {
      Setting.showMessage("error", i18next.t("general:Failed to delete"));
      return;
    }

    user.cart.splice(index, 1);

    UserBackend.updateUser(user.owner, user.name, user)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.fetch();
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  updateCartItemQuantity(record, newQuantity) {
    if (newQuantity < 1) {
      return;
    }

    const itemKey = `${record.name}-${record.price}-${record.pricingName || ""}-${record.planName || ""}`;
    if (this.updatingCartItemsRef?.[itemKey]) {
      return;
    }

    this.updatingCartItemsRef[itemKey] = true;

    const user = Setting.deepCopy(this.state.user);
    const index = user.cart.findIndex(item => item.name === record.name && item.price === record.price && (item.pricingName || "") === (record.pricingName || "") && (item.planName || "") === (record.planName || ""));
    if (index === -1) {
      delete this.updatingCartItemsRef[itemKey];
      return;
    }

    if (index !== -1) {
      user.cart[index].quantity = newQuantity;

      const newData = [...this.state.data];
      const dataIndex = newData.findIndex(item => item.name === record.name && item.price === record.price && (item.pricingName || "") === (record.pricingName || "") && (item.planName || "") === (record.planName || ""));
      if (dataIndex !== -1) {
        newData[dataIndex].quantity = newQuantity;
        this.setState({data: newData});
      }

      this.setState(prevState => ({
        updatingCartItems: {
          ...(prevState.updatingCartItems || {}),
          [itemKey]: true,
        },
      }));

      UserBackend.updateUser(user.owner, user.name, user)
        .then((res) => {
          if (res.status === "ok") {
            this.setState({user: user});
          } else {
            Setting.showMessage("error", res.msg);
            this.fetch();
          }
        })
        .catch(error => {
          Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
          this.fetch();
        })
        .finally(() => {
          delete this.updatingCartItemsRef[itemKey];
          this.setState(prevState => {
            const updatingCartItems = {...(prevState.updatingCartItems || {})};
            delete updatingCartItems[itemKey];
            return {updatingCartItems};
          });
        });
    }
  }

  renderTable(carts) {
    const isEmpty = carts === undefined || carts === null || carts.length === 0;
    const owner = this.state.user?.owner || this.props.account.owner;

    let total = 0;
    let currency = "";
    if (carts && carts.length > 0) {
      carts.forEach(item => {
        total += item.price * item.quantity;
      });
      currency = carts[0].currency;
    }

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
            <Link to={`/products/${owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "170px",
        sorter: true,
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
        },
      },
      {
        title: i18next.t("product:Price"),
        dataIndex: "price",
        key: "price",
        width: "160px",
        sorter: true,
        render: (text, record) => {
          const subtotal = (record.price * record.quantity).toFixed(2);
          return Setting.getPriceDisplay(subtotal, record.currency);
        },
      },
      {
        title: i18next.t("pricing:Pricing name"),
        dataIndex: "pricingName",
        key: "pricingName",
        width: "140px",
        sorter: true,
        render: (text, record) => {
          if (!text) {return null;}
          return (
            <Link to={`/pricings/${owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("plan:Plan name"),
        dataIndex: "planName",
        key: "planName",
        width: "140px",
        sorter: true,
        render: (text, record) => {
          if (!text) {return null;}
          return (
            <Link to={`/plans/${owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("product:Quantity"),
        dataIndex: "quantity",
        key: "quantity",
        width: "100px",
        sorter: true,
        render: (text, record) => {
          const itemKey = `${record.name}-${record.price}-${record.pricingName || ""}-${record.planName || ""}`;
          const isUpdating = this.state.updatingCartItems?.[itemKey] === true;
          return (
            <QuantityStepper
              value={text}
              min={1}
              onIncrease={() => this.updateCartItemQuantity(record, text + 1)}
              onDecrease={() => this.updateCartItemQuantity(record, text - 1)}
              onChange={null}
              disabled={isUpdating}
            />
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "160px",
        fixed: Setting.isMobile() ? false : "right",
        render: (text, record, index) => {
          return (
            <div style={{display: "flex", flexWrap: "wrap", gap: "8px"}}>
              <Button type="primary" onClick={() => this.props.history.push(`/products/${owner}/${record.name}/buy`)}>
                {i18next.t("product:Detail")}
              </Button>
              <PopconfirmModal
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteCart(record)}
              >
              </PopconfirmModal>
            </div>
          );
        },
      },
    ];

    return (
      <div>
        <Table
          scroll={{x: "max-content"}}
          columns={columns}
          dataSource={carts}
          rowKey={(record, index) => `${record.name}-${record.pricingName}-${record.planName}-${index}`}
          size="middle"
          bordered
          pagination={false}
          title={() => {
            return (
              <div>
                {i18next.t("general:Cart")}&nbsp;&nbsp;&nbsp;&nbsp;
                <Button size="small" onClick={() => this.props.history.push("/product-store")}>{i18next.t("general:Add")}</Button>
                &nbsp;&nbsp;
                <PopconfirmModal
                  size="small"
                  style={{marginRight: "8px"}}
                  text={i18next.t("general:Clear")}
                  title={i18next.t("general:Sure to delete") + `: ${i18next.t("general:Cart")} ?`}
                  onConfirm={() => this.clearCart()}
                  disabled={isEmpty}
                />
                <Button type="primary" size="small" onClick={() => this.placeOrder()} disabled={isEmpty || this.state.isPlacingOrder} loading={this.state.isPlacingOrder}>{i18next.t("general:Place Order")}</Button>
              </div>
            );
          }}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />

        {!isEmpty && (
          <div style={{marginTop: "20px", display: "flex", flexDirection: "column", justifyContent: "center", alignItems: "center", gap: "20px"}}>
            <div style={{display: "flex", alignItems: "center", fontSize: "18px", fontWeight: "bold"}}>
              {i18next.t("product:Total Price")}:&nbsp;
              <span style={{color: "red", fontSize: "28px"}}>
                {Setting.getCurrencySymbol(currency)}{total.toFixed(2)} ({Setting.getCurrencyText(currency)})
              </span>
            </div>
            <Button
              type="primary"
              size="large"
              style={{height: "50px", fontSize: "20px", padding: "0 40px", borderRadius: "5px"}}
              onClick={() => this.placeOrder()}
              disabled={this.state.isPlacingOrder}
              loading={this.state.isPlacingOrder}
            >
              {i18next.t("general:Place Order")}
            </Button>
          </div>
        )}
      </div>
    );
  }

  fetch = (params = {}) => {
    this.setState({loading: true});
    const organizationName = this.props.account.owner;
    const userName = this.props.account.name;

    UserBackend.getUser(organizationName, userName)
      .then(async(res) => {
        if (res.status === "ok") {
          const cartData = res.data.cart || [];

          const productPromises = cartData.map(item =>
            ProductBackend.getProduct(organizationName, item.name)
              .then(pRes => {
                if (pRes.status === "ok" && pRes.data) {
                  return {
                    ...pRes.data,
                    pricingName: item.pricingName,
                    planName: item.planName,
                    quantity: item.quantity,
                    price: pRes.data.isRecharge ? item.price : pRes.data.price,
                  };
                }
                return item;
              })
              .catch(() => item)
          );

          const fullCartData = await Promise.all(productPromises);

          const sortedData = [...fullCartData];
          if (params.sortField && params.sortOrder) {
            sortedData.sort((a, b) => {
              const aValue = a[params.sortField];
              const bValue = b[params.sortField];

              if (aValue === bValue) {
                return 0;
              }

              const comparison = aValue > bValue ? 1 : -1;
              return params.sortOrder === "ascend" ? comparison : -comparison;
            });
          }

          this.setState({
            loading: false,
            data: sortedData,
            user: res.data,
            pagination: {
              ...params.pagination,
              total: sortedData.length,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
          this.setState({loading: false});
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => {
        this.setState({
          loading: false,
        });
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };
}

export default CartListPage;
