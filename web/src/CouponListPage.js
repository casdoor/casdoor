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
import {Link} from "react-router-dom";
import {Button, Table, Tag} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as CouponBackend from "./backend/CouponBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class CouponListPage extends BaseListPage {
  newCoupon() {
    const randomName = Setting.getRandomName();
    const owner = Setting.getRequestOrganization(this.props.account);
    return {
      owner: owner,
      name: `coupon_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Coupon - ${randomName}`,
      description: "",
      code: `CODE_${randomName}`.toUpperCase(),
      discountType: "percentage",
      discount: 10,
      maxDiscount: 0,
      scope: "universal",
      products: [],
      users: [],
      quantity: 100,
      usedCount: 0,
      maxUsagePerUser: 1,
      startTime: moment().format(),
      expireTime: moment().add(30, "days").format(),
      minOrderAmount: 0,
      currency: "USD",
      state: "Active",
    };
  }

  addCoupon() {
    const newCoupon = this.newCoupon();
    CouponBackend.addCoupon(newCoupon)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/coupons/${newCoupon.owner}/${newCoupon.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteCoupon(i) {
    CouponBackend.deleteCoupon(this.state.data[i])
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.fetch({
            pagination: {
              ...this.state.pagination,
              current: this.state.pagination.current > 1 && this.state.data.length === 1 ? this.state.pagination.current - 1 : this.state.pagination.current,
            },
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderTable(coupons) {
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
            <Link to={`/coupons/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("owner"),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
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
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("invitation:Code"),
        dataIndex: "code",
        key: "code",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("code"),
      },
      {
        title: i18next.t("coupon:Discount type"),
        dataIndex: "discountType",
        key: "discountType",
        width: "130px",
        sorter: true,
        render: (text, record, index) => {
          return text === "percentage" ? i18next.t("coupon:Percentage") : i18next.t("coupon:Fixed");
        },
      },
      {
        title: i18next.t("coupon:Discount"),
        dataIndex: "discount",
        key: "discount",
        width: "120px",
        sorter: true,
        render: (text, record, index) => {
          if (record.discountType === "percentage") {
            return `${text}%`;
          }
          return Setting.getPriceDisplay(text, record.currency);
        },
      },
      {
        title: i18next.t("provider:Scope"),
        dataIndex: "scope",
        key: "scope",
        width: "120px",
        sorter: true,
        render: (text) => {
          const colorMap = {universal: "blue", product: "green", user: "orange"};
          return <Tag color={colorMap[text] || "default"}>{text}</Tag>;
        },
      },
      {
        title: i18next.t("coupon:Usage"),
        dataIndex: "usedCount",
        key: "usedCount",
        width: "120px",
        render: (text, record) => {
          return record.quantity > 0 ? `${text}/${record.quantity}` : `${text}/\u221E`;
        },
      },
      {
        title: i18next.t("general:Expire time"),
        dataIndex: "expireTime",
        key: "expireTime",
        width: "160px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:State"),
        dataIndex: "state",
        key: "state",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("state"),
        render: (text) => {
          const colorMap = {Active: "green", Inactive: "default", Expired: "red"};
          return <Tag color={colorMap[text] || "default"}>{text}</Tag>;
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "200px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          const isAdmin = Setting.isLocalAdminUser(this.props.account);
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push({pathname: `/coupons/${record.owner}/${record.name}`, mode: isAdmin ? "edit" : "view"})}>{isAdmin ? i18next.t("general:Edit") : i18next.t("general:View")}</Button>
              <PopconfirmModal
                disabled={!isAdmin}
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteCoupon(index)}
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={coupons} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => {
            const isAdmin = Setting.isLocalAdminUser(this.props.account);
            return (
              <div>
                {i18next.t("general:Coupons")}&nbsp;&nbsp;&nbsp;&nbsp;
                <Button type="primary" size="small" disabled={!isAdmin} onClick={this.addCoupon.bind(this)}>{i18next.t("general:Add")}</Button>
              </div>
            );
          }}
          loading={this.getTableLoading()}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    let field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (params.type !== undefined && params.type !== null) {
      field = "type";
      value = params.type;
    }
    this.setState({loading: true});
    CouponBackend.getCoupons(Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        this.setState({
          loading: false,
        });
        if (res.status === "ok") {
          this.setState({
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
          if (Setting.isResponseDenied(res)) {
            this.setState({
              isAuthorized: false,
            });
          } else {
            Setting.showMessage("error", res.msg);
          }
        }
      });
  };
}

export default CouponListPage;
