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
import {Table} from 'antd';
import * as Setting from "./Setting";
import * as RecordBackend from "./backend/RecordBackend";
import i18next from "i18next";

class RecordListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      records: null,
      total: 0,
    };
  }

  UNSAFE_componentWillMount() {
    this.getRecords(1, 10);
  }

  getRecords(page, pageSize) {
    RecordBackend.getRecords(page, pageSize)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            records: res.data,
            total: res.data2
          });
        }
      });
  }

  newRecord() {
    return {
      id  : "",
      clientIp: "",
      timestamp: "",
      organization: "",
      username: "",
      requestUri: "",
      action: "login",
    }
  }

  renderTable(records) {
    const columns = [
      {
        title: i18next.t("general:ID"),
        dataIndex: 'id',
        key: 'id',
        width: '120px',
        sorter: (a, b) => a.id - b.id,
      },
      {
        title: i18next.t("general:Client IP"),
        dataIndex: 'clientIp',
        key: 'clientIp',
        width: '120px',
        sorter: (a, b) => a.clientIp.localeCompare(b.clientIp),
      },
      {
        title: i18next.t("general:Timestamp"),
        dataIndex: 'timestamp',
        key: 'timestamp',
        width: '160px',
        sorter: (a, b) => a.timestamp.localeCompare(b.timestamp),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: 'organization',
        key: 'organization',
        width: '120px',
        sorter: (a, b) => a.organization.localeCompare(b.organization),
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
        width: '160px',
        sorter: (a, b) => a.user.localeCompare(b.user),
      },
      {
        title: i18next.t("general:Request URI"),
        dataIndex: 'requestUri',
        key: 'requestUri',
        width: '160px',
        sorter: (a, b) => a.requestUri.localeCompare(b.requestUri),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: 'action',
        key: 'action',
        width: '160px',
        sorter: (a, b) => a.action.localeCompare(b.action),
        render: (text, record, index) => {
          return text;
        }
      },
    ];

    const paginationProps = {
      total: this.state.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.total),
      onChange: (page, pageSize) => this.getRecords(page, pageSize),
      onShowSizeChange: (current, size) => this.getRecords(current, size),
    };

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={records} rowKey="id" size="middle" bordered pagination={paginationProps}
               title={() => (
                 <div>
                   {i18next.t("general:Records")}&nbsp;&nbsp;&nbsp;&nbsp;
                 </div>
               )}
               loading={records === null}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.state.records)
        }
      </div>
    );
  }
}

export default RecordListPage;
