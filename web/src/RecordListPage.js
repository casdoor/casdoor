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
import {Col, Row, Table} from 'antd';
import * as Setting from "./Setting";
import * as RecordBackend from "./backend/RecordBackend";

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
        title: Setting.I18n("general:Client ip"),
        dataIndex: 'Record',
        key: 'Record',
        width: '120px',
        sorter: (a, b) => a.Record.clientIp.localeCompare(b.Record.clientIp),
        render: (text, record, index) => {
          return text.clientIp;
        }
      },
      {
        title: Setting.I18n("general:Timestamp"),
        dataIndex: 'Record',
        key: 'Record',
        width: '160px',
        sorter: (a, b) => a.Record.timestamp.localeCompare(b.Record.timestamp),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text.timestamp);
        }
      },
      {
        title: Setting.I18n("general:Organization"),
        dataIndex: 'Record',
        key: 'Record',
        width: '120px',
        sorter: (a, b) => a.Record.organization.localeCompare(b.Record.organization),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text.organization}`}>
              {text.organization}
            </Link>
          )
        }
      },
      {
        title: Setting.I18n("general:Username"),
        dataIndex: 'Record',
        key: 'Record',
        width: '160px',
        sorter: (a, b) => a.Record.username.localeCompare(b.Record.username),
        render: (text, record, index) => {
          return text.username;
        }
      },
      {
        title: Setting.I18n("general:Request uri"),
        dataIndex: 'Record',
        key: 'Record',
        width: '160px',
        sorter: (a, b) => a.Record.requestUri.localeCompare(b.Record.requestUri),
        render: (text, record, index) => {
          return text.requestUri;
        }
      },
      {
        title: Setting.I18n("general:Action"),
        dataIndex: 'Record',
        key: 'Record',
        width: '160px',
        sorter: (a, b) => a.Record.action.localeCompare(b.Record.action),
        render: (text, record, index) => {
          return text.action;
        }
      },
    ];

    return (
      <div>
        <Table columns={columns} dataSource={records} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   {Setting.I18n("general:Records")}&nbsp;&nbsp;&nbsp;&nbsp;
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
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.renderTable(this.state.records)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default RecordListPage;
