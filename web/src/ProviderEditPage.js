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
import * as ProviderBackend from "./backend/ProviderBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const { Option } = Select;

class ProviderEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      providerName: props.match.params.providerName,
      provider: null,
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
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
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

  renderProvider() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("provider:Edit Provider")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitProviderEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.name} onChange={e => {
              this.updateProviderField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Display Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.displayName} onChange={e => {
              this.updateProviderField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("provider:Type")}:
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.provider.type} onChange={(value => {this.updateProviderField('type', value);})}>
              {
                [
                  {id: 'google', name: 'Google'},
                  {id: 'github', name: 'GitHub'},
                  {id: 'qq', name: 'QQ'},
                  {id: 'wechat', name: 'WeChat'},
                ].map((providerType, index) => <Option key={index} value={providerType.id}>{providerType.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("provider:Client ID")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.clientId} onChange={e => {
              this.updateProviderField('clientId', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("provider:Client Secret")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.clientSecret} onChange={e => {
              this.updateProviderField('clientSecret', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("provider:Provider URL")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.providerUrl} onChange={e => {
              this.updateProviderField('providerUrl', e.target.value);
            }} />
          </Col>
        </Row>
      </Card>
    )
  }

  submitProviderEdit() {
    let provider = Setting.deepCopy(this.state.provider);
    ProviderBackend.updateProvider(this.state.provider.owner, this.state.providerName, provider)
      .then((res) => {
        if (res.msg === "") {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            providerName: this.state.provider.name,
          });
          this.props.history.push(`/providers/${this.state.provider.name}`);
        } else {
          Setting.showMessage("error", res.msg);
          this.updateProviderField('name', this.state.providerName);
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
              this.state.provider !== null ? this.renderProvider() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitProviderEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default ProviderEditPage;
