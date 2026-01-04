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
import {Card} from "antd";
import {withRouter} from "react-router-dom";
import * as Util from "./Util";
import * as Setting from "../Setting";
import i18next from "i18next";

class TelegramLogin extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      applicationName: "",
      providerName: "",
      botUsername: "",
      authUrl: "",
    };
  }

  UNSAFE_componentWillMount() {
    const params = new URLSearchParams(this.props.location.search);
    const state = params.get("state");
    const queryString = Util.getQueryParamsFromState(state);
    const innerParams = new URLSearchParams(queryString);
    
    const applicationName = innerParams.get("application");
    const providerName = innerParams.get("provider");
    
    // Get provider info to retrieve bot username
    Setting.getProvider("admin", providerName).then((provider) => {
      if (provider) {
        const redirectOrigin = window.location.origin;
        const redirectUri = `${redirectOrigin}/callback`;
        
        this.setState({
          applicationName: applicationName,
          providerName: providerName,
          botUsername: provider.clientId,
          authUrl: `${redirectUri}?state=${state}`,
        }, () => {
          this.loadTelegramWidget();
        });
      }
    });
  }

  loadTelegramWidget() {
    if (!this.state.botUsername || !this.state.authUrl) {
      return;
    }

    // Remove any existing Telegram script
    const existingScript = document.querySelector('script[src*="telegram-widget"]');
    if (existingScript) {
      existingScript.remove();
    }

    // Create and load the Telegram widget script
    const script = document.createElement("script");
    script.src = "https://telegram.org/js/telegram-widget.js?22";
    script.setAttribute("data-telegram-login", this.state.botUsername);
    script.setAttribute("data-size", "large");
    script.setAttribute("data-auth-url", this.state.authUrl);
    script.setAttribute("data-request-access", "write");
    script.async = true;

    const container = document.getElementById("telegram-login-container");
    if (container) {
      container.innerHTML = "";
      container.appendChild(script);
    }
  }

  render() {
    return (
      <div className="login-content" style={{margin: "auto"}}>
        <div style={{marginBottom: "10px", textAlign: "center"}}>
          <Card
            style={{
              width: "400px",
              margin: "0 auto",
              marginTop: "100px",
            }}
            title={
              <div>
                <img
                  width={40}
                  height={40}
                  src={Setting.getProviderLogoURL({type: "Telegram"})}
                  alt="Telegram"
                  style={{marginRight: "10px"}}
                />
                {i18next.t("login:Sign in with Telegram")}
              </div>
            }
          >
            <div style={{textAlign: "center", padding: "20px"}}>
              <p>{i18next.t("login:Click the button below to sign in with Telegram")}</p>
              <div
                id="telegram-login-container"
                style={{
                  display: "flex",
                  justifyContent: "center",
                  marginTop: "20px",
                }}
              />
            </div>
          </Card>
        </div>
      </div>
    );
  }
}

export default withRouter(TelegramLogin);
