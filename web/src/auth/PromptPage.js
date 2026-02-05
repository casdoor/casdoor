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
import {Button, Card, Col, Input, Result, Row, Form} from "antd";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as UserBackend from "../backend/UserBackend";
import * as Setting from "../Setting";
import i18next from "i18next";
import AffiliationSelect from "../common/select/AffiliationSelect";
import OAuthWidget from "../common/OAuthWidget";
import RegionSelect from "../common/select/RegionSelect";
import {withRouter} from "react-router-dom";
import * as AuthBackend from "./AuthBackend";
import {CountryCodeSelect} from "../common/select/CountryCodeSelect";

class PromptPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      type: props.type,
      applicationName: props.applicationName ?? (props.match === undefined ? null : props.match.params.applicationName),
      application: null,
      user: null,
      steps: null,
      current: 0,
      finished: false,
      validPhone: true,
    };
    this.form = React.createRef();
  }

  UNSAFE_componentWillMount() {
    this.getUser();
    if (this.getApplicationObj() === null) {
      this.getApplication();
    }
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    if (this.state.user !== null && this.getApplicationObj() !== null && this.state.steps === null) {
      this.initSteps(this.state.user, this.getApplicationObj());
    }
  }

  getUser() {
    const organizationName = this.props.account.owner;
    const userName = this.props.account.name;
    UserBackend.getUser(organizationName, userName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          user: res.data,
        });
      });
  }

  getApplication() {
    if (this.state.applicationName === null) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.onUpdateApplication(res.data);
        this.setState({
          application: res.data,
        });
      });
  }

  getApplicationObj() {
    return this.props.application ?? this.state.application;
  }

  onUpdateApplication(application) {
    this.props.onUpdateApplication(application);
  }

  parseUserField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateUserField(key, value) {
    value = this.parseUserField(key, value);

    const user = this.state.user;
    user[key] = value;
    this.setState({
      user: user,
    });

    this.submitUserEdit(false);
  }

  updateUserFieldWithoutSubmit(key, value) {
    value = this.parseUserField(key, value);

    const user = this.state.user;
    user[key] = value;
    this.setState({
      user: user,
    });
  }

  unlinked() {
    this.getUser();
  }

  renderSignupItem(signupItem) {
    const required = signupItem.required;
    const user = this.state.user;

    if (signupItem.name === "Display name") {
      const displayNameRules = [
        {
          required: required,
          message: i18next.t("signup:Please input your display name!"),
          whitespace: true,
        },
      ];
      if (signupItem.regex) {
        displayNameRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }

      return (
        <Form.Item
          key={signupItem.name}
          name="name"
          label={signupItem.label ? signupItem.label : i18next.t("general:Display name")}
          rules={displayNameRules}
          initialValue={user?.displayName || ""}
        >
          <Input placeholder={signupItem.placeholder} onChange={e => this.updateUserFieldWithoutSubmit("displayName", e.target.value)} />
        </Form.Item>
      );
    } else if (signupItem.name === "First name") {
      const firstNameRules = [
        {
          required: required,
          message: i18next.t("signup:Please input your first name!"),
          whitespace: true,
        },
      ];
      if (signupItem.regex) {
        firstNameRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }
      return (
        <Form.Item
          key={signupItem.name}
          name="firstName"
          label={signupItem.label ? signupItem.label : i18next.t("general:First name")}
          rules={firstNameRules}
          initialValue={user?.firstName || ""}
        >
          <Input placeholder={signupItem.placeholder} onChange={e => this.updateUserFieldWithoutSubmit("firstName", e.target.value)} />
        </Form.Item>
      );
    } else if (signupItem.name === "Last name") {
      const lastNameRules = [
        {
          required: required,
          message: i18next.t("signup:Please input your last name!"),
          whitespace: true,
        },
      ];
      if (signupItem.regex) {
        lastNameRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }
      return (
        <Form.Item
          key={signupItem.name}
          name="lastName"
          label={signupItem.label ? signupItem.label : i18next.t("general:Last name")}
          rules={lastNameRules}
          initialValue={user?.lastName || ""}
        >
          <Input placeholder={signupItem.placeholder} onChange={e => this.updateUserFieldWithoutSubmit("lastName", e.target.value)} />
        </Form.Item>
      );
    } else if (signupItem.name === "Email") {
      const emailRules = [
        {
          required: required,
          message: i18next.t("signup:Please input your Email!"),
        },
        {
          validator: (_, value) => {
            if (value !== "" && !Setting.isValidEmail(value)) {
              return Promise.reject(i18next.t("signup:The input is not valid Email!"));
            }

            if (signupItem.regex) {
              const reg = new RegExp(signupItem.regex);
              if (!reg.test(value)) {
                return Promise.reject(i18next.t("signup:The input Email doesn't match the signup item regex!"));
              }
            }

            return Promise.resolve();
          },
        },
      ];
      return (
        <Form.Item
          key={signupItem.name}
          name="email"
          label={signupItem.label ? signupItem.label : i18next.t("general:Email")}
          rules={emailRules}
          initialValue={user?.email || ""}
        >
          <Input placeholder={signupItem.placeholder} onChange={e => this.updateUserFieldWithoutSubmit("email", e.target.value)} />
        </Form.Item>
      );
    } else if (signupItem.name === "Phone") {
      return (
        <Form.Item key={signupItem.name} label={signupItem.label ? signupItem.label : i18next.t("general:Phone")} required={required}>
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
              initialValue={user?.countryCode || ""}
            >
              <CountryCodeSelect
                style={{width: "35%"}}
                countryCodes={this.getApplicationObj().organizationObj.countryCodes}
                onChange={value => this.updateUserFieldWithoutSubmit("countryCode", value)}
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
              initialValue={user?.phone || ""}
            >
              <Input
                placeholder={signupItem.placeholder}
                style={{width: "65%"}}
                onChange={e => this.updateUserFieldWithoutSubmit("phone", e.target.value)}
              />
            </Form.Item>
          </Input.Group>
        </Form.Item>
      );
    } else if (signupItem.name === "Affiliation") {
      const affiliationRules = [
        {
          required: required,
          message: i18next.t("signup:Please input your affiliation!"),
          whitespace: true,
        },
      ];
      if (signupItem.regex) {
        affiliationRules.push({
          pattern: new RegExp(signupItem.regex),
          message: i18next.t("signup:The input doesn't match the signup item regex!"),
        });
      }
      return (
        <Form.Item
          key={signupItem.name}
          name="affiliation"
          label={signupItem.label ? signupItem.label : i18next.t("user:Affiliation")}
          rules={affiliationRules}
          initialValue={user?.affiliation || ""}
        >
          <Input placeholder={signupItem.placeholder} onChange={e => this.updateUserFieldWithoutSubmit("affiliation", e.target.value)} />
        </Form.Item>
      );
    } else if (signupItem.name === "Country/Region") {
      return (
        <Form.Item
          key={signupItem.name}
          name="region"
          label={signupItem.label ? signupItem.label : i18next.t("user:Country/Region")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please select your country/region!"),
            },
          ]}
          initialValue={user?.region || ""}
        >
          <RegionSelect onChange={(value) => {
            this.updateUserFieldWithoutSubmit("region", value);
          }} />
        </Form.Item>
      );
    } else if (signupItem.name === "ID card") {
      return (
        <Form.Item
          key={signupItem.name}
          name="idCard"
          label={signupItem.label ? signupItem.label : i18next.t("user:ID card")}
          rules={[
            {
              required: required,
              message: i18next.t("signup:Please input your ID card number!"),
              whitespace: true,
            },
          ]}
          initialValue={user?.idCard || ""}
        >
          <Input placeholder={signupItem.placeholder} onChange={e => this.updateUserFieldWithoutSubmit("idCard", e.target.value)} />
        </Form.Item>
      );
    }
    return null;
  }

  renderContent(application) {
    const promptedSignupItems = application?.signupItems?.filter(signupItem => Setting.isSignupItemPrompted(signupItem)) || [];

    return (
      <div style={{width: "500px"}}>
        <Form ref={this.form}>
          {
            (application === null || this.state.user === null) ? null : (
              application?.providers.filter(providerItem => Setting.isProviderPrompted(providerItem)).map((providerItem, index) => <OAuthWidget key={providerItem.name} labelSpan={6} user={this.state.user} application={application} providerItem={providerItem} account={this.props.account} onUnlinked={() => {return this.unlinked();}} />)
            )
          }
          {
            (application === null || this.state.user === null) ? null : (
              promptedSignupItems.map((signupItem, index) => this.renderSignupItem(signupItem))
            )
          }
        </Form>
      </div>
    );
  }

  onUpdateAccount(account) {
    this.props.onUpdateAccount(account);
  }

  getRedirectUrl() {
    // "/prompt/app-example?redirectUri=http://localhost:2000/callback&code=8eb113b072296818f090&state=app-example"
    const params = new URLSearchParams(this.props.location.search);
    const redirectUri = params.get("redirectUri");
    const code = params.get("code");
    const state = params.get("state");
    const oauth = params.get("oauth");
    if (redirectUri === null || code === null || state === null) {
      const signInUrl = sessionStorage.getItem("signinUrl");
      return oauth === "true" ? signInUrl : "";
    }
    return `${redirectUri}?code=${code}&state=${state}`;
  }

  logout() {
    AuthBackend.logout()
      .then((res) => {
        if (res.status === "ok") {
          this.onUpdateAccount(null);
        } else {
          Setting.showMessage("error", res.msg);
        }
      });
  }

  finishAndJump() {
    this.setState({
      finished: true,
    }, () => {
      const redirectUrl = this.getRedirectUrl();
      if (redirectUrl !== "" && redirectUrl !== null) {
        Setting.goToLink(redirectUrl);
      } else {
        Setting.redirectToLoginPage(this.getApplicationObj(), this.props.history);
      }
    });
  }

  performUserUpdate(isFinal) {
    const user = Setting.deepCopy(this.state.user);
    UserBackend.updateUser(this.state.user.owner, this.state.user.name, user)
      .then((res) => {
        if (res.status === "ok") {
          if (isFinal) {
            Setting.showMessage("success", i18next.t("general:Successfully saved"));
            this.finishAndJump();
          }
        } else {
          if (isFinal) {
            Setting.showMessage("error", res.msg);
          }
        }
      })
      .catch(error => {
        if (isFinal) {
          Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        }
      });
  }

  submitUserEdit(isFinal) {
    if (isFinal && this.form.current) {
      // Validate all form fields before submission
      this.form.current.validateFields()
        .then(values => {
          this.performUserUpdate(isFinal);
        })
        .catch(errorInfo => {
          // Extract field-specific error messages for better user feedback
          const errors = errorInfo.errorFields || [];
          if (errors.length > 0) {
            const firstError = errors[0];
            const fieldName = firstError.name.join(".");
            const errorMsg = firstError.errors[0] || i18next.t("signup:Please fill in all required fields!");
            Setting.showMessage("error", `${fieldName}: ${errorMsg}`);
          } else {
            Setting.showMessage("error", i18next.t("signup:Please fill in all required fields!"));
          }
        });
    } else {
      this.performUserUpdate(isFinal);
    }
  }

  renderPromptProvider(application) {
    return (
      <div style={{display: "flex", alignItems: "center", flexDirection: "column"}}>
        {this.renderContent(application)}
        <Button style={{marginTop: "50px", width: "200px"}}
          disabled={!Setting.isPromptAnswered(this.state.user, application)}
          type="primary" size="large" onClick={() => {
            this.submitUserEdit(true);
          }}>
          {i18next.t("code:Submit and complete")}
        </Button>
      </div>);
  }

  initSteps(user, application) {
    const steps = [];
    if (Setting.hasPromptPage(application)) {
      steps.push({
        content: this.renderPromptProvider(application),
        name: "provider",
        title: i18next.t("application:Binding providers"),
      });
    }

    this.setState({
      steps: steps,
    });
  }

  renderSteps() {
    if (this.state.steps === null || this.state.steps?.length === 0) {
      return null;
    }

    return (
      <Card style={{marginTop: "20px", marginBottom: "20px"}}
        title={this.state.steps[this.state.current].title}
      >
        <div >{this.state.steps[this.state.current].content}</div>
      </Card>
    );
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return null;
    }

    if (this.state.steps?.length === 0) {
      return (
        <Result
          style={{display: "flex", flex: "1 1 0%", justifyContent: "center", flexDirection: "column"}}
          status="error"
          title={i18next.t("application:Sign Up Error")}
          subTitle={i18next.t("application:You are unexpected to see this prompt page")}
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

    return (
      <div style={{display: "flex", flex: "1", justifyContent: "center"}}>
        {this.renderSteps()}
      </div>
    );
  }
}

export default withRouter(PromptPage);
