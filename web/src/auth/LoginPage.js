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
import {Button, Checkbox, Col, Form, Input, Result, Row, Spin} from "antd";
import {LockOutlined, UserOutlined} from "@ant-design/icons";
import * as AuthBackend from "./AuthBackend";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Provider from "./Provider";
import * as Util from "./Util";
import * as Setting from "../Setting";
import {GithubLoginButton, GoogleLoginButton} from "react-social-login-buttons";
import FacebookLoginButton from "./FacebookLoginButton";
import QqLoginButton from "./QqLoginButton";
import DingTalkLoginButton from "./DingTalkLoginButton";
import GiteeLoginButton from "./GiteeLoginButton";
import WechatLoginButton from "./WechatLoginButton";
import WeiboLoginButton from "./WeiboLoginButton";
import i18next from "i18next";
import LinkedInLoginButton from "./LinkedInLoginButton";
import WeComLoginButton from "./WeComLoginButton";

class LoginPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      type: props.type,
      applicationName: props.applicationName !== undefined ? props.applicationName : (props.match === undefined ? null : props.match.params.applicationName),
      application: null,
      mode: props.mode !== undefined ? props.mode : (props.match === undefined ? null : props.match.params.mode), // "signup" or "signin"
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

    ApplicationBackend.getApplication("admin", this.state.applicationName)
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

  onUpdateAccount(account) {
    this.props.onUpdateAccount(account);
  }

  onFinish(values) {
    const application = this.getApplicationObj();
    const ths = this;
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

            if (Setting.hasPromptPage(application)) {
              AuthBackend.getAccount("")
                .then((res) => {
                  let account = null;
                  if (res.status === "ok") {
                    account = res.data;
                    account.organization = res.data2;

                    this.onUpdateAccount(account);

                    if (Setting.isPromptAnswered(account, application)) {
                      Setting.goToLink(`${oAuthParams.redirectUri}?code=${code}&state=${oAuthParams.state}`);
                    } else {
                      Setting.goToLinkSoft(ths, `/prompt/${application.name}?redirectUri=${oAuthParams.redirectUri}&code=${code}&state=${oAuthParams.state}`);
                    }
                  } else {
                    Setting.showMessage("error", `Failed to sign in: ${res.msg}`);
                  }
                });
            } else {
              Setting.goToLink(`${oAuthParams.redirectUri}?code=${code}&state=${oAuthParams.state}`);
            }

            // Util.showMessage("success", `Authorization code: ${res.data}`);
          }
        } else {
          Util.showMessage("error", `Failed to log in: ${res.msg}`);
        }
      });
  };

  getSigninButton(type) {
    const text = i18next.t("login:Sign in with {type}").replace("{type}", type);
    if (type === "GitHub") {
      return <GithubLoginButton text={text} align={"center"} />
    } else if (type === "Google") {
      return <GoogleLoginButton text={text} align={"center"} />
    } else if (type === "QQ") {
      return <QqLoginButton text={text} align={"center"} />
    } else if (type === "Facebook") {
      return <FacebookLoginButton text={text} align={"center"} />
    } else if (type === "Weibo") {
      return <WeiboLoginButton text={text} align={"center"} />
    } else if (type === "Gitee") {
      return <GiteeLoginButton text={text} align={"center"} />
    } else if (type === "WeChat") {
      return <WechatLoginButton text={text} align={"center"} />
    } else if (type === "DingTalk") {
      return <DingTalkLoginButton text={text} align={"center"} />
    } else if (type === "LinkedIn"){
      return <LinkedInLoginButton text={text} align={"center"} />
    } else if (type === "WeCom") {
      return <WeComLoginButton text={text} align={"center"} />
    }

    return text;
  }

  renderProviderLogo(provider, application, width, margin, size) {
    if (size === "small") {
      return (
        <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "signup")}>
          <img width={width} height={width} src={Provider.getAuthLogo(provider)} alt={provider.displayName} style={{margin: margin}} />
        </a>
      )
    } else {
      return (
        <div key={provider.displayName} style={{marginBottom: "10px"}}>
          <a href={Provider.getAuthUrl(application, provider, "signup")}>
            {
              this.getSigninButton(provider.type)
            }
          </a>
        </div>
      )
    }
  }

  isProviderVisible(providerItem) {
    if (this.state.mode === "signup") {
      return Setting.isProviderVisibleForSignUp(providerItem);
    } else {
      return Setting.isProviderVisibleForSignIn(providerItem);
    }
  }

  renderForm(application) {
    if (this.state.msg !== null) {
      return Util.renderMessage(this.state.msg)
    }

    if (this.state.mode === "signup" && !application.enableSignUp) {
      return (
        <Result
          status="error"
          title="Sign Up Error"
          subTitle={"The application does not allow to sign up new account"}
          extra={[
            <Link onClick={() => {
              Setting.goToLogin(this, application);
            }}>
              <Button type="primary" key="signin">
                Sign In
              </Button>
            </Link>
          ]}
        >
        </Result>
      )
    }

    if (application.enablePassword) {
      return (
        <Form
          name="normal_login"
          initialValues={{
            organization: application.organization,
            application: application.name,
            remember: true
          }}
          onFinish={(values) => {this.onFinish(values)}}
          style={{width: "250px"}}
          size="large"
        >
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="application"
            rules={[
              {
                required: true,
                message: 'Please input your application!',
              },
            ]}
          >
          </Form.Item>
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="organization"
            rules={[
              {
                required: true,
                message: 'Please input your organization!',
              },
            ]}
          >
          </Form.Item>
          <Form.Item
            name="username"
            rules={[{ required: true, message: i18next.t("login:Please input your username, Email or phone!") }]}
          >
            <Input
              prefix={<UserOutlined className="site-form-item-icon" />}
              placeholder={i18next.t("login:username, Email or phone")}
              disabled={!application.enablePassword}
            />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[{ required: true, message: i18next.t("login:Please input your password!") }]}
          >
            <Input
              prefix={<LockOutlined className="site-form-item-icon" />}
              type="password"
              placeholder={i18next.t("login:Password")}
              disabled={!application.enablePassword}
            />
          </Form.Item>
          <Form.Item>
            <Form.Item name="autoSignin" valuePropName="checked" noStyle>
              <Checkbox style={{float: "left"}} disabled={!application.enablePassword}>
                {i18next.t("login:Auto login")}
              </Checkbox>
            </Form.Item>
            <a style={{float: "right"}} onClick={() => {
              Setting.goToForget(this, application);
            }}>
              {i18next.t("login:Forgot password?")}
            </a>
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              style={{width: "100%", marginBottom: '5px'}}
              disabled={!application.enablePassword}
            >
              {i18next.t("login:Sign In")}
            </Button>
            {
              !application.enableSignUp ? null : this.renderFooter(application)
            }
          </Form.Item>
          <Form.Item>
            {
              application.providers.filter(providerItem => this.isProviderVisible(providerItem)).map(providerItem => {
                return this.renderProviderLogo(providerItem.provider, application, 30, 5, "small");
              })
            }
          </Form.Item>
        </Form>
      );
    } else {
      return (
        <div style={{marginTop: "20px"}}>
          <div style={{fontSize: 16, textAlign: "left"}}>
            {i18next.t("login:To access")}&nbsp;
            <a target="_blank" rel="noreferrer" href={application.homepageUrl}>
              <span style={{fontWeight: "bold"}}>
                {application.displayName}
              </span>
            </a>
            :
          </div>
          <br/>
          {
            application.providers.filter(providerItem => this.isProviderVisible(providerItem)).map(providerItem => {
              return this.renderProviderLogo(providerItem.provider, application, 40, 10, "big");
            })
          }
          {
            !application.enableSignUp ? null : (
              <div>
                <br/>
                {
                  this.renderFooter(application)
                }
              </div>
            )
          }
        </div>
      )
    }
  }

  renderFooter(application) {
    if (this.state.mode === "signup") {
      return (
        <div style={{float: "right"}}>
          {i18next.t("signup:Have account?")}&nbsp;
          <Link onClick={() => {
            Setting.goToLogin(this, application);
          }}>
            {i18next.t("signup:sign in now")}
          </Link>
        </div>
      )
    } else {
      return (
        <div style={{float: "right"}}>
          {i18next.t("login:No account yet?")}&nbsp;
          <a onClick={() => {
            Setting.goToSignup(this, application);
          }}>
            {i18next.t("login:sign up now")}
          </a>
        </div>
      )
    }
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return Util.renderMessageLarge(this, this.state.msg);
    }

    let defaultProvider = null;
    application.providers?.forEach(provider => {
      if (provider.isDefault) {
        defaultProvider = provider;
      }
    })

    console.log(this.props.application)

    if (this.props.application === undefined && defaultProvider !== null) {
      return (
        <React.Fragment load={window.location.href = Provider.getAuthUrl(application, defaultProvider.provider, "signup")}>
          {/*TODO: Friendly transition page?*/}
          {/*<Result*/}
          {/*  style={{marginTop: "20vh"}}*/}
          {/*  icon={<Spin />}*/}
          {/*  title={`Login Casdoor with ${defaultProvider.provider.type}`}*/}
          {/*/>*/}
        </React.Fragment>
      )
    } else {
      return (
        <Row>
          <Col span={24} style={{display: "flex", justifyContent: "center"}}>
            <div style={{marginTop: "80px", marginBottom: "50px", textAlign: "center"}}>
              {
                Setting.renderHelmet(application)
              }
              {
                Setting.renderLogo(application)
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
}

export default LoginPage;
