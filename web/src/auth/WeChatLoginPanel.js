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
      }, 30000);
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
    const loginPage = mode === "loginPage";
    return (
      <div style={{width: loginWidth}}>
        {(!loginPage) && (
          <>
            {application.signinItems?.filter(item => item.name === "Logo").map(signinItem => this.props.renderFormItem(application, signinItem))}
            {this.props.renderMethodChoiceBox()}
            {application.signinItems?.filter(item => item.name === "Languages").map(signinItem => this.props.renderFormItem(application, signinItem))}
          </>
        )}

        {loading ? (
          <div style={{width: loginWidth, height: 350, marginTop: loginPage ? 30 : 0, borderLeft: loginPage ? "1px solid #ccc" : "none", padding: loginPage ? "0 50 0 30" : "0", marginLeft: loginPage ? 40 : 0, display: "flex", justifyContent: "center", alignItems: "center"}}>
            <span>{i18next.t("login:Loading")}</span>
          </div>
        ) : qrCode ? (
          <>
            {loginPage ? (
              <div style={{width: loginWidth, height: 350, marginTop: 30, paddingBottom: 30, borderLeft: "1px solid #ccc", paddingLeft: 50, marginLeft: 40, display: "flex", justifyContent: "center", alignItems: "center", flexDirection: "column"}}>
                <span style={{fontSize: 13, marginBottom: 4}}>{i18next.t("login:Scan QR code with Wechat app to login")}</span>
                <img src={`data:image/png;base64,${qrCode}`} alt="WeChat QR code" style={{width: 190, height: 190, border: "1px solid #ccc"}} />
                <span style={{fontSize: 9, marginTop: 2}}>{i18next.t("login:QR code refreshes in 30s")}</span>
              </div>
            ) : (
              <div style={{display: "flex", flexDirection: "column", alignItems: "center", width: 320}}>
                <img src={`data:image/png;base64,${qrCode}`} alt="WeChat QR code" style={{width: 250, height: 250}} />
                <a style={{paddingTop: 10}} onClick={e => {e.preventDefault(); this.fetchQrCode();}}>
                  {i18next.t("login:Refresh")}
                </a>
              </div>
            )}
          </>
        ) : null}
      </div>
    );
  }
}

export default WeChatLoginPanel;
