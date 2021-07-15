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
import CropperDiv from "./CropperDiv.js";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import PasswordModal from "./PasswordModal";
import ResetModal from "./ResetModal";
import AffiliationSelect from "./common/AffiliationSelect";
import OAuthWidget from "./common/OAuthWidget";

import "codemirror/lib/codemirror.css"
require('codemirror/theme/material-darker.css');
require("codemirror/mode/javascript/javascript");

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
    };
  }

  UNSAFE_componentWillMount() {
    this.getUser();
    this.getOrganizations();
    this.getUserApplication();
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

  getUserApplication() {
    ApplicationBackend.getUserApplication(this.state.organizationName, this.state.userName)
      .then((application) => {
        this.setState({
          application: application,
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

  unlinked() {
    this.getUser();
  }

  isSelfOrAdmin() {
    return (this.state.user.id === this.props.account?.id) || Setting.isAdminUser(this.props.account);
  }

  renderUser() {
    return (
      <Card size="small" title={
        <div>
          {Setting.I18n("user:Edit User")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitUserEdit.bind(this)}>{Setting.I18n("general:Save")}</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(Setting.I18n("general:Organization"), Setting.I18n("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.user.owner} onChange={(value => {this.updateUserField('owner', value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel("ID", Setting.I18n("general:ID - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.id} disabled={true} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(Setting.I18n("general:Name"), Setting.I18n("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.name} disabled={true} onChange={e => {
              this.updateUserField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(Setting.I18n("general:Display name"), Setting.I18n("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.displayName} onChange={e => {
              this.updateUserField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(Setting.I18n("general:Avatar"), Setting.I18n("general:Avatar - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                {Setting.I18n("general:URL")}:
              </Col>
              <Col span={22} >
                <Input prefix={<LinkOutlined/>} value={this.state.user.avatar} onChange={e => {
                  this.updateUserField('avatar', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                {Setting.I18n("general:Preview")}:
              </Col>
              <Col span={22} >
                <a target="_blank" rel="noreferrer" href={this.state.user.avatar}>
                  <img src={this.state.user.avatar} alt={this.state.user.avatar} height={90} style={{marginBottom: '20px'}}/>
                </a>
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}}>
              <CropperDiv buttonText={`${Setting.I18n("user:Upload a photo")}...`} title={Setting.I18n("user:Upload a photo")} targetFunction={UserBackend.uploadAvatar} />
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(Setting.I18n("general:User type"), Setting.I18n("general:User type - Tooltip"))} :
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
            {Setting.getLabel(Setting.I18n("general:Password"), Setting.I18n("general:Password - Tooltip"))} :
          </Col>
          <Col span={22} >
            <PasswordModal user={this.state.user} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(Setting.I18n("general:Email"), Setting.I18n("general:Email - Tooltip"))} :
          </Col>
          <Col style={{paddingRight: '20px'}} span={11} >
            <Input value={this.state.user.email} disabled />
          </Col>
          <Col span={11} >
            { this.state.user.id === this.props.account?.id ? (<ResetModal org={this.state.application?.organizationObj} buttonText={Setting.I18n("user:Reset Email...")} destType={"email"} coolDownTime={60}/>) : null}
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(Setting.I18n("general:Phone"), Setting.I18n("general:Phone - Tooltip"))} :
          </Col>
          <Col style={{paddingRight: '20px'}} span={11} >
            <Input value={this.state.user.phone} addonBefore={`+${this.state.application?.organizationObj.phonePrefix}`} disabled />
          </Col>
          <Col span={11} >
            { this.state.user.id === this.props.account?.id ? (<ResetModal org={this.state.application?.organizationObj} buttonText={Setting.I18n("user:Reset Phone...")} destType={"phone"} coolDownTime={60}/>) : null}
          </Col>
        </Row>
        {
          (this.state.application === null || this.state.user === null) ? null : (
            <AffiliationSelect labelSpan={2} application={this.state.application} user={this.state.user} onUpdateUserField={(key, value) => { return this.updateUserField(key, value)}} />
          )
        }
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(Setting.I18n("user:Tag"), Setting.I18n("user:Tag - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.user.tag} onChange={e => {
              this.updateUserField('tag', e.target.value);
            }} />
          </Col>
        </Row>
        {
          !this.isSelfOrAdmin() ? null : (
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={2}>
                {Setting.getLabel(Setting.I18n("user:Third-party logins"), Setting.I18n("user:Third-party logins - Tooltip"))} :
              </Col>
              <Col span={22} >
                <div style={{marginBottom: 20}}>
                  {
                    (this.state.application === null || this.state.user === null) ? null : (
                      this.state.application?.providers.filter(providerItem => Setting.isProviderVisible(providerItem)).map((providerItem, index) => <OAuthWidget key={providerItem.name} labelSpan={3} user={this.state.user} application={this.state.application} providerItem={providerItem} onUnlinked={() => { return this.unlinked()}} />)
                    )
                  }
                </div>
              </Col>
            </Row>
          )
        }
        {
          !Setting.isAdminUser(this.props.account) ? null : (
            <React.Fragment>
              {/*<Row style={{marginTop: '20px'}} >*/}
              {/*  <Col style={{marginTop: '5px'}} span={2}>*/}
              {/*    {Setting.I18n("user:Properties")}:*/}
              {/*  </Col>*/}
              {/*  <Col span={22} >*/}
              {/*    <CodeMirror*/}
              {/*      value={JSON.stringify(this.state.user.properties, null, 4)}*/}
              {/*      options={{mode: 'javascript', theme: "material-darker"}}*/}
              {/*    />*/}
              {/*  </Col>*/}
              {/*</Row>*/}
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(Setting.I18n("user:Is admin"), Setting.I18n("user:Is admin - Tooltip"))} :
                </Col>
                <Col span={1} >
                  <Switch checked={this.state.user.isAdmin} onChange={checked => {
                    this.updateUserField('isAdmin', checked);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(Setting.I18n("user:Is global admin"), Setting.I18n("user:Is global admin - Tooltip"))} :
                </Col>
                <Col span={1} >
                  <Switch checked={this.state.user.isGlobalAdmin} onChange={checked => {
                    this.updateUserField('isGlobalAdmin', checked);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(Setting.I18n("user:Is forbidden"), Setting.I18n("user:Is forbidden - Tooltip"))} :
                </Col>
                <Col span={1} >
                  <Switch checked={this.state.user.isForbidden} onChange={checked => {
                    this.updateUserField('isForbidden', checked);
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
            <Button type="primary" size="large" onClick={this.submitUserEdit.bind(this)}>{Setting.I18n("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default UserEditPage;
