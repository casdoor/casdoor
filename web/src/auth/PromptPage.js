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
import {Button, Card, Col, Result, Row} from "antd";
import * as ApplicationBackend from "../backend/ApplicationBackend";
import * as UserBackend from "../backend/UserBackend";
import * as AuthBackend from "./AuthBackend";
import * as Setting from "../Setting";
import i18next from "i18next";
import AffiliationSelect from "../common/select/AffiliationSelect";
import OAuthWidget from "../common/OAuthWidget";
import RegionSelect from "../common/select/RegionSelect";
import {withRouter} from "react-router-dom";
import MfaSetupPage, {EmailMfaType, SmsMfaType, TotpMfaType} from "./MfaSetupPage";

class PromptPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      type: props.type,
      applicationName: props.applicationName ?? (props.match === undefined ? null : props.match.params.applicationName),
      application: null,
      user: null,
      promptType: new URLSearchParams(this.props.location.search).get("promptType") || "provider",
      steps: null,
      current: 0,
    };
  }

  UNSAFE_componentWillMount() {
    this.getUser();
    if (this.getApplicationObj() === null) {
      this.getApplication();
    }
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    if (this.state.user !== null && this.getApplicationObj() !== null && this.state.steps === null) {
      // eslint-disable-next-line no-console
      console.log(this.getApplicationObj());
      this.initSteps(this.state.user, this.getApplicationObj());
    }
  }

  getUser() {
    const organizationName = this.props.account.owner;
    const userName = this.props.account.name;
    UserBackend.getUser(organizationName, userName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          user: res,
        });
      });
  }

  getApplication() {
    if (this.state.applicationName === null) {
      return;
    }

    ApplicationBackend.getApplication("admin", this.state.applicationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.onUpdateApplication(res);
        this.setState({
          application: res,
        });
      });
  }

  getApplicationObj() {
    return this.props.application ?? this.state.application;
  }

  onUpdateApplication(application) {
    this.props.onUpdateApplication(application);
  }

  parseUserField(key, value) {
    // if ([].includes(key)) {
    //   value = Setting.myParseInt(value);
    // }
    return value;
  }

  updateUserField(key, value) {
    value = this.parseUserField(key, value);

    const user = this.state.user;
    user[key] = value;
    this.setState({
      user: user,
    });

    this.submitUserEdit(false);
  }

  updateUserFieldWithoutSubmit(key, value) {
    value = this.parseUserField(key, value);

    const user = this.state.user;
    user[key] = value;
    this.setState({
      user: user,
    });
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
      <div style={{width: "500px"}}>
        {
          this.renderAffiliation(application)
        }
        <div>
          {
            (application === null || this.state.user === null) ? null : (
              application?.providers.filter(providerItem => Setting.isProviderPrompted(providerItem)).map((providerItem, index) => <OAuthWidget key={providerItem.name} labelSpan={6} user={this.state.user} application={application} providerItem={providerItem} account={this.props.account} onUnlinked={() => {return this.unlinked();}} />)
            )
          }
          {
            (application === null || this.state.user === null) ? null : (
              application?.signupItems.filter(signupItem => Setting.isSignupItemPrompted(signupItem)).map((signupItem, index) => {
                if (signupItem.name !== "Country/Region") {
                  return null;
                }
                return (
                  <Row key={signupItem.name} style={{marginTop: "20px", justifyContent: "space-between"}} >
                    <Col style={{marginTop: "5px"}} >
                      <span style={{marginLeft: "5px"}}>
                        {
                          i18next.t("user:Country/Region")
                        }:
                      </span>
                    </Col>
                    <Col >
                      <RegionSelect defaultValue={this.state.user.region} onChange={(value) => {
                        this.updateUserFieldWithoutSubmit("region", value);
                      }} />
                    </Col>
                  </Row>
                );
              })
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

  finishAndJump() {
    AuthBackend.logout()
      .then((res) => {
        if (res.status === "ok") {
          this.onUpdateAccount(null);

          let redirectUrl = this.getRedirectUrl();
          if (redirectUrl === "") {
            redirectUrl = res.data2;
          }
          if (redirectUrl !== "" && redirectUrl !== null) {
            Setting.goToLink(redirectUrl);
          } else {
            Setting.redirectToLoginPage(this.getApplicationObj(), this.props.history);
          }
        } else {
          Setting.showMessage("error", `Failed to log out: ${res.msg}`);
        }
      });
  }

  submitUserEdit(isFinal) {
    const user = Setting.deepCopy(this.state.user);
    UserBackend.updateUser(this.state.user.owner, this.state.user.name, user)
      .then((res) => {
        if (res.status === "ok") {
          if (isFinal) {
            Setting.showMessage("success", i18next.t("general:Successfully saved"));
            this.finishAndJump();
          }
        } else {
          if (isFinal) {
            Setting.showMessage("error", res.msg);
          }
        }
      })
      .catch(error => {
        if (isFinal) {
          Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
        }
      });
  }

  isFinishedAllPrompts(application) {
    if (application === null || this.state.user === null) {
      return false;
    }

    return !this.state.steps.some((step) => {return step.finished !== true;});
  }

  renderPromptProvider(application) {
    return (
      <div style={{display: "flex", alignItems: "center", flexDirection: "column"}}>
        {this.renderContent(application)}
        <Button style={{marginTop: "50px", width: "200px"}}
          disabled={!Setting.isPromptAnswered(this.state.user, application)}
          type="primary" size="large" onClick={() => {
            this.submitUserEdit(true);
          }}>
          {i18next.t("code:Submit and complete")}
        </Button>
      </div>);
  }

  getRequiredMfaTypes(organization) {
    const types = organization.mfaItems.map((mfaItem) => {
      if (mfaItem.rule === "Required") {
        return mfaItem.name;
      } else {
        return "";
      }
    });

    return types.filter((type) => {return type !== "";});
  }

  isPromptMfaEnable(organization) {
    const requiredMfaTypes = this.getRequiredMfaTypes(organization);
    if (requiredMfaTypes.length === 0) {
      return false;
    }

    return requiredMfaTypes.some((mfaType) => {
      if (mfaType === EmailMfaType && !this.state.user.mfaEmailEnabled) {
        return true;
      }
      if (mfaType === SmsMfaType && !this.state.user.mfaPhoneEnabled) {
        return true;
      }
      if (mfaType === TotpMfaType && this.state.user.totpSecret === "") {
        return true;
      }
      return false;
    });
  }

  renderPromptMfa() {
    return (
      <MfaSetupPage
        application={this.getApplicationObj()}
        account={this.props.account}
        current={1}
        isAuthenticated={true}
        isPromptPage={true}
        redirectUri={this.getRedirectUrl()}
        requiredMfaTypes={this.getRequiredMfaTypes(this.getApplicationObj().organizationObj)}
        onFinished={() => {
          this.finishAndJump();
        }}
        {...this.props}
      />
    );
  }

  initSteps(user, application) {
    const steps = [];

    if (this.isPromptMfaEnable(application.organizationObj) && this.state.promptType === "mfa") {
      steps.push({
        content: this.renderPromptMfa(application),
        name: "mfa",
        title: i18next.t("mfa:Protect your account with Multi-factor authentication"),
      });
    }

    if (!Setting.isPromptAnswered(user, application) && this.state.promptType === "provider") {
      steps.push({
        content: this.renderPromptProvider(application),
        name: "provider",
        title: i18next.t("application:Binding providers"),
      });
    }

    this.setState({
      steps: steps,
    });
  }

  renderSteps() {
    if (this.state.steps === null || this.state.steps?.length === 0) {
      return null;
    }

    return (
      <Card
        title={this.state.steps[this.state.current].title}
      >
        <div >{this.state.steps[this.state.current].content}</div>
      </Card>
    );
  }

  render() {
    const application = this.getApplicationObj();
    if (application === null) {
      return null;
    }

    if (this.state.steps?.length === 0) {
      return (
        <Result
          style={{display: "flex", flex: "1 1 0%", justifyContent: "center", flexDirection: "column"}}
          status="error"
          title={i18next.t("application:Sign Up Error")}
          subTitle={i18next.t("application:You are unexpected to see this prompt page")}
          extra={[
            <Button type="primary" key="signin" onClick={() => Setting.redirectToLoginPage(application, this.props.history)}>
              {
                i18next.t("login:Sign In")
              }
            </Button>,
          ]}
        >
        </Result>
      );
    }

    return (
      <div style={{display: "flex", flex: "1", justifyContent: "center"}}>
        {this.renderSteps()}
      </div>
    );
  }
}

export default withRouter(PromptPage);
