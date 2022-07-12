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

import React, {Component} from "react";
import "./App.less";
import {Helmet} from "react-helmet";
import * as Setting from "./Setting";
import {DownOutlined, LogoutOutlined, SettingOutlined} from "@ant-design/icons";
import {Avatar, BackTop, Dropdown, Layout, Menu, Card, Result, Button} from "antd";
import {Link, Redirect, Route, Switch, withRouter} from "react-router-dom";
import OrganizationListPage from "./OrganizationListPage";
import OrganizationEditPage from "./OrganizationEditPage";
import UserListPage from "./UserListPage";
import UserEditPage from "./UserEditPage";
import RoleListPage from "./RoleListPage";
import RoleEditPage from "./RoleEditPage";
import PermissionListPage from "./PermissionListPage";
import PermissionEditPage from "./PermissionEditPage";
import ProviderListPage from "./ProviderListPage";
import ProviderEditPage from "./ProviderEditPage";
import ApplicationListPage from "./ApplicationListPage";
import ApplicationEditPage from "./ApplicationEditPage";
import ResourceListPage from "./ResourceListPage";
import LdapEditPage from "./LdapEditPage";
import LdapSyncPage from "./LdapSyncPage";
import TokenListPage from "./TokenListPage";
import TokenEditPage from "./TokenEditPage";
import RecordListPage from "./RecordListPage";
import WebhookListPage from "./WebhookListPage";
import WebhookEditPage from "./WebhookEditPage";
import SyncerListPage from "./SyncerListPage";
import SyncerEditPage from "./SyncerEditPage";
import CertListPage from "./CertListPage";
import CertEditPage from "./CertEditPage";
import ProductListPage from "./ProductListPage";
import ProductEditPage from "./ProductEditPage";
import ProductBuyPage from "./ProductBuyPage";
import PaymentListPage from "./PaymentListPage";
import PaymentEditPage from "./PaymentEditPage";
import PaymentResultPage from "./PaymentResultPage";
import AccountPage from "./account/AccountPage";
import HomePage from "./basic/HomePage";
import CustomGithubCorner from "./CustomGithubCorner";
import * as Conf from "./Conf";

import * as Auth from "./auth/Auth";
import SignupPage from "./auth/SignupPage";
import ResultPage from "./auth/ResultPage";
import LoginPage from "./auth/LoginPage";
import SelfLoginPage from "./auth/SelfLoginPage";
import SelfForgetPage from "./auth/SelfForgetPage";
import ForgetPage from "./auth/ForgetPage";
import * as AuthBackend from "./auth/AuthBackend";
import AuthCallback from "./auth/AuthCallback";
import SelectLanguageBox from "./SelectLanguageBox";
import i18next from "i18next";
import PromptPage from "./auth/PromptPage";
import OdicDiscoveryPage from "./auth/OidcDiscoveryPage";
import SamlCallback from "./auth/SamlCallback";
import CasLogout from "./auth/CasLogout";
import ModelListPage from "./ModelListPage";
import ModelEditPage from "./ModelEditPage";

const {Header, Footer} = Layout;

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      selectedMenuKey: 0,
      account: undefined,
      uri: null,
    };

    Setting.initServerUrl();
    Auth.initAuthWithConfig({
      serverUrl: Setting.ServerUrl,
      appName: "app-built-in", // the application name of Casdoor itself, do not change it
    });
  }

  UNSAFE_componentWillMount() {
    this.updateMenuKey();
    this.getAccount();
  }

  componentDidUpdate() {
    // eslint-disable-next-line no-restricted-globals
    const uri = location.pathname;
    if (this.state.uri !== uri) {
      this.updateMenuKey();
    }
  }

  updateMenuKey() {
    // eslint-disable-next-line no-restricted-globals
    const uri = location.pathname;
    this.setState({
      uri: uri,
    });
    if (uri === "/") {
      this.setState({selectedMenuKey: "/"});
    } else if (uri.includes("/organizations")) {
      this.setState({selectedMenuKey: "/organizations"});
    } else if (uri.includes("/users")) {
      this.setState({selectedMenuKey: "/users"});
    } else if (uri.includes("/roles")) {
      this.setState({selectedMenuKey: "/roles"});
    } else if (uri.includes("/permissions")) {
      this.setState({selectedMenuKey: "/permissions"});
    } else if (uri.includes("/models")) {
      this.setState({selectedMenuKey: "/models"});
    } else if (uri.includes("/providers")) {
      this.setState({selectedMenuKey: "/providers"});
    } else if (uri.includes("/applications")) {
      this.setState({selectedMenuKey: "/applications"});
    } else if (uri.includes("/resources")) {
      this.setState({selectedMenuKey: "/resources"});
    } else if (uri.includes("/tokens")) {
      this.setState({selectedMenuKey: "/tokens"});
    } else if (uri.includes("/records")) {
      this.setState({selectedMenuKey: "/records"});
    } else if (uri.includes("/webhooks")) {
      this.setState({selectedMenuKey: "/webhooks"});
    } else if (uri.includes("/syncers")) {
      this.setState({selectedMenuKey: "/syncers"});
    } else if (uri.includes("/certs")) {
      this.setState({selectedMenuKey: "/certs"});
    } else if (uri.includes("/products")) {
      this.setState({selectedMenuKey: "/products"});
    } else if (uri.includes("/payments")) {
      this.setState({selectedMenuKey: "/payments"});
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

  setLanguage(account) {
    let language = account?.language;
    if (language !== "" && language !== i18next.language) {
      Setting.setLanguage(language);
    }
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
        if (res.status === "ok") {
          account = res.data;
          account.organization = res.data2;

          this.setLanguage(account);
        } else {
          if (res.msg !== "Please sign in first") {
            Setting.showMessage("error", `Failed to sign in: ${res.msg}`);
          }
        }

        this.setState({
          account: account,
        });
      });
  }

  logout() {
    this.setState({
      expired: false,
      submitted: false,
    });

    AuthBackend.logout()
      .then((res) => {
        if (res.status === "ok") {
          this.setState({
            account: null
          });

          Setting.showMessage("success", "Logged out successfully");
          let redirectUri = res.data2;
          if (redirectUri !== null && redirectUri !== undefined && redirectUri !== "") {
            Setting.goToLink(redirectUri);
          }else{
            Setting.goToLinkSoft(this, "/");
          }
        } else {
          Setting.showMessage("error", `Failed to log out: ${res.msg}`);
        }
      });
  }

  onUpdateAccount(account) {
    this.setState({
      account: account
    });
  }

  handleRightDropdownClick(e) {
    if (e.key === "/account") {
      this.props.history.push("/account");
    } else if (e.key === "/logout") {
      this.logout();
    }
  }

  renderAvatar() {
    if (this.state.account.avatar === "") {
      return (
        <Avatar style={{backgroundColor: Setting.getAvatarColor(this.state.account.name), verticalAlign: "middle"}} size="large">
          {Setting.getShortName(this.state.account.name)}
        </Avatar>
      );
    } else {
      return (
        <Avatar src={this.state.account.avatar} style={{verticalAlign: "middle"}} size="large">
          {Setting.getShortName(this.state.account.name)}
        </Avatar>
      );
    }
  }

  renderRightDropdown() {
    const menu = (
      <Menu onClick={this.handleRightDropdownClick.bind(this)}>
        <Menu.Item key="/account">
          <SettingOutlined />
          &nbsp;
          {i18next.t("account:My Account")}
        </Menu.Item>
        <Menu.Item key="/logout">
          <LogoutOutlined />
          &nbsp;
          {i18next.t("account:Logout")}
        </Menu.Item>
      </Menu>
    );

    return (
      <Dropdown key="/rightDropDown" overlay={menu} className="rightDropDown">
        <div className="ant-dropdown-link" style={{float: "right", cursor: "pointer"}}>
          &nbsp;
          &nbsp;
          {
            this.renderAvatar()
          }
          &nbsp;
          &nbsp;
          {Setting.isMobile() ? null : Setting.getShortName(this.state.account.displayName)} &nbsp; <DownOutlined />
          &nbsp;
          &nbsp;
          &nbsp;
        </div>
      </Dropdown>
    );
  }

  renderAccount() {
    let res = [];

    if (this.state.account === undefined) {
      return null;
    } else if (this.state.account === null) {
      // res.push(
      //   <Menu.Item key="/signup" style={{float: 'right', marginRight: '20px'}}>
      //     <Link to="/signup">
      //       {i18next.t("account:Sign Up")}
      //     </Link>
      //   </Menu.Item>
      // );
      // res.push(
      //   <Menu.Item key="/login" style={{float: 'right'}}>
      //     <Link to="/login">
      //       {i18next.t("account:Login")}
      //     </Link>
      //   </Menu.Item>
      // );
    } else {
      res.push(this.renderRightDropdown());
    }

    return res;
  }

  renderMenu() {
    let res = [];

    if (this.state.account === null || this.state.account === undefined) {
      return [];
    }

    res.push(
      <Menu.Item key="/">
        <Link to="/">
          {i18next.t("general:Home")}
        </Link>
      </Menu.Item>
    );

    if (Setting.isAdminUser(this.state.account)) {
      res.push(
        <Menu.Item key="/organizations">
          <Link to="/organizations">
            {i18next.t("general:Organizations")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/users">
          <Link to="/users">
            {i18next.t("general:Users")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/roles">
          <Link to="/roles">
            {i18next.t("general:Roles")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/permissions">
          <Link to="/permissions">
            {i18next.t("general:Permissions")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/models">
          <Link to="/models">
            {i18next.t("general:Models")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/providers">
          <Link to="/providers">
            {i18next.t("general:Providers")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/applications">
          <Link to="/applications">
            {i18next.t("general:Applications")}
          </Link>
        </Menu.Item>
      );
    }

    if (Setting.isAdminUser(this.state.account)) {
      res.push(
        <Menu.Item key="/resources">
          <Link to="/resources">
            {i18next.t("general:Resources")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/tokens">
          <Link to="/tokens">
            {i18next.t("general:Tokens")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/records">
          <Link to="/records">
            {i18next.t("general:Records")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/webhooks">
          <Link to="/webhooks">
            {i18next.t("general:Webhooks")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/syncers">
          <Link to="/syncers">
            {i18next.t("general:Syncers")}
          </Link>
        </Menu.Item>
      );
      res.push(
        <Menu.Item key="/certs">
          <Link to="/certs">
            {i18next.t("general:Certs")}
          </Link>
        </Menu.Item>
      );

      if (Conf.EnableExtraPages) {
        res.push(
          <Menu.Item key="/products">
            <Link to="/products">
              {i18next.t("general:Products")}
            </Link>
          </Menu.Item>
        );
        res.push(
          <Menu.Item key="/payments">
            <Link to="/payments">
              {i18next.t("general:Payments")}
            </Link>
          </Menu.Item>
        );
      }

      res.push(
        <Menu.Item key="/swagger">
          <a target="_blank" rel="noreferrer" href={Setting.isLocalhost() ? `${Setting.ServerUrl}/swagger` : "/swagger"}>
            {i18next.t("general:Swagger")}
          </a>
        </Menu.Item>
      );
    }

    return res;
  }

  renderHomeIfLoggedIn(component) {
    if (this.state.account !== null && this.state.account !== undefined) {
      return <Redirect to="/" />;
    } else {
      return component;
    }
  }

  renderLoginIfNotLoggedIn(component) {
    if (this.state.account === null) {
      sessionStorage.setItem("from", window.location.pathname);
      return <Redirect to="/login" />;
    } else if (this.state.account === undefined) {
      return null;
    } else {
      return component;
    }
  }

  isStartPages() {
    return window.location.pathname.startsWith("/login") ||
      window.location.pathname.startsWith("/signup") ||
      window.location.pathname === "/";
  }

  renderRouter() {
    return(
      <div>
        <Switch>
          <Route exact path="/result" render={(props) => this.renderHomeIfLoggedIn(<ResultPage {...props} />)} />
          <Route exact path="/result/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<ResultPage {...props} />)} />
          <Route exact path="/" render={(props) => this.renderLoginIfNotLoggedIn(<HomePage account={this.state.account} {...props} />)} />
          <Route exact path="/account" render={(props) => this.renderLoginIfNotLoggedIn(<AccountPage account={this.state.account} {...props} />)} />
          <Route exact path="/organizations" render={(props) => this.renderLoginIfNotLoggedIn(<OrganizationListPage account={this.state.account} {...props} />)} />
          <Route exact path="/organizations/:organizationName" render={(props) => this.renderLoginIfNotLoggedIn(<OrganizationEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/organizations/:organizationName/users" render={(props) => this.renderLoginIfNotLoggedIn(<UserListPage account={this.state.account} {...props} />)} />
          <Route exact path="/users" render={(props) => this.renderLoginIfNotLoggedIn(<UserListPage account={this.state.account} {...props} />)} />
          <Route exact path="/users/:organizationName/:userName" render={(props) => <UserEditPage account={this.state.account} {...props} />} />
          <Route exact path="/roles" render={(props) => this.renderLoginIfNotLoggedIn(<RoleListPage account={this.state.account} {...props} />)} />
          <Route exact path="/roles/:organizationName/:roleName" render={(props) => this.renderLoginIfNotLoggedIn(<RoleEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/permissions" render={(props) => this.renderLoginIfNotLoggedIn(<PermissionListPage account={this.state.account} {...props} />)} />
          <Route exact path="/permissions/:organizationName/:permissionName" render={(props) => this.renderLoginIfNotLoggedIn(<PermissionEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/models" render={(props) => this.renderLoginIfNotLoggedIn(<ModelListPage account={this.state.account} {...props} />)} />
          <Route exact path="/models/:organizationName/:modelName" render={(props) => this.renderLoginIfNotLoggedIn(<ModelEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/providers" render={(props) => this.renderLoginIfNotLoggedIn(<ProviderListPage account={this.state.account} {...props} />)} />
          <Route exact path="/providers/:providerName" render={(props) => this.renderLoginIfNotLoggedIn(<ProviderEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/applications" render={(props) => this.renderLoginIfNotLoggedIn(<ApplicationListPage account={this.state.account} {...props} />)} />
          <Route exact path="/applications/:applicationName" render={(props) => this.renderLoginIfNotLoggedIn(<ApplicationEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/resources" render={(props) => this.renderLoginIfNotLoggedIn(<ResourceListPage account={this.state.account} {...props} />)} />
          {/* <Route exact path="/resources/:resourceName" render={(props) => this.renderLoginIfNotLoggedIn(<ResourceEditPage account={this.state.account} {...props} />)}/>*/}
          <Route exact path="/ldap/:ldapId" render={(props) => this.renderLoginIfNotLoggedIn(<LdapEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/ldap/sync/:ldapId" render={(props) => this.renderLoginIfNotLoggedIn(<LdapSyncPage account={this.state.account} {...props} />)} />
          <Route exact path="/tokens" render={(props) => this.renderLoginIfNotLoggedIn(<TokenListPage account={this.state.account} {...props} />)} />
          <Route exact path="/tokens/:tokenName" render={(props) => this.renderLoginIfNotLoggedIn(<TokenEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/webhooks" render={(props) => this.renderLoginIfNotLoggedIn(<WebhookListPage account={this.state.account} {...props} />)} />
          <Route exact path="/webhooks/:webhookName" render={(props) => this.renderLoginIfNotLoggedIn(<WebhookEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/syncers" render={(props) => this.renderLoginIfNotLoggedIn(<SyncerListPage account={this.state.account} {...props} />)} />
          <Route exact path="/syncers/:syncerName" render={(props) => this.renderLoginIfNotLoggedIn(<SyncerEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/certs" render={(props) => this.renderLoginIfNotLoggedIn(<CertListPage account={this.state.account} {...props} />)} />
          <Route exact path="/certs/:certName" render={(props) => this.renderLoginIfNotLoggedIn(<CertEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/products" render={(props) => this.renderLoginIfNotLoggedIn(<ProductListPage account={this.state.account} {...props} />)} />
          <Route exact path="/products/:productName" render={(props) => this.renderLoginIfNotLoggedIn(<ProductEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/products/:productName/buy" render={(props) => this.renderLoginIfNotLoggedIn(<ProductBuyPage account={this.state.account} {...props} />)} />
          <Route exact path="/payments" render={(props) => this.renderLoginIfNotLoggedIn(<PaymentListPage account={this.state.account} {...props} />)} />
          <Route exact path="/payments/:paymentName" render={(props) => this.renderLoginIfNotLoggedIn(<PaymentEditPage account={this.state.account} {...props} />)} />
          <Route exact path="/payments/:paymentName/result" render={(props) => this.renderLoginIfNotLoggedIn(<PaymentResultPage account={this.state.account} {...props} />)} />
          <Route exact path="/records" render={(props) => this.renderLoginIfNotLoggedIn(<RecordListPage account={this.state.account} {...props} />)} />
          <Route exact path="/.well-known/openid-configuration" render={(props) => <OdicDiscoveryPage />} />
          <Route path="" render={() => <Result status="404" title="404 NOT FOUND" subTitle={i18next.t("general:Sorry, the page you visited does not exist.")}
            extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>} />} />
        </Switch>
      </div>
    );
  }

  renderContent() {
    if (!Setting.isMobile()) {
      return (
        <div style={{display: "flex", flex: "auto", width:"100%", flexDirection: "column"}}>
          <Layout style={{display: "flex", alignItems: "stretch"}}>
            <Header style={{padding: "0", marginBottom: "3px"}}>
              {
                Setting.isMobile() ? null : (
                  <Link to={"/"}>
                    <div className="logo" />
                  </Link>
                )
              }
              <div>
                <Menu
                  // theme="dark"
                  mode={(Setting.isMobile() && this.isStartPages()) ? "inline" : "horizontal"}
                  selectedKeys={[`${this.state.selectedMenuKey}`]}
                  style={{lineHeight: "64px", width: "80%", position: "absolute"}}
                >
                  {
                    this.renderMenu()
                  }
                </Menu>
                {
                  this.renderAccount()
                }
                <SelectLanguageBox />
              </div>
            </Header>
            <Layout style={{backgroundColor: "#f5f5f5", alignItems: "stretch"}}>
              <Card className="content-warp-card">
                {
                  this.renderRouter()
                }
              </Card>
            </Layout>
          </Layout>
        </div>
      );
    } else {
      return(
        <div>
          <Header style={{padding: "0", marginBottom: "3px"}}>
            {
              Setting.isMobile() ? null : (
                <Link to={"/"}>
                  <div className="logo" />
                </Link>
              )
            }
            <Menu
            // theme="dark"
              mode={(Setting.isMobile() && this.isStartPages()) ? "inline" : "horizontal"}
              selectedKeys={[`${this.state.selectedMenuKey}`]}
              style={{lineHeight: "64px"}}
            >
              {
                this.renderMenu()
              }
              <div style = {{float: "right"}}>
                {
                  this.renderAccount()
                }
                <SelectLanguageBox />
              </div>
            </Menu>
          </Header>
          {
            this.renderRouter()
          }
        </div>
      );
    }
  }

  renderFooter() {
    // How to keep your footer where it belongs ?
    // https://www.freecodecamp.org/neyarnws/how-to-keep-your-footer-where-it-belongs-59c6aa05c59c/

    return (
      <Footer id="footer" style={
        {
          borderTop: "1px solid #e8e8e8",
          backgroundColor: "white",
          textAlign: "center",
        }
      }>
        Made with <span style={{color: "rgb(255, 255, 255)"}}>❤️</span> by <a style={{fontWeight: "bold", color: "black"}} target="_blank" href="https://casdoor.org" rel="noreferrer">Casdoor</a>
      </Footer>
    );
  }

  isDoorPages() {
    return window.location.pathname.startsWith("/signup") ||
      window.location.pathname.startsWith("/login") ||
      window.location.pathname.startsWith("/callback") ||
      window.location.pathname.startsWith("/prompt") ||
      window.location.pathname.startsWith("/forget") ||
      window.location.pathname.startsWith("/cas");
  }

  renderPage() {
    if (this.isDoorPages()) {
      return (
        <div>
          <Switch>
            <Route exact path="/signup" render={(props) => this.renderHomeIfLoggedIn(<SignupPage account={this.state.account} {...props} />)} />
            <Route exact path="/signup/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<SignupPage account={this.state.account} {...props} onUpdateAccount={(account) => {this.onUpdateAccount(account);}} />)} />
            <Route exact path="/login" render={(props) => this.renderHomeIfLoggedIn(<SelfLoginPage account={this.state.account} {...props} />)} />
            <Route exact path="/signup/oauth/authorize" render={(props) => <LoginPage account={this.state.account} type={"code"} mode={"signup"} {...props} onUpdateAccount={(account) => {this.onUpdateAccount(account);}} />} />
            <Route exact path="/login/oauth/authorize" render={(props) => <LoginPage account={this.state.account} type={"code"} mode={"signin"} {...props} onUpdateAccount={(account) => {this.onUpdateAccount(account);}} />} />
            <Route exact path="/login/saml/authorize/:owner/:applicationName" render={(props) => <LoginPage account={this.state.account} type={"saml"} mode={"signin"} {...props} onUpdateAccount={(account) => {this.onUpdateAccount(account);}} />} />
            <Route exact path="/cas/:owner/:casApplicationName/logout" render={(props) => this.renderHomeIfLoggedIn(<CasLogout clearAccount={() => this.setState({account: null})} {...props} />)} />
            <Route exact path="/cas/:owner/:casApplicationName/login" render={(props) => {return (<LoginPage type={"cas"} mode={"signup"} account={this.state.account} {...props} />);}} />
            <Route exact path="/callback" component={AuthCallback} />
            <Route exact path="/callback/saml" component={SamlCallback} />
            <Route exact path="/forget" render={(props) => this.renderHomeIfLoggedIn(<SelfForgetPage {...props} />)} />
            <Route exact path="/forget/:applicationName" render={(props) => this.renderHomeIfLoggedIn(<ForgetPage {...props} />)} />
            <Route exact path="/prompt" render={(props) => this.renderLoginIfNotLoggedIn(<PromptPage account={this.state.account} {...props} />)} />
            <Route exact path="/prompt/:applicationName" render={(props) => this.renderLoginIfNotLoggedIn(<PromptPage account={this.state.account} onUpdateAccount={(account) => {this.onUpdateAccount(account);}} {...props} />)} />
            <Route path="" render={() => <Result status="404" title="404 NOT FOUND" subTitle={i18next.t("general:Sorry, the page you visited does not exist.")}
              extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>} />} />
          </Switch>
          {
            this.renderFooter()
          }
        </div>
      );
    }

    return (
      <div id="parent-area">
        <BackTop />
        <CustomGithubCorner />
        <div id="content-wrap" style={{flexDirection: "column"}}>
          {
            this.renderContent()
          }
        </div>
        {
          this.renderFooter()
        }
      </div>
    );
  }

  render() {
    if (this.state.account === undefined || this.state.account === null) {
      return (
        <React.Fragment>
          <Helmet>
            <link rel="icon" href={"https://cdn.casdoor.com/static/favicon.png"} />
          </Helmet>
          {
            this.renderPage()
          }
        </React.Fragment>
      );
    }

    const organization = this.state.account.organization;
    return (
      <React.Fragment>
        <Helmet>
          <title>{organization.displayName}</title>
          <link rel="icon" href={organization.favicon} />
        </Helmet>
        {
          this.renderPage()
        }
      </React.Fragment>
    );
  }
}

export default withRouter(App);
