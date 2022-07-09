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
import {Button, Col, Result, Row} from "antd";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as UserBackend from "../backend/UserBackend";
import * as AuthBackend from "./AuthBackend";
import * as Setting from "../Setting";
import i18next from "i18next";
import AffiliationSelect from "../common/AffiliationSelect";
import OAuthWidget from "../common/OAuthWidget";

class PromptPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      type: props.type,
      applicationName: props.applicationName !== undefined ? props.applicationName : (props.match === undefined ? null : props.match.params.applicationName),
      application: null,
      user: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getUser();
    this.getApplication();
  }

  getUser() {
    const organizationName = this.props.account.owner;
    const userName = this.props.account.name;
    UserBackend.getUser(organizationName, userName)
      .then((user) => {
        this.setState({
          user: user,
        });
      });
  }

  getApplication() {
    if (this.state.applicationName === null) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((application) => {
        this.setState({
          application: application,
        });
      });
  }

  getApplicationObj() {
    if (this.props.application !== undefined) {
      return this.props.application;
    } else {
      return this.state.application;
    }
  }

  parseUserField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateUserField(key, value) {
    value = this.parseUserField(key, value);

    let user = this.state.user;
    user[key] = value;
    this.setState({
      user: user,
    });

    this.submitUserEdit(false);
  }

  renderAffiliation(application) {
    if (!Setting.isAffiliationPrompted(application)) {
      return null;
    }

    if (application === null || this.state.user === null) {
      return null;
    }

    return (
      <AffiliationSelect labelSpan={6} application={application} user={this.state.user} onUpdateUserField={(key, value) => {return this.updateUserField(key, value);}} />
    );
  }

  unlinked() {
    this.getUser();
  }

  renderContent(application) {
    return (
      <div style={{width: "400px"}}>
        {
          this.renderAffiliation(application)
        }
        <div>
          {
            (application === null || this.state.user === null) ? null : (
              application?.providers.filter(providerItem => Setting.isProviderPrompted(providerItem)).map((providerItem, index) => <OAuthWidget key={providerItem.name} labelSpan={6} user={this.state.user} application={application} providerItem={providerItem} onUnlinked={() => {return this.unlinked();}} />)
            )
          }
        </div>
      </div>
    );
  }

  onUpdateAccount(account) {
    this.props.onUpdateAccount(account);
  }

  getRedirectUrl() {
    // "/prompt/app-example?redirectUri=http://localhost:2000/callback&code=8eb113b072296818f090&state=app-example"
    const params = new URLSearchParams(this.props.location.search);
    const redirectUri = params.get("redirectUri");
    const code = params.get("code");
    const state = params.get("state");
    if (redirectUri === null || code === null || state === null) {
      return "";
    }
    return `${redirectUri}?code=${code}&state=${state}`;
  }

  logout() {
    AuthBackend.logout()
      .then((res) => {
        if (res.status === "ok") {
          this.onUpdateAccount(null);

          let redirectUrl = this.getRedirectUrl();
          if (redirectUrl === "") {
            redirectUrl = res.data2;
          }
          if (redirectUrl !== "") {
            Setting.goToLink(redirectUrl);
          } else {
            Setting.goToLogin(this, this.getApplicationObj());
          }
        } else {
          Setting.showMessage("error", `Failed to log out: ${res.msg}`);
        }
      });
  }

  submitUserEdit(isFinal) {
    let user = Setting.deepCopy(this.state.user);
    UserBackend.updateUser(this.state.user.owner, this.state.user.name, user)
      .then((res) => {
        if (res.msg === "") {
          if (isFinal) {
            Setting.showMessage("success", "Successfully saved");

            this.logout();
          }
        } else {
          if (isFinal) {
            Setting.showMessage("error", res.msg);
          }
        }
      })
      .catch(error => {
        if (isFinal) {
          Setting.showMessage("error", `Failed to connect to server: ${error}`);
        }
      });
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return null;
    }

    if (!Setting.hasPromptPage(application)) {
      return (
        <Result
          status="error"
          title="Sign Up Error"
          subTitle={"You are unexpected to see this prompt page"}
          extra={[
            <Button type="primary" key="signin" onClick={() => {
              Setting.goToLogin(this, application);
            }}>
              Sign In
            </Button>
          ]}
        >
        </Result>
      );
    }

    return (
      <Row>
        <Col span={24} style={{display: "flex", justifyContent: "center"}}>
          <div style={{marginTop: "80px", marginBottom: "50px", textAlign: "center"}}>
            {
              Setting.renderHelmet(application)
            }
            {
              Setting.renderLogo(application)
            }
            {
              this.renderContent(application)
            }
            <Row style={{margin: 10}}>
              <Col span={18}>
              </Col>
            </Row>
            <div style={{marginTop: "50px"}}>
              <Button disabled={!Setting.isPromptAnswered(this.state.user, application)} type="primary" size="large" onClick={() => {this.submitUserEdit(true);}}>{i18next.t("code:Submit and complete")}</Button>
            </div>
          </div>
        </Col>
      </Row>
    );
  }
}

export default PromptPage;
