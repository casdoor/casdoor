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
import {Select, Spin} from "antd";
import {Switch as AntSwitch} from "antd";
import i18next from "i18next";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import PricingPage from "./pricing/PricingPage";
import * as Setting from "./Setting";
import * as Conf from "./Conf";
import SignupPage from "./auth/SignupPage";
import SelfLoginPage from "./auth/SelfLoginPage";
import LoginPage from "./auth/LoginPage";
import SelfForgetPage from "./auth/SelfForgetPage";
import ForgetPage from "./auth/ForgetPage";
import PromptPage from "./auth/PromptPage";
import ResultPage from "./auth/ResultPage";
import CasLogout from "./auth/CasLogout";
import {authConfig} from "./auth/Auth";
import ProductBuyPage from "./ProductBuyPage";
import PaymentResultPage from "./PaymentResultPage";
import QrCodePage from "./QrCodePage";
import CaptchaPage from "./CaptchaPage";
import CustomHead from "./basic/CustomHead";
import * as Util from "./auth/Util";

class EntryPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      application: undefined,
      pricing: undefined,
      carouselApplications: [],
      carouselCurrentIndex: 0,
      carouselAutoScroll: true,
    };
    this.carouselTimer = null;
  }

  componentDidUpdate(prevProps, prevState) {
    // Load carousel applications when main application changes
    if (this.state.application !== prevState.application && this.state.application?.carouselApplications?.length > 0) {
      this.loadCarouselApplications(this.state.application.carouselApplications);
    }

    // Start/stop carousel timer based on auto-scroll state
    if (this.state.carouselAutoScroll !== prevState.carouselAutoScroll) {
      if (this.state.carouselAutoScroll) {
        this.startCarouselTimer();
      } else {
        this.stopCarouselTimer();
      }
    }
  }

  componentWillUnmount() {
    this.stopCarouselTimer();
  }

  loadCarouselApplications = async(carouselAppNames) => {
    const apps = [];
    for (const appName of carouselAppNames) {
      const parts = appName.split("/");
      if (parts.length === 2) {
        const [owner, name] = parts;
        try {
          const res = await ApplicationBackend.getApplication(owner, name);
          if (res.status === "ok" && res.data) {
            apps.push(res.data);
          }
        } catch (error) {
          // Failed to load carousel application
        }
      }
    }
    this.setState({carouselApplications: apps}, () => {
      if (this.state.carouselAutoScroll) {
        this.startCarouselTimer();
      }
    });
  };

  startCarouselTimer = () => {
    this.stopCarouselTimer();
    if (this.state.carouselApplications.length > 0) {
      this.carouselTimer = setInterval(() => {
        this.setState(prevState => ({
          carouselCurrentIndex: (prevState.carouselCurrentIndex + 1) % prevState.carouselApplications.length,
        }));
      }, 5000); // 5 seconds
    }
  };

  stopCarouselTimer = () => {
    if (this.carouselTimer) {
      clearInterval(this.carouselTimer);
      this.carouselTimer = null;
    }
  };

  getDisplayApplication = () => {
    if (this.state.carouselApplications.length > 0) {
      return this.state.carouselApplications[this.state.carouselCurrentIndex];
    }
    return this.state.application;
  };

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
      this.props.updataThemeData(themeData);
      this.props.updateApplication(application);

      if (application) {
        localStorage.setItem("applicationName", application.name);
      }
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
          this.props.updataThemeData(themeData);
        });
    };

    if (this.state.application?.ipRestriction) {
      return Util.renderMessageLarge(this, this.state.application.ipRestriction);
    }

    if (this.state.application?.organizationObj?.ipRestriction) {
      return Util.renderMessageLarge(this, this.state.application.organizationObj.ipRestriction);
    }

    const isDarkMode = this.props.themeAlgorithm.includes("dark");
    const displayApplication = this.getDisplayApplication();

    return (
      <React.Fragment>
        <CustomHead headerHtml={this.state.application?.headerHtml} />
        {this.state.carouselApplications.length > 0 && (
          <div style={{
            position: "fixed",
            top: 20,
            right: 20,
            zIndex: 1000,
            backgroundColor: "rgba(255, 255, 255, 0.9)",
            padding: "15px",
            borderRadius: "8px",
            boxShadow: "0 2px 8px rgba(0,0,0,0.15)",
            display: "flex",
            flexDirection: "column",
            gap: "10px",
            minWidth: "250px",
          }}>
            <div style={{display: "flex", alignItems: "center", justifyContent: "space-between"}}>
              <span style={{fontWeight: "500", marginRight: "10px"}}>
                {i18next.t("login:Auto scroll")}:
              </span>
              <AntSwitch
                checked={this.state.carouselAutoScroll}
                onChange={(checked) => this.setState({carouselAutoScroll: checked})}
              />
            </div>
            <div>
              <div style={{fontWeight: "500", marginBottom: "5px"}}>
                {i18next.t("login:Current application")}:
              </div>
              <Select
                style={{width: "100%"}}
                value={this.state.carouselCurrentIndex}
                onChange={(index) => {
                  this.setState({carouselCurrentIndex: index});
                  if (this.state.carouselAutoScroll) {
                    this.startCarouselTimer();
                  }
                }}
              >
                {this.state.carouselApplications.map((app, index) => (
                  <Select.Option key={index} value={index}>
                    {app.displayName || app.name}
                  </Select.Option>
                ))}
              </Select>
            </div>
          </div>
        )}
        <div className={`${isDarkMode ? "loginBackgroundDark" : "loginBackground"}`}
          style={{backgroundImage: Setting.inIframe() ? null : (Setting.isMobile() ? `url(${displayApplication?.formBackgroundUrlMobile})` : `url(${displayApplication?.formBackgroundUrl})`)}}>
          <Spin size="large" spinning={this.state.application === undefined && this.state.pricing === undefined} tip={i18next.t("login:Loading")}
            style={{width: "100%", margin: "0 auto", position: "absolute"}} />
          <Switch>
            <Route exact path="/signup" render={(props) => this.renderHomeIfLoggedIn(<SignupPage {...this.props} application={displayApplication} applicationName={authConfig.appName} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/signup/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<SignupPage {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/login" render={(props) => this.renderHomeIfLoggedIn(<SelfLoginPage {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/login/:owner" render={(props) => this.renderHomeIfLoggedIn(<SelfLoginPage {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/signup/oauth/authorize" render={(props) => <SignupPage {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/login/oauth/authorize" render={(props) => <LoginPage {...this.props} application={displayApplication} type={"code"} mode={"signin"} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/login/oauth/device/:userCode" render={(props) => <LoginPage {...this.props} application={displayApplication} type={"device"} mode={"signin"} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/login/saml/authorize/:owner/:applicationName" render={(props) => <LoginPage {...this.props} application={displayApplication} type={"saml"} mode={"signin"} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/forget" render={(props) => <SelfForgetPage {...this.props} account={this.props.account} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/forget/:applicationName" render={(props) => <ForgetPage {...this.props} account={this.props.account} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/prompt" render={(props) => this.renderLoginIfNotLoggedIn(<PromptPage {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/prompt/:applicationName" render={(props) => this.renderLoginIfNotLoggedIn(<PromptPage {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/result" render={(props) => this.renderHomeIfLoggedIn(<ResultPage {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/result/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<ResultPage {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/cas/:owner/:casApplicationName/logout" render={(props) => this.renderHomeIfLoggedIn(<CasLogout {...this.props} application={displayApplication} onUpdateApplication={onUpdateApplication} {...props} />)} />
            <Route exact path="/cas/:owner/:casApplicationName/login" render={(props) => {return (<LoginPage {...this.props} application={displayApplication} type={"cas"} mode={"signin"} onUpdateApplication={onUpdateApplication} {...props} />);}} />
            <Route exact path="/select-plan/:owner/:pricingName" render={(props) => <PricingPage {...this.props} pricing={this.state.pricing} onUpdatePricing={onUpdatePricing} {...props} />} />
            <Route exact path="/buy-plan/:owner/:pricingName" render={(props) => <ProductBuyPage {...this.props} pricing={this.state.pricing} onUpdatePricing={onUpdatePricing} {...props} />} />
            <Route exact path="/buy-plan/:owner/:pricingName/result" render={(props) => <PaymentResultPage {...this.props} pricing={this.state.pricing} onUpdatePricing={onUpdatePricing} {...props} />} />
            <Route exact path="/qrcode/:owner/:paymentName" render={(props) => <QrCodePage {...this.props} onUpdateApplication={onUpdateApplication} {...props} />} />
            <Route exact path="/captcha" render={(props) => <CaptchaPage {...props} />} />
          </Switch>
        </div>

      </React.Fragment>
    );
  }
}

export default EntryPage;
