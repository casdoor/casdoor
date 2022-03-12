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
import {Button, Popconfirm, Table} from 'antd';
import moment from "moment";
import * as Setting from "./Setting";
import * as PaymentBackend from "./backend/PaymentBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import * as Provider from "./auth/Provider";

class PaymentListPage extends BaseListPage {
  newPayment() {
    const randomName = Setting.getRandomName();
    return {
      owner: "admin",
      name: `payment_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Payment - ${randomName}`,
      provider: "provider_pay_paypal",
      type: "PayPal",
      organization: "built-in",
      user: "admin",
      productId: "computer-1",
      productName: "A notebook computer",
      price: 300.00,
      currency: "USD",
      state: "Paid",
    }
  }

  addPayment() {
    const newPayment = this.newPayment();
    PaymentBackend.addPayment(newPayment)
      .then((res) => {
          this.props.history.push({pathname: `/payments/${newPayment.name}`, mode: "add"});
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Payment failed to add: ${error}`);
      });
  }

  deletePayment(i) {
    PaymentBackend.deletePayment(this.state.data[i])
      .then((res) => {
          Setting.showMessage("success", `Payment deleted successfully`);
          this.setState({
            data: Setting.deleteRow(this.state.data, i),
            pagination: {total: this.state.pagination.total - 1},
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Payment failed to delete: ${error}`);
      });
  }

  renderTable(payments) {
    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: 'organization',
        key: 'organization',
        width: '120px',
        sorter: true,
        ...this.getColumnSearchProps('organization'),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:User"),
        dataIndex: 'user',
        key: 'user',
        width: '120px',
        sorter: true,
        ...this.getColumnSearchProps('user'),
        render: (text, record, index) => {
          return (
            <Link to={`/users/${record.organization}/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '180px',
        fixed: 'left',
        sorter: true,
        ...this.getColumnSearchProps('name'),
        render: (text, record, index) => {
          return (
            <Link to={`/payments/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: 'createdTime',
        key: 'createdTime',
        width: '160px',
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      // {
      //   title: i18next.t("general:Display name"),
      //   dataIndex: 'displayName',
      //   key: 'displayName',
      //   width: '160px',
      //   sorter: true,
      //   ...this.getColumnSearchProps('displayName'),
      // },
      {
        title: i18next.t("general:Provider"),
        dataIndex: 'provider',
        key: 'provider',
        width: '150px',
        fixed: 'left',
        sorter: true,
        ...this.getColumnSearchProps('provider'),
        render: (text, record, index) => {
          return (
            <Link to={`/providers/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("payment:Type"),
        dataIndex: 'type',
        key: 'type',
        width: '140px',
        align: 'center',
        filterMultiple: false,
        filters: Setting.getProviderTypeOptions('Payment').map((o) => {return {text:o.id, value:o.name}}),
        sorter: true,
        render: (text, record, index) => {
          record.category = "Payment";
          return Provider.getProviderLogoWidget(record);
        }
      },
      {
        title: i18next.t("payment:Product"),
        dataIndex: 'productName',
        key: 'productName',
        // width: '160px',
        sorter: true,
        ...this.getColumnSearchProps('productName'),
      },
      {
        title: i18next.t("payment:Price"),
        dataIndex: 'price',
        key: 'price',
        width: '120px',
        sorter: true,
        ...this.getColumnSearchProps('price'),
      },
      {
        title: i18next.t("payment:Currency"),
        dataIndex: 'currency',
        key: 'currency',
        width: '120px',
        sorter: true,
        ...this.getColumnSearchProps('currency'),
      },
      {
        title: i18next.t("payment:State"),
        dataIndex: 'state',
        key: 'state',
        width: '120px',
        sorter: true,
        ...this.getColumnSearchProps('state'),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: '',
        key: 'op',
        width: '170px',
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/payments/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete payment: ${record.name} ?`}
                onConfirm={() => this.deletePayment(index)}
              >
                <Button style={{marginBottom: '10px'}} type="danger">{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
          )
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
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={payments} rowKey="name" size="middle" bordered pagination={paginationProps}
               title={() => (
                 <div>
                   {i18next.t("general:Payments")}&nbsp;&nbsp;&nbsp;&nbsp;
                   <Button type="primary" size="small" onClick={this.addPayment.bind(this)}>{i18next.t("general:Add")}</Button>
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
    PaymentBackend.getPayments("", params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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

export default PaymentListPage;
