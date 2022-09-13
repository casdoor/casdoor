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
import {Button, Card, Col, Input, InputNumber, Row, Select, Switch, Table, Tooltip} from "antd";
import * as AdapterBackend from "./backend/AdapterBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

import "codemirror/lib/codemirror.css";
import * as ModelBackend from "./backend/ModelBackend";
import {EditOutlined, MinusOutlined} from "@ant-design/icons";
require("codemirror/theme/material-darker.css");
require("codemirror/mode/javascript/javascript");

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
      models: [],
      policyLists: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getAdapter();
    this.getOrganizations();
  }

  getAdapter() {
    AdapterBackend.getAdapter(this.state.organizationName, this.state.adapterName)
      .then((adapter) => {
        this.setState({
          adapter: adapter,
        });

        this.getModels(adapter.owner);
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations(this.state.organizationName)
      .then((res) => {
        this.setState({
          organizations: (res.msg === undefined) ? res : [],
        });
      });
  }

  getModels(organizationName) {
    ModelBackend.getModels(organizationName)
      .then((res) => {
        this.setState({
          models: res,
        });
      });
  }

  parseAdapterField(key, value) {
    if (["port"].includes(key)) {
      value = Setting.myParseInt(value);
    }
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

  synPolicies() {
    this.setState({loading: true});
    AdapterBackend.syncPolicies(this.state.adapter.owner, this.state.adapter.name)
      .then((res) => {
        this.setState({loading: false, policyLists: res});
      })
      .catch(error => {
        this.setState({loading: false});
        Setting.showMessage("error", `Adapter failed to get policies: ${error}`);
      });
  }

  renderTable(table) {
    const columns = [
      {
        title: "Rule Type",
        dataIndex: "PType",
        key: "PType",
        width: "100px",
      },
      {
        title: "V0",
        dataIndex: "V0",
        key: "V0",
        width: "100px",
      },
      {
        title: "V1",
        dataIndex: "V1",
        key: "V1",
        width: "100px",
      },
      {
        title: "V2",
        dataIndex: "V2",
        key: "V2",
        width: "100px",
      },
      {
        title: "V3",
        dataIndex: "V3",
        key: "V3",
        width: "100px",
      },
      {
        title: "V4",
        dataIndex: "V4",
        key: "V4",
        width: "100px",
      },
      {
        title: "V5",
        dataIndex: "V5",
        key: "V5",
        width: "100px",
      },
      {
        title: "Option",
        key: "option",
        width: "100px",
        render: (text, record, index) => {
          return (
            <div>
              <Tooltip placement="topLeft" title="Edit">
                <Button style={{marginRight: "0.5rem"}} icon={<EditOutlined />} size="small" />
              </Tooltip>
              <Tooltip placement="topLeft" title="Delete">
                <Button icon={<MinusOutlined />} size="small" />
              </Tooltip>
            </div>
          );
        },
      }];

    return (
      <div>
        <Table
          pagination={{
            defaultPageSize: 10,
          }}
          columns={columns} dataSource={table} rowKey="name" size="middle" bordered
          loading={this.state.loading}
        />
      </div>
    );
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
            <Select virtual={false} style={{width: "100%"}} value={this.state.adapter.organization} onChange={(value => {this.updateadapterField("organization", value);})}>
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
            <Input value={this.state.adapter.name} onChange={e => {
              this.updateAdapterField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Type"), i18next.t("provider:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.adapter.type} onChange={(value => {
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
            <InputNumber value={this.state.adapter.port} onChange={value => {
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
            {Setting.getLabel(i18next.t("adapter:Database type"), i18next.t("adapter:Database type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.adapter.databaseType} onChange={(value => {this.updateAdapterField("databaseType", value);})}>
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
            {Setting.getLabel(i18next.t("adapter:Database"), i18next.t("adapter:Database - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.adapter.database} onChange={e => {
              this.updateAdapterField("database", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("adapter:Table"), i18next.t("adapter:Table - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.adapter.table}
              disabled={this.state.adapter.type === "Keycloak"} onChange={e => {
                this.updateAdapterField("table", e.target.value);
              }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("adapter:Model"), i18next.t("adapter:Model - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.adapter.model} onChange={(model => {
              this.updateAdapterField("model", model);
            })}>
              {
                this.state.models.map((model, index) => <Option key={index} value={model.name}>{model.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("adapter:Policies"), i18next.t("adapter:Policies - Tooltip"))} :
          </Col>
          <Col span={2}>
            <Button type="primary" onClick={() => {this.synPolicies();}}>
              {i18next.t("adapter:Sync")}
            </Button>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          </Col>
          <Col span={22} >
            {
              this.renderTable(this.state.policyLists)
            }
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("general:Is enabled"), i18next.t("general:Is enabled - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.adapter.isEnabled} onChange={checked => {
              this.updateAdapterField("isEnabled", checked);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitAdapterEdit(willExist) {
    const adapter = Setting.deepCopy(this.state.adapter);
    AdapterBackend.updateAdapter(this.state.adapter.owner, this.state.adapterName, adapter)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", "Successfully saved");
          this.setState({
            adapterName: this.state.adapter.name,
          });

          if (willExist) {
            this.props.history.push("/adapters");
          } else {
            this.props.history.push(`/adapters/${this.state.adapter.name}`);
          }
        } else {
          Setting.showMessage("error", res.msg);
          this.updateAdapterField("name", this.state.adapterName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  deleteAdapter() {
    AdapterBackend.deleteAdapter(this.state.adapter)
      .then(() => {
        this.props.history.push("/adapters");
      })
      .catch(error => {
        Setting.showMessage("error", `adapter failed to delete: ${error}`);
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
