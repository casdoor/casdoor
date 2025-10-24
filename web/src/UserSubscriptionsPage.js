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
import {Button, Card, Col, Row, Table, Popconfirm, Tag} from "antd";
import * as Setting from "./Setting";
import i18next from "i18next";
import * as SubscriptionBackend from "./backend/SubscriptionBackend";

class UserSubscriptionsPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      subscriptions: [],
      loading: false,
    };
  }

  componentDidMount() {
    this.getSubscriptions();
  }

  getSubscriptions() {
    this.setState({loading: true});
    SubscriptionBackend.getUserSubscriptions(this.props.account.owner, this.props.account.name)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            subscriptions: res.data,
            loading: false,
          });
        } else {
          Setting.showMessage("error", res.msg);
          this.setState({loading: false});
        }
      });
  }

  cancelSubscription(subscription) {
    // Update subscription state to Suspended
    const updatedSubscription = {...subscription, state: "Suspended"};
    SubscriptionBackend.updateSubscription(subscription.owner, subscription.name, updatedSubscription)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.getSubscriptions();
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  getStateTag(state) {
    let color = "default";
    if (state === "Active") {
      color = "green";
    } else if (state === "Upcoming") {
      color = "blue";
    } else if (state === "Expired") {
      color = "red";
    } else if (state === "Suspended") {
      color = "orange";
    } else if (state === "Error") {
      color = "red";
    } else if (state === "Pending") {
      color = "gold";
    }
    return <Tag color={color}>{state}</Tag>;
  }

  render() {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        width: "150px",
      },
      {
        title: i18next.t("subscription:Plan"),
        dataIndex: "plan",
        key: "plan",
        width: "150px",
      },
      {
        title: i18next.t("subscription:Start time"),
        dataIndex: "startTime",
        key: "startTime",
        width: "180px",
        render: (text) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("subscription:End time"),
        dataIndex: "endTime",
        key: "endTime",
        width: "180px",
        render: (text) => {
          return Setting.getFormattedDate(text);
        },
      },
      {
        title: i18next.t("subscription:Period"),
        dataIndex: "period",
        key: "period",
        width: "120px",
      },
      {
        title: i18next.t("general:State"),
        dataIndex: "state",
        key: "state",
        width: "120px",
        render: (text) => {
          return this.getStateTag(text);
        },
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: "150px",
        render: (text, record) => {
          return (
            <div>
              {record.state === "Active" || record.state === "Upcoming" ? (
                <Popconfirm
                  title={i18next.t("subscription:Are you sure to cancel this subscription?")}
                  onConfirm={() => this.cancelSubscription(record)}
                  okText={i18next.t("general:OK")}
                  cancelText={i18next.t("general:Cancel")}
                >
                  <Button type="primary" danger size="small">
                    {i18next.t("subscription:Cancel")}
                  </Button>
                </Popconfirm>
              ) : null}
            </div>
          );
        },
      },
    ];

    return (
      <div>
        <Row style={{marginTop: "20px"}}>
          <Col span={24}>
            <Card
              title={i18next.t("subscription:My Subscriptions")}
              bordered={false}
            >
              <Table
                columns={columns}
                dataSource={this.state.subscriptions}
                rowKey="name"
                size="middle"
                bordered
                pagination={{pageSize: 10}}
                loading={this.state.loading}
              />
            </Card>
          </Col>
        </Row>
      </div>
    );
  }
}

export default UserSubscriptionsPage;
