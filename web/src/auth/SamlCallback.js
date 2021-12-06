// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import * as Setting from "../Setting";
import i18next from "i18next";

class SamlCallback extends React.Component {
    constructor(props) {
      super(props);
      this.state = {
        classes: props,
        msg: null,
      };
    }

    UNSAFE_componentWillMount() {
        const params = new URLSearchParams(this.props.location.search);
        let relayState = params.get('relayState')
        let samlResponse = params.get('samlResponse')
        let redirectUri = `${window.location.origin}/callback`;
        const applicationName = "app-built-in"
        const body = {
            type: "login",
            application: applicationName,
            provider: "aliyun-idaas",
            state: applicationName,
            redirectUri: redirectUri,
            method: "signup",
            relayState: relayState,
            samlResponse: encodeURIComponent(samlResponse),
          };
        AuthBackend.loginWithSaml(body)
          .then((res) => {
            if (res.status === 'ok') {
                  Util.showMessage("success", `Logged in successfully`);
                  // Setting.goToLinkSoft(this, "/");
                  Setting.goToLink("/");
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
export default withRouter(SamlCallback);