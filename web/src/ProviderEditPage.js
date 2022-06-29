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
import {Button, Card, Col, Input, InputNumber, Row, Select, Switch} from 'antd';
import {LinkOutlined} from "@ant-design/icons";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import { authConfig } from "./auth/Auth";
import * as ProviderEditTestEmail from "./TestEmailWidget";
import copy from 'copy-to-clipboard';
import { CaptchaPreview } from "./common/CaptchaPreview";

const { Option } = Select;
const { TextArea } = Input;

class ProviderEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      providerName: props.match.params.providerName,
      provider: null,
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
      testEmail: this.props.account["email"] !== undefined ? this.props.account["email"] : "",
    };
  }

  UNSAFE_componentWillMount() {
    this.getProvider();
  }

  getProvider() {
    ProviderBackend.getProvider("admin", this.state.providerName)
      .then((provider) => {
        this.setState({
          provider: provider,
        });
      });
  }

  parseProviderField(key, value) {
    if (["port"].includes(key)) {
      value = Setting.myParseInt(value);
    }
    return value;
  }

  updateProviderField(key, value) {
    value = this.parseProviderField(key, value);

    let provider = this.state.provider;
    provider[key] = value;
    this.setState({
      provider: provider,
    });
  }

  getClientIdLabel() {
    switch (this.state.provider.category) {
      case "Email":
        return Setting.getLabel(i18next.t("signup:Username"), i18next.t("signup:Username - Tooltip"));
      case "SMS":
        if (this.state.provider.type === "Volc Engine SMS") {
          return Setting.getLabel(i18next.t("provider:Access key"), i18next.t("provider:Access key - Tooltip"));
        } else if (this.state.provider.type === "Huawei Cloud SMS") {
          return Setting.getLabel(i18next.t("provider:App key"), i18next.t("provider:App key - Tooltip"));
        } else {
          return Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
        }
      case "Captcha":
        if (this.state.provider.type === "Aliyun Captcha") {
          return Setting.getLabel(i18next.t("provider:Access key"), i18next.t("provider:Access key - Tooltip"));
        } else {
          return Setting.getLabel(i18next.t("provider:Site key"), i18next.t("provider:Site key - Tooltip"));
        }
      default:
        return Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"));
    }
  }

  getClientSecretLabel() {
    switch (this.state.provider.category) {
      case "Email":
        return Setting.getLabel(i18next.t("login:Password"), i18next.t("login:Password - Tooltip"));
      case "SMS":
        if (this.state.provider.type === "Volc Engine SMS") {
          return Setting.getLabel(i18next.t("provider:Secret access key"), i18next.t("provider:SecretAccessKey - Tooltip"));
        } else if (this.state.provider.type === "Huawei Cloud SMS") {
          return Setting.getLabel(i18next.t("provider:App secret"), i18next.t("provider:AppSecret - Tooltip"));
        } else {
          return Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
        }
      case "Captcha":
        if (this.state.provider.type === "Aliyun Captcha") {
          return Setting.getLabel(i18next.t("provider:Secret access key"), i18next.t("provider:SecretAccessKey - Tooltip"));
        } else {
          return Setting.getLabel(i18next.t("provider:Secret key"), i18next.t("provider:Secret key - Tooltip"));
        }
      default:
        return Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"));
    }
  }

  getAppIdRow() {
    let text, tooltip;
    if (this.state.provider.category === "SMS" && this.state.provider.type === "Tencent Cloud SMS") {
      text = i18next.t("provider:App ID");
      tooltip = i18next.t("provider:App ID - Tooltip");
    } else if (this.state.provider.type === "WeCom" && this.state.provider.subType === "Internal") {
      text = i18next.t("provider:Agent ID");
      tooltip = i18next.t("provider:Agent ID - Tooltip");
    } else if (this.state.provider.type === "Infoflow"){
      text = i18next.t("provider:Agent ID");
      tooltip = i18next.t("provider:Agent ID - Tooltip");
    } else if (this.state.provider.category === "SMS" && this.state.provider.type === "Volc Engine SMS") {
      text = i18next.t("provider:SMS account");
      tooltip = i18next.t("provider:SMS account - Tooltip");
    } else if (this.state.provider.category === "SMS" && this.state.provider.type === "Huawei Cloud SMS") {
      text = i18next.t("provider:Channel No.");
      tooltip = i18next.t("provider:Channel No. - Tooltip");
    } else {
      return null;
    }

    return <Row style={{marginTop: '20px'}} >
      <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
        {Setting.getLabel(text, tooltip)} :
      </Col>
      <Col span={22} >
        <Input value={this.state.provider.appId} onChange={e => {
          this.updateProviderField('appId', e.target.value);
        }} />
      </Col>
    </Row>;
  }

  loadSamlConfiguration() {
    var parser = new DOMParser();
    var xmlDoc = parser.parseFromString(this.state.provider.metadata, "text/xml");
    var cert = xmlDoc.getElementsByTagName("ds:X509Certificate")[0].childNodes[0].nodeValue;
    var endpoint = xmlDoc.getElementsByTagName("md:SingleSignOnService")[0].getAttribute("Location");
    var issuerUrl = xmlDoc.getElementsByTagName("md:EntityDescriptor")[0].getAttribute("entityID");
    this.updateProviderField("idP", cert);
    this.updateProviderField("endpoint", endpoint);
    this.updateProviderField("issuerUrl", issuerUrl);
  }

  renderProvider() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("provider:New Provider") : i18next.t("provider:Edit Provider")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitProviderEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" onClick={() => this.submitProviderEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: '20px'}} onClick={() => this.deleteProvider()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile())? {margin: '5px'}:{}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.name} onChange={e => {
              this.updateProviderField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.displayName} onChange={e => {
              this.updateProviderField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Category"), i18next.t("provider:Category - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.provider.category} onChange={(value => {
              this.updateProviderField('category', value);
              if (value === "OAuth") {
                this.updateProviderField('type', 'GitHub');
              } else if (value === "Email") {
                this.updateProviderField('type', 'Default');
                this.updateProviderField('title', 'Casdoor Verification Code');
                this.updateProviderField('content', 'You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes.');
              } else if (value === "SMS") {
                this.updateProviderField('type', 'Aliyun SMS');
              } else if (value === "Storage") {
                this.updateProviderField('type', 'Local File System');
                this.updateProviderField('domain', Setting.getFullServerUrl());
              } else if (value === "SAML") {
                this.updateProviderField('type', 'Aliyun IDaaS');
              } else if (value === "Captcha") {
                this.updateProviderField('type', 'Default');
              }
            })}>
              {
                [
                  {id: 'OAuth', name: 'OAuth'},
                  {id: 'Email', name: 'Email'},
                  {id: 'SMS', name: 'SMS'},
                  {id: 'Storage', name: 'Storage'},
                  {id: 'SAML', name: 'SAML'},
                  {id: 'Payment', name: 'Payment'},
                  {id: 'Captcha', name: 'Captcha'},
                ].map((providerCategory, index) => <Option key={index} value={providerCategory.id}>{providerCategory.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Type"), i18next.t("provider:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.provider.type} onChange={(value => {
              this.updateProviderField('type', value);
              if (value === "Local File System") {
                this.updateProviderField('domain', Setting.getFullServerUrl());
              }
              if (value === "Custom") {
                this.updateProviderField('customAuthUrl', 'https://door.casdoor.com/login/oauth/authorize');
                this.updateProviderField('customScope', 'openid profile email');
                this.updateProviderField('customTokenUrl', 'https://door.casdoor.com/api/login/oauth/access_token');
                this.updateProviderField('customUserInfoUrl', 'https://door.casdoor.com/api/userinfo');
              }
            })}>
              {
                Setting.getProviderTypeOptions(this.state.provider.category).map((providerType, index) => <Option key={index} value={providerType.id}>{providerType.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        {
          this.state.provider.type !== "WeCom" && this.state.provider.type !== "Infoflow" && this.state.provider.type !== "Aliyun Captcha" ? null : (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Sub type"), i18next.t("provider:Sub type - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Select virtual={false} style={{width: '100%'}} value={this.state.provider.subType} onChange={value => {
                    this.updateProviderField('subType', value);
                  }}>
                    {
                      Setting.getProviderSubTypeOptions(this.state.provider.type).map((providerSubType, index) => <Option key={index} value={providerSubType.id}>{providerSubType.name}</Option>)
                    }
                  </Select>
                </Col>
              </Row>
              {
                this.state.provider.type !== "WeCom"  ? null : (
                  <Row style={{marginTop: '20px'}} >
                    <Col style={{marginTop: '5px'}} span={2}>
                      {Setting.getLabel(i18next.t("provider:Method"), i18next.t("provider:Method - Tooltip"))} :
                    </Col>
                    <Col span={22} >
                      <Select virtual={false} style={{width: '100%'}} value={this.state.provider.method} onChange={value => {
                        this.updateProviderField('method', value);
                      }}>
                        {
                          [{name: "Normal"}, {name: "Silent"}].map((method, index) => <Option key={index} value={method.name}>{method.name}</Option>)
                        }
                      </Select>
                    </Col>
                  </Row>)
              }
            </React.Fragment>
          )
        }
        {
          this.state.provider.type !== "Custom" ? null : (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Auth URL"), i18next.t("provider:Auth URL - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.customAuthUrl} onChange={e => {
                    this.updateProviderField('customAuthUrl', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Scope"), i18next.t("provider:Scope - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.customScope} onChange={e => {
                    this.updateProviderField('customScope', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Token URL"), i18next.t("provider:Token URL - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.customTokenUrl} onChange={e => {
                    this.updateProviderField('customTokenUrl', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:UserInfo URL"), i18next.t("provider:UserInfo URL - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.customUserInfoUrl} onChange={e => {
                    this.updateProviderField('customUserInfoUrl', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel( i18next.t("general:Favicon"), i18next.t("general:Favicon - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Row style={{marginTop: '20px'}} >
                    <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 1}>
                      {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
                    </Col>
                    <Col span={23} >
                      <Input prefix={<LinkOutlined/>} value={this.state.provider.customLogo} onChange={e => {
                        this.updateProviderField('customLogo', e.target.value);
                      }} />
                    </Col>
                  </Row>
                  <Row style={{marginTop: '20px'}} >
                    <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 1}>
                      {i18next.t("general:Preview")}:
                    </Col>
                    <Col span={23} >
                      <a target="_blank" rel="noreferrer" href={this.state.provider.customLogo}>
                        <img src={this.state.provider.customLogo} alt={this.state.provider.customLogo} height={90} style={{marginBottom: '20px'}}/>
                      </a>
                    </Col>
                  </Row>
                </Col>
              </Row>
            </React.Fragment>
          )
        }
        {
          this.state.provider.category === "Captcha" && this.state.provider.type === "Default" ? null : (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {this.getClientIdLabel()}
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.clientId} onChange={e => {
                    this.updateProviderField('clientId', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {this.getClientSecretLabel()}
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.clientSecret} onChange={e => {
                    this.updateProviderField('clientSecret', e.target.value);
                  }} />
                </Col>
              </Row>
            </React.Fragment>
          )
        }
        {
          this.state.provider.type !== "WeChat" && this.state.provider.type !== "Aliyun Captcha" ? null : (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {this.state.provider.type === "Aliyun Captcha"
                    ? Setting.getLabel(i18next.t("provider:Scene"), i18next.t("provider:Scene - Tooltip"))
                    : Setting.getLabel(i18next.t("provider:Client ID 2"), i18next.t("provider:Client ID 2 - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.clientId2} onChange={e => {
                    this.updateProviderField('clientId2', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {this.state.provider.type === "Aliyun Captcha"
                    ? Setting.getLabel(i18next.t("provider:App key"), i18next.t("provider:App key - Tooltip"))
                    : Setting.getLabel(i18next.t("provider:Client secret 2"), i18next.t("provider:Client secret 2 - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.clientSecret2} onChange={e => {
                    this.updateProviderField('clientSecret2', e.target.value);
                  }} />
                </Col>
              </Row>
            </React.Fragment>
          )
        }
        {
          this.state.provider.type !== "Adfs" &&  this.state.provider.type !== "Casdoor" && this.state.provider.type !== "Okta" ? null : (
            <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={2}>
              {Setting.getLabel(i18next.t("provider:Domain"), i18next.t("provider:Domain - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.provider.domain} onChange={e => {
                this.updateProviderField('domain', e.target.value);
              }} />
            </Col>
          </Row>
          )
        }
        {this.state.provider.category === "Storage" ? (
          <div>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                {Setting.getLabel(i18next.t("provider:Endpoint"), i18next.t("provider:Region endpoint for Internet"))} :
              </Col>
              <Col span={22} >
                <Input value={this.state.provider.endpoint} onChange={e => {
                  this.updateProviderField('endpoint', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                {Setting.getLabel(i18next.t("provider:Endpoint (Intranet)"), i18next.t("provider:Region endpoint for Intranet"))} :
              </Col>
              <Col span={22} >
                <Input value={this.state.provider.intranetEndpoint} onChange={e => {
                  this.updateProviderField('intranetEndpoint', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                {Setting.getLabel(i18next.t("provider:Bucket"), i18next.t("provider:Bucket - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input value={this.state.provider.bucket} onChange={e => {
                  this.updateProviderField('bucket', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                {Setting.getLabel(i18next.t("provider:Domain"), i18next.t("provider:Domain - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input value={this.state.provider.domain} onChange={e => {
                  this.updateProviderField('domain', e.target.value);
                }} />
              </Col>
            </Row>
            {this.state.provider.type === "AWS S3" || this.state.provider.type === "Tencent Cloud COS" ? (
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Region ID"), i18next.t("provider:Region ID - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.regionId} onChange={e => {
                    this.updateProviderField('regionId', e.target.value);
                  }} />
                </Col>
              </Row>
            ) : null}
          </div>
        ) : null}
        {
          this.state.provider.category === "Email" ? (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input prefix={<LinkOutlined/>} value={this.state.provider.host} onChange={e => {
                    this.updateProviderField('host', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Port"), i18next.t("provider:Port - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <InputNumber value={this.state.provider.port} onChange={value => {
                    this.updateProviderField('port', value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Email Title"), i18next.t("provider:Email Title - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.title} onChange={e => {
                    this.updateProviderField('title', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Email Content"), i18next.t("provider:Email Content - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <TextArea autoSize={{minRows: 1, maxRows: 100}} value={this.state.provider.content} onChange={e => {
                    this.updateProviderField('content', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Test Email"), i18next.t("provider:Test Email - Tooltip"))} :
                </Col>
                <Col span={4} >
                  <Input value={this.state.testEmail}
                         placeHolder = {i18next.t("user:Input your email")}
                         onChange={e => {
                    this.setState({testEmail: e.target.value})
                  }} />
                </Col>
                <Button style={{marginLeft: '10px', marginBottom: "5px"}} type="primary"
                        onClick={() => ProviderEditTestEmail.connectSmtpServer(this.state.provider)} >
                  {i18next.t("provider:Test Connection")}
                </Button>
                <Button style={{marginLeft: '10px', marginBottom: "5px"}} type="primary"
                        disabled={!Setting.isValidEmail(this.state.testEmail)}
                        onClick={() => ProviderEditTestEmail.sendTestEmail(this.state.provider, this.state.testEmail)} >
                  {i18next.t("provider:Send Test Email")}
                </Button>
              </Row>
            </React.Fragment>
          ) : this.state.provider.category === "SMS" ? (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Sign Name"), i18next.t("provider:Sign Name - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.signName} onChange={e => {
                    this.updateProviderField('signName', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Template Code"), i18next.t("provider:Template Code - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.templateCode} onChange={e => {
                    this.updateProviderField('templateCode', e.target.value);
                  }} />
                </Col>
              </Row>
            </React.Fragment>
          ) : this.state.provider.category === "SAML" ? (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Sign request"), i18next.t("provider:Sign request - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Switch checked={this.state.provider.enableSignAuthnRequest} onChange={checked => {
                    this.updateProviderField('enableSignAuthnRequest', checked);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Metadata"), i18next.t("provider:Metadata - Tooltip"))} :
                </Col>
                <Col span={22}>
                  <TextArea rows={4} value={this.state.provider.metadata} onChange={e => {
                    this.updateProviderField('metadata', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}}>
                <Col style={{marginTop: '5px'}} span={2}></Col>
                <Col span={2}>
                  <Button type="primary" onClick={() => {
                      try {
                        this.loadSamlConfiguration();
                        Setting.showMessage("success", i18next.t("provider:Parse Metadata successfully"));
                      } catch (err) {
                        Setting.showMessage("error", i18next.t("provider:Can not parse Metadata"));
                      }
                    }}>
                    {i18next.t("provider:Parse")}
                  </Button>
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Endpoint"), i18next.t("provider:SAML 2.0 Endpoint (HTTP)"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.endpoint} onChange={e => {
                    this.updateProviderField('endpoint', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:IdP"), i18next.t("provider:IdP public key"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.idP} onChange={e => {
                    this.updateProviderField('idP', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Issuer URL"), i18next.t("provider:Issuer URL - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.issuerUrl} onChange={e => {
                    this.updateProviderField('issuerUrl', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:SP ACS URL"), i18next.t("provider:SP ACS URL - Tooltip"))} :
                </Col>
                <Col span={21} >
                  <Input value={`${authConfig.serverUrl}/api/acs`} readOnly="readonly" />
                </Col>
                <Col span={1}>
                  <Button type="primary" onClick={() => {
                    copy(`${authConfig.serverUrl}/api/acs`);
                    Setting.showMessage("success", i18next.t("provider:Link copied to clipboard successfully"));
                  }}>
                    {i18next.t("provider:Copy")}
                  </Button>
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:SP Entity ID"), i18next.t("provider:SP ACS URL - Tooltip"))} :
                </Col>
                <Col span={21} >
                  <Input value={`${authConfig.serverUrl}/api/acs`} readOnly="readonly" />
                </Col>
                <Col span={1}>
                  <Button type="primary" onClick={() => {
                    copy(`${authConfig.serverUrl}/api/acs`);
                    Setting.showMessage("success", i18next.t("provider:Link copied to clipboard successfully"));
                  }}>
                    {i18next.t("provider:Copy")}
                  </Button>
                </Col>
              </Row>
            </React.Fragment>
          ) : null
        }
        {this.getAppIdRow()}
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Provider URL"), i18next.t("provider:Provider URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.provider.providerUrl} onChange={e => {
              this.updateProviderField('providerUrl', e.target.value);
            }} />
          </Col>
        </Row>
        {
          this.state.provider.category !== "Captcha" ? null : (
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
              </Col>
              <Col span={22} >
                <CaptchaPreview
                  provider={this.state.provider}
                  providerName={this.state.providerName}
                  clientSecret={this.state.provider.clientSecret}
                  captchaType={this.state.provider.type}
                  subType={this.state.provider.subType}
                  owner={this.state.provider.owner}
                  clientId={this.state.provider.clientId}
                  name={this.state.provider.name}
                  providerUrl={this.state.provider.providerUrl}
                  clientId2={this.state.provider.clientId2}
                  clientSecret2={this.state.provider.clientSecret2}
                />
              </Col>
            </Row>
          )
        }
      </Card>
    )
  }

  submitProviderEdit(willExist) {
    let provider = Setting.deepCopy(this.state.provider);
    ProviderBackend.updateProvider(this.state.provider.owner, this.state.providerName, provider)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            providerName: this.state.provider.name,
          });

          if (willExist) {
            this.props.history.push(`/providers`);
          } else {
            this.props.history.push(`/providers/${this.state.provider.name}`);
          }
        } else {
          Setting.showMessage("error", res.msg);
          this.updateProviderField('name', this.state.providerName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  deleteProvider() {
    ProviderBackend.deleteProvider(this.state.provider)
      .then(() => {
        this.props.history.push(`/providers`);
      })
      .catch(error => {
        Setting.showMessage("error", `Provider failed to delete: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.provider !== null ? this.renderProvider() : null
        }
        <div style={{marginTop: '20px', marginLeft: '40px'}}>
          <Button size="large" onClick={() => this.submitProviderEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" size="large" onClick={() => this.submitProviderEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: '20px'}} size="large" onClick={() => this.deleteProvider()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default ProviderEditPage;
