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
import { Link } from "react-router-dom";
import { Form, Input, Checkbox, Button, Row, Col, Result, Avatar } from "antd";
import * as Setting from "../Setting";
import * as AuthBackend from "./AuthBackend";
import i18next from "i18next";
import * as Util from "./Util";
import { authConfig } from "./Auth";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as UserBackend from "../backend/UserBackend";
import { CountDownInput } from "../component/CountDownInput";
import { getUser } from "../backend/UserBackend";

const formItemLayout = {
  labelCol: {
    xs: {
      span: 24,
    },
    sm: {
      span: 6,
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

    this.form = React.createRef();

    const params = new URLSearchParams(this.props.location?.search);
    const paramUser = params.get("user");
    const paramOwner = params.get("owner");

    let stateUrlParams = {
      urlWithProvider: false,
      user: "",
      owner: "",
    };

    if (paramUser && paramOwner) {
      stateUrlParams.urlWithProvider = true;
      stateUrlParams.user = paramUser;
      stateUrlParams.owner = paramOwner;
    }

    this.state = {
      classes: props,
      applicationName:
        props.match?.params.applicationName !== undefined
          ? props.match.params.applicationName
          : authConfig.appName,
      application: null,
      email: "",
      phone: "",
      emailCode: "",
      phoneCode: "",
      provider: {
        type: "",
        user: "",
        username: "",
        email: "",
        avatar: "",
      },
      urlParams: stateUrlParams,
    };

    this.form = React.createRef();
  }

  UNSAFE_componentWillMount() {
    if (this.state.applicationName !== undefined) {
      this.getApplication();
    } else {
      Util.showMessage(
        "error",
        `Unknown application name: ${this.state.applicationName}`
      );
    }
  }

  getApplication() {
    if (this.state.applicationName === undefined) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });
      });
  }

  getResultPath(application) {
    if (authConfig.appName === application.name) {
      return "/result";
    } else {
      return `/result/${application.name}`;
    }
  }

  getApplicationObj() {
    if (this.props.application !== undefined) {
      return this.props.application;
    } else {
      return this.state.application;
    }
  }

  onFinish(values) {
    const application = this.getApplicationObj();
    values.phonePrefix = application.organizationObj.phonePrefix;
    values.organization = application.organization;
    values.application = application.name;
    values.withProvider = this.state.urlParams.urlWithProvider;
    values.avatar = this.state.provider.avatar;
    AuthBackend.signup(values)
      .then((res) => {
        if (res.status === "ok") {
          Setting.goToLinkSoft(this, this.getResultPath(application));
        } else {
          Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
        }
      });
  }

  onFinishFailed(values, errorFields, outOfDate) {
    this.form.current.scrollToField(errorFields[0].name);
  }

  renderForm(application) {
    if (!application.enableSignUp) {
      return (
        <Result
          status="error"
          title="Sign Up Error"
          subTitle={"The application does not allow to sign up new account"}
          extra={[
            <Link
              onClick={() => {
                Setting.goToLogin(this, application);
              }}
            >
              <Button type="primary" key="signin">
                Sign In
              </Button>
            </Link>,
          ]}
        ></Result>
      );
    }

    let providerItem;
    if (this.state.urlParams.urlWithProvider) {
      providerItem = (
        <Form.Item label={this.state.provider.type}>
          <div>
            <Avatar
              size={40}
              src={this.state.provider.avatar}
              alt={this.state.provider.user}
            />
            <span style={{ marginLeft: 10 }}>{this.state.provider.user}</span>
          </div>
        </Form.Item>
      );
    }

    return (
      <Form
        {...formItemLayout}
        ref={this.form}
        name="signup"
        onFinish={(values) => this.onFinish(values)}
        onFinishFailed={(errorInfo) => this.onFinishFailed(
            errorInfo.values,
            errorInfo.errorFields,
            errorInfo.outOfDate
          )}
        initialValues={{
          application: application.name,
          organization: application.organization,
        }}
        style={{ width: !Setting.isMobile() ? "400px" : "250px" }}
        size="large"
      >
        {providerItem}
        <Form.Item
          name="username"
          label={i18next.t("signup:Username")}
          rules={[
            {
              required: true,
              message: i18next.t("login:Please input your username!"),
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="name"
          label={i18next.t("general:Display name")}
          rules={[
            {
              required: true,
              message: i18next.t("signup:Please input your display name!"),
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="affiliation"
          label={i18next.t("user:Affiliation")}
          rules={[
            {
              required: true,
              message: i18next.t("signup:Please input your affiliation!"),
              whitespace: true,
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="email"
          label={i18next.t("general:Email")}
          rules={[
            {
              type: "email",
              message: i18next.t("signup:The input is not valid Email!"),
            },
            {
              required: true,
              message: i18next.t("signup:Please input your Email!"),
            },
          ]}
        >
          <Input onChange={(e) => this.setState({ email: e.target.value })} />
        </Form.Item>
        <Form.Item
          name="emailCode"
          label={i18next.t("signup:email code")}
          rules={[
            {
              required: true,
              message: i18next.t("signup:Please input your verification code!"),
            },
          ]}
        >
          <CountDownInput
            defaultButtonText={i18next.t("signup:send code")}
            onButtonClick={UserBackend.sendCode}
            onButtonClickArgs={[
              this.state.email,
              "email",
              application?.organizationObj.owner +
                "/" +
                application?.organizationObj.name,
            ]}
            coolDownTime={60}
          />
        </Form.Item>
        <Form.Item
          name="password"
          label={i18next.t("general:Password")}
          rules={[
            {
              required: true,
              message: i18next.t("login:Please input your password!"),
            },
          ]}
          hasFeedback
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="confirm"
          label={i18next.t("signup:Confirm")}
          dependencies={["password"]}
          hasFeedback
          rules={[
            {
              required: true,
              message: i18next.t("signup:Please confirm your password!"),
            },
            ({ getFieldValue }) => ({
              validator(rule, value) {
                if (!value || getFieldValue("password") === value) {
                  return Promise.resolve();
                }

                return Promise.reject(
                  i18next.t(
                    "signup:Your confirmed password is inconsistent with the password!"
                  )
                );
              },
            }),
          ]}
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="phone"
          label={i18next.t("general:Phone")}
          rules={[
            {
              required: true,
              message: i18next.t("signup:Please input your phone number!"),
            },
          ]}
        >
          <Input
            style={{
              width: "100%",
            }}
            addonBefore={`+${this.state.application?.organizationObj.phonePrefix}`}
            onChange={(e) => this.setState({ phone: e.target.value })}
          />
        </Form.Item>
        <Form.Item
          name="phoneCode"
          label={i18next.t("signup:phone code")}
          rules={[
            {
              required: true,
              message: i18next.t(
                "signup:Please input your phone verification code!"
              ),
            },
          ]}
        >
          <CountDownInput
            defaultButtonText={i18next.t("signup:send code")}
            onButtonClick={UserBackend.sendCode}
            onButtonClickArgs={[
              this.state.phone,
              "phone",
              application.organizationObj.owner +
                "/" +
                application.organizationObj.name,
            ]}
            coolDownTime={60}
          />
        </Form.Item>
        <Form.Item
          name="agreement"
          valuePropName="checked"
          {...tailFormItemLayout}
        >
          <Checkbox>
            {i18next.t("signup:Accept")}&nbsp;
            <Link to={"/agreement"}>{i18next.t("signup:Terms of Use")}</Link>
          </Checkbox>
        </Form.Item>
        <Form.Item {...tailFormItemLayout}>
          <Button type="primary" htmlType="submit">
            {i18next.t("account:Sign Up")}
          </Button>
          &nbsp;&nbsp;{i18next.t("signup:Have account?")}&nbsp;
          <Link
            onClick={() => {
              Setting.goToLogin(this, application);
            }}>
            {i18next.t("signup:sign in now")}
          </Link>
        </Form.Item>
      </Form>
    );
  }

  componentWillReceiveProps(nextProps, nextContext) {
    if (this.state.urlParams.urlWithProvider) {
      getUser(this.state.urlParams.owner, this.state.urlParams.user).then(
        (user) => {
          this.form.current.setFieldsValue({
            username: user.displayName.toLowerCase().replace(" ", "_"),
            name: user.displayName,
            email: user.email,
          });

          this.setState({
            provider: {
              type: this.getProviderFromUser(user),
              user: user.displayName,
              avatar: user.avatar,
            },
          });
        }
      );
    }
  }

  getProviderFromUser(userInfo) {
    const providerType = [
      { key: "google", name: "Google" },
      { key: "github", name: "GitHub" },
      { key: "qq", name: "QQ" },
      { key: "wechat", name: "WeChat" },
      { key: "facebook", name: "Facebook" },
      { key: "dingtalk", name: "DingTalk" },
      { key: "weibo", name: "Weibo" },
    ];

    let providerName = "";

    providerType.forEach((provider) => {
      if (userInfo[provider.key]) {
        providerName = provider.name;
      }
    });
    return providerName;
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return null;
    }

    return (
      <div>
        &nbsp;
        <Row>
          <Col span={24} style={{ display: "flex", justifyContent: "center" }}>
            <div style={{ marginTop: "10px", textAlign: "center" }}>
              {Setting.renderHelmet(application)}
              {Setting.renderLogo(application)}
              {this.renderForm(application)}
            </div>
          </Col>
        </Row>
      </div>
    );
  }
}

export default SignupPage;
