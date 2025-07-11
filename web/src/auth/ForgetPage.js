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
import {Button, Col, Form, Input, Popover, Row, Select, Steps} from "antd";
import * as AuthBackend from "./AuthBackend";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Util from "./Util";
import * as Setting from "../Setting";
import i18next from "i18next";
import {SendCodeInput} from "../common/SendCodeInput";
import * as UserBackend from "../backend/UserBackend";
import {ArrowLeftOutlined, CheckCircleOutlined, KeyOutlined, LockOutlined, SolutionOutlined, UserOutlined} from "@ant-design/icons";
import CustomGithubCorner from "../common/CustomGithubCorner";
import {withRouter} from "react-router-dom";
import * as PasswordChecker from "../common/PasswordChecker";

const {Option} = Select;

class ForgetPage extends React.Component {
  constructor(props) {
    super(props);
    const queryParams = new URLSearchParams(location.search);
    this.state = {
      classes: props,
      applicationName: props.applicationName ?? props.match.params?.applicationName,
      msg: null,
      name: props.account ? props.account.name : queryParams.get("username"),
      username: props.account ? props.account.name : "",
      phone: "",
      email: "",
      dest: "",
      isVerifyTypeFixed: false,
      verifyType: "", // "email", "phone"
      current: queryParams.get("code") ? 2 : 0,
      code: queryParams.get("code"),
      queryParams: queryParams,
    };
    this.form = React.createRef();
  }

  componentDidMount() {
    if (this.getApplicationObj() === undefined) {
      if (this.state.applicationName !== undefined) {
        this.getApplication();
      } else {
        Setting.showMessage("error", i18next.t("forget:Unknown forget type") + ": " + this.state.type);
      }
    }
  }

  getApplication() {
    if (this.state.applicationName === undefined) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }
        this.onUpdateApplication(res.data);
      });
  }
  getApplicationObj() {
    return this.props.application;
  }

  onUpdateApplication(application) {
    this.props.onUpdateApplication(application);
  }

  onFormFinish(name, info, forms) {
    switch (name) {
    case "step1":
      const username = forms.step1.getFieldValue("username");
      AuthBackend.getEmailAndPhone(forms.step1.getFieldValue("organization"), username)
        .then((res) => {
          if (res.status === "ok") {
            const phone = res.data.phone;
            const email = res.data.email;

            if (!phone && !email) {
              Setting.showMessage("error", "no verification method!");
            } else {
              this.setState({
                name: res.data.name,
                phone: phone,
                email: email,
              });

              const saveFields = (type, dest, fixed) => {
                this.setState({
                  verifyType: type,
                  isVerifyTypeFixed: fixed,
                  dest: dest,
                });
              };

              switch (res.data2) {
              case "email":
                saveFields("email", email, true);
                break;
              case "phone":
                saveFields("phone", phone, true);
                break;
              case "username":
                phone !== "" ? saveFields("phone", phone, false) : saveFields("email", email, false);
              }

              this.setState({
                current: 1,
              });
            }
          } else {
            Setting.showMessage("error", res.msg);
          }
        });
      break;
    case "step2":
      UserBackend.verifyCode({
        application: forms.step2.getFieldValue("application"),
        organization: forms.step2.getFieldValue("organization"),
        username: forms.step2.getFieldValue("dest"),
        name: this.state.name,
        code: forms.step2.getFieldValue("code"),
        type: "login",
      }).then(res => {
        if (res.status === "ok") {
          this.setState({current: 2, code: forms.step2.getFieldValue("code")});
        } else {
          Setting.showMessage("error", res.msg);
        }
      });

      break;
    default:
      break;
    }
  }

  async onFinish(values) {
    values.username = this.state.name;
    values.userOwner = this.getApplicationObj()?.organizationObj.name;

    if (this.state.queryParams.get("code")) {
      const res = await UserBackend.verifyCode({
        application: this.getApplicationObj().name,
        organization: values.userOwner,
        username: this.state.queryParams.get("dest"),
        name: this.state.name,
        code: this.state.code,
        type: "login",
      });

      if (res.status !== "ok") {
        Setting.showMessage("error", res.msg);
        return;
      }
    }

    UserBackend.setPassword(values.userOwner, values.username, "", values?.newPassword, this.state.code).then(res => {
      if (res.status === "ok") {
        const linkInStorage = sessionStorage.getItem("signinUrl");
        if (linkInStorage !== null && linkInStorage !== "") {
          Setting.goToLinkSoft(this, linkInStorage);
        } else {
          Setting.redirectToLoginPage(this.getApplicationObj(), this.props.history);
        }
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  }

  onFinishFailed(values, errorFields) {}

  renderOptions() {
    const options = [];

    if (this.state.phone !== "") {
      options.push(
        <Option key={"phone"} value={this.state.phone} >
          &nbsp;&nbsp;{this.state.phone}
        </Option>
      );
    }

    if (this.state.email !== "") {
      options.push(
        <Option key={"email"} value={this.state.email} >
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
        {this.state.current === 0 ?
          <Form
            ref={this.form}
            name="step1"
            // eslint-disable-next-line no-console
            onFinishFailed={(errorInfo) => console.log(errorInfo)}
            initialValues={{
              application: application.name,
              organization: application.organization,
              username: this.state.name,
            }}
            style={{width: "300px"}}
            size="large"
          >
            <Form.Item
              hidden
              name="application"
              rules={[
                {
                  required: true,
                  message: i18next.t("application:Please input your application!"),
                },
              ]}
            />
            <Form.Item
              hidden
              name="organization"
              rules={[
                {
                  required: true,
                  message: i18next.t("application:Please input your organization!"),
                },
              ]}
            />
            <Form.Item
              name="username"
              rules={[
                {
                  required: true,
                  message: i18next.t("forget:Please input your username!"),
                  whitespace: true,
                },
              ]}
            >
              <Input
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
          </Form> : null}

        {/* STEP 2: verify email or phone */}
        {this.state.current === 1 ? <Form
          ref={this.form}
          name="step2"
          onFinishFailed={(errorInfo) =>
            this.onFinishFailed(
              errorInfo.values,
              errorInfo.errorFields,
              errorInfo.outOfDate
            )
          }
          onValuesChange={(changedValues, allValues) => {
            if (!changedValues.dest) {
              return;
            }
            const verifyType = changedValues.dest?.indexOf("@") === -1 ? "phone" : "email";
            this.setState({
              dest: changedValues.dest,
              verifyType: verifyType,
            });
          }}
          initialValues={{
            application: application.name,
            organization: application.organization,
            dest: this.state.dest,
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
                message: i18next.t("application:Please input your application!"),
              },
            ]}
          />
          <Form.Item
            hidden
            name="organization"
            rules={[
              {
                required: true,
                message: i18next.t("application:Please input your organization!"),
              },
            ]}
          />
          <Form.Item
            name="dest"
            validateFirst
            hasFeedback
          >
            {
              <Select virtual={false}
                disabled={this.state.isVerifyTypeFixed}
                style={{textAlign: "left"}}
                placeholder={i18next.t("forget:Choose email or phone")}
              >
                {
                  this.renderOptions()
                }
              </Select>
            }
          </Form.Item>
          <Form.Item
            name="code"
            rules={[
              {
                required: true,
                message: i18next.t("code:Please input your verification code!"),
              },
            ]}
          >
            <SendCodeInput disabled={this.state.dest === ""}
              method={"forget"}
              onButtonClickArgs={[this.state.dest, this.state.verifyType, Setting.getApplicationName(this.getApplicationObj()), this.state.name]}
              application={application}
            />
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
        </Form> : null}

        {/* STEP 3 */}
        {this.state.current === 2 ?
          <Form
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
              hidden
              name="application"
              rules={[
                {
                  required: true,
                  message: i18next.t("application:Please input your application!"),
                },
              ]}
            />
            <Form.Item
              hidden
              name="organization"
              rules={[
                {
                  required: true,
                  message: i18next.t("application:Please input your organization!"),
                },
              ]}
            />
            <Popover placement={window.innerWidth >= 960 ? "right" : "top"} content={this.state.passwordPopover} open={this.state.passwordPopoverOpen}>
              <Form.Item
                name="newPassword"
                hidden={this.state.current !== 2}
                rules={[
                  {
                    required: true,
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
                <Input.Password
                  prefix={<LockOutlined />}
                  placeholder={i18next.t("general:Password")}
                  onChange={(e) => {
                    this.setState({
                      passwordPopover: PasswordChecker.renderPasswordPopover(application.organizationObj.passwordOptions, e.target.value),
                    });
                  }}
                  onFocus={() => {
                    this.setState({
                      passwordPopoverOpen: application.organizationObj.passwordOptions?.length > 0,
                      passwordPopover: PasswordChecker.renderPasswordPopover(application.organizationObj.passwordOptions, this.form.current?.getFieldValue("newPassword") ?? ""),
                    });
                  }}
                  onBlur={() => {
                    this.setState({
                      passwordPopoverOpen: false,
                    });
                  }}
                />
              </Form.Item>
            </Popover>
            <Form.Item
              name="confirm"
              dependencies={["newPassword"]}
              hasFeedback
              rules={[
                {
                  required: true,
                  message: i18next.t("signup:Please confirm your password!"),
                },
                ({getFieldValue}) => ({
                  validator(rule, value) {
                    if (!value || getFieldValue("newPassword") === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(
                      i18next.t("signup:Your confirmed password is inconsistent with the password!")
                    );
                  },
                }),
              ]}
            >
              <Input.Password
                prefix={<CheckCircleOutlined />}
                placeholder={i18next.t("signup:Confirm")}
              />
            </Form.Item>
            <br />
            <Form.Item hidden={this.state.current !== 2}>
              <Button block type="primary" htmlType="submit">
                {i18next.t("forget:Change Password")}
              </Button>
            </Form.Item>
          </Form> : null}
      </Form.Provider>
    );
  }

  stepBack() {
    if (this.state.current > 0) {
      this.setState({
        current: this.state.current - 1,
      });
    } else if (this.props.history.length > 1) {
      this.props.history.goBack();
    } else {
      Setting.redirectToLoginPage(this.getApplicationObj(), this.props.history);
    }
  }

  render() {
    const application = this.getApplicationObj();
    if (application === undefined) {
      return null;
    }
    if (application === null) {
      return Util.renderMessageLarge(this, this.state.msg);
    }

    return (
      <React.Fragment>
        <CustomGithubCorner />
        <div className="forget-content" style={{padding: Setting.isMobile() ? "0" : null, boxShadow: Setting.isMobile() ? "none" : null}}>
          {Setting.inIframe() || Setting.isMobile() ? null : <div dangerouslySetInnerHTML={{__html: application.formCss}} />}
          {Setting.inIframe() || !Setting.isMobile() ? null : <div dangerouslySetInnerHTML={{__html: application.formCssMobile}} />}
          <Button type="text"
            style={{position: "relative", left: Setting.isMobile() ? "10px" : "-90px", top: 0}}
            icon={<ArrowLeftOutlined style={{fontSize: "24px"}} />}
            size={"large"}
            onClick={() => {this.stepBack();}}
          />
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
                    {i18next.t("forget:Reset password")}
                  </div>
                </Col>
              </Row>
              <Row>
                <Col span={24}>
                  <Steps
                    current={this.state.current}
                    items={[
                      {
                        title: i18next.t("forget:Account"),
                        icon: <UserOutlined />,
                      },
                      {
                        title: i18next.t("forget:Verify"),
                        icon: <SolutionOutlined />,
                      },
                      {
                        title: i18next.t("forget:Reset"),
                        icon: <KeyOutlined />,
                      },
                    ]}
                    style={{
                      width: "90%",
                      maxWidth: "500px",
                      margin: "auto",
                      marginTop: "80px",
                    }}
                  >
                  </Steps>
                </Col>
              </Row>
            </Col>
            <Col span={24} style={{display: "flex", justifyContent: "center"}}>
              <div style={{marginTop: "40px", textAlign: "center"}}>
                {this.renderForm(application)}
              </div>
            </Col>
          </Row>
        </div>
      </React.Fragment>
    );
  }
}

export default withRouter(ForgetPage);
