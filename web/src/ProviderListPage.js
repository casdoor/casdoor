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
import {Button, Col, Popconfirm, Row, Table} from 'antd';
import moment from "moment";
import * as Setting from "./Setting";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as Provider from "./auth/Provider";
import i18next from "i18next";

class ProviderListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      providers: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getProviders();
  }

  getProviders() {
    ProviderBackend.getProviders("admin")
      .then((res) => {
        this.setState({
          providers: res,
        });
      });
  }

  newProvider() {
    return {
      owner: "admin", // this.props.account.providername,
      name: `provider_${this.state.providers.length}`,
      createdTime: moment().format(),
      displayName: `New Provider - ${this.state.providers.length}`,
      category: "OAuth",
      type: "GitHub",
      clientId: "",
      clientSecret: "",
      host: "",
      port: 0,
      providerUrl: "https://github.com/organizations/xxx/settings/applications/1234567",
    }
  }

  addProvider() {
    const newProvider = this.newProvider();
    ProviderBackend.addProvider(newProvider)
      .then((res) => {
          Setting.showMessage("success", `Provider added successfully`);
          this.setState({
            providers: Setting.prependRow(this.state.providers, newProvider),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Provider failed to add: ${error}`);
      });
  }

  deleteProvider(i) {
    ProviderBackend.deleteProvider(this.state.providers[i])
      .then((res) => {
          Setting.showMessage("success", `Provider deleted successfully`);
          this.setState({
            providers: Setting.deleteRow(this.state.providers, i),
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Provider failed to delete: ${error}`);
      });
  }

  renderTable(providers) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '120px',
        sorter: (a, b) => a.name.localeCompare(b.name),
        render: (text, record, index) => {
          return (
            <Link to={`/providers/${text}`}>
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
        sorter: (a, b) => a.createdTime.localeCompare(b.createdTime),
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        }
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: 'displayName',
        key: 'displayName',
        // width: '100px',
        sorter: (a, b) => a.displayName.localeCompare(b.displayName),
      },
      {
        title: i18next.t("provider:Category"),
        dataIndex: 'category',
        key: 'category',
        width: '100px',
        sorter: (a, b) => a.category.localeCompare(b.category),
      },
      {
        title: i18next.t("provider:Type"),
        dataIndex: 'type',
        key: 'type',
        width: '80px',
        sorter: (a, b) => a.type.localeCompare(b.type),
        render: (text, record, index) => {
          if (record.category !== "OAuth") {
            return text;
          } else {
            return (
              <img width={30} height={30} src={Provider.getAuthLogo(record)} alt={record.displayName} />
            )
          }
        }
      },
      {
        title: i18next.t("provider:Client ID"),
        dataIndex: 'clientId',
        key: 'clientId',
        width: '150px',
        sorter: (a, b) => a.clientId.localeCompare(b.clientId),
      },
      // {
      //   title: 'Client secret',
      //   dataIndex: 'clientSecret',
      //   key: 'clientSecret',
      //   width: '150px',
      //   sorter: (a, b) => a.clientSecret.localeCompare(b.clientSecret),
      // },
      {
        title: i18next.t("provider:Provider URL"),
        dataIndex: 'providerUrl',
        key: 'providerUrl',
        width: '150px',
        sorter: (a, b) => a.providerUrl.localeCompare(b.providerUrl),
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              {
                Setting.getShortText(text)
              }
            </a>
          )
        }
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: '',
        key: 'op',
        width: '170px',
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/providers/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete provider: ${record.name} ?`}
                onConfirm={() => this.deleteProvider(index)}
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
        <Table columns={columns} dataSource={providers} rowKey="name" size="middle" bordered pagination={{pageSize: 100}}
               title={() => (
                 <div>
                   {i18next.t("general:Providers")}&nbsp;&nbsp;&nbsp;&nbsp;
                   <Button type="primary" size="small" onClick={this.addProvider.bind(this)}>{i18next.t("general:Add")}</Button>
                 </div>
               )}
               loading={providers === null}
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
              this.renderTable(this.state.providers)
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    );
  }
}

export default ProviderListPage;
