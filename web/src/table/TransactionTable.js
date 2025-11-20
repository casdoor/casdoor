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

import React from "react";
import {Table} from "antd";
import {Link} from "react-router-dom";
import * as Setting from "../Setting";
import i18next from "i18next";

class TransactionTable extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
    };
  }

  render() {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "180px",
        render: (text, record) => {
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
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "120px",
        render: (text, record) => {
          if (!text) {
            return text;
          }
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
        render: (text) => {
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
        render: (text, record) => {
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
        render: (text, record) => {
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
        render: (text, record) => {
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
      ...(!this.props.hideTag ? [{
        title: i18next.t("user:Tag"),
        dataIndex: "tag",
        key: "tag",
        width: "120px",
      }] : []),
      {
        title: i18next.t("transaction:Amount"),
        dataIndex: "amount",
        key: "amount",
        width: "120px",
      },
      {
        title: i18next.t("payment:Currency"),
        dataIndex: "currency",
        key: "currency",
        width: "120px",
        render: (text, record, index) => {
          return Setting.getCurrencyWithFlag(text);
        },
      },
      {
        title: i18next.t("general:Payment"),
        dataIndex: "payment",
        key: "payment",
        width: "120px",
        render: (text, record) => {
          if (!text) {
            return text;
          }
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
      },
    ];

    return (
      <Table
        scroll={{x: "max-content"}}
        columns={columns}
        dataSource={this.props.transactions}
        rowKey={(record) => `${record.owner}/${record.name}`}
        size="middle"
        bordered
        pagination={{pageSize: 10}}
      />
    );
  }
}

export default TransactionTable;
