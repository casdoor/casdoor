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
import {Button, Card, Col, Input, Popover, Row, Select, Switch, Upload} from 'antd';
import {LinkOutlined, UploadOutlined} from "@ant-design/icons";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as CertBackend from "./backend/CertBackend";
import * as Setting from "./Setting";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ResourceBackend from "./backend/ResourceBackend";
import SignupPage from "./auth/SignupPage";
import LoginPage from "./auth/LoginPage";
import i18next from "i18next";
import UrlTable from "./UrlTable";
import ProviderTable from "./ProviderTable";
import SignupTable from "./SignupTable";
import PromptPage from "./auth/PromptPage";

import {Controlled as CodeMirror} from 'react-codemirror2';
import "codemirror/lib/codemirror.css";
require('codemirror/theme/material-darker.css');
require("codemirror/mode/htmlmixed/htmlmixed");

const { Option } = Select;

class ApplicationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applicationName: props.match.params.applicationName,
      application: null,
      organizations: [],
      certs: [],
      providers: [],
      uploading: false,
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getApplication();
    this.getOrganizations();
    this.getCerts();
    this.getProviders();
  }

  getApplication() {
    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        if (application.grantTypes === null || application.grantTypes === undefined || application.grantTypes.length === 0) {
          application.grantTypes = ["authorization_code"];
        }
        this.setState({
          application: application,
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

  getCerts() {
    CertBackend.getCerts("admin")
      .then((res) => {
        this.setState({
          certs: (res.msg === undefined) ? res : [],
        });
      });
  }

  getProviders() {
    ProviderBackend.getProviders("admin")
      .then((res) => {
        this.setState({
          providers: res,
        });
      });
  }

  parseApplicationField(key, value) {
    if (["expireInHours", "refreshExpireInHours"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateApplicationField(key, value) {
    value = this.parseApplicationField(key, value);

    let application = this.state.application;
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
          Setting.showMessage("error", res.msg);
        }
      }).finally(() => {
        this.setState({uploading: false});
    })
  }

  renderApplication() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("application:New Application") : i18next.t("application:Edit Application")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitApplicationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" onClick={() => this.submitApplicationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: '20px'}} onClick={() => this.deleteApplication()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile())? {margin: '5px'}:{}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.name} disabled={this.state.application.name === "app-built-in"} onChange={e => {
              this.updateApplicationField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.displayName} onChange={e => {
              this.updateApplicationField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Logo"), i18next.t("general:Logo - Tooltip"))} :
          </Col>
          <Col span={22} style={(Setting.isMobile()) ? {maxWidth:'100%'} :{}}>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined/>} value={this.state.application.logo} onChange={e => {
                  this.updateApplicationField('logo', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 1}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" rel="noreferrer" href={this.state.application.logo}>
                  <img src={this.state.application.logo} alt={this.state.application.logo} height={90} style={{marginBottom: '20px'}}/>
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Home"), i18next.t("general:Home - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.homepageUrl} onChange={e => {
              this.updateApplicationField('homepageUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.description} onChange={e => {
              this.updateApplicationField('description', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.application.organization} onChange={(value => {this.updateApplicationField('organization', value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.clientId} onChange={e => {
              this.updateApplicationField('clientId', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.clientSecret} onChange={e => {
              this.updateApplicationField('clientSecret', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Cert"), i18next.t("general:Cert - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.application.cert} onChange={(value => {this.updateApplicationField('cert', value);})}>
              {
                this.state.certs.map((cert, index) => <Option key={index} value={cert.name}>{cert.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Redirect URLs"), i18next.t("application:Redirect URLs - Tooltip"))} :
          </Col>
          <Col span={22} >
            <UrlTable
              title={i18next.t("application:Redirect URLs")}
              table={this.state.application.redirectUris}
              onUpdateTable={(value) => { this.updateApplicationField('redirectUris', value)}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Token format"), i18next.t("application:Token format - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.application.tokenFormat} onChange={(value => {this.updateApplicationField('tokenFormat', value);})}>
              {
                ['JWT', 'JWT-Empty']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Token expire"), i18next.t("application:Token expire - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input style={{width: "150px"}} value={this.state.application.expireInHours} suffix="Hours" onChange={e => {
              this.updateApplicationField('expireInHours', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Refresh token expire"), i18next.t("application:Refresh token expire - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input style={{width: "150px"}} value={this.state.application.refreshExpireInHours} suffix="Hours" onChange={e => {
              this.updateApplicationField('refreshExpireInHours', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Password ON"), i18next.t("application:Password ON - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enablePassword} onChange={checked => {
              this.updateApplicationField('enablePassword', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Enable signup"), i18next.t("application:Enable signup - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableSignUp} onChange={checked => {
              this.updateApplicationField('enableSignUp', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Signin session"), i18next.t("application:Enable signin session - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableSigninSession} onChange={checked => {
              this.updateApplicationField('enableSigninSession', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("application:Enable code signin"), i18next.t("application:Enable code signin - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableCodeSignin} onChange={checked => {
              this.updateApplicationField('enableCodeSignin', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Signup URL"), i18next.t("general:Signup URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.signupUrl} onChange={e => {
              this.updateApplicationField('signupUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Signin URL"), i18next.t("general:Signin URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.signinUrl} onChange={e => {
              this.updateApplicationField('signinUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Forget URL"), i18next.t("general:Forget URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.forgetUrl} onChange={e => {
              this.updateApplicationField('forgetUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Affiliation URL"), i18next.t("general:Affiliation URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.affiliationUrl} onChange={e => {
              this.updateApplicationField('affiliationUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Terms of Use"), i18next.t("provider:Terms of Use - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.termsOfUse} style={{marginBottom: "10px"}} onChange={e => {
              this.updateApplicationField("termsOfUse", e.target.value);
            }}/>
            <Upload maxCount={1} accept=".html" showUploadList={false}
                    beforeUpload={file => {return false}} onChange={info => {this.handleUpload(info)}}>
              <Button icon={<UploadOutlined />} loading={this.state.uploading}>Click to Upload</Button>
            </Upload>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Signup HTML"), i18next.t("provider:Signup HTML - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}} >
                <CodeMirror
                  value={this.state.application.signupHtml}
                  options={{mode: 'htmlmixed', theme: "material-darker"}}
                  onBeforeChange={(editor, data, value) => {
                    this.updateApplicationField("signupHtml", value);
                  }}
                />
              </div>
            } title={i18next.t("provider:Signup HTML - Edit")} trigger="click">
              <Input value={this.state.application.signupHtml} style={{marginBottom: "10px"}} onChange={e => {
                this.updateApplicationField("signupHtml", e.target.value)
              }}/>
            </Popover>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Signin HTML"), i18next.t("provider:Signin HTML - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Popover placement="right" content={
              <div style={{width: "900px", height: "300px"}} >
                <CodeMirror
                  value={this.state.application.signinHtml}
                  options={{mode: 'htmlmixed', theme: "material-darker"}}
                  onBeforeChange={(editor, data, value) => {
                    this.updateApplicationField("signinHtml", value);
                  }}
                />
              </div>
            } title={i18next.t("provider:Signin HTML - Edit")} trigger="click">
              <Input value={this.state.application.signinHtml} style={{marginBottom: "10px"}} onChange={e => {
                this.updateApplicationField("signinHtml", e.target.value)
              }}/>
            </Popover>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("application:Grant types"), i18next.t("application:Grant types - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: '100%'}}
                    value={this.state.application.grantTypes}
                    onChange={(value => {
                      this.updateApplicationField('grantTypes', value);
                    })} >
                      {
                        [
                          {id: "authorization_code", name: "Authorization Code"},
                          {id: "password", name: "Password"},
                          {id: "client_credentials", name: "Client Credentials"},
                          {id: "token", name: "Token"},
                          {id: "id_token",name:"ID Token"},
                          {id: "wechat_miniprogram", name:"WeChat Mini Program"}
                        ].map((item, index)=><Option key={index} value={item.id}>{item.name}</Option>)
                      }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Providers"), i18next.t("general:Providers - Tooltip"))} :
          </Col>
          <Col span={22} >
            <ProviderTable
              title={i18next.t("general:Providers")}
              table={this.state.application.providers}
              providers={this.state.providers}
              application={this.state.application}
              onUpdateTable={(value) => { this.updateApplicationField('providers', value)}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
          </Col>
          {
            this.renderPreview()
          }
        </Row>
        {
          !this.state.application.enableSignUp ? null : (
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("application:Signup items"), i18next.t("application:Signup items - Tooltip"))} :
              </Col>
              <Col span={22} >
                <SignupTable
                  title={i18next.t("application:Signup items")}
                  table={this.state.application.signupItems}
                  onUpdateTable={(value) => { this.updateApplicationField('signupItems', value)}}
                />
              </Col>
            </Row>
          )
        }
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
          </Col>
          {
            this.renderPreview2()
          }
        </Row>
      </Card>
    )
  }

  renderPreview() {
    let signUpUrl = `/signup/${this.state.application.name}`;
    let signInUrl = `/login/oauth/authorize?client_id=${this.state.application.clientId}&response_type=code&redirect_uri=${this.state.application.redirectUris[0]}&scope=read&state=casdoor`;
    if (!this.state.application.enablePassword) {
      signUpUrl = signInUrl.replace("/login/oauth/authorize", "/signup/oauth/authorize");
    }
    if (!Setting.isMobile()) {
    return (
      <React.Fragment>
        <Col span={11} style={{display:"flex", flexDirection: "column"}}>
          <a style={{marginBottom: "10px", display: "flex"}} target="_blank" rel="noreferrer" href={signUpUrl}>
            <Button type="primary">{i18next.t("application:Test signup page..")}</Button>
          </a>
          <br/>
          <br/>
          <div style={{width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", alignItems:"center", overflow:"auto", flexDirection:"column", flex: "auto"}}>
            {
              this.state.application.enablePassword ? (
                <SignupPage application={this.state.application} />
              ) : (
                <LoginPage type={"login"} mode={"signup"} application={this.state.application} />
              )
            }
          </div>
        </Col>
        <Col span={11} style={{display:"flex", flexDirection: "column"}}>
          <a style={{marginBottom: "10px", display: "flex"}} target="_blank" rel="noreferrer" href={signInUrl}>
            <Button type="primary">{i18next.t("application:Test signin page..")}</Button>
          </a>
          <br/>
          <br/>
          <div style={{width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", alignItems:"center", overflow:"auto", flexDirection:"column", flex: "auto"}}>
            <LoginPage type={"login"} mode={"signin"} application={this.state.application} />
          </div>
        </Col>
      </React.Fragment>
    )
  } else{
    return(
      <React.Fragment>
        <Col span={24} style={{display:"flex", flexDirection: "column"}}>
          <a style={{marginBottom: "10px", display: "flex"}} target="_blank" rel="noreferrer" href={signUpUrl}>
            <Button type="primary">{i18next.t("application:Test signup page..")}</Button>
          </a>
          <div style={{marginBottom:"10px", width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", alignItems: "center", overflow: "auto", flexDirection: "column", flex: "auto"}}>
            {
              this.state.application.enablePassword ? (
                <SignupPage application={this.state.application} />
              ) : (
                <LoginPage type={"login"} mode={"signup"} application={this.state.application} />
              )
            }
          </div>
          <a style={{marginBottom: "10px", display: "flex"}} target="_blank" rel="noreferrer" href={signInUrl}>
            <Button type="primary">{i18next.t("application:Test signin page..")}</Button>
          </a>
          <div style={{width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", alignItems: "center", overflow: "auto", flexDirection: "column", flex: "auto"}}>
            <LoginPage type={"login"} mode={"signin"} application={this.state.application} />
          </div>
        </Col>
      </React.Fragment>
    )
  }
}

  renderPreview2() {
    let promptUrl = `/prompt/${this.state.application.name}`;

    return (
      <React.Fragment>
        <Col span={(Setting.isMobile()) ? 24 : 11} style={{display:"flex", flexDirection: "column", flex: "auto"}} >
          <a style={{marginBottom: "10px"}} target="_blank" rel="noreferrer" href={promptUrl}>
            <Button type="primary">{i18next.t("application:Test prompt page..")}</Button>
          </a>
          <br style={(Setting.isMobile()) ? {display: "none"} : {}} />
          <br style={(Setting.isMobile()) ? {display: "none"} : {}} />
          <div style={{width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888", flexDirection: "column", flex: "auto"}}>
            <PromptPage application={this.state.application} account={this.props.account} />
          </div>
        </Col>
      </React.Fragment>
    )
  }

  submitApplicationEdit(willExist) {
    let application = Setting.deepCopy(this.state.application);
    ApplicationBackend.updateApplication(this.state.application.owner, this.state.applicationName, application)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            applicationName: this.state.application.name,
          });

          if (willExist) {
            this.props.history.push(`/applications`);
          } else {
            this.props.history.push(`/applications/${this.state.application.name}`);
          }
        } else {
          Setting.showMessage("error", res.msg);
          this.updateApplicationField('name', this.state.applicationName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  deleteApplication() {
    ApplicationBackend.deleteApplication(this.state.application)
      .then(() => {
        this.props.history.push(`/applications`);
      })
      .catch(error => {
        Setting.showMessage("error", `Application failed to delete: ${error}`);
      });
  }

  render() {
    return (
      <div>
      {
        this.state.application !== null ? this.renderApplication() : null
      }
      <div style={{marginTop: '20px', marginLeft: '40px'}}>
        <Button size="large" onClick={() => this.submitApplicationEdit(false)}>{i18next.t("general:Save")}</Button>
        <Button style={{marginLeft: '20px'}} type="primary" size="large" onClick={() => this.submitApplicationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        {this.state.mode === "add" ? <Button style={{marginLeft: '20px'}} size="large" onClick={() => this.deleteApplication()}>{i18next.t("general:Cancel")}</Button> : null}
      </div>
    </div>
    );
  }
}

export default ApplicationEditPage;
