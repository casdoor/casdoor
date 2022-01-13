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
import {Button, Card, Input, Select, Switch, Form, Space, Image} from 'antd';
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as LdapBackend from "./backend/LdapBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import {LinkOutlined} from "@ant-design/icons";
import LdapTable from "./LdapTable";

const { Option } = Select;

class OrganizationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.match.params.organizationName,
      organization: null,
      ldaps: null,
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

  renderOrganization() {
    return (
      <Card size="small" title={
        <Space size={10}>
          {i18next.t("organization:Edit Organization")}
          <Button onClick={() => this.submitOrganizationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button type="primary" onClick={() => this.submitOrganizationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </Space>
        }
        type="inner"
      >
        <Form
          labelCol={{ span: 3 }}
          wrapperCol={{ span: 21 }}
        >
          <Form.Item label={Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))}>
            <Input value={this.state.organization.name} disabled={this.state.organization.name === "built-in"}
              onChange={e => {
              this.updateOrganizationField('name', e.target.value);
            }} />
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))}>
            <Input value={this.state.organization.displayName}
              onChange={e => {
              this.updateOrganizationField('displayName', e.target.value);
            }} />
          </Form.Item>
          <Form.Item label={Setting.getLabel("Favicon", i18next.t("general:Favicon - Tooltip"))}>
            <Form.Item>
              <Input prefix={<LinkOutlined/>} value={this.state.organization.favicon}
                onChange={e => {
                this.updateOrganizationField('favicon', e.target.value);
              }} />
            </Form.Item>
            <Form.Item label={i18next.t("general:Preview")}>
              <Image
                src={this.state.organization.favicon}
                alt={this.state.organization.favicon}
                height={90}
              ></Image>
            </Form.Item>
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("organization:Website URL"), i18next.t("organization:Website URL - Tooltip"))}>
              <Input prefix={<LinkOutlined/>} value={this.state.organization.websiteUrl}
                onChange={e => {
                this.updateOrganizationField('websiteUrl', e.target.value);
              }} />
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("general:Password type"), i18next.t("general:Password type - Tooltip"))}>
              <Select virtual={false} style={{width: '100%'}} value={this.state.organization.passwordType}
                onChange={(value => {this.updateOrganizationField('passwordType', value);})}>
                {
                  ['plain', 'salt', 'md5-salt', 'bcrypt']
                    .map((item, index) => <Option key={index} value={item}>{item}</Option>)
                }
              </Select>
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("general:Password salt"), i18next.t("general:Password salt - Tooltip"))}>
              <Input value={this.state.organization.passwordSalt}
                onChange={e => {
                this.updateOrganizationField('passwordSalt', e.target.value);
              }} />
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("general:Phone prefix"), i18next.t("general:Phone prefix - Tooltip"))}>
              <Input addonBefore={"+"} value={this.state.organization.phonePrefix}
                onChange={e => {
                this.updateOrganizationField('phonePrefix', e.target.value);
              }} />
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("general:Default avatar"), i18next.t("general:Default avatar - Tooltip"))}>
              <Form.Item>
                <Input prefix={<LinkOutlined/>} value={this.state.organization.defaultAvatar}
                  onChange={e => {
                  this.updateOrganizationField('defaultAvatar', e.target.value);
                }} />
              </Form.Item>
              <Form.Item label={i18next.t("general:Preview")}>
                <Image
                  src={this.state.organization.defaultAvatar}
                  alt={this.state.organization.defaultAvatar}
                  height={90}
                ></Image>
              </Form.Item>
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("general:Master password"), i18next.t("general:Master password - Tooltip"))}>
            <Input value={this.state.organization.masterPassword}
              onChange={e => {
              this.updateOrganizationField('masterPassword', e.target.value);
            }} />
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("organization:Soft deletion"), i18next.t("organization:Soft deletion - Tooltip"))}>
            <Switch checked={this.state.organization.enableSoftDeletion}
              onChange={checked => {
              this.updateOrganizationField('enableSoftDeletion', checked);
            }} />
          </Form.Item>
          <Form.Item label={Setting.getLabel(i18next.t("general:LDAPs"), i18next.t("general:LDAPs - Tooltip"))}>
            <LdapTable
              title={i18next.t("general:LDAPs")}
              table={this.state.ldaps}
              organizationName={this.state.organizationName}
              onUpdateTable={(value) => {
                this.setState({ldaps: value}) }}
            />
          </Form.Item>
        </Form>
      </Card>
    )
  }

  submitOrganizationEdit(willExist) {
    let organization = Setting.deepCopy(this.state.organization);
    OrganizationBackend.updateOrganization(this.state.organization.owner, this.state.organizationName, organization)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            organizationName: this.state.organization.name,
          });

          if (willExist) {
            this.props.history.push(`/organizations`);
          } else {
            this.props.history.push(`/organizations/${this.state.organization.name}`);
          }
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
          <Button size="large" onClick={() => this.submitOrganizationEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" size="large" onClick={() => this.submitOrganizationEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
        </div>
      </div>
    );
  }
}

export default OrganizationEditPage;
