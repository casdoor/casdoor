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
import {Button, Popconfirm, Table} from 'antd';
import moment from "moment";
import * as Setting from "./Setting";
import * as WebhookBackend from "./backend/WebhookBackend";
import i18next from "i18next";

class WebhookListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      webhooks: null,
      total: 0,
    };
  }

  UNSAFE_componentWillMount() {
    this.getWebhooks(1, 10);
  }

  getWebhooks(page, pageSize) {
    WebhookBackend.getWebhooks("admin", page, pageSize)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            webhooks: res.data,
            total: res.data2
          });
        }
      });
  }

  newWebhook() {
    var randomName = Math.random().toString(36).slice(-6)
    return {
      owner: "admin", // this.props.account.webhookname,
      name: `webhook_${randomName}`,
      createdTime: moment().format(),
      url: "https://example.com/callback",
      contentType: "application/json",
      events: [],
      organization: "built-in",
    }
  }

  addWebhook() {
    const newWebhook = this.newWebhook();
    WebhookBackend.addWebhook(newWebhook)
      .then((res) => {
          Setting.showMessage("success", `Webhook added successfully`);
          this.setState({
            webhooks: Setting.prependRow(this.state.webhooks, newWebhook),
            total: this.state.total + 1
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Webhook failed to add: ${error}`);
      });
  }

  deleteWebhook(i) {
    WebhookBackend.deleteWebhook(this.state.webhooks[i])
      .then((res) => {
          Setting.showMessage("success", `Webhook deleted successfully`);
          this.setState({
            webhooks: Setting.deleteRow(this.state.webhooks, i),
            total: this.state.total - 1
          });
        }
      )
      .catch(error => {
        Setting.showMessage("error", `Webhook failed to delete: ${error}`);
      });
  }

  renderTable(webhooks) {
    const columns = [
      {
        title: i18next.t("general:Organization"),
        dataIndex: 'organization',
        key: 'organization',
        width: '80px',
        sorter: (a, b) => a.organization.localeCompare(b.organization),
        render: (text, record, index) => {
          return (
            <Link to={`/organizations/${text}`}>
              {text}
            </Link>
          )
        }
      },
      {
        title: i18next.t("general:Name"),
        dataIndex: 'name',
        key: 'name',
        width: '150px',
        fixed: 'left',
        sorter: (a, b) => a.name.localeCompare(b.name),
        render: (text, record, index) => {
          return (
            <Link to={`/webhooks/${text}`}>
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
        title: i18next.t("webhook:URL"),
        dataIndex: 'url',
        key: 'url',
        width: '300px',
        sorter: (a, b) => a.url.localeCompare(b.url),
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
        title: i18next.t("webhook:Content type"),
        dataIndex: 'contentType',
        key: 'contentType',
        width: '150px',
        sorter: (a, b) => a.contentType.localeCompare(b.contentType),
      },
      {
        title: i18next.t("webhook:Events"),
        dataIndex: 'events',
        key: 'events',
        // width: '100px',
        sorter: (a, b) => a.events.localeCompare(b.events),
        render: (text, record, index) => {
          return Setting.getTags(text);
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
              <Button style={{marginTop: '10px', marginBottom: '10px', marginRight: '10px'}} type="primary" onClick={() => this.props.history.push(`/webhooks/${record.name}`)}>{i18next.t("general:Edit")}</Button>
              <Popconfirm
                title={`Sure to delete webhook: ${record.name} ?`}
                onConfirm={() => this.deleteWebhook(index)}
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
      onChange: (page, pageSize) => this.getWebhooks(page, pageSize),
      onShowSizeChange: (current, size) => this.getWebhooks(current, size),
    };

    return (
      <div>
        <Table scroll={{x: 'max-content'}} columns={columns} dataSource={webhooks} rowKey="name" size="middle" bordered pagination={paginationProps}
               title={() => (
                 <div>
                   {i18next.t("general:Webhooks")}&nbsp;&nbsp;&nbsp;&nbsp;
                   <Button type="primary" size="small" onClick={this.addWebhook.bind(this)}>{i18next.t("general:Add")}</Button>
                 </div>
               )}
               loading={webhooks === null}
        />
      </div>
    );
  }

  render() {
    return (
      <div>
        {
          this.renderTable(this.state.webhooks)
        }
      </div>
    );
  }
}

export default WebhookListPage;
