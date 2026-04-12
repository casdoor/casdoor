// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
import * as ServerBackend from "./backend/ServerBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";
import ScanServerModal from "./common/modal/ScanServerModal";

class ServerListPage extends BaseListPage {
  constructor(props) {
    super(props);
    this.state = {
      ...this.state,
      scanLoading: false,
      scanResult: null,
      scanServers: [],
      showScanModal: false,
      scanFilters: {
        cidrs: ["127.0.0.1/32"],
        ports: ["1-65535"],
        paths: ["/", "/mcp", "/sse", "/mcp/sse"],
      },
    };
  }

  newServer() {
    const randomName = Setting.getRandomName();
    const owner = Setting.getRequestOrganization(this.props.account);
    return {
      owner: owner,
      name: `server_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Server - ${randomName}`,
      url: "",
      application: "",
    };
  }

  addServer() {
    const newServer = this.newServer();
    ServerBackend.addServer(newServer)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/servers/${newServer.owner}/${newServer.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteServer(i) {
    ServerBackend.deleteServer(this.state.data[i])
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.fetch({
            pagination: {
              ...this.state.pagination,
              current: this.state.pagination.current > 1 && this.state.data.length === 1 ? this.state.pagination.current - 1 : this.state.pagination.current,
            },
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  fetch = (params = {}) => {
    const field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    if (!params.pagination) {
      params.pagination = {current: 1, pageSize: 10};
    }

    this.setState({loading: true});
    ServerBackend.getServers(Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
      .then((res) => {
        this.setState({loading: false});
        if (res.status === "ok") {
          this.setState({
            data: res.data,
            pagination: {
              ...params.pagination,
              total: res.data2,
            },
            searchText: params.searchText,
            searchedColumn: params.searchedColumn,
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      });
  };

  scanIntranetServers = (scanRequest) => {
    this.setState({scanLoading: true});
    ServerBackend.syncIntranetServers(scanRequest)
      .then((res) => {
        this.setState({scanLoading: false});
        if (res.status === "ok") {
          const scanResult = res.data ?? {};
          const scanServers = scanResult.servers ?? [];
          this.setState({scanResult: scanResult, scanServers: scanServers});
          Setting.showMessage("success", `${i18next.t("general:Successfully got")}: ${scanServers.length} server(s)`);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      })
      .catch(error => {
        this.setState({scanLoading: false});
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  openScanModal = () => {
    this.setState({showScanModal: true});
  };

  closeScanModal = () => {
    if (this.state.scanLoading) {
      return;
    }
    this.setState({showScanModal: false});
  };

  submitScan = () => {
    const cidr = this.state.scanFilters.cidrs
      .map(item => item.trim())
      .filter(item => item !== "");
    const ports = this.state.scanFilters.ports
      .map(item => `${item}`.trim())
      .filter(item => item !== "");
    const paths = this.state.scanFilters.paths
      .map(item => item.trim())
      .filter(item => item !== "");

    if (cidr.length === 0) {
      Setting.showMessage("error", i18next.t("server:Please select at least one IP range"));
      return;
    }
    if (ports.length === 0) {
      Setting.showMessage("error", i18next.t("server:Please select at least one port"));
      return;
    }

    const invalidPort = ports.find(item => !/^\d+$|^\d+\s*-\s*\d+$/.test(item));
    if (invalidPort !== undefined) {
      Setting.showMessage("error", `Invalid port expression: ${invalidPort}`);
      return;
    }

    this.scanIntranetServers({cidr: cidr, ports: ports, paths: paths});
  };

  addScannedServer = (scanServer) => {
    const owner = Setting.getRequestOrganization(this.props.account);
    const randomName = Setting.getRandomName();
    const newServer = {
      owner: owner,
      name: `server_${randomName}`,
      createdTime: moment().format(),
      displayName: `Scanned MCP ${scanServer.host}:${scanServer.port}`,
      url: scanServer.url,
      application: "",
    };

    ServerBackend.addServer(newServer)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully added"));
          const {pagination} = this.state;
          this.fetch({pagination});
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  renderTable(servers) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "160px",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/servers/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "owner",
        key: "owner",
        width: "130px",
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
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("general:URL"),
        dataIndex: "url",
        key: "url",
        sorter: true,
        ...this.getColumnSearchProps("url"),
        render: (text) => {
          if (!text) {
            return null;
          }

          return (
            <a target="_blank" rel="noreferrer" href={text}>
              {Setting.getShortText(text, 40)}
            </a>
          );
        },
      },
      {
        title: i18next.t("general:Application"),
        dataIndex: "application",
        key: "application",
        width: "140px",
        sorter: true,
        ...this.getColumnSearchProps("application"),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "op",
        key: "op",
        width: "180px",
        fixed: (Setting.isMobile()) ? false : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" onClick={() => this.props.history.push(`/servers/${record.owner}/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <PopconfirmModal title={i18next.t("general:Sure to delete") + `: ${record.name} ?`} onConfirm={() => this.deleteServer(index)}>
              </PopconfirmModal>
            </div>
          );
        },
      },
    ];

    const filteredColumns = Setting.filterTableColumns(columns, this.props.formItems ?? this.state.formItems);
    const paginationProps = {
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <>
        <Table
          scroll={{x: "max-content"}}
          dataSource={servers}
          columns={filteredColumns}
          rowKey={record => `${record.owner}/${record.name}`}
          pagination={{...this.state.pagination, ...paginationProps}}
          loading={this.getTableLoading()}
          onChange={this.handleTableChange}
          size="middle"
          bordered
          title={() => (
            <div>
              {i18next.t("server:Edit MCP Server")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={() => this.addServer()}>{i18next.t("general:Add")}</Button>
            &nbsp;
              <Button size="small" onClick={this.openScanModal}>{i18next.t("server:Scan server")}</Button>
            &nbsp;
              <Button size="small" onClick={() => this.props.history.push("/server-store")}>{i18next.t("general:MCP Store")}</Button>
            </div>
          )}
        />
        <ScanServerModal
          open={this.state.showScanModal}
          loading={this.state.scanLoading}
          scanFilters={this.state.scanFilters}
          scanResult={this.state.scanResult}
          scanServers={this.state.scanServers}
          onSubmit={this.submitScan}
          onCancel={this.closeScanModal}
          onChangeScanFilters={(patch) => this.setState(prevState => ({scanFilters: {...prevState.scanFilters, ...patch}}))}
          onAddScannedServer={this.addScannedServer}
        />
      </>
    );
  }
}

export default ServerListPage;
