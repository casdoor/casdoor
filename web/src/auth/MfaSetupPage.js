// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import React, {useEffect, useState} from "react";
import {Button, Col, Form, Input, Result, Row, Steps} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as MfaBackend from "../backend/MfaBackend";
import {CheckOutlined, KeyOutlined, LockOutlined, UserOutlined} from "@ant-design/icons";

import * as UserBackend from "../backend/UserBackend";
import {MfaSmsVerifyForm, MfaTotpVerifyForm, mfaSetup} from "./MfaVerifyForm";
import * as ApplicationBackend from "../backend/ApplicationBackend";

export const EmailMfaType = "email";
export const SmsMfaType = "sms";
export const TotpMfaType = "app";

function CheckPasswordForm({user, onSuccess, onFail}) {
  const [form] = Form.useForm();

  const onFinish = ({password}) => {
    const data = {...user, password};
    UserBackend.checkUserPassword(data)
      .then((res) => {
        if (res.status === "ok") {
          onSuccess(res);
        } else {
          onFail(res);
        }
      })
      .finally(() => {
        form.setFieldsValue({password: ""});
      });
  };

  return (
    <Form
      form={form}
      style={{width: "300px", marginTop: "20px"}}
      onFinish={onFinish}
    >
      <Form.Item
        name="password"
        rules={[{required: true, message: i18next.t("login:Please input your password!")}]}
      >
        <Input.Password
          prefix={<LockOutlined />}
          placeholder={i18next.t("general:Password")}
        />
      </Form.Item>

      <Form.Item>
        <Button
          style={{marginTop: 24}}
          loading={false}
          block
          type="primary"
          htmlType="submit"
        >
          {i18next.t("forget:Next Step")}
        </Button>
      </Form.Item>
    </Form>
  );
}

export function MfaVerifyForm({mfaType, application, user, onSuccess, onFail}) {
  const [form] = Form.useForm();
  const [mfaProps, setMfaProps] = useState({mfaType: mfaType});

  useEffect(() => {
    if (mfaType === SmsMfaType) {
      setMfaProps({
        mfaType: mfaType,
        secret: user.phone,
        countryCode: user.countryCode,
      });
    }

    if (mfaType === EmailMfaType) {
      setMfaProps({
        mfaType: mfaType,
        secret: user.email,
      });
    }
  }, [mfaType]);

  const onFinish = ({passcode}) => {
    const data = {passcode, mfaType: mfaType, ...user};
    MfaBackend.MfaSetupVerify(data)
      .then((res) => {
        if (res.status === "ok") {
          onSuccess(res);
        } else {
          onFail(res);
        }
      })
      .catch((error) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      })
      .finally(() => {
        form.setFieldsValue({passcode: ""});
      });
  };

  if (mfaType === null || mfaType === undefined || mfaProps.secret === undefined) {
    return <div></div>;
  }

  if (mfaType === SmsMfaType || mfaType === EmailMfaType) {
    return <MfaSmsVerifyForm onFinish={onFinish} application={application} user={user} method={mfaSetup} mfaProps={mfaProps} />;
  } else if (mfaType === TotpMfaType) {
    return <MfaTotpVerifyForm onFinish={onFinish} />;
  } else {
    return <div></div>;
  }
}

function EnableMfaForm({user, mfaType, recoveryCodes, onSuccess, onFail}) {
  const [loading, setLoading] = useState(false);
  const requestEnableTotp = () => {
    const data = {
      mfaType,
      ...user,
    };
    setLoading(true);
    MfaBackend.MfaSetupEnable(data).then(res => {
      if (res.status === "ok") {
        onSuccess(res);
      } else {
        onFail(res);
      }
    }
    ).finally(() => {
      setLoading(false);
    });
  };

  return (
    <div style={{width: "400px"}}>
      <p>{i18next.t("mfa:Please save this recovery code. Once your device cannot provide an authentication code, you can reset mfa authentication by this recovery code")}</p>
      <br />
      <code style={{fontStyle: "solid"}}>{recoveryCodes[0]}</code>
      <Button style={{marginTop: 24}} loading={loading} onClick={() => {
        requestEnableTotp();
      }} block type="primary">
        {i18next.t("general:Enable")}
      </Button>
    </div>
  );
}

class MfaSetupPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      account: props.account,
      applicationName: (props.applicationName ?? props.account?.signupApplication) ?? "",
      isAuthenticated: props.isAuthenticated ?? false,
      isPromptPage: props.isPromptPage,
      redirectUri: props.redirectUri,
      current: props.current ?? 0,
      mfaType: props.mfaType ?? new URLSearchParams(props.location?.search)?.get("mfaType") ?? SmsMfaType,
      mfaProps: null,
    };
  }

  componentDidMount() {
    this.getApplication();
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    if (this.state.isAuthenticated === true && this.state.mfaProps === null) {
      MfaBackend.MfaSetupInitiate({
        mfaType: this.state.mfaType,
        ...this.getUser(),
      }).then((res) => {
        if (res.status === "ok") {
          this.setState({
            mfaProps: res.data,
          });
        } else {
          Setting.showMessage("error", i18next.t("mfa:Failed to initiate MFA"));
        }
      });
    }
  }

  getApplication() {
    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        if (application !== null) {
          this.setState({
            application: application,
          });
        } else {
          Setting.showMessage("error", i18next.t("mfa:Failed to get application"));
        }
      });
  }

  getUser() {
    return {
      name: this.state.account.name,
      owner: this.state.account.owner,
    };
  }

  renderStep() {
    switch (this.state.current) {
    case 0:
      return <CheckPasswordForm
        user={this.getUser()}
        onSuccess={() => {
          this.setState({
            current: this.state.current + 1,
            isAuthenticated: true,
          });
        }}
        onFail={(res) => {
          Setting.showMessage("error", i18next.t("mfa:Failed to initiate MFA"));
        }}
      />;
    case 1:
      if (!this.state.isAuthenticated) {
        return null;
      }

      return (
        <div>
          <MfaVerifyForm
            mfaType={this.state.mfaType}
            application={this.state.application}
            user={this.props.account}
            onSuccess={() => {
              this.setState({
                current: this.state.current + 1,
              });
            }}
            onFail={(res) => {
              Setting.showMessage("error", i18next.t("general:Failed to verify"));
            }}
          />
          <Col span={24} style={{display: "flex", justifyContent: "left"}}>
            {(this.state.mfaType === EmailMfaType || this.props.account.mfaEmailEnabled) ? null :
              <Button type={"link"} onClick={() => {
                this.props.history.push(`/mfa-authentication/setup?mfaType=${EmailMfaType}`);
                this.setState({
                  mfaType: EmailMfaType,
                });
              }
              }>{i18next.t("mfa:Use Email")}</Button>
            }
            {
              (this.state.mfaType === SmsMfaType || this.props.account.mfaPhoneEnabled) ? null :
                <Button type={"link"} onClick={() => {
                  this.props.history.push(`/mfa-authentication/setup?mfaType=${SmsMfaType}`);
                  this.setState({
                    mfaType: SmsMfaType,
                  });
                }
                }>{i18next.t("mfa:Use SMS")}</Button>
            }
          </Col>
        </div>
      );
    case 2:
      if (!this.state.isAuthenticated) {
        return null;
      }

      return <EnableMfaForm user={this.getUser()} mfaType={this.state.mfaType} recoveryCodes={this.state.mfaProps.recoveryCodes}
        onSuccess={() => {
          Setting.showMessage("success", i18next.t("general:Enabled successfully"));
          if (this.state.isPromptPage && this.state.redirectUri) {
            Setting.goToLink(this.state.redirectUri);
          } else {
            Setting.goToLink("/account");
          }
        }}
        onFail={(res) => {
          Setting.showMessage("error", `${i18next.t("general:Failed to enable")}: ${res.msg}`);
        }} />;
    default:
      return null;
    }
  }

  render() {
    if (!this.props.account) {
      return (
        <Result
          status="403"
          title="403 Unauthorized"
          subTitle={i18next.t("general:Sorry, you do not have permission to access this page or logged in status invalid.")}
          extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>}
        />
      );
    }

    return (
      <Row>
        <Col span={24} style={{justifyContent: "center"}}>
          <Row>
            <Col span={24}>
              <div style={{textAlign: "center", fontSize: "28px"}}>
                {i18next.t("mfa:Protect your account with Multi-factor authentication")}</div>
              <div style={{textAlign: "center", fontSize: "16px", marginTop: "10px"}}>{i18next.t("mfa:Each time you sign in to your Account, you'll need your password and a authentication code")}</div>
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <Steps current={this.state.current}
                items={[
                  {title: i18next.t("mfa:Verify Password"), icon: <UserOutlined />},
                  {title: i18next.t("mfa:Verify Code"), icon: <KeyOutlined />},
                  {title: i18next.t("general:Enable"), icon: <CheckOutlined />},
                ]}
                style={{width: "90%", maxWidth: "500px", margin: "auto", marginTop: "80px",
                }} >
              </Steps>
            </Col>
          </Row>
        </Col>
        <Col span={24} style={{display: "flex", justifyContent: "center"}}>
          <div style={{marginTop: "10px", textAlign: "center"}}>
            {this.renderStep()}
          </div>
        </Col>
      </Row>
    );
  }
}

export default MfaSetupPage;
