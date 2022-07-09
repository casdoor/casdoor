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

import React from 'react';
import {Button, Card, Col, Input, Row, Select, Switch} from 'antd';
import * as OrganizationBackend from './backend/OrganizationBackend';
import * as LdapBackend from './backend/LdapBackend';
import * as Setting from './Setting';
import i18next from 'i18next';
import {LinkOutlined} from '@ant-design/icons';
import LdapTable from './LdapTable';
import AccountTable from './AccountTable';

const { Option } = Select;

class OrganizationEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      organizationName: props.match.params.organizationName,
      organization: null,
      ldaps: null,
      mode: props.location.mode !== undefined ? props.location.mode : 'edit',
    };
  }

  UNSAFE_componentWillMount() {
    this.getOrganization();
    this.getLdaps();
  }

  getOrganization() {
    OrganizationBackend.getOrganization('admin', this.state.organizationName)
      .then((organization) => {
        this.setState({
          organization: organization,
        });
      });
  }

  getLdaps() {
    LdapBackend.getLdaps(this.state.organizationName)
      .then(res => {
        let resdata = [];
        if (res.status === 'ok') {
          if (res.data !== null) {
            resdata = res.data;
          }
        }
        this.setState({
          ldaps: resdata
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

    let organization = this.state.organization;
    organization[key] = value;
    this.setState({
      organization: organization,
    });
  }

  renderOrganization() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === 'add' ? i18next.t('organization:New Organization') : i18next.t('organization:Edit Organization')}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitOrganizationEdit(false)}>{i18next.t('general:Save')}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" onClick={() => this.submitOrganizationEdit(true)}>{i18next.t('general:Save & Exit')}</Button>
          {this.state.mode === 'add' ? <Button style={{marginLeft: '20px'}} onClick={() => this.deleteOrganization()}>{i18next.t('general:Cancel')}</Button> : null}
        </div>
      } style={(Setting.isMobile())? {margin: '5px'}:{}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('general:Name'), i18next.t('general:Name - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.name} disabled={this.state.organization.name === 'built-in'} onChange={e => {
              this.updateOrganizationField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('general:Display name'), i18next.t('general:Display name - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.displayName} onChange={e => {
              this.updateOrganizationField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel( i18next.t('general:Favicon'), i18next.t('general:Favicon - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t('general:URL'), i18next.t('general:URL - Tooltip'))} :
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined/>} value={this.state.organization.favicon} onChange={e => {
                  this.updateOrganizationField('favicon', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 1}>
                {i18next.t('general:Preview')}:
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
            {Setting.getLabel(i18next.t('organization:Website URL'), i18next.t('organization:Website URL - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.organization.websiteUrl} onChange={e => {
              this.updateOrganizationField('websiteUrl', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('general:Password type'), i18next.t('general:Password type - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.organization.passwordType} onChange={(value => {this.updateOrganizationField('passwordType', value);})}>
              {
                ['plain', 'salt', 'md5-salt', 'bcrypt', 'pbkdf2-salt', 'argon2id']
                  .map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('general:Password salt'), i18next.t('general:Password salt - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.passwordSalt} onChange={e => {
              this.updateOrganizationField('passwordSalt', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('general:Phone prefix'), i18next.t('general:Phone prefix - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Input addonBefore={'+'} value={this.state.organization.phonePrefix} onChange={e => {
              this.updateOrganizationField('phonePrefix', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('general:Default avatar'), i18next.t('general:Default avatar - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 1}>
                {Setting.getLabel(i18next.t('general:URL'), i18next.t('general:URL - Tooltip'))} :
              </Col>
              <Col span={23} >
                <Input prefix={<LinkOutlined/>} value={this.state.organization.defaultAvatar} onChange={e => {
                  this.updateOrganizationField('defaultAvatar', e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: '20px'}} >
              <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 1}>
                {i18next.t('general:Preview')}:
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
            {Setting.getLabel(i18next.t('organization:Tags'), i18next.t('organization:Tags - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} mode="tags" style={{width: '100%'}} value={this.state.organization.tags} onChange={(value => {this.updateOrganizationField('tags', value);})}>
              {
                this.state.organization.tags?.map((item, index) => <Option key={index} value={item}>{item}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('general:Master password'), i18next.t('general:Master password - Tooltip'))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.organization.masterPassword} onChange={e => {
              this.updateOrganizationField('masterPassword', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t('organization:Soft deletion'), i18next.t('organization:Soft deletion - Tooltip'))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.organization.enableSoftDeletion} onChange={checked => {
              this.updateOrganizationField('enableSoftDeletion', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 19 : 2}>
            {Setting.getLabel(i18next.t('organization:Is profile public'), i18next.t('organization:Is profile public - Tooltip'))} :
          </Col>
          <Col span={1} >
            <Switch checked={this.state.organization.isProfilePublic} onChange={checked => {
              this.updateOrganizationField('isProfilePublic', checked);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('organization:Account items'), i18next.t('organization:Account items - Tooltip'))} :
          </Col>
          <Col span={22} >
            <AccountTable
              title={i18next.t('organization:Account items')}
              table={this.state.organization.accountItems}
              onUpdateTable={(value) => { this.updateOrganizationField('accountItems', value);}}
            />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}}>
          <Col style={{marginTop: '5px'}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t('general:LDAPs'), i18next.t('general:LDAPs - Tooltip'))} :
          </Col>
          <Col span={22}>
            <LdapTable
              title={i18next.t('general:LDAPs')}
              table={this.state.ldaps}
              organizationName={this.state.organizationName}
              onUpdateTable={(value) => {
                this.setState({ldaps: value}); }}
            />
          </Col>
        </Row>
      </Card>
    );
  }

  submitOrganizationEdit(willExist) {
    let organization = Setting.deepCopy(this.state.organization);
    OrganizationBackend.updateOrganization(this.state.organization.owner, this.state.organizationName, organization)
      .then((res) => {
        if (res.msg === '') {
          Setting.showMessage('success', 'Successfully saved');
          this.setState({
            organizationName: this.state.organization.name,
          });

          if (willExist) {
            this.props.history.push('/organizations');
          } else {
            this.props.history.push(`/organizations/${this.state.organization.name}`);
          }
        } else {
          Setting.showMessage('error', res.msg);
          this.updateOrganizationField('name', this.state.organizationName);
        }
      })
      .catch(error => {
        Setting.showMessage('error', `Failed to connect to server: ${error}`);
      });
  }

  deleteOrganization() {
    OrganizationBackend.deleteOrganization(this.state.organization)
      .then(() => {
        this.props.history.push('/organizations');
      })
      .catch(error => {
        Setting.showMessage('error', `Failed to connect to server: ${error}`);
      });
  }

  render() {
    return (
      <div>
        {
          this.state.organization !== null ? this.renderOrganization() : null
        }
        <div style={{marginTop: '20px', marginLeft: '40px'}}>
          <Button size="large" onClick={() => this.submitOrganizationEdit(false)}>{i18next.t('general:Save')}</Button>
          <Button style={{marginLeft: '20px'}} type="primary" size="large" onClick={() => this.submitOrganizationEdit(true)}>{i18next.t('general:Save & Exit')}</Button>
          {this.state.mode === 'add' ? <Button style={{marginLeft: '20px'}} size="large" onClick={() => this.deleteOrganization()}>{i18next.t('general:Cancel')}</Button> : null}
        </div>
      </div>
    );
  }
}

export default OrganizationEditPage;
