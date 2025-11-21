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
import {Link} from "react-router-dom";
import * as Setting from "./Setting";
import {Button, Table} from "antd";
import PopconfirmModal from "./common/modal/PopconfirmModal";
import RechargeModal from "./common/modal/RechargeModal";
import React from "react";
import * as TransactionBackend from "./backend/TransactionBackend";
import moment from "moment/moment";

class TransactionListPage extends BaseListPage {
  constructor(props) {
    super(props);
    this.state = {
      ...this.state,
      rechargeModalVisible: false,
    };
  }
  newTransaction() {
    const organizationName = Setting.getRequestOrganization(this.props.account);
    return {
      owner: organizationName,
      createdTime: moment().format(),
      application: "app-built-in",
      domain: "https://ai-admin.casibase.com",
      category: "chat_id",
      type: "message_id",
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

  showRechargeModal = () => {
    this.setState({rechargeModalVisible: true});
  };

  handleRechargeOk = (rechargeData) => {
    const organizationName = rechargeData.organization || Setting.getRequestOrganization(this.props.account);
    // Create a recharge transaction with minimal required fields
    const newTransaction = {
      owner: organizationName,
      createdTime: moment().format(),
      application: rechargeData.application || "",
      domain: "",
      category: "Recharge",
      type: "Manual",
      provider: "",
      user: "",
      tag: rechargeData.tag,
      amount: rechargeData.amount,
      currency: rechargeData.currency,
      payment: "",
      state: "Paid", // Recharge transactions are considered completed immediately
    };

    TransactionBackend.addTransaction(newTransaction)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully added"));
          this.setState({rechargeModalVisible: false});
          this.fetch({
            pagination: this.state.pagination,
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  handleRechargeCancel = () => {
    this.setState({rechargeModalVisible: false});
  };

  renderTable(transactions) {
    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "120px",
        fixed: "left",
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
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "180px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/transactions/${record.owner}/${record.name}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "160px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${record.owner}/${record.application}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("provider:Domain"),
        dataIndex: "domain",
        key: "domain",
        width: "200px",
        sorter: true,
        ...this.getColumnSearchProps("domain"),
        render: (text, record, index) => {
          if (!text) {
            return null;
          }

          return (
            <a href={text} target="_blank" rel="noopener noreferrer">
              {text}
            </a>
          );
        },
      },
      {
        title: i18next.t("provider:Category"),
        dataIndex: "category",
        key: "category",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("category"),
        render: (text, record, index) => {
          if (text && record.domain) {
            const chatUrl = `${record.domain}/chats/${text}`;
            return (
              <a href={chatUrl} target="_blank" rel="noopener noreferrer">
                {text}
              </a>
            );
          }
          return text;
        },
      },
      {
        title: i18next.t("provider:Type"),
        dataIndex: "type",
        key: "type",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("type"),
        render: (text, record, index) => {
          if (text && record.domain) {
            const messageUrl = `${record.domain}/messages/${text}`;
            return (
              <a href={messageUrl} target="_blank" rel="noopener noreferrer">
                {text}
              </a>
            );
          }
          return text;
        },
      },
      {
        title: i18next.t("general:Provider"),
        dataIndex: "provider",
        key: "provider",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("provider"),
        render: (text, record, index) => {
          if (!text) {
            return text;
          }
          if (record.domain) {
            const casibaseUrl = `${record.domain}/providers/${text}`;
            return (
              <a href={casibaseUrl} target="_blank" rel="noopener noreferrer">
                {text}
              </a>
            );
          }
          return (
            <Link to={`/providers/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:User"),
        dataIndex: "user",
        key: "user",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("user"),
        render: (text, record, index) => {
          if (!text || Setting.isAnonymousUserName(text)) {
            return text;
          }

          return (
            <Link to={`/users/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("user:Tag"),
        dataIndex: "tag",
        key: "tag",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("tag"),
      },
      {
        title: i18next.t("transaction:Amount"),
        dataIndex: "amount",
        key: "amount",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("amount"),
      },
      {
        title: i18next.t("payment:Currency"),
        dataIndex: "currency",
        key: "currency",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("currency"),
        render: (text, record, index) => {
          return Setting.getCurrencyWithFlag(text);
        },
      },
      {
        title: i18next.t("general:Payment"),
        dataIndex: "payment",
        key: "payment",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("payment"),
        render: (text, record, index) => {
          return (
            <Link to={`/payments/${record.owner}/${record.payment}`}>
              {text}
            </Link>
          );
        },
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
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "240px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          const isAdmin = Setting.isLocalAdminUser(this.props.account);
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push({pathname: `/transactions/${record.owner}/${record.name}`, mode: isAdmin ? "edit" : "view"})}>{isAdmin ? i18next.t("general:Edit") : i18next.t("general:View")}</Button>
              <PopconfirmModal
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteTransaction(index)}
                disabled={!isAdmin}
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={transactions} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => {
            const isAdmin = Setting.isLocalAdminUser(this.props.account);
            return (
              <div>
                {i18next.t("general:Transactions")}&nbsp;&nbsp;&nbsp;&nbsp;
                <Button type="primary" size="small" disabled={!isAdmin} onClick={this.showRechargeModal}>{i18next.t("transaction:Recharge")}</Button>
                &nbsp;&nbsp;
                <Button type="primary" size="small" disabled={!isAdmin} onClick={this.addTransaction.bind(this)}>{i18next.t("general:Add")}</Button>
              </div>
            );
          }}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
        <RechargeModal
          visible={this.state.rechargeModalVisible}
          onOk={this.handleRechargeOk}
          onCancel={this.handleRechargeCancel}
          account={this.props.account}
          currentOrganization={Setting.getRequestOrganization(this.props.account)}
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
