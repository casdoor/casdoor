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
import {Button, Table, Tag} from "antd";
import moment from "moment";
import * as Setting from "./Setting";
import * as BackupBackend from "./backend/BackupBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class BackupListPage extends BaseListPage {
  constructor(props) {
    super(props);
  }

  componentDidMount() {
    super.componentDidMount();
    this.setState({
      owner: Setting.isAdminUser(this.props.account) ? "admin" : this.props.account.owner,
    });
  }

  newBackup() {
    const randomName = Setting.getRandomName();
    const owner = Setting.isDefaultOrganizationSelected(this.props.account) ? this.state.owner : Setting.getRequestOrganization(this.props.account);
    return {
      owner: owner,
      name: `backup_${randomName}`,
      createdTime: moment().format(),
      displayName: `New Backup - ${randomName}`,
      description: "",
      host: "",
      port: 3306,
      database: "",
      username: "",
      password: "",
      backupFile: "",
      fileSize: 0,
      status: "Created",
    };
  }

  addBackup() {
    const newBackup = this.newBackup();
    BackupBackend.addBackup(newBackup)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/backups/${newBackup.owner}/${newBackup.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteBackup(i) {
    BackupBackend.deleteBackup(this.state.data[i])
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

  executeBackup(record) {
    BackupBackend.executeBackup(record.owner, record.name)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("backup:Backup started successfully"));
          this.fetch();
        } else {
          Setting.showMessage("error", `${i18next.t("backup:Failed to start backup")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  restoreBackup(record) {
    BackupBackend.restoreBackup(record.owner, record.name)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("backup:Backup restored successfully"));
        } else {
          Setting.showMessage("error", `${i18next.t("backup:Failed to restore backup")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderTable(backups) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "150px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/backups/${record.owner}/${text}`}>
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
        render: (text, record, index) => {
          return (text !== "admin") ? text : i18next.t("provider:admin (Shared)");
        },
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
        width: "150px",
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("backup:Database"),
        dataIndex: "database",
        key: "database",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("database"),
      },
      {
        title: i18next.t("backup:File size"),
        dataIndex: "fileSize",
        key: "fileSize",
        width: "120px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFriendlyFileSize(text);
        },
      },
      {
        title: i18next.t("general:Status"),
        dataIndex: "status",
        key: "status",
        width: "130px",
        sorter: true,
        render: (text, record, index) => {
          let color = "default";
          if (text === "Completed") {
            color = "success";
          } else if (text === "Failed") {
            color = "error";
          } else if (text === "InProgress") {
            color = "processing";
          }
          return <Tag color={color}>{text}</Tag>;
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "300px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button 
                disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner)} 
                style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} 
                onClick={() => this.executeBackup(record)}
              >
                {i18next.t("backup:Execute")}
              </Button>
              <PopconfirmModal
                disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner) || record.status !== "Completed"}
                title={i18next.t("backup:Sure to restore") + `: ${record.name} ?`}
                onConfirm={() => this.restoreBackup(record)}
              >
                <Button 
                  disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner) || record.status !== "Completed"}
                  style={{marginBottom: "10px", marginRight: "10px"}}
                  type="default"
                >
                  {i18next.t("backup:Restore")}
                </Button>
              </PopconfirmModal>
              <Button 
                disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner)} 
                style={{marginBottom: "10px", marginRight: "10px"}} 
                type="primary" 
                onClick={() => this.props.history.push(`/backups/${record.owner}/${record.name}`)}
              >
                {i18next.t("general:Edit")}
              </Button>
              <PopconfirmModal
                disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner)}
                title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                onConfirm={() => this.deleteBackup(index)}
              >
                <Button 
                  disabled={!Setting.isAdminUser(this.props.account) && (record.owner !== this.props.account.owner)} 
                  style={{marginBottom: "10px"}} 
                  type="danger"
                >
                  {i18next.t("general:Delete")}
                </Button>
              </PopconfirmModal>
            </div>
          );
        },
      },
    ];

    const paginationProps = {
      pageSize: this.state.pagination.pageSize,
      total: this.state.pagination.total,
      showSizeChanger: true,
      showQuickJumper: true,
      current: this.state.pagination.current,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
      position: ["topRight"],
      locale: {items_per_page: ""},
    };

    return (
      <div>
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={backups} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Backups")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button disabled={!Setting.isAdminUser(this.props.account) && (this.state.owner !== this.props.account.owner)} type="primary" size="small" onClick={this.addBackup.bind(this)}>{i18next.t("general:Add")}</Button>
            </div>
          )}
          loading={this.state.loading}
          onChange={this.handleTableChange}
        />
      </div>
    );
  }

  fetch = (params = {}) => {
    const field = params.searchedColumn, value = params.searchText;
    const sortField = params.sortField, sortOrder = params.sortOrder;
    this.setState({loading: true});
    if (Setting.isDefaultOrganizationSelected(this.props.account)) {
      BackupBackend.getBackups(this.state.owner, params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
        .then((res) => {
          this.setState({
            loading: false,
          });
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
            if (res.msg !== "Unauthorized operation") {
              Setting.showMessage("error", `Failed to get backups: ${res.msg}`);
            }
          }
        });
    } else {
      BackupBackend.getGlobalBackups(params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
        .then((res) => {
          this.setState({
            loading: false,
          });
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
            if (res.msg !== "Unauthorized operation") {
              Setting.showMessage("error", `Failed to get backups: ${res.msg}`);
            }
          }
        });
    }
  };
}

export default BackupListPage;
