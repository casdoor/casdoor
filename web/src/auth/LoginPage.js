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

import React from "react";
import {Link} from "react-router-dom";
import {Button, Checkbox, Col, Form, Input, Result, Row, Spin, Tabs} from "antd";
import {LockOutlined, UserOutlined} from "@ant-design/icons";
import * as UserWebauthnBackend from "../backend/UserWebauthnBackend";
import * as AuthBackend from "./AuthBackend";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Provider from "./Provider";
import * as ProviderButton from "./ProviderButton";
import * as Util from "./Util";
import * as Setting from "../Setting";
import SelfLoginButton from "./SelfLoginButton";
import i18next from "i18next";
import CustomGithubCorner from "../CustomGithubCorner";
import {CountDownInput} from "../common/CountDownInput";

const {TabPane} = Tabs;

class LoginPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      type: props.type,
      applicationName: props.applicationName !== undefined ? props.applicationName : (props.match === undefined ? null : props.match.params.applicationName),
      owner : props.owner !== undefined ? props.owner : (props.match === undefined ? null : props.match.params.owner),
      application: null,
      mode: props.mode !== undefined ? props.mode : (props.match === undefined ? null : props.match.params.mode), // "signup" or "signin"
      isCodeSignin: false,
      msg: null,
      username: null,
      validEmailOrPhone: false,
      validEmail: false,
      validPhone: false,
      loginMethod: "password"
    };

    if (this.state.type === "cas" && props.match?.params.casApplicationName !== undefined) {
      this.state.owner = props.match?.params.owner;
      this.state.applicationName = props.match?.params.casApplicationName;
    }
  }

  UNSAFE_componentWillMount() {
    if (this.state.type === "login" || this.state.type === "cas") {
      this.getApplication();
    } else if (this.state.type === "code") {
      this.getApplicationLogin();
    } else if (this.state.type === "saml") {
      this.getSamlApplication();
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

  getSamlApplication() {
    if (this.state.applicationName === null) {
      return;
    }
    ApplicationBackend.getApplication(this.state.owner, this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });
      }
      );
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
    if (this.state.loginMethod === "webAuthn") {
      let username = this.state.username;
      if (username === null || username === "") {
        username = values["username"];
      }

      this.signInWithWebAuthn(username);
      return;
    }

    const application = this.getApplicationObj();
    const ths = this;

    // here we are supposed to determine whether Casdoor is working as an OAuth server or CAS server
    if (this.state.type === "cas") {
      // CAS
      const casParams = Util.getCasParameters();
      values["type"] = this.state.type;
      AuthBackend.loginCas(values, casParams).then((res) => {
        if (res.status === "ok") {
          let msg = "Logged in successfully. ";
          if (casParams.service === "") {
            // If service was not specified, Casdoor must display a message notifying the client that it has successfully initiated a single sign-on session.
            msg += "Now you can visit apps protected by Casdoor.";
          }
          Util.showMessage("success", msg);

          if (casParams.service !== "") {
            let st = res.data;
            let newUrl = new URL(casParams.service);
            newUrl.searchParams.append("ticket", st);
            window.location.href = newUrl.toString();
          }
        } else {
          Util.showMessage("error", `Failed to log in: ${res.msg}`);
        }
      });
    } else {
      // OAuth
      const oAuthParams = Util.getOAuthGetParameters();
      if (oAuthParams !== null && oAuthParams.responseType != null && oAuthParams.responseType !== "") {
        values["type"] = oAuthParams.responseType;
      } else {
        values["type"] = this.state.type;
      }
      values["phonePrefix"] = this.getApplicationObj()?.organizationObj.phonePrefix;

      if (oAuthParams !== null) {
        values["samlRequest"] = oAuthParams.samlRequest;
      }

      if (values["samlRequest"] != null && values["samlRequest"] !== "") {
        values["type"] = "saml";
      }

      if (this.state.owner != null) {
        values["organization"] = this.state.owner;
      }

      AuthBackend.login(values, oAuthParams)
        .then((res) => {
          if (res.status === "ok") {
            const responseType = values["type"];
            if (responseType === "login") {
              Util.showMessage("success", "Logged in successfully");

              const link = Setting.getFromLink();
              Setting.goToLink(link);
            } else if (responseType === "code") {
              const code = res.data;
              const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
              const noRedirect = oAuthParams.noRedirect;

              if (Setting.hasPromptPage(application)) {
                AuthBackend.getAccount("")
                  .then((res) => {
                    let account = null;
                    if (res.status === "ok") {
                      account = res.data;
                      account.organization = res.data2;

                      this.onUpdateAccount(account);

                      if (Setting.isPromptAnswered(account, application)) {
                        Setting.goToLink(`${oAuthParams.redirectUri}${concatChar}code=${code}&state=${oAuthParams.state}`);
                      } else {
                        Setting.goToLinkSoft(ths, `/prompt/${application.name}?redirectUri=${oAuthParams.redirectUri}&code=${code}&state=${oAuthParams.state}`);
                      }
                    } else {
                      Setting.showMessage("error", `Failed to sign in: ${res.msg}`);
                    }
                  });
              } else {
                if (noRedirect === "true") {
                  window.close();
                  const newWindow = window.open(`${oAuthParams.redirectUri}${concatChar}code=${code}&state=${oAuthParams.state}`);
                  if (newWindow) {
                    setInterval(() => {
                      if (!newWindow.closed) {
                        newWindow.close();
                      }
                    }, 1000);
                  }
                } else {
                  Setting.goToLink(`${oAuthParams.redirectUri}${concatChar}code=${code}&state=${oAuthParams.state}`);
                }
              }

              // Util.showMessage("success", `Authorization code: ${res.data}`);
            } else if (responseType === "token" || responseType === "id_token") {
              const accessToken = res.data;
              Setting.goToLink(`${oAuthParams.redirectUri}#${responseType}=${accessToken}?state=${oAuthParams.state}&token_type=bearer`);
            } else if (responseType === "saml") {
              const SAMLResponse = res.data;
              const redirectUri = res.data2;
              Setting.goToLink(`${redirectUri}?SAMLResponse=${encodeURIComponent(SAMLResponse)}&RelayState=${oAuthParams.relayState}`);
            }
          } else {
            Util.showMessage("error", `Failed to log in: ${res.msg}`);
          }
        });
    }
  }

  getSamlUrl(provider) {
    const params = new URLSearchParams(this.props.location.search);
    let clientId = params.get("client_id");
    let application = params.get("state");
    let realRedirectUri = params.get("redirect_uri");
    let redirectUri = `${window.location.origin}/callback/saml`;
    let providerName = provider.name;
    let relayState = `${clientId}&${application}&${providerName}&${realRedirectUri}&${redirectUri}`;
    AuthBackend.getSamlLogin(`${provider.owner}/${providerName}`, btoa(relayState)).then((res) => {
      if (res.data2 === "POST") {
        document.write(res.data);
      } else {
        window.location.href = res.data;
      }
    });
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
      return Util.renderMessage(this.state.msg);
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
      );
    }

    if (application.enablePassword) {
      return (
        <Form
          name="normal_login"
          initialValues={{
            organization: application.organization,
            application: application.name,
            autoSignin: true,
          }}
          onFinish={(values) => {this.onFinish(values);}}
          style={{width: "300px"}}
          size="large"
        >
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="application"
            rules={[
              {
                required: true,
                message: "Please input your application!",
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
                message: "Please input your organization!",
              },
            ]}
          >
          </Form.Item>
          {this.renderMethodChoiceBox()}
          <Form.Item
            name="username"
            rules={[
              {
                required: true,
                message: i18next.t("login:Please input your username, Email or phone!")
              },
              {
                validator: (_, value) => {
                  if (this.state.isCodeSignin) {
                    if (this.state.email !== "" && !Setting.isValidEmail(this.state.username) && !Setting.isValidPhone(this.state.username)) {
                      this.setState({validEmailOrPhone: false});
                      return Promise.reject(i18next.t("login:The input is not valid Email or Phone!"));
                    }

                    if (Setting.isValidPhone(this.state.username)) {
                      this.setState({validPhone: true});
                    }
                    if (Setting.isValidEmail(this.state.username)) {
                      this.setState({validEmail: true});
                    }
                  }

                  this.setState({validEmailOrPhone: true});
                  return Promise.resolve();
                }
              }
            ]}
          >
            <Input
              prefix={<UserOutlined className="site-form-item-icon" />}
              placeholder={this.state.isCodeSignin ? i18next.t("login:Email or phone") : i18next.t("login:username, Email or phone")}
              disabled={!application.enablePassword}
              onChange={e => {
                this.setState({
                  username: e.target.value,
                });
              }}
            />
          </Form.Item>
          {
            this.renderPasswordOrCodeInput()
          }
          <Form.Item>
            <Form.Item name="autoSignin" valuePropName="checked" noStyle>
              <Checkbox style={{float: "left"}} disabled={!application.enablePassword}>
                {i18next.t("login:Auto sign in")}
              </Checkbox>
            </Form.Item>
            <a style={{float: "right"}} onClick={() => {
              Setting.goToForget(this, application);
            }}>
              {i18next.t("login:Forgot password?")}
            </a>
          </Form.Item>
          <Form.Item>
            {
              this.state.loginMethod === "password" ?
                (
                  <Button
                    type="primary"
                    htmlType="submit"
                    style={{width: "100%", marginBottom: "5px"}}
                    disabled={!application.enablePassword}
                  >
                    {i18next.t("login:Sign In")}
                  </Button>
                ) :
                (
                <Button
                  type="primary"
                  htmlType="submit"
                  style={{width: "100%", marginBottom: "5px"}}
                  disabled={!application.enablePassword}
                >
                  {i18next.t("login:Sign in with WebAuthn")}
                </Button>
                )
            }
            {
              this.renderFooter(application)
            }
          </Form.Item>
          <Form.Item>
            {
              application.providers.filter(providerItem => this.isProviderVisible(providerItem)).map(providerItem => {
                return ProviderButton.renderProviderLogo(providerItem.provider, application, 30, 5, "small");
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
          <br />
          {
            application.providers.filter(providerItem => this.isProviderVisible(providerItem)).map(providerItem => {
              return ProviderButton.renderProviderLogo(providerItem.provider, application, 40, 10, "big");
            })
          }
          <div>
            <br />
            {
              this.renderFooter(application)
            }
          </div>
        </div>
      );
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
      );
    } else {
      return (
        <React.Fragment>
          <span style={{float: "left"}}>
            {
              !application.enableCodeSignin ? null : (
                <a onClick={() => {
                  this.setState({
                    isCodeSignin: !this.state.isCodeSignin,
                  });
                }}>
                  {this.state.isCodeSignin ? i18next.t("login:Sign in with password") : i18next.t("login:Sign in with code")}
                </a>
              )
            }
          </span>
          <span style={{float: "right"}}>
            {
              !application.enableSignUp ? null : (
                <>
                  {i18next.t("login:No account?")}&nbsp;
                  <a onClick={() => {
                    sessionStorage.setItem("signinUrl", window.location.href);
                    Setting.goToSignup(this, application);
                  }}>
                    {i18next.t("login:sign up now")}
                  </a>
                </>
              )
            }
          </span>
        </React.Fragment>
      );
    }
  }

  renderSignedInBox() {
    if (this.props.account === undefined || this.props.account === null) {
      return null;
    }
    let application = this.getApplicationObj();
    if (this.props.account.owner !== application.organization) {
      return null;
    }

    const params = new URLSearchParams(this.props.location.search);
    let silentSignin = params.get("silentSignin");
    if (silentSignin !== null) {
      if (window !== window.parent) {
        const message = {tag: "Casdoor", type: "SilentSignin", data: "signing-in"};
        window.parent.postMessage(message, "*");
      }

      let values = {};
      values["application"] = this.state.application.name;
      this.onFinish(values);
    }

    return (
      <div>
        {/* {*/}
        {/*  JSON.stringify(silentSignin)*/}
        {/* }*/}
        <div style={{fontSize: 16, textAlign: "left"}}>
          {i18next.t("login:Continue with")}&nbsp;:
        </div>
        <br />
        <SelfLoginButton account={this.props.account} onClick={() => {
          let values = {};
          values["application"] = this.state.application.name;
          this.onFinish(values);
        }} />
        <br />
        <br />
        <div style={{fontSize: 16, textAlign: "left"}}>
          {i18next.t("login:Or sign in with another account")}&nbsp;:
        </div>
      </div>
    );
  }

  signInWithWebAuthn(username) {
    if (username === null || username === "") {
      Setting.showMessage("error", "username is required for webauthn login");
      return;
    }

    let application = this.getApplicationObj();
    return fetch(`${Setting.ServerUrl}/api/webauthn/signin/begin?owner=${application.organization}&name=${username}`, {
      method: "GET",
      credentials: "include"
    })
      .then(res => res.json())
      .then((credentialRequestOptions) => {
        if ("status" in credentialRequestOptions) {
          Setting.showMessage("error", credentialRequestOptions.msg);
          throw credentialRequestOptions.status.msg;
        }

        credentialRequestOptions.publicKey.challenge = UserWebauthnBackend.webAuthnBufferDecode(credentialRequestOptions.publicKey.challenge);
        credentialRequestOptions.publicKey.allowCredentials.forEach(function(listItem) {
          listItem.id = UserWebauthnBackend.webAuthnBufferDecode(listItem.id);
        });

        return navigator.credentials.get({
          publicKey: credentialRequestOptions.publicKey
        });
      })
      .then((assertion) => {
        let authData = assertion.response.authenticatorData;
        let clientDataJSON = assertion.response.clientDataJSON;
        let rawId = assertion.rawId;
        let sig = assertion.response.signature;
        let userHandle = assertion.response.userHandle;
        return fetch(`${Setting.ServerUrl}/api/webauthn/signin/finish`, {
          method: "POST",
          credentials: "include",
          body: JSON.stringify({
            id: assertion.id,
            rawId: UserWebauthnBackend.webAuthnBufferEncode(rawId),
            type: assertion.type,
            response: {
              authenticatorData: UserWebauthnBackend.webAuthnBufferEncode(authData),
              clientDataJSON: UserWebauthnBackend.webAuthnBufferEncode(clientDataJSON),
              signature: UserWebauthnBackend.webAuthnBufferEncode(sig),
              userHandle: UserWebauthnBackend.webAuthnBufferEncode(userHandle),
            },
          })
        })
          .then(res => res.json()).then((res) => {
            if (res.msg === "") {
              Setting.showMessage("success", "Successfully logged in with webauthn credentials");
              Setting.goToLink("/");
            } else {
              Setting.showMessage("error", res.msg);
            }
          })
          .catch(error => {
            Setting.showMessage("error", `Failed to connect to server: ${error}`);
          });
      });
  }

  renderPasswordOrCodeInput() {
    let application = this.getApplicationObj();
    if (this.state.loginMethod === "password") {
      return this.state.isCodeSignin ? (
        <Form.Item
          name="code"
          rules={[{required: true, message: i18next.t("login:Please input your code!")}]}
        >
          <CountDownInput
            disabled={this.state.username?.length === 0 || !this.state.validEmailOrPhone}
            onButtonClickArgs={[this.state.username, this.state.validEmail ? "email" : "phone", Setting.getApplicationName(application)]}
          />
        </Form.Item>
      ) : (
        <Form.Item
          name="password"
          rules={[{required: true, message: i18next.t("login:Please input your password!")}]}
        >
          <Input
            prefix={<LockOutlined className="site-form-item-icon" />}
            type="password"
            placeholder={i18next.t("login:Password")}
            disabled={!application.enablePassword}
          />
        </Form.Item>
      );
    }
  }

  renderMethodChoiceBox() {
    let application = this.getApplicationObj();
    if (application.enableWebAuthn) {
      return (
        <div>
          <Tabs defaultActiveKey="password" onChange={(key) => {this.setState({loginMethod: key});}} centered>
            <TabPane tab={i18next.t("login:Password")} key="password" />
            <TabPane tab={"WebAuthn"} key="webAuthn" />
          </Tabs>
        </div>
      );
    }
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return Util.renderMessageLarge(this, this.state.msg);
    }

    if (application.signinHtml !== "") {
      return (
        <div dangerouslySetInnerHTML={{__html: application.signinHtml}} />
      );
    }

    const visibleOAuthProviderItems = application.providers.filter(providerItem => this.isProviderVisible(providerItem));
    if (this.props.application === undefined && !application.enablePassword && visibleOAuthProviderItems.length === 1) {
      Setting.goToLink(Provider.getAuthUrl(application, visibleOAuthProviderItems[0].provider, "signup"));
      return (
        <div style={{textAlign: "center"}}>
          <Spin size="large" tip={i18next.t("login:Signing in...")} style={{paddingTop: "10%"}} />
        </div>
      );
    }
    const FormStyle = {
      padding: "30px",
      border: "2px solid #ffffff",
      borderRadius: "7px",
      backgroundColor:"#ffffff",
      boxShadow:"rgba(0, 0, 0, 0.19) 0px 10px 20px, rgba(0, 0, 0, 0.23) 0px 6px 6px"
    };
    return (
      <div style={{backgroundColor:"#f0f0f0",
      // background: "url(https://desk-fd.zol-img.com.cn/t_s960x600c5/g5/M00/0A/0C/ChMkJlwbaPmIfjr8AAHXNlQ9tyQAAt5PAMXql0AAddO522.jpg)",
      backgroundRepeat:"no-repeat",
      backgroundSize: "100% 100%"
      }}>
      <Row>
        <Col span={24} style={{display: "flex", justifyContent: "center"}}>
          <div style={{marginTop: "80px", marginBottom: "90px", textAlign: "center"}}>
            <div style={{...FormStyle}}>
            {
              Setting.renderHelmet(application)
            }
            <CustomGithubCorner />
            {
              Setting.renderLogo(application)
            }
            {/* {*/}
            {/*  this.state.clientId !== null ? "Redirect" : null*/}
            {/* }*/}
            {
              this.renderSignedInBox()
            }
            {
              this.renderForm(application)
            }
            </div>
          </div>
        </Col>
      </Row>
      </div>
    );
  }
}

export default LoginPage;
