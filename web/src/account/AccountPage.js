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
import {Col, Descriptions, Row} from 'antd';
import * as AccountBackend from "../backend/AccountBackend";
import * as Setting from "../Setting";

class AccountPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      username: props.match.params.username,
      user: null,
    };
  }

  componentWillMount() {
    this.getUser();
  }

  getUser() {
    if (this.state.username !== undefined) {
      AccountBackend.getUser(this.state.username)
        .then((user) => {
          this.setState({
            user: user,
          });
        });
    }
  }

  renderValue(key) {
    if (this.props.account === null || this.props.account === undefined) {
      return <a href={"/login"}>Please sign in first</a>
    } else if (this.state.user !== null) {
      return this.state.user[key];
    } else {
      return this.props.account[key];
    }
  }

  renderContent() {
    return (
      <div>
        &nbsp;
        <Descriptions title="My Account" bordered>
          <Descriptions.Item label="Username">{this.renderValue("name")}</Descriptions.Item>
          <Descriptions.Item label="Organization">{this.renderValue("owner")}</Descriptions.Item>
          <Descriptions.Item label="Created At">{Setting.getFormattedDate(this.renderValue("createdTime"))}</Descriptions.Item>
          <Descriptions.Item label="Password Type">{this.renderValue("passwordType")}</Descriptions.Item>
          <Descriptions.Item label="Display Name">{this.renderValue("displayName")}</Descriptions.Item>
          <Descriptions.Item label="E-mail">{this.renderValue("email")}</Descriptions.Item>
          <Descriptions.Item label="Phone">{this.renderValue("phone")}</Descriptions.Item>
        </Descriptions>
      </div>
    );
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.renderContent()
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
      </div>
    )
  }
}

export default AccountPage;
