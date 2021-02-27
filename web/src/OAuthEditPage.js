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
import {Button, Card, Col, Input, Row} from 'antd';
import * as OAuthAppBackend from "./backend/OAuthAppBackend";
import * as Setting from "./Setting";
import i18next from "i18next";

class OAuthEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      oauthAppName: props.match.params.oauthAppName,
      app: null,
      account: props.account,
    };
  }

  componentWillMount() {
    this.getOAuthApp();
  }

  getOAuthApp() {
    OAuthAppBackend.getOAuthApp(this.state.account.name, this.state.oauthAppName)
      .then((app) => {
        this.setState({
          app: app,
        });
      });
  }

  parseOAuthAppField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateOAuthAppField(key, value) {
    value = this.parseOAuthAppField(key, value);

    let app = this.state.app;
    app[key] = value;
    this.setState({
      app: app,
    });
  }

  renderOAuthApp() {
    return (
      <Card size="small" title={
        <div>
          {i18next.t("oauth:Edit OAuth App")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button type="primary" onClick={this.submitOAuthAppEdit.bind(this)}>{i18next.t("general:Save")}</Button>
        </div>
      } style={{marginLeft: '5px'}} type="inner">
        <Row style={{marginTop: '10px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("general:Name")}: 
          </Col>
          <Col span={22} >
            <Input value={this.state.app.name} onChange={e => {
              this.updateOAuthAppField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("oauth:Homepage URL")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.app.domain} onChange={e => {
              this.updateOAuthAppField('domain', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {i18next.t("oauth:Callback URL")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.app.callback} onChange={e => {
              this.updateOAuthAppField('callback', e.target.value);
            }} />
          </Col>
        </Row>
      </Card>
    )
  }

  submitOAuthAppEdit() {
    let oauthApp = Setting.deepCopy(this.state.app);
    OAuthAppBackend.updateOAuthApp(this.state.app.clientId, oauthApp)
      .then((res) => {
        if (res) {
          Setting.showMessage("success", `Successfully saved`);
          this.setState({
            oauthAppName: this.state.app.name,
          });
          this.props.history.push(`/oauth`);
        } else {
          Setting.showMessage("error", `failed to save: server side failure`);
          this.updateOAuthAppField('name', this.state.oauthAppName);
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
              this.state.app !== null ? this.renderOAuthApp() : null
            }
          </Col>
          <Col span={1}>
          </Col>
        </Row>
        <Row style={{margin: 10}}>
          <Col span={2}>
          </Col>
          <Col span={18}>
            <Button type="primary" size="large" onClick={this.submitOAuthAppEdit.bind(this)}>{i18next.t("general:Save")}</Button>
          </Col>
        </Row>
      </div>
    );
  }
}

export default OAuthEditPage;