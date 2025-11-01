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
        width: "150px",
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "160px",
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("transaction:Category"),
        dataIndex: "category",
        key: "category",
        width: "120px",
      },
      {
        title: i18next.t("transaction:Type"),
        dataIndex: "type",
        key: "type",
        width: "120px",
      },
      {
        title: i18next.t("transaction:Amount"),
        dataIndex: "amount",
        key: "amount",
        width: "100px",
        render: (text, record, index) => {
          return `${record.currency} ${text}`;
        },
      },
      {
        title: i18next.t("transaction:State"),
        dataIndex: "state",
        key: "state",
        width: "100px",
      },
    ];

    return (
      <Table
        scroll={{x: "max-content"}}
        columns={columns}
        dataSource={this.props.transactions}
        rowKey="name"
        size="middle"
        bordered
        pagination={{pageSize: 10}}
      />
    );
  }
}

export default TransactionTable;
