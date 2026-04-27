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
import {Avatar, Button, Typography} from "antd";
import i18next from "i18next";
import * as Setting from "../Setting";
import * as Util from "./Util";

const nativeSsoCandidatePorts = [47321, 47322, 47323, 47324, 47325];
const nativeSsoStatusPath = "/native-sso/status";
const nativeSsoAuthorizePath = "/native-sso/authorize";

class NativeSsoPanel extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      nativeSsoAgent: null,
      active: false,
      authorizing: false,
      error: "",
    };
    this.disposed = false;
  }

  componentDidMount() {
    this.startNativeSsoDiscovery();
  }

  componentDidUpdate(prevProps) {
    if (prevProps.application?.clientId !== this.props.application?.clientId || prevProps.restartKey !== this.props.restartKey) {
      this.startNativeSsoDiscovery();
    }
  }

  componentWillUnmount() {
    this.disposed = true;
  }

  setActive(active) {
    if (this.state.active !== active) {
      this.setState({active});
      this.props.onActiveChange?.(active);
    }
  }

  getNativeSsoBaseUrl(port) {
    return `http://127.0.0.1:${port}`;
  }

  async fetchNativeSsoJson(url, options = {}) {
    const timeoutMs = options.timeoutMs ?? 1200;
    const controller = timeoutMs > 0 ? new AbortController() : null;
    const timer = controller === null ? null : setTimeout(() => controller.abort(), timeoutMs);
    const headers = {
      ...(options.headers || {}),
    };
    if (options.body !== undefined) {
      headers["Content-Type"] = "application/json";
    }
    const {timeoutMs: _timeoutMs, ...requestOptions} = options;
    try {
      const response = await fetch(url, {
        ...requestOptions,
        signal: controller?.signal,
        headers: headers,
      });
      return await response.json();
    } finally {
      if (timer !== null) {
        clearTimeout(timer);
      }
    }
  }

  getNativeSsoRequestContext() {
    const oAuthParams = Util.getOAuthGetParameters();
    return {
      serverUrl: Setting.getFullServerUrl(),
      clientId: this.props.application?.clientId || oAuthParams?.clientId || "",
      applicationName: this.props.application?.name || "",
      organization: this.props.application?.organization || "",
      responseType: oAuthParams?.responseType || this.props.type || "login",
      redirectUri: oAuthParams?.redirectUri || "",
      scope: oAuthParams?.scope || "openid profile email device_sso",
      state: oAuthParams?.state || "",
      nonce: oAuthParams?.nonce || "",
      codeChallenge: oAuthParams?.codeChallenge || "",
      challengeMethod: oAuthParams?.challengeMethod || "",
      resource: oAuthParams?.resource || "",
    };
  }

  async startNativeSsoDiscovery() {
    if (!this.props.application?.clientId) {
      this.setActive(false);
      return;
    }

    this.setState({authorizing: false, error: ""});
    const initialAgent = this.props.initialAgent;
    const candidatePorts = initialAgent?.port
      ? [initialAgent.port, ...nativeSsoCandidatePorts.filter(port => port !== initialAgent.port)]
      : nativeSsoCandidatePorts;

    if (initialAgent?.port) {
      this.setState({nativeSsoAgent: initialAgent, active: true});
      this.props.onActiveChange?.(true);
    } else {
      this.setActive(false);
    }

    for (const port of candidatePorts) {
      const status = await this.getNativeSsoStatus(port);
      if (this.disposed) {
        return;
      }

      if (status?.available === true) {
        this.setState({
          nativeSsoAgent: {
            ...initialAgent,
            ...status,
            port: port,
          },
          error: "",
        });
        this.setActive(true);
        return;
      }
    }

    this.setState({nativeSsoAgent: null});
    this.setActive(false);
  }

  async getNativeSsoStatus(port) {
    try {
      const query = new URLSearchParams({
        serverUrl: Setting.getFullServerUrl(),
        clientId: this.props.application?.clientId || "",
      });
      const status = await this.fetchNativeSsoJson(`${this.getNativeSsoBaseUrl(port)}${nativeSsoStatusPath}?${query.toString()}`, {
        method: "GET",
      });
      if (!this.isNativeSsoStatusValid(status)) {
        return null;
      }
      return status;
    } catch {
      return null;
    }
  }

  isNativeSsoStatusValid(status) {
    return status?.available === true &&
      String(status.serverUrl || "").replace(/\/+$/, "") === Setting.getFullServerUrl().replace(/\/+$/, "");
  }

  async authorizeNativeSso() {
    const {nativeSsoAgent, authorizing} = this.state;
    if (!nativeSsoAgent?.port || authorizing) {
      return;
    }

    this.setState({authorizing: true, error: ""});
    try {
      const result = await this.fetchNativeSsoJson(`${this.getNativeSsoBaseUrl(nativeSsoAgent.port)}${nativeSsoAuthorizePath}`, {
        method: "POST",
        body: JSON.stringify(this.getNativeSsoRequestContext()),
        timeoutMs: 120000,
      });

      if (result?.status !== "approved") {
        this.fallbackToPasswordLogin(result?.message || result?.msg || i18next.t("login:Native SSO was denied"));
        return;
      }

      this.props.onSuccess?.(result);
    } catch (error) {
      this.fallbackToPasswordLogin(error.message || i18next.t("login:Native SSO was denied"));
    }
  }

  fallbackToPasswordLogin(message) {
    this.setState({authorizing: false, error: ""});
    this.setActive(false);
    this.props.onFallback?.(message, this.state.nativeSsoAgent);
  }

  useOtherLoginMethods() {
    this.setState({authorizing: false, error: ""});
    this.setActive(false);
    this.props.onFallback?.("", this.state.nativeSsoAgent);
  }

  render() {
    const {nativeSsoAgent, active, authorizing, error} = this.state;
    if (!active || !nativeSsoAgent) {
      return null;
    }

    const appName = nativeSsoAgent.applicationName || this.props.application?.displayName || this.props.application?.name || "";
    const agentName = nativeSsoAgent.displayName || nativeSsoAgent.userName || nativeSsoAgent.username || nativeSsoAgent.name || "";
    const agentAvatar = nativeSsoAgent.avatar || "";

    return (
      <div style={{width: 320, margin: "0 auto", textAlign: "center"}}>
        <div style={{marginBottom: 12}}>
          <Typography.Title level={4} style={{marginBottom: 8}}>
            {i18next.t("login:Native SSO")}
          </Typography.Title>
          <Typography.Text>{i18next.t("login:Signed in on this device with {app}").replace("{app}", appName)}</Typography.Text>
        </div>
        <div style={{border: "1px solid #f0f0f0", borderRadius: 8, padding: 24, background: "#fafafa"}}>
          <Avatar src={agentAvatar || undefined} size={72} style={{marginBottom: 12}}>
            {agentName ? agentName.substring(0, 1).toUpperCase() : null}
          </Avatar>
          {agentName ? (
            <div style={{fontSize: 18, fontWeight: 600, marginBottom: 8}}>
              {agentName}
            </div>
          ) : null}
          <Typography.Text type="secondary">
            {i18next.t("login:Native SSO is ready")}
          </Typography.Text>
          <div style={{marginTop: 20}}>
            <Button type="primary" size="large" block onClick={() => this.authorizeNativeSso()} disabled={authorizing}>
              {i18next.t("login:Native SSO")}
            </Button>
          </div>
          <div style={{marginTop: 12}}>
            <Button type="link" onClick={() => this.useOtherLoginMethods()}>
              {i18next.t("login:Use other login methods")}
            </Button>
          </div>
        </div>
        {error === "" ? null : (
          <div style={{marginTop: 12}}>
            <Typography.Text type="danger">{error}</Typography.Text>
          </div>
        )}
      </div>
    );
  }
}

export default NativeSsoPanel;
