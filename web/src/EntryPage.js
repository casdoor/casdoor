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

import React, {Suspense} from "react";
import {Redirect, Route, Switch} from "react-router-dom";
import {Spin} from "antd";
import i18next from "i18next";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as Setting from "./Setting";
import * as Conf from "./Conf";
import {setTheme} from "./store/themeSlice";
import {connect} from "react-redux";
import entryRoutes from "./routers/entry";
import CustomHead from "./basic/CustomHead";

class EntryPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      application: undefined,
      pricing: undefined,
    };
  }

  renderHomeIfLoggedIn(component) {
    if (this.props.account !== null && this.props.account !== undefined) {
      return <Redirect to={{pathname: "/", state: {from: "/login"}}} />;
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

  render() {
    const onUpdateApplication = (application) => {
      this.setState({
        application: application,
      });

      const themeData = application !== null ? Setting.getThemeData(application.organizationObj, application) : Conf.ThemeDefault;
      this.props.setTheme(themeData);
    };

    const onUpdatePricing = (pricing) => {
      this.setState({
        pricing: pricing,
      });

      ApplicationBackend.getApplication("admin", pricing.application)
        .then((res) => {
          if (res.status === "error") {
            Setting.showMessage("error", res.msg);
            return;
          }

          const application = res.data;
          const themeData = application !== null ? Setting.getThemeData(application.organizationObj, application) : Conf.ThemeDefault;
          this.props.setTheme(themeData);
        });
    };

    return (
      <React.Fragment>
        <CustomHead headerHtml={this.state.application?.headerHtml} />
        <div className="loginBackground"
          style={{backgroundImage: Setting.inIframe() || Setting.isMobile() ? null : `url(${this.state.application?.formBackgroundUrl})`}}>
          <Spin size="large" spinning={this.state.application === undefined && this.state.pricing === undefined} tip={i18next.t("login:Loading")}
            style={{margin: "0 auto"}} />
          <Suspense fallback={<div></div>}>
            <Switch>
              {entryRoutes.map((el, index) => {
                const Element = el.component;
                return (<Route key="" exact={el.exact} path={el.path} render={(props) => {
                  if (el.auth && el.unAuthRedirect === "home") {
                    props.onUpdateApplication = onUpdateApplication;
                    props.application = this.state.application;
                    return this.renderHomeIfLoggedIn(<Element {...this.props} {...props} {...el.props} />);
                  } else if (el.auth && el.unAuthRedirect === "login") {
                    props.onUpdateApplication = onUpdateApplication;
                    props.application = this.state.application;
                    return this.renderLoginIfNotLoggedIn(<Element {...this.props} {...props} {...el.props} />);
                  } else {
                    props.pricing = this.state.pricing;
                    props.onUpdatePricing = onUpdatePricing;
                    props.onUpdateApplication = onUpdateApplication;
                    return <Element {...this.props} {...props} {...el.props} />;
                  }
                }} />);
              })}
            </Switch>
          </Suspense>
        </div>
      </React.Fragment>
    );
  }
}

export default connect(null, {setTheme})(EntryPage);
