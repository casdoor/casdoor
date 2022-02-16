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
import {Spin} from "antd";
import {withRouter} from "react-router-dom";
import * as AuthBackend from "./AuthBackend";
import * as Util from "./Util";
import {authConfig} from "./Auth";
import * as Setting from "../Setting";
import i18next from "i18next";

class AuthCallback extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      msg: null,
    };
  }

  getInnerParams() {
    // For example, for Casbin-OA, realRedirectUri = "http://localhost:9000/login"
    // realRedirectUrl = "http://localhost:9000"
    const params = new URLSearchParams(this.props.location.search);
    const state = params.get("state");
    const queryString = Util.stateToGetQueryParams(state);
    return new URLSearchParams(queryString);
  }

  getResponseType() {
    // "http://localhost:8000"
    const authServerUrl = authConfig.serverUrl;

    const innerParams = this.getInnerParams();
    const method = innerParams.get("method");
    if (method === "signup") {
      const realRedirectUri = innerParams.get("redirect_uri");
      // Casdoor's own login page, so "code" is not necessary
      if (realRedirectUri === null) {
        return "login";
      }

      const realRedirectUrl = new URL(realRedirectUri).origin;

      // For Casdoor itself, we use "login" directly
      if (authServerUrl === realRedirectUrl) {
        return "login";
      } else {
        return "code";
      }
    } else if (method === "link") {
      return "link";
    } else {
      return "unknown";
    }
  }

  UNSAFE_componentWillMount() {
    const params = new URLSearchParams(this.props.location.search);
    let isSteam = params.get("openid.mode")
    let code = params.get("code");
    // WeCom returns "auth_code=xxx" instead of "code=xxx"
    if (code === null) {
      code = params.get("auth_code");
    }
    // Dingtalk now  returns "authCode=xxx" instead of "code=xxx"
    if (code === null) {
      code = params.get("authCode")
    }
    //Steam don't use code, so we should use all params as code.
    if (isSteam !== null && code === null) {
      code = this.props.location.search
    }

    const innerParams = this.getInnerParams();
    const applicationName = innerParams.get("application");
    const providerName = innerParams.get("provider");
    const method = innerParams.get("method");

    let redirectUri = `${window.location.origin}/callback`;

    const body = {
      type: this.getResponseType(),
      application: applicationName,
      provider: providerName,
      code: code,
      // state: innerParams.get("state"),
      state: applicationName,
      redirectUri: redirectUri,
      method: method,
    };
    const oAuthParams = Util.getOAuthGetParameters(innerParams);
    AuthBackend.login(body, oAuthParams)
      .then((res) => {
        if (res.status === 'ok') {
          const responseType = this.getResponseType();
          if (responseType === "login") {
            Util.showMessage("success", `Logged in successfully`);
            // Setting.goToLinkSoft(this, "/");

            const link = Setting.getFromLink();
            Setting.goToLink(link);
          } else if (responseType === "code") {
            const code = res.data;
            Setting.goToLink(`${oAuthParams.redirectUri}?code=${code}&state=${oAuthParams.state}`);
            // Util.showMessage("success", `Authorization code: ${res.data}`);
          } else if (responseType === "link") {
            const from = innerParams.get("from");
            Setting.goToLinkSoft(this, from);
          }
        } else {
          this.setState({
            msg: res.msg,
          });
        }
      });
  }

  render() {
    return (
      <div style={{textAlign: "center"}}>
        {
          (this.state.msg === null) ? (
            <Spin size="large" tip={i18next.t("login:Signing in...")} style={{paddingTop: "10%"}} />
          ) : (
            Util.renderMessageLarge(this, this.state.msg)
          )
        }
      </div>
    )
  }
}

export default withRouter(AuthCallback);
