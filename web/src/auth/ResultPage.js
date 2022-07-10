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
import {Result, Button} from "antd";
import i18next from "i18next";
import {authConfig} from "./Auth";
import * as Util from "./Util";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as Setting from "../Setting";

class ResultPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      applicationName: props.match.params.applicationName !== undefined ? props.match.params.applicationName : authConfig.appName,
      application: null,
    };
  }

  UNSAFE_componentWillMount() {
    if (this.state.applicationName !== undefined) {
      this.getApplication();
    } else {
      Util.showMessage("error", `Unknown application name: ${this.state.applicationName}`);
    }
  }

  getApplication() {
    if (this.state.applicationName === undefined) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });
      });
  }

  render() {
    const application = this.state.application;

    return (
      <div>
        {
          Setting.renderHelmet(application)
        }
        <Result
          status="success"
          title={i18next.t("signup:Your account has been created!")}
          subTitle={i18next.t("signup:Please click the below button to sign in")}
          extra={[
            <Button type="primary" key="login" onClick={() => {
              let linkInStorage = sessionStorage.getItem("signinUrl");
              if (linkInStorage !== null && linkInStorage !== "") {
                Setting.goToLink(linkInStorage);
              } else {
                Setting.goToLogin(this, application);
              }
            }}>
              {i18next.t("login:Sign In")}
            </Button>
          ]}
        />
      </div>
    );
  }
}

export default ResultPage;
