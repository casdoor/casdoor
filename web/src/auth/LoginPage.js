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
import {Link} from "react-router-dom";
import {Button, Card, Checkbox, Col, Form, Input, Row} from "antd";
import {LockOutlined, UserOutlined} from "@ant-design/icons";
import * as AuthBackend from "./AuthBackend";
import * as Provider from "./Provider";
import * as Util from "./Util";
import * as Setting from "../Setting";

class LoginPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      type: props.type,
      applicationName: props.applicationName !== undefined ? props.applicationName : (props.match === undefined ? null : props.match.params.applicationName),
      application: null,
      msg: null,
    };
  }

  UNSAFE_componentWillMount() {
    if (this.state.type === "login") {
      this.getApplication();
    } else if (this.state.type === "code") {
      this.getApplicationLogin();
    } else {
      Util.showMessage("error", `Unknown authentication type: ${this.state.type}`);
    }
  }

  getApplicationLogin() {
    const oAuthParams = Util.getOAuthGetParameters();
    AuthBackend.getApplicationLogin(oAuthParams)
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            application: res.data,
          });
        } else {
          // Util.showMessage("error", res.msg);
          this.setState({
            application: res.data,
            msg: res.msg,
          });
        }
      });
  }

  getApplication() {
    if (this.state.applicationName === null) {
      return;
    }

    AuthBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });
      });
  }

  getApplicationObj() {
    if (this.props.application !== undefined) {
      return this.props.application;
    } else {
      return this.state.application;
    }
  }

  onFinish(values) {
    values["type"] = this.state.type;
    const oAuthParams = Util.getOAuthGetParameters();
    AuthBackend.login(values, oAuthParams)
      .then((res) => {
        if (res.status === 'ok') {
          const responseType = this.state.type;
          if (responseType === "login") {
            Util.showMessage("success", `Logged in successfully`);
            Setting.goToLink("/");
          } else if (responseType === "code") {
            const code = res.data;
            Setting.goToLink(`${oAuthParams.redirectUri}?code=${code}&state=${oAuthParams.state}`);
            // Util.showMessage("success", `Authorization code: ${res.data}`);
          }
        } else {
          Util.showMessage("error", `Failed to log in: ${res.msg}`);
        }
      });
  };

  renderProviderLogo(provider, application, width, margin, size) {
    if (size === "small") {
      return (
        <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "signup")}>
          <img width={width} height={width} src={Provider.getAuthLogo(provider)} alt={provider.displayName} style={{margin: margin}} />
        </a>
      )
    } else {
      return (
        <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "signup")}>
          <Card
            hoverable
            bodyStyle={{padding: 0}}
            style={{height: 60, marginBottom: 10, paddingTop: 10}}
          >
            <Row>
              <Col span={3}>
              </Col>
              <Col span={3}>
                <img width={width} height={width} src={Provider.getAuthLogo(provider)} alt={provider.displayName} />
              </Col>
              <Col span={18}>
                <div style={{marginTop: 8, fontWeight: 500, color: "#757575"}}>
                  Login by {provider.type}
                </div>
              </Col>
            </Row>

          </Card>
        </a>
      )
    }
  }

  renderForm(application) {
    if (this.state.msg !== null) {
      return Util.renderMessage(this.state.msg)
    }

    if (application.enablePassword) {
      return (
        <Form
          name="normal_login"
          initialValues={{
            organization: application.organization,
            remember: true
          }}
          onFinish={this.onFinish.bind(this)}
          style={{width: "250px"}}
          size="large"
        >
          <Form.Item style={{height: 0, visibility: "hidden"}}
                     name="organization"
                     rules={[{ required: true, message: 'Please input your organization!' }]}
          >
            <Input
              prefix={<UserOutlined className="site-form-item-icon" />}
              placeholder="organization"
              disabled={!application.enablePassword}
            />
          </Form.Item>
          <Form.Item
            name="username"
            rules={[{ required: true, message: 'Please input your Username!' }]}
          >
            <Input
              prefix={<UserOutlined className="site-form-item-icon" />}
              placeholder="username"
              disabled={!application.enablePassword}
            />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[{ required: true, message: 'Please input your Password!' }]}
          >
            <Input
              prefix={<LockOutlined className="site-form-item-icon" />}
              type="password"
              placeholder="password"
              disabled={!application.enablePassword}
            />
          </Form.Item>
          <Form.Item>
            <Form.Item name="remember" valuePropName="checked" noStyle>
              <Checkbox style={{float: "left"}} disabled={!application.enablePassword}>
                Auto login
              </Checkbox>
            </Form.Item>
            <Link style={{float: "right"}} to="/forgot">
              Forgot password?
            </Link>
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              style={{width: "100%"}}
              disabled={!application.enablePassword}
            >
              Sign In
            </Button>
            {
              !application.enableSignUp ? null : (
                <div style={{float: "right"}}>
                  No account yet?&nbsp;
                  <Link to={"/register"}>
                    sign up now
                  </Link>
                </div>
              )
            }
          </Form.Item>
          <Form.Item>
            {
              application.providerObjs.map(provider => {
                return this.renderProviderLogo(provider, application, 30, 5, "small");
              })
            }
          </Form.Item>
        </Form>
      );
    } else {
      return (
        <div style={{marginTop: "20px"}}>
          <div style={{fontSize: 16, textAlign: "left"}}>
            Please click to login&nbsp;
            <a target="_blank" rel="noreferrer" href={application.homepageUrl}>
              {application.displayName}
            </a>
            :
          </div>
          <br/>
          {
            application.providerObjs.map(provider => {
              return this.renderProviderLogo(provider, application, 40, 10, "big");
            })
          }
          {
            !application.enableSignUp ? null : (
              <div>
                <br/>
                <div style={{float: "right"}}>
                  No account yet?&nbsp;
                  <Link to={"/register"}>
                    sign up now
                  </Link>
                </div>
              </div>
            )
          }
        </div>
      )
    }
  }

  renderLogo(application) {
    if (application.homepageUrl !== "") {
      return (
        <a target="_blank" rel="noreferrer" href={application.homepageUrl}>
          <img width={250} src={application.logo} alt={application.displayName} style={{marginBottom: '30px'}}/>
        </a>
      )
    } else {
      return (
        <img width={250} src={application.logo} alt={application.displayName} style={{marginBottom: '30px'}}/>
      );
    }
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return Util.renderMessageLarge(this, this.state.msg);
    }

    return (
      <Row>
        <Col span={24} style={{display: "flex", justifyContent: "center"}}>
          <div style={{marginTop: "80px", textAlign: "center"}}>
            {
              this.renderLogo(application)
            }
            {/*{*/}
            {/*  this.state.clientId !== null ? "Redirect" : null*/}
            {/*}*/}
            {
              this.renderForm(application)
            }
          </div>
        </Col>
      </Row>
    )
  }
}

export default LoginPage;
