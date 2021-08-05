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
import {Button, Card, Col, Input, Row, Select} from 'antd';
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as LdapBackend from "./backend/LdapBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import {LinkOutlined} from "@ant-design/icons";
import LdapTable from "./LdapTable";
import AccountTable from "./AccountTable";
import UserEditPage from "./UserEditPage";

const { Option } = Select;

class OrganizationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.match.params.organizationName,
      organization: null,
      ldaps: null,
      previewRole: "normal-user"
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrganization();
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

  getLdaps() {
    LdapBackend.getLdaps(this.state.organizationName)
      .then(res => {
        let resdata = []
        if (res.status === "ok") {
          if (res.data !== null) {
            resdata = res.data;
          }
        }
        this.setState({
          ldaps: resdata
        })
      })
  }

  parseOrganizationField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateOrganizationField(key, value) {
    value = this.parseOrganizationField(key, value);

    let organization = this.state.organization;
    organization[key] = value;
    this.setState({
      organization: organization,
    });
  }

  renderPreview() {
    return (
      <React.Fragment>
        <Col offset={2} span={20} style={{marginTop: "10px"}}>
          <Button type="primary" style={{width: "160px"}}>
            <a target="_blank" rel="noreferrer" href={"/account"}>{i18next.t("organization:Test user edit page..")}</a>
          </Button>
          <Select value={this.state.previewRole || undefined}
                  onChange={value => this.setState({previewRole: value})}
                  style={{width: "140px", marginLeft: "10px"}}>
            <Option value="admin">Admin</Option>
            <Option value="normal-user">Normal User</Option>
          </Select>
        </Col>
        <Col offset={2} span={20} style={{display: "flex", flexDirection: "column", marginTop: "10px"}}>
          <div style={{
            marginTop: "10px",
            border: "1px solid rgb(217,217,217)",
            boxShadow: "10px 10px 5px #888888",
            alignItems: "center",
            overflow: "auto",
            flexDirection: "column",
            flex: "auto"
          }}>
            <UserEditPage organizationName={this.props.account.owner}
                          accountItems={this.state.organization.accountItems} userName={this.props.account.name}
                          account={this.state.previewRole === "admin" ? this.props.account : ""}  isPreview={true}/>
          </div>
        </Col>
      </React.Fragment>
    )
  }

  renderOrganization() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("organization:Edit Organization")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitOrganizationEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={(Setting.isMobile())? {margin: '5px'}:{}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.name} onChange={e => {
              this.updateOrganizationField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.displayName} onChange={e => {
              this.updateOrganizationField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel("Favicon", i18next.t("general:Favicon - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                URL:
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined/>} value={this.state.organization.favicon} onChange={e => {
                  this.updateOrganizationField('favicon', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" rel="noreferrer" href={this.state.organization.favicon}>
                  <img src={this.state.organization.favicon} alt={this.state.organization.favicon} height={90} style={{marginBottom: '20px'}}/>
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("organization:Website URL"), i18next.t("organization:Website URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.organization.websiteUrl} onChange={e => {
              this.updateOrganizationField('websiteUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Password type"), i18next.t("general:Password type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.organization.passwordType} onChange={(value => {this.updateOrganizationField('passwordType', value);})}>
              {
                ['plain', 'salt']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Password salt"), i18next.t("general:Password salt - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.passwordSalt} onChange={e => {
              this.updateOrganizationField('passwordSalt', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Phone Prefix"), i18next.t("general:Phone Prefix - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input addonBefore={"+"} value={this.state.organization.phonePrefix} onChange={e => {
              this.updateOrganizationField('phonePrefix', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Default avatar"), i18next.t("general:Default avatar - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                URL:
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined/>} value={this.state.organization.defaultAvatar} onChange={e => {
                  this.updateOrganizationField('defaultAvatar', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
                {i18next.t("general:Preview")}:
              </Col>
              <Col span={23} >
                <a target="_blank" rel="noreferrer" href={this.state.organization.defaultAvatar}>
                  <img src={this.state.organization.defaultAvatar} alt={this.state.organization.defaultAvatar} height={90} style={{marginBottom: '20px'}}/>
                </a>
              </Col>
            </Row>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("organization:Account items"), i18next.t("organization:Account items - Tooltip"))} :
          </Col>
          <Col span={22} >
            <AccountTable
                title={i18next.t("application:Account items")}
                table={this.state.organization.accountItems}
                onUpdateTable={(value) => { this.updateOrganizationField('accountItems', value)}}
            />
            {
              this.renderPreview()
            }
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}}>
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:LDAPs"), i18next.t("general:LDAPs - Tooltip"))} :
          </Col>
          <Col span={22}>
            <LdapTable
              title={i18next.t("general:LDAPs")}
              table={this.state.ldaps}
              organizationName={this.state.organizationName}
              onUpdateTable={(value) => {
                this.setState({ldaps: value}) }}
            />
          </Col>
        </Row>
      </Card>
    )
  }

  submitOrganizationEdit() {
    let organization = Setting.deepCopy(this.state.organization);
    OrganizationBackend.updateOrganization(this.state.organization.owner, this.state.organizationName, organization)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            organizationName: this.state.organization.name,
          });
          this.props.history.push(`/organizations/${this.state.organization.name}`);
        } else {
          Setting.showMessage("error", res.msg);
          this.updateOrganizationField('name', this.state.organizationName);
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
          this.state.organization !== null ? this.renderOrganization() : null
        }
        <div style={{marginTop: '20px', marginLeft: '40px'}}>
          <Button type="primary" size="large" onClick={this.submitOrganizationEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      </div>
    );
  }
}

export default OrganizationEditPage;
