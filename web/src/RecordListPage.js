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
import {Switch, Table} from 'antd';
import * as Setting from "./Setting";
import * as RecordBackend from "./backend/RecordBackend";
import i18next from "i18next";
import moment from "moment";

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
      owner: "built-in",
      name: "1234",
      id : "1234",
      clientIp: "::1",
      timestamp: moment().format(),
      organization: "built-in",
      username: "admin",
      requestUri: "/api/get-account",
      action: "login",
      isTriggered: false,
    }
  }

  renderTable(records) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '320px',
        sorter: (a, b) => a.name.localeCompare(b.name),
      },
      {
        title: i18next.t("general:ID"),
        dataIndex: 'id',
        key: 'id',
        width: '90px',
        sorter: (a, b) => a.id - b.id,
      },
      {
        title: i18next.t("general:Client IP"),
        dataIndex: 'clientIp',
        key: 'clientIp',
        width: '150px',
        sorter: (a, b) => a.clientIp.localeCompare(b.clientIp),
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={`https://db-ip.com/${text}`}>
              {text}
            </a>
          )
        }
      },
      {
        title: i18next.t("general:Timestamp"),
        dataIndex: 'createdTime',
        key: 'createdTime',
        width: '180px',
        sorter: (a, b) => a.createdTime.localeCompare(b.createdTime),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: 'organization',
        key: 'organization',
        width: '80px',
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
        width: '120px',
        sorter: (a, b) => a.user.localeCompare(b.user),
        render: (text, record, index) => {
          return (
            <Link to={`/users/${record.organization}/${record.user}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Request URI"),
        dataIndex: 'requestUri',
        key: 'requestUri',
        // width: '300px',
        sorter: (a, b) => a.requestUri.localeCompare(b.requestUri),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: 'action',
        key: 'action',
        width: '200px',
        sorter: (a, b) => a.action.localeCompare(b.action),
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return text;
        }
      },
      {
        title: i18next.t("record:Is Triggered"),
        dataIndex: 'isTriggered',
        key: 'isTriggered',
        width: '140px',
        sorter: (a, b) => a.isTriggered - b.isTriggered,
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          if (!["signup", "login", "logout", "update-user"].includes(record.action)) {
            return null;
          }

          return (
            <Switch disabled checkedChildren="ON" unCheckedChildren="OFF" checked={text} />
          )
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
