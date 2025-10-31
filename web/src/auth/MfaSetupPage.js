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

import React from "react";
import {Button, Col, Result, Row, Spin, Steps} from "antd";
import {withRouter} from "react-router-dom";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as MfaBackend from "../backend/MfaBackend";
import {CheckOutlined, KeyOutlined, UserOutlined} from "@ant-design/icons";
import CheckPasswordForm from "./mfa/CheckPasswordForm";
import MfaEnableForm from "./mfa/MfaEnableForm";
import {MfaVerifyForm} from "./mfa/MfaVerifyForm";

export const EmailMfaType = "email";
export const SmsMfaType = "sms";
export const TotpMfaType = "app";
export const RadiusMfaType = "radius";
export const PushMfaType = "push";
export const RecoveryMfaType = "recovery";

class MfaSetupPage extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(props.location.search);
    const {location} = this.props;
    this.state = {
      account: props.account,
      application: null,
      applicationName: props.account.signupApplication ?? localStorage.getItem("applicationName") ?? "",
      current: location.state?.from !== undefined ? 1 : 0,
      mfaProps: null,
      mfaType: params.get("mfaType") ?? SmsMfaType,
      isPromptPage: props.isPromptPage || location.state?.from !== undefined,
      loading: false,
    };
  }

  componentDidMount() {
    this.getApplication();
    if (this.state.current === 1) {
      this.setState({
        loading: true,
      });

      setTimeout(() => {
        this.initMfaProps();
      }, 200);
    }
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    if (this.state.mfaType !== prevState.mfaType || this.state.current !== prevState.current) {
      if (this.state.current === 1) {
        this.initMfaProps();
      }
    }
  }

  getApplication() {
    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((res) => {
        if (res !== null) {
          if (res.status === "error") {
            Setting.showMessage("error", res.msg);
            return;
          }
          this.setState({
            application: res.data,
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
          loading: false,
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
      if (this.state.mfaType === SmsMfaType) {
        return null;
      }
      return (<Button type={"link"} onClick={() => {
        this.setState({
          mfaType: SmsMfaType,
        });
        this.props.history.push(`/mfa/setup?mfaType=${SmsMfaType}`);
      }
      }>{i18next.t("mfa:Use SMS")}</Button>
      );
    };

    const renderEmailLink = () => {
      if (this.state.mfaType === EmailMfaType) {
        return null;
      }
      return (<Button type={"link"} onClick={() => {
        this.setState({
          mfaType: EmailMfaType,
        });
        this.props.history.push(`/mfa/setup?mfaType=${EmailMfaType}`);
      }
      }>{i18next.t("mfa:Use Email")}</Button>
      );
    };

    const renderTotpLink = () => {
      if (this.state.mfaType === TotpMfaType) {
        return null;
      }
      return (<Button type={"link"} onClick={() => {
        this.setState({
          mfaType: TotpMfaType,
        });
        this.props.history.push(`/mfa/setup?mfaType=${TotpMfaType}`);
      }
      }>{i18next.t("mfa:Use Authenticator App")}</Button>
      );
    };

    const renderRadiusLink = () => {
      if (this.state.mfaType === RadiusMfaType) {
        return null;
      }
      return (<Button type={"link"} onClick={() => {
        this.setState({
          mfaType: RadiusMfaType,
        });
        this.props.history.push(`/mfa/setup?mfaType=${RadiusMfaType}`);
      }
      }>{i18next.t("mfa:Use Radius")}</Button>
      );
    };

    const renderPushLink = () => {
      if (this.state.mfaType === PushMfaType) {
        return null;
      }
      return (<Button type={"link"} onClick={() => {
        this.setState({
          mfaType: PushMfaType,
        });
        this.props.history.push(`/mfa/setup?mfaType=${PushMfaType}`);
      }
      }>{i18next.t("mfa:Use Push Notification")}</Button>
      );
    };

    return !this.state.isPromptPage ? (
      <React.Fragment>
        {renderSmsLink()}
        {renderEmailLink()}
        {renderTotpLink()}
        {renderRadiusLink()}
        {renderPushLink()}
      </React.Fragment>
    ) : null;
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
            });
          }}
          onFail={(res) => {
            Setting.showMessage("error", i18next.t("mfa:Failed to initiate MFA") + ": " + res.msg);
          }}
        />
      );
    case 1:
      return (
        <div>
          <MfaVerifyForm
            mfaProps={this.state.mfaProps}
            application={this.state.application}
            user={this.props.account}
            onSuccess={(res) => {
              this.setState({
                dest: res.dest,
                countryCode: res.countryCode,
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
      return (
        <MfaEnableForm user={this.getUser()} mfaType={this.state.mfaType} secret={this.state.mfaProps.secret} recoveryCodes={this.state.mfaProps.recoveryCodes} dest={this.state.dest} countryCode={this.state.countryCode}
          onSuccess={() => {
            Setting.showMessage("success", i18next.t("general:Enabled successfully"));
            this.props.onfinish();

            const mfaRedirectUrl = localStorage.getItem("mfaRedirectUrl");
            if (mfaRedirectUrl !== undefined && mfaRedirectUrl !== null) {
              Setting.goToLink(localStorage.getItem("mfaRedirectUrl"));
              localStorage.removeItem("mfaRedirectUrl");
            } else {
              this.props.history.push("/account");
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
          extra={<a href="/web/public"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>}
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
          <Spin spinning={this.state.loading}>
            <Steps current={this.state.current}
              items={[
                {title: i18next.t("mfa:Verify Password"), icon: <UserOutlined />},
                {title: i18next.t("mfa:Verify Code"), icon: <KeyOutlined />},
                {title: i18next.t("general:Enable"), icon: <CheckOutlined />},
              ]}
              style={{width: "90%", maxWidth: "500px", margin: "auto", marginTop: "50px",
              }} >
            </Steps>
          </Spin>
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

export default withRouter(MfaSetupPage);
