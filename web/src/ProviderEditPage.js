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
import {Button, Card, Col, Input, InputNumber, Row, Select, Switch} from 'antd';
import {LinkOutlined} from "@ant-design/icons";
import * as ProviderBackend from "./backend/ProviderBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import ProviderPropertiesTable from "./ProviderPropertiesTable";
import * as Provider from "./auth/Provider";
import {DangerousAreaStart, DangerousAreaEnd} from "./component/DangerousArea";

const { Option } = Select;
const { TextArea } = Input;

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
    if (["port"].includes(key)) {
      value = Setting.myParseInt(value);
    }
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

  getProviderTypeOptions(provider) {
    if (provider.category === "OAuth") {
      return (
        [
          {id: 'Other', name: 'Other'},
          {id: 'Google', name: 'Google'},
          {id: 'GitHub', name: 'GitHub'},
          {id: 'QQ', name: 'QQ'},
          {id: 'WeChat', name: 'WeChat'},
          {id: 'Facebook', name: 'Facebook'},
          {id: 'DingTalk', name: 'DingTalk'},
          {id: 'Weibo', name: 'Weibo'},
          {id: 'Gitee', name: 'Gitee'},
          {id: 'LinkedIn', name: 'LinkedIn'},
        ]
      );
    } else if (provider.category === "Email") {
      return (
        [
          {id: 'Default', name: 'Default'},
        ]
      );
    } else {
      return (
        [
          {id: 'aliyun', name: 'Aliyun'},
          {id: 'tencent', name: 'Tencent Cloud'},
        ]
      );
    }
  }

  getOAuthUrlTemplate(type) {
    if (type === "Google") {
      return `${Provider.GoogleAuthUri}?client_id=` + "${clientId}" + `&scope=${Provider.GoogleAuthScope}&response_type=code`;
    } else if (type === "GitHub") {
      return `${Provider.GithubAuthUri}?client_id=` + "${clientId}" + `&scope=${Provider.GithubAuthScope}&response_type=code`;
    } else if (type === "QQ") {
      return `${Provider.QqAuthUri}?client_id=` + "${clientId}" + `&scope=${Provider.QqAuthScope}&response_type=code`;
    } else if (type === "WeChat") {
      return `${Provider.WeChatAuthUri}?appid=` + "${clientId}" + `&scope=${Provider.WeChatAuthScope}&response_type=code#wechat_redirect`;
    } else if (type === "Facebook") {
      return `${Provider.FacebookAuthUri}?client_id=` + "${clientId}" + `&scope=${Provider.FacebookAuthScope}&response_type=code`;
    } else if (type === "DingTalk") {
      return `${Provider.DingTalkAuthUri}?appid=` + "${clientId}" + `&scope=snsapi_login&response_type=code`;
    } else if (type === "Weibo") {
      return `${Provider.WeiboAuthUri}?client_id=` + "${clientId}" + `&scope=${Provider.WeiboAuthScope}&response_type=code`;
    } else if (type === "Gitee") {
      return `${Provider.GiteeAuthUri}?client_id=` + "${clientId}" + `&scope=${Provider.GiteeAuthScope}&response_type=code`;
    } else if (type === "LinkedIn") {
      return `${Provider.LinkedInAuthUri}?client_id=` + "${clientId}" + `&scope=${Provider.LinkedInAuthScope}&response_type=code`
    }
    return "https://example.com/$client_id=${clientId}";
  }

  getOAuthScope(type) {
    if (type === "Google") {
      return Provider.GoogleAuthScope;
    } else if (type === "GitHub") {
      return Provider.GithubAuthScope;
    } else if (type === "QQ") {
      return Provider.QqAuthScope;
    } else if (type === "WeChat") {
      return Provider.WeChatAuthScope;
    } else if (type === "Facebook") {
      return Provider.FacebookAuthScope;
    } else if (type === "DingTalk") {
      return `snsapi_login`;
    } else if (type === "Weibo") {
      return Provider.WeiboAuthScope;
    } else if (type === "Gitee") {
      return Provider.GiteeAuthScope;
    } else if (type === "LinkedIn") {
      return Provider.LinkedInAuthScope;
    }

    return "scope"
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
            {Setting.getLabel(i18next.t("general:Name"), i18next.t("general:Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.name} onChange={e => {
              this.updateProviderField('name', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("general:Display name"), i18next.t("general:Display name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.displayName} onChange={e => {
              this.updateProviderField('displayName', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("provider:Category"), i18next.t("provider:Category - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: '100%'}} value={this.state.provider.category} onChange={(value => {
              this.updateProviderField('category', value);
              if (value === "OAuth") {
                this.updateProviderField('type', 'GitHub');
              } else if (value === "Email") {
                this.updateProviderField('type', 'Default');
                this.updateProviderField('title', 'Casdoor Verification Code');
                this.updateProviderField('content', 'You have requested a verification code at Casdoor. Here is your code: %s, please enter in 5 minutes.');
              } else if (value === "SMS") {
                this.updateProviderField('type', 'aliyun');
              }
            })}>
              {
                [
                  {id: 'OAuth', name: 'OAuth'},
                  {id: 'Email', name: 'Email'},
                  {id: 'SMS', name: 'SMS'},
                ].map((providerCategory, index) => <Option key={index} value={providerCategory.id}>{providerCategory.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("provider:Type"), i18next.t("provider:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select
                virtual={false}
                style={{width: '100%'}}
                value={this.state.provider.type}
                onChange={value => {
                  if (this.state.provider.category === "OAuth") {
                    this.updateProviderField('OAuthUrlTemplate', this.getOAuthUrlTemplate(value));
                    this.updateProviderField('scope', this.getOAuthScope(value));
                  }
                  this.updateProviderField('type', value);
                }}
            >
              {
                this.getProviderTypeOptions(this.state.provider).map((providerType, index) => <Option key={index} value={providerType.id}>{providerType.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {this.state.provider.category === "Email" ? Setting.getLabel(i18next.t("signup:Username"), i18next.t("signup:Username - Tooltip")) : Setting.getLabel(i18next.t("provider:Client ID"), i18next.t("provider:Client ID - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.clientId} onChange={e => {
              this.updateProviderField('clientId', e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {this.state.provider.category === "Email" ? Setting.getLabel(i18next.t("login:Password"), i18next.t("login:Password - Tooltip")) : Setting.getLabel(i18next.t("provider:Client secret"), i18next.t("provider:Client secret - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={this.state.provider.clientSecret} onChange={e => {
              this.updateProviderField('clientSecret', e.target.value);
            }} />
          </Col>
        </Row>
        {
          this.state.provider.category === "Email" ? (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input prefix={<LinkOutlined/>} value={this.state.provider.host} onChange={e => {
                    this.updateProviderField('host', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Port"), i18next.t("provider:Port - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <InputNumber value={this.state.provider.port} onChange={value => {
                    this.updateProviderField('port', value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Email Title"), i18next.t("provider:Email Title - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.title} onChange={e => {
                    this.updateProviderField('title', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Email Content"), i18next.t("provider:Email Content - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <TextArea autoSize={{minRows: 1, maxRows: 100}} value={this.state.provider.content} onChange={e => {
                    this.updateProviderField('content', e.target.value);
                  }} />
                </Col>
              </Row>
            </React.Fragment>
          ) : this.state.provider.category === "SMS" ? (
            <React.Fragment>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Region ID"), i18next.t("provider:Region ID - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.regionId} onChange={e => {
                    this.updateProviderField('regionId', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Sign Name"), i18next.t("provider:Sign Name - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.signName} onChange={e => {
                    this.updateProviderField('signName', e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: '20px'}} >
                <Col style={{marginTop: '5px'}} span={2}>
                  {Setting.getLabel(i18next.t("provider:Template Code"), i18next.t("provider:Template Code - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Input value={this.state.provider.templateCode} onChange={e => {
                    this.updateProviderField('templateCode', e.target.value);
                  }} />
                </Col>
              </Row>
            </React.Fragment>
          ) : null
        }
        {this.state.provider.category === "SMS" && this.state.provider.type === "tencent" ? (
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={2}>
              {Setting.getLabel(i18next.t("provider:App ID"), i18next.t("provider:App ID - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={this.state.provider.appId} onChange={e => {
                this.updateProviderField('appId', e.target.value);
              }} />
            </Col>
          </Row>
        ) : null}
        <Row style={{marginTop: '20px'}} >
          <Col style={{marginTop: '5px'}} span={2}>
            {Setting.getLabel(i18next.t("provider:Provider URL"), i18next.t("provider:Provider URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined/>} value={this.state.provider.providerUrl} onChange={e => {
              this.updateProviderField('providerUrl', e.target.value);
            }} />
          </Col>
        </Row>
        {this.state.provider.category === "OAuth" && this.state.provider.type !== "Other" ?
            <DangerousAreaStart /> : null}
        {this.state.provider.category === "OAuth" ? (<div>
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={2}>
              {Setting.getLabel(i18next.t("provider:Scope"), i18next.t("provider:Scope - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input prefix={<LinkOutlined/>} value={this.state.provider.scope} onChange={e => {
                this.updateProviderField('scope', e.target.value);
              }} />
            </Col>
          </Row>
          <Row style={{marginTop: '20px'}} >
            <Col style={{marginTop: '5px'}} span={2}>
              {Setting.getLabel(i18next.t("provider:OAuth Url Template"), i18next.t("provider:OAuthUrlTemplate - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input prefix={<LinkOutlined/>} value={this.state.provider.OAuthUrlTemplate} onChange={e => {
                this.updateProviderField('OAuthUrlTemplate', e.target.value);
              }} />
            </Col>
          </Row>
            </div>) : null}
        <ProviderPropertiesTable
            properties={this.state.provider.properties}
            onPropertyChange={properties => this.updateProviderField("properties", properties)}
        />
        {this.state.provider.category === "OAuth" && this.state.provider.type !== "Other" ?
            <DangerousAreaEnd /> : null}
      </Card>
    )
  }

  submitProviderEdit() {
    let provider = Setting.deepCopy(this.state.provider);
    delete provider.properties[""];
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
