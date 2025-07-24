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

  componentWillUnmount() {
    this.clearPolling();
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
    const {application, loginWidth = 320} = this.props;
    const {loading, qrCode} = this.state;
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

export default WeChatLoginPanel;
