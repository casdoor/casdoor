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

import React from "react";
import {Button, Card, Col, Input, InputNumber, Row, Select, Switch} from "antd";
import * as AdapterBackend from "./backend/AdapterBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const {Option} = Select;

class AdapterEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      adapterName: props.match.params.adapterName,
      adapter: null,
      organizations: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getAdapter();
    this.getOrganizations();
  }

  getAdapter() {
    AdapterBackend.getAdapter(this.state.organizationName, this.state.adapterName)
      .then((res) => {
        if (res.status === "ok") {
          if (res.data === null) {
            this.props.history.push("/404");
            return;
          }

          this.setState({
            adapter: res.data,
          });
        }
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

  parseAdapterField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateAdapterField(key, value) {
    value = this.parseAdapterField(key, value);

    const adapter = this.state.adapter;
    adapter[key] = value;
    this.setState({
      adapter: adapter,
    });
  }

  renderAdapter() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("adapter:New Adapter") : i18next.t("adapter:Edit Adapter")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitAdapterEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitAdapterEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteAdapter()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account) || Setting.builtInObject(this.state.adapter)} value={this.state.adapter.owner} onChange={(value => {
              this.updateAdapterField("owner", value);
            })}>
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
            <Input disabled={Setting.builtInObject(this.state.adapter)} value={this.state.adapter.name} onChange={e => {
              this.updateAdapterField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Table"), i18next.t("syncer:Table - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.adapter.table}
              disabled={Setting.builtInObject(this.state.adapter)} onChange={e => {
                this.updateAdapterField("table", e.target.value);
              }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("adapter:Use same DB"), i18next.t("adapter:Use same DB - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch disabled={Setting.builtInObject(this.state.adapter)} checked={this.state.adapter.useSameDb || Setting.builtInObject(this.state.adapter)} onChange={checked => {
              this.updateAdapterField("useSameDb", checked);
              if (checked) {
                this.updateAdapterField("type", "");
                this.updateAdapterField("databaseType", "");
                this.updateAdapterField("host", "");
                this.updateAdapterField("port", 0);
                this.updateAdapterField("user", "");
                this.updateAdapterField("password", "");
                this.updateAdapterField("database", "");
              } else {
                this.updateAdapterField("type", "Database");
                this.updateAdapterField("databaseType", "mysql");
                this.updateAdapterField("host", "localhost");
                this.updateAdapterField("port", 3306);
                this.updateAdapterField("user", "root");
                this.updateAdapterField("password", "123456");
                this.updateAdapterField("database", "dbName");
              }
            }} />
          </Col>
        </Row>
        {
          (this.state.adapter.useSameDb || Setting.builtInObject(this.state.adapter)) ? null : (
            <React.Fragment>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Type"), i18next.t("provider:Type - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Select virtual={false} disabled={Setting.builtInObject(this.state.adapter)} style={{width: "100%"}} value={this.state.adapter.type} onChange={(value => {
                    this.updateAdapterField("type", value);
                    const adapter = this.state.adapter;
                    // adapter["tableColumns"] = Setting.getAdapterTableColumns(this.state.adapter);
                    this.setState({
                      adapter: adapter,
                    });
                  })}>
                    {
                      ["Database"]
                        .map((item, index) => <Option key={index} value={item}>{item}</Option>)
                    }
                  </Select>
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("syncer:Database type"), i18next.t("syncer:Database type - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Select virtual={false} disabled={Setting.builtInObject(this.state.adapter)} style={{width: "100%"}} value={this.state.adapter.databaseType} onChange={(value => {this.updateAdapterField("databaseType", value);})}>
                    {
                      [
                        {id: "mysql", name: "MySQL"},
                        {id: "postgres", name: "PostgreSQL"},
                        {id: "mssql", name: "SQL Server"},
                        {id: "oracle", name: "Oracle"},
                        {id: "sqlite3", name: "Sqlite 3"},
                      ].map((databaseType, index) => <Option key={index} value={databaseType.id}>{databaseType.name}</Option>)
                    }
                  </Select>
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.adapter.host} onChange={e => {
                    this.updateAdapterField("host", e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Port"), i18next.t("provider:Port - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <InputNumber value={this.state.adapter.port} min={0} max={65535} onChange={value => {
                    this.updateAdapterField("port", value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("general:User"), i18next.t("general:User - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.adapter.user} onChange={e => {
                    this.updateAdapterField("user", e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("general:Password"), i18next.t("general:Password - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.adapter.password} onChange={e => {
                    this.updateAdapterField("password", e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("syncer:Database"), i18next.t("syncer:Database - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input disabled={Setting.builtInObject(this.state.adapter)} value={this.state.adapter.database} onChange={e => {
                    this.updateAdapterField("database", e.target.value);
                  }} />
                </Col>
              </Row>
            </React.Fragment>
          )
        }
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:DB test"), i18next.t("provider:DB test - Tooltip"))} :
          </Col>
          <Col span={2} >
            <Button disabled={this.state.organizationName !== this.state.adapter.owner} type={"primary"} onClick={() => {
              AdapterBackend.getPolicies("", "", `${this.state.adapter.owner}/${this.state.adapter.name}`)
                .then((res) => {
                  if (res.status === "ok") {
                    Setting.showMessage("success", i18next.t("syncer:Connect successfully"));
                  } else {
                    Setting.showMessage("error", i18next.t("syncer:Failed to connect") + ": " + res.msg);
                  }
                })
                .catch(error => {
                  Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
                });
            }
            }>{i18next.t("syncer:Test DB Connection")}</Button>
          </Col>
        </Row>
      </Card>
    );
  }

  submitAdapterEdit(exitAfterSave) {
    const adapter = Setting.deepCopy(this.state.adapter);
    AdapterBackend.updateAdapter(this.state.organizationName, this.state.adapterName, adapter)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            organizationName: this.state.adapter.owner,
            adapterName: this.state.adapter.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/adapters");
          } else {
            this.props.history.push(`/adapters/${this.state.adapter.owner}/${this.state.adapter.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateAdapterField("name", this.state.adapterName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteAdapter() {
    AdapterBackend.deleteAdapter(this.state.adapter)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/adapters");
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
          this.state.adapter !== null ? this.renderAdapter() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitAdapterEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitAdapterEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteAdapter()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default AdapterEditPage;
