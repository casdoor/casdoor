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
import {Redirect, Route, Switch} from "react-router-dom";
import {ConfigProvider, Spin, theme} from "antd";
import i18next from "i18next";
import * as Setting from "./Setting";
import SignupPage from "./auth/SignupPage";
import SelfLoginPage from "./auth/SelfLoginPage";
import LoginPage from "./auth/LoginPage";
import SelfForgetPage from "./auth/SelfForgetPage";
import ForgetPage from "./auth/ForgetPage";
import PromptPage from "./auth/PromptPage";
import CasLogout from "./auth/CasLogout";
import {ThemeDefault} from "./common/theme/ThemeEditor";

class EntryPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      application: undefined,
    };
  }

  renderHomeIfLoggedIn(component) {
    if (this.props.account !== null && this.props.account !== undefined) {
      return <Redirect to="/" />;
    } else {
      return component;
    }
  }

  renderLoginIfNotLoggedIn(component) {
    if (this.props.account === null) {
      sessionStorage.setItem("from", window.location.pathname);
      return <Redirect to="/login" />;
    } else if (this.props.account === undefined) {
      return null;
    } else {
      return component;
    }
  }

  getApplicationObj() {
    return this.state.application || null;
  }

  initAlgorithm = (themeData) => {
    // init Algorithm according the theme type
    const algorithms = [themeData.themeType !== "dark" ? theme.defaultAlgorithm : theme.darkAlgorithm];

    if (themeData.isCompact === true) {
      algorithms.push(theme.compactAlgorithm);
    }

    return algorithms;
  };

  render() {
    const application = this.getApplicationObj();
    const themeData = application !== null ? Setting.getThemeData(application.organizationObj, application) : ThemeDefault;

    const onUpdateApplication = (application) => {
      this.setState({
        application: application,
      });
    };

    return (
      <ConfigProvider theme={{
        token: {
          colorPrimary: themeData.colorPrimary,
          colorInfo: themeData.colorPrimary,
        },
        algorithm: this.initAlgorithm(themeData),
      }}>
        <div className="loginBackground" style={{backgroundImage: Setting.inIframe() || Setting.isMobile() ? null : `url(${this.state.application?.formBackgroundUrl})`}}>
          <Spin spinning={this.state.application === undefined} tip={i18next.t("login:Loading")} style={{margin: "0 auto"}} />
          <Switch>
            <Route exact path="/signup" render={(props) => this.renderHomeIfLoggedIn(<SignupPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/signup/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<SignupPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/login" render={(props) => this.renderHomeIfLoggedIn(<SelfLoginPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/login/:owner" render={(props) => this.renderHomeIfLoggedIn(<SelfLoginPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/auto-signup/oauth/authorize" render={(props) => <LoginPage {...this.props} application={this.state.application} type={"code"} mode={"signup"} onUpdateApplication={onUpdateApplication}{...props} />} />
            <Route exact path="/signup/oauth/authorize" render={(props) => <SignupPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/login/oauth/authorize" render={(props) => <LoginPage {...this.props} application={this.state.application} type={"code"} mode={"signin"} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/login/saml/authorize/:owner/:applicationName" render={(props) => <LoginPage {...this.props} application={this.state.application} type={"saml"} mode={"signin"} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/forget" render={(props) => this.renderHomeIfLoggedIn(<SelfForgetPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/forget/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<ForgetPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/prompt" render={(props) => this.renderLoginIfNotLoggedIn(<PromptPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/prompt/:applicationName" render={(props) => this.renderLoginIfNotLoggedIn(<PromptPage {...this.props} application={this.state.application} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/cas/:owner/:casApplicationName/logout" render={(props) => this.renderHomeIfLoggedIn(<CasLogout {...this.props} application={this.state.application} {...props} />)} />
            <Route exact path="/cas/:owner/:casApplicationName/login" render={(props) => {return (<LoginPage {...this.props} application={this.state.application} type={"cas"} mode={"signup"} {...props} />);}} />
          </Switch>
        </div>
      </ConfigProvider>
    );
  }
}

export default EntryPage;
