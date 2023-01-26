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
import {BarsOutlined, DownOutlined, LogoutOutlined, SettingOutlined} from "@ant-design/icons";
import {Avatar, Button, Card, ConfigProvider, Drawer, Dropdown, FloatButton, Layout, Menu, Result, theme} from "antd";
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
import EntryPage from "./EntryPage";
import ResultPage from "./auth/ResultPage";
import * as AuthBackend from "./auth/AuthBackend";
import AuthCallback from "./auth/AuthCallback";
import SelectLanguageBox from "./SelectLanguageBox";
import i18next from "i18next";
import OdicDiscoveryPage from "./auth/OidcDiscoveryPage";
import SamlCallback from "./auth/SamlCallback";
import ModelListPage from "./ModelListPage";
import ModelEditPage from "./ModelEditPage";
import SystemInfo from "./SystemInfo";
import AdapterListPage from "./AdapterListPage";
import AdapterEditPage from "./AdapterEditPage";
import {withTranslation} from "react-i18next";
import SelectThemeBox from "./SelectThemeBox";
import SessionListPage from "./SessionListPage";

const {Header, Footer, Content} = Layout;

class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      selectedMenuKey: 0,
      account: undefined,
      uri: null,
      menuVisible: false,
      themeAlgorithm: null,
      logo: null,
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

  componentDidMount() {
    this.setState({
      themeAlgorithm: this.getTheme(),
      logo: Setting.getLogo(localStorage.getItem("theme")),
    });

    addEventListener("changeTheme", (e) => {
      this.setState({
        themeAlgorithm: this.getTheme(),
        logo: Setting.getLogo(localStorage.getItem("theme")),
      });
    });
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
    } else if (uri.includes("/adapters")) {
      this.setState({selectedMenuKey: "/adapters"});
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
    } else if (uri.includes("/sysinfo")) {
      this.setState({selectedMenuKey: "/sysinfo"});
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

  getTheme() {
    return localStorage.getItem("theme") === null ? theme.defaultAlgorithm :
      Setting.Themes.find(t => t.key === localStorage.getItem("theme")).theme;
  }

  setLanguage(account) {
    const language = account?.language;
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
          if (res.data !== "Please login first") {
            Setting.showMessage("error", `${i18next.t("application:Failed to sign in")}: ${res.msg}`);
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
          const owner = this.state.account.owner;

          this.setState({
            account: null,
          });

          Setting.showMessage("success", i18next.t("application:Logged out successfully"));
          const redirectUri = res.data2;
          if (redirectUri !== null && redirectUri !== undefined && redirectUri !== "") {
            Setting.goToLink(redirectUri);
          } else if (owner !== "built-in") {
            Setting.goToLink(`${window.location.origin}/login/${owner}`);
          } else {
            Setting.goToLinkSoft(this, "/");
          }
        } else {
          Setting.showMessage("error", `Failed to log out: ${res.msg}`);
        }
      });
  }

  onUpdateAccount(account) {
    this.setState({
      account: account,
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
    const items = [];
    items.push(Setting.getItem(<><SettingOutlined />&nbsp;&nbsp;{i18next.t("account:My Account")}</>,
      "/account"
    ));
    items.push(Setting.getItem(<><LogoutOutlined />&nbsp;&nbsp;{i18next.t("account:Logout")}</>,
      "/logout"));
    const onClick = this.handleRightDropdownClick.bind(this);

    return (
      <Dropdown key="/rightDropDown" menu={{items, onClick}} className="rightDropDown">
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
    const res = [];

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
    const res = [];

    if (this.state.account === null || this.state.account === undefined) {
      return [];
    }

    res.push(Setting.getItem(<Link to="/">{i18next.t("general:Home")}</Link>, "/"));

    if (Setting.isAdminUser(this.state.account)) {
      res.push(Setting.getItem(<Link to="/organizations">{i18next.t("general:Organizations")}</Link>,
        "/organizations"));
    }

    if (Setting.isLocalAdminUser(this.state.account)) {
      res.push(Setting.getItem(<Link to="/users">{i18next.t("general:Users")}</Link>,
        "/users"
      ));

      res.push(Setting.getItem(<Link to="/roles">{i18next.t("general:Roles")}</Link>,
        "/roles"
      ));

      res.push(Setting.getItem(<Link to="/permissions">{i18next.t("general:Permissions")}</Link>,
        "/permissions"
      ));
    }

    if (Setting.isAdminUser(this.state.account)) {
      res.push(Setting.getItem(<Link to="/models">{i18next.t("general:Models")}</Link>,
        "/models"
      ));

      res.push(Setting.getItem(<Link to="/adapters">{i18next.t("general:Adapters")}</Link>,
        "/adapters"
      ));
    }

    if (Setting.isLocalAdminUser(this.state.account)) {
      res.push(Setting.getItem(<Link to="/applications">{i18next.t("general:Applications")}</Link>,
        "/applications"
      ));

      res.push(Setting.getItem(<Link to="/providers">{i18next.t("general:Providers")}</Link>,
        "/providers"
      ));

      res.push(Setting.getItem(<Link to="/resources">{i18next.t("general:Resources")}</Link>,
        "/resources"
      ));

      res.push(Setting.getItem(<Link to="/records">{i18next.t("general:Records")}</Link>,
        "/records"
      ));
    }

    if (Setting.isAdminUser(this.state.account)) {
      res.push(Setting.getItem(<Link to="/tokens">{i18next.t("general:Tokens")}</Link>,
        "/tokens"
      ));

      res.push(Setting.getItem(<Link to="/sessions">{i18next.t("general:Sessions")}</Link>,
        "/sessions"
      ));

      res.push(Setting.getItem(<Link to="/webhooks">{i18next.t("general:Webhooks")}</Link>,
        "/webhooks"
      ));

      res.push(Setting.getItem(<Link to="/syncers">{i18next.t("general:Syncers")}</Link>,
        "/syncers"
      ));

      res.push(Setting.getItem(<Link to="/certs">{i18next.t("general:Certs")}</Link>,
        "/certs"
      ));

      if (Conf.EnableExtraPages) {
        res.push(Setting.getItem(<Link to="/products">{i18next.t("general:Products")}</Link>,
          "/products"
        ));

        res.push(Setting.getItem(<Link to="/payments">{i18next.t("general:Payments")}</Link>,
          "/payments"
        ));

        res.push(Setting.getItem(<Link to="/sysinfo">{i18next.t("general:SysInfo")}</Link>,
          "/sysinfo"
        ));
      }
      res.push(Setting.getItem(<a target="_blank" rel="noreferrer"
        href={Setting.isLocalhost() ? `${Setting.ServerUrl}/swagger` : "/swagger"}>{i18next.t("general:Swagger")}</a>,
      "/swagger"
      ));
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
    return (
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
        <Route exact path="/adapters" render={(props) => this.renderLoginIfNotLoggedIn(<AdapterListPage account={this.state.account} {...props} />)} />
        <Route exact path="/adapters/:organizationName/:adapterName" render={(props) => this.renderLoginIfNotLoggedIn(<AdapterEditPage account={this.state.account} {...props} />)} />
        <Route exact path="/providers" render={(props) => this.renderLoginIfNotLoggedIn(<ProviderListPage account={this.state.account} {...props} />)} />
        <Route exact path="/providers/:organizationName/:providerName" render={(props) => this.renderLoginIfNotLoggedIn(<ProviderEditPage account={this.state.account} {...props} />)} />
        <Route exact path="/applications" render={(props) => this.renderLoginIfNotLoggedIn(<ApplicationListPage account={this.state.account} {...props} />)} />
        <Route exact path="/applications/:organizationName/:applicationName" render={(props) => this.renderLoginIfNotLoggedIn(<ApplicationEditPage account={this.state.account} {...props} />)} />
        <Route exact path="/resources" render={(props) => this.renderLoginIfNotLoggedIn(<ResourceListPage account={this.state.account} {...props} />)} />
        {/* <Route exact path="/resources/:resourceName" render={(props) => this.renderLoginIfNotLoggedIn(<ResourceEditPage account={this.state.account} {...props} />)}/>*/}
        <Route exact path="/ldap/:ldapId" render={(props) => this.renderLoginIfNotLoggedIn(<LdapEditPage account={this.state.account} {...props} />)} />
        <Route exact path="/ldap/sync/:ldapId" render={(props) => this.renderLoginIfNotLoggedIn(<LdapSyncPage account={this.state.account} {...props} />)} />
        <Route exact path="/tokens" render={(props) => this.renderLoginIfNotLoggedIn(<TokenListPage account={this.state.account} {...props} />)} />
        <Route exact path="/sessions" render={(props) => this.renderLoginIfNotLoggedIn(<SessionListPage account={this.state.account} {...props} />)} />
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
        <Route exact path="/sysinfo" render={(props) => this.renderLoginIfNotLoggedIn(<SystemInfo account={this.state.account} {...props} />)} />
        <Route path="" render={() => <Result status="404" title="404 NOT FOUND" subTitle={i18next.t("general:Sorry, the page you visited does not exist.")}
          extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>} />} />
      </Switch>
    );
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

  renderContent() {
    return (
      <Layout id="parent-area">
        {/* https://github.com/ant-design/ant-design/issues/40394 ant design bug. If it will be fixed, we can delete the code for control the color of Header*/}
        <Header style={{padding: "0", marginBottom: "3px", backgroundColor: this.state.themeAlgorithm === theme.darkAlgorithm ? "black" : "white"}}>
          {Setting.isMobile() ? null : (
            <Link to={"/"}>
              <div className="logo" style={{background: `url(${this.state.logo})`}} />
            </Link>
          )}
          {Setting.isMobile() ?
            <React.Fragment>
              <Drawer title={i18next.t("general:Close")} placement="left" visible={this.state.menuVisible} onClose={this.onClose}>
                <Menu
                  items={this.renderMenu()}
                  mode={"inline"}
                  selectedKeys={[`${this.state.selectedMenuKey}`]}
                  style={{lineHeight: "64px"}}
                  onClick={this.onClose}
                >
                </Menu>
              </Drawer>
              <Button icon={<BarsOutlined />} onClick={this.showMenu} type="text">
                {i18next.t("general:Menu")}
              </Button>
            </React.Fragment>
            :
            <Menu
              items={this.renderMenu()}
              mode={"horizontal"}
              selectedKeys={[`${this.state.selectedMenuKey}`]}
              style={{position: "absolute", left: "145px"}}
            />
          }
          {this.renderAccount()}
          {this.state.account &&
              <React.Fragment>
                <SelectThemeBox themes={this.state.account.organization.themes} />
                <SelectLanguageBox languages={this.state.account.organization.languages} />
              </React.Fragment>
          }
        </Header>
        <Content style={{display: "flex", flexDirection: "column"}} >
          {Setting.isMobile() ?
            this.renderRouter()
            :
            <Card className="content-warp-card">
              {this.renderRouter()}
            </Card>
          }
        </Content>
        {this.renderFooter()}
      </Layout>
    );
  }

  renderFooter() {
    return (
      <React.Fragment>
        {!this.state.account ? null : <div style={{display: "none"}} id="CasdoorApplicationName" value={this.state.account.signupApplication} />}
        <Footer id="footer" style={
          {
            textAlign: "center",
          }
        }>
          Powered by <a target="_blank" href="https://casdoor.org" rel="noreferrer"><img style={{paddingBottom: "3px"}} height={"20px"} alt={"Casdoor"} src={this.state.logo} /></a>
        </Footer>
      </React.Fragment>
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
      window.location.pathname.startsWith("/cas") ||
      window.location.pathname.startsWith("/auto-signup");
  }

  renderPage() {
    if (this.isDoorPages()) {
      return (
        <Layout id="parent-area">
          <Content style={{display: "flex", justifyContent: "center"}}>
            {
              this.isEntryPages() ?
                <EntryPage account={this.state.account} onUpdateAccount={(account) => {
                  this.onUpdateAccount(account);
                }} /> :
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
        </Layout>
      );
    }

    return (
      <React.Fragment>
        <FloatButton.BackTop />
        <CustomGithubCorner />
        {
          this.renderContent()
        }
      </React.Fragment>
    );
  }

  render() {
    return (
      <React.Fragment>
        <Helmet>
          {(this.state.account === undefined || this.state.account === null) ?
            <link rel="icon" href={"https://cdn.casdoor.com/static/favicon.png"} />
            : <React.Fragment>
              <title>{this.state.account.organization?.displayName}</title>
              <link rel="icon" href={this.state.account.organization?.favicon} />
            </React.Fragment>
          }
        </Helmet>
        <ConfigProvider theme={{
          token: {
            colorPrimary: "rgb(89,54,213)",
            colorInfo: "rgb(89,54,213)",
          },
          algorithm: this.state.themeAlgorithm,
        }}>
          {
            this.renderPage()
          }
        </ConfigProvider>
      </React.Fragment>
    );
  }
}

export default withRouter(withTranslation()(App));
