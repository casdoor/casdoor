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
import { Button, Col, Divider, Form, Input, Row, Steps } from "antd";
import * as AuthBackend from "./AuthBackend";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Util from "./Util";
import * as Setting from "../Setting";
import i18next from "i18next";
import { CountDownInput } from "../component/CountDownInput";
import * as UserBackend from "../backend/UserBackend";
import {
  CheckCircleOutlined,
  KeyOutlined,
  LockOutlined,
  MailOutlined,
  SolutionOutlined,
  UserOutlined,
} from "@ant-design/icons";

const { Step } = Steps;

class ForgetPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      account: props.account,
      applicationName:
        props.applicationName !== undefined
          ? props.applicationName
          : props.match === undefined
          ? null
          : props.match.params.applicationName,
      application: null,
      msg: null,
      email: "",
      phone: "",
      emailCode: "",
      phoneCode: "",
      verifyType: "phone", // "email" or "phone"
      current: 0,
    };
  }

  UNSAFE_componentWillMount() {
    if (this.state.applicationName !== undefined) {
      this.getApplication();
    } else {
      Util.showMessage(
        "error",
        i18next.t(`forgetPassword:Unknown forgot type: `) + this.state.type
      );
    }
  }

  getApplication() {
    if (this.state.applicationName === null) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName).then(
      (application) => {
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

  onFinish(values) {
    values.phonePrefix = this.state.application?.organizationObj.phonePrefix;
    AuthBackend.forgetPassword(values).then((res) => {
      if (res.status === "ok") {
        Setting.goToLogin(this, this.state.application);
      } else {
        Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
      }
    });
  }

  onFinishFailed(values, errorFields) {}

  onChange = (current) => {
    this.setState({ current: current });
  };

  renderForm(application) {
    return (
      <Form
        ref={this.form}
        name="forgetPassword"
        onFinish={(values) => this.onFinish(values)}
        onFinishFailed={(errorInfo) =>
          this.onFinishFailed(
            errorInfo.values,
            errorInfo.errorFields,
            errorInfo.outOfDate
          )
        }
        initialValues={{
          application: application.name,
          organization: application.organization,
        }}
        style={{ width: "300px" }}
        size="large"
      >
        <Form.Item
          style={{ height: 0, visibility: "hidden" }}
          name="application"
          rules={[
            {
              required: true,
              message: i18next.t(
                `forgetPassword:Please input your application!`
              ),
            },
          ]}
        />
        <Form.Item
          style={{ height: 0, visibility: "hidden" }}
          name="organization"
          rules={[
            {
              required: true,
              message: i18next.t(
                `forgetPassword:Please input your organization!`
              ),
            },
          ]}
        />
        <Form.Item
          name="username"
          hidden={this.state.current !== 0}
          rules={[
            {
              required: true,
              message: i18next.t("forgetPassword:Please input your username!"),
              whitespace: true,
            },
          ]}
        >
          <Input
            prefix={<UserOutlined />}
            placeholder={i18next.t("signup:Username")}
          />
        </Form.Item>
        <Form.Item
          name="email" //use email instead of email/phone to adapt to RequestForm in account.go
          validateFirst
          hasFeedback
          hidden={this.state.current !== 1}
          rules={[
            {
              message: i18next.t(
                "forgetPassword:Please input your Email/Phone string!"
              ),
              required: true,
              validator: (_, value) =>
                /(^1[0-9]{10}$)|(^\w+([-+.']\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$)/.test(
                  value
                )
                  ? Promise.resolve()
                  : Promise.reject("Email/Phone's format wrong!"),
            },
          ]}
        >
          <Input
            prefix={<MailOutlined />}
            placeholder={i18next.t("forgetPassword:Email/Phone")}
            onChange={(e) => {
              if (e.target.value.indexOf("@") !== -1) {
                this.setState({ verifyType: "email", email: e.target.value });
              } else {
                this.setState({ verifyType: "phone", phone: e.target.value });
              }
            }}
          />
        </Form.Item>
        <Form.Item
          hidden={this.state.current !== 1}
          name="emailCode" //use emailCode instead of email/phoneCode to adapt to RequestForm in account.go
          rules={[
            {
              required: true,
              message: i18next.t(
                "forgetPassword:Please input your verification code!"
              ),
            },
          ]}
        >
          {this.state.verifyType === "email" ? (
            <CountDownInput
              prefix={"AuditOutlined"}
              placeHolder={i18next.t("forgetPassword:Verify code")}
              defaultButtonText={i18next.t("forgetPassword:send code")}
              onButtonClick={UserBackend.sendCode}
              onButtonClickArgs={[
                this.state.email,
                "email",
                this.state.application?.organizationObj.owner +
                  "/" +
                  this.state.application?.organizationObj.name,
              ]}
              coolDownTime={60}
            />
          ) : (
            <CountDownInput
              prefix={"AuditOutlined"}
              placeHolder={i18next.t("forgetPassword:Verify code")}
              defaultButtonText={i18next.t("forgetPassword:send code")}
              onButtonClick={UserBackend.sendCode}
              onButtonClickArgs={[
                this.state.phone,
                "phone",
                this.state.application?.organizationObj.owner +
                  "/" +
                  this.state.application?.organizationObj.name,
              ]}
              coolDownTime={60}
            />
          )}
        </Form.Item>
        <Form.Item
          name="password"
          hidden={this.state.current !== 2}
          rules={[
            {
              required: true,
              message: i18next.t("forgetPassword:Please input your password!"),
            },
          ]}
          hasFeedback
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder={i18next.t("forgetPassword:Password")}
          />
        </Form.Item>
        <Form.Item
          name="confirm"
          dependencies={["password"]}
          hidden={this.state.current !== 2}
          hasFeedback
          rules={[
            {
              required: true,
              message: i18next.t(
                "forgetPassword:Please confirm your password!"
              ),
            },
            ({ getFieldValue }) => ({
              validator(rule, value) {
                if (!value || getFieldValue("password") === value) {
                  return Promise.resolve();
                }
                return Promise.reject(
                  i18next.t(
                    "forgetPassword:Your confirmed password is inconsistent with the password!"
                  )
                );
              },
            }),
          ]}
        >
          <Input.Password
            prefix={<CheckCircleOutlined />}
            placeholder={i18next.t("forgetPassword:Confirm")}
          />
        </Form.Item>
        <br />
        <div hidden={this.state.current === 2}>
          <Button
            block
            type="primary"
            onClick={() => {
              this.setState({ current: this.state.current + 1 });
            }}
          >
            {i18next.t("forgetPassword:Next Step")}
          </Button>
        </div>
        <Form.Item hidden={this.state.current !== 2}>
          <Button block type="primary" htmlType="submit">
            {i18next.t("forgetPassword:Change Password")}
          </Button>
        </Form.Item>
      </Form>
    );
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return Util.renderMessageLarge(this, this.state.msg);
    }

    return (
      <>
        <Divider style={{ fontSize: "28px" }}>
          {i18next.t("forgetPassword:Retrieve password")}
        </Divider>
        <Row>
          <Col span={24} style={{ display: "flex", justifyContent: "center" }}>
            <Steps
              current={this.state.current}
              onChange={this.onChange}
              style={{
                width: "90%",
                maxWidth: "500px",
                margin: "auto",
                marginTop: "80px",
              }}
            >
              <Step
                title={i18next.t("forgetPassword:Account")}
                icon={<UserOutlined />}
              />
              <Step
                title={i18next.t("forgetPassword:Verify")}
                icon={<SolutionOutlined />}
              />
              <Step
                title={i18next.t("forgetPassword:Reset")}
                icon={<KeyOutlined />}
              />
            </Steps>
          </Col>
        </Row>
        <Row>
          <Col span={24} style={{ display: "flex", justifyContent: "center" }}>
            <div style={{ marginTop: "10px", textAlign: "center" }}>
              {Setting.renderHelmet(application)}
              {this.renderForm(application)}
            </div>
          </Col>
        </Row>
      </>
    );
  }
}

export default ForgetPage;
