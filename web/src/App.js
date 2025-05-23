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

import React, {Component, Suspense, lazy} from "react";
import "./App.less";
import {Helmet} from "react-helmet";
import * as Setting from "./Setting";
import {setOrgIsTourVisible, setTourLogo} from "./TourConfig";
import {StyleProvider, legacyLogicalPropertiesTransformer} from "@ant-design/cssinjs";
import {GithubOutlined, InfoCircleFilled, ShareAltOutlined} from "@ant-design/icons";
import {Alert, Button, ConfigProvider, Drawer, FloatButton, Layout, Result, Tooltip} from "antd";
import {Route, Switch, withRouter} from "react-router-dom";
import CustomGithubCorner from "./common/CustomGithubCorner";
import * as Conf from "./Conf";

import * as Auth from "./auth/Auth";
import EntryPage from "./EntryPage";
import * as AuthBackend from "./auth/AuthBackend";
import AuthCallback from "./auth/AuthCallback";
import SamlCallback from "./auth/SamlCallback";
import i18next from "i18next";
import {withTranslation} from "react-i18next";
const ManagementPage = lazy(() => import("./ManagementPage"));
const {Footer, Content} = Layout;

import {setTwoToneColor} from "@ant-design/icons";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import * as Cookie from "cookie";
import PrivacyPolicyEn from "./static/privacy-policy-en.pdf";
import PrivacyPolicyZh from "./static/privacy-policy-zh.pdf";
import TermsOfServiceEn from "./static/terms-of-service-en.pdf";
import TermsOfServiceZh from "./static/terms-of-service-zh.pdf";

setTwoToneColor("rgb(87,52,211)");

class App extends Component {
  constructor(props) {
    super(props);
    this.setThemeAlgorithm();
    let storageThemeAlgorithm = [];
    try {
      storageThemeAlgorithm = localStorage.getItem("themeAlgorithm") ? JSON.parse(localStorage.getItem("themeAlgorithm")) : ["default"];
    } catch {
      storageThemeAlgorithm = ["default"];
    }
    this.state = {
      classes: props,
      selectedMenuKey: 0,
      account: undefined,
      accessToken: undefined,
      uri: null,
      themeAlgorithm: storageThemeAlgorithm,
      themeData: Conf.ThemeDefault,
      logo: this.getLogo(storageThemeAlgorithm),
      requiredEnableMfa: false,
      isAiAssistantOpen: false,
      application: undefined,
    };
    Setting.initServerUrl();
    Auth.initAuthWithConfig({
      serverUrl: Setting.ServerUrl,
      appName: Conf.DefaultApplication, // the application used in Casdoor root path: "/"
    });
  }

  UNSAFE_componentWillMount() {
    this.updateMenuKey();
    this.getAccount();
    this.getApplication();
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    const uri = location.pathname;
    if (this.state.uri !== uri) {
      this.updateMenuKey();
    }

    if (this.state.account !== prevState.account) {
      const requiredEnableMfa = Setting.isRequiredEnableMfa(this.state.account, this.state.account?.organization);
      this.setState({
        requiredEnableMfa: requiredEnableMfa,
      });

      if (requiredEnableMfa === true) {
        const mfaType = Setting.getMfaItemsByRules(this.state.account, this.state.account?.organization, [Setting.MfaRuleRequired])
          .find((item) => item.rule === Setting.MfaRuleRequired)?.name;
        if (mfaType !== undefined) {
          this.props.history.push(`/mfa/setup?mfaType=${mfaType}`, {from: "/login"});
        }
      }
    }
  }

  updateMenuKey() {
    const uri = location.pathname;
    this.setState({
      uri: uri,
    });
    if (uri === "/" || uri.includes("/shortcuts") || uri.includes("/apps")) {
      this.setState({selectedMenuKey: "/home"});
    } else if (uri.includes("/organizations") || uri.includes("/trees") || uri.includes("/groups") || uri.includes("/users") || uri.includes("/invitations")) {
      this.setState({selectedMenuKey: "/orgs"});
    } else if (uri.includes("/applications") || uri.includes("/providers") || uri.includes("/resources") || uri.includes("/certs")) {
      this.setState({selectedMenuKey: "/identity"});
    } else if (uri.includes("/roles") || uri.includes("/permissions") || uri.includes("/models") || uri.includes("/adapters") || uri.includes("/enforcers")) {
      this.setState({selectedMenuKey: "/auth"});
    } else if (uri.includes("/records") || uri.includes("/tokens") || uri.includes("/sessions")) {
      this.setState({selectedMenuKey: "/logs"});
    } else if (uri.includes("/products") || uri.includes("/payments") || uri.includes("/plans") || uri.includes("/pricings") || uri.includes("/subscriptions")) {
      this.setState({selectedMenuKey: "/business"});
    } else if (uri.includes("/sysinfo") || uri.includes("/syncers") || uri.includes("/webhooks")) {
      this.setState({selectedMenuKey: "/admin"});
    } else if (uri.includes("/signup")) {
      this.setState({selectedMenuKey: "/signup"});
    } else if (uri.includes("/login")) {
      this.setState({selectedMenuKey: "/login"});
    } else if (uri.includes("/result")) {
      this.setState({selectedMenuKey: "/result"});
    } else {
      this.setState({selectedMenuKey: -1});
    }
  }

  getAccessTokenParam(params) {
    // "/page?access_token=123"
    const accessToken = params.get("access_token");
    return accessToken === null ? "" : `?accessToken=${accessToken}`;
  }

  getCredentialParams(params) {
    // "/page?username=abc&password=123"
    if (params.get("username") === null || params.get("password") === null) {
      return "";
    }
    return `?username=${params.get("username")}&password=${params.get("password")}`;
  }

  getUrlWithoutQuery() {
    return window.location.toString().replace(window.location.search, "");
  }

  getLanguageParam(params) {
    // "/page?language=en"
    const language = params.get("language");
    if (language !== null) {
      Setting.setLanguage(language);
      return `language=${language}`;
    }
    return "";
  }

  getLogo(themes) {
    return Setting.getLogo(themes);
  }

  setThemeAlgorithm() {
    const currentUrl = window.location.href;
    const url = new URL(currentUrl);
    const themeType = url.searchParams.get("theme");
    if (themeType === "dark" || themeType === "default") {
      localStorage.setItem("themeAlgorithm", JSON.stringify([themeType]));
    }
  }

  setLanguage(account) {
    const language = account?.language;
    if (language !== null && language !== "" && language !== i18next.language) {
      Setting.setLanguage(language);
    }
  }

  setTheme = (theme, initThemeAlgorithm) => {
    this.setState({
      themeData: theme,
    });

    if (initThemeAlgorithm) {
      if (localStorage.getItem("themeAlgorithm")) {
        let storageThemeAlgorithm = [];
        try {
          storageThemeAlgorithm = JSON.parse(localStorage.getItem("themeAlgorithm"));
        } catch {
          storageThemeAlgorithm = ["default"];
        }
        this.setState({
          logo: this.getLogo(storageThemeAlgorithm),
          themeAlgorithm: storageThemeAlgorithm,
        });
        return;
      }
      this.setState({
        logo: this.getLogo(Setting.getAlgorithmNames(theme)),
        themeAlgorithm: Setting.getAlgorithmNames(theme),
      });
    }
  };

  getApplication() {
    const applicationName = localStorage.getItem("applicationName");
    if (!applicationName) {
      return;
    }
    ApplicationBackend.getApplication("admin", applicationName)
      .then((res) => {
        if (res.status === "error") {
          Setting.showMessage("error", res.msg);
          return;
        }

        this.setState({
          application: res.data,
        });
      });
  }

  getAccount() {
    const params = new URLSearchParams(this.props.location.search);

    let query = this.getAccessTokenParam(params);
    if (query === "") {
      query = this.getCredentialParams(params);
    }

    const query2 = this.getLanguageParam(params);
    if (query2 !== "") {
      const url = window.location.toString().replace(new RegExp(`[?&]${query2}`), "");
      window.history.replaceState({}, document.title, url);
    }

    if (query !== "") {
      window.history.replaceState({}, document.title, this.getUrlWithoutQuery());
    }

    AuthBackend.getAccount(query)
      .then((res) => {
        let account = null;
        let accessToken = null;
        if (res.status === "ok") {
          account = res.data;
          account.organization = res.data2;
          accessToken = res.data.accessToken;

          this.setLanguage(account);
          this.setTheme(Setting.getThemeData(account.organization), Conf.InitThemeAlgorithm);
          setTourLogo(account.organization.logo);
          setOrgIsTourVisible(account.organization.enableTour);
        } else {
          if (res.data !== "Please login first") {
            Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
          }
        }

        this.setState({
          account: account,
          accessToken: accessToken,
        });
      });
  }

  onUpdateAccount(account) {
    this.setState({
      account: account,
    });
  }

  renderFooter(logo, footerHtml) {
    logo = logo ?? this.state.logo;
    footerHtml = footerHtml ?? this.state.application?.footerHtml;
    const language = Setting.getLanguage();
    const privacyPolicy = language === "en" ? PrivacyPolicyEn : PrivacyPolicyZh;
    const termsOfService = language === "en" ? TermsOfServiceEn : TermsOfServiceZh;
    const isLoginSuccessPage = this.props.location.pathname === "/login/success";
    const commonStyle = {
      fontWeight: 600, color: isLoginSuccessPage ? "#fff" : "",
    };
    return (
      <React.Fragment>
        {!this.state.account ? null : <div style={{display: "none"}} id="CasdoorApplicationName" value={this.state.account.signupApplication} />}
        {!this.state.account ? null : <div style={{display: "none"}} id="CasdoorAccessToken" value={this.state.accessToken} />}
        <Footer id="footer" style={
          {
            textAlign: "center",
            zIndex: 1000,
            ...(isLoginSuccessPage
              ? {
                backgroundColor: "#000",
                color: "#fff",
              }
              : {}),
          }
        }>
          {
            footerHtml && footerHtml !== "" ?
              <React.Fragment>
                <div dangerouslySetInnerHTML={{__html: footerHtml}} />
              </React.Fragment>
              : (
                Conf.CustomFooter !== null ? Conf.CustomFooter : (
                  <React.Fragment>
                    {/* Powered by <a target="_blank" href="https://casdoor.org" rel="noreferrer"><img style={{paddingBottom: "3px"}} height={"20px"} alt={"Casdoor"} src={logo} /></a> */}
                    <div className="terms-privacy" style={{display: "flex", justifyContent: "center", fontSize: "14px"}}>
                      <div style={{opacity: "0.5"}}>{i18next.t("login:Login by acceptance")}</div>
                      <a
                        href={termsOfService}
                        className="terms-link"
                        style={commonStyle}
                        target="open"
                      >
                        {i18next.t("login:Terms of Service")}
                      </a>
                      <a
                        href={privacyPolicy}
                        className="privacy-link"
                        style={commonStyle}
                        target="open"
                      >
                        {i18next.t("login:Privacy Policy")}
                      </a>
                    </div>
                  </React.Fragment>
                )
              )
          }
        </Footer>
      </React.Fragment>
    );
  }

  renderAiAssistant() {
    return (
      <Drawer
        title={
          <React.Fragment>
            <Tooltip title="Want to deploy your own AI assistant? Click to learn more!">
              <a target="_blank" rel="noreferrer" href={"https://casdoor.com"}>
                <img style={{width: "20px", marginRight: "10px", marginBottom: "2px"}} alt="help" src="https://casbin.org/img/casbin.svg" />
                AI Assistant
              </a>
            </Tooltip>
            <a className="custom-link" style={{float: "right", marginTop: "2px"}} target="_blank" rel="noreferrer" href={`${Conf.AiAssistantUrl}`}>
              <ShareAltOutlined className="custom-link" style={{fontSize: "20px", color: "rgb(140,140,140)"}} />
            </a>
            <a className="custom-link" style={{float: "right", marginRight: "30px", marginTop: "2px"}} target="_blank" rel="noreferrer" href={"https://github.com/casibase/casibase"}>
              <GithubOutlined className="custom-link" style={{fontSize: "20px", color: "rgb(140,140,140)"}} />
            </a>
          </React.Fragment>
        }
        placement="right"
        width={500}
        mask={false}
        onClose={() => {
          this.setState({
            isAiAssistantOpen: false,
          });
        }}
        open={this.state.isAiAssistantOpen}
      >
        <iframe id="iframeHelper" title={"iframeHelper"} src={`${Conf.AiAssistantUrl}/?isRaw=1`} width="100%" height="100%" scrolling="no" frameBorder="no" />
      </Drawer>
    );
  }

  isDoorPages() {
    return this.isEntryPages() || window.location.pathname.startsWith("/callback");
  }

  isEntryPages() {
    return window.location.pathname.startsWith("/signup") ||
        window.location.pathname.startsWith("/login") ||
        window.location.pathname.startsWith("/forget") ||
        window.location.pathname.startsWith("/prompt") ||
        window.location.pathname.startsWith("/result") ||
        window.location.pathname.startsWith("/cas") ||
        window.location.pathname.startsWith("/select-plan") ||
        window.location.pathname.startsWith("/buy-plan") ||
        window.location.pathname.startsWith("/qrcode") ||
        window.location.pathname.startsWith("/captcha");
  }

  onClick = ({key}) => {
    if (key !== "/swagger" && key !== "/records") {
      if (this.state.requiredEnableMfa) {
        Setting.showMessage("info", "Please enable MFA first!");
      } else {
        this.props.history.push(key);
      }
    }
  };

  onLoginSuccess(redirectUrl) {
    window.google?.accounts?.id?.cancel();
    if (redirectUrl) {
      localStorage.setItem("mfaRedirectUrl", redirectUrl);
    }
    this.getAccount();
  }

  renderPage() {
    if (this.isDoorPages()) {
      let themeData = this.state.themeData;
      let logo = this.state.logo;
      let footerHtml = null;
      if (this.state.organization === undefined) {
        const curCookie = Cookie.parse(document.cookie);
        if (curCookie["organizationTheme"] && curCookie["organizationTheme"] !== "null") {
          themeData = JSON.parse(curCookie["organizationTheme"]);
        }
        if (curCookie["organizationLogo"] && curCookie["organizationLogo"] !== "") {
          logo = curCookie["organizationLogo"];
        }
        if (curCookie["organizationFootHtml"] && curCookie["organizationFootHtml"] !== "") {
          footerHtml = curCookie["organizationFootHtml"];
        }
      }

      return (
        <ConfigProvider theme={{
          token: {
            colorPrimary: themeData.colorPrimary,
            borderRadius: themeData.borderRadius,
          },
          algorithm: Setting.getAlgorithm(this.state.themeAlgorithm),
        }}>
          <StyleProvider hashPriority="high" transformers={[legacyLogicalPropertiesTransformer]}>
            <Layout id="parent-area">
              <Content style={{display: "flex", justifyContent: "center"}}>
                {
                  this.isEntryPages() ?
                    <EntryPage
                      account={this.state.account}
                      theme={this.state.themeData}
                      themeAlgorithm={this.state.themeAlgorithm}
                      updateApplication={(application) => {
                        this.setState({
                          application: application,
                        });
                      }}
                      onLoginSuccess={(redirectUrl) => {this.onLoginSuccess(redirectUrl);}}
                      onUpdateAccount={(account) => this.onUpdateAccount(account)}
                      updataThemeData={this.setTheme}
                    /> :
                    <Switch>
                      <Route exact path="/callback" render={(props) => <AuthCallback {...props} {...this.props} application={this.state.application} onLoginSuccess={(redirectUrl) => {this.onLoginSuccess(redirectUrl);}} />} />
                      <Route exact path="/callback/saml" render={(props) => <SamlCallback {...props} {...this.props} application={this.state.application} onLoginSuccess={(redirectUrl) => {this.onLoginSuccess(redirectUrl);}} />} />
                      <Route path="" render={() => <Result status="404" title="404 NOT FOUND" subTitle={i18next.t("general:Sorry, the page you visited does not exist.")}
                        extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>} />} />
                    </Switch>
                }
              </Content>
              {
                this.renderFooter(logo, footerHtml)
              }
              {
                this.renderAiAssistant()
              }
            </Layout>
          </StyleProvider>
        </ConfigProvider>
      );
    }
    return (
      <React.Fragment>
        {/* { */}
        {/*   this.renderBanner() */}
        {/* } */}
        <FloatButton.BackTop />
        <CustomGithubCorner />
        {
          <Suspense fallback={null}>
            <Layout id="parent-area">
              <ManagementPage
                account={this.state.account}
                application={this.state.application}
                uri={this.state.uri}
                themeData={this.state.themeData}
                themeAlgorithm={this.state.themeAlgorithm}
                selectedMenuKey={this.state.selectedMenuKey}
                requiredEnableMfa={this.state.requiredEnableMfa}
                menuVisible={this.state.menuVisible}
                logo={this.state.logo}
                onChangeTheme={this.setTheme}
                onClick = {this.onClick}
                onfinish={() => {
                  this.setState({requiredEnableMfa: false});
                }}
                openAiAssistant={() => {
                  this.setState({
                    isAiAssistantOpen: true,
                  });
                }}
                setLogoAndThemeAlgorithm={(nextThemeAlgorithm) => {
                  this.setState({
                    themeAlgorithm: nextThemeAlgorithm,
                    logo: this.getLogo(nextThemeAlgorithm),
                  });
                  localStorage.setItem("themeAlgorithm", JSON.stringify(nextThemeAlgorithm));
                }}
                setLogoutState={() => {
                  this.setState({
                    account: null,
                  });
                }}
              />
              {
                this.renderFooter()
              }
              {
                this.renderAiAssistant()
              }
            </Layout>
          </Suspense>
        }
      </React.Fragment>
    );
  }

  renderBanner() {
    if (!Conf.IsDemoMode) {
      return null;
    }

    const language = Setting.getLanguage();
    if (language === "en" || language === "zh") {
      return null;
    }

    return (
      <Alert type="info" banner showIcon={false} closable message={
        <div style={{textAlign: "center"}}>
          <InfoCircleFilled style={{color: "rgb(87,52,211)"}} />
          &nbsp;&nbsp;
          {i18next.t("general:Found some texts still not translated? Please help us translate at")}
          &nbsp;
          <a target="_blank" rel="noreferrer" href={"https://crowdin.com/project/casdoor-site"}>
            Crowdin
          </a>
          &nbsp;!&nbsp;🙏
        </div>
      } />
    );
  }

  render() {
    return (
      <React.Fragment>
        {(this.state.account === undefined || this.state.account === null) ?
          <Helmet>
            <link rel="icon" href={"https://cdn.casdoor.com/static/favicon.png"} />
          </Helmet> :
          <Helmet>
            <title>{this.state.account.organization?.displayName}</title>
            <link rel="icon" href={this.state.account.organization?.favicon} />
          </Helmet>
        }
        <ConfigProvider theme={{
          token: {
            colorPrimary: this.state.themeData.colorPrimary,
            colorInfo: this.state.themeData.colorPrimary,
            borderRadius: this.state.themeData.borderRadius,
          },
          algorithm: Setting.getAlgorithm(this.state.themeAlgorithm),
        }}>
          <StyleProvider hashPriority="high" transformers={[legacyLogicalPropertiesTransformer]}>
            {
              this.renderPage()
            }
          </StyleProvider>
        </ConfigProvider>
      </React.Fragment>
    );
  }
}

export default withRouter(withTranslation()(App));
