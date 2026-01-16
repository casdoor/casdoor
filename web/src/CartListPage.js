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
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class CartListPage extends BaseListPage {
  deleteCart(record) {
    const user = Setting.deepCopy(this.state.user);
    if (user === undefined || user === null || !Array.isArray(user.cart)) {
      Setting.showMessage("error", i18next.t("general:Failed to delete"));
      return;
    }

    const index = user.cart.findIndex(item => item.name === record.name && item.price === record.price);
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

  renderTable(carts) {
    const owner = this.state.user?.owner || this.props.account.owner;

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
        title: i18next.t("payment:Currency"),
        dataIndex: "currency",
        key: "currency",
        width: "120px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getCurrencyWithFlag(text);
        },
      },
      {
        title: i18next.t("product:Price"),
        dataIndex: "price",
        key: "price",
        width: "120px",
        sorter: true,
      },
      {
        title: i18next.t("product:Quantity"),
        dataIndex: "quantity",
        key: "quantity",
        width: "120px",
        sorter: true,
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
                {i18next.t("product:Buy")}
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

    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <div>
        <Table
          scroll={{x: "max-content"}}
          columns={columns}
          dataSource={carts}
          rowKey={(record, index) => `${record.name}-${index}`}
          size="middle"
          bordered
          pagination={paginationProps}
          title={() => {
            return (
              <div>
                {i18next.t("general:Carts")}&nbsp;&nbsp;&nbsp;&nbsp;
                <Button type="primary" size="small" onClick={() => this.props.history.push("/product-store")}>{i18next.t("general:Add")}</Button>
                &nbsp;&nbsp;
                <Button size="small">{i18next.t("general:Place Order")}</Button>
              </div>
            );
          }}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    this.setState({loading: true});
    const organizationName = this.props.account.owner;
    const userName = this.props.account.name;

    UserBackend.getUser(organizationName, userName)
      .then((res) => {
        this.setState({
          loading: false,
        });
        if (res.status === "ok") {
          const cartData = res.data.cart || [];
          this.setState({
            data: cartData,
            user: res.data,
            pagination: {
              ...params.pagination,
              total: cartData.length,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
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
