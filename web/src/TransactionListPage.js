// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

import BaseListPage from "./BaseListPage";
import i18next from "i18next";
import * as Setting from "./Setting";
import {Button, Table} from "antd";
import React from "react";
import * as TransactionBackend from "./backend/TransactionBackend";
import moment from "moment/moment";
import {getTransactionTableColumns} from "./table/TransactionTableColumns";

class TransactionListPage extends BaseListPage {
  newTransaction() {
    const organizationName = Setting.getRequestOrganization(this.props.account);
    return {
      owner: organizationName,
      createdTime: moment().format(),
      application: "app-built-in",
      domain: "https://ai-admin.casibase.com",
      category: "",
      type: "chat_id",
      subtype: "message_id",
      provider: "provider_chatgpt",
      user: "admin",
      tag: "AI message",
      amount: 0.1,
      currency: "USD",
      payment: "payment_paypal_001",
      state: "Paid",
    };
  }

  deleteTransaction(i) {
    TransactionBackend.deleteTransaction(this.state.data[i])
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

  addTransaction() {
    const newTransaction = this.newTransaction();
    TransactionBackend.addTransaction(newTransaction)
      .then((res) => {
        if (res.status === "ok") {
          const transactionId = res.data;
          this.props.history.push({pathname: `/transactions/${newTransaction.owner}/${transactionId}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      }
      )
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  rechargeTransaction() {
    const organizationName = Setting.getRequestOrganization(this.props.account);
    const newTransaction = {
      owner: organizationName,
      createdTime: moment().format(),
      application: this.props.account.signupApplication || "",
      domain: "",
      category: "Recharge",
      type: "",
      subtype: "",
      provider: "",
      user: this.props.account.name || "",
      tag: "User",
      amount: 100,
      currency: "USD",
      payment: "",
      state: "Paid",
    };
    TransactionBackend.addTransaction(newTransaction)
      .then((res) => {
        if (res.status === "ok") {
          const transactionId = res.data;
          this.props.history.push({pathname: `/transactions/${newTransaction.owner}/${transactionId}`, mode: "recharge"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderTable(transactions) {
    const columns = getTransactionTableColumns({
      includeOrganization: true,
      includeUser: true,
      includeTag: true,
      includeActions: true,
      getColumnSearchProps: this.getColumnSearchProps,
      account: this.props.account,
      onEdit: (record, isAdmin) => {
        this.props.history.push({pathname: `/transactions/${record.owner}/${record.name}`, mode: isAdmin ? "edit" : "view"});
      },
      onDelete: (index) => {
        this.deleteTransaction(index);
      },
    });

    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <div>
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={transactions} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => {
            const isAdmin = Setting.isLocalAdminUser(this.props.account);
            return (
              <div>
                {i18next.t("general:Transactions")}&nbsp;&nbsp;&nbsp;&nbsp;
                <Button size="small" disabled={!isAdmin} onClick={this.addTransaction.bind(this)}>{i18next.t("general:Add")}</Button>
                &nbsp;&nbsp;
                <Button type="primary" size="small" disabled={!isAdmin} onClick={this.rechargeTransaction.bind(this)}>{i18next.t("transaction:Recharge")}</Button>
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
    let field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (params.type !== undefined && params.type !== null) {
      field = "type";
      value = params.type;
    }
    this.setState({loading: true});
    TransactionBackend.getTransactions(Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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

export default TransactionListPage;
