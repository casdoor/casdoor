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
import {Button, Descriptions, Drawer, Result, Table, Tag, Tooltip} from "antd";
import i18next from "i18next";
import * as Setting from "./Setting";
import * as WebhookEventBackend from "./backend/WebhookEventBackend";
import Editor from "./common/Editor";

class WebhookEventListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      data: [],
      loading: false,
      replayingId: "",
      isAuthorized: true,
      statusFilter: "",
      sortField: "",
      sortOrder: "",
      detailShow: false,
      detailRecord: null,
      pagination: {
        current: 1,
        pageSize: 10,
        showQuickJumper: true,
        showSizeChanger: true,
        total: 0,
      },
    };
  }

  componentDidMount() {
    window.addEventListener("storageOrganizationChanged", this.handleOrganizationChange);
    this.fetchWebhookEvents(this.state.pagination);
  }

  componentWillUnmount() {
    window.removeEventListener("storageOrganizationChanged", this.handleOrganizationChange);
  }

  handleOrganizationChange = () => {
    const pagination = {
      ...this.state.pagination,
      current: 1,
    };
    this.fetchWebhookEvents(pagination, this.state.statusFilter, this.state.sortField, this.state.sortOrder);
  };

  getStatusTag = (status) => {
    const statusConfig = {
      pending: {color: "gold", text: i18next.t("webhook:Pending")},
      success: {color: "green", text: i18next.t("webhook:Success")},
      failed: {color: "red", text: i18next.t("webhook:Failed")},
      retrying: {color: "blue", text: i18next.t("webhook:Retrying")},
    };

    const config = statusConfig[status] || {color: "default", text: status || i18next.t("webhook:Unknown")};

    return <Tag color={config.color}>{config.text}</Tag>;
  };

  getWebhookLink = (webhookName) => {
    if (!webhookName) {
      return "-";
    }

    const shortName = Setting.getShortName(webhookName);

    return (
      <Tooltip title={webhookName}>
        <Link to={`/webhooks/${encodeURIComponent(shortName)}`}>
          {shortName}
        </Link>
      </Tooltip>
    );
  };

  getOrganizationFilter = () => {
    if (!this.props.account) {
      return "";
    }

    return Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account);
  };

  fetchWebhookEvents = (pagination = this.state.pagination, statusFilter = this.state.statusFilter, sortField = this.state.sortField, sortOrder = this.state.sortOrder) => {
    this.setState({loading: true});

    WebhookEventBackend.getWebhookEvents("", this.getOrganizationFilter(), pagination.current, pagination.pageSize, "", statusFilter, sortField, sortOrder)
      .then((res) => {
        this.setState({loading: false});

        if (res.status === "ok") {
          this.setState({
            data: res.data || [],
            statusFilter,
            sortField,
            sortOrder,
            pagination: {
              ...pagination,
              total: res.data2 ?? 0,
            },
          });
        } else if (Setting.isResponseDenied(res)) {
          this.setState({isAuthorized: false});
        } else {
          Setting.showMessage("error", res.msg);
        }
      })
      .catch((error) => {
        this.setState({loading: false});
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  replayWebhookEvent = (event) => {
    const eventId = `${event.owner}/${event.name}`;
    this.setState({replayingId: eventId});

    WebhookEventBackend.replayWebhookEvent(eventId)
      .then((res) => {
        this.setState({replayingId: ""});

        if (res.status === "ok") {
          Setting.showMessage("success", typeof res.data === "string" ? res.data : i18next.t("webhook:Webhook event replay triggered"));
          this.fetchWebhookEvents(this.state.pagination, this.state.statusFilter, this.state.sortField, this.state.sortOrder);
        } else {
          Setting.showMessage("error", `${i18next.t("webhook:Failed to replay webhook event")}: ${res.msg}`);
        }
      })
      .catch((error) => {
        this.setState({replayingId: ""});
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  };

  handleTableChange = (pagination, filters, sorter) => {
    const statusFilter = Array.isArray(filters?.status) ? (filters.status[0] ?? "") : (filters?.status ?? "");
    const sortField = Array.isArray(sorter) ? "" : sorter?.field ?? "";
    const sortOrder = Array.isArray(sorter) ? "" : sorter?.order ?? "";
    const nextPagination = statusFilter !== this.state.statusFilter ? {
      ...pagination,
      current: 1,
    } : pagination;

    this.fetchWebhookEvents(nextPagination, statusFilter, sortField, sortOrder);
  };

  openDetailDrawer = (record) => {
    this.setState({
      detailRecord: record,
      detailShow: true,
    });
  };

  closeDetailDrawer = () => {
    this.setState({
      detailShow: false,
      detailRecord: null,
    });
  };

  getEditorMaxWidth = () => {
    return Setting.isMobile() ? window.innerWidth - 80 : 520;
  };

  jsonStrFormatter = (str) => {
    if (!str) {
      return "";
    }

    try {
      return JSON.stringify(JSON.parse(str), null, 2);
    } catch (e) {
      return str;
    }
  };

  getDetailField = (field) => {
    return this.state.detailRecord ? this.state.detailRecord[field] ?? "" : "";
  };

  renderTable = () => {
    const columns = [
      {
        title: i18next.t("webhook:Webhook Name"),
        dataIndex: "webhookName",
        key: "webhookName",
        width: 220,
        render: (text) => this.getWebhookLink(text),
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: 160,
        render: (text) => text ? <Link to={`/organizations/${text}`}>{text}</Link> : "-",
      },
      {
        title: i18next.t("general:Status"),
        dataIndex: "status",
        key: "status",
        width: 140,
        filters: [
          {text: i18next.t("webhook:Pending"), value: "pending"},
          {text: i18next.t("webhook:Success"), value: "success"},
          {text: i18next.t("webhook:Failed"), value: "failed"},
          {text: i18next.t("webhook:Retrying"), value: "retrying"},
        ],
        filterMultiple: false,
        filteredValue: this.state.statusFilter ? [this.state.statusFilter] : null,
        render: (text) => this.getStatusTag(text),
      },
      {
        title: i18next.t("webhook:Attempt Count"),
        dataIndex: "attemptCount",
        key: "attemptCount",
        width: 140,
        sorter: true,
        sortOrder: this.state.sortField === "attemptCount" ? this.state.sortOrder : null,
      },
      {
        title: i18next.t("webhook:Next Retry Time"),
        dataIndex: "nextRetryTime",
        key: "nextRetryTime",
        width: 180,
        sorter: true,
        sortOrder: this.state.sortField === "nextRetryTime" ? this.state.sortOrder : null,
        render: (text) => text ? Setting.getFormattedDate(text) : "-",
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "action",
        key: "action",
        width: 180,
        fixed: Setting.isMobile() ? false : "right",
        render: (_, record) => {
          const eventId = `${record.owner}/${record.name}`;
          return (
            <>
              <Button
                type="link"
                style={{paddingLeft: 0}}
                onClick={() => this.openDetailDrawer(record)}
              >
                {i18next.t("general:View")}
              </Button>
              <Button
                type="primary"
                loading={this.state.replayingId === eventId}
                onClick={() => this.replayWebhookEvent(record)}
              >
                {i18next.t("webhook:Replay")}
              </Button>
            </>
          );
        },
      },
    ];

    return (
      <Table
        rowKey={(record) => `${record.owner}/${record.name}`}
        columns={columns}
        dataSource={this.state.data}
        loading={this.getTableLoading()}
        pagination={{
          ...this.state.pagination,
          showTotal: (total) => i18next.t("general:{total} in total").replace("{total}", total),
        }}
        scroll={{x: "max-content"}}
        size="middle"
        bordered
        title={() => i18next.t("webhook:Webhook Event Logs")}
        onChange={this.handleTableChange}
      />
    );
  };

  render() {
    if (!this.state.isAuthorized) {
      return (
        <Result
          status="403"
          title={`403 ${i18next.t("general:Unauthorized")}`}
          subTitle={i18next.t("general:Sorry, you do not have permission to access this page or logged in status invalid.")}
          extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>}
        />
      );
    }

    return (
      <>
        {this.renderTable()}
        <Drawer
          title={i18next.t("webhook:Webhook Event Detail")}
          width={Setting.isMobile() ? "100%" : 720}
          placement="right"
          destroyOnClose
          onClose={this.closeDetailDrawer}
          open={this.state.detailShow}
        >
          <Descriptions
            bordered
            size="small"
            column={1}
            layout={Setting.isMobile() ? "vertical" : "horizontal"}
            style={{padding: "12px", height: "100%", overflowY: "auto"}}
          >
            <Descriptions.Item label={i18next.t("webhook:Webhook Name")}>
              {this.getDetailField("webhookName") ? this.getWebhookLink(this.getDetailField("webhookName")) : "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Organization")}>
              {this.getDetailField("organization") ? (
                <Link to={`/organizations/${this.getDetailField("organization")}`}>
                  {this.getDetailField("organization")}
                </Link>
              ) : "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Status")}>
              {this.getStatusTag(this.getDetailField("status"))}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("webhook:Attempt Count")}>
              {this.getDetailField("attemptCount") || 0}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("webhook:Next Retry Time")}>
              {this.getDetailField("nextRetryTime") ? Setting.getFormattedDate(this.getDetailField("nextRetryTime")) : "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("webhook:Payload")}>
              <Editor
                value={this.jsonStrFormatter(this.getDetailField("payload"))}
                lang="json"
                fillHeight
                fillWidth
                maxWidth={this.getEditorMaxWidth()}
                dark
                readOnly
              />
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("webhook:Last Error")}>
              <Editor
                value={this.getDetailField("lastError") || "-"}
                fillHeight
                fillWidth
                maxWidth={this.getEditorMaxWidth()}
                dark
                readOnly
              />
            </Descriptions.Item>
          </Descriptions>
        </Drawer>
      </>
    );
  }
}

export default WebhookEventListPage;
