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
import {Button, Form, Input, Popover, Radio, Result, Row, Select, message} from "antd";
import * as Setting from "../Setting";
import * as AuthBackend from "./AuthBackend";
import * as ProviderButton from "./ProviderButton";
import i18next from "i18next";
import * as Util from "./Util";
import {authConfig} from "./Auth";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as AgreementModal from "../common/modal/AgreementModal";
import {SendCodeInput} from "../common/SendCodeInput";
import RegionSelect from "../common/select/RegionSelect";
import CustomGithubCorner from "../common/CustomGithubCorner";
import LanguageSelect from "../common/select/LanguageSelect";
import {withRouter} from "react-router-dom";
import {CountryCodeSelect} from "../common/select/CountryCodeSelect";
import * as PasswordChecker from "../common/PasswordChecker";
import * as InvitationBackend from "../backend/InvitationBackend";

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
      span: 16,
    },
  },
};

const renderFormItem = (signupItem) => {
  const commonProps = {
    name: signupItem.name.toLowerCase(),
    label: signupItem.label || signupItem.name,
    rules: [
      {
        required: signupItem.required,
        message: i18next.t(`signup:Please input your ${signupItem.label || signupItem.name}!`),
      },
    ],
  };

  if (!signupItem.type || signupItem.type === "Input") {
    return (
      <Form.Item {...commonProps}>
        <Input placeholder={signupItem.placeholder} />
      </Form.Item>
    );
  } else if (signupItem.type === "Single Choice" || signupItem.type === "Multiple Choices") {
    return (
      <Form.Item {...commonProps}>
        <Select
          mode={signupItem.type === "Multiple Choices" ? "multiple" : "single"}
          placeholder={signupItem.placeholder}
          showSearch={false}
          options={signupItem.options.map(option => ({label: option, value: option}))}
        />
      </Form.Item>
    );
  }
};

export const tailFormItemLayout = {
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
      applicationName: (props.applicationName ?? props.match?.params?.applicationName) ?? null,
      email: "",
      phone: "",
      emailOrPhoneMode: "",
      countryCode: "",
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

  componentDidMount() {
    const oAuthParams = Util.getOAuthGetParameters();
    if (oAuthParams !== null) {
      const signinUrl = window.location.pathname.replace("/signup/oauth/authorize", "/login/oauth/authorize");
      sessionStorage.setItem("signinUrl", signinUrl + window.location.search);
    }

    if (this.getApplicationObj() === undefined) {
      if (this.state.applicationName !== null) {
        this.getApplication(this.state.applicationName);

        const sp = new URLSearchParams(window.location.search);
        if (sp.has("invitationCode")) {
          const invitationCode = sp.get("invitationCode");
          this.setState({invitationCode: invitationCode});
          if (invitationCode !== "") {
            this.getInvitationCodeInfo(invitationCode, "admin/" + this.state.applicationName);
          }
        }
      } else if (oAuthParams !== null) {
        this.getApplicationLogin(oAuthParams);
      } else {
        Setting.showMessage("error", `Unknown application name: ${this.state.applicationName}`);
        this.onUpdateApplication(null);
      }
    }
  }

  getApplication(applicationName) {
    if (applicationName === undefined) {
      return;
    }

    ApplicationBackend.getApplication("admin", applicationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.onUpdateApplication(res.data);
      });
  }

  getApplicationLogin(oAuthParams) {
    AuthBackend.getApplicationLogin(oAuthParams)
      .then((res) => {
        if (res.status === "ok") {
          const application = res.data;
          this.onUpdateApplication(application);
        } else {
          this.onUpdateApplication(null);
          this.setState({
            msg: res.msg,
          });
        }
      });
  }

  getInvitationCodeInfo(invitationCode, application) {
    InvitationBackend.getInvitationCodeInfo(invitationCode, application)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }
        this.setState({invitation: res.data});
      });
  }

  getResultPath(application, signupParams) {
    if (signupParams?.plan && signupParams?.pricing) {
      // the prompt page needs the user to be signed in, so for paid-user sign up, just go to buy-plan page
      return `/buy-plan/${application.organization}/${signupParams?.pricing}?user=${signupParams.username}&plan=${signupParams.plan}`;
    }
    if (authConfig.appName === application.name) {
      return "/result";
    } else {
      const oAuthParams = Util.getOAuthGetParameters();
      if (Setting.hasPromptPage(application)) {
        return `/prompt/${application.name}?oauth=${oAuthParams !== null}`;
      } else {
        return `/result/${application.name}`;
      }
    }
  }

  getApplicationObj() {
    return this.props.application;
  }

  onUpdateAccount(account) {
    this.props.onUpdateAccount(account);
  }

  onUpdateApplication(application) {
    this.props.onUpdateApplication(application);
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

  onFinish(values) {
    const application = this.getApplicationObj();

    if (Array.isArray(values.gender)) {
      values.gender = values.gender.join(", ");
    }

    if (Array.isArray(values.bio)) {
      values.bio = values.bio.join(", ");
    }

    if (Array.isArray(values.tag)) {
      values.tag = values.tag.join(", ");
    }

    if (Array.isArray(values.education)) {
      values.education = values.education.join(", ");
    }

    const params = new URLSearchParams(window.location.search);
    values.plan = params.get("plan");
    values.pricing = params.get("pricing");
    AuthBackend.signup(values)
      .then((res) => {
        if (res.status === "ok") {
          // the user's id will be returned by `signup()`, if user signup by phone, the `username` in `values` is undefined.
          values.username = res.data.split("/")[1];
          if (Setting.hasPromptPage(application) && (!values.plan || !values.pricing)) {
            AuthBackend.getAccount("")
              .then((res) => {
                let account = null;
                if (res.status === "ok") {
                  account = res.data;
                  account.organization = res.data2;

                  this.onUpdateAccount(account);
                  Setting.goToLinkSoft(this, this.getResultPath(application, values));
                } else {
                  Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
                }
              });
          } else {
            Setting.goToLinkSoft(this, this.getResultPath(application, values));
          }
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  onFinishFailed(values, errorFields, outOfDate) {
    this.form.current.scrollToField(errorFields[0].name);
  }

  isProviderVisible(providerItem) {
    return Setting.isProviderVisibleForSignUp(providerItem);
  }

  renderFormItem(application, signupItem) {
    const validItems = ["Gender", "Bio", "Tag", "Education"];
    if (!signupItem.visible) {
      return null;
    }

    const required = signupItem.required;

    if (signupItem.name === "Username") {
      return (
        <Form.Item
          name="username"
          className="signup-username"
          label={signupItem.label ? signupItem.label : i18next.t("signup:Username")}
          rules={[
            {
              required: required,
              message: i18next.t("forget:Please input your username!"),
              whitespace: true,
            },
          ]}
        >
          <Input className="signup-username-input" placeholder={signupItem.placeholder}
            disabled={this.state.invitation !== undefined && this.state.invitation.username !== ""} />
        </Form.Item>
      );
    } else if (signupItem.name === "Display name") {
      if (signupItem.rule === "First, last" && Setting.getLanguage() !== "zh") {
        return (
          <React.Fragment>
            <Form.Item
              name="firstName"
              className="signup-first-name"
              label={signupItem.label ? signupItem.label : i18next.t("general:First name")}
              rules={[
                {
                  required: required,
                  message: i18next.t("signup:Please input your first name!"),
                  whitespace: true,
                },
              ]}
            >
              <Input className="signup-first-name-input" placeholder={signupItem.placeholder} />
            </Form.Item>
            <Form.Item
              name="lastName"
              className="signup-last-name"
              label={signupItem.label ? signupItem.label : i18next.t("general:Last name")}
              rules={[
                {
                  required: required,
                  message: i18next.t("signup:Please input your last name!"),
                  whitespace: true,
                },
              ]}
            >
              <Input className="signup-last-name-input" placeholder={signupItem.placeholder} />
            </Form.Item>
          </React.Fragment>
        );
      }

      return (
        <Form.Item
          name="name"
          className="signup-name"
          label={(signupItem.label ? signupItem.label : (signupItem.rule === "Real name" || signupItem.rule === "First, last") ? i18next.t("general:Real name") : i18next.t("general:Display name"))}
          rules={[
            {
              required: required,
              message: (signupItem.rule === "Real name" || signupItem.rule === "First, last") ? i18next.t("signup:Please input your real name!") : i18next.t("signup:Please input your display name!"),
              whitespace: true,
            },
          ]}
        >
          <Input className="signup-name-input" placeholder={signupItem.placeholder} />
        </Form.Item>
      );
    } else if (signupItem.name === "Affiliation") {
      return (
        <Form.Item
          name="affiliation"
          className="signup-affiliation"
          label={signupItem.label ? signupItem.label : i18next.t("user:Affiliation")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please input your affiliation!"),
              whitespace: true,
            },
          ]}
        >
          <Input className="signup-affiliation-input" placeholder={signupItem.placeholder} />
        </Form.Item>
      );
    } else if (signupItem.name === "ID card") {
      return (
        <Form.Item
          name="idCard"
          className="signup-idcard"
          label={signupItem.label ? signupItem.label : i18next.t("user:ID card")}
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
          <Input className="signup-idcard-input" placeholder={signupItem.placeholder} />
        </Form.Item>
      );
    } else if (signupItem.name === "Country/Region") {
      return (
        <Form.Item
          name="country_region"
          className="signup-country-region"
          label={signupItem.label ? signupItem.label : i18next.t("user:Country/Region")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please select your country/region!"),
            },
          ]}
        >
          <RegionSelect className="signup-region-select" onChange={(value) => {
            this.setState({region: value});
          }} />
        </Form.Item>
      );
    } else if (signupItem.name === "Email" || signupItem.name === "Phone" || signupItem.name === "Email or Phone" || signupItem.name === "Phone or Email") {
      const renderEmailItem = () => {
        return (
          <React.Fragment>
            <Form.Item
              name="email"
              className="signup-email"
              label={signupItem.label ? signupItem.label : i18next.t("general:Email")}
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

                    if (signupItem.regex) {
                      const reg = new RegExp(signupItem.regex);
                      if (!reg.test(this.state.email)) {
                        this.setState({validEmail: false});
                        return Promise.reject(i18next.t("signup:The input Email doesn't match the signup item regex!"));
                      }
                    }

                    this.setState({validEmail: true});
                    return Promise.resolve();
                  },
                },
              ]}
            >
              <Input className="signup-email-input" placeholder={signupItem.placeholder} disabled={this.state.invitation !== undefined && this.state.invitation.email !== ""} onChange={e => this.setState({email: e.target.value})} />
            </Form.Item>
            {
              signupItem.rule !== "No verification" &&
              <Form.Item
                name="emailCode"
                className="signup-email-code"
                label={signupItem.label ? signupItem.label : i18next.t("code:Email code")}
                rules={[{
                  required: required,
                  message: i18next.t("code:Please input your verification code!"),
                }]}
              >
                <SendCodeInput
                  className="signup-email-code-input"
                  disabled={!this.state.validEmail}
                  method={"signup"}
                  onButtonClickArgs={[this.state.email, "email", Setting.getApplicationName(application)]}
                  application={application}
                />
              </Form.Item>
            }
          </React.Fragment>
        );
      };

      const renderPhoneItem = () => {
        return (
          <React.Fragment>
            <Form.Item className="signup-phone" label={signupItem.label ? signupItem.label : i18next.t("general:Phone")} required={required}>
              <Input.Group compact>
                <Form.Item
                  name="countryCode"
                  noStyle
                  rules={[
                    {
                      required: required,
                      message: i18next.t("signup:Please select your country code!"),
                    },
                  ]}
                >
                  <CountryCodeSelect
                    style={{width: "35%"}}
                    countryCodes={this.getApplicationObj().organizationObj.countryCodes}
                  />
                </Form.Item>
                <Form.Item
                  name="phone"
                  dependencies={["countryCode"]}
                  noStyle
                  rules={[
                    {
                      required: required,
                      message: i18next.t("signup:Please input your phone number!"),
                    },
                    ({getFieldValue}) => ({
                      validator: (_, value) => {
                        if (!required && !value) {
                          return Promise.resolve();
                        }

                        if (value && !Setting.isValidPhone(value, getFieldValue("countryCode"))) {
                          this.setState({validPhone: false});
                          return Promise.reject(i18next.t("signup:The input is not valid Phone!"));
                        }

                        this.setState({validPhone: true});
                        return Promise.resolve();
                      },
                    }),
                  ]}
                >
                  <Input
                    className="signup-phone-input"
                    placeholder={signupItem.placeholder}
                    style={{width: "65%"}}
                    disabled={this.state.invitation !== undefined && this.state.invitation.phone !== ""}
                    onChange={e => this.setState({phone: e.target.value})}
                  />
                </Form.Item>
              </Input.Group>
            </Form.Item>
            {
              signupItem.rule !== "No verification" &&
              <Form.Item
                name="phoneCode"
                className="phone-code"
                label={signupItem.label ? signupItem.label : i18next.t("code:Phone code")}
                rules={[
                  {
                    required: required,
                    message: i18next.t("code:Please input your phone verification code!"),
                  },
                ]}
              >
                <SendCodeInput
                  className="signup-phone-code-input"
                  disabled={!this.state.validPhone}
                  method={"signup"}
                  onButtonClickArgs={[this.state.phone, "phone", Setting.getApplicationName(application)]}
                  application={application}
                  countryCode={this.form.current?.getFieldValue("countryCode")}
                />
              </Form.Item>
            }
          </React.Fragment>
        );
      };

      if (signupItem.name === "Email") {
        return renderEmailItem();
      } else if (signupItem.name === "Phone") {
        return renderPhoneItem();
      } else if (signupItem.name === "Email or Phone" || signupItem.name === "Phone or Email") {
        let emailOrPhoneMode = this.state.emailOrPhoneMode;
        if (emailOrPhoneMode === "") {
          emailOrPhoneMode = signupItem.name === "Email or Phone" ? "Email" : "Phone";
        }

        return (
          <React.Fragment>
            <Row style={{marginTop: "30px", marginBottom: "20px"}} >
              <Radio.Group style={{width: "400px"}} buttonStyle="solid" onChange={e => {
                this.setState({
                  emailOrPhoneMode: e.target.value,
                });
              }} value={emailOrPhoneMode}>
                {
                  signupItem.name === "Email or Phone" ? (
                    <React.Fragment>
                      <Radio.Button value={"Email"}>{i18next.t("general:Email")}</Radio.Button>
                      <Radio.Button value={"Phone"}>{i18next.t("general:Phone")}</Radio.Button>
                    </React.Fragment>
                  ) : (
                    <React.Fragment>
                      <Radio.Button value={"Phone"}>{i18next.t("general:Phone")}</Radio.Button>
                      <Radio.Button value={"Email"}>{i18next.t("general:Email")}</Radio.Button>
                    </React.Fragment>
                  )
                }
              </Radio.Group>
            </Row>
            {
              emailOrPhoneMode === "Email" ? renderEmailItem() : renderPhoneItem()
            }
          </React.Fragment>
        );
      } else {
        return null;
      }
    } else if (signupItem.name === "Password") {
      return (
        <Popover placement={window.innerWidth >= 960 ? "right" : "top"} content={this.state.passwordPopover} open={this.state.passwordPopoverOpen}>
          <Form.Item
            name="password"
            className="signup-password"
            label={signupItem.label ? signupItem.label : i18next.t("general:Password")}
            rules={[
              {
                required: required,
                validateTrigger: "onChange",
                validator: (rule, value) => {
                  const errorMsg = PasswordChecker.checkPasswordComplexity(value, application.organizationObj.passwordOptions);
                  if (errorMsg === "") {
                    return Promise.resolve();
                  } else {
                    return Promise.reject(errorMsg);
                  }
                },
              },
            ]}
            hasFeedback
          >
            <Input.Password className="signup-password-input" placeholder={signupItem.placeholder} onChange={(e) => {
              this.setState({
                passwordPopover: PasswordChecker.renderPasswordPopover(application.organizationObj.passwordOptions, e.target.value),
              });
            }}
            onFocus={() => {
              this.setState({
                passwordPopoverOpen: application.organizationObj.passwordOptions?.length > 0,
                passwordPopover: PasswordChecker.renderPasswordPopover(application.organizationObj.passwordOptions, this.form.current?.getFieldValue("password") ?? ""),
              });
            }}
            onBlur={() => {
              this.setState({
                passwordPopoverOpen: false,
              });
            }} />
          </Form.Item>
        </Popover>
      );
    } else if (signupItem.name === "Confirm password") {
      return (
        <Form.Item
          name="confirm"
          className="signup-confirm"
          label={signupItem.label ? signupItem.label : i18next.t("signup:Confirm")}
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
          <Input.Password placeholder={signupItem.placeholder} />
        </Form.Item>
      );
    } else if (signupItem.name === "Invitation code") {
      return (
        <Form.Item
          name="invitationCode"
          className="signup-invitation-code"
          label={signupItem.label ? signupItem.label : i18next.t("application:Invitation code")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please input your invitation code!"),
            },
          ]}
        >
          <Input className="signup-invitation-code-input" placeholder={signupItem.placeholder} disabled={this.state.invitation !== undefined && this.state.invitation !== ""} />
        </Form.Item>
      );
    } else if (signupItem.name === "Agreement") {
      return AgreementModal.renderAgreementFormItem(application, required, tailFormItemLayout, this);
    } else if (signupItem.name.startsWith("Text ")) {
      return (
        <div dangerouslySetInnerHTML={{__html: signupItem.label}} />
      );
    } else if (signupItem.name === "Signup button") {
      return (
        <Form.Item {...tailFormItemLayout}>
          <Button type="primary" htmlType="submit" className="signup-button">
            {i18next.t("account:Sign Up")}
          </Button>
          &nbsp;&nbsp;{i18next.t("signup:Have account?")}&nbsp;
          <a className="signup-link" onClick={() => {
            const linkInStorage = sessionStorage.getItem("signinUrl");
            if (linkInStorage !== null && linkInStorage !== "") {
              Setting.goToLinkSoft(this, linkInStorage);
            } else {
              Setting.redirectToLoginPage(application, this.props.history);
            }
          }}>
            {i18next.t("signup:sign in now")}
          </a>
        </Form.Item>
      );
    } else if (signupItem.name === "Providers") {
      const showForm = Setting.isPasswordEnabled(application) || Setting.isCodeSigninEnabled(application) || Setting.isWebAuthnEnabled(application) || Setting.isLdapEnabled(application);
      if (signupItem.rule === "None" || signupItem.rule === "") {
        signupItem.rule = showForm ? "small" : "big";
      }
      return (

        application.providers.filter(providerItem => this.isProviderVisible(providerItem)).map((providerItem, id) => {
          return (
            <span key={id} onClick={(e) => {
              const agreementChecked = this.form.current.getFieldValue("agreement");

              if (agreementChecked !== undefined && typeof agreementChecked === "boolean" && !agreementChecked) {
                e.preventDefault();
                message.error(i18next.t("signup:Please accept the agreement!"));
              }
            }}>
              {
                ProviderButton.renderProviderLogo(providerItem.provider, application, null, null, signupItem.rule, this.props.location)
              }
            </span>
          );
        })
      );
    } else if (validItems.includes(signupItem.name)) {
      return renderFormItem(signupItem);
    }
  }

  renderForm(application) {
    if (!application.enableSignUp) {
      return (
        <Result
          status="error"
          title={i18next.t("application:Sign Up Error")}
          subTitle={i18next.t("application:The application does not allow to sign up new account")}
          extra={[
            <Button type="primary" key="signin" onClick={() => Setting.redirectToLoginPage(application, this.props.history)}>
              {
                i18next.t("login:Sign In")
              }
            </Button>,
          ]}
        >
        </Result>
      );
    }
    if (this.state.invitation !== undefined) {
      if (this.state.invitation.username !== "") {
        this.form.current?.setFieldValue("username", this.state.invitation.username);
      }
      if (this.state.invitation.email !== "") {
        this.form.current?.setFieldValue("email", this.state.invitation.email);
      }
      if (this.state.invitation.phone !== "") {
        this.form.current?.setFieldValue("phone", this.state.invitation.phone);
      }
      if (this.state.invitationCode !== "") {
        this.form.current?.setFieldValue("invitationCode", this.state.invitationCode);
      }
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
          countryCode: application.organizationObj.countryCodes?.[0],
        }}
        size="large"
        layout={Setting.isMobile() ? "vertical" : "horizontal"}
        style={{width: Setting.isMobile() ? "300px" : "400px"}}
      >
        <Form.Item
          name="application"
          hidden={true}
          rules={[
            {
              required: true,
              message: "Please input your application!",
            },
          ]}
        >
        </Form.Item>
        <Form.Item
          name="organization"
          hidden={true}
          rules={[
            {
              required: true,
              message: "Please input your organization!",
            },
          ]}
        >
        </Form.Item>
        {
          application.signupItems?.map((signupItem, idx) => {
            return (
              <div key={idx}>
                <div dangerouslySetInnerHTML={{__html: ("<style>" + signupItem.customCss + "</style>")}} />
                {this.renderFormItem(application, signupItem)}
              </div>
            );
          })
        }
      </Form>
    );
  }

  render() {
    const application = this.getApplicationObj();
    if (application === undefined || application === null) {
      return null;
    }

    let existSignupButton = false;
    application.signupItems?.map(item => {
      item.name === "Signup button" ? existSignupButton = true : null;
    });
    if (!existSignupButton) {
      application.signupItems?.push({
        customCss: "",
        label: "",
        name: "Signup button",
        placeholder: "",
        visible: true,
      });
    }

    if (application.signupHtml !== "") {
      return (
        <div dangerouslySetInnerHTML={{__html: application.signupHtml}} />
      );
    }

    return (
      <React.Fragment>
        <CustomGithubCorner />
        <div className="login-content" style={{margin: this.props.preview ?? this.parseOffset(application.formOffset)}}>
          {Setting.inIframe() || Setting.isMobile() ? null : <div dangerouslySetInnerHTML={{__html: application.formCss}} />}
          {Setting.inIframe() || !Setting.isMobile() ? null : <div dangerouslySetInnerHTML={{__html: application.formCssMobile}} />}
          <div className={Setting.isDarkTheme(this.props.themeAlgorithm) ? "login-panel-dark" : "login-panel"}>
            <div className="side-image" style={{display: application.formOffset !== 4 ? "none" : null}}>
              <div dangerouslySetInnerHTML={{__html: application.formSideHtml}} />
            </div>
            <div className="login-form">
              {
                Setting.renderHelmet(application)
              }
              {
                Setting.renderLogo(application)
              }
              <LanguageSelect languages={application.organizationObj.languages} style={{top: "55px", right: "5px", position: "absolute"}} />
              {
                this.renderForm(application)
              }
            </div>
          </div>
        </div>
      </React.Fragment>
    );
  }
}

export default withRouter(SignupPage);
