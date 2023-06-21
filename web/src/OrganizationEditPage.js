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
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as LdapBackend from "./backend/LdapBackend";
import * as Setting from "./Setting";
import * as Conf from "./Conf";
import i18next from "i18next";
import {LinkOutlined} from "@ant-design/icons";
import LdapTable from "./table/LdapTable";
import AccountTable from "./table/AccountTable";
import ThemeEditor from "./common/theme/ThemeEditor";
import MfaTable from "./table/MfaTable";

const {Option} = Select;

class OrganizationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.match.params.organizationName,
      organization: null,
      applications: [],
      ldaps: null,
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrganization();
    this.getApplications();
    this.getLdaps();
  }

  getOrganization() {
    OrganizationBackend.getOrganization("admin", this.state.organizationName)
      .then((res) => {
        if (res.status === "ok") {
          const organization = res.data;
          if (organization === null) {
            this.props.history.push("/404");
            return;
          }

          this.setState({
            organization: organization,
          });
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  getApplications() {
    ApplicationBackend.getApplicationsByOrganization("admin", this.state.organizationName)
      .then((applications) => {
        this.setState({
          applications: applications,
        });
      });
  }

  getLdaps() {
    LdapBackend.getLdaps(this.state.organizationName)
      .then(res => {
        let resdata = [];
        if (res.status === "ok") {
          if (res.data !== null) {
            resdata = res.data;
          }
        }
        this.setState({
          ldaps: resdata,
        });
      });
  }

  parseOrganizationField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateOrganizationField(key, value) {
    value = this.parseOrganizationField(key, value);
    const organization = this.state.organization;
    organization[key] = value;
    this.setState({
      organization: organization,
    });
  }

  renderOrganization() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("organization:New Organization") : i18next.t("organization:Edit Organization")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitOrganizationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitOrganizationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteOrganization()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.name} disabled={this.state.organization.name === "built-in"} onChange={e => {
              this.updateOrganizationField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.displayName} onChange={e => {
              this.updateOrganizationField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Favicon"), i18next.t("general:Favicon - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined />} value={this.state.organization.favicon} onChange={e => {
                  this.updateOrganizationField("favicon", e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" rel="noreferrer" href={this.state.organization.favicon}>
                  <img src={this.state.organization.favicon} alt={this.state.organization.favicon} height={90} style={{marginBottom: "20px"}} />
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("organization:Website URL"), i18next.t("organization:Website URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.organization.websiteUrl} onChange={e => {
              this.updateOrganizationField("websiteUrl", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Password type"), i18next.t("general:Password type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.organization.passwordType} onChange={(value => {this.updateOrganizationField("passwordType", value);})}
              options={["plain", "salt", "md5-salt", "bcrypt", "pbkdf2-salt", "argon2id"].map(item => Setting.getOption(item, item))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Password salt"), i18next.t("general:Password salt - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.passwordSalt} onChange={e => {
              this.updateOrganizationField("passwordSalt", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Password complexity options"), i18next.t("general:Password complexity options - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select
              virtual={false}
              style={{width: "100%"}}
              mode="multiple"
              value={this.state.organization.passwordOptions}
              onChange={(value => {
                this.updateOrganizationField("passwordOptions", value);
              })}
              options={[
                {value: "AtLeast6", name: i18next.t("user:The password must have at least 6 characters")},
                {value: "AtLeast8", name: i18next.t("user:The password must have at least 8 characters")},
                {value: "Aa123", name: i18next.t("user:The password must contain at least one uppercase letter, one lowercase letter and one digit")},
                {value: "SpecialChar", name: i18next.t("user:The password must contain at least one special character")},
                {value: "NoRepeat", name: i18next.t("user:The password must not contain any repeated characters")},
              ].map((item) => Setting.getOption(item.name, item.value))}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Supported country codes"), i18next.t("general:Supported country codes - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode={"multiple"} style={{width: "100%"}} value={this.state.organization.countryCodes ?? []}
              onChange={value => {
                this.updateOrganizationField("countryCodes", value);
              }}
              filterOption={(input, option) => (option?.text ?? "").toLowerCase().includes(input.toLowerCase())}
            >
              {
                Setting.getCountryCodeData().map((country) => Setting.getCountryCodeOption(country))
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Languages"), i18next.t("general:Languages - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="multiple" style={{width: "100%"}}
              options={Setting.Countries.map((item) => {
                return Setting.getOption(item.label, item.key);
              })}
              value={this.state.organization.languages ?? []}
              onChange={(value => {
                this.updateOrganizationField("languages", value);
              })} >
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Default avatar"), i18next.t("general:Default avatar - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined />} value={this.state.organization.defaultAvatar} onChange={e => {
                  this.updateOrganizationField("defaultAvatar", e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" rel="noreferrer" href={this.state.organization.defaultAvatar}>
                  <img src={this.state.organization.defaultAvatar} alt={this.state.organization.defaultAvatar} height={90} style={{marginBottom: "20px"}} />
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Default application"), i18next.t("general:Default application - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.organization.defaultApplication} onChange={(value => {this.updateOrganizationField("defaultApplication", value);})}
              options={this.state.applications?.map((item) => Setting.getOption(item.name, item.name))
              } />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("organization:Tags"), i18next.t("organization:Tags - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: "100%"}} value={this.state.organization.tags} onChange={(value => {this.updateOrganizationField("tags", value);})}>
              {
                this.state.organization.tags?.map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Master password"), i18next.t("general:Master password - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.masterPassword} onChange={e => {
              this.updateOrganizationField("masterPassword", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("organization:Init score"), i18next.t("organization:Init score - Tooltip"))} :
          </Col>
          <Col span={4} >
            <InputNumber value={this.state.organization.initScore} onChange={value => {
              this.updateOrganizationField("initScore", value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("organization:Soft deletion"), i18next.t("organization:Soft deletion - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.organization.enableSoftDeletion} onChange={checked => {
              this.updateOrganizationField("enableSoftDeletion", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t("organization:Is profile public"), i18next.t("organization:Is profile public - Tooltip"))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.organization.isProfilePublic} onChange={checked => {
              this.updateOrganizationField("isProfilePublic", checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("organization:Account items"), i18next.t("organization:Account items - Tooltip"))} :
          </Col>
          <Col span={22} >
            <AccountTable
              title={i18next.t("organization:Account items")}
              table={this.state.organization.accountItems}
              onUpdateTable={(value) => {this.updateOrganizationField("accountItems", value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:MFA items"), i18next.t("general:MFA items - Tooltip"))} :
          </Col>
          <Col span={22} >
            <MfaTable
              title={i18next.t("general:MFA items")}
              table={this.state.organization.mfaItems ?? []}
              onUpdateTable={(value) => {this.updateOrganizationField("mfaItems", value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("theme:Theme"), i18next.t("theme:Theme - Tooltip"))} :
          </Col>
          <Col span={22} style={{marginTop: "5px"}}>
            <Row>
              <Radio.Group value={this.state.organization.themeData?.isEnabled ?? false} onChange={e => {
                const {_, ...theme} = this.state.organization.themeData ?? {...Conf.ThemeDefault, isEnabled: false};
                this.updateOrganizationField("themeData", {...theme, isEnabled: e.target.value});
              }} >
                <Radio.Button value={false}>{i18next.t("organization:Follow global theme")}</Radio.Button>
                <Radio.Button value={true}>{i18next.t("theme:Customize theme")}</Radio.Button>
              </Radio.Group>
            </Row>
            {
              this.state.organization.themeData?.isEnabled ?
                <Row style={{marginTop: "20px"}}>
                  <ThemeEditor themeData={this.state.organization.themeData} onThemeChange={(_, nextThemeData) => {
                    const {isEnabled} = this.state.organization.themeData ?? {...Conf.ThemeDefault, isEnabled: false};
                    this.updateOrganizationField("themeData", {...nextThemeData, isEnabled});
                  }} />
                </Row> : null
            }
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:LDAPs"), i18next.t("general:LDAPs - Tooltip"))} :
          </Col>
          <Col span={22}>
            <LdapTable
              title={i18next.t("general:LDAPs")}
              table={this.state.ldaps}
              organizationName={this.state.organizationName}
              onUpdateTable={(value) => {
                this.setState({ldaps: value});
              }}
            />
          </Col>
        </Row>
      </Card>
    );
  }

  submitOrganizationEdit(willExist) {
    const organization = Setting.deepCopy(this.state.organization);
    organization.accountItems = organization.accountItems?.filter(accountItem => accountItem.name !== "Please select an account item");

    OrganizationBackend.updateOrganization(this.state.organization.owner, this.state.organizationName, organization)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully saved"));

          if (this.props.account.organization.name === this.state.organizationName) {
            this.props.onChangeTheme(Setting.getThemeData(this.state.organization));
          }

          this.setState({
            organizationName: this.state.organization.name,
          });

          if (willExist) {
            this.props.history.push("/organizations");
          } else {
            this.props.history.push(`/organizations/${this.state.organization.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to save")}: ${res.msg}`);
          this.updateOrganizationField("name", this.state.organizationName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteOrganization() {
    OrganizationBackend.deleteOrganization(this.state.organization)
      .then((res) => {
        if (res.status === "ok") {
          this.props.history.push("/organizations");
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
          this.state.organization !== null ? this.renderOrganization() : null
        }
        <div style={{marginTop: "20px", marginLeft: "40px"}}>
          <Button size="large" onClick={() => this.submitOrganizationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" size="large" onClick={() => this.submitOrganizationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} size="large" onClick={() => this.deleteOrganization()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      </div>
    );
  }
}

export default OrganizationEditPage;
