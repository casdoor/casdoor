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
import {Button, Card, Col, Input, InputNumber, Row, Select, Switch} from 'antd';
import {LinkOutlined} from "@ant-design/icons";
import * as SyncerBackend from "./backend/SyncerBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import SyncerTableColumnTable from "./SyncerTableColumnTable";

const { Option } = Select;

class SyncerEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      syncerName: props.match.params.syncerName,
      syncer: null,
      organizations: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getSyncer();
    this.getOrganizations();
  }

  getSyncer() {
    SyncerBackend.getSyncer("admin", this.state.syncerName)
      .then((syncer) => {
        this.setState({
          syncer: syncer,
        });
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        this.setState({
          organizations: (res.msg === undefined) ? res : [],
        });
      });
  }

  parseSyncerField(key, value) {
    if (["port"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateSyncerField(key, value) {
    value = this.parseSyncerField(key, value);

    let syncer = this.state.syncer;
    syncer[key] = value;
    this.setState({
      syncer: syncer,
    });
  }

  renderSyncer() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("syncer:Edit Syncer")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitSyncerEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" onClick={() => this.submitSyncerEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </div>
      } style={(Setting.isMobile())? {margin: '5px'}:{}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.syncer.organization} onChange={(value => {this.updateSyncerField('organization', value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.syncer.name} onChange={e => {
              this.updateSyncerField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Type"), i18next.t("provider:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.syncer.type} onChange={(value => {this.updateSyncerField('type', value);})}>
              {
                ['Database', 'LDAP']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.syncer.host} onChange={e => {
              this.updateSyncerField('host', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Port"), i18next.t("provider:Port - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.syncer.port} onChange={value => {
              this.updateSyncerField('port', value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:User"), i18next.t("general:User - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.syncer.user} onChange={e => {
              this.updateSyncerField('user', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Password"), i18next.t("general:Password - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.syncer.password} onChange={e => {
              this.updateSyncerField('password', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Database"), i18next.t("syncer:Database - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.syncer.database} onChange={e => {
              this.updateSyncerField('database', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Table"), i18next.t("syncer:Table - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.syncer.table} onChange={e => {
              this.updateSyncerField('table', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Table columns"), i18next.t("syncer:Table columns - Tooltip"))} :
          </Col>
          <Col span={22} >
            <SyncerTableColumnTable
              title={i18next.t("syncer:Table columns")}
              table={this.state.syncer.tableColumns}
              onUpdateTable={(value) => { this.updateSyncerField('tableColumns', value)}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Affiliation table"), i18next.t("syncer:Affiliation table - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.syncer.affiliationTable} onChange={e => {
              this.updateSyncerField('affiliationTable', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Avatar base URL"), i18next.t("syncer:Avatar base URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.syncer.avatarBaseUrl} onChange={e => {
              this.updateSyncerField('avatarBaseUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Sync interval"), i18next.t("syncer:Sync interval - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.syncer.syncInterval} onChange={value => {
              this.updateSyncerField('syncInterval', value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("syncer:Is enabled"), i18next.t("syncer:Is enabled - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.syncer.isEnabled} onChange={checked => {
              this.updateSyncerField('isEnabled', checked);
            }} />
          </Col>
        </Row>
      </Card>
    )
  }

  submitSyncerEdit(willExist) {
    let syncer = Setting.deepCopy(this.state.syncer);
    SyncerBackend.updateSyncer(this.state.syncer.owner, this.state.syncerName, syncer)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            syncerName: this.state.syncer.name,
          });

          if (willExist) {
            this.props.history.push(`/syncers`);
          } else {
            this.props.history.push(`/syncers/${this.state.syncer.name}`);
          }
        } else {
          Setting.showMessage("error", res.msg);
          this.updateSyncerField('name', this.state.syncerName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.syncer !== null ? this.renderSyncer() : null
        }
        <div style={{marginTop: '20px', marginLeft: '40px'}}>
          <Button size="large" onClick={() => this.submitSyncerEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" size="large" onClick={() => this.submitSyncerEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </div>
      </div>
    );
  }
}

export default SyncerEditPage;
