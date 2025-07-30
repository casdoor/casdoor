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
import {Button, Card, Col, ConfigProvider, Input, InputNumber, Popover, Radio, Result, Row, Select, Space, Switch, Upload, message} from "antd";
import {CopyOutlined, HolderOutlined, LinkOutlined, UploadOutlined, UsergroupAddOutlined} from "@ant-design/icons";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as CertBackend from "./backend/CertBackend";
import * as Setting from "./Setting";
import * as Conf from "./Conf";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ResourceBackend from "./backend/ResourceBackend";
import SignupPage from "./auth/SignupPage";
import LoginPage from "./auth/LoginPage";
import i18next from "i18next";
import UrlTable from "./table/UrlTable";
import ProviderTable from "./table/ProviderTable";
import SigninMethodTable from "./table/SigninMethodTable";
import SignupTable from "./table/SignupTable";
import SamlAttributeTable from "./table/SamlAttributeTable";
import PromptPage from "./auth/PromptPage";
import copy from "copy-to-clipboard";
import ThemeEditor from "./common/theme/ThemeEditor";

import SigninTable from "./table/SigninTable";
import Editor from "./common/Editor";
import * as GroupBackend from "./backend/GroupBackend";

const {Option} = Select;

const template = `<style>
  .login-panel {
    padding: 40px 70px 0 70px;
    border-radius: 10px;
    background-color: #ffffff;
    box-shadow: 0 0 30px 20px rgba(0, 0, 0, 0.20);
  }
  .login-panel-dark {
    padding: 40px 70px 0 70px;
    border-radius: 10px;
    background-color: #333333;
    box-shadow: 0 0 30px 20px rgba(255, 255, 255, 0.20);
  }
  .forget-content {
    padding: 10px 100px 20px;
    margin: 30px auto;
    border: 2px solid #fff;
    border-radius: 7px;
    background-color: rgb(255 255 255);
    box-shadow: 0 0 20px rgb(0 0 0 / 20%);
  }
</style>`;

const previewGrid = Setting.isMobile() ? 22 : 11;
const previewWidth = Setting.isMobile() ? "110%" : "90%";

const sideTemplate = `<style>
  .left-model{
    text-align: center;
    padding: 30px;
    background-color: #8ca0ed;
    position: absolute;
    transform: none;
    width: 100%;
    height: 100%;
  }
  .side-logo{
    display: flex;
    align-items: center;
  }
  .side-logo span {
    font-family: Montserrat, sans-serif;
    font-weight: 900;
    font-size: 2.4rem;
    line-height: 1.3;
    margin-left: 16px;
    color: #404040;
  }
  .img{
    max-width: none;
    margin: 41px 0 13px;
  }
</style>
<div class="left-model">
  <span class="side-logo"> <img src="${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png" alt="Casdoor" style="width: 120px"> 
    <span>SSO</span> 
  </span>
  <div class="img">
    <img src="${Setting.StaticBaseUrl}/img/casbin.svg" alt="Casdoor"/>
  </div>
</div>
`;

class ApplicationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      owner: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      applicationName: props.match.params.applicationName,
      application: null,
      organizations: [],
      certs: [],
      providers: [],
      uploading: false,
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
      samlAttributes: [],
      samlMetadata: null,
      isAuthorized: true,
    };
  }

  UNSAFE_componentWillMount() {
    this.getApplication();
    this.getOrganizations();
    this.getGroups();
  }

  getApplication() {
    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        const application = res.data;
        if (application.grantTypes === null || application.grantTypes === undefined || application.grantTypes.length === 0) {
          application.grantTypes = ["authorization_code"];
        }

        if (application.tags === null || application.tags === undefined) {
          application.tags = [];
        }

        this.setState({
          application: application,
        });

        this.getProviders(application);

        this.getCerts(application);

        this.getSamlMetadata(application.enableSamlPostBinding);
      });
  }

  getOrganizations() {
    OrganizationBackend.getOrganizations("admin")
      .then((res) => {
        if (res.status === "error") {
          this.setState({
            isAuthorized: false,
          });
        } else {
          this.setState({
            organizations: res.data || [],
          });
        }
      });
  }

  getGroups() {
    GroupBackend.getGroups(this.state.organizationName)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            groups: res.data,
          });
        }
      });
  }

  getCerts(application) {
    let owner = application.organization;
    if (application.isShared) {
      owner = this.props.owner;
    }
    CertBackend.getCerts(owner)
      .then((res) => {
        this.setState({
          certs: res.data || [],
        });
      });
  }

  getProviders(application) {
    let owner = application.organization;
    if (application.isShared) {
      owner = this.props.account.owner;
    }
    ProviderBackend.getProviders(owner)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            providers: res.data,
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  getSamlMetadata(checked) {
    ApplicationBackend.getSamlMetadata("admin", this.state.applicationName, checked)
      .then((data) => {
        this.setState({
          samlMetadata: data,
        });
      });
  }

  parseApplicationField(key, value) {
    if (["offset"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateApplicationField(key, value) {
    value = this.parseApplicationField(key, value);
    const application = this.state.application;
    application[key] = value;
    this.setState({
      application: application,
    });
  }

  handleUpload(info) {
    if (info.file.type !== "text/html") {
      Setting.showMessage("error", i18next.t("application:Please select a HTML file"));
      return;
    }
    this.setState({uploading: true});
    const fullFilePath = `termsOfUse/${this.state.application.owner}/${this.state.application.name}.html`;
    ResourceBackend.uploadResource(this.props.account.owner, this.props.account.name, "termsOfUse", "ApplicationEditPage", fullFilePath, info.file)
      .then(res => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("application:File uploaded successfully"));
          this.updateApplicationField("termsOfUse", res.data);
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
        }
      }).finally(() => {
        this.setState({uploading: false});
      });
  }

  renderApplication() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("application:New Application") : i18next.t("application:Edit Application")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitApplicationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitApplicationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteApplication()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.name} disabled={this.state.application.name === "app-built-in"} onChange={e => {
              const value = e.target.value;
              if (/[/?:@#&%=+;]/.test(value)) {
                const invalidChars = "/ ? : @ # & % = + ;";
                const messageText = i18next.t("application:Invalid characters in application name") + ":" + " " + invalidChars;
                message.error(messageText);
                return;
              }
              this.updateApplicationField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.displayName} onChange={e => {
              this.updateApplicationField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Is shared"), i18next.t("general:Is shared - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Switch disabled={Setting.isAdminUser()} checked={this.state.application.isShared} onChange={checked => {
              this.updateApplicationField("isShared", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Logo"), i18next.t("general:Logo - Tooltip"))} :
          </Col>
          <Col span={22} style={(Setting.isMobile()) ? {maxWidth: "100%"} : {}}>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined />} value={this.state.application.logo} onChange={e => {
                  this.updateApplicationField("logo", e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" rel="noreferrer" href={this.state.application.logo}>
                  <img src={this.state.application.logo} alt={this.state.application.logo} height={90} style={{marginBottom: "20px"}} />
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Home"), i18next.t("general:Home - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.application.homepageUrl} onChange={e => {
              this.updateApplicationField("homepageUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.description} onChange={e => {
              this.updateApplicationField("description", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.application.organization} onChange={(value => {this.updateApplicationField("organization", value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("organization:Tags"), i18next.t("application:Tags - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: "100%"}} value={this.state.application.tags} onChange={(value => {this.updateApplicationField("tags", value);})}>
              {
                this.state.application.tags?.map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.clientId} onChange={e => {
              this.updateApplicationField("clientId", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.clientSecret} onChange={e => {
              this.updateApplicationField("clientSecret", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Cert"), i18next.t("general:Cert - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.application.cert} onChange={(value => {this.updateApplicationField("cert", value);})}>
              {
                this.state.certs.map((cert, index) => <Option key={index} value={cert.name}>{cert.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Redirect URLs"), i18next.t("application:Redirect URLs - Tooltip"))} :
          </Col>
          <Col span={22} >
            <UrlTable
              title={i18next.t("application:Redirect URLs")}
              table={this.state.application.redirectUris}
              onUpdateTable={(value) => {this.updateApplicationField("redirectUris", value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Forced redirect origin"), i18next.t("general:Forced redirect origin - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.application.forcedRedirectOrigin} onChange={e => {
              this.updateApplicationField("forcedRedirectOrigin", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Token format"), i18next.t("application:Token format - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.application.tokenFormat} onChange={(value => {this.updateApplicationField("tokenFormat", value);})}
              options={["JWT", "JWT-Empty", "JWT-Custom", "JWT-Standard"].map((item) => Setting.getOption(item, item))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Token signing method"), i18next.t("application:Token signing method - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.application.tokenSigningMethod === "" ? "RS256" : this.state.application.tokenSigningMethod} onChange={(value => {this.updateApplicationField("tokenSigningMethod", value);})}
              options={["RS256", "RS512", "ES256", "ES512", "ES384"].map((item) => Setting.getOption(item, item))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Token fields"), i18next.t("application:Token fields - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} disabled={this.state.application.tokenFormat !== "JWT-Custom"} mode="tags" showSearch style={{width: "100%"}} value={this.state.application.tokenFields} onChange={(value => {this.updateApplicationField("tokenFields", value);})}>
              <Option key={"provider"} value={"provider"}>{"Provider"}</Option>)
              {
                Setting.getUserCommonFields().map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Token expire"), i18next.t("application:Token expire - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber style={{width: "150px"}} value={this.state.application.expireInHours} min={1} step={1} precision={0} addonAfter="Hours" onChange={value => {
              this.updateApplicationField("expireInHours", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Refresh token expire"), i18next.t("application:Refresh token expire - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber style={{width: "150px"}} value={this.state.application.refreshExpireInHours} min={1} step={1} precision={0} addonAfter="Hours" onChange={value => {
              this.updateApplicationField("refreshExpireInHours", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Failed signin limit"), i18next.t("application:Failed signin limit - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber style={{width: "150px"}} value={this.state.application.failedSigninLimit} min={1} step={1} precision={0} addonAfter="Times" onChange={value => {
              this.updateApplicationField("failedSigninLimit", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Failed signin frozen time"), i18next.t("application:Failed signin frozen time - Tooltip"))} :
          </Col>
          <Col span={22} >
            <InputNumber style={{width: "150px"}} value={this.state.application.failedSigninFrozenTime} min={1} step={1} precision={0} addonAfter="Minutes" onChange={value => {
              this.updateApplicationField("failedSigninFrozenTime", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("ldap:Default group"), i18next.t("ldap:Default group - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Select virtual={false} style={{width: "100%"}} value={this.state.application.defaultGroup ?? []} onChange={(value => {
              this.updateApplicationField("defaultGroup", value);
            })}
            >
              <Option key={""} value={""}>
                <Space>
                  {i18next.t("general:Default")}
                </Space>
              </Option>
              {
                this.state.groups?.map((group) => <Option key={group.name} value={`${group.owner}/${group.name}`}>
                  <Space>
                    {group.type === "Physical" ? <UsergroupAddOutlined /> : <HolderOutlined />}
                    {group.displayName}
                  </Space>
                </Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Enable signup"), i18next.t("application:Enable signup - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableSignUp} onChange={checked => {
              this.updateApplicationField("enableSignUp", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Disable signin"), i18next.t("application:Disable signin - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.disableSignin} onChange={checked => {
              this.updateApplicationField("disableSignin", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Signin session"), i18next.t("application:Enable signin session - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableSigninSession} onChange={checked => {
              if (!checked) {
                this.updateApplicationField("enableAutoSignin", false);
              }

              this.updateApplicationField("enableSigninSession", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Auto signin"), i18next.t("application:Auto signin - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableAutoSignin} onChange={checked => {
              if (!this.state.application.enableSigninSession && checked) {
                Setting.showMessage("error", i18next.t("application:Please enable \"Signin session\" first before enabling \"Auto signin\""));
                return;
              }

              this.updateApplicationField("enableAutoSignin", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Enable Email linking"), i18next.t("application:Enable Email linking - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableLinkWithEmail} onChange={checked => {
              this.updateApplicationField("enableLinkWithEmail", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Signin methods"), i18next.t("application:Signin methods - Tooltip"))} :
          </Col>
          <Col span={22} >
            <SigninMethodTable
              title={i18next.t("application:Signin methods")}
              table={this.state.application.signinMethods}
              onUpdateTable={(value) => {
                this.updateApplicationField("signinMethods", value);
              }}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Org choice mode"), i18next.t("application:Org choice mode - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}}
              options={[
                {label: i18next.t("general:None"), value: "None"},
                {label: i18next.t("application:Select"), value: "Select"},
                {label: i18next.t("application:Input"), value: "Input"},
              ].map((item) => {
                return Setting.getOption(item.label, item.value);
              })}
              value={this.state.application.orgChoiceMode ?? []}
              onChange={(value => {
                this.updateApplicationField("orgChoiceMode", value);
              })} >
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Signup URL"), i18next.t("general:Signup URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.application.signupUrl} onChange={e => {
              this.updateApplicationField("signupUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Signin URL"), i18next.t("general:Signin URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.application.signinUrl} onChange={e => {
              this.updateApplicationField("signinUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Forget URL"), i18next.t("general:Forget URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.application.forgetUrl} onChange={e => {
              this.updateApplicationField("forgetUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Affiliation URL"), i18next.t("general:Affiliation URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.application.affiliationUrl} onChange={e => {
              this.updateApplicationField("affiliationUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:IP whitelist"), i18next.t("general:IP whitelist - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input placeholder = {this.state.application.organizationObj?.ipWhitelist} value={this.state.application.ipWhitelist} onChange={e => {
              this.updateApplicationField("ipWhitelist", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("signup:Terms of Use"), i18next.t("signup:Terms of Use - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.application.termsOfUse} style={{marginBottom: "10px"}} onChange={e => {
              this.updateApplicationField("termsOfUse", e.target.value);
            }} />
            <Upload maxCount={1} accept=".html" showUploadList={false}
              beforeUpload={file => {return false;}} onChange={info => {this.handleUpload(info);}}>
              <Button icon={<UploadOutlined />} loading={this.state.uploading}>{i18next.t("general:Click to Upload")}</Button>
            </Upload>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Signup HTML"), i18next.t("provider:Signup HTML - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}} >
                <Editor value={this.state.application.signupHtml} lang="html" fillHeight dark onChange={value => {
                  this.updateApplicationField("signupHtml", value);
                }} />
              </div>
            } title={i18next.t("provider:Signup HTML - Edit")} trigger="click">
              <Input value={this.state.application.signupHtml} style={{marginBottom: "10px"}} onChange={e => {
                this.updateApplicationField("signupHtml", e.target.value);
              }} />
            </Popover>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Signin HTML"), i18next.t("provider:Signin HTML - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}} >
                <Editor value={this.state.application.signinHtml} lang="html" fillHeight dark onChange={value => {
                  this.updateApplicationField("signinHtml", value);
                }} />
              </div>
            } title={i18next.t("provider:Signin HTML - Edit")} trigger="click">
              <Input value={this.state.application.signinHtml} style={{marginBottom: "10px"}} onChange={e => {
                this.updateApplicationField("signinHtml", e.target.value);
              }} />
            </Popover>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Grant types"), i18next.t("application:Grant types - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}}
              value={this.state.application.grantTypes}
              onChange={(value => {
                this.updateApplicationField("grantTypes", value);
              })} >
              {
                [
                  {id: "authorization_code", name: "Authorization Code"},
                  {id: "password", name: "Password"},
                  {id: "client_credentials", name: "Client Credentials"},
                  {id: "token", name: "Token"},
                  {id: "id_token", name: "ID Token"},
                  {id: "refresh_token", name: "Refresh Token"},
                  {id: "urn:ietf:params:oauth:grant-type:device_code", name: "Device Code"},
                ].map((item, index) => <Option key={index} value={item.id}>{item.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:SAML reply URL"), i18next.t("application:Redirect URL (Assertion Consumer Service POST Binding URL) - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.application.samlReplyUrl} onChange={e => {
              this.updateApplicationField("samlReplyUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Enable SAML compression"), i18next.t("application:Enable SAML compression - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableSamlCompress} onChange={checked => {
              this.updateApplicationField("enableSamlCompress", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Enable SAML C14N10"), i18next.t("application:Enable SAML C14N10 - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableSamlC14n10} onChange={checked => {
              this.updateApplicationField("enableSamlC14n10", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Use Email as NameID"), i18next.t("application:Use Email as NameID - Tooltip"))} :
          </Col>
          <Col span={1}>
            <Switch checked={this.state.application.useEmailAsSamlNameId} onChange={checked => {
              this.updateApplicationField("useEmailAsSamlNameId", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Enable SAML POST binding"), i18next.t("application:Enable SAML POST binding - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableSamlPostBinding} onChange={checked => {
              this.updateApplicationField("enableSamlPostBinding", checked);
              this.getSamlMetadata(checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:SAML attributes"), i18next.t("general:SAML attributes - Tooltip"))} :
          </Col>
          <Col span={22} >
            <SamlAttributeTable
              title={i18next.t("general:SAML attributes")}
              table={this.state.application.samlAttributes}
              application={this.state.application}
              onUpdateTable={(value) => {this.updateApplicationField("samlAttributes", value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:SAML metadata"), i18next.t("application:SAML metadata - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Editor value={this.state.samlMetadata?.toString() ?? ""} lang="xml" readOnly />
            <br />
            <Button style={{marginBottom: "10px"}} type="primary" shape="round" icon={<CopyOutlined />} onClick={() => {
              copy(`${window.location.origin}/api/saml/metadata?application=admin/${encodeURIComponent(this.state.applicationName)}&enablePostBinding=${this.state.application.enableSamlPostBinding}`);
              Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
            }}
            >
              {i18next.t("application:Copy SAML metadata URL")}
            </Button>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Providers"), i18next.t("general:Providers - Tooltip"))} :
          </Col>
          <Col span={22} >
            <ProviderTable
              title={i18next.t("general:Providers")}
              table={this.state.application.providers}
              providers={this.state.providers}
              application={this.state.application}
              onUpdateTable={(value) => {this.updateApplicationField("providers", value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
          </Col>
          {
            this.renderSignupSigninPreview()
          }
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Background URL"), i18next.t("application:Background URL - Tooltip"))} :
          </Col>
          <Col span={22} style={(Setting.isMobile()) ? {maxWidth: "100%"} : {}}>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input prefix={<LinkOutlined />} value={this.state.application.formBackgroundUrl} onChange={e => {
                  this.updateApplicationField("formBackgroundUrl", e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={22} >
                <a target="_blank" rel="noreferrer" href={this.state.application.formBackgroundUrl}>
                  <img src={this.state.application.formBackgroundUrl} alt={this.state.application.formBackgroundUrl} height={90} style={{marginBottom: "20px"}} />
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Background URL Mobile"), i18next.t("application:Background URL Mobile - Tooltip"))} :
          </Col>
          <Col span={22} style={(Setting.isMobile()) ? {maxWidth: "100%"} : {}}>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input prefix={<LinkOutlined />} value={this.state.application.formBackgroundUrlMobile} onChange={e => {
                  this.updateApplicationField("formBackgroundUrlMobile", e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={22} >
                <a target="_blank" rel="noreferrer" href={this.state.application.formBackgroundUrlMobile}>
                  <img src={this.state.application.formBackgroundUrlMobile} alt={this.state.application.formBackgroundUrlMobile} height={90} style={{marginBottom: "20px"}} />
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Custom CSS"), i18next.t("application:Custom CSS - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}} >
                <Editor
                  value={this.state.application.formCss === "" ? template : this.state.application.formCss}
                  lang="css"
                  fillHeight
                  dark
                  onChange={value => {
                    this.updateApplicationField("formCss", value);
                  }}
                />
              </div>
            } title={i18next.t("application:Custom CSS - Edit")} trigger="click">
              <Input value={this.state.application.formCss} style={{marginBottom: "10px"}} onChange={e => {
                this.updateApplicationField("formCss", e.target.value);
              }} />
            </Popover>
          </Col>
        </Row>
        <Row>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Custom CSS Mobile"), i18next.t("application:Custom CSS Mobile - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}} >
                <Editor
                  value={this.state.application.formCssMobile === "" ? template : this.state.application.formCssMobile}
                  lang="css"
                  fillHeight
                  dark
                  onChange={value => {
                    this.updateApplicationField("formCssMobile", value);
                  }}
                />
              </div>
            } title={i18next.t("application:Custom CSS Mobile - Edit")} trigger="click">
              <Input value={this.state.application.formCssMobile} style={{marginBottom: "10px"}} onChange={e => {
                this.updateApplicationField("formCssMobile", e.target.value);
              }} />
            </Popover>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Form position"), i18next.t("application:Form position - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: "20px"}} >
              <Radio.Group buttonStyle="solid" onChange={e => {this.updateApplicationField("formOffset", e.target.value);}} value={this.state.application.formOffset}>
                <Radio.Button value={1}>{i18next.t("application:Left")}</Radio.Button>
                <Radio.Button value={2}>{i18next.t("application:Center")}</Radio.Button>
                <Radio.Button value={3}>{i18next.t("application:Right")}</Radio.Button>
                <Radio.Button value={4}>
                  {i18next.t("application:Enable side panel")}
                </Radio.Button>
              </Radio.Group>
            </Row>
            {this.state.application.formOffset === 4 ?
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 3}>
                  {Setting.getLabel(i18next.t("application:Side panel HTML"), i18next.t("application:Side panel HTML - Tooltip"))} :
                </Col>
                <Col span={21} >
                  <Popover placement="right" content={
                    <div style={{width: "900px", height: "300px"}} >
                      <Editor
                        value={this.state.application.formSideHtml === "" ? sideTemplate : this.state.application.formSideHtml}
                        lang="html"
                        fillHeight
                        dark
                        onChange={value => {
                          this.updateApplicationField("formSideHtml", value);
                        }}
                      />
                    </div>
                  } title={i18next.t("application:Side panel HTML - Edit")} trigger="click">
                    <Input value={this.state.application.formSideHtml} style={{marginBottom: "10px"}} onChange={e => {
                      this.updateApplicationField("formSideHtml", e.target.value);
                    }} />
                  </Popover>
                </Col>
              </Row>
              : null}
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("theme:Theme"), i18next.t("theme:Theme - Tooltip"))} :
          </Col>
          <Col span={22} style={{marginTop: "5px"}}>
            <Row>
              <Radio.Group buttonStyle="solid" value={this.state.application.themeData?.isEnabled ?? false} onChange={e => {
                const {_, ...theme} = this.state.application.themeData ?? {...Conf.ThemeDefault, isEnabled: false};
                this.updateApplicationField("themeData", {...theme, isEnabled: e.target.value});
              }} >
                <Radio.Button value={false}>{i18next.t("application:Follow organization theme")}</Radio.Button>
                <Radio.Button value={true}>{i18next.t("theme:Customize theme")}</Radio.Button>
              </Radio.Group>
            </Row>
            {
              this.state.application.themeData?.isEnabled ?
                <Row style={{marginTop: "20px"}}>
                  <ThemeEditor themeData={this.state.application.themeData} onThemeChange={(_, nextThemeData) => {
                    const {isEnabled} = this.state.application.themeData ?? {...Conf.ThemeDefault, isEnabled: false};
                    this.updateApplicationField("themeData", {...nextThemeData, isEnabled});
                  }} />
                </Row> : null
            }
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Header HTML"), i18next.t("application:Header HTML - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}} >
                <Editor
                  value={this.state.application.headerHtml}
                  lang="html"
                  fillHeight
                  dark
                  onChange={value => {
                    this.updateApplicationField("headerHtml", value);
                  }}
                />
              </div>
            } title={i18next.t("application:Header HTML - Edit")} trigger="click">
              <Input value={this.state.application.headerHtml} style={{marginBottom: "10px"}} onChange={e => {
                this.updateApplicationField("headerHtml", e.target.value);
              }} />
            </Popover>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Footer HTML"), i18next.t("application:Footer HTML - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}} >
                <Editor
                  value={this.state.application.footerHtml}
                  lang="html"
                  fillHeight
                  dark
                  onChange={value => {
                    this.updateApplicationField("footerHtml", value);
                  }}
                />
              </div>
            } title={i18next.t("application:Footer HTML - Edit")} trigger="click">
              <Input value={this.state.application.footerHtml} style={{marginBottom: "10px"}} onChange={e => {
                this.updateApplicationField("footerHtml", e.target.value);
              }} />
            </Popover>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          </Col>
          <Button style={{marginLeft: "10px", marginBottom: "5px"}} onClick={() => this.updateApplicationField("footerHtml", Setting.getDefaultFooterContent())} >
            {i18next.t("provider:Reset to Default HTML")}
          </Button>
          <Button style={{marginLeft: "10px", marginBottom: "5px"}} onClick={() => this.updateApplicationField("footerHtml", Setting.getEmptyFooterContent())} >
            {i18next.t("application:Reset to Empty")}
          </Button>
        </Row>
        {
          <React.Fragment>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("application:Signin items"), i18next.t("application:Signin items - Tooltip"))} :
              </Col>
              <Col span={22} >
                <SigninTable
                  title={i18next.t("application:Signin items")}
                  table={this.state.application.signinItems}
                  themeAlgorithm={this.state.themeAlgorithm}
                  onUpdateTable={(value) => {
                    this.updateApplicationField("signinItems", value);
                  }}
                />
              </Col>
            </Row>
          </React.Fragment>
        }
        {
          !this.state.application.enableSignUp ? null : (
            <React.Fragment>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("application:Signup items"), i18next.t("application:Signup items - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <SignupTable
                    title={i18next.t("application:Signup items")}
                    table={this.state.application.signupItems}
                    onUpdateTable={(value) => {
                      this.updateApplicationField("signupItems", value);
                    }}
                  />
                </Col>
              </Row>
            </React.Fragment>
          )
        }
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
          </Col>
          {
            this.renderPromptPreview()
          }
        </Row>
      </Card>
    );
  }

  renderSignupSigninPreview() {
    const themeData = this.state.application.themeData ?? Conf.ThemeDefault;
    let signUpUrl = `/signup/${this.state.application.name}`;

    let redirectUri;
    if (this.state.application.redirectUris?.length > 0) {
      redirectUri = this.state.application.redirectUris[0];
    } else {
      redirectUri = "\"ERROR: You must specify at least one Redirect URL in 'Redirect URLs'\"";
    }

    let clientId = this.state.application.clientId;
    if (this.state.application.isShared) {
      clientId += `-org-${this.props.account.owner}`;
    }
    const signInUrl = `/login/oauth/authorize?client_id=${clientId}&response_type=code&redirect_uri=${redirectUri}&scope=read&state=casdoor`;
    const maskStyle = {position: "absolute", top: "0px", left: "0px", zIndex: 10, height: "97%", width: "100%", background: "rgba(0,0,0,0.4)"};
    if (!Setting.isPasswordEnabled(this.state.application)) {
      signUpUrl = signInUrl.replace("/login/oauth/authorize", "/signup/oauth/authorize");
    }

    return (
      <React.Fragment>
        <Col span={previewGrid}>
          <Button style={{marginBottom: "10px"}} type="primary" shape="round" icon={<CopyOutlined />} onClick={() => {
            copy(`${window.location.origin}${signUpUrl}`);
            Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
          }}
          >
            {i18next.t("application:Copy signup page URL")}
          </Button>
          <br />
          <ConfigProvider theme={{
            token: {
              colorPrimary: themeData.colorPrimary,
              colorInfo: themeData.colorPrimary,
              borderRadius: themeData.borderRadius,
            },
          }}>
            <div style={{position: "relative", width: previewWidth, border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", overflow: "auto"}}>
              {
                Setting.isPasswordEnabled(this.state.application) ? (
                  <div className="loginBackground" style={{backgroundImage: `url(${this.state.application?.formBackgroundUrl})`, overflow: "auto"}}>
                    <SignupPage application={this.state.application} preview = "auto" />
                  </div>
                ) : (
                  <div className="loginBackground" style={{backgroundImage: `url(${this.state.application?.formBackgroundUrl})`, overflow: "auto"}}>
                    <LoginPage type={"login"} mode={"signup"} application={this.state.application} preview = "auto" />
                  </div>
                )
              }
              <div style={{overflow: "auto", ...maskStyle}} />
            </div>
          </ConfigProvider>
        </Col>
        <Col span={previewGrid}>
          <Button style={{marginBottom: "10px", marginTop: Setting.isMobile() ? "15px" : "0"}} type="primary" shape="round" icon={<CopyOutlined />} onClick={() => {
            copy(`${window.location.origin}${signInUrl}`);
            Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
          }}
          >
            {i18next.t("application:Copy signin page URL")}
          </Button>
          <br />
          <ConfigProvider theme={{
            token: {
              colorPrimary: themeData.colorPrimary,
              colorInfo: themeData.colorPrimary,
              borderRadius: themeData.borderRadius,
            },
          }}>
            <div style={{position: "relative", width: previewWidth, border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", overflow: "auto"}}>
              <div className="loginBackground" style={{backgroundImage: `url(${this.state.application?.formBackgroundUrl})`, overflow: "auto"}}>
                <LoginPage type={"login"} mode={"signin"} application={this.state.application} preview = "auto" />
              </div>
              <div style={{overflow: "auto", ...maskStyle}} />
            </div>
          </ConfigProvider>
        </Col>
      </React.Fragment>
    );
  }

  renderPromptPreview() {
    const themeData = this.state.application.themeData ?? Conf.ThemeDefault;
    const promptUrl = `/prompt/${this.state.application.name}`;
    const maskStyle = {position: "absolute", top: "0px", left: "0px", zIndex: 10, height: "100%", width: "100%", background: "rgba(0,0,0,0.4)"};
    return (
      <Col span={previewGrid}>
        <Button style={{marginBottom: "10px"}} type="primary" shape="round" icon={<CopyOutlined />} onClick={() => {
          copy(`${window.location.origin}${promptUrl}`);
          Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
        }}
        >
          {i18next.t("application:Copy prompt page URL")}
        </Button>
        <br />
        <ConfigProvider theme={{
          token: {
            colorPrimary: themeData.colorPrimary,
            colorInfo: themeData.colorPrimary,
            borderRadius: themeData.borderRadius,
          },
        }}>
          <div style={{position: "relative", width: previewWidth, border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", flexDirection: "column", flex: "auto"}}>
            <PromptPage application={this.state.application} account={this.props.account} />
            <div style={maskStyle} />
          </div>
        </ConfigProvider>
      </Col>
    );
  }

  submitApplicationEdit(exitAfterSave) {
    const application = Setting.deepCopy(this.state.application);
    application.providers = application.providers?.filter(provider => this.state.providers.map(provider => provider.name).includes(provider.name));
    application.signinMethods = application.signinMethods?.filter(signinMethod => ["Password", "Verification code", "WebAuthn", "LDAP", "Face ID", "WeChat"].includes(signinMethod.name));

    ApplicationBackend.updateApplication("admin", this.state.applicationName, application)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));
          this.setState({
            applicationName: this.state.application.name,
          });

          if (exitAfterSave) {
            this.props.history.push("/applications");
          } else {
            this.props.history.push(`/applications/${this.state.application.organization}/${this.state.application.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateApplicationField("name", this.state.applicationName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteApplication() {
    ApplicationBackend.deleteApplication(this.state.application)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/applications");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  render() {
    if (!this.state.isAuthorized) {
      return (
        <Result
          status="403"
          title="403 Unauthorized"
          subTitle={i18next.t("general:Sorry, you do not have permission to access this page or logged in status invalid.")}
          extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>}
        />
      );
    }

    return (
      <div>
        {
          this.state.application !== null ? this.renderApplication() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitApplicationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitApplicationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteApplication()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default ApplicationEditPage;
