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
import {Alert, Button, Checkbox, Col, Form, Input, Result, Row} from "antd";
import {LockOutlined, UserOutlined} from "@ant-design/icons";
import * as AuthBackend from "./AuthBackend";
import * as Provider from "./Provider";
import * as Util from "./Util";

class Face extends React.Component {
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

  componentWillMount() {
    if (this.state.type === "login") {
      this.getApplication();
    } else if (this.state.type === "code") {
      this.getApplicationLogin();
    } else {
      Util.showMessage("error", `Unknown authentication type: ${this.state.type}`);
    }
  }

  getApplicationLogin() {
    const queries = new URLSearchParams(window.location.search);
    const clientId = queries.get("client_id");
    const responseType = queries.get("response_type");
    const redirectUri = queries.get("redirect_uri");
    const scope = queries.get("scope");
    const state = queries.get("state");
    AuthBackend.getApplicationLogin(clientId, responseType, redirectUri, scope, state)
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
    values.type = this.state.type;
    AuthBackend.login(values)
      .then((res) => {
        if (res.status === 'ok') {
          if (this.state.type === "login") {
            this.props.onLoggedIn();
            Util.showMessage("success", `Logged in successfully`);
            Util.goToLink("/");
          } else if (this.state.type === "code") {
            Util.showMessage("success", `Authorization code: ${res.data}`);
          }
        } else {
          Util.showMessage("error", `Log in failed：${res.msg}`);
        }
      });
  };

  renderMessage() {
    if (this.state.msg !== null) {
      return (
        <div style={{display: "inline"}}>
          <Alert
            message="Login Error"
            showIcon
            description={this.state.msg}
            type="error"
            action={
              <Button size="small" danger>
                Detail
              </Button>
            }
          />
        </div>
      )
    } else {
      return null;
    }
  }

  renderMessageLarge() {
    if (this.state.msg !== null) {
      return (
        <div style={{display: "inline"}}>
          <Result
            status="error"
            title="Login Error"
            subTitle={this.state.msg}
            extra={[
              <Button type="primary" key="details">
                Details
              </Button>,
              <Button key="help">Help</Button>,
            ]}
          >
          </Result>
        </div>
      )
    } else {
      return null;
    }
  }

  renderForm(application) {
    if (this.state.msg !== null) {
      return this.renderMessage()
    }

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
          <a style={{float: "right"}} href="">
            Forgot password?
          </a>
        </Form.Item>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            style={{width: "100%"}}
            disabled={!application.enablePassword}
          >
            Login
          </Button>
          <div style={{float: "right"}}>
            No account yet, <a href="/register">sign up now</a>
          </div>
        </Form.Item>
        <Form.Item>
          {
            application.providerObjs.map(provider => {
              return (
                <a href={Provider.getAuthUrl(application, provider, "signup")}>
                  <img width={30} height={30} src={Provider.getAuthLogo(provider)} alt={provider.displayName} style={{margin: "3px"}} />
                </a>
              );
            })
          }
        </Form.Item>
      </Form>
    );
  }

  renderLogo(application) {
    if (application.homepageUrl !== "") {
      return (
        <a target="_blank" href={application.homepageUrl}>
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
      return this.renderMessageLarge();
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

export default Face;
