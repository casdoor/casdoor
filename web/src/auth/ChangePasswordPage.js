// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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
import {Button, Card, Col, Input, Row} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as ApplicationBackend from "../backend/ApplicationBackend";

class ChangePasswordPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      users: [],
      applicationName: props.applicationName ?? props.match.params?.applicationName,
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  componentDidMount() {
    if (this.getApplicationObj() === undefined) {
      if (this.state.applicationName !== undefined) {
        this.getApplication();
      } else {
        Setting.showMessage("error", i18next.t("forget:Unknown forget type") + ": " + this.state.type);
      }
    }
  }

  getApplication() {
    if (this.state.applicationName === undefined) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.onUpdateApplication(application);
      });
  }

  getApplicationObj() {
    return this.props.application;
  }

  onUpdateApplication(application) {
    this.props.onUpdateApplication(application);
  }

  renderChat() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("general:Password")}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "20px"}} >
          <Col span={22}>
            <Input.Password addonBefore={i18next.t("user:New Password")} placeholder={i18next.t("user:input password")} onChange={(e) => {}} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col span={22}>
            <Input.Password addonBefore={i18next.t("user:Re-enter New")} placeholder={i18next.t("user:input password")} onChange={(e) => {}} />
          </Col>
        </Row>
      </Card>
    );
  }

  render() {
    return (
      <div>
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large">{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large">{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large">{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default ChangePasswordPage;
