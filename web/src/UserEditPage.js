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
import * as UserBackend from "./backend/UserBackend";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as Setting from "./Setting";
import {LinkOutlined} from "@ant-design/icons";
import i18next from "i18next";
import CropperDiv from "./CropperDiv.js";
import * as AuthBackend from "./auth/AuthBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as Provider from "./auth/Provider";

const { Option } = Select;

class UserEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.organizationName !== undefined ? props.organizationName : props.match.params.organizationName,
      userName: props.userName !== undefined ? props.userName : props.match.params.userName,
      user: null,
      application: null,
      organizations: [],
      providers: [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getUser();
    this.getOrganizations();
    this.getDefaultApplication();
    this.getDefaultProviders();
  }

  getUser() {
    UserBackend.getUser(this.state.organizationName, this.state.userName)
      .then((user) => {
        this.setState({
          user: user,
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

  getDefaultApplication() {
    ApplicationBackend.getDefaultApplication("admin")
      .then((application) => {
        this.setState({
          application: application,
        });
      });
  }

  getDefaultProviders() {
    ProviderBackend.getDefaultProviders("admin")
      .then((res) => {
        this.setState({
          providers: res,
        });
      });
  }

  parseUserField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateUserField(key, value) {
    value = this.parseUserField(key, value);

    let user = this.state.user;
    user[key] = value;
    this.setState({
      user: user,
    });
  }

  unlinkUser(providerType) {
    const body = {
      providerType: providerType,
    };
    AuthBackend.unlink(body)
      .then((res) => {
        if (res.status === 'ok') {
          Setting.showMessage("success", `Linked successfully`);

          this.getUser();
        } else {
          Setting.showMessage("error", `Failed to unlink: ${res.msg}`);
        }
      });
  }

  getProviderLink(provider, linkedValue) {
    if (provider.type === "GitHub") {
      return `https://github.com/${linkedValue}`;
    } else if (provider.type === "Google") {
      return "https://mail.google.com";
    } else {
      return "";
    }
  }

  renderIdp(provider) {
    const linkedValue = this.state.user[provider.type.toLowerCase()];

    return (
      <Row style={{marginTop: '20px'}} >
        <Col style={{marginTop: '5px'}} span={2}>
          {
            Setting.getProviderLogo(provider)
          }
          <span style={{marginLeft: '5px'}}>
            {
              `${provider.type}:`
            }
          </span>
        </Col>
        <Col span={22} >
          <span style={{width: '200px', display: "inline-block"}}>
            {
              linkedValue === "" ? (
                "(empty)"
              ) : (
                <a target="_blank" rel="noreferrer" href={this.getProviderLink(provider, linkedValue)}>
                  {
                    linkedValue
                  }
                </a>
              )
            }
          </span>
          {
            linkedValue === "" ? (
              <a key={provider.displayName} href={Provider.getAuthUrl(this.state.application, provider, "link")}>
                <Button style={{marginLeft: '20px', width: '80px'}} type="primary">Link</Button>
              </a>
            ) : (
              <Button style={{marginLeft: '20px', width: '80px'}} onClick={() => this.unlinkUser(provider.type)}>Unlink</Button>
            )
          }
        </Col>
      </Row>
    )
  }

  renderUser() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("user:Edit User")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitUserEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Organization")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.user.owner} onChange={(value => {this.updateUserField('owner', value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            ID:
          </Col>
          <Col span={22} >
            <Input value={this.state.user.id} disabled={true} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.user.name} disabled={true} onChange={e => {
              this.updateUserField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Display Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.user.displayName} onChange={e => {
              this.updateUserField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Avatar")}:
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                URL:
              </Col>
              <Col span={22} >
                <Input prefix={<LinkOutlined/>} value={this.state.user.avatar} onChange={e => {
                  this.updateUserField('avatar', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={22} >
                <a target="_blank" rel="noreferrer" href={this.state.user.avatar}>
                  <img src={this.state.user.avatar} alt={this.state.user.avatar} height={90} style={{marginBottom: '20px'}}/>
                </a>
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}}>
              <CropperDiv buttonText={"Upload a phone..."} title={"Crop your new profile picture"} targetFunction={UserBackend.uploadAvatar} />
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:User Type")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.user.type} onChange={(value => {this.updateUserField('type', value);})}>
              {
                ['normal-user']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Password Type")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.user.passwordType} onChange={(value => {this.updateUserField('passwordType', value);})}>
              {
                ['plain']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Password")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.user.password} onChange={e => {
              this.updateUserField('password', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Email")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.user.email} onChange={e => {
              this.updateUserField('email', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Phone")}:
          </Col>
          <Col span={22} >
            <Input addonBefore={
              <Select virtual={false} style={{width: 70}}  value={this.state.user.phonePrefix} onChange={(value => {this.updateUserField('phonePrefix', value);})}>
                <Option value="1">+1</Option>
                <Option value="86">+86</Option>
              </Select>
            } value={this.state.user.phone} onChange={e => {
              this.updateUserField('phone', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("user:Affiliation")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.user.affiliation} onChange={e => {
              this.updateUserField('affiliation', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("user:Tag")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.user.tag} onChange={e => {
              this.updateUserField('tag', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("user:Third-party Logins")}:
          </Col>
          <Col span={22} >
            <div style={{marginBottom: 20}}>
              {
                this.state.providers.map((provider, index) => this.renderIdp(provider))
              }
            </div>
          </Col>
        </Row>
        {
          !Setting.isAdminUser(this.props.account) ? null : (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {i18next.t("user:Is Admin")}:
                </Col>
                <Col span={1} >
                  <Switch checked={this.state.user.isAdmin} onChange={checked => {
                    this.updateUserField('isAdmin', checked);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {i18next.t("user:Is Global Admin")}:
                </Col>
                <Col span={1} >
                  <Switch checked={this.state.user.isGlobalAdmin} onChange={checked => {
                    this.updateUserField('isGlobalAdmin', checked);
                  }} />
                </Col>
              </Row>
            </React.Fragment>
          )
        }
      </Card>
    )
  }

  submitUserEdit() {
    let user = Setting.deepCopy(this.state.user);
    UserBackend.updateUser(this.state.organizationName, this.state.userName, user)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            organizationName: this.state.user.owner,
            userName: this.state.user.name,
          });

          if (this.props.history !== undefined) {
            this.props.history.push(`/users/${this.state.user.owner}/${this.state.user.name}`);
          }
        } else {
          Setting.showMessage("error", res.msg);
          this.updateUserField('owner', this.state.organizationName);
          this.updateUserField('name', this.state.userName);
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
              this.state.user !== null ? this.renderUser() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitUserEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default UserEditPage;
