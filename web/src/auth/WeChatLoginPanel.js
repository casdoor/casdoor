// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
import * as AuthBackend from "./AuthBackend";
import i18next from "i18next";
import * as Util from "./Util";
import {Form} from "antd";

class WeChatLoginPanel extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      qrCode: null,
      loading: false,
      ticket: null,
    };
    this.pollingTimer = null;
  }

  UNSAFE_componentWillMount() {
    this.fetchQrCode();
  }

  componentDidUpdate(prevProps) {
    if (this.props.loginMethod === "wechat" && prevProps.loginMethod !== "wechat") {
      this.fetchQrCode();
    }
    if (prevProps.loginMethod === "wechat" && this.props.loginMethod !== "wechat") {
      this.setState({qrCode: null, loading: false, ticket: null});
      this.clearPolling();
    }
  }

  componentDidMount() {
    if (this.props.mode === "loginPage") {
      this.autoRefreshTimer = setInterval(() => {
        this.fetchQrCode();
      }, 300000);
    }
  }

  componentWillUnmount() {
    this.clearPolling();
    if (this.autoRefreshTimer) {
      clearInterval(this.autoRefreshTimer);
      this.autoRefreshTimer = null;
    }
  }

  clearPolling() {
    if (this.pollingTimer) {
      clearInterval(this.pollingTimer);
      this.pollingTimer = null;
    }
  }

  fetchQrCode() {
    const {application} = this.props;
    const wechatProviderItem = application?.providers?.find(p => p.provider?.type === "WeChat");
    if (wechatProviderItem) {
      this.setState({loading: true, qrCode: null, ticket: null});
      AuthBackend.getWechatQRCode(`${wechatProviderItem.provider.owner}/${wechatProviderItem.provider.name}`).then(res => {
        if (res.status === "ok" && res.data) {
          this.setState({qrCode: res.data, loading: false, ticket: res.data2});
          this.clearPolling();
          this.pollingTimer = setInterval(() => {
            Util.getEvent(application, wechatProviderItem.provider, res.data2, "signup");
          }, 1000);
        } else {
          this.setState({qrCode: null, loading: false, ticket: null});
          this.clearPolling();
        }
      }).catch(() => {
        this.setState({qrCode: null, loading: false, ticket: null});
        this.clearPolling();
      });
    }
  }

  render() {
    const {application, loginWidth = 320, mode} = this.props;
    const {loading, qrCode} = this.state;
    if (mode === "loginPage") {
      return (
        <div style={{width: loginWidth, margin: "0 auto", textAlign: "center", marginTop: 16}}>
          {loading ? (
            <div style={{marginTop: 120}}>
              <span style={{fontSize: 16}}>{i18next.t("login:Loading...")}</span>
            </div>
          ) : qrCode ? (
            <div style={{marginTop: 10, width: {loginWidth}, display: "flex", flexDirection: "column", alignItems: "center"}}>
              <span style={{fontSize: 14, marginBottom: 6, marginTop: 4, width: 200}}>{i18next.t("login:Scan QR code with Wechat app to login")}</span>
              <img src={`data:image/png;base64,${qrCode}`} alt="WeChat QR code" style={{width: 190, height: 190, border: "1px solid #ccc"}} />
              <span style={{fontSize: 10, marginTop: 5}}>{i18next.t("login:QR code refreshes in 30s")}</span>
            </div>
          ) : null}
        </div>
      );
    }
    return (
      <div style={{width: loginWidth, margin: "0 auto", textAlign: "center", marginTop: 16}}>
        {application.signinItems?.filter(item => item.name === "Logo").map(signinItem => this.props.renderFormItem(application, signinItem))}
        {this.props.renderMethodChoiceBox()}
        {application.signinItems?.filter(item => item.name === "Languages").map(signinItem => this.props.renderFormItem(application, signinItem))}
        {loading ? (
          <div style={{marginTop: 16}}>
            <span>{i18next.t("login:Loading")}</span>
          </div>
        ) : qrCode ? (
          <div style={{marginTop: 2}}>
            <img src={`data:image/png;base64,${qrCode}`} alt="WeChat QR code" style={{width: 250, height: 250}} />
            <div style={{marginTop: 8}}>
              <a onClick={e => {e.preventDefault(); this.fetchQrCode();}}>
                {i18next.t("login:Refresh")}
              </a>
            </div>
          </div>
        ) : null}
      </div>
    );
  }
}

export function WeChatLoginWithForm({application, loginWidth = 320, formProps = {}, children}) {
  return (
    <div style={{display: "flex", flexDirection: "row", justifyContent: "center", alignItems: "flex-start", background: "#fff", borderRadius: 8, boxShadow: "0 0 10px rgba(0,0,0,0.1)", width: 800, padding: "40px 60px", margin: "20px auto", position: "relative"}}>
      <Form
        name="normal_login"
        initialValues={{
          organization: application.organization,
          application: application.name,
          autoSignin: true,
          username: formProps.username ?? "",
          password: formProps.password ?? "",
        }}
        onFinish={formProps.onFinish}
        style={{width: `${loginWidth}px`, paddingTop: "20px", marginLeft: 70}}
        size="large"
        ref={formProps.formRef}
      >
        <Form.Item
          hidden={true}
          name="application"
          rules={[
            {
              required: true,
              message: formProps.appMsg || "Please input your application!",
            },
          ]}
        >
        </Form.Item>
        <Form.Item
          hidden={true}
          name="organization"
          rules={[
            {
              required: true,
              message: formProps.orgMsg || "Please input your organization!",
            },
          ]}
        >
        </Form.Item>
        {children}
      </Form>
      <div style={{marginTop: 70, marginLeft: 60, paddingTop: "20px", borderLeft: "1px solid #ccc", width: loginWidth, height: 340, display: "flex", flexDirection: "column", alignItems: "center"}}>
        <WeChatLoginPanel application={application} mode="loginPage" loginWidth={loginWidth} />
      </div>
    </div>
  );
}

export default WeChatLoginPanel;
