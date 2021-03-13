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
import * as TokenBackend from "./backend/TokenBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

const { Option } = Select;

class TokenEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      tokenName: props.match.params.tokenName,
      token: null,
    };
  }

  componentWillMount() {
    this.getToken();
  }

  getToken() {
    TokenBackend.getToken("admin", this.state.tokenName)
      .then((token) => {
        this.setState({
          token: token,
        });
      });
  }

  parseTokenField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateTokenField(key, value) {
    value = this.parseTokenField(key, value);

    let token = this.state.token;
    token[key] = value;
    this.setState({
      token: token,
    });
  }

  renderToken() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("token:Edit Token")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitTokenEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.token.name} onChange={e => {
              this.updateTokenField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Application")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.token.application} onChange={e => {
              this.updateTokenField('application', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Access Token")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.token.accessToken} onChange={e => {
              this.updateTokenField('accessToken', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Expires In")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.token.expiresIn} onChange={e => {
              this.updateTokenField('expiresIn', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Scope")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.token.scope} onChange={e => {
              this.updateTokenField('scope', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Token Type")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.token.tokenType} onChange={e => {
              this.updateTokenField('tokenType', e.target.value);
            }} />
          </Col>
        </Row>
      </Card>
    )
  }

  submitTokenEdit() {
    let token = Setting.deepCopy(this.state.token);
    TokenBackend.updateToken(this.state.token.owner, this.state.tokenName, token)
      .then((res) => {
        if (res) {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            tokenName: this.state.token.name,
          });
          this.props.history.push(`/tokens/${this.state.token.name}`);
        } else {
          Setting.showMessage("error", `failed to save: server side failure`);
          this.updateTokenField('name', this.state.tokenName);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `failed to save: ${error}`);
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
              this.state.token !== null ? this.renderToken() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitTokenEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default TokenEditPage;
