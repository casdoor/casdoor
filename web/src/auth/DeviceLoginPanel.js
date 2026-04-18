// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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
import {Button, QRCode, Typography} from "antd";
import i18next from "i18next";
import * as AuthBackend from "./AuthBackend";
import * as Util from "./Util";

class DeviceLoginPanel extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      session: null,
      qrStatus: "loading",
      phase: "loading",
      error: "",
    };
    this.pollingTimer = null;
    this.refreshTimer = null;
  }

  componentDidMount() {
    this.startDeviceLogin();
  }

  componentDidUpdate(prevProps) {
    if (prevProps.application?.clientId !== this.props.application?.clientId) {
      this.startDeviceLogin();
    }
  }

  componentWillUnmount() {
    this.clearPolling();
    this.clearRefreshTimer();
  }

  clearPolling() {
    if (this.pollingTimer) {
      clearInterval(this.pollingTimer);
      this.pollingTimer = null;
    }
  }

  clearRefreshTimer() {
    if (this.refreshTimer) {
      clearTimeout(this.refreshTimer);
      this.refreshTimer = null;
    }
  }

  startDeviceLogin() {
    const {application} = this.props;
    if (!application?.clientId) {
      return;
    }

    this.clearPolling();
    this.clearRefreshTimer();
    const oAuthParams = Util.getOAuthGetParameters();
    const scope = oAuthParams?.scope || "openid profile email";

    this.setState({session: null, qrStatus: "loading", phase: "loading", error: ""});
    AuthBackend.startDeviceLogin(application.clientId, scope).then((res) => {
      if (res.error || !res.device_code || !res.verification_uri) {
        this.setState({session: null, qrStatus: "expired", phase: "error", error: res.error_description || res.error || i18next.t("login:Device login is unavailable")});
        return;
      }

      this.setState({session: res, qrStatus: "active", phase: "pending", error: ""});
      const intervalSeconds = res.interval || 5;
      this.pollingTimer = setInterval(() => {
        this.pollDeviceToken();
      }, intervalSeconds * 1000);
    }).catch((error) => {
      this.setState({session: null, qrStatus: "expired", phase: "error", error: error.message});
    });
  }

  pollDeviceToken() {
    const {application, onSuccess} = this.props;
    const {session, phase} = this.state;
    if (!session?.device_code || phase === "success" || phase === "expired" || phase === "denied") {
      return;
    }

    AuthBackend.pollDeviceLoginToken(application.clientId, session.device_code).then((res) => {
      if (res.access_token) {
        this.clearPolling();
        this.setState({phase: "success", qrStatus: "active", error: ""});
        if (onSuccess) {
          onSuccess(this.state.session.device_code);
        }
        return;
      }

      if (res.error === "authorization_pending") {
        return;
      }

      this.clearPolling();
      if (res.error === "access_denied") {
        this.setState({phase: "denied", qrStatus: "expired", error: res.error_description || i18next.t("login:Device login was canceled")});
      } else if (res.error === "expired_token") {
        this.setState({phase: "expired", qrStatus: "expired", error: res.error_description || i18next.t("login:Device login expired")});
        this.refreshTimer = setTimeout(() => {
          this.startDeviceLogin();
        }, 800);
      } else {
        this.setState({phase: "error", qrStatus: "expired", error: res.error_description || res.error || i18next.t("login:Device login is unavailable")});
      }
    }).catch((error) => {
      this.clearPolling();
      this.setState({phase: "error", qrStatus: "expired", error: error.message});
    });
  }

  renderDescription() {
    const {session, phase, error} = this.state;

    if (phase === "success") {
      return <Typography.Text>{i18next.t("application:Logged in successfully")}</Typography.Text>;
    }

    if (phase === "denied" || phase === "expired" || phase === "error") {
      return <Typography.Text type="danger">{error}</Typography.Text>;
    }

    return (
      <React.Fragment>
        <Typography.Text>{i18next.t("login:Scan this QR code with a signed-in device to continue")}</Typography.Text>
        <br />
        {
          session?.user_code ? (
            <Typography.Text type="secondary">
              {i18next.t("login:Confirmation code")}: {session.user_code}
            </Typography.Text>
          ) : null
        }
      </React.Fragment>
    );
  }

  render() {
    const {session, qrStatus} = this.state;

    return (
      <div style={{width: 320, margin: "0 auto", textAlign: "center"}}>
        <div style={{marginBottom: 12}}>
          <Typography.Title level={4} style={{marginBottom: 8}}>
            {i18next.t("login:Device login")}
          </Typography.Title>
          {this.renderDescription()}
        </div>
        <QRCode
          style={{margin: "auto", marginTop: "12px", marginBottom: "16px"}}
          bordered={false}
          status={qrStatus}
          value={session?.verification_uri ?? " "}
          size={230}
        />
        <Button type="link" onClick={() => this.startDeviceLogin()}>
          {i18next.t("general:Refresh")}
        </Button>
      </div>
    );
  }
}

export default DeviceLoginPanel;
