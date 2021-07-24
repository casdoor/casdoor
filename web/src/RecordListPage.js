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
    };
  }

  UNSAFE_componentWillMount() {
    this.getRecords();
  }

  getRecords() {
    RecordBackend.getRecords()
      .then((res) => {
        this.setState({
          records: res,
        });
      });
  }

  newRecord() {
    return {
      id  : "",
      Record:{
        clientIp:"",
        timestamp:"",
        organization:"",
        username:"",
        requestUri:"",
        action:"login",
      },
    }
  }

  renderTable(records) {
    const columns = [
      {
        title: i18next.t("general:Client ip"),
        dataIndex: ['Record', 'clientIp'],
        key: 'id',
        width: '120px',
        fixed: 'left',
        sorter: (a, b) => a.Record.clientIp.localeCompare(b.Record.clientIp),
        render: (text, record, index) => {
          return text;
        }
      },
      {
        title: i18next.t("general:Timestamp"),
        dataIndex: ['Record', 'timestamp'],
        key: 'id',
        width: '160px',
        sorter: (a, b) => a.Record.timestamp.localeCompare(b.Record.timestamp),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: ['Record', 'organization'],
        key: 'id',
        width: '120px',
        sorter: (a, b) => a.Record.organization.localeCompare(b.Record.organization),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Username"),
        dataIndex: ['Record', 'username'],
        key: 'id',
        width: '160px',
        sorter: (a, b) => a.Record.username.localeCompare(b.Record.username),
        render: (text, record, index) => {
          return text;
        }
      },
      {
        title: i18next.t("general:Request uri"),
        dataIndex: ['Record', 'requestUri'],
        key: 'id',
        width: '160px',
        sorter: (a, b) => a.Record.requestUri.localeCompare(b.Record.requestUri),
        render: (text, record, index) => {
          return text;
        }
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: ['Record', 'action'],
        key: 'id',
        width: '160px',
        sorter: (a, b) => a.Record.action.localeCompare(b.Record.action),
        render: (text, record, index) => {
          return text;
        }
      },
    ];

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={records} rowKey="id" size="middle" bordered pagination={{pageSize: 100}}
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
