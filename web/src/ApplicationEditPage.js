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
import {Button, Card, Col, Input, Row, Select, Switch} from 'antd';
import {LinkOutlined} from "@ant-design/icons";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as Setting from "./Setting";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import SignupPage from "./auth/SignupPage";
import LoginPage from "./auth/LoginPage";
import i18next from "i18next";
import UrlTable from "./UrlTable";
import ProviderTable from "./ProviderTable";
import SignupTable from "./SignupTable";
import PromptPage from "./auth/PromptPage";

const { Option } = Select;

class ApplicationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applicationName: props.match.params.applicationName,
      application: null,
      organizations: [],
      providers: [],
      mfaProviders: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getApplication();
    this.getOrganizations();
    this.getProviders();
  }

  getApplication() {
    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        }, () => {this.updateMFAProviders(application?.providers?.map(item => item.provider))});
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

  getProviders() {
    ProviderBackend.getProviders("admin")
      .then((res) => {
        this.setState({
          providers: res,
        })
      });
  }

  parseApplicationField(key, value) {
    if (["expireInHours"].includes(key)) {
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

  updateMFAProviders(providers) {
    this.setState({
      mfaProviders: Setting.getDuplicatedArray(providers, [{category: "Email"}, {category: "SMS"}], "category")
    })
  }

  renderApplication() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("application:Edit Application")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitApplicationEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.name} onChange={e => {
              this.updateApplicationField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.displayName} onChange={e => {
              this.updateApplicationField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel("Logo", i18next.t("general:Logo - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={1}>
                URL:
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined/>} value={this.state.application.logo} onChange={e => {
                  this.updateApplicationField('logo', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={1}>
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
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Home"), i18next.t("general:Home - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.homepageUrl} onChange={e => {
              this.updateApplicationField('homepageUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Description"), i18next.t("general:Description - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.description} onChange={e => {
              this.updateApplicationField('description', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
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
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.clientId} onChange={e => {
              this.updateApplicationField('clientId', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.application.clientSecret} onChange={e => {
              this.updateApplicationField('clientSecret', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
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
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Token expire"), i18next.t("general:Token expire - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input style={{width: "150px"}} value={this.state.application.expireInHours} suffix="Hours" onChange={e => {
              this.updateApplicationField('expireInHours', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("application:Password ON"), i18next.t("application:Password ON - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enablePassword} onChange={checked => {
              this.updateApplicationField('enablePassword', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("application:Enable signup"), i18next.t("application:Enable signup - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableSignUp} onChange={checked => {
              this.updateApplicationField('enableSignUp', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} align="middle">
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("application:Enable MFA"), i18next.t("application:Enable MFA - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.application.enableMfa && this.state.mfaProviders?.length > 0} onChange={checked => {
              this.updateApplicationField('enableMfa', checked);
            }}/>
          </Col>
          <Col span={2} >
            {
              !this.state.application.enableMfa || this.state.mfaProviders?.length === 0 ? <div style={{height: '32px'}}> </div> :
                <Select virtual={false} style={{width: '100%'}} value={this.state.application.mfa_method?this.state.application.mfa_method:this.state.mfaProviders[0]?.category} onChange={(value => {this.updateApplicationField('mfa_method', value);})}>
                  {
                    this.state.mfaProviders?.map((provider, index) => <Option key={index} value={provider.category}>{provider.category}</Option>)
                  }
                </Select>
            }
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Signup URL"), i18next.t("general:Signup URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.signupUrl} onChange={e => {
              this.updateApplicationField('signupUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Signin URL"), i18next.t("general:Signin URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.signinUrl} onChange={e => {
              this.updateApplicationField('signinUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Forget URL"), i18next.t("general:Forget URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.forgetUrl} onChange={e => {
              this.updateApplicationField('forgetUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Affiliation URL"), i18next.t("general:Affiliation URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.application.affiliationUrl} onChange={e => {
              this.updateApplicationField('affiliationUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Providers"), i18next.t("general:Providers - Tooltip"))} :
          </Col>
          <Col span={22} >
            <ProviderTable
              title={i18next.t("general:Providers")}
              table={this.state.application.providers}
              providers={this.state.providers}
              application={this.state.application}
              onUpdateTable={(value) => { this.updateApplicationField('providers', value);this.updateMFAProviders(value.map(item => item.provider))}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
          </Col>
          {
            this.renderPreview()
          }
        </Row>
        {
          !this.state.application.enableSignUp ? null : (
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
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
          <Col style={{marginTop: '5px'}} span={2}>
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

    return (
      <React.Fragment>
        <Col span={11} >
          <a style={{marginBottom: '10px'}} target="_blank" rel="noreferrer" href={signUpUrl}>
            <Button type="primary">{i18next.t("application:Test signup page..")}</Button>
          </a>
          <br/>
          <br/>
          <div style={{width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888"}}>
            {
              this.state.application.enablePassword ? (
                <SignupPage application={this.state.application} />
              ) : (
                <LoginPage type={"login"} mode={"signup"} application={this.state.application} />
              )
            }
          </div>
        </Col>
        <Col span={11} >
          <a style={{marginBottom: '10px'}} target="_blank" rel="noreferrer" href={signInUrl}>
            <Button type="primary">{i18next.t("application:Test signin page..")}</Button>
          </a>
          <br/>
          <br/>
          <div style={{width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888"}}>
            <LoginPage type={"login"} mode={"signin"} application={this.state.application} />
          </div>
        </Col>
      </React.Fragment>
    )
  }

  renderPreview2() {
    let promptUrl = `/prompt/${this.state.application.name}`;

    return (
      <React.Fragment>
        <Col span={11} >
          <a style={{marginBottom: '10px'}} target="_blank" rel="noreferrer" href={promptUrl}>
            <Button type="primary">{i18next.t("application:Test prompt page..")}</Button>
          </a>
          <br/>
          <br/>
          <div style={{width: "90%", border: "1px solid rgb(217,217,217)", boxShadow: "10px 10px 5px #888888"}}>
            <PromptPage application={this.state.application} account={this.props.account} />
          </div>
        </Col>
      </React.Fragment>
    )
  }

  submitApplicationEdit() {
    let application = Setting.deepCopy(this.state.application);
    ApplicationBackend.updateApplication(this.state.application.owner, this.state.applicationName, application)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            applicationName: this.state.application.name,
          });
          this.props.history.push(`/applications/${this.state.application.name}`);
        } else {
          Setting.showMessage("error", res.msg);
          this.updateApplicationField('name', this.state.applicationName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `Failed to connect to server: ${error}`);
      });
  }

  render() {
    return (
      <div>
        <Row style={{width: "100%"}}>
          <Col span={1}>
          </Col>
          <Col span={22}>
            {
              this.state.application !== null ? this.renderApplication() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitApplicationEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default ApplicationEditPage;
