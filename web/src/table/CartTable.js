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
import {Table} from "antd";
import i18next from "i18next";
import * as Setting from "../Setting";

class CartTable extends React.Component {
  render() {
    const columns = [
      {
        title: i18next.t("product:Name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "200px",
      },
      {
        title: i18next.t("product:Image"),
        dataIndex: "image",
        key: "image",
        width: "80px",
        render: (text, record, index) => {
          if (!text) {
            return null;
          }
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              <img src={text} alt={record.displayName} width={40} />
            </a>
          );
        },
      },
      {
        title: i18next.t("product:Price"),
        dataIndex: "price",
        key: "price",
        width: "120px",
        render: (text, record, index) => {
          return Setting.getCurrencySymbol(record.currency) + text;
        },
      },
      {
        title: i18next.t("product:Quantity"),
        dataIndex: "quantity",
        key: "quantity",
        width: "100px",
      },
      {
        title: i18next.t("general:Detail"),
        dataIndex: "detail",
        key: "detail",
      },
    ];

    const cart = this.props.cart || [];

    return (
      <Table
        columns={columns}
        dataSource={cart}
        rowKey={(record) => `${record.owner}/${record.name}`}
        size="small"
        pagination={false}
      />
    );
  }
}

export default CartTable;
