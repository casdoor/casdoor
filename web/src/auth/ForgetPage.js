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
import {Button, Col, Form, Input, Row, Select, Steps} from "antd";
import * as AuthBackend from "./AuthBackend";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Util from "./Util";
import * as Setting from "../Setting";
import i18next from "i18next";
import {CountDownInput} from "../common/CountDownInput";
import * as UserBackend from "../backend/UserBackend";
import {CheckCircleOutlined, KeyOutlined, LockOutlined, SolutionOutlined, UserOutlined} from "@ant-design/icons";
import CustomGithubCorner from "../CustomGithubCorner";
import {withRouter} from "react-router-dom";

const {Step} = Steps;
const {Option} = Select;

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
      name: "",
      email: "",
      isFixed: false,
      fixedContent: "",
      token: "",
      phone: "",
      emailCode: "",
      phoneCode: "",
      verifyType: null, // "email" or "phone"
      current: 0,
    };
  }

  UNSAFE_componentWillMount() {
    if (this.state.applicationName !== undefined) {
      this.getApplication();
    } else {
      Setting.showMessage("error", i18next.t("forget:Unknown forget type: ") + this.state.type);
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
      const username = forms.step1.getFieldValue("username");
      AuthBackend.getEmailAndPhone({
        application: forms.step1.getFieldValue("application"),
        organization: forms.step1.getFieldValue("organization"),
        username: username,
      }).then((res) => {
        if (res.status === "ok") {
          const phone = res.data.phone;
          const email = res.data.email;
          const saveFields = () => {
            if (this.state.isFixed) {
              forms.step2.setFieldsValue({email: this.state.fixedContent});
              this.setState({username: this.state.fixedContent});
            }
            this.setState({current: 1});
          };
          this.setState({phone: phone, email: email, username: res.data.name, name: res.data.name});

          if (phone !== "" && email === "") {
            this.setState({
              verifyType: "phone",
            });
          } else if (phone === "" && email !== "") {
            this.setState({
              verifyType: "email",
            });
          }

          switch (res.data2) {
          case "email":
            this.setState({isFixed: true, fixedContent: email, verifyType: "email"}, () => {saveFields();});
            break;
          case "phone":
            this.setState({isFixed: true, fixedContent: phone, verifyType: "phone"}, () => {saveFields();});
            break;
          default:
            saveFields();
            break;
          }
        } else {
          Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
        }
      });
      break;
    case "step2":
      const oAuthParams = Util.getOAuthGetParameters();
      const login = () => {
        AuthBackend.login({
          application: forms.step2.getFieldValue("application"),
          organization: forms.step2.getFieldValue("organization"),
          username: this.state.username,
          name: this.state.name,
          code: forms.step2.getFieldValue("emailCode"),
          phonePrefix: this.state.application?.organizationObj.phonePrefix,
          type: "login",
        }, oAuthParams).then(res => {
          if (res.status === "ok") {
            this.setState({current: 2, userId: res.data, username: res.data.split("/")[1]});
          } else {
            Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
          }
        });
      };
      if (this.state.verifyType === "email") {
        this.setState({username: this.state.email}, () => {login();});
      } else if (this.state.verifyType === "phone") {
        this.setState({username: this.state.phone}, () => {login();});
      }
      break;
    default:
      break;
    }
  }

  onFinish(values) {
    values.username = this.state.username;
    values.userOwner = this.state.application?.organizationObj.name;
    UserBackend.setPassword(values.userOwner, values.username, "", values?.newPassword).then(res => {
      if (res.status === "ok") {
        Setting.redirectToLoginPage(this.state.application, this.props.history);
      } else {
        Setting.showMessage("error", i18next.t(`signup:${res.msg}`));
      }
    });
  }

  onFinishFailed(values, errorFields) {}

  renderOptions() {
    const options = [];

    if (this.state.phone !== "") {
      options.push(
        <Option key={"phone"} value={"phone"}>
          &nbsp;&nbsp;{this.state.phone}
        </Option>
      );
    }

    if (this.state.email !== "") {
      options.push(
        <Option key={"email"} value={"email"}>
          &nbsp;&nbsp;{this.state.email}
        </Option>
      );
    }

    return options;
  }

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
          // eslint-disable-next-line no-console
          onFinishFailed={(errorInfo) => console.log(errorInfo)}
          initialValues={{
            application: application.name,
            organization: application.organization,
          }}
          style={{width: "300px"}}
          size="large"
        >
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="application"
            rules={[
              {
                required: true,
                message: i18next.t(
                  "forget:Please input your application!"
                ),
              },
            ]}
          />
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="organization"
            rules={[
              {
                required: true,
                message: i18next.t(
                  "forget:Please input your organization!"
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
          style={{width: "300px"}}
          size="large"
        >
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="application"
            rules={[
              {
                required: true,
                message: i18next.t(
                  "forget:Please input your application!"
                ),
              },
            ]}
          />
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="organization"
            rules={[
              {
                required: true,
                message: i18next.t(
                  "forget:Please input your organization!"
                ),
              },
            ]}
          />
          <Form.Item
            name="email" // use email instead of email/phone to adapt to RequestForm in account.go
            validateFirst
            hasFeedback
          >
            {
              this.state.isFixed ? <Input disabled /> :
                <Select
                  key={this.state.verifyType}
                  virtual={false} style={{textAlign: "left"}}
                  defaultValue={this.state.verifyType}
                  disabled={this.state.username === ""}
                  placeholder={i18next.t("forget:Choose email or phone")}
                  onChange={(value) => {
                    this.setState({
                      verifyType: value,
                    });
                  }}
                >
                  {
                    this.renderOptions()
                  }
                </Select>
            }
          </Form.Item>
          <Form.Item
            name="emailCode" // use emailCode instead of email/phoneCode to adapt to RequestForm in account.go
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
                onButtonClickArgs={[this.state.email, "email", Setting.getApplicationName(this.state.application), this.state.name]}
                application={application}
              />
            ) : (
              <CountDownInput
                disabled={this.state.username === "" || this.state.verifyType === ""}
                onButtonClickArgs={[this.state.phone, "phone", Setting.getApplicationName(this.state.application), this.state.name]}
                application={application}
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
          style={{width: "300px"}}
          size="large"
        >
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="application"
            rules={[
              {
                required: true,
                message: i18next.t(
                  "forget:Please input your application!"
                ),
              },
            ]}
          />
          <Form.Item
            style={{height: 0, visibility: "hidden"}}
            name="organization"
            rules={[
              {
                required: true,
                message: i18next.t(
                  "forget:Please input your organization!"
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
              ({getFieldValue}) => ({
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
            <Button block type="primary" htmlType="submit" disabled={this.state.userId === ""}>
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
      <div className="loginBackground" style={{backgroundImage: Setting.inIframe() || Setting.isMobile() ? null : `url(${application.formBackgroundUrl})`}}>
        <CustomGithubCorner />
        <div className="login-content forget-content" style={{padding: Setting.isMobile() ? "0" : null, boxShadow: Setting.isMobile() ? "none" : null}}>
          <Row>
            <Col span={24} style={{justifyContent: "center"}}>
              <Row>
                <Col span={24}>
                  <div style={{marginTop: "80px", marginBottom: "10px", textAlign: "center"}}>
                    {
                      Setting.renderHelmet(application)
                    }
                    {
                      Setting.renderLogo(application)
                    }
                  </div>
                </Col>
              </Row>
              <Row>
                <Col span={24}>
                  <div style={{textAlign: "center", fontSize: "28px"}}>
                    {i18next.t("forget:Retrieve password")}
                  </div>
                </Col>
              </Row>
              <Row>
                <Col span={24}>
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
            </Col>
            <Col span={24} style={{display: "flex", justifyContent: "center"}}>
              <div style={{marginTop: "10px", textAlign: "center"}}>
                {this.renderForm(application)}
              </div>
            </Col>
          </Row>
        </div>
      </div>
    );
  }
}

export default withRouter(ForgetPage);
