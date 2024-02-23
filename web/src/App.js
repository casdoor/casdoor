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

const {Footer, Content} = Layout;

import {setTwoToneColor} from "@ant-design/icons";
import {setLogo, setTheme, setThemeAlgorithm} from "./store/themeSlice";
import {connect} from "react-redux";
import {setAccount} from "./store/accountSlice";

const AppContent = lazy(() => import("./AppContent"));

setTwoToneColor("rgb(87,52,211)");

const mapStateToProps = (state) => ({
  themeData: state.theme.value,
  logo: state.theme.logo,
  themeAlgorithm: state.theme.themeAlgorithm,
  account: state.account.value,
});

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      selectedMenuKey: 0,
      uri: null,
      menuVisible: false,
      requiredEnableMfa: false,
      isAiAssistantOpen: false,
    };
    Setting.initServerUrl();
    Auth.initAuthWithConfig({
      serverUrl: Setting.ServerUrl,
      appName: Conf.DefaultApplication, // the application used in Casdoor root path: "/"
    });
  }

  componentDidMount() {
    this.updateMenuKey();
    this.getAccount();
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    const uri = location.pathname;
    if (this.state.uri !== uri) {
      this.updateMenuKey();
    }

    if (this.props.account !== prevProps.account) {
      const requiredEnableMfa = Setting.isRequiredEnableMfa(this.props.account, this.props.account?.organization);
      this.setState({
        requiredEnableMfa: requiredEnableMfa,
      });

      if (requiredEnableMfa === true) {
        const mfaType = Setting.getMfaItemsByRules(this.props.account, this.props.account?.organization, [Setting.MfaRuleRequired])
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
      this.setState({selectedMenuKeyselectedMenuKey: "/home"});
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
    if (themes.includes("dark")) {
      return `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256_dark.png`;
    } else {
      return `${Setting.StaticBaseUrl}/img/casdoor-logo_1185x256.png`;
    }
  }

  setLanguage(account) {
    const language = account?.language;
    if (language !== null && language !== "" && language !== i18next.language) {
      Setting.setLanguage(language);
    }
  }

  setTheme = (theme, initThemeAlgorithm) => {
    this.props.setTheme(theme);
    if (initThemeAlgorithm) {
      this.props.setLogo(theme);
      this.props.setThemeAlgorithm(theme);
    }
  };

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
        if (res.status === "ok") {
          account = res.data;
          account.organization = res.data2;

          this.setLanguage(account);
          this.setTheme(Setting.getThemeData(account.organization), Conf.InitThemeAlgorithm);
        } else {
          if (res.data !== "Please login first") {
            Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
          }
        }

        this.props.setAccount(account);
      });
  }

  onUpdateAccount(account) {
    // this.setState({
    //   account: account,
    // });
    this.props.setAccount(account);
  }

  renderContent() {

  }

  renderFooter() {
    return (
      <React.Fragment>
        {!this.props.account ? null : <div style={{display: "none"}} id="CasdoorApplicationName" value={this.props.account.signupApplication} />}
        <Footer id="footer" style={
          {
            textAlign: "center",
          }
        }>
          {
            Conf.CustomFooter !== null ? Conf.CustomFooter : (
              <React.Fragment>
                Powered by <a target="_blank" href="https://casdoor.org" rel="noreferrer"><img style={{paddingBottom: "3px"}} height={"20px"} alt={"Casdoor"} src={this.props.logo} /></a>
              </React.Fragment>
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
            <a className="custom-link" style={{float: "right", marginTop: "2px"}} target="_blank" rel="noreferrer" href={"https://ai.casbin.com"}>
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
        visible={this.state.isAiAssistantOpen}
      >
        <iframe id="iframeHelper" title={"iframeHelper"} src={"https://ai.casbin.com/?isRaw=1"} width="100%" height="100%" scrolling="no" frameBorder="no" />
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
        window.location.pathname.startsWith("/qrcode") ;
  }

  onClose = () => {
    this.setState({
      menuVisible: false,
    });
  };

  showMenu = () => {
    this.setState({
      menuVisible: true,
    });
  };

  setExpiredAndSubmitted() {
    this.setState({
      expired: false,
      submitted: false,
    });
  }

  setRequireMfaFalse() {
    this.setState({
      requiredEnableMfa: false,
    });
  }

  renderPage() {
    if (this.isDoorPages()) {
      return (
        <Layout id="parent-area">
          <Content style={{display: "flex", justifyContent: "center"}}>
            {
              this.isEntryPages() ?
                <EntryPage
                  account={this.props.account}
                  theme={this.props.themeData}
                  onLoginSuccess={(redirectUrl) => {
                    if (redirectUrl) {
                      localStorage.setItem("mfaRedirectUrl", redirectUrl);
                    }
                    this.getAccount();
                  }}
                  onUpdateAccount={(account) => this.onUpdateAccount(account)}
                  updataThemeData={this.setTheme}
                  t={this.props.t}
                /> :
                <Switch>
                  <Route exact path="/callback" component={AuthCallback} />
                  <Route exact path="/callback/saml" component={SamlCallback} />
                  <Route path="" render={() => <Result status="404" title="404 NOT FOUND" subTitle={i18next.t("general:Sorry, the page you visited does not exist.")}
                    extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>} />} />
                </Switch>
            }
          </Content>
          {
            this.renderFooter()
          }
          {
            this.renderAiAssistant()
          }
        </Layout>
      );
    }

    return (
      <React.Fragment>
        {/* { */}
        {/*   this.renderBanner() */}
        {/* } */}
        <FloatButton.BackTop />
        <CustomGithubCorner />
        <Layout id="parent-area">
          {
            <AppContent
              menuVisible={this.state.menuVisible}
              uri={this.state.uri}
              requiredEnableMfa={this.state.requiredEnableMfa}
              selectedMenuKey={this.state.selectedMenuKey}
              setExpiredAndSubmitted={() => {this.setExpiredAndSubmitted();}}
              setRequireMfaFalse={() => this.setRequireMfaFalse()}
              onClose = {() => {this.onClose();}}
              showMenu = {() => this.showMenu()}
              OpenAiAssistant = {() => {
                this.setState({
                  isAiAssistantOpen: true,
                });
              }}
              {...this.props}
            />
          }
          {
            this.renderFooter()
          }
          {
            this.renderAiAssistant()
          }
        </Layout>
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
        {(this.props.account === undefined || this.props.account === null) ?
          <Helmet>
            <link rel="icon" href={"https://cdn.casdoor.com/static/favicon.png"} />
          </Helmet> :
          <Helmet>
            <title>{this.props.account.organization?.displayName}</title>
            <link rel="icon" href={this.props.account.organization?.favicon} />
          </Helmet>
        }
        <ConfigProvider theme={{
          token: {
            colorPrimary: this.props.themeData.colorPrimary,
            colorInfo: this.props.themeData.colorPrimary,
            borderRadius: this.props.themeData.borderRadius,
          },
          algorithm: Setting.getAlgorithm(this.props.themeAlgorithm),
        }}>
          <StyleProvider hashPriority="high" transformers={[legacyLogicalPropertiesTransformer]}>
            <Suspense fallback={<div></div>}>
              {
                this.renderPage()
              }
            </Suspense>
          </StyleProvider>
        </ConfigProvider>
      </React.Fragment>
    );
  }
}

export default connect(mapStateToProps, {setAccount, setTheme, setLogo, setThemeAlgorithm})(withRouter(withTranslation()(App)));
