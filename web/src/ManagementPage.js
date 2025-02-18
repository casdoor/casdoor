// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

import * as Setting from "./Setting";
import {Avatar, Button, Card, Drawer, Dropdown, Menu, Result, Tooltip} from "antd";
import EnableMfaNotification from "./common/notifaction/EnableMfaNotification";
import {Link, Redirect, Route, Switch, withRouter} from "react-router-dom";
import React, {useState} from "react";
import i18next from "i18next";
import {
  AppstoreTwoTone,
  BarsOutlined, DeploymentUnitOutlined, DollarTwoTone, DownOutlined,
  HomeTwoTone,
  LockTwoTone, LogoutOutlined,
  SafetyCertificateTwoTone, SettingOutlined, SettingTwoTone,
  WalletTwoTone
} from "@ant-design/icons";
import Dashboard from "./basic/Dashboard";
import AppListPage from "./basic/AppListPage";
import ShortcutsPage from "./basic/ShortcutsPage";
import AccountPage from "./account/AccountPage";
import OrganizationListPage from "./OrganizationListPage";
import OrganizationEditPage from "./OrganizationEditPage";
import UserListPage from "./UserListPage";
import GroupTreePage from "./GroupTreePage";
import GroupListPage from "./GroupListPage";
import GroupEditPage from "./GroupEditPage";
import UserEditPage from "./UserEditPage";
import InvitationListPage from "./InvitationListPage";
import InvitationEditPage from "./InvitationEditPage";
import ApplicationListPage from "./ApplicationListPage";
import ApplicationEditPage from "./ApplicationEditPage";
import ProviderListPage from "./ProviderListPage";
import ProviderEditPage from "./ProviderEditPage";
import RecordListPage from "./RecordListPage";
import ResourceListPage from "./ResourceListPage";
import CertListPage from "./CertListPage";
import CertEditPage from "./CertEditPage";
import RoleListPage from "./RoleListPage";
import RoleEditPage from "./RoleEditPage";
import PermissionListPage from "./PermissionListPage";
import PermissionEditPage from "./PermissionEditPage";
import ModelListPage from "./ModelListPage";
import ModelEditPage from "./ModelEditPage";
import AdapterListPage from "./AdapterListPage";
import AdapterEditPage from "./AdapterEditPage";
import EnforcerListPage from "./EnforcerListPage";
import EnforcerEditPage from "./EnforcerEditPage";
import SessionListPage from "./SessionListPage";
import TokenListPage from "./TokenListPage";
import TokenEditPage from "./TokenEditPage";
import ProductListPage from "./ProductListPage";
import ProductEditPage from "./ProductEditPage";
import ProductBuyPage from "./ProductBuyPage";
import PaymentListPage from "./PaymentListPage";
import PaymentEditPage from "./PaymentEditPage";
import PaymentResultPage from "./PaymentResultPage";
import PlanListPage from "./PlanListPage";
import PlanEditPage from "./PlanEditPage";
import PricingListPage from "./PricingListPage";
import PricingEditPage from "./PricingEditPage";
import SubscriptionListPage from "./SubscriptionListPage";
import SubscriptionEditPage from "./SubscriptionEditPage";
import SystemInfo from "./SystemInfo";
import SyncerListPage from "./SyncerListPage";
import SyncerEditPage from "./SyncerEditPage";
import WebhookListPage from "./WebhookListPage";
import WebhookEditPage from "./WebhookEditPage";
import LdapEditPage from "./LdapEditPage";
import LdapSyncPage from "./LdapSyncPage";
import MfaSetupPage from "./auth/MfaSetupPage";
import OdicDiscoveryPage from "./auth/OidcDiscoveryPage";
import * as Conf from "./Conf";
import LanguageSelect from "./common/select/LanguageSelect";
import ThemeSelect from "./common/select/ThemeSelect";
import OpenTour from "./common/OpenTour";
import OrganizationSelect from "./common/select/OrganizationSelect";
import AccountAvatar from "./account/AccountAvatar";
import {Content, Header} from "antd/es/layout/layout";
import * as AuthBackend from "./auth/AuthBackend";
import {clearWeb3AuthToken} from "./auth/Web3Auth";
import TransactionListPage from "./TransactionListPage";
import TransactionEditPage from "./TransactionEditPage";
import VerificationListPage from "./VerificationListPage";

function ManagementPage(props) {

  const [menuVisible, setMenuVisible] = useState(false);

  function logout() {
    AuthBackend.logout()
      .then((res) => {
        if (res.status === "ok") {
          const owner = props.account.owner;
          props.setLogoutState();
          clearWeb3AuthToken();
          Setting.showMessage("success", i18next.t("application:Logged out successfully"));
          const redirectUri = res.data2;
          if (redirectUri !== null && redirectUri !== undefined && redirectUri !== "") {
            Setting.goToLink(redirectUri);
          } else if (owner !== "built-in") {
            Setting.goToLink(`${window.location.origin}/login/${owner}`);
          } else {
            Setting.goToLinkSoft({props}, "/");
          }
        } else {
          Setting.showMessage("error", `Failed to log out: ${res.msg}`);
        }
      });
  }

  function renderAvatar() {
    if (props.account.avatar === "") {
      return (
        <Avatar style={{backgroundColor: Setting.getAvatarColor(props.account.name), verticalAlign: "middle"}} size="large">
          {Setting.getShortName(props.account.name)}
        </Avatar>
      );
    } else {
      return (
        <Avatar src={props.account.avatar} style={{verticalAlign: "middle"}} size="large"
          icon={<AccountAvatar src={props.account.avatar} style={{verticalAlign: "middle"}} size={40} />}
        >
          {Setting.getShortName(props.account.name)}
        </Avatar>
      );
    }
  }

  function renderRightDropdown() {
    const items = [];
    if (props.requiredEnableMfa === false) {
      items.push(Setting.getItem(<><SettingOutlined />&nbsp;&nbsp;{i18next.t("account:My Account")}</>,
        "/account"
      ));
    }
    items.push(Setting.getItem(<><LogoutOutlined />&nbsp;&nbsp;{i18next.t("account:Logout")}</>,
      "/logout"));

    const onClick = (e) => {
      if (e.key === "/account") {
        props.history.push("/account");
      } else if (e.key === "/subscription") {
        props.history.push("/subscription");
      } else if (e.key === "/logout") {
        logout();
      }
    };

    return (
      <Dropdown key="/rightDropDown" menu={{items, onClick}} >
        <div className="rightDropDown">
          {
            renderAvatar()
          }
                    &nbsp;
                    &nbsp;
          {Setting.isMobile() ? null : Setting.getShortText(Setting.getNameAtLeast(props.account.displayName), 30)} &nbsp; <DownOutlined />
                    &nbsp;
                    &nbsp;
                    &nbsp;
        </div>
      </Dropdown>
    );
  }

  function renderAccountMenu() {
    if (props.account === undefined) {
      return null;
    } else if (props.account === null) {
      return (
        <React.Fragment>
          <LanguageSelect />
        </React.Fragment>
      );
    } else {
      return (
        <React.Fragment>
          {renderRightDropdown()}
          <ThemeSelect
            themeAlgorithm={props.themeAlgorithm}
            onChange={props.setLogoAndThemeAlgorithm} />
          <LanguageSelect languages={props.account.organization.languages} />
          {
            Conf.AiAssistantUrl?.trim() && (
              <Tooltip title="Click to open AI assistant">
                <div className="select-box" onClick={props.openAiAssistant}>
                  <DeploymentUnitOutlined style={{fontSize: "24px"}} />
                </div>
              </Tooltip>
            )
          }
          <OpenTour />
          {Setting.isAdminUser(props.account) && (props.uri.indexOf("/trees") === -1) &&
                        <OrganizationSelect
                          initValue={Setting.getOrganization()}
                          withAll={true}
                          className="org-select"
                          style={{display: Setting.isMobile() ? "none" : "flex"}}
                          onChange={(value) => {
                            Setting.setOrganization(value);
                          }}
                        />
          }
        </React.Fragment>
      );
    }
  }

  function getMenuItems() {
    const res = [];

    if (props.account === null || props.account === undefined) {
      return [];
    }

    let textColor = "black";
    const twoToneColor = props.themeData.colorPrimary;

    let logo = props.account.organization.logo ? props.account.organization.logo : Setting.getLogo(props.themeAlgorithm);
    if (props.themeAlgorithm.includes("dark")) {
      if (props.account.organization.logoDark) {
        logo = props.account.organization.logoDark;
      }
      textColor = "white";
    }

    !Setting.isMobile() ? res.push({
      label:
            <Link to="/">
              <img className="logo" src={logo ?? props.logo} alt="logo" />
            </Link>,
      disabled: true, key: "logo",
      style: {
        padding: 0,
        height: "auto",
      },
    }) : null;

    res.push(Setting.getItem(<Link style={{color: textColor}} to="/">{i18next.t("general:Home")}</Link>, "/home", <HomeTwoTone twoToneColor={twoToneColor} />, [
      Setting.getItem(<Link to="/">{i18next.t("general:Dashboard")}</Link>, "/"),
      Setting.getItem(<Link to="/shortcuts">{i18next.t("general:Shortcuts")}</Link>, "/shortcuts"),
      Setting.getItem(<Link to="/apps">{i18next.t("general:Apps")}</Link>, "/apps"),
    ].filter(item => {
      return Setting.isLocalAdminUser(props.account);
    })));

    if (Setting.isLocalAdminUser(props.account)) {
      if (Conf.ShowGithubCorner) {
        res.push(Setting.getItem(<a href={"https://casdoor.com"}>
          <span style={{fontWeight: "bold", backgroundColor: "rgba(87,52,211,0.4)", marginTop: "12px", paddingLeft: "5px", paddingRight: "5px", display: "flex", alignItems: "center", height: "40px", borderRadius: "5px"}}>
            🚀 SaaS Hosting 🔥
          </span>
        </a>, "#"));
      }

      res.push(Setting.getItem(<Link style={{color: textColor}} to="/organizations">{i18next.t("general:User Management")}</Link>, "/orgs", <AppstoreTwoTone twoToneColor={twoToneColor} />, [
        Setting.getItem(<Link to="/organizations">{i18next.t("general:Organizations")}</Link>, "/organizations"),
        Setting.getItem(<Link to="/groups">{i18next.t("general:Groups")}</Link>, "/groups"),
        Setting.getItem(<Link to="/users">{i18next.t("general:Users")}</Link>, "/users"),
        Setting.getItem(<Link to="/invitations">{i18next.t("general:Invitations")}</Link>, "/invitations"),
      ]));

      res.push(Setting.getItem(<Link style={{color: textColor}} to="/applications">{i18next.t("general:Identity")}</Link>, "/identity", <LockTwoTone twoToneColor={twoToneColor} />, [
        Setting.getItem(<Link to="/applications">{i18next.t("general:Applications")}</Link>, "/applications"),
        Setting.getItem(<Link to="/providers">{i18next.t("general:Providers")}</Link>, "/providers"),
        Setting.getItem(<Link to="/resources">{i18next.t("general:Resources")}</Link>, "/resources"),
        Setting.getItem(<Link to="/certs">{i18next.t("general:Certs")}</Link>, "/certs"),
      ]));

      res.push(Setting.getItem(<Link style={{color: textColor}} to="/roles">{i18next.t("general:Authorization")}</Link>, "/auth", <SafetyCertificateTwoTone twoToneColor={twoToneColor} />, [
        Setting.getItem(<Link to="/roles">{i18next.t("general:Roles")}</Link>, "/roles"),
        Setting.getItem(<Link to="/permissions">{i18next.t("general:Permissions")}</Link>, "/permissions"),
        Setting.getItem(<Link to="/models">{i18next.t("general:Models")}</Link>, "/models"),
        Setting.getItem(<Link to="/adapters">{i18next.t("general:Adapters")}</Link>, "/adapters"),
        Setting.getItem(<Link to="/enforcers">{i18next.t("general:Enforcers")}</Link>, "/enforcers"),
      ].filter(item => {
        if (!Setting.isLocalAdminUser(props.account) && ["/models", "/adapters", "/enforcers"].includes(item.key)) {
          return false;
        } else {
          return true;
        }
      })));

      res.push(Setting.getItem(<Link style={{color: textColor}} to="/sessions">{i18next.t("general:Logging & Auditing")}</Link>, "/logs", <WalletTwoTone twoToneColor={twoToneColor} />, [
        Setting.getItem(<Link to="/sessions">{i18next.t("general:Sessions")}</Link>, "/sessions"),
        Conf.CasvisorUrl ? Setting.getItem(<a target="_blank" rel="noreferrer" href={Conf.CasvisorUrl}>{i18next.t("general:Records")}</a>, "/records")
          : Setting.getItem(<Link to="/records">{i18next.t("general:Records")}</Link>, "/records"),
        Setting.getItem(<Link to="/tokens">{i18next.t("general:Tokens")}</Link>, "/tokens"),
        Setting.getItem(<Link to="/verifications">{i18next.t("general:Verifications")}</Link>, "/verifications"),
      ]));

      res.push(Setting.getItem(<Link style={{color: textColor}} to="/products">{i18next.t("general:Business & Payments")}</Link>, "/business", <DollarTwoTone twoToneColor={twoToneColor} />, [
        Setting.getItem(<Link to="/products">{i18next.t("general:Products")}</Link>, "/products"),
        Setting.getItem(<Link to="/payments">{i18next.t("general:Payments")}</Link>, "/payments"),
        Setting.getItem(<Link to="/plans">{i18next.t("general:Plans")}</Link>, "/plans"),
        Setting.getItem(<Link to="/pricings">{i18next.t("general:Pricings")}</Link>, "/pricings"),
        Setting.getItem(<Link to="/subscriptions">{i18next.t("general:Subscriptions")}</Link>, "/subscriptions"),
        Setting.getItem(<Link to="/transactions">{i18next.t("general:Transactions")}</Link>, "/transactions"),
      ]));

      if (Setting.isAdminUser(props.account)) {
        res.push(Setting.getItem(<Link style={{color: textColor}} to="/sysinfo">{i18next.t("general:Admin")}</Link>, "/admin", <SettingTwoTone twoToneColor={twoToneColor} />, [
          Setting.getItem(<Link to="/sysinfo">{i18next.t("general:System Info")}</Link>, "/sysinfo"),
          Setting.getItem(<Link to="/syncers">{i18next.t("general:Syncers")}</Link>, "/syncers"),
          Setting.getItem(<Link to="/webhooks">{i18next.t("general:Webhooks")}</Link>, "/webhooks"),
          Setting.getItem(<a target="_blank" rel="noreferrer" href={Setting.isLocalhost() ? `${Setting.ServerUrl}/swagger` : "/swagger"}>{i18next.t("general:Swagger")}</a>, "/swagger")]));
      } else {
        res.push(Setting.getItem(<Link style={{color: textColor}} to="/syncers">{i18next.t("general:Admin")}</Link>, "/admin", <SettingTwoTone twoToneColor={twoToneColor} />, [
          Setting.getItem(<Link to="/syncers">{i18next.t("general:Syncers")}</Link>, "/syncers"),
          Setting.getItem(<Link to="/webhooks">{i18next.t("general:Webhooks")}</Link>, "/webhooks")]));
      }
    }

    const navItems = props.account.organization.navItems;

    if (!Array.isArray(navItems)) {
      return res;
    }

    if (navItems.includes("all")) {
      return res;
    }

    const resFiltered = res.map(item => {
      if (!Array.isArray(item.children)) {
        return item;
      }
      const filteredChildren = [];
      item.children.forEach(itemChild => {
        if (navItems.includes(itemChild.key)) {
          filteredChildren.push(itemChild);
        }
      });

      item.children = filteredChildren;
      return item;
    });

    return resFiltered.filter(item => {
      if (item.key === "#" || item.key === "logo") {return true;}
      return Array.isArray(item.children) && item.children.length > 0;
    });
  }

  function renderLoginIfNotLoggedIn(component) {
    if (props.account === null) {
      sessionStorage.setItem("from", window.location.pathname);
      return <Redirect to="/login" />;
    } else if (props.account === undefined) {
      return null;
    } else if (props.account.needUpdatePassword) {
      return <Redirect to={"/forget/" + props.application.name} />;
    } else {
      return component;
    }
  }

  function renderRouter() {
    const account = props.account;
    const onChangeTheme = props.onChangeTheme;
    const onfinish = props.onfinish;
    return (
      <Switch>
        <Route exact path="/" render={(props) => renderLoginIfNotLoggedIn(<Dashboard account={account} {...props} />)} />
        <Route exact path="/apps" render={(props) => renderLoginIfNotLoggedIn(<AppListPage account={account} {...props} />)} />
        <Route exact path="/shortcuts" render={(props) => renderLoginIfNotLoggedIn(<ShortcutsPage account={account} {...props} />)} />
        <Route exact path="/account" render={(props) => renderLoginIfNotLoggedIn(<AccountPage account={account} {...props} />)} />
        <Route exact path="/organizations" render={(props) => renderLoginIfNotLoggedIn(<OrganizationListPage account={account} {...props} />)} />
        <Route exact path="/organizations/:organizationName" render={(props) => renderLoginIfNotLoggedIn(<OrganizationEditPage account={account} onChangeTheme={onChangeTheme} {...props} />)} />
        <Route exact path="/organizations/:organizationName/users" render={(props) => renderLoginIfNotLoggedIn(<UserListPage account={account} {...props} />)} />
        <Route exact path="/trees/:organizationName" render={(props) => renderLoginIfNotLoggedIn(<GroupTreePage account={account} {...props} />)} />
        <Route exact path="/trees/:organizationName/:groupName" render={(props) => renderLoginIfNotLoggedIn(<GroupTreePage account={account} {...props} />)} />
        <Route exact path="/groups" render={(props) => renderLoginIfNotLoggedIn(<GroupListPage account={account} {...props} />)} />
        <Route exact path="/groups/:organizationName/:groupName" render={(props) => renderLoginIfNotLoggedIn(<GroupEditPage account={account} {...props} />)} />
        <Route exact path="/users" render={(props) => renderLoginIfNotLoggedIn(<UserListPage account={account} {...props} />)} />
        <Route exact path="/users/:organizationName/:userName" render={(props) => <UserEditPage account={account} {...props} />} />
        <Route exact path="/invitations" render={(props) => renderLoginIfNotLoggedIn(<InvitationListPage account={account} {...props} />)} />
        <Route exact path="/invitations/:organizationName/:invitationName" render={(props) => renderLoginIfNotLoggedIn(<InvitationEditPage account={account} {...props} />)} />
        <Route exact path="/applications" render={(props) => renderLoginIfNotLoggedIn(<ApplicationListPage account={account} {...props} />)} />
        <Route exact path="/applications/:organizationName/:applicationName" render={(props) => renderLoginIfNotLoggedIn(<ApplicationEditPage account={account} {...props} />)} />
        <Route exact path="/providers" render={(props) => renderLoginIfNotLoggedIn(<ProviderListPage account={account} {...props} />)} />
        <Route exact path="/providers/:organizationName/:providerName" render={(props) => renderLoginIfNotLoggedIn(<ProviderEditPage account={account} {...props} />)} />
        <Route exact path="/records" render={(props) => renderLoginIfNotLoggedIn(<RecordListPage account={account} {...props} />)} />
        <Route exact path="/resources" render={(props) => renderLoginIfNotLoggedIn(<ResourceListPage account={account} {...props} />)} />
        <Route exact path="/certs" render={(props) => renderLoginIfNotLoggedIn(<CertListPage account={account} {...props} />)} />
        <Route exact path="/certs/:organizationName/:certName" render={(props) => renderLoginIfNotLoggedIn(<CertEditPage account={account} {...props} />)} />
        <Route exact path="/verifications" render={(props) => renderLoginIfNotLoggedIn(<VerificationListPage account={account} {...props} />)} />
        <Route exact path="/roles" render={(props) => renderLoginIfNotLoggedIn(<RoleListPage account={account} {...props} />)} />
        <Route exact path="/roles/:organizationName/:roleName" render={(props) => renderLoginIfNotLoggedIn(<RoleEditPage account={account} {...props} />)} />
        <Route exact path="/permissions" render={(props) => renderLoginIfNotLoggedIn(<PermissionListPage account={account} {...props} />)} />
        <Route exact path="/permissions/:organizationName/:permissionName" render={(props) => renderLoginIfNotLoggedIn(<PermissionEditPage account={account} {...props} />)} />
        <Route exact path="/models" render={(props) => renderLoginIfNotLoggedIn(<ModelListPage account={account} {...props} />)} />
        <Route exact path="/models/:organizationName/:modelName" render={(props) => renderLoginIfNotLoggedIn(<ModelEditPage account={account} {...props} />)} />
        <Route exact path="/adapters" render={(props) => renderLoginIfNotLoggedIn(<AdapterListPage account={account} {...props} />)} />
        <Route exact path="/adapters/:organizationName/:adapterName" render={(props) => renderLoginIfNotLoggedIn(<AdapterEditPage account={account} {...props} />)} />
        <Route exact path="/enforcers" render={(props) => renderLoginIfNotLoggedIn(<EnforcerListPage account={account} {...props} />)} />
        <Route exact path="/enforcers/:organizationName/:enforcerName" render={(props) => renderLoginIfNotLoggedIn(<EnforcerEditPage account={account} {...props} />)} />
        <Route exact path="/sessions" render={(props) => renderLoginIfNotLoggedIn(<SessionListPage account={account} {...props} />)} />
        <Route exact path="/tokens" render={(props) => renderLoginIfNotLoggedIn(<TokenListPage account={account} {...props} />)} />
        <Route exact path="/tokens/:tokenName" render={(props) => renderLoginIfNotLoggedIn(<TokenEditPage account={account} {...props} />)} />
        <Route exact path="/products" render={(props) => renderLoginIfNotLoggedIn(<ProductListPage account={account} {...props} />)} />
        <Route exact path="/products/:organizationName/:productName" render={(props) => renderLoginIfNotLoggedIn(<ProductEditPage account={account} {...props} />)} />
        <Route exact path="/products/:organizationName/:productName/buy" render={(props) => renderLoginIfNotLoggedIn(<ProductBuyPage account={account} {...props} />)} />
        <Route exact path="/payments" render={(props) => renderLoginIfNotLoggedIn(<PaymentListPage account={account} {...props} />)} />
        <Route exact path="/payments/:organizationName/:paymentName" render={(props) => renderLoginIfNotLoggedIn(<PaymentEditPage account={account} {...props} />)} />
        <Route exact path="/payments/:organizationName/:paymentName/result" render={(props) => renderLoginIfNotLoggedIn(<PaymentResultPage account={account} {...props} />)} />
        <Route exact path="/plans" render={(props) => renderLoginIfNotLoggedIn(<PlanListPage account={account} {...props} />)} />
        <Route exact path="/plans/:organizationName/:planName" render={(props) => renderLoginIfNotLoggedIn(<PlanEditPage account={account} {...props} />)} />
        <Route exact path="/pricings" render={(props) => renderLoginIfNotLoggedIn(<PricingListPage account={account} {...props} />)} />
        <Route exact path="/pricings/:organizationName/:pricingName" render={(props) => renderLoginIfNotLoggedIn(<PricingEditPage account={account} {...props} />)} />
        <Route exact path="/subscriptions" render={(props) => renderLoginIfNotLoggedIn(<SubscriptionListPage account={account} {...props} />)} />
        <Route exact path="/subscriptions/:organizationName/:subscriptionName" render={(props) => renderLoginIfNotLoggedIn(<SubscriptionEditPage account={account} {...props} />)} />
        <Route exact path="/sysinfo" render={(props) => renderLoginIfNotLoggedIn(<SystemInfo account={account} {...props} />)} />
        <Route exact path="/syncers" render={(props) => renderLoginIfNotLoggedIn(<SyncerListPage account={account} {...props} />)} />
        <Route exact path="/syncers/:syncerName" render={(props) => renderLoginIfNotLoggedIn(<SyncerEditPage account={account} {...props} />)} />
        <Route exact path="/transactions" render={(props) => renderLoginIfNotLoggedIn(<TransactionListPage account={account} {...props} />)} />
        <Route exact path="/transactions/:organizationName/:transactionName" render={(props) => renderLoginIfNotLoggedIn(<TransactionEditPage account={account} {...props} />)} />
        <Route exact path="/webhooks" render={(props) => renderLoginIfNotLoggedIn(<WebhookListPage account={account} {...props} />)} />
        <Route exact path="/webhooks/:webhookName" render={(props) => renderLoginIfNotLoggedIn(<WebhookEditPage account={account} {...props} />)} />
        <Route exact path="/ldap/:organizationName/:ldapId" render={(props) => renderLoginIfNotLoggedIn(<LdapEditPage account={account} {...props} />)} />
        <Route exact path="/ldap/sync/:organizationName/:ldapId" render={(props) => renderLoginIfNotLoggedIn(<LdapSyncPage account={account} {...props} />)} />
        <Route exact path="/mfa/setup" render={(props) => renderLoginIfNotLoggedIn(<MfaSetupPage account={account} onfinish={onfinish} {...props} />)} />
        <Route exact path="/.well-known/openid-configuration" render={(props) => <OdicDiscoveryPage />} />
        <Route path="" render={() => <Result status="404" title="404 NOT FOUND" subTitle={i18next.t("general:Sorry, the page you visited does not exist.")}
          extra={<a href="/"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>} />} />
      </Switch>
    );
  }

  function isWithoutCard() {
    return Setting.isMobile() || window.location.pathname.startsWith("/trees");
  }

  const menuStyleRight = Setting.isAdminUser(props.account) && !Setting.isMobile() ? "calc(180px + 280px)" : "320px";

  const onClose = () => {
    setMenuVisible(false);
  };

  const showMenu = () => {
    setMenuVisible(true);
  };

  return (
    <React.Fragment>
      <EnableMfaNotification account={props.account} />
      <Header style={{padding: "0", marginBottom: "3px", backgroundColor: props.themeAlgorithm.includes("dark") ? "black" : "white"}} >
        {props.requiredEnableMfa || (Setting.isMobile() ?
          <React.Fragment>
            <Drawer title={i18next.t("general:Close")} placement="left" open={menuVisible} onClose={onClose}>
              <Menu
                items={getMenuItems()}
                mode={"inline"}
                selectedKeys={[props.selectedMenuKey]}
                style={{lineHeight: "64px"}}
                onClick={onClose}
              >
              </Menu>
            </Drawer>
            <Button icon={<BarsOutlined />} onClick={showMenu} type="text">
              {i18next.t("general:Menu")}
            </Button>
          </React.Fragment> :
          <Menu
            onClick={onClose}
            items={getMenuItems()}
            mode={"horizontal"}
            selectedKeys={[props.selectedMenuKey]}
            style={{position: "absolute", left: 0, right: menuStyleRight, backgroundColor: props.themeAlgorithm.includes("dark") ? "black" : "white"}}
          />
        )}
        {
          renderAccountMenu()
        }
      </Header>
      <Content style={{display: "flex", flexDirection: "column"}} >
        {isWithoutCard() ?
          renderRouter() :
          <Card className="content-warp-card">
            {renderRouter()}
          </Card>
        }
      </Content>
    </React.Fragment>
  );
}

export default withRouter(ManagementPage);
