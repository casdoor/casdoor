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

import React, {useState} from "react";
import {Button, Col, Form, Input, Result, Row, Steps} from "antd";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as MfaBackend from "../backend/MfaBackend";
import {CheckOutlined, KeyOutlined, LockOutlined, UserOutlined} from "@ant-design/icons";

import * as UserBackend from "../backend/UserBackend";
import {MfaSmsVerifyForm, MfaTotpVerifyForm, mfaSetup} from "./MfaVerifyForm";

export const EmailMfaType = "email";
export const SmsMfaType = "sms";
export const TotpMfaType = "app";
export const RecoveryMfaType = "recovery";

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

export function MfaVerifyForm({mfaProps, application, user, onSuccess, onFail}) {
  const [form] = Form.useForm();
  const onFinish = ({passcode}) => {
    const data = {passcode, mfaType: mfaProps.mfaType, ...user};
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

  if (mfaProps === undefined || mfaProps === null || application === undefined || application === null || user === undefined || user === null) {
    return <div></div>;
  }

  if (mfaProps.mfaType === SmsMfaType || mfaProps.mfaType === EmailMfaType) {
    return <MfaSmsVerifyForm mfaProps={mfaProps} onFinish={onFinish} application={application} method={mfaSetup} user={user} />;
  } else if (mfaProps.mfaType === TotpMfaType) {
    return <MfaTotpVerifyForm mfaProps={mfaProps} onFinish={onFinish} />;
  } else {
    return <div></div>;
  }
}

function EnableMfaForm({user, mfaType, recoveryCodes, onSuccess, onFail}) {
  const [loading, setLoading] = useState(false);
  const requestEnableMfa = () => {
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
        requestEnableMfa();
      }} block type="primary">
        {i18next.t("general:Enable")}
      </Button>
    </div>
  );
}

class MfaSetupPage extends React.Component {
  constructor(props) {
    super(props);
    const requiredMfaTypes = Setting.getRequiredMfaTypes(props.account.organization);
    this.state = {
      account: props.account,
      application: this.props.application ?? null,
      applicationName: props.account.signupApplication ?? "",
      isAuthenticated: props.isAuthenticated ?? false,
      isPromptPage: props.isPromptPage,
      redirectUri: props.redirectUri,
      current: props.current ?? 0,
      mfaType: requiredMfaTypes?.[0] ?? new URLSearchParams(props.location?.search)?.get("mfaType") ?? SmsMfaType,
      mfaProps: null,
      requiredMfaTypes: requiredMfaTypes,
    };
    // eslint-disable-next-line no-console
    console.log(this.props);
  }

  componentDidMount() {
    this.getApplication();
    // eslint-disable-next-line no-console
    console.log(this.state);
    if (this.state.isPromptPage === true) {
      this.initMfaProps();
    }
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    if (this.state.isAuthenticated === true && (this.state.mfaProps === null || this.state.mfaType !== prevState.mfaType)) {
      this.initMfaProps();
    }
  }

  getApplication() {
    if (this.state.application !== null) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((res) => {
        if (res !== null) {
          if (res.status === "error") {
            Setting.showMessage("error", res.msg);
            return;
          }
          this.setState({
            application: res,
          });
        } else {
          Setting.showMessage("error", i18next.t("mfa:Failed to get application"));
        }
      });
  }

  initMfaProps() {
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

  getUser() {
    return this.props.account;
  }

  renderMfaTypeSwitch() {
    const renderSmsLink = () => {
      if ((this.state.isPromptPage && !this.state.requiredMfaTypes.includes(SmsMfaType)) || this.state.mfaType === SmsMfaType || this.props.account.mfaPhoneEnabled) {
        return null;
      }
      return (<Button type={"link"} onClick={() => {
        this.setState({
          mfaType: SmsMfaType,
        });
      }
      }>{i18next.t("mfa:Use SMS")}</Button>
      );
    };

    const renderEmailLink = () => {
      if ((this.state.isPromptPage && !this.state.requiredMfaTypes.includes(EmailMfaType)) || this.state.mfaType === EmailMfaType || this.props.account.mfaEmailEnabled) {
        return null;
      }
      return (<Button type={"link"} onClick={() => {
        this.setState({
          mfaType: EmailMfaType,
        });
      }
      }>{i18next.t("mfa:Use Email")}</Button>
      );
    };

    const renderTotpLink = () => {
      if ((this.state.isPromptPage && !this.state.requiredMfaTypes.includes(TotpMfaType)) || this.state.mfaType === TotpMfaType || this.props.account.totpSecret !== "") {
        return null;
      }
      return (<Button type={"link"} onClick={() => {
        this.setState({
          mfaType: TotpMfaType,
        });
      }
      }>{i18next.t("mfa:Use Authenticator App")}</Button>
      );
    };

    return (
      <React.Fragment>
        {renderSmsLink()}
        {renderEmailLink()}
        {renderTotpLink()}
      </React.Fragment>
    );
  }

  renderStep() {
    switch (this.state.current) {
    case 0:
      return (
        <CheckPasswordForm
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
        />
      );
    case 1:
      if (!this.state.isAuthenticated) {
        return null;
      }

      return (
        <div>
          <MfaVerifyForm
            mfaProps={this.state.mfaProps}
            application={this.state.application}
            user={this.props.account}
            onSuccess={() => {
              this.setState({
                current: this.state.current + 1,
              });
            }}
            onFail={(res) => {
              Setting.showMessage("error", i18next.t("general:Failed to verify") + ": " + res.msg);
            }}
          />
          <Col span={24} style={{display: "flex", justifyContent: "left"}}>
            {this.renderMfaTypeSwitch()}
          </Col>
        </div>
      );
    case 2:
      if (!this.state.isAuthenticated) {
        return null;
      }

      return (
        <EnableMfaForm user={this.getUser()} mfaType={this.state.mfaType} recoveryCodes={this.state.mfaProps.recoveryCodes}
          onSuccess={() => {
            Setting.showMessage("success", i18next.t("general:Enabled successfully"));
            if (this.state.isPromptPage) {
              this.props.onfinish("mfa");
            } else {
              Setting.goToLink("/account");
            }
          }}
          onFail={(res) => {
            Setting.showMessage("error", `${i18next.t("general:Failed to enable")}: ${res.msg}`);
          }} />
      );
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
              <p style={{textAlign: "center", fontSize: "28px"}}>
                {i18next.t("mfa:Protect your account with Multi-factor authentication")}</p>
              <p style={{textAlign: "center", fontSize: "16px", marginTop: "10px"}}>{i18next.t("mfa:Each time you sign in to your Account, you'll need your password and a authentication code")}</p>
            </Col>
          </Row>
          <Steps current={this.state.current}
            items={[
              {title: i18next.t("mfa:Verify Password"), icon: <UserOutlined />},
              {title: i18next.t("mfa:Verify Code"), icon: <KeyOutlined />},
              {title: i18next.t("general:Enable"), icon: <CheckOutlined />},
            ]}
            style={{width: "90%", maxWidth: "500px", margin: "auto", marginTop: "50px",
            }} >
          </Steps>
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
