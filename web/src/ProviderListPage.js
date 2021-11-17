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
import {Button, Popconfirm, Table, Tooltip} from 'antd';
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
      total: 0,
    };
  }

  UNSAFE_componentWillMount() {
    this.getProviders(1, 10);
  }

  getProviders(page, pageSize) {
    ProviderBackend.getProviders("admin", page, pageSize)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            providers: res.data,
            total: res.data2
          });
        }
      });
  }

  newProvider() {
    var randomName = Math.random().toString(36).slice(-6)
    return {
      owner: "admin", // this.props.account.providername,
      name: `provider_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Provider - ${randomName}`,
      category: "OAuth",
      type: "GitHub",
      method: "Normal",
      clientId: "",
      clientSecret: "",
      enableSignUp: true,
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
            total: this.state.total + 1
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
            total: this.state.total - 1
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
        fixed: 'left',
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
        width: '180px',
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
        align: 'center',
        sorter: (a, b) => a.type.localeCompare(b.type),
        render: (text, record, index) => {
          const url = Provider.getProviderUrl(record);
          if (url !== "") {
            return (
              <Tooltip title={record.type}>
                <a target="_blank" rel="noreferrer" href={Provider.getProviderUrl(record)}>
                  <img width={36} height={36} src={Provider.getProviderLogo(record)} alt={record.displayName} />
                </a>
              </Tooltip>
            )
          } else {
            return (
              <Tooltip title={record.type}>
                <img width={36} height={36} src={Provider.getProviderLogo(record)} alt={record.displayName} />
              </Tooltip>
            )
          }
        }
      },
      {
        title: i18next.t("provider:Client ID"),
        dataIndex: 'clientId',
        key: 'clientId',
        width: '100px',
        sorter: (a, b) => a.clientId.localeCompare(b.clientId),
        render: (text, record, index) => {
          return Setting.getShortText(text);
        }
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
        fixed: (Setting.isMobile()) ? "false" : "right",
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

    const paginationProps = {
      total: this.state.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.total),
      onChange: (page, pageSize) => this.getProviders(page, pageSize),
      onShowSizeChange: (current, size) => this.getProviders(current, size),
    };

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={providers} rowKey="name" size="middle" bordered pagination={paginationProps}
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
        {
          this.renderTable(this.state.providers)
        }
      </div>
    );
  }
}

export default ProviderListPage;
