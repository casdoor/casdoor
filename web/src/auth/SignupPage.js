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
import {Button, Checkbox, Col, Form, Input, Modal, Result, Row} from "antd";
import * as Setting from "../Setting";
import * as AuthBackend from "./AuthBackend";
import * as ProviderButton from "./ProviderButton";
import i18next from "i18next";
import * as Util from "./Util";
import {authConfig} from "./Auth";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import {CountDownInput} from "../common/CountDownInput";
import SelectRegionBox from "../SelectRegionBox";
import CustomGithubCorner from "../CustomGithubCorner";

const formItemLayout = {
  labelCol: {
    xs: {
      span: 24,
    },
    sm: {
      span: 8,
    },
  },
  wrapperCol: {
    xs: {
      span: 24,
    },
    sm: {
      span: 18,
    },
  },
};

const tailFormItemLayout = {
  wrapperCol: {
    xs: {
      span: 24,
      offset: 0,
    },
    sm: {
      span: 16,
      offset: 8,
    },
  },
};

class SignupPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applicationName: props.match?.params.applicationName !== undefined ? props.match.params.applicationName : authConfig.appName,
      application: null,
      email: "",
      phone: "",
      emailCode: "",
      phoneCode: "",
      validEmail: false,
      validPhone: false,
      region: "",
      isTermsOfUseVisible: false,
      termsOfUseContent: "",
    };

    this.form = React.createRef();
  }

  UNSAFE_componentWillMount() {
    let applicationName = this.state.applicationName;
    const oAuthParams = Util.getOAuthGetParameters();
    if (oAuthParams !== null) {
      applicationName = oAuthParams.state;
      this.setState({applicationName: oAuthParams.state});
      const signinUrl = window.location.href.replace("/signup/oauth/authorize", "/login/oauth/authorize");
      sessionStorage.setItem("signinUrl", signinUrl);
    }

    if (applicationName !== undefined) {
      this.getApplication(applicationName);
    } else {
      Util.showMessage("error", `Unknown application name: ${applicationName}`);
    }
  }

  getApplication(applicationName) {
    if (applicationName === undefined) {
      return;
    }

    ApplicationBackend.getApplication("admin", applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });

        if (application !== null && application !== undefined) {
          this.getTermsofuseContent(application.termsOfUse);
        }
      });
  }

  getResultPath(application) {
    if (authConfig.appName === application.name) {
      return "/result";
    } else {
      if (Setting.hasPromptPage(application)) {
        return `/prompt/${application.name}`;
      } else {
        return `/result/${application.name}`;
      }
    }
  }

  getApplicationObj() {
    if (this.props.application !== undefined) {
      return this.props.application;
    } else {
      return this.state.application;
    }
  }

  getTermsofuseContent(url) {
    fetch(url, {
      method: "GET",
    }).then(r => {
      r.text().then(res => {
        this.setState({termsOfUseContent: res});
      });
    });
  }

  onUpdateAccount(account) {
    this.props.onUpdateAccount(account);
  }

  onFinish(values) {
    const application = this.getApplicationObj();
    values.phonePrefix = application.organizationObj.phonePrefix;
    AuthBackend.signup(values)
      .then((res) => {
        if (res.status === "ok") {
          if (Setting.hasPromptPage(application)) {
            AuthBackend.getAccount("")
              .then((res) => {
                let account = null;
                if (res.status === "ok") {
                  account = res.data;
                  account.organization = res.data2;

                  this.onUpdateAccount(account);
                  Setting.goToLinkSoft(this, this.getResultPath(application));
                } else {
                  Setting.showMessage("error", `Failed to sign in: ${res.msg}`);
                }
              });
          } else {
            Setting.goToLinkSoft(this, this.getResultPath(application));
          }
        } else {
          Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
        }
      });
  }

  onFinishFailed(values, errorFields, outOfDate) {
    this.form.current.scrollToField(errorFields[0].name);
  }

  isProviderVisible(providerItem) {
    if (this.state.mode === "signup") {
      return Setting.isProviderVisibleForSignUp(providerItem);
    } else {
      return Setting.isProviderVisibleForSignIn(providerItem);
    }
  }

  renderFormItem(application, signupItem) {
    if (!signupItem.visible) {
      return null;
    }

    const required = signupItem.required;

    if (signupItem.name === "Username") {
      return (
        <Form.Item
          name="username"
          key="username"
          label={i18next.t("signup:Username")}
          rules={[
            {
              required: required,
              message: i18next.t("forget:Please input your username!"),
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
      );
    } else if (signupItem.name === "Display name") {
      if (signupItem.rule === "First, last" && Setting.getLanguage() !== "zh") {
        return (
          <React.Fragment>
            <Form.Item
              name="firstName"
              key="firstName"
              label={i18next.t("general:First name")}
              rules={[
                {
                  required: required,
                  message: i18next.t("signup:Please input your first name!"),
                  whitespace: true,
                },
              ]}
            >
              <Input />
            </Form.Item>
            <Form.Item
              name="lastName"
              key="lastName"
              label={i18next.t("general:Last name")}
              rules={[
                {
                  required: required,
                  message: i18next.t("signup:Please input your last name!"),
                  whitespace: true,
                },
              ]}
            >
              <Input />
            </Form.Item>
          </React.Fragment>
        );
      }

      return (
        <Form.Item
          name="name"
          key="name"
          label={(signupItem.rule === "Real name" || signupItem.rule === "First, last") ? i18next.t("general:Real name") : i18next.t("general:Display name")}
          rules={[
            {
              required: required,
              message: (signupItem.rule === "Real name" || signupItem.rule === "First, last") ? i18next.t("signup:Please input your real name!") : i18next.t("signup:Please input your display name!"),
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
      );
    } else if (signupItem.name === "Affiliation") {
      return (
        <Form.Item
          name="affiliation"
          key="affiliation"
          label={i18next.t("user:Affiliation")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please input your affiliation!"),
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
      );
    } else if (signupItem.name === "ID card") {
      return (
        <Form.Item
          name="idCard"
          key="idCard"
          label={i18next.t("user:ID card")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please input your ID card number!"),
              whitespace: true,
            },
            {
              required: required,
              pattern: new RegExp(/^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9X]$/, "g"),
              message: i18next.t("signup:Please input the correct ID card number!"),
            },
          ]}
        >
          <Input />
        </Form.Item>
      );
    } else if (signupItem.name === "Country/Region") {
      return (
        <Form.Item
          name="country_region"
          key="region"
          label={i18next.t("user:Country/Region")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please select your country/region!"),
            },
          ]}
        >
          <SelectRegionBox onChange={(value) => {this.setState({region: value});}} />
        </Form.Item>
      );
    } else if (signupItem.name === "Email") {
      return (
        <React.Fragment>
          <Form.Item
            name="email"
            key="email"
            label={i18next.t("general:Email")}
            rules={[
              {
                required: required,
                message: i18next.t("signup:Please input your Email!"),
              },
              {
                validator: (_, value) => {
                  if (this.state.email !== "" && !Setting.isValidEmail(this.state.email)) {
                    this.setState({validEmail: false});
                    return Promise.reject(i18next.t("signup:The input is not valid Email!"));
                  }

                  this.setState({validEmail: true});
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Input onChange={e => this.setState({email: e.target.value})} />
          </Form.Item>
          {
            signupItem.rule !== "No verification" &&
            <Form.Item
              name="emailCode"
              key="emailCode"
              label={i18next.t("code:Email code")}
              rules={[{
                required: required,
                message: i18next.t("code:Please input your verification code!"),
              }]}
            >
              <CountDownInput
                disabled={!this.state.validEmail}
                onButtonClickArgs={[this.state.email, "email", Setting.getApplicationName(application)]}
                application={application}
              />
            </Form.Item>
          }
        </React.Fragment>
      );
    } else if (signupItem.name === "Phone") {
      return (
        <React.Fragment>
          <Form.Item
            name="phone"
            key="phone"
            label={i18next.t("general:Phone")}
            rules={[
              {
                required: required,
                message: i18next.t("signup:Please input your phone number!"),
              },
              {
                validator: (_, value) => {
                  if (this.state.phone !== "" && !Setting.isValidPhone(this.state.phone)) {
                    this.setState({validPhone: false});
                    return Promise.reject(i18next.t("signup:The input is not valid Phone!"));
                  }

                  this.setState({validPhone: true});
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Input
              style={{
                width: "100%",
              }}
              addonBefore={`+${this.state.application?.organizationObj.phonePrefix}`}
              onChange={e => this.setState({phone: e.target.value})}
            />
          </Form.Item>
          <Form.Item
            name="phoneCode"
            key="phoneCode"
            label={i18next.t("code:Phone code")}
            rules={[
              {
                required: required,
                message: i18next.t("code:Please input your phone verification code!"),
              },
            ]}
          >
            <CountDownInput
              disabled={!this.state.validPhone}
              onButtonClickArgs={[this.state.phone, "phone", Setting.getApplicationName(application)]}
              application={application}
            />
          </Form.Item>
        </React.Fragment>
      );
    } else if (signupItem.name === "Password") {
      return (
        <Form.Item
          name="password"
          key="password"
          label={i18next.t("general:Password")}
          rules={[
            {
              required: required,
              min: 6,
              message: i18next.t("login:Please input your password, at least 6 characters!"),
            },
          ]}
          hasFeedback
        >
          <Input.Password />
        </Form.Item>
      );
    } else if (signupItem.name === "Confirm password") {
      return (
        <Form.Item
          name="confirm"
          key="confirm"
          label={i18next.t("signup:Confirm")}
          dependencies={["password"]}
          hasFeedback
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please confirm your password!"),
            },
            ({getFieldValue}) => ({
              validator(rule, value) {
                if (!value || getFieldValue("password") === value) {
                  return Promise.resolve();
                }

                return Promise.reject(i18next.t("signup:Your confirmed password is inconsistent with the password!"));
              },
            }),
          ]}
        >
          <Input.Password />
        </Form.Item>
      );
    } else if (signupItem.name === "Agreement") {
      return (
        <Form.Item
          name="agreement"
          key="agreement"
          valuePropName="checked"
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please accept the agreement!"),
            },
          ]}
          {...tailFormItemLayout}
        >
          <Checkbox>
            {i18next.t("signup:Accept")}&nbsp;
            <Link onClick={() => {
              this.setState({
                isTermsOfUseVisible: true,
              });
            }}>
              {i18next.t("signup:Terms of Use")}
            </Link>
          </Checkbox>
        </Form.Item>
      );
    }
  }

  renderModal() {
    return (
      <Modal
        title={i18next.t("signup:Terms of Use")}
        visible={this.state.isTermsOfUseVisible}
        width={"55vw"}
        closable={false}
        okText={i18next.t("signup:Accept")}
        cancelText={i18next.t("signup:Decline")}
        onOk={() => {
          this.form.current.setFieldsValue({agreement: true});
          this.setState({
            isTermsOfUseVisible: false,
          });
        }}
        onCancel={() => {
          this.form.current.setFieldsValue({agreement: false});
          this.setState({
            isTermsOfUseVisible: false,
          });
          this.props.history.goBack();
        }}
      >
        <iframe title={"terms"} style={{border: 0, width: "100%", height: "60vh"}} srcDoc={this.state.termsOfUseContent} />
      </Modal>
    );
  }

  renderForm(application) {
    if (!application.enableSignUp) {
      return (
        <Result
          status="error"
          title="Sign Up Error"
          subTitle={"The application does not allow to sign up new account"}
          extra={[
            <Button type="primary" key="signin" onClick={() => {
              Setting.goToLogin(this, application);
            }}>
              Sign In
            </Button>,
          ]}
        >
        </Result>
      );
    }
    return (
      <Form
        {...formItemLayout}
        ref={this.form}
        name="signup"
        onFinish={(values) => this.onFinish(values)}
        onFinishFailed={(errorInfo) => this.onFinishFailed(errorInfo.values, errorInfo.errorFields, errorInfo.outOfDate)}
        initialValues={{
          application: application.name,
          organization: application.organization,
        }}
        style={{width: !Setting.isMobile() ? "400px" : "250px"}}
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
        {
          application.signupItems?.map(signupItem => this.renderFormItem(application, signupItem))
        }
        <Form.Item {...tailFormItemLayout}>
          <Button type="primary" htmlType="submit">
            {i18next.t("account:Sign Up")}
          </Button>
          &nbsp;&nbsp;{i18next.t("signup:Have account?")}&nbsp;
          <a onClick={() => {
            const linkInStorage = sessionStorage.getItem("signinUrl");
            if (linkInStorage !== null && linkInStorage !== "") {
              Setting.goToLink(linkInStorage);
            } else {
              Setting.goToLogin(this, application);
            }
          }}>
            {i18next.t("signup:sign in now")}
          </a>
        </Form.Item>
        {
          application.providers.filter(providerItem => this.isProviderVisible(providerItem)).map(providerItem => {
            return ProviderButton.renderProviderLogo(providerItem.provider, application, 30, 5, "small");
          })
        }
      </Form>
    );
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return null;
    }

    if (application.signupHtml !== "") {
      return (
        <div dangerouslySetInnerHTML={{__html: application.signupHtml}} />
      );
    }

    return (
      <div className="loginBackground" style={{
        backgroundImage: "url(" + application.backgroundUrl + ")"}}
      >
        <CustomGithubCorner />
        &nbsp;
        <Row>
          <Col span={8} offset={application.offset === 0 ? 8 : application.offset} style={{display: "flex", justifyContent: "center"}} >
            <div className="loginFrom" style={{marginTop: "10px", marginBottom: "30px", textAlign: "center"}}>
              {
                Setting.renderHelmet(application)
              }
              {
                Setting.renderLogo(application)
              }
              {
                this.renderForm(application)
              }
            </div>
          </Col>
        </Row>
        {
          this.renderModal()
        }
      </div>
    );
  }
}

export default SignupPage;
