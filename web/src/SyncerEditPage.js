// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
import {Button, Card, Col, Input, InputNumber, Radio, Row, Select, Switch} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as SyncerBackend from "./backend/SyncerBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import SyncerTableColumnTable from "./table/SyncerTableColumnTable";

import * as CertBackend from "./backend/CertBackend";
import Editor from "./common/Editor";

const {Option} = Select;

class SyncerEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      certs: [],
      classes: props,
      syncerName: props.match.params.syncerName,
      syncer: null,
      organizations: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
      testDbLoading: false,
    };
  }

  UNSAFE_componentWillMount() {
    this.getSyncer();
    this.getOrganizations();
  }

  getSyncer() {
    SyncerBackend.getSyncer("admin", this.state.syncerName)
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
          syncer: res.data,
        });

        if (res.data && res.data.organization) {
          this.getCerts(res.data.organization);
        }
      });
  }

  getCerts(owner) {
    // Load certificates for the given organization
    CertBackend.getCerts(owner)
      .then((res) => {
        this.setState({
          certs: res.data || [],
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

  parseSyncerField(key, value) {
    if (["port"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateSyncerField(key, value) {
    value = this.parseSyncerField(key, value);

    const syncer = this.state.syncer;
    if (key === "organization" && syncer["organization"] !== value) {
      // the syncer changed the organization, reset the cert and reload certs
      syncer["cert"] = "";
      this.getCerts(value);
    }

    syncer[key] = value;
    this.setState({
      syncer: syncer,
    });
  }

  getSyncerTableColumns(syncer) {
    switch (syncer.type) {
    case "Keycloak":
      return [
        {
          "name": "ID",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [

          ],
        },
        {
          "name": "USERNAME",
          "type": "string",
          "casdoorName": "Name",
          "isHashed": true,
          "values": [

          ],
        },
        {
          "name": "LAST_NAME+FIRST_NAME",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [

          ],
        },
        {
          "name": "EMAIL",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [

          ],
        },
        {
          "name": "EMAIL_VERIFIED",
          "type": "boolean",
          "casdoorName": "EmailVerified",
          "isHashed": true,
          "values": [

          ],
        },
        {
          "name": "FIRST_NAME",
          "type": "string",
          "casdoorName": "FirstName",
          "isHashed": true,
          "values": [

          ],
        },
        {
          "name": "LAST_NAME",
          "type": "string",
          "casdoorName": "LastName",
          "isHashed": true,
          "values": [

          ],
        },
        {
          "name": "CREATED_TIMESTAMP",
          "type": "string",
          "casdoorName": "CreatedTime",
          "isHashed": true,
          "values": [

          ],
        },
        {
          "name": "ENABLED",
          "type": "boolean",
          "casdoorName": "IsForbidden",
          "isHashed": true,
          "values": [

          ],
        },
      ];
    case "WeCom":
      return [
        {
          "name": "userid",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "name",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "email",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "mobile",
          "type": "string",
          "casdoorName": "Phone",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "avatar",
          "type": "string",
          "casdoorName": "Avatar",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "position",
          "type": "string",
          "casdoorName": "Title",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "gender",
          "type": "string",
          "casdoorName": "Gender",
          "isHashed": true,
          "values": [],
        },
      ];
    case "Azure AD":
      return [
        {
          "name": "id",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "userPrincipalName",
          "type": "string",
          "casdoorName": "Name",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "displayName",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "givenName",
          "type": "string",
          "casdoorName": "FirstName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "surname",
          "type": "string",
          "casdoorName": "LastName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "mail",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "mobilePhone",
          "type": "string",
          "casdoorName": "Phone",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "jobTitle",
          "type": "string",
          "casdoorName": "Title",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "officeLocation",
          "type": "string",
          "casdoorName": "Location",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "preferredLanguage",
          "type": "string",
          "casdoorName": "Language",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "accountEnabled",
          "type": "boolean",
          "casdoorName": "IsForbidden",
          "isHashed": true,
          "values": [],
        },
      ];
    case "Google Workspace":
      return [
        {
          "name": "id",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "primaryEmail",
          "type": "string",
          "casdoorName": "Name",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "name.fullName",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "name.givenName",
          "type": "string",
          "casdoorName": "FirstName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "name.familyName",
          "type": "string",
          "casdoorName": "LastName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "suspended",
          "type": "boolean",
          "casdoorName": "IsForbidden",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "isAdmin",
          "type": "boolean",
          "casdoorName": "IsAdmin",
          "isHashed": true,
          "values": [],
        },
      ];
    case "DingTalk":
      return [
        {
          "name": "userid",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "name",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "email",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "mobile",
          "type": "string",
          "casdoorName": "Phone",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "avatar",
          "type": "string",
          "casdoorName": "Avatar",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "title",
          "type": "string",
          "casdoorName": "Title",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "job_number",
          "type": "string",
          "casdoorName": "Name",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "active",
          "type": "boolean",
          "casdoorName": "IsForbidden",
          "isHashed": true,
          "values": [],
        },
      ];
    case "Active Directory":
      return [
        {
          "name": "objectGUID",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "sAMAccountName",
          "type": "string",
          "casdoorName": "Name",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "displayName",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "givenName",
          "type": "string",
          "casdoorName": "FirstName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "sn",
          "type": "string",
          "casdoorName": "LastName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "mail",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "mobile",
          "type": "string",
          "casdoorName": "Phone",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "title",
          "type": "string",
          "casdoorName": "Title",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "department",
          "type": "string",
          "casdoorName": "Affiliation",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "userAccountControl",
          "type": "string",
          "casdoorName": "IsForbidden",
          "isHashed": true,
          "values": [],
        },
      ];
    case "Lark":
      return [
        {
          "name": "user_id",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "name",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "email",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "mobile",
          "type": "string",
          "casdoorName": "Phone",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "avatar",
          "type": "string",
          "casdoorName": "Avatar",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "job_title",
          "type": "string",
          "casdoorName": "Title",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "gender",
          "type": "number",
          "casdoorName": "Gender",
          "isHashed": true,
          "values": [],
        },
      ];
    case "Okta":
      return [
        {
          "name": "id",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "profile.login",
          "type": "string",
          "casdoorName": "Name",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "profile.displayName",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "profile.firstName",
          "type": "string",
          "casdoorName": "FirstName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "profile.lastName",
          "type": "string",
          "casdoorName": "LastName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "profile.email",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "profile.mobilePhone",
          "type": "string",
          "casdoorName": "Phone",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "profile.title",
          "type": "string",
          "casdoorName": "Title",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "profile.preferredLanguage",
          "type": "string",
          "casdoorName": "Language",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "status",
          "type": "string",
          "casdoorName": "IsForbidden",
          "isHashed": true,
          "values": [],
        },
      ];
    case "SCIM":
      return [
        {
          "name": "id",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "userName",
          "type": "string",
          "casdoorName": "Name",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "displayName",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "name.givenName",
          "type": "string",
          "casdoorName": "FirstName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "name.familyName",
          "type": "string",
          "casdoorName": "LastName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "emails",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "phoneNumbers",
          "type": "string",
          "casdoorName": "Phone",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "title",
          "type": "string",
          "casdoorName": "Title",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "preferredLanguage",
          "type": "string",
          "casdoorName": "Language",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "active",
          "type": "boolean",
          "casdoorName": "IsForbidden",
          "isHashed": true,
          "values": [],
        },
      ];
    case "AWS IAM":
      return [
        {
          "name": "UserId",
          "type": "string",
          "casdoorName": "Id",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "UserName",
          "type": "string",
          "casdoorName": "Name",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "UserName",
          "type": "string",
          "casdoorName": "DisplayName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "Tags.Email",
          "type": "string",
          "casdoorName": "Email",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "Tags.Phone",
          "type": "string",
          "casdoorName": "Phone",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "Tags.FirstName",
          "type": "string",
          "casdoorName": "FirstName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "Tags.LastName",
          "type": "string",
          "casdoorName": "LastName",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "Tags.Title",
          "type": "string",
          "casdoorName": "Title",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "Tags.Department",
          "type": "string",
          "casdoorName": "Affiliation",
          "isHashed": true,
          "values": [],
        },
        {
          "name": "CreateDate",
          "type": "string",
          "casdoorName": "CreatedTime",
          "isHashed": true,
          "values": [],
        },
      ];
    default:
      return [];
    }
  }

  needSshfields() {
    return this.state.syncer.type === "Database" && (this.state.syncer.databaseType === "mysql" || this.state.syncer.databaseType === "mssql" || this.state.syncer.databaseType === "postgres");
  }

  renderSyncer() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("syncer:New Syncer") : i18next.t("syncer:Edit Syncer")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitSyncerEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitSyncerEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteSyncer()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.syncer.organization} onChange={(value => {this.updateSyncerField("organization", value);})}>
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
            <Input value={this.state.syncer.name} onChange={e => {
              this.updateSyncerField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Type"), i18next.t("general:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.syncer.type} onChange={(value => {
              this.updateSyncerField("type", value);
              const syncer = this.state.syncer;
              syncer["tableColumns"] = this.getSyncerTableColumns(this.state.syncer);
              syncer.table = (value === "Keycloak") ? "user_entity" : this.state.syncer.table;
              this.setState({
                syncer: syncer,
              });
            })}>
              {
                ["Database", "Keycloak", "WeCom", "Azure AD", "Active Directory", "Google Workspace", "DingTalk", "Lark", "Okta", "SCIM", "AWS IAM"]
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        {
          this.state.syncer.type === "WeCom" || this.state.syncer.type === "Azure AD" || this.state.syncer.type === "Active Directory" || this.state.syncer.type === "Google Workspace" || this.state.syncer.type === "DingTalk" || this.state.syncer.type === "Lark" || this.state.syncer.type === "Okta" || this.state.syncer.type === "SCIM" || this.state.syncer.type === "AWS IAM" ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("syncer:Database type"), i18next.t("syncer:Database type - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Select virtual={false} style={{width: "100%"}} value={this.state.syncer.databaseType} onChange={(value => {
                  this.updateSyncerField("databaseType", value);
                  if (value === "postgres") {
                    this.updateSyncerField("sslMode", "disable");
                  } else {
                    this.updateSyncerField("sslMode", "");
                  }
                })}>
                  {
                    [
                      {id: "mysql", name: "MySQL"},
                      {id: "postgres", name: "PostgreSQL"},
                      {id: "mssql", name: "SQL Server"},
                      {id: "oracle", name: "Oracle"},
                      {id: "sqlite3", name: "Sqlite 3"},
                    ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
                  }
                </Select>
              </Col>
            </Row>
          )
        }
        {
          this.state.syncer.databaseType !== "postgres" ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("provider:SSL mode"), i18next.t("provider:SSL mode - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Select virtual={false} style={{width: "100%"}} value={this.state.syncer.sslMode} onChange={(value => {this.updateSyncerField("sslMode", value);})}>
                  {
                    [
                      {id: "disable", name: "disable"},
                      // {id: "allow", name: "allow"},
                      // {id: "prefer", name: "prefer"},
                      {id: "require", name: "require"},
                      {id: "verify-ca", name: "verify-ca"},
                      {id: "verify-full", name: "verify-full"},
                    ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
                  }
                </Select>
              </Col>
            </Row>
          )
        }
        {
          this.state.syncer.type === "WeCom" || this.state.syncer.type === "DingTalk" || this.state.syncer.type === "Lark" ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(this.state.syncer.type === "Azure AD" ? i18next.t("provider:Tenant ID") : this.state.syncer.type === "Google Workspace" ? i18next.t("syncer:Admin Email") : this.state.syncer.type === "Active Directory" ? i18next.t("ldap:Server") : this.state.syncer.type === "SCIM" ? i18next.t("syncer:SCIM Server URL") : this.state.syncer.type === "AWS IAM" ? i18next.t("syncer:AWS Region") : i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input prefix={<LinkOutlined />} value={this.state.syncer.host} onChange={e => {
                  this.updateSyncerField("host", e.target.value);
                }} />
              </Col>
            </Row>
          )
        }
        {
          this.state.syncer.type === "WeCom" || this.state.syncer.type === "Azure AD" || this.state.syncer.type === "Google Workspace" || this.state.syncer.type === "DingTalk" || this.state.syncer.type === "Lark" || this.state.syncer.type === "Okta" || this.state.syncer.type === "SCIM" || this.state.syncer.type === "AWS IAM" ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(this.state.syncer.type === "Active Directory" ? i18next.t("provider:LDAP port") : i18next.t("provider:Port"), i18next.t("provider:Port - Tooltip"))} :
              </Col>
              <Col span={22} >
                <InputNumber value={this.state.syncer.port} onChange={value => {
                  this.updateSyncerField("port", value);
                }} />
              </Col>
            </Row>
          )
        }
        {
          this.state.syncer.type === "Google Workspace" ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(
                  this.state.syncer.type === "WeCom" ? i18next.t("syncer:Corp ID") :
                    this.state.syncer.type === "DingTalk" ? i18next.t("provider:App Key") :
                      this.state.syncer.type === "Lark" ? i18next.t("provider:App ID") :
                        this.state.syncer.type === "Azure AD" ? i18next.t("provider:Client ID") :
                          this.state.syncer.type === "Active Directory" ? i18next.t("syncer:Bind DN") :
                            this.state.syncer.type === "SCIM" ? i18next.t("syncer:Username (optional)") :
                              this.state.syncer.type === "AWS IAM" ? i18next.t("syncer:AWS Access Key ID") :
                                i18next.t("general:User"),
                  i18next.t("general:User - Tooltip")
                )} :
              </Col>
              <Col span={22} >
                <Input value={this.state.syncer.user} onChange={e => {
                  this.updateSyncerField("user", e.target.value);
                }} />
              </Col>
            </Row>
          )
        }
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(
              this.state.syncer.type === "WeCom" ? i18next.t("syncer:Corp secret") :
                this.state.syncer.type === "DingTalk" ? i18next.t("provider:App secret") :
                  this.state.syncer.type === "Lark" ? i18next.t("provider:App secret") :
                    this.state.syncer.type === "Azure AD" ? i18next.t("provider:Client secret") :
                      this.state.syncer.type === "Google Workspace" ? i18next.t("syncer:Service account key") :
                        this.state.syncer.type === "SCIM" ? i18next.t("syncer:API Token / Password") :
                          this.state.syncer.type === "AWS IAM" ? i18next.t("syncer:AWS Secret Access Key") :
                            i18next.t("general:Password"),
              i18next.t("general:Password - Tooltip")
            )} :
          </Col>
          <Col span={22} >
            {
              this.state.syncer.type === "Google Workspace" ? (
                <Input.TextArea rows={4} value={this.state.syncer.password} onChange={e => {
                  this.updateSyncerField("password", e.target.value);
                }} placeholder={i18next.t("syncer:Paste your Google Workspace service account JSON key here")} />
              ) : (
                <Input.Password value={this.state.syncer.password} onChange={e => {
                  this.updateSyncerField("password", e.target.value);
                }} />
              )
            }
          </Col>
        </Row>
        {
          this.state.syncer.type === "WeCom" || this.state.syncer.type === "Azure AD" || this.state.syncer.type === "Google Workspace" || this.state.syncer.type === "DingTalk" || this.state.syncer.type === "Lark" || this.state.syncer.type === "Okta" || this.state.syncer.type === "SCIM" || this.state.syncer.type === "AWS IAM" ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(this.state.syncer.type === "Active Directory" ? i18next.t("ldap:Base DN") : i18next.t("syncer:Database"), i18next.t("syncer:Database - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input value={this.state.syncer.database} onChange={e => {
                  this.updateSyncerField("database", e.target.value);
                }} />
              </Col>
            </Row>
          )
        }
        {
          this.needSshfields() ? (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("general:SSH type"), i18next.t("general:SSH type - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Radio.Group value={this.state.syncer.sshType} buttonStyle="solid" onChange={e => {
                  this.updateSyncerField("sshType", e.target.value);
                }}>
                  <Radio.Button value="">{i18next.t("general:None")}</Radio.Button>
                  <Radio.Button value="password">{i18next.t("general:Password")}</Radio.Button>
                  <Radio.Button value="cert">{i18next.t("general:Cert")}</Radio.Button>
                </Radio.Group>
              </Col>
            </Row>
          ) : null
        }
        {
          this.state.syncer.sshType && this.needSshfields() ? (
            <React.Fragment>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("syncer:SSH host"), i18next.t("provider:Host - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input prefix={<LinkOutlined />} value={this.state.syncer.sshHost} onChange={e => {
                    this.updateSyncerField("sshHost", e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("syncer:SSH port"), i18next.t("provider:Port - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <InputNumber value={this.state.syncer.sshPort} onChange={value => {
                    this.updateSyncerField("sshPort", value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("syncer:SSH user"), i18next.t("general:User - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.syncer.sshUser} onChange={e => {
                    this.updateSyncerField("sshUser", e.target.value);
                  }} />
                </Col>
              </Row>
              {
                this.state.syncer.sshType === "password" && this.needSshfields() ?
                  (
                    <Row style={{marginTop: "20px"}} >
                      <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                        {Setting.getLabel(i18next.t("syncer:SSH password"), i18next.t("general:Password - Tooltip"))} :
                      </Col>
                      <Col span={22} >
                        <Input.Password value={this.state.syncer.sshPassword} onChange={e => {
                          this.updateSyncerField("ssh " + "sshPassword", e.target.value);
                        }} />
                      </Col>
                    </Row>
                  ) : (
                    <Row style={{marginTop: "20px"}} >
                      <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                        {Setting.getLabel(i18next.t("general:SSH cert"), i18next.t("general:Cert - Tooltip"))} :
                      </Col>
                      <Col span={22} >
                        <Select virtual={false} style={{width: "100%"}} value={this.state.syncer.cert} onChange={(value => {this.updateSyncerField("cert", value);})}>
                          {
                            this.state?.certs.map((cert, index) => <Option key={index} value={cert.name}>{cert.name}</Option>)
                          }
                        </Select>
                      </Col>
                    </Row>
                  )
              }
            </React.Fragment>
          ) : null
        }
        {
          this.state.syncer.type === "WeCom" || this.state.syncer.type === "Azure AD" || this.state.syncer.type === "Google Workspace" || this.state.syncer.type === "DingTalk" || this.state.syncer.type === "Lark" || this.state.syncer.type === "Okta" || this.state.syncer.type === "SCIM" || this.state.syncer.type === "AWS IAM" ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("syncer:Table"), i18next.t("syncer:Table - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input value={this.state.syncer.table} onChange={e => {
                  this.updateSyncerField("table", e.target.value);
                }} />
              </Col>
            </Row>
          )
        }
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Syncer test"), i18next.t("provider:Syncer test - Tooltip"))} :
          </Col>
          <Col span={2} >
            <Button type={"primary"} loading={this.state.testDbLoading} onClick={() => {
              this.setState({testDbLoading: true});
              SyncerBackend.testSyncerDb(this.state.syncer)
                .then((res) => {
                  if (res.status === "ok") {
                    this.setState({testDbLoading: false});
                    Setting.showMessage("success", i18next.t("syncer:Connect successfully"));
                  } else {
                    this.setState({testDbLoading: false});
                    Setting.showMessage("error", `${i18next.t("syncer:Failed to connect")}: ${res.msg}`);
                  }
                })
                .catch(error => {
                  this.setState({testDbLoading: false});
                  Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
                });
            }
            }>{i18next.t("syncer:Test Connection")}</Button>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Table columns"), i18next.t("syncer:Table columns - Tooltip"))} :
          </Col>
          <Col span={22} >
            <SyncerTableColumnTable
              title={i18next.t("syncer:Table columns")}
              table={this.state.syncer.tableColumns}
              onUpdateTable={(value) => {this.updateSyncerField("tableColumns", value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Affiliation table"), i18next.t("syncer:Affiliation table - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.syncer.affiliationTable} onChange={e => {
              this.updateSyncerField("affiliationTable", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Avatar base URL"), i18next.t("syncer:Avatar base URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.syncer.avatarBaseUrl} onChange={e => {
              this.updateSyncerField("avatarBaseUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Sync interval"), i18next.t("syncer:Sync interval - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber value={this.state.syncer.syncInterval} onChange={value => {
              this.updateSyncerField("syncInterval", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("syncer:Error text"), i18next.t("syncer:Error text - Tooltip"))} :
          </Col>
          <Col span={22} >
            <div style={{width: "100%", height: "300px"}} >
              <Editor
                value={this.state.syncer.errorText}
                fillHeight
                readOnly
                dark
                lang="js"
                onChange={value => {
                  this.updateSyncerField("errorText", value);
                }}
              />
            </div>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("syncer:Is read-only"), i18next.t("syncer:Is read-only - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.syncer.isReadOnly} onChange={checked => {
              this.updateSyncerField("isReadOnly", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("general:Is enabled"), i18next.t("general:Is enabled - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.syncer.isEnabled} onChange={checked => {
              this.updateSyncerField("isEnabled", checked);
            }} />
          </Col>
        </Row>
      </Card>
    );
  }

  submitSyncerEdit(exitAfterSave) {
    const syncer = Setting.deepCopy(this.state.syncer);
    SyncerBackend.updateSyncer(this.state.syncer.owner, this.state.syncerName, syncer)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            syncerName: this.state.syncer.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/syncers");
          } else {
            this.props.history.push(`/syncers/${this.state.syncer.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateSyncerField("name", this.state.syncerName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteSyncer() {
    SyncerBackend.deleteSyncer(this.state.syncer)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/syncers");
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
          this.state.syncer !== null ? this.renderSyncer() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitSyncerEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitSyncerEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteSyncer()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default SyncerEditPage;
