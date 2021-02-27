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

import React from 'react';
import * as OAuthAppBackend from "./backend/OAuthAppBackend";
import { parseJson } from './Setting';
import {Link} from "react-router-dom";
import {Button, Col, Popconfirm, Row, Table} from 'antd';
import moment from "moment";
import * as Setting from "./Setting";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import i18next from "i18next";

class OAuthListPage extends React.Component {
  constructor(props) {
    super(props)
    this.state = {
      classes: props,
      apps: null,
      account: props.account,
    }
  }

  componentWillMount() {
    this.getOAuthApps();
  }

  getOAuthApps() {
    OAuthAppBackend.getOAuthApps(this.state.account.name)
    .then((res) => {
      this.setState({
        apps: res,
      });
    });
  }

  addOAuthApp() {
    this.props.history.push(`/oauth/newoauthapp`);
  }

  deleteOAuthApp(i) {
    console.log(this.state.apps[i])
    OAuthAppBackend.deleteOAuthApp(this.state.apps[i])
      .then((res) => {
          Setting.showMessage("success", `OAuth App deleted successfully`);
          this.setState({
            apps: Setting.deleteRow(this.state.apps, i),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `OAuth App failed to delete: ${error}`);
      });
  }

  renderTable(apps) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '120px',
        sorter: (a, b) => a.name.localeCompare(b.name),
      },
      {
        title: i18next.t("oauth:Client ID"),
        dataIndex: 'clientId',
        key: 'clientId',
        width: '300px',
      },
      {
        title: i18next.t("oauth:Client Secret"),
        dataIndex: 'clientSecret',
        key: 'clientSecret',
        width: '300px',
      },
      {
        title: i18next.t("oauth:Homepage URL"),
        dataIndex: 'domain',
        key: 'domain',
        width: '300px',
        sorter: (a, b) => a.domain.localeCompare(b.domain),
      },
      {
        title: i18next.t("oauth:Callback"),
        dataIndex: 'callback',
        key: 'callback',
        width: '300px',
        sorter: (a, b) => a.callback.localeCompare(b.callback),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: '',
        key: 'op',
        width: '170px',
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/oauth/editapp/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete OAuthApp: ${record.name} ?`}
                onConfirm={() => this.deleteOAuthApp(index)}
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
        <Table columns={columns} dataSource={apps} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                  {i18next.t("general:OAuth Apps")}&nbsp;&nbsp;&nbsp;&nbsp;
                  <Button type="primary" size="small" onClick={this.addOAuthApp.bind(this)}>{i18next.t("general:Add")}</Button>
                 </div>
               )}
               
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
              this.renderTable(this.state.apps)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default OAuthListPage;