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
import {Button, Table} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as CertBackend from "./backend/CertBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class CertListPage extends BaseListPage {
  newCert() {
    const randomName = Setting.getRandomName();
    return {
      owner: this.props.account.owner, // this.props.account.certname,
      name: `cert_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Cert - ${randomName}`,
      scope: "JWT",
      type: "x509",
      cryptoAlgorithm: "RS256",
      bitSize: 4096,
      expireInYears: 20,
      certificate: "",
      privateKey: "",
    };
  }

  addCert() {
    const newCert = this.newCert();
    CertBackend.addCert(newCert)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/certs/${newCert.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteCert(i) {
    CertBackend.deleteCert(this.state.data[i])
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

  renderTable(certs) {
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
            <Link to={`/certs/${text}`}>
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
        ...this.getColumnSearchProps("organization"),
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
        title: i18next.t("provider:Scope"),
        dataIndex: "scope",
        key: "scope",
        filterMultiple: false,
        filters: [
          {text: "JWT", value: "JWT"},
        ],
        width: "110px",
        sorter: true,
      },
      {
        title: i18next.t("provider:Type"),
        dataIndex: "type",
        key: "type",
        filterMultiple: false,
        filters: [
          {text: "x509", value: "x509"},
        ],
        width: "110px",
        sorter: true,
      },
      {
        title: i18next.t("cert:Crypto algorithm"),
        dataIndex: "cryptoAlgorithm",
        key: "cryptoAlgorithm",
        filterMultiple: false,
        filters: [
          {text: "RS256", value: "RS256"},
        ],
        width: "190px",
        sorter: true,
      },
      {
        title: i18next.t("cert:Bit size"),
        dataIndex: "bitSize",
        key: "bitSize",
        width: "130px",
        sorter: true,
        ...this.getColumnSearchProps("bitSize"),
      },
      {
        title: i18next.t("cert:Expire in years"),
        dataIndex: "expireInYears",
        key: "expireInYears",
        width: "170px",
        sorter: true,
        ...this.getColumnSearchProps("expireInYears"),
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
              <Button disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner)} style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/certs/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <PopconfirmModal
                disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner)}
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteCert(index)}
              >
              </PopconfirmModal>
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={certs} rowKey="name" size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Certs")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addCert.bind(this)}>{i18next.t("general:Add")}</Button>
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
    (Setting.isAdminUser(this.props.account) ? CertBackend.getGlobleCerts(params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      : CertBackend.getCerts(this.props.account.owner, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder))
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

export default CertListPage;
