// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
import {CheckCircleOutlined, ClockCircleOutlined, CloseCircleOutlined, SyncOutlined} from "@ant-design/icons";
import moment from "moment";
import * as Setting from "./Setting";
import * as TicketBackend from "./backend/TicketBackend";
import i18next from "i18next";
import BaseListPage from "./BaseListPage";
import PopconfirmModal from "./common/modal/PopconfirmModal";

class TicketListPage extends BaseListPage {
  newTicket() {
    const randomName = Setting.getRandomName();
    const owner = Setting.getRequestOrganization(this.props.account);
    return {
      owner: owner,
      name: `ticket_${randomName}`,
      createdTime: moment().format(),
      updatedTime: moment().format(),
      displayName: `New Ticket - ${randomName}`,
      user: this.props.account.name,
      title: "",
      content: "",
      state: "Open",
      messages: [],
    };
  }

  addTicket() {
    const newTicket = this.newTicket();
    TicketBackend.addTicket(newTicket)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push({pathname: `/tickets/${newTicket.owner}/${newTicket.name}`, mode: "add"});
          Setting.showMessage("success", i18next.t("general:Successfully added"));
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to add")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteTicket(i) {
    TicketBackend.deleteTicket(this.state.data[i])
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

  renderTable(tickets) {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "140px",
        fixed: "left",
        sorter: true,
        ...this.getColumnSearchProps("name"),
        render: (text, record, index) => {
          return (
            <Link to={`/tickets/${record.owner}/${text}`}>
              {text}
            </Link>
          );
        },
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
        width: "200px",
        sorter: true,
        ...this.getColumnSearchProps("displayName"),
      },
      {
        title: i18next.t("general:Title"),
        dataIndex: "title",
        key: "title",
        width: "200px",
        sorter: true,
        ...this.getColumnSearchProps("title"),
      },
      {
        title: i18next.t("general:User"),
        dataIndex: "user",
        key: "user",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("user"),
        render: (text, record, index) => {
          return (
            <a target="_blank" rel="noreferrer" href={Setting.getUserProfileUrl(text, this.props.account)}>
              {text}
            </a>
          );
        },
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "160px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:Updated time"),
        dataIndex: "updatedTime",
        key: "updatedTime",
        width: "160px",
        sorter: true,
        render: (text, record, index) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("general:State"),
        dataIndex: "state",
        key: "state",
        width: "120px",
        sorter: true,
        ...this.getColumnSearchProps("state"),
        render: (text, record, index) => {
          switch (text) {
          case "Open":
            return Setting.getTag("processing", i18next.t("ticket:Open"), <ClockCircleOutlined />);
          case "In Progress":
            return Setting.getTag("warning", i18next.t("ticket:In Progress"), <SyncOutlined spin />);
          case "Resolved":
            return Setting.getTag("success", i18next.t("ticket:Resolved"), <CheckCircleOutlined />);
          case "Closed":
            return Setting.getTag("default", i18next.t("ticket:Closed"), <CloseCircleOutlined />);
          default:
            return null;
          }
        },
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "180px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record, index) => {
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} onClick={() => this.props.history.push(`/tickets/${record.owner}/${record.name}`)}>{i18next.t("general:View")}</Button>
              {Setting.isAdminUser(this.props.account) ? (
                <PopconfirmModal
                  title={i18next.t("general:Sure to delete") + `: ${record.name} ?`}
                  onConfirm={() => this.deleteTicket(index)}
                >
                </PopconfirmModal>
              ) : null}
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
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={tickets} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Tickets")}&nbsp;&nbsp;&nbsp;&nbsp;
              <Button type="primary" size="small" onClick={this.addTicket.bind(this)}>{i18next.t("general:Add")}</Button>
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
    if (params.type !== undefined && params.type !== null) {
      field = "type";
      value = params.type;
    }
    this.setState({loading: true});
    TicketBackend.getTickets(Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account), params.pagination.current, params.pagination.pageSize, field, value, sortField, sortOrder)
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
          if (Setting.isResponseDenied(res)) {
            this.setState({
              isAuthorized: false,
            });
          } else {
            Setting.showMessage("error", res.msg);
          }
        }
      });
  };
}

export default TicketListPage;
