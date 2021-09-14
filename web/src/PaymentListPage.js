// Copyright 2021 The casbin Authors. All Rights Reserved.
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

class PaymentListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      payments: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getPayments();
  }

  getPayments() {
    PaymentBackend.getPayments().then(res => {
      this.setState({
        payments: res
      })
    })
  }

  deletePayment(i) {
    PaymentBackend.deletePayment(this.state.payments[i]).then(res => {
      Setting.showMessage("success", `Payment deleted successfully`);
      this.setState({
        payments: Setting.deleteRow(this.state.payments, i),
      });
    }).catch(error => {
      Setting.showMessage("error", `Payment failed to delete: ${error}`);
    });
  }

  renderTable(payments) {
    const columns = [
      {
        title: "Id",
        dataIndex: 'id',
        key: 'id',
        width: (Setting.isMobile()) ? "100px" : "170px",
        fixed: 'left',
        sorter: (a, b) => a.id.localeCompare(b.id),
      },
      {
        title: "invoice",
        dataIndex: 'invoice',
        key: 'invoice',
        width: '150px',
        sorter: (a, b) => a.invoice.localeCompare(b.invoice),
      },
      {
        title: "Application",
        dataIndex: 'application',
        key: 'application',
        width: '120px',
        sorter: (a, b) => a.application.localeCompare(b.application),
        render: (text, record, index) => {
          return (
            <Link to={`/applications/${text.substring(text.indexOf("/")+1)}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: "Amount",
        dataIndex: 'pay_item',
        key: 'pay_item',
        width: '150px',
        ellipsis: true,
        render: (text, record, index) => {
          return (
              `amount: ${text.currency} ${text.price}`
          )
        }
      },
      {
        title: "Status",
        dataIndex: 'status',
        key: 'status',
        width: '120px',
        sorter: (a, b) => a.status.localeCompare(b.status),
      },
      {
        title: "Payer",
        dataIndex: 'payer',
        key: 'payer',
        width: '120px',
        render: (text, record, index) => {
          return (
              text === null ? "" : (
                  `${text?.name.surname} ${text?.name.given_name}`
              )
          )
        }
      },
      {
        title: "Payer Email",
        dataIndex: 'payer',
        key: 'payer',
        width: '160px',
        ellipsis: true,
        render: (text, record, index) => {
          return (
              text === null ? "" : text.email_address
          )
        }
      },
      {
        title: "Create_time",
        dataIndex: 'create_time',
        key: 'create_time',
        width: '150px',
        sorter: (a, b) => a.create_time.localeCompare(b.create_time),
      },
      {
        title: "update_time",
        dataIndex: 'update_time',
        key: 'update_time',
        width: '150px',
        sorter: (a, b) => a.update_time.localeCompare(b.update_time),
      },
      {
        title: "description",
        dataIndex: 'pay_item',
        key: 'pay_item',
        ellipsis: true,
        width: '150px',
        render: (text, record, index) => {
          return (
              text.description
          )
        }
      },
      {
        title: "Callback",
        dataIndex: 'callback',
        key: 'callback',
        width: '100px',
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: '',
        key: 'op',
        width: '80px',
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Popconfirm
                title={`Sure to delete this payment ?`}
                onConfirm={() => this.deletePayment(index)}
              >
                <Button style={{marginBottom: '10px'}} type="danger">{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
          )
        }
      },
    ];

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={payments} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   {"Payments"}&nbsp;&nbsp;&nbsp;&nbsp;
                 </div>
               )}
               loading={payments === null}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.state.payments)
        }
      </div>
    );
  }
}

export default PaymentListPage;
