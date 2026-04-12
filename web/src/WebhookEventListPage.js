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
import {Button, Descriptions, Drawer, Result, Table, Tag} from "antd";
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
      stateFilter: "",
      sortField: "",
      sortOrder: "",
      detailShow: false,
      detailRecord: null,
      pagination: {
        current: 1,
        pageSize: 10,
        total: 0,
      },
    };
  }

  getTableLoading = () => {
    return this.state.loading ? {tip: i18next.t("login:Loading")} : false;
  };

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
    this.fetchWebhookEvents(pagination, this.state.stateFilter, this.state.sortField, this.state.sortOrder);
  };

  getStateTag = (state) => {
    const stateConfig = {
      Pending: {color: "gold", text: i18next.t("webhook:Pending")},
      Success: {color: "green", text: i18next.t("webhook:Success")},
      Failed: {color: "red", text: i18next.t("webhook:Failed")},
      Retrying: {color: "blue", text: i18next.t("webhook:Retrying")},
    };

    const config = stateConfig[state] || {color: "default", text: state || i18next.t("webhook:Unknown")};

    return <Tag color={config.color}>{config.text}</Tag>;
  };

  getWebhookEditLink = (webhookId) => {
    if (!webhookId) {
      return "-";
    }

    const pathName = Setting.getShortName(webhookId);

    return (
      <Link to={`/webhooks/${encodeURIComponent(pathName)}`}>
        {pathName}
      </Link>
    );
  };

  getOrganizationFilter = () => {
    if (!this.props.account) {
      return "";
    }

    return Setting.isDefaultOrganizationSelected(this.props.account) ? "" : Setting.getRequestOrganization(this.props.account);
  };

  fetchWebhookEvents = (pagination = this.state.pagination, stateFilter = this.state.stateFilter, sortField = this.state.sortField, sortOrder = this.state.sortOrder) => {
    this.setState({loading: true});

    WebhookEventBackend.getWebhookEvents("", this.getOrganizationFilter(), pagination.current, pagination.pageSize, "", stateFilter, sortField, sortOrder)
      .then((res) => {
        this.setState({loading: false});

        if (res.status === "ok") {
          this.setState({
            data: res.data || [],
            stateFilter,
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
          this.fetchWebhookEvents(this.state.pagination, this.state.stateFilter, this.state.sortField, this.state.sortOrder);
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
    const stateFilter = Array.isArray(filters?.state) ? (filters.state[0] ?? "") : (filters?.state ?? "");
    const sortField = Array.isArray(sorter) ? "" : sorter?.field ?? "";
    const sortOrder = Array.isArray(sorter) ? "" : sorter?.order ?? "";
    const nextPagination = stateFilter !== this.state.stateFilter ? {
      ...pagination,
      current: 1,
    } : pagination;

    this.fetchWebhookEvents(nextPagination, stateFilter, sortField, sortOrder);
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
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "150px",
        fixed: "left",
        render: (text, record) => (
          <Button type="link" style={{paddingLeft: 0}} onClick={() => this.openDetailDrawer(record)}>
            {text}
          </Button>
        ),
      },
      {
        title: i18next.t("general:Webhook"),
        dataIndex: "webhook",
        key: "webhook",
        width: "160px",
        render: (text) => this.getWebhookEditLink(text),
      },
      {
        title: i18next.t("general:Organization"),
        dataIndex: "organization",
        key: "organization",
        width: "110px",
        render: (text) => {
          return text ? (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          ) : "-";
        },
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: "150px",
        sorter: true,
        sortOrder: this.state.sortField === "createdTime" ? this.state.sortOrder : null,
        render: (text) => (text ? Setting.getFormattedDate(text) : "-"),
      },
      {
        title: i18next.t("webhook:Attempt Count"),
        dataIndex: "attemptCount",
        key: "attemptCount",
        width: "150px",
        sorter: true,
        sortOrder: this.state.sortField === "attemptCount" ? this.state.sortOrder : null,
      },
      {
        title: i18next.t("webhook:Next Retry Time"),
        dataIndex: "nextRetryTime",
        key: "nextRetryTime",
        width: "150px",
        sorter: true,
        sortOrder: this.state.sortField === "nextRetryTime" ? this.state.sortOrder : null,
        render: (text) => {
          return text ? Setting.getFormattedDate(text) : "-";
        },
      },
      {
        title: i18next.t("general:State"),
        dataIndex: "state",
        key: "state",
        width: "120px",
        filters: [
          {text: i18next.t("webhook:Pending"), value: "Pending"},
          {text: i18next.t("webhook:Success"), value: "Success"},
          {text: i18next.t("webhook:Failed"), value: "Failed"},
          {text: i18next.t("webhook:Retrying"), value: "Retrying"},
        ],
        filterMultiple: false,
        filteredValue: this.state.stateFilter ? [this.state.stateFilter] : null,
        render: (text) => this.getStateTag(text),
      },
      {
        title: i18next.t("general:Action"),
        dataIndex: "",
        key: "op",
        width: "170px",
        fixed: (Setting.isMobile()) ? "false" : "right",
        render: (text, record) => {
          const eventId = `${record.owner}/${record.name}`;
          const canReplay = record.state !== "Success";
          return (
            <div>
              <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} onClick={() => this.openDetailDrawer(record)}>{i18next.t("general:View")}</Button>
              {canReplay ? (
                <Button style={{marginTop: "10px", marginBottom: "10px", marginRight: "10px"}} type="primary" loading={this.state.replayingId === eventId} onClick={() => this.replayWebhookEvent(record)}>{i18next.t("webhook:Replay")}</Button>
              ) : null}
            </div>
          );
        },
      },
    ];

    const paginationProps = {
      current: this.state.pagination.current,
      pageSize: this.state.pagination.pageSize,
      total: this.state.pagination.total,
      showQuickJumper: true,
      showSizeChanger: true,
      showTotal: () => i18next.t("general:{total} in total").replace("{total}", this.state.pagination.total),
    };

    return (
      <div>
        <Table scroll={{x: "max-content"}} columns={columns} dataSource={this.state.data} rowKey={(record) => `${record.owner}/${record.name}`} size="middle" bordered pagination={paginationProps}
          title={() => (
            <div>
              {i18next.t("general:Webhook Events")}
            </div>
          )}
          loading={this.getTableLoading()}
          onChange={this.handleTableChange}
        />
      </div>
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
      <div>
        {this.renderTable()}
        <Drawer
          title={i18next.t("general:Detail")}
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
            <Descriptions.Item label={i18next.t("general:Name")}>
              {this.getDetailField("name") || "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Webhook")}>
              {this.getWebhookEditLink(this.getDetailField("webhook"))}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Organization")}>
              {this.getDetailField("organization") ? (
                <Link to={`/organizations/${this.getDetailField("organization")}`}>
                  {this.getDetailField("organization")}
                </Link>
              ) : "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Created time")}>
              {this.getDetailField("createdTime") ? Setting.getFormattedDate(this.getDetailField("createdTime")) : "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:State")}>
              {this.getStateTag(this.getDetailField("state"))}
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
      </div>
    );
  }
}

export default WebhookEventListPage;
