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
import {withRouter} from "react-router-dom";
import * as Setting from "../Setting";
import * as AccountBackend from "../backend/AccountBackend";

class CallbackBox extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      providerType: props.match.params.providerType,
      providerName: props.match.params.providerName,
      addition: props.match.params.addition,
      state: "",
      code: "",
      isAuthenticated: false,
      isSignedUp: false,
      email: ""
    };
    const params = new URLSearchParams(this.props.location.search);
    this.state.code = params.get("code");
    this.state.state = params.get("state");
  }

  getAuthenticatedInfo() {
    let redirectUrl;
    redirectUrl = `${Setting.ClientUrl}/callback/${this.state.providerType}/${this.state.providerName}/${this.state.addition}`;
    switch (this.state.providerType) {
      case "github":
        AccountBackend.githubLogin(this.state.providerName, this.state.code, this.state.state, redirectUrl, this.state.addition)
          .then((res) => {
            if (res.status === "ok") {
              window.location.href = '/';
            }else {
              Setting.showMessage("error", res?.msg);
            }
          });
        break;
    }
  }

  componentDidMount() {
    this.getAuthenticatedInfo();
  }

  render() {
    return (
      <div>
        <h3>
          Logging in ...
        </h3>
      </div>
    )
  }
}

export default withRouter(CallbackBox);
