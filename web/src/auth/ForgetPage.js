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
import { Button, Col, Divider, Form, Select, Input, Row, Steps } from "antd";
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
  SolutionOutlined,
  UserOutlined,
} from "@ant-design/icons";

const { Step } = Steps;
const { Option } = Select;

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
      userId: "",
      username: "",
      email: "",
      isFixed: false,
      fixedContent: "",
      token: "",
      phone: "",
      emailCode: "",
      phoneCode: "",
      verifyType: "", // "email" or "phone"
      current: 0,
    };
  }

  UNSAFE_componentWillMount() {
    if (this.state.applicationName !== undefined) {
      this.getApplication();
    } else {
      Util.showMessage(
          "error",
          i18next.t(`forget:Unknown forgot type: `) + this.state.type
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

  onFormFinish(name, info, forms) {
      switch (name) {
          case "step1":
              const username = forms.step1.getFieldValue("username")
              AuthBackend.getEmailAndPhone({
                  application: forms.step1.getFieldValue("application"),
                  organization: forms.step1.getFieldValue("organization"),
                  username: username
              }).then((res) => {
                  if (res.status === "ok") {
                      this.setState({phone: res.data.phone, email: res.data.email, username: res.data.name});
                      switch (res.data2) {
                          case "email":
                              this.setState({isFixed: true, fixedContent: res.data.email, verifyType: "email"});
                              break
                          case "phone":
                              this.setState({isFixed: true, fixedContent: res.data.phone, verifyType: "phone"});
                              break
                      }
                      if (this.state.isFixed) {
                          forms.step2.setFieldsValue({email: this.state.fixedContent})
                      }
                      this.setState({current: 1})
                  } else {
                      Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
                  }
              });
              break;
          case "step2":
              const oAuthParams = Util.getOAuthGetParameters();
              AuthBackend.login({
                  application: forms.step2.getFieldValue("application"),
                  organization: forms.step2.getFieldValue("organization"),
                  email: forms.step2.getFieldValue("email"),
                  emailCode: forms.step2.getFieldValue("emailCode"),
                  phonePrefix: this.state.application?.organizationObj.phonePrefix,
                  username: this.state.username,
                  type: "login"
              }, oAuthParams).then(res => {
                  if (res.status === "ok") {
                      this.setState({current: 2, userId: res.data, username: res.data.split("/")[1]})
                  } else {
                      Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
                  }
              })
              break
      }
  }

  onFinish(values) {
    values.username = this.state.username;
    values.userOwner = this.state.application?.organizationObj.name
    UserBackend.setPassword(values.userOwner, values.username, "", values?.newPassword).then(res => {
        if (res.status === "ok") {
            Setting.goToLogin(this, this.state.application);
        } else {
            Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
        }
    })
  }

  onFinishFailed(values, errorFields) {}

  renderForm(application) {
    return (
        <Form.Provider onFormFinish={(name, {info, forms}) => {
            this.onFormFinish(name, info, forms);
        }}>
          {/* STEP 1: input username -> get email & phone */}
          <Form
              hidden={this.state.current !== 0}
              ref={this.form}
              name="step1"
              onFinishFailed={(errorInfo) => console.log(errorInfo)}
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
                        `forget:Please input your application!`
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
                        `forget:Please input your organization!`
                    ),
                  },
                ]}
            />
            <Form.Item
                name="username"
                rules={[
                  {
                    required: true,
                    message: i18next.t(
                        "forget:Please input your username!"
                    ),
                    whitespace: true,
                  },
                ]}
            >
              <Input
                  onChange={(e) => {
                    this.setState({
                      username: e.target.value,
                    });
                  }}
                  prefix={<UserOutlined />}
                  placeholder={i18next.t("login:username, Email or phone")}
              />
            </Form.Item>
            <br />
            <Form.Item>
              <Button block type="primary" htmlType="submit">
                {i18next.t("forget:Next Step")}
              </Button>
            </Form.Item>
          </Form>

          {/* STEP 2: verify email or phone */}
          <Form
              hidden={this.state.current !== 1}
              ref={this.form}
              name="step2"
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
                        `forget:Please input your application!`
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
                        `forget:Please input your organization!`
                    ),
                  },
                ]}
            />
            <Form.Item
                name="email" //use email instead of email/phone to adapt to RequestForm in account.go
                validateFirst
                hasFeedback
            >
               {
                  this.state.isFixed ? <Input disabled/> :
                     <Select
                         disabled={this.state.username === ""}
                         placeholder={i18next.t(
                             "forget:Choose email verification or mobile verification"
                         )}
                         onChange={(value) => {
                           if (value === this.state.phone) {
                             this.setState({ verifyType: "phone" });
                           }
                           if (value === this.state.email) {
                             this.setState({ verifyType: "email" });
                           }
                         }}
                         allowClear
                         style={{ textAlign: "left" }}
                     >
                       <Option key={1} value={this.state.phone}>
                         {this.state.phone.replace(/(\d{3})\d*(\d{4})/,'$1****$2')}
                       </Option>
                       <Option key={2} value={this.state.email}>
                         {this.state.email.split("@")[0].length>2?
                             this.state.email.replace(/(?<=.)[^@]+(?=.@)/, "*****"):
                             this.state.email.replace(/(\w?@)/, "*@")}
                       </Option>
                     </Select>
               }
            </Form.Item>
            <Form.Item
                name="emailCode" //use emailCode instead of email/phoneCode to adapt to RequestForm in account.go
                rules={[
                  {
                    required: true,
                    message: i18next.t(
                        "code:Please input your verification code!"
                    ),
                  },
                ]}
            >
              {this.state.verifyType === "email" ? (
                  <CountDownInput
                      disabled={this.state.username === "" || this.state.verifyType === ""}
                      placeHolder={i18next.t("code:Verify code")}
                      defaultButtonText={i18next.t("code:Send code")}
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
                      disabled={this.state.username === "" || this.state.verifyType === ""}
                      placeHolder={i18next.t("code:Verify code")}
                      defaultButtonText={i18next.t("code:Send code")}
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
            <br />
            <Form.Item>
              <Button
                  block
                  type="primary"
                  htmlType="submit"
              >
                {i18next.t("forget:Next Step")}
              </Button>
            </Form.Item>
          </Form>

          {/* STEP 3 */}
          <Form
              hidden={this.state.current !== 2}
              ref={this.form}
              name="step3"
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
                        `forget:Please input your application!`
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
                        `forget:Please input your organization!`
                    ),
                  },
                ]}
            />
            <Form.Item
                name="newPassword"
                hidden={this.state.current !== 2}
                rules={[
                  {
                    required: true,
                    message: i18next.t(
                        "forget:Please input your password!"
                    ),
                  },
                ]}
                hasFeedback
            >
              <Input.Password
                  disabled={this.state.userId === ""}
                  prefix={<LockOutlined />}
                  placeholder={i18next.t("forget:Password")}
              />
            </Form.Item>
            <Form.Item
                name="confirm"
                dependencies={["newPassword"]}
                hasFeedback
                rules={[
                  {
                    required: true,
                    message: i18next.t(
                        "forget:Please confirm your password!"
                    ),
                  },
                  ({ getFieldValue }) => ({
                    validator(rule, value) {
                      if (!value || getFieldValue("newPassword") === value) {
                        return Promise.resolve();
                      }
                      return Promise.reject(
                          i18next.t(
                              "forget:Your confirmed password is inconsistent with the password!"
                          )
                      );
                    },
                  }),
                ]}
            >
              <Input.Password
                  disabled={this.state.userId === ""}
                  prefix={<CheckCircleOutlined />}
                  placeholder={i18next.t("forget:Confirm")}
              />
            </Form.Item>
            <br />
            <Form.Item hidden={this.state.current !== 2}>
              <Button block type="primary"  htmlType="submit" disabled={this.state.userId === ""}>
                {i18next.t("forget:Change Password")}
              </Button>
            </Form.Item>
          </Form>
        </Form.Provider>
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
            {i18next.t("forget:Retrieve password")}
          </Divider>
          <Row>
            <Col span={24} style={{ display: "flex", justifyContent: "center" }}>
              <Steps
                  current={this.state.current}
                  style={{
                    width: "90%",
                    maxWidth: "500px",
                    margin: "auto",
                    marginTop: "80px",
                  }}
              >
                <Step
                    title={i18next.t("forget:Account")}
                    icon={<UserOutlined />}
                />
                <Step
                    title={i18next.t("forget:Verify")}
                    icon={<SolutionOutlined />}
                />
                <Step
                    title={i18next.t("forget:Reset")}
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
