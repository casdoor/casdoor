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
import {message, Spin} from "antd";
import {withRouter} from "react-router-dom";
import * as AuthBackend from "./AuthBackend";
import {getClientUrl} from "./Auth";

class AuthCallback extends React.Component {
  constructor(props) {
    super(props);
    const params = new URLSearchParams(this.props.location.search);
    this.state = {
      classes: props,
      providerType: props.match.params.providerType,
      providerName: props.match.params.providerName,
      addition: props.match.params.addition,
      state: params.get("state"),
      code: params.get("code"),
      isAuthenticated: false,
      isSignedUp: false,
      email: ""
    };
  }

  componentWillMount() {
    this.authLogin();
  }

  showMessage(type, text) {
    if (type === "success") {
      message.success(text);
    } else if (type === "error") {
      message.error(text);
    }
  }

  authLogin() {
    let redirectUrl;
    redirectUrl = `${getClientUrl()}/callback/${this.state.providerType}/${this.state.providerName}/${this.state.addition}`;
    AuthBackend.authLogin(this.state.providerName, this.state.code, this.state.state, redirectUrl, this.state.addition)
      .then((res) => {
        if (res.status === "ok") {
          window.location.href = '/';
        } else {
          this.showMessage("error", res?.msg);
        }
      });
  }

  render() {
    return (
      <div style={{textAlign: "center"}}>
        <Spin size="large" tip="Signing in..." style={{paddingTop: "10%"}} />
      </div>
    )
  }
}

export default withRouter(AuthCallback);
