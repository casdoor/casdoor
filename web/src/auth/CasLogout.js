// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
import {Card, Spin} from "antd";
import {withRouter} from "react-router-dom";
import * as AuthBackend from "./AuthBackend";
import * as Setting from "../Setting";
import i18next from "i18next";

class CasLogout extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      msg: null,
    };
    if (props.match?.params.casApplicationName !== undefined) {
      this.state.owner = props.match?.params.owner;
      this.state.applicationName = props.match?.params.casApplicationName;
    }
  }

  UNSAFE_componentWillMount() {
    const params = new URLSearchParams(this.props.location.search);

    const logoutInterval = 100;
    const logoutTimeOut = (s, redirectUri, linkSoftFlag) => {
      setTimeout(() => {
        AuthBackend.getAccount().then((accountRes) => {
          if (accountRes.status === "ok") {
            AuthBackend.logout().then((logoutRes) => {
              if (logoutRes.status === "ok") {
                const redirectData = logoutRes.data2;
                if (redirectData !== null && redirectData !== undefined && redirectData !== "") {
                  logoutTimeOut(s, redirectData, false);
                } else if (params.has("service")) {
                  logoutTimeOut(s, params.get("service"), false);
                } else {
                  logoutTimeOut(s, `/cas/${this.state.owner}/${this.state.applicationName}/login`, true);
                }
              } else {
                Setting.showMessage("error", `Failed to log out: ${logoutRes.msg}`);
              }
            });
          } else {
            if (linkSoftFlag) {
              Setting.goToLinkSoft(this, redirectUri);
            } else {
              Setting.goToLink(redirectUri);
            }
            Setting.showMessage("success", "Logged out successfully");
            this.props.onUpdateAccount(null);
          }
        });
      }, s);
    };

    AuthBackend.logout()
      .then((res) => {
        if (res.status === "ok") {
          const redirectUri = res.data2;
          if (redirectUri !== null && redirectUri !== undefined && redirectUri !== "") {
            logoutTimeOut(logoutInterval, redirectUri, false);
          } else if (params.has("service")) {
            logoutTimeOut(logoutInterval, params.get("service"), false);
          } else {
            logoutTimeOut(logoutInterval, `/cas/${this.state.owner}/${this.state.applicationName}/login`, true);
          }
        } else {
          Setting.showMessage("error", `Failed to log out: ${res.msg}`);
        }
      });
  }

  render() {
    return (
      <Card>
        <div style={{display: "flex", justifyContent: "center", alignItems: "center"}}>
          {
            <Spin size="large" tip={i18next.t("login:Logging out...")} style={{paddingTop: "10%"}} />
          }
        </div>
      </Card>
    );
  }
}
export default withRouter(CasLogout);
