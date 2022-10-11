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
import * as OrganizationBackend from "../backend/OrganizationBackend";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Provider from "./Provider";
import * as ProviderButton from "./ProviderButton";
import * as Util from "./Util";
import * as Setting from "../Setting";
import SelfLoginButton from "./SelfLoginButton";
import i18next from "i18next";
import CustomGithubCorner from "../CustomGithubCorner";
import {CountDownInput} from "../common/CountDownInput";
import SelectLanguageBox from "../SelectLanguageBox";
import {withTranslation} from "react-i18next";
import {CaptchaModal} from "../common/CaptchaModal";

const {TabPane} = Tabs;

class LoginPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      type: props.type,
      applicationName: props.applicationName !== undefined ? props.applicationName : (props.match === undefined ? null : props.match.params.applicationName),
      owner: props.owner !== undefined ? props.owner : (props.match === undefined ? null : props.match.params.owner),
      application: null,
      mode: props.mode !== undefined ? props.mode : (props.match === undefined ? null : props.match.params.mode), // "signup" or "signin"
      msg: null,
      username: null,
      validEmailOrPhone: false,
      validEmail: false,
      validPhone: false,
      loginMethod: "password",
    };
    this.captchaModal = React.createRef();

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

    if (this.state.owner === null || this.state.owner === undefined || this.state.owner === "") {
      ApplicationBackend.getApplication("admin", this.state.applicationName)
        .then((application) => {
          this.setState({
            application: application,
          });
        });
    } else {
      OrganizationBackend.getDefaultApplication("admin", this.state.owner)
        .then((res) => {
          if (res.status === "ok") {
            this.setState({
              application: res.data,
              applicationName: res.data.name,
            });
          } else {
            Util.showMessage("error", res.msg);
          }
        });
    }
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

  parseOffset(offset) {
    if (offset === 2 || offset === 4 || Setting.inIframe() || Setting.isMobile()) {
      return "0 auto";
    }
    if (offset === 1) {
      return "0 10%";
    }
    if (offset === 3) {
      return "0 60%";
    }
  }

  populateOauthValues(values) {
    const oAuthParams = Util.getOAuthGetParameters();
    if (oAuthParams !== null && oAuthParams.responseType !== null && oAuthParams.responseType !== "") {
      values["type"] = oAuthParams.responseType;
    } else {
      values["type"] = this.state.type;
    }
    values["phonePrefix"] = this.getApplicationObj()?.organizationObj.phonePrefix;

    if (oAuthParams !== null) {
      values["samlRequest"] = oAuthParams.samlRequest;
    }

    if (values["samlRequest"] !== null && values["samlRequest"] !== "" && values["samlRequest"] !== undefined) {
      values["type"] = "saml";
    }

    if (this.state.application.organization !== null && this.state.application.organization !== undefined) {
      values["organization"] = this.state.application.organization;
    }
  }
  postCodeLoginAction(res) {
    const application = this.getApplicationObj();
    const ths = this;
    const oAuthParams = Util.getOAuthGetParameters();
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
  }

  onFinish(values) {
    if (this.state.loginMethod === "webAuthn") {
      let username = this.state.username;
      if (username === null || username === "") {
        username = values["username"];
      }

      this.signInWithWebAuthn(username, values);
      return;
    }

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
            const st = res.data;
            const newUrl = new URL(casParams.service);
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
      this.populateOauthValues(values);

      if (this.getLoginFailureCount() > 0 && this.captchaModal.current) {
        this.captchaModal.current.showCaptcha(() => {
          this.oauthLogin(values, oAuthParams);
        });
      } else {
        this.oauthLogin(values, oAuthParams);
      }
    }
  }

  oauthLogin(values, oAuthParams) {
    AuthBackend.login(values, oAuthParams)
      .then((res) => {
        if (res.status === "ok") {
          const responseType = values["type"];

          this.resetLoginFailureCount();

          if (responseType === "login") {
            Util.showMessage("success", "Logged in successfully");

            const link = Setting.getFromLink();
            Setting.goToLink(link);
          } else if (responseType === "code") {
            this.postCodeLoginAction(res);
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
          this.incrementLoginFailureCount();
          Util.showMessage("error", `Failed to log in: ${res.msg}`);
        }
      });
  }

  incrementLoginFailureCount() {
    if (localStorage) {
      localStorage.setItem("loginFailureCount", (this.getLoginFailureCount() + 1).toString());
    } else {
      this.setState((state) => {
        return {loginFailureCount: state.loginFailureCount + 1};
      });
    }
  }

  resetLoginFailureCount() {
    if (localStorage) {
      localStorage.setItem("loginFailureCount", "0");
    } else {
      this.setState(() => {
        return {loginFailureCount: 0};
      });
    }
  }

  getLoginFailureCount() {
    let loginFailureCount;
    if (localStorage) {
      loginFailureCount = Number.parseInt(localStorage.getItem("loginFailureCount"));
    } else {
      loginFailureCount = this.state.loginFailureCount ? this.state.loginFailureCount : 0;
    }
    if (loginFailureCount === null || loginFailureCount === undefined || isNaN(loginFailureCount)) {
      loginFailureCount = 0;
    }
    return loginFailureCount;
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
            <Link key="login" onClick={() => {
              Setting.goToLogin(this, application);
            }}>
              <Button type="primary" key="signin">
                Sign In
              </Button>
            </Link>,
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
          <Row style={{minHeight: 130, alignItems: "center"}}>
            <Col span={24}>
              <Form.Item
                name="username"
                rules={[
                  {
                    required: true,
                    message: i18next.t("login:Please input your username, Email or phone!"),
                  },
                  {
                    validator: (_, value) => {
                      if (this.state.loginMethod === "verificationCode") {
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
                    },
                  },
                ]}
              >
                <Input
                  id = "input"
                  prefix={<UserOutlined className="site-form-item-icon" />}
                  placeholder={(this.state.loginMethod === "verificationCode") ? i18next.t("login:Email or phone") : i18next.t("login:username, Email or phone")}
                  disabled={!application.enablePassword}
                  onChange={e => {
                    this.setState({
                      username: e.target.value,
                    });
                  }}
                />
              </Form.Item>
            </Col>
            {
              this.renderPasswordOrCodeInput()
            }
          </Row>
          <Form.Item>
            <Form.Item name="autoSignin" valuePropName="checked" noStyle>
              <Checkbox style={{float: "left"}} disabled={!application.enablePassword}>
                {i18next.t("login:Auto sign in")}
              </Checkbox>
            </Form.Item>
            {
              Setting.renderForgetLink(application, i18next.t("login:Forgot password?"))
            }
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              style={{width: "100%", marginBottom: "5px"}}
              disabled={!application.enablePassword}
            >
              {
                this.state.loginMethod === "webAuthn" ? i18next.t("login:Sign in with WebAuthn") :
                  i18next.t("login:Sign In")
              }
            </Button>
            {
              this.renderCaptchaModal(application)
            }
            {
              this.renderFooter(application)
            }
          </Form.Item>
          <Form.Item>
            {
              application.providers.filter(providerItem => this.isProviderVisible(providerItem)).map(providerItem => {
                return ProviderButton.renderProviderLogo(providerItem.provider, application, 30, 5, "small", this.props.location);
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
              return ProviderButton.renderProviderLogo(providerItem.provider, application, 40, 10, "big", this.props.location);
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

  renderCaptchaModal(application) {
    if (application && application.enableCaptcha && application.providers) {
      for (const idx in application.providers) {
        const provider = application.providers[idx].provider;
        if (provider && provider.category === "Captcha" && provider.type === "Default") {
          return (
            <CaptchaModal
              provider={provider}
              providerName={provider.name}
              clientSecret={provider.clientSecret}
              captchaType={provider.type}
              subType={provider.subType}
              owner={provider.owner}
              clientId={provider.clientId}
              name={provider.name}
              providerUrl={provider.providerUrl}
              clientId2={provider.clientId2}
              clientSecret2={provider.clientSecret2}
              preview={false}
              ref={this.captchaModal}
            />
          );
        }
      }
    }
    return null;
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
          <span style={{float: "right"}}>
            {
              !application.enableSignUp ? null : (
                <>
                  {i18next.t("login:No account?")}&nbsp;
                  {
                    Setting.renderSignupLink(application, i18next.t("login:sign up now"))
                  }
                </>
              )
            }
          </span>
        </React.Fragment>
      );
    }
  }

  sendSilentSigninData(data) {
    if (Setting.inIframe()) {
      const message = {tag: "Casdoor", type: "SilentSignin", data: data};
      window.parent.postMessage(message, "*");
    }
  }

  renderSignedInBox() {
    if (this.props.account === undefined || this.props.account === null) {
      this.sendSilentSigninData("user-not-logged-in");
      return null;
    }

    const application = this.getApplicationObj();
    if (this.props.account.owner !== application.organization) {
      return null;
    }

    const params = new URLSearchParams(this.props.location.search);
    const silentSignin = params.get("silentSignin");
    if (silentSignin !== null) {
      this.sendSilentSigninData("signing-in");

      const values = {};
      values["application"] = this.state.application.name;
      this.onFinish(values);
    }

    if (application.enableAutoSignin) {
      const values = {};
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
          const values = {};
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

  signInWithWebAuthn(username, values) {
    const oAuthParams = Util.getOAuthGetParameters();
    this.populateOauthValues(values);
    const application = this.getApplicationObj();
    return fetch(`${Setting.ServerUrl}/api/webauthn/signin/begin?owner=${application.organization}&name=${username}`, {
      method: "GET",
      credentials: "include",
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
          publicKey: credentialRequestOptions.publicKey,
        });
      })
      .then((assertion) => {
        const authData = assertion.response.authenticatorData;
        const clientDataJSON = assertion.response.clientDataJSON;
        const rawId = assertion.rawId;
        const sig = assertion.response.signature;
        const userHandle = assertion.response.userHandle;
        return fetch(`${Setting.ServerUrl}/api/webauthn/signin/finish${AuthBackend.oAuthParamsToQuery(oAuthParams)}`, {
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
          }),
        })
          .then(res => res.json()).then((res) => {
            if (res.msg === "") {
              const responseType = values["type"];
              if (responseType === "code") {
                this.postCodeLoginAction(res);
              } else if (responseType === "token" || responseType === "id_token") {
                const accessToken = res.data;
                Setting.goToLink(`${oAuthParams.redirectUri}#${responseType}=${accessToken}?state=${oAuthParams.state}&token_type=bearer`);
              } else {
                Setting.showMessage("success", "Successfully logged in with webauthn credentials");
                Setting.goToLink("/");
              }
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
    const application = this.getApplicationObj();
    if (this.state.loginMethod === "password") {
      return (
        <Col span={24}>
          <Form.Item
            name="password"
            rules={[{required: true, message: i18next.t("login:Please input your password!")}]}
          >
            <Input.Password
              prefix={<LockOutlined className="site-form-item-icon" />}
              type="password"
              placeholder={i18next.t("login:Password")}
              disabled={!application.enablePassword}
            />
          </Form.Item>
        </Col>
      );
    } else if (this.state.loginMethod === "verificationCode") {
      return (
        <Col span={24}>
          <Form.Item
            name="code"
            rules={[{required: true, message: i18next.t("login:Please input your code!")}]}
          >
            <CountDownInput
              disabled={this.state.username?.length === 0 || !this.state.validEmailOrPhone}
              onButtonClickArgs={[this.state.username, this.state.validEmail ? "email" : "phone", Setting.getApplicationName(application)]}
              application={application}
            />
          </Form.Item>
        </Col>
      );
    } else {
      return null;
    }
  }

  renderMethodChoiceBox() {
    const application = this.getApplicationObj();
    if (application.enableCodeSignin || application.enableWebAuthn) {
      return (
        <div>
          <Tabs size={"small"} defaultActiveKey="password" onChange={(key) => {this.setState({loginMethod: key});}} centered>
            <TabPane tab={i18next.t("login:Password")} key="password" />
            {
              !application.enableCodeSignin ? null : (
                <TabPane tab={i18next.t("login:Verification Code")} key="verificationCode" />
              )
            }
            {
              !application.enableWebAuthn ? null : (
                <TabPane tab={i18next.t("login:WebAuthn")} key="webAuthn" />
              )
            }
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

    return (
      <div className="loginBackground" style={{backgroundImage: Setting.inIframe() || Setting.isMobile() ? null : `url(${application.formBackgroundUrl})`}}>
        <CustomGithubCorner />
        <div className="login-content" style={{margin: this.parseOffset(application.formOffset)}}>
          {Setting.inIframe() ? null : <div dangerouslySetInnerHTML={{__html: application.formCss}} />}
          <div className="login-panel">
            <SelectLanguageBox id="language-box-corner" style={{top: "50px"}} />
            <div className="side-image" style={{display: application.formOffset !== 4 ? "none" : null}}>
              <div dangerouslySetInnerHTML={{__html: application.formSideHtml}} />
            </div>
            <div className="login-form">
              <div >
                <div>
                  {
                    Setting.renderHelmet(application)
                  }
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
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default withTranslation()(LoginPage);
