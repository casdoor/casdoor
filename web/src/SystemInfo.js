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

import {Card, Col, Divider, Progress, Row, Spin} from "antd";
import * as SystemBackend from "./backend/SystemInfo";
import React from "react";
import * as Setting from "./Setting";
import i18next from "i18next";

class SystemInfo extends React.Component {

  constructor(props) {
    super(props);
    this.state = {
      systemInfo: {cpuUsage: [], memoryUsed: 0, memoryTotal: 0},
      versionInfo: {},
      intervalId: null,
      loading: true,
    };
  }

  UNSAFE_componentWillMount() {
    SystemBackend.getSystemInfo().then(res => {
      this.setState({
        systemInfo: res.data,
        loading: false,
      });

      const id = setInterval(() => {
        SystemBackend.getSystemInfo().then(res => {
          this.setState({
            systemInfo: res.data,
          });
        }).catch(error => {
          Setting.showMessage("error", `System info failed to get: ${error}`);
        });
      }, 1000 * 2);
      this.setState({intervalId: id});
    }).catch(error => {
      Setting.showMessage("error", `System info failed to get: ${error}`);
    });

    SystemBackend.getVersionInfo().then(res => {
      this.setState({
        versionInfo: res.data,
      });
    }).catch(err => {
      Setting.showMessage("error", `Version info failed to get: ${err}`);
    });
  }

  componentWillUnmount() {
    if (this.state.intervalId !== null) {
      clearInterval(this.state.intervalId);
    }
  }

  render() {
    const cpuUi = this.state.systemInfo.cpuUsage?.length <= 0 ? i18next.t("system:Failed to get CPU usage") :
      this.state.systemInfo.cpuUsage.map((usage, i) => {
        return (
          <Progress key={i} percent={Number(usage.toFixed(1))} />
        );
      });

    const memUi = this.state.systemInfo.memoryUsed && this.state.systemInfo.memoryTotal && this.state.systemInfo.memoryTotal <= 0 ? i18next.t("system:Failed to get memory usage") :
      <div>
        {Setting.getFriendlyFileSize(this.state.systemInfo.memoryUsed)} / {Setting.getFriendlyFileSize(this.state.systemInfo.memoryTotal)}
        <br /> <br />
        <Progress type="circle" percent={Number((Number(this.state.systemInfo.memoryUsed) / Number(this.state.systemInfo.memoryTotal) * 100).toFixed(2))} />
      </div>;

    const link = this.state.versionInfo?.version !== "" ? `https://github.com/casdoor/casdoor/releases/tag/${this.state.versionInfo?.version}` : "";
    let versionText = this.state.versionInfo?.version !== "" ? this.state.versionInfo?.version : i18next.t("system:Unknown version");
    if (this.state.versionInfo?.commitOffset > 0) {
      versionText += ` (ahead+${this.state.versionInfo?.commitOffset})`;
    }

    if (!Setting.isMobile()) {
      return (
        <Row>
          <Col span={6}></Col>
          <Col span={12}>
            <Row gutter={[10, 10]}>
              <Col span={12}>
                <Card title={i18next.t("system:CPU Usage")} bordered={true} style={{textAlign: "center", height: "100%"}}>
                  {this.state.loading ? <Spin size="large" /> : cpuUi}
                </Card>
              </Col>
              <Col span={12}>
                <Card title={i18next.t("system:Memory Usage")} bordered={true} style={{textAlign: "center", height: "100%"}}>
                  {this.state.loading ? <Spin size="large" /> : memUi}
                </Card>
              </Col>
            </Row>
            <Divider />
            <Card title={i18next.t("system:About Casdoor")} bordered={true} style={{textAlign: "center"}}>
              <div>{i18next.t("system:An Identity and Access Management (IAM) / Single-Sign-On (SSO) platform with web UI supporting OAuth 2.0, OIDC, SAML and CAS")}</div>
              GitHub: <a target="_blank" rel="noreferrer" href="https://github.com/casdoor/casdoor">Casdoor</a>
              <br />
              {i18next.t("system:Version")}: <a target="_blank" rel="noreferrer" href={link}>{versionText}</a>
              <br />
              {i18next.t("system:Official website")}: <a target="_blank" rel="noreferrer" href="https://casdoor.org">https://casdoor.org</a>
              <br />
              {i18next.t("system:Community")}: <a target="_blank" rel="noreferrer" href="https://casdoor.org/#:~:text=Casdoor%20API-,Community,-GitHub">Get in Touch!</a>
            </Card>
          </Col>
          <Col span={6}></Col>
        </Row>
      );
    } else {
      return (
        <Row gutter={[16, 0]}>
          <Col span={24}>
            <Card title={i18next.t("system:CPU Usage")} bordered={true} style={{textAlign: "center", width: "100%"}}>
              {this.state.loading ? <Spin size="large" /> : cpuUi}
            </Card>
          </Col>
          <Col span={24}>
            <Card title={i18next.t("system:Memory Usage")} bordered={true} style={{textAlign: "center", width: "100%"}}>
              {this.state.loading ? <Spin size="large" /> : memUi}
            </Card>
          </Col>
          <Col span={24}>
            <Card title={i18next.t("system:About Casdoor")} bordered={true} style={{textAlign: "center"}}>
              <div>{i18next.t("system:An Identity and Access Management (IAM) / Single-Sign-On (SSO) platform with web UI supporting OAuth 2.0, OIDC, SAML and CAS")}</div>
              GitHub: <a target="_blank" rel="noreferrer" href="https://github.com/casdoor/casdoor">Casdoor</a>
              <br />
              {i18next.t("system:Version")}: <a target="_blank" rel="noreferrer" href={link}>{versionText}</a>
              <br />
              {i18next.t("system:Official website")}: <a target="_blank" rel="noreferrer" href="https://casdoor.org">https://casdoor.org</a>
              <br />
              {i18next.t("system:Community")}: <a target="_blank" rel="noreferrer" href="https://casdoor.org/#:~:text=Casdoor%20API-,Community,-GitHub">Get in Touch!</a>
            </Card>
          </Col>
        </Row>
      );
    }
  }
}

export default SystemInfo;
