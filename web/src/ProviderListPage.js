// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
import {Button, Popconfirm, Table} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as Provider from "./auth/Provider";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";

class ProviderListPage extends BaseListPage {
  constructor(props) {
    super(props);
  }

  componentDidMount() {
    this.setState({
      owner: Setting.isAdminUser(this.props.account) ? "admin" : this.props.account.owner,
    });
  }

  newProvider() {
    const randomName = Setting.getRandomName();
    return {
      owner: this.state.owner,
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
    };
  }

  addProvider() {
    const newProvider = this.newProvider();
    ProviderBackend.addProvider(newProvider)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/providers/${newProvider.owner}/${newProvider.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteProvider(i) {
    ProviderBackend.deleteProvider(this.state.data[i])
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.setState({
            data: Setting.deleteRow(this.state.data, i),
            pagination: {total: this.state.pagination.total - 1},
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderTable(providers) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "120px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/providers/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("owner"),
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "180px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        // width: '100px',
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("provider:Category"),
        dataIndex: "category",
        key: "category",
        filterMultiple: false,
        filters: [
          {text: "OAuth", value: "OAuth"},
          {text: "Email", value: "Email"},
          {text: "SMS", value: "SMS"},
          {text: "Storage", value: "Storage"},
          {text: "SAML", value: "SAML"},
          {text: "Captcha", value: "Captcha"},
          {text: "Payment", value: "Payment"},
        ],
        width: "110px",
        sorter: true,
      },
      {
        title: i18next.t("provider:Type"),
        dataIndex: "type",
        key: "type",
        width: "110px",
        align: "center",
        filterMultiple: false,
        filters: [
          {text: "OAuth", value: "OAuth", children: Setting.getProviderTypeOptions("OAuth").map((o) => {return {text: o.id, value: o.name};})},
          {text: "Email", value: "Email", children: Setting.getProviderTypeOptions("Email").map((o) => {return {text: o.id, value: o.name};})},
          {text: "SMS", value: "SMS", children: Setting.getProviderTypeOptions("SMS").map((o) => {return {text: o.id, value: o.name};})},
          {text: "Storage", value: "Storage", children: Setting.getProviderTypeOptions("Storage").map((o) => {return {text: o.id, value: o.name};})},
          {text: "SAML", value: "SAML", children: Setting.getProviderTypeOptions("SAML").map((o) => {return {text: o.id, value: o.name};})},
          {text: "Captcha", value: "Captcha", children: Setting.getProviderTypeOptions("Captcha").map((o) => {return {text: o.id, value: o.name};})},
          {text: "Payment", value: "Payment", children: Setting.getProviderTypeOptions("Payment").map((o) => {return {text: o.id, value: o.name};})},
        ],
        sorter: true,
        render: (text, record, index) => {
          return Provider.getProviderLogoWidget(record);
        },
      },
      {
        title: i18next.t("provider:Client ID"),
        dataIndex: "clientId",
        key: "clientId",
        width: "100px",
        sorter: true,
        ...this.getColumnSearchProps("clientId"),
        render: (text, record, index) => {
          return Setting.getShortText(text);
        },
      },
      {
        title: i18next.t("provider:Provider URL"),
        dataIndex: "providerUrl",
        key: "providerUrl",
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("providerUrl"),
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={text}>
              {
                Setting.getShortText(text)
              }
            </a>
          );
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "170px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner)} style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/providers/${record.owner}/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteProvider(index)}
              >
                <Button disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner)} style={{marginBottom: "10px"}} type="primary" danger>{i18next.t("general:Delete")}</Button>
              </Popconfirm>
            </div>
          );
        },
      },
    ];

    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <div>
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={providers} rowKey="name" size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Providers")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addProvider.bind(this)}>{i18next.t("general:Add")}</Button>
            </div>
          )}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    let field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (params.category !== undefined && params.category !== null) {
      field = "category";
      value = params.category;
    } else if (params.type !== undefined && params.type !== null) {
      field = "type";
      value = params.type;
    }
    this.setState({loading: true});
    (Setting.isAdminUser(this.props.account) ? ProviderBackend.getGlobalProviders(params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      : ProviderBackend.getProviders(this.state.owner, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder))
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            loading: false,
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
          if (Setting.isResponseDenied(res)) {
            this.setState({
              loading: false,
              isAuthorized: false,
            });
          }
        }
      });
  };
}

export default ProviderListPage;
