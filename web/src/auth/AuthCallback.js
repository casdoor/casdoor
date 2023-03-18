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
import RedirectForm from "../common/RedirectForm";

class AuthCallback extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      msg: null,
      samlResponse: "",
      relayState: "",
      redirectUrl: "",
    };
  }

  getInnerParams() {
    // For example, for Casbin-OA, realRedirectUri = "http://localhost:9000/login"
    // realRedirectUrl = "http://localhost:9000"
    const params = new URLSearchParams(this.props.location.search);
    const state = params.get("state");
    const queryString = Util.getQueryParamsFromState(state);
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
        const samlRequest = innerParams.get("SAMLRequest");
        // cas don't use 'redirect_url', it is called 'service'
        const casService = innerParams.get("service");
        if (samlRequest !== null && samlRequest !== undefined && samlRequest !== "") {
          return "saml";
        } else if (casService !== null && casService !== undefined && casService !== "") {
          return "cas";
        }
        return "login";
      }

      const realRedirectUrl = new URL(realRedirectUri).origin;

      // For Casdoor itself, we use "login" directly
      if (authServerUrl === realRedirectUrl) {
        return "login";
      } else {
        const responseType = innerParams.get("response_type");
        if (responseType !== null) {
          return responseType;
        }
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
    const isSteam = params.get("openid.mode");
    let code = params.get("code");
    // WeCom returns "auth_code=xxx" instead of "code=xxx"
    if (code === null) {
      code = params.get("auth_code");
    }
    // Dingtalk now  returns "authCode=xxx" instead of "code=xxx"
    if (code === null) {
      code = params.get("authCode");
    }
    // Steam don't use code, so we should use all params as code.
    if (isSteam !== null && code === null) {
      code = this.props.location.search;
    }

    const innerParams = this.getInnerParams();
    const applicationName = innerParams.get("application");
    const providerName = innerParams.get("provider");
    const method = innerParams.get("method");
    const samlRequest = innerParams.get("SAMLRequest");
    const casService = innerParams.get("service");

    const redirectUri = `${window.location.origin}/callback`;

    const body = {
      type: this.getResponseType(),
      application: applicationName,
      provider: providerName,
      code: code,
      samlRequest: samlRequest,
      // state: innerParams.get("state"),
      state: applicationName,
      redirectUri: redirectUri,
      method: method,
    };

    if (this.getResponseType() === "cas") {
      // user is using casdoor as cas sso server, and wants the ticket to be acquired
      AuthBackend.loginCas(body, {"service": casService}).then((res) => {
        if (res.status === "ok") {
          let msg = "Logged in successfully.";
          if (casService === "") {
            // If service was not specified, Casdoor must display a message notifying the client that it has successfully initiated a single sign-on session.
            msg += "Now you can visit apps protected by Casdoor.";
          }
          Setting.showMessage("success", msg);

          if (casService !== "") {
            const st = res.data;
            const newUrl = new URL(casService);
            newUrl.searchParams.append("ticket", st);
            window.location.href = newUrl.toString();
          }
        } else {
          Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
        }
      });
      return;
    }
    // OAuth
    const oAuthParams = Util.getOAuthGetParameters(innerParams);
    const concatChar = oAuthParams?.redirectUri?.includes("?") ? "&" : "?";
    AuthBackend.login(body, oAuthParams)
      .then((res) => {
        if (res.status === "ok") {
          const responseType = this.getResponseType();
          if (responseType === "login") {
            Setting.showMessage("success", "Logged in successfully");
            // Setting.goToLinkSoft(this, "/");

            const link = Setting.getFromLink();
            Setting.goToLink(link);
          } else if (responseType === "code") {
            const code = res.data;
            Setting.goToLink(`${oAuthParams.redirectUri}${concatChar}code=${code}&state=${oAuthParams.state}`);
            // Setting.showMessage("success", `Authorization code: ${res.data}`);
          } else if (responseType === "token" || responseType === "id_token") {
            const token = res.data;
            Setting.goToLink(`${oAuthParams.redirectUri}${concatChar}${responseType}=${token}&state=${oAuthParams.state}&token_type=bearer`);
          } else if (responseType === "link") {
            const from = innerParams.get("from");
            Setting.goToLinkSoft(this, from);
          } else if (responseType === "saml") {
            if (res.data2.method === "POST") {
              this.setState({
                samlResponse: res.data,
                redirectUrl: res.data2.redirectUrl,
                relayState: oAuthParams.relayState,
              });
            } else {
              const SAMLResponse = res.data;
              const redirectUri = res.data2.redirectUrl;
              Setting.goToLink(`${redirectUri}?SAMLResponse=${encodeURIComponent(SAMLResponse)}&RelayState=${oAuthParams.relayState}`);
            }
          }
        } else {
          this.setState({
            msg: res.msg,
          });
        }
      });
  }

  render() {
    if (this.state.samlResponse !== "") {
      return <RedirectForm samlResponse={this.state.samlResponse} redirectUrl={this.state.redirectUrl} relayState={this.state.relayState} />;
    }

    return (
      <div style={{display: "flex", justifyContent: "center", alignItems: "center"}}>
        {
          (this.state.msg === null) ? (
            <Spin size="large" tip={i18next.t("login:Signing in...")} style={{paddingTop: "10%"}} />
          ) : (
            Util.renderMessageLarge(this, this.state.msg)
          )
        }
      </div>
    );
  }
}

export default withRouter(AuthCallback);
