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
import {Button, Card, Input, InputNumber, Radio, Row, Select, Switch} from "antd";
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
import {FormCol, FormCol2, FormRow} from "./Setting";

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
      .then((organization) => {
        this.setState({
          organization: organization,
        });
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
        <FormRow marginTop="10px">
          <FormCol label={i18next.t("general:Name")} tooltip={i18next.t("general:Name - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Input value={this.state.organization.name} disabled={this.state.organization.name === "built-in"} onChange={e => {
              this.updateOrganizationField("name", e.target.value);
            }} />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Display name")} tooltip={i18next.t("general:Display name - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Input value={this.state.organization.name} disabled={this.state.organization.name === "built-in"} onChange={e => {
              this.updateOrganizationField("name", e.target.value);
            }} />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Favicon")} tooltip={i18next.t("general:Favicon - Tooltip")}>
          </FormCol>
          <FormCol2>
            <FormRow>
              <FormCol label={i18next.t("general:URL")} tooltip={i18next.t("general:URL - Tooltip")} minWidth={70}>
              </FormCol>
              <FormCol2 span={23}>
                <Input prefix={<LinkOutlined />} value={this.state.organization.favicon} onChange={e => {
                  this.updateOrganizationField("favicon", e.target.value);
                }} />
              </FormCol2>
            </FormRow>
            <FormRow>
              <FormCol label={i18next.t("general:Preview")} span={(Setting.isMobile()) ? 22 : 1} minWidth={70}>
              </FormCol>
              <FormCol2>
                <a target="_blank" rel="noreferrer" href={this.state.organization.favicon}>
                  <img src={this.state.organization.favicon} alt={this.state.organization.favicon} height={90} style={{marginBottom: "20px"}} />
                </a>
              </FormCol2>
            </FormRow>
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("organization:Website URL")} tooltip={i18next.t("organization:Website URL - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Input prefix={<LinkOutlined />} value={this.state.organization.websiteUrl} onChange={e => {
              this.updateOrganizationField("websiteUrl", e.target.value);
            }} />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Password type")} tooltip={i18next.t("general:Password type - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Select virtual={false} style={{width: "100%"}} value={this.state.organization.passwordType} onChange={(value => {this.updateOrganizationField("passwordType", value);})}
              options={["plain", "salt", "md5-salt", "bcrypt", "pbkdf2-salt", "argon2id"].map(item => Setting.getOption(item, item))}
            />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Password salt")} tooltip={i18next.t("general:Password salt - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Input value={this.state.organization.passwordSalt} onChange={e => {
              this.updateOrganizationField("passwordSalt", e.target.value);
            }} />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Supported country codes")} tooltip={i18next.t("general:Supported country codes - Tooltip")}>
          </FormCol>
          <FormCol2>
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
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Default avatar")} tooltip={i18next.t("general:Default avatar - Tooltip")}>
          </FormCol>
          <FormCol2>
            <FormRow>
              <FormCol label={i18next.t("general:URL")} tooltip={i18next.t("general:URL - Tooltip")} minWidth={70}>
              </FormCol>
              <FormCol2 span={23}>
                <Input prefix={<LinkOutlined />} value={this.state.organization.defaultAvatar} onChange={e => {
                  this.updateOrganizationField("defaultAvatar", e.target.value);
                }} />
              </FormCol2>
            </FormRow>
            <FormRow>
              <FormCol label={i18next.t("general:Preview")} span={(Setting.isMobile()) ? 22 : 1} minWidth={70}>
              </FormCol>
              <FormCol2>
                <a target="_blank" rel="noreferrer" href={this.state.organization.defaultAvatar}>
                  <img src={this.state.organization.defaultAvatar} alt={this.state.organization.defaultAvatar} height={90} style={{marginBottom: "20px"}} />
                </a>
              </FormCol2>
            </FormRow>
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Default application")} tooltip={i18next.t("general:Default application - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Select virtual={false} style={{width: "100%"}} value={this.state.organization.defaultApplication} onChange={(value => {this.updateOrganizationField("defaultApplication", value);})}
              options={this.state.applications?.map((item) => Setting.getOption(item.name, item.name))}
            />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("organization:Tags")} tooltip={i18next.t("organization:Tags - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Select virtual={false} mode="tags" style={{width: "100%"}} value={this.state.organization.tags} onChange={(value => {this.updateOrganizationField("tags", value);})}>
              {
                this.state.organization.tags?.map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Master password")} tooltip={i18next.t("general:Master password - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Input value={this.state.organization.masterPassword} onChange={e => {
              this.updateOrganizationField("masterPassword", e.target.value);
            }} />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:Languages")} tooltip={i18next.t("general:Languages - Tooltip")}>
          </FormCol>
          <FormCol2>
            <Select virtual={false} mode="tags" style={{width: "100%"}}
              options={Setting.Countries.map((item) => {
                return Setting.getOption(item.label, item.key);
              })}
              value={this.state.organization.languages ?? []}
              onChange={(value => {
                this.updateOrganizationField("languages", value);
              })} >
            </Select>
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("organization:Init score")} tooltip={i18next.t("organization:Init score - Tooltip")} span={(Setting.isMobile()) ? 19 : 2}>
          </FormCol>
          <FormCol2 span={4}>
            <InputNumber value={this.state.organization.initScore} onChange={value => {
              this.updateOrganizationField("initScore", value);
            }} />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("organization:Soft deletion")} tooltip={i18next.t("organization:Soft deletion - Tooltip")} span={(Setting.isMobile()) ? 19 : 2}>
          </FormCol>
          <FormCol2 span={1}>
            <Switch checked={this.state.organization.enableSoftDeletion} onChange={checked => {
              this.updateOrganizationField("enableSoftDeletion", checked);
            }} />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("organization:Is profile public")} tooltip={i18next.t("organization:Is profile public - Tooltip")} span={(Setting.isMobile()) ? 19 : 2}>
          </FormCol>
          <FormCol2 span={1}>
            <Switch checked={this.state.organization.isProfilePublic} onChange={checked => {
              this.updateOrganizationField("isProfilePublic", checked);
            }} />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("organization:Account items")} tooltip={i18next.t("organization:Account items - Tooltip")} >
          </FormCol>
          <FormCol2>
            <AccountTable
              title={i18next.t("organization:Account items")}
              table={this.state.organization.accountItems}
              onUpdateTable={(value) => {this.updateOrganizationField("accountItems", value);}}
            />
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("theme:Theme")} tooltip={i18next.t("theme:Theme - Tooltip")} >
          </FormCol>
          <FormCol2 marginTop={"5px"}>
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
          </FormCol2>
        </FormRow>
        <FormRow>
          <FormCol label={i18next.t("general:LDAPs")} tooltip={i18next.t("general:LDAPs - Tooltip")} >
          </FormCol>
          <FormCol2>
            <LdapTable
              title={i18next.t("general:LDAPs")}
              table={this.state.ldaps}
              organizationName={this.state.organizationName}
              onUpdateTable={(value) => {
                this.setState({ldaps: value});
              }}
            />
          </FormCol2>
        </FormRow>
      </Card>
    );
  }

  submitOrganizationEdit(willExist) {
    const organization = Setting.deepCopy(this.state.organization);
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
