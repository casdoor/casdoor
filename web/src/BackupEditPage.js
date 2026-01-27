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
import {Button, Card, Col, Input, InputNumber, Row, Select} from "antd";
import * as BackupBackend from "./backend/BackupBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const {Option} = Select;
const {TextArea} = Input;

class BackupEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      backupName: props.match.params.backupName,
      owner: props.match.params.organizationName,
      backup: null,
      organizations: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getBackup();
    this.getOrganizations();
  }

  getBackup() {
    BackupBackend.getBackup(this.state.owner, this.state.backupName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          backup: res.data,
        });
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: res.data || [],
        });
      });
  }

  parseBackupField(key, value) {
    if (["port", "fileSize"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateBackupField(key, value) {
    value = this.parseBackupField(key, value);

    const backup = this.state.backup;
    backup[key] = value;
    this.setState({
      backup: backup,
    });
  }

  renderBackup() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("backup:New Backup") : i18next.t("backup:Edit Backup")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitBackupEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitBackupEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteBackup()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.backup.owner} onChange={(value => {this.updateBackupField("owner", value);})}>
              {Setting.isAdminUser(this.props.account) ? <Option key={"admin"} value={"admin"}>{i18next.t("provider:admin (Shared)")}</Option> : null}
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.backup.name} onChange={e => {
              this.updateBackupField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.backup.displayName} onChange={e => {
              this.updateBackupField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <TextArea autoSize={{minRows: 1, maxRows: 100}} value={this.state.backup.description} onChange={e => {
              this.updateBackupField("description", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("backup:Host"), i18next.t("backup:Host - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.backup.host} onChange={e => {
              this.updateBackupField("host", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("backup:Port"), i18next.t("backup:Port - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.backup.port} onChange={value => {
              this.updateBackupField("port", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("backup:Database"), i18next.t("backup:Database - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.backup.database} onChange={e => {
              this.updateBackupField("database", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("backup:Username"), i18next.t("backup:Username - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.backup.username} onChange={e => {
              this.updateBackupField("username", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("backup:Password"), i18next.t("backup:Password - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input.Password value={this.state.backup.password} onChange={e => {
              this.updateBackupField("password", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("backup:Backup file"), i18next.t("backup:Backup file - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.backup.backupFile} disabled />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("backup:File size"), i18next.t("backup:File size - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={Setting.getFriendlyFileSize(this.state.backup.fileSize)} disabled />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Status"), i18next.t("general:Status - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.backup.status} disabled />
          </Col>
        </Row>
      </Card>
    );
  }

  submitBackupEdit(exitAfterSave) {
    const backup = Setting.deepCopy(this.state.backup);
    BackupBackend.updateBackup(this.state.backup.owner, this.state.backupName, backup)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            backupName: this.state.backup.name,
            owner: this.state.backup.owner,
          });

          if (exitAfterSave) {
            this.props.history.push("/backups");
          } else {
            this.props.history.push(`/backups/${this.state.backup.owner}/${this.state.backup.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateBackupField("name", this.state.backupName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteBackup() {
    BackupBackend.deleteBackup(this.state.backup)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/backups");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.backup !== null ? this.renderBackup() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitBackupEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitBackupEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteBackup()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default BackupEditPage;
