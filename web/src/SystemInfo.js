// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

import {Card, Col, Divider, Progress, Row} from "antd";
import {getGitHubLatestReleaseVersion, getSystemInfo} from "./backend/SystemInfo";
import React from "react";
import * as Setting from "./Setting";
import i18next from "i18next";

class SystemInfo extends React.Component {

  constructor(props) {
    super(props);
    this.state = {
      cpuUsage: 0,
      memUsage: 0,
      latestReleaseVersion: "v1.0.0",
      intervalId: null,
    };
  }

  componentDidMount() {
    // eslint-disable-next-line no-console
    getSystemInfo(this.props.account?.owner, this.props.account?.name).then(res => {
      this.setState({
        cpuUsage: Math.floor(res.cpuUsage),
        memUsage: Math.floor(res.memoryUsage),
      });

      const id = setInterval(() => {
        getSystemInfo(this.props.account?.owner, this.props.account?.name).then(res => {
          this.setState({
            cpuUsage: Math.floor(res.cpuUsage),
            memUsage: Math.floor(res.memoryUsage),
          });
        });
      }, 1000 * 3);
      this.setState({intervalId: id});
    }).catch(error => {
      Setting.showMessage("error", `System info failed to get: ${error}`);
    });

    getGitHubLatestReleaseVersion("casdoor/casdoor").then(res => {
      this.setState({latestReleaseVersion: res});
    }).catch(err => {
      Setting.showMessage("error", `get latest release version failed: ${err}`);
    });
  }

  componentWillUnmount() {
    clearInterval(this.state.intervalId);
  }

  render() {
    return (
      <Row>
        <Col span={6}></Col>
        <Col span={12}>
          <Row gutter={[10, 10]}>
            <Col span={12}>
              <Card title={i18next.t("system:Cpu Usage")} bordered={true} style={{textAlign: "center"}}>
                <Progress type="circle" percent={this.state.cpuUsage} />
              </Card>
            </Col>
            <Col span={12}>
              <Card title={i18next.t("system:Memory Usage")} bordered={true} style={{textAlign: "center"}}>
                <Progress type="circle" percent={this.state.memUsage} />
              </Card>
            </Col>
          </Row>
          <Divider />
          <Card title="About Casdoor" bordered={true} style={{textAlign: "center"}}>
            <div>{i18next.t("system:An Identity and Access Management (IAM) / Single-Sign-On (SSO) platform with web UI supporting OAuth 2.0, OIDC, SAML and CAS, QQ group 645200447")}</div>
            GitHub: <a href="https://github.com/casdoor/casdoor">casdoor</a>
            <br />
            {i18next.t("system:Latest Release Version")}: <a href={`https://github.com/casdoor/casdoor/releases/tag/${this.state.latestReleaseVersion}`}>{this.state.latestReleaseVersion}</a>
            <br />
            {i18next.t("system:Official Website")}: <a href="https://casdoor.org/">casdoor.org</a>
          </Card>
        </Col>
        <Col span={6}></Col>
      </Row>
    );
  }
}

export default SystemInfo;
