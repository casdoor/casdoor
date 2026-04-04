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
import {Link} from "react-router-dom";
import {Button, Card, Col, Input, Row, Select} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as EntryBackend from "./backend/EntryBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import EntryMessageViewer from "./EntryMessageViewer";

const {Option} = Select;
class EntryEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      entryName: props.match.params.entryName,
      owner: props.match.params.organizationName,
      entry: null,
      organizations: [],
      applications: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getEntry();
    this.getOrganizations();
    this.getApplications(this.state.owner);
  }

  getEntry() {
    EntryBackend.getEntry(this.state.entry?.owner || this.state.owner, this.state.entryName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "ok") {
          this.setState({
            entry: res.data,
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

  updateEntryField(key, value) {
    const entry = this.state.entry;
    if (key === "owner" && entry.owner !== value) {
      entry.application = "";
      this.getApplications(value);
    }

    entry[key] = value;
    this.setState({
      entry: entry,
    });
  }

  submitEntryEdit(willExit) {
    const entry = Setting.deepCopy(this.state.entry);
    EntryBackend.updateEntry(this.state.owner, this.state.entryName, entry)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully modified"));
          if (willExit) {
            this.props.history.push("/entries");
          } else {
            this.setState({
              mode: "edit",
              owner: entry.owner,
              entryName: entry.name,
            }, () => {this.getEntry();});
            this.props.history.push(`/entries/${entry.owner}/${entry.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to update")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteEntry() {
    EntryBackend.deleteEntry(this.state.entry)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.props.history.push("/entries");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  renderEntry() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("entry:New Entry") : i18next.t("entry:Edit Entry")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitEntryEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitEntryEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteEntry()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.entry.owner} onChange={(value => {this.updateEntryField("owner", value);})}>
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
            <Input value={this.state.entry.name} onChange={e => {
              this.updateEntryField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Display name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.entry.displayName} onChange={e => {
              this.updateEntryField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Provider")}:
          </Col>
          <Col span={22} >
            {this.state.entry.provider ? (
              <Link to={`/providers/${this.state.entry.owner}/${this.state.entry.provider}`}>
                {this.state.entry.provider}
              </Link>
            ) : null}
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Type"), i18next.t("general:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled value={this.state.entry.type ?? ""} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Listening URL"), i18next.t("general:Listening URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.entry.url} onChange={e => {
              this.updateEntryField("url", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("token:Access token"), i18next.t("token:Access token - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input.Password placeholder={"***"} value={this.state.entry.token} onChange={e => {
              this.updateEntryField("token", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.entry.application} onChange={(value => {this.updateEntryField("application", value);})}>
              {
                this.state.applications.map((application, index) => <Option key={index} value={application.name}>{application.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <EntryMessageViewer entry={this.state.entry} labelSpan={(Setting.isMobile()) ? 22 : 2} contentSpan={22} />
      </Card>
    );
  }

  render() {
    if (this.state.entry === null) {
      return null;
    }

    return (
      <div>
        {this.renderEntry()}
      </div>
    );
  }
}

export default EntryEditPage;
