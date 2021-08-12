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
import * as ApplicationBackend from "./backend/ApplicationBackend";
import PasswordModal from "./PasswordModal";
import ResetModal from "./ResetModal";
import AffiliationSelect from "./common/AffiliationSelect";
import OAuthWidget from "./common/OAuthWidget";
import SelectRegionBox from "./SelectRegionBox";

import {Controlled as CodeMirror} from 'react-codemirror2';
import "codemirror/lib/codemirror.css";
require('codemirror/theme/material-darker.css');
require("codemirror/mode/javascript/javascript");

const {Option} = Select;

const accountMap = {
  "Organization": ["owner"],
  "ID": ["id"],
  "Name": ["name"],
  "Display name": ["displayName"],
  "Avatar": ["avatar"],
  "User type": ["type"],
  "Password": ["password"],
  "Email": ["email"],
  "Phone": ["phone"],
  "Country/Region": ["region"],
  "Affiliation": ["affiliation"],
  "Tag": ["tag"],
  "3rd-party logins": ["github", "google", "qq", "wechat", "facebook", "dingtalk", "weibo", "gitee", "linkedin", "wecom"],
}

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
      accountItems: props.isPreview ? props.accountItems : [],
    };
  }

  UNSAFE_componentWillMount() {
    this.getUser();
    this.getOrganizations();
    this.getUserApplication();
    let accountItems = [];
    if (Setting.isAdminUser(this.props.account) && !this.props.isPreview) {
      for (let accountItem in accountMap) {
        accountItems.push({"name": accountItem});
      }
    } else {
      accountItems = this.props.accountItems;
    }
    this.setState({accountItems: accountItems});
  }

  componentWillReceiveProps(nextProps, nextContent) {
    if (this.props.isPreview) {
      this.setState({accountItems: nextProps.accountItems});
    }
  }

  getUser() {
    UserBackend.getUser(this.state.organizationName, this.state.userName)
      .then((res) => {
        this.setState({
          user: res.data,
        });
        if (this.props.account === null) {
          this.setState({accountItems: res.data2})
        }
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

  renderRequired() {
    return (
      <span style={{marginRight: "4px", color: "#ff4d4f"}}>*</span>
    )
  }

  renderFormItem(accountItem) {
    if (Setting.isAdminUser(this.props.account)) {
      accountItem.visible = true;
      accountItem.editable = true;
    }
    if (!accountItem.visible) {
      return null;
    }

    switch (accountItem.name) {
      case "Organization":
        return (
          <Row style={{marginTop: '10px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
            </Col>
            <Col span={22}>
              <Select virtual={false} style={{width: '100%'}} disabled={!Setting.isAdminUser(this.props.account)}
                      value={this.state.user.owner} onChange={(value => {
                this.updateUserField('owner', value);
              })}>
                {
                  this.state.organizations.map((organization, index) => <Option key={index}
                                                                                value={organization.name}>{organization.name}</Option>)
                }
              </Select>
            </Col>
          </Row>
        )
      case "ID":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel("ID", i18next.t("general:ID - Tooltip"))} :
            </Col>
            <Col span={22}>
              <Input value={this.state.user.id} disabled={true}/>
            </Col>
          </Row>
        )
      case "Name":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
            </Col>
            <Col span={22}>
              <Input value={this.state.user.name} disabled={true} onChange={e => {
                this.updateUserField('name', e.target.value);
              }}/>
            </Col>
          </Row>
        )
      case "Display name":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
            </Col>
            <Col span={22}>
              <Input value={this.state.user.displayName} disabled={!accountItem.editable} onChange={e => {
                this.updateUserField('displayName', e.target.value);
              }}/>
            </Col>
          </Row>
        )
      case "Avatar":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("general:Avatar"), i18next.t("general:Avatar - Tooltip"))} :
            </Col>
            <Col span={22}>
              <Row style={{marginTop: '20px'}}>
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {i18next.t("general:URL")}:
                </Col>
                <Col span={22}>
                  <Input prefix={<LinkOutlined/>} value={this.state.user.avatar} disabled={!accountItem.editable}
                         onChange={e => {
                           this.updateUserField('avatar', e.target.value);
                         }}/>
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}}>
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {i18next.t("general:Preview")}:
                </Col>
                <Col span={22}>
                  <a target="_blank" rel="noreferrer" href={this.state.user.avatar}>
                    <img src={this.state.user.avatar} alt={this.state.user.avatar} height={90}
                         style={{marginBottom: '20px'}}/>
                  </a>
                </Col>
              </Row>
              {
                accountItem.editable ? (
                  <Row style={{marginTop: '20px'}}>
                    <CropperDiv buttonText={`${i18next.t("user:Upload a photo")}...`} title={i18next.t("user:Upload a photo")} user={this.state.user} />
                  </Row>) : null
              }
            </Col>
          </Row>
        )
      case "User type":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("general:User type"), i18next.t("general:User type - Tooltip"))} :
            </Col>
            <Col span={22}>
              <Select virtual={false} style={{width: '100%'}} value={this.state.user.type}
                      disabled={!accountItem.editable} onChange={(value => {
                this.updateUserField('type', value);
              })}>
                {
                  ['normal-user']
                    .map((item, index) => <Option key={index} value={item}>{item}</Option>)
                }
              </Select>
            </Col>
          </Row>
        )
      case "Password":
        return accountItem.editable ? (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("general:Password"), i18next.t("general:Password - Tooltip"))} :
            </Col>
            <Col span={22}>
              <PasswordModal user={this.state.user}/>
            </Col>
          </Row>
        ) : null
      case "Email":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("general:Email"), i18next.t("general:Email - Tooltip"))} :
            </Col>
            <Col style={{paddingRight: '20px'}} span={11}>
              <Input value={this.state.user.email} disabled/>
            </Col>
            {
              accountItem.editable ? (
                <Col span={11}>
                  {this.state.user.id === this.props.account?.id ? (
                    <ResetModal org={this.state.application?.organizationObj}
                                buttonText={i18next.t("user:Reset Email...")} destType={"email"}
                                coolDownTime={60}/>) : null}
                </Col>) : null
            }
          </Row>
        )
      case "Phone":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("general:Phone"), i18next.t("general:Phone - Tooltip"))} :
            </Col>
            <Col style={{paddingRight: '20px'}} span={11}>
              <Input value={this.state.user.phone}
                     addonBefore={`+${this.state.application?.organizationObj.phonePrefix}`} disabled/>
            </Col>
            {
              accountItem.editable ? (
                <Col span={11}>
                  {this.state.user.id === this.props.account?.id ? (
                    <ResetModal org={this.state.application?.organizationObj}
                                buttonText={i18next.t("user:Reset Phone...")}
                                destType={"phone"} coolDownTime={60}/>) : null}
                </Col>
              ) : null
            }
          </Row>
        )
      case "Country/Region":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("user:Country/Region"), i18next.t("user:Country/Region - Tooltip"))} :
            </Col>
            <Col span={22}>
              <SelectRegionBox defaultValue={this.state.user.region} onChange={(value) => {
                this.updateUserField("region", value);
              }} disabled={!accountItem.editable}/>
            </Col>
          </Row>
        )
      case "Affiliation":
        //TODO: Add prop "disabled" for AffiliationSelect
        if (!accountItem.editable) {
          return (
            <Row style={{marginTop: '20px'}}>
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                {accountItem.required ? this.renderRequired() : null}
                {Setting.getLabel(i18next.t("user:Affiliation"), i18next.t("user:Affiliation - Tooltip"))} :
              </Col>
              <Col span={22}>
                <Input value={this.state.user.affiliation} disabled={true}/>
              </Col>
            </Row>
          )
        } else {
          return (this.state.application === null || this.state.user === null) ? null : (
            <React.Fragment>
              <AffiliationSelect labelSpan={(Setting.isMobile()) ? 22 : 2}
                                 application={this.state.application}
                                 user={this.state.user}
                                 onUpdateUserField={(key, value) => {
                                   return this.updateUserField(key, value)
                                 }}
                                 required={accountItem.required}
              />
            </React.Fragment>
          )
        }
      case "Tag":
        return (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("user:Tag"), i18next.t("user:Tag - Tooltip"))} :
            </Col>
            <Col span={22}>
              <Input value={this.state.user.tag} disabled={!accountItem.editable} onChange={e => {
                this.updateUserField('tag', e.target.value);
              }}/>
            </Col>
          </Row>
        )
      case "3rd-party logins":
        return !this.isSelfOrAdmin() && !this.props.isPreview ? null : (
          <Row style={{marginTop: '20px'}}>
            <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
              {accountItem.required ? this.renderRequired() : null}
              {Setting.getLabel(i18next.t("user:3rd-party logins"), i18next.t("user:3rd-party logins - Tooltip"))} :
            </Col>
            <Col span={22}>
              <div style={{marginBottom: 20}}>
                {
                  (this.state.application === null || this.state.user === null) ? null : (
                    this.state.application?.providers.filter(providerItem => Setting.isProviderVisible(providerItem)).map((providerItem, index) =>
                      <OAuthWidget key={providerItem.name} labelSpan={(Setting.isMobile()) ? 10 : 3}
                                   user={this.state.user} application={this.state.application}
                                   providerItem={providerItem} disabled={!accountItem.editable} onUnlinked={() => {
                        return this.unlinked()
                      }}/>)
                  )
                }
              </div>
            </Col>
          </Row>
        )
      default:
        return null
    }
  }

  renderUser() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("user:Edit User")}&nbsp;&nbsp;&nbsp;&nbsp;
          {
            this.props.isPreview ? null : (
              <Button type="primary" onClick={this.submitUserEdit.bind(this)}>{i18next.t("general:Save")}</Button>
            )
          }
        </div>
      } style={(Setting.isMobile()) ? {margin: '5px'} : {}} type="inner">
        {
          this.state.accountItems?.map(accountItem => this.renderFormItem(accountItem))
        }
        {
          !Setting.isAdminUser(this.props.account) ? null : (
            <React.Fragment>
              {/*<Row style={{marginTop: '20px'}} >*/}
              {/*  <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>*/}
              {/*    {i18next.t("user:Properties")}:*/}
              {/*  </Col>*/}
              {/*  <Col span={22} >*/}
              {/*    <CodeMirror*/}
              {/*      value={JSON.stringify(this.state.user.properties, null, 4)}*/}
              {/*      options={{mode: 'javascript', theme: "material-darker"}}*/}
              {/*    />*/}
              {/*  </Col>*/}
              {/*</Row>*/}
              <Row style={{marginTop: '20px'}}>
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("user:Is admin"), i18next.t("user:Is admin - Tooltip"))} :
                </Col>
                <Col span={(Setting.isMobile()) ? 22 : 2}>
                  <Switch checked={this.state.user.isAdmin} onChange={checked => {
                    this.updateUserField('isAdmin', checked);
                  }}/>
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}}>
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("user:Is global admin"), i18next.t("user:Is global admin - Tooltip"))} :
                </Col>
                <Col span={(Setting.isMobile()) ? 22 : 2}>
                  <Switch checked={this.state.user.isGlobalAdmin} onChange={checked => {
                    this.updateUserField('isGlobalAdmin', checked);
                  }}/>
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}}>
                <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("user:Is forbidden"), i18next.t("user:Is forbidden - Tooltip"))} :
                </Col>
                <Col span={(Setting.isMobile()) ? 22 : 2}>
                  <Switch checked={this.state.user.isForbidden} onChange={checked => {
                    this.updateUserField('isForbidden', checked);
                  }}/>
                </Col>
              </Row>
            </React.Fragment>
          )
        }
      </Card>
    )
  }

  submitUserEdit() {
    let emptyFieldName;
    let user = Setting.deepCopy(this.state.user);
    try {
      this.state.accountItems.forEach(item => {
        if (item.required) {
          try {
            accountMap[item.name].forEach(userItem => {
              if (user[userItem].length === 0) {
                emptyFieldName = item.name;
                throw new Error();
              }
            })
          } finally {}
        }
      })
    } catch (e) {
      Setting.showMessage("error", i18next.t("user:Missing parameter."));
      return
    }
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
        {
          this.state.user !== null ? this.renderUser() : null
        }
        {
          this.props.isPreview ? null : (
            <div style={{marginTop: '20px', marginLeft: '40px'}}>
              <Button type="primary" size="large"
                      onClick={this.submitUserEdit.bind(this)}>{i18next.t("general:Save")}</Button>
            </div>
          )
        }
      </div>
    );
  }
}

export default UserEditPage;
