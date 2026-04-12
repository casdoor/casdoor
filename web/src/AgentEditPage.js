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
import Loading from "./common/Loading";
import {Button, Card, Col, Input, Row, Select} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as AgentBackend from "./backend/AgentBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";

const {Option} = Select;

class AgentEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      agentName: props.match.params.agentName,
      owner: props.match.params.organizationName,
      agent: null,
      organizations: [],
      applications: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getAgent();
    this.getOrganizations();
    this.getApplications(this.state.owner);
  }

  getAgent() {
    AgentBackend.getAgent(this.state.agent?.owner || this.state.owner, this.state.agentName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "ok") {
          this.setState({
            agent: res.data,
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      });
  }

  getOrganizations() {
    if (Setting.isAdminUser(this.props.account)) {
      OrganizationBackend.getOrganizations("admin")
        .then((res) => {
          this.setState({
            organizations: res.data || [],
          });
        });
    }
  }

  getApplications(owner) {
    ApplicationBackend.getApplicationsByOrganization("admin", owner)
      .then((res) => {
        this.setState({
          applications: res.data || [],
        });
      });
  }

  updateAgentField(key, value) {
    const agent = this.state.agent;
    if (key === "owner" && agent.owner !== value) {
      agent.application = "";
      this.getApplications(value);
    }

    agent[key] = value;
    this.setState({
      agent: agent,
    });
  }

  submitAgentEdit(willExit) {
    const agent = Setting.deepCopy(this.state.agent);
    AgentBackend.updateAgent(this.state.owner, this.state.agentName, agent)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully modified"));
          if (willExit) {
            this.props.history.push("/agents");
          } else {
            this.setState({
              mode: "edit",
              owner: agent.owner,
              agentName: agent.name,
            }, () => {this.getAgent();});
            this.props.history.push(`/agents/${agent.owner}/${agent.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to update")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteAgent() {
    AgentBackend.deleteAgent(this.state.agent)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.props.history.push("/agents");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderAgent() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("agent:New Agent") : i18next.t("agent:Edit Agent")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitAgentEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitAgentEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteAgent()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.agent.owner} onChange={(value => {this.updateAgentField("owner", value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.agent.name} onChange={e => {
              this.updateAgentField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Display name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.agent.displayName} onChange={e => {
              this.updateAgentField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Listening URL"), i18next.t("general:Listening URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.agent.url} onChange={e => {
              this.updateAgentField("url", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("token:Access token"), i18next.t("token:Access token - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input.Password placeholder={"***"} value={this.state.agent.token} onChange={e => {
              this.updateAgentField("token", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.agent.application} onChange={(value => {this.updateAgentField("application", value);})}>
              {
                this.state.applications.map((application, index) => <Option key={index} value={application.name}>{application.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
      </Card>
    );
  }

  render() {
    if (this.state.agent === null) {
      return <Loading type="page" tip={i18next.t("login:Loading")} />;
    }

    return (
      <div>
        {this.renderAgent()}
      </div>
    );
  }
}

export default AgentEditPage;
