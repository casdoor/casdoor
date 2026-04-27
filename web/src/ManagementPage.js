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
import Sider from "antd/es/layout/Sider";
import EnableMfaNotification from "./common/notifaction/EnableMfaNotification";
import {Link, Redirect, Route, Switch, withRouter} from "react-router-dom";
import React, {useEffect, useState} from "react";
import i18next from "i18next";
import {
  AppstoreOutlined,
  BarsOutlined, CheckCircleOutlined, DeploymentUnitOutlined, DollarOutlined, DownOutlined,
  HomeOutlined,
  LockOutlined, LogoutOutlined,
  MenuFoldOutlined, MenuUnfoldOutlined,
  SafetyCertificateOutlined, SettingOutlined,
  WalletOutlined
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
import KeyListPage from "./KeyListPage";
import KeyEditPage from "./KeyEditPage";
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
import ProductStorePage from "./ProductStorePage";
import ProductEditPage from "./ProductEditPage";
import ProductBuyPage from "./ProductBuyPage";
import CartListPage from "./CartListPage";
import CouponListPage from "./CouponListPage";
import CouponEditPage from "./CouponEditPage";
import OrderListPage from "./OrderListPage";
import OrderEditPage from "./OrderEditPage";
import OrderPayPage from "./OrderPayPage";
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
import FormListPage from "./FormListPage";
import FormEditPage from "./FormEditPage";
import SyncerListPage from "./SyncerListPage";
import SyncerEditPage from "./SyncerEditPage";
import WebhookListPage from "./WebhookListPage";
import WebhookEventListPage from "./WebhookEventListPage";
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
import BreadcrumbBar from "./common/BreadcrumbBar";
import {Content, Header} from "antd/es/layout/layout";
import * as AuthBackend from "./auth/AuthBackend";
import {clearWeb3AuthToken} from "./auth/Web3Auth";
import TransactionListPage from "./TransactionListPage";
import TransactionEditPage from "./TransactionEditPage";
import VerificationListPage from "./VerificationListPage";
import TicketListPage from "./TicketListPage";
import TicketEditPage from "./TicketEditPage";
import * as Cookie from "cookie";
import * as UserBackend from "./backend/UserBackend";
import AgentListPage from "./AgentListPage";
import AgentEditPage from "./AgentEditPage";
import ServerListPage from "./ServerListPage";
import ServerStorePage from "./ServerStorePage";
import ServerEditPage from "./ServerEditPage";
import EntryListPage from "./EntryListPage";
import EntryEditPage from "./EntryEditPage";
import OpenClawSessionTranscriptPage from "./OpenClawSessionTranscriptPage";
import SiteListPage from "./SiteListPage";
import SiteEditPage from "./SiteEditPage";
import RuleListPage from "./RuleListPage";
import RuleEditPage from "./RuleEditPage";

function getMenuParentKey(uri) {
  if (!uri) {return null;}
  if (uri === "/" || uri.includes("/shortcuts") || uri.includes("/apps")) {return "/home";}
  if (uri.includes("/organizations") || uri.includes("/trees") || uri.includes("/groups") || uri.includes("/users") || uri.includes("/invitations")) {return "/orgs";}
  if (uri.includes("/applications") || uri.includes("/providers") || uri.includes("/resources") || uri.includes("/certs") || uri.includes("/keys")) {return "/identity";}
  if (uri.includes("/agents") || uri.includes("/servers") || uri.includes("/server-store") || uri.includes("/entries") || uri.includes("/sites") || uri.includes("/rules")) {return "/gateway";}
  if (uri.includes("/roles") || uri.includes("/permissions") || uri.includes("/models") || uri.includes("/adapters") || uri.includes("/enforcers")) {return "/auth";}
  if (uri.includes("/records") || uri.includes("/tokens") || uri.includes("/sessions") || uri.includes("/verifications")) {return "/logs";}
  if (uri.includes("/product-store") || uri.includes("/products") || uri.includes("/coupons") || uri.includes("/orders") || uri.includes("/payments") || uri.includes("/plans") || uri.includes("/pricings") || uri.includes("/subscriptions") || uri.includes("/transactions") || uri.includes("/cart")) {return "/business";}
  if (uri.includes("/sysinfo") || uri.includes("/forms") || uri.includes("/syncers") || uri.includes("/webhooks") || uri.includes("/webhook-events") || uri.includes("/tickets")) {return "/admin";}
  return null;
}

function ManagementPage(props) {
  const [menuVisible, setMenuVisible] = useState(false);
  const [siderCollapsed, setSiderCollapsed] = useState(() => localStorage.getItem("siderCollapsed") === "true");
  const [menuOpenKeys, setMenuOpenKeys] = useState(() => {
    if (localStorage.getItem("siderCollapsed") === "true") {
      return [];
    }
    const parentKey = getMenuParentKey(props.uri || location.pathname);
    return parentKey ? [parentKey] : [];
  });

  useEffect(() => {
    if (siderCollapsed) {
      setMenuOpenKeys([]);
      return;
    }
    const parentKey = getMenuParentKey(props.uri);
    if (parentKey) {
      setMenuOpenKeys(prev =>
        prev.includes(parentKey) ? prev : [...prev, parentKey]
      );
    }
  }, [props.uri, siderCollapsed]);
  const organization = props.account?.organization;
  const navItems = Setting.isLocalAdminUser(props.account) ? organization?.navItems : (organization?.userNavItems ?? []);
  const widgetItems = organization?.widgetItems;
  const currentUri = props.uri || location.pathname;
  const selectedLeafKey = "/" + (currentUri.split("/").filter(Boolean)[0] || "");

  const isDark = props.themeAlgorithm.includes("dark");
  const textColor = isDark ? "white" : "black";
  const siderLogo = (() => {
    if (!props.account?.organization) {return Setting.getLogo(props.themeAlgorithm);}
    let logo = props.account.organization.logo || Setting.getLogo(props.themeAlgorithm);
    if (isDark && props.account.organization.logoDark) {
      logo = props.account.organization.logoDark;
    }
    return logo;
  })();

  const toggleSider = () => {
    const next = !siderCollapsed;
    setSiderCollapsed(next);
    localStorage.setItem("siderCollapsed", String(next));
  };

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
          Setting.showMessage("error", `${i18next.t("general:Failed to log out")}: ${res.msg}`);
        }
      });
  }

  function renderAvatar() {
    if (props.account.avatar === "") {
      return (
        <div style={{verticalAlign: "middle", marginLeft: 8, display: "inline-flex"}}>
          {Setting.getAvatarPlaceholder(props.account.name, 40)}
        </div>
      );
    } else {
      return (
        <Avatar src={props.account.avatar} style={{verticalAlign: "middle", marginLeft: 8}} size="large"
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
    const curCookie = Cookie.parse(document.cookie);
    if (curCookie["impersonateUser"]) {
      items.push(Setting.getItem(<><LogoutOutlined />&nbsp;&nbsp;{i18next.t("account:Exit impersonation")}</>,
        "/exit-impersonation"));
    } else {
      items.push(Setting.getItem(<><LogoutOutlined />&nbsp;&nbsp;{i18next.t("account:Logout")}</>,
        "/logout"));
    }

    const onClick = (e) => {
      if (e.key === "/account") {
        props.history.push("/account");
      } else if (e.key === "/subscription") {
        props.history.push("/subscription");
      } else if (e.key === "/logout") {
        logout();
      } else if (e.key === "/exit-impersonation") {
        UserBackend.exitImpersonateUser().then((res) => {
          if (res.status === "ok") {
            Setting.showMessage("success", i18next.t("account:Exit impersonation"));
            Setting.goToLinkSoft({props}, "/");
            window.location.reload();
          } else {
            Setting.showMessage("error", res.msg);
          }
        });
      }
    };

    return (
      <Dropdown key="/rightDropDown" menu={{items, onClick}} placement="bottomRight" >
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

  function navItemsIsAll() {
    return !Array.isArray(navItems) || !!navItems?.includes("all");
  }

  function widgetItemsIsAll() {
    return !Array.isArray(widgetItems) || !!widgetItems?.includes("all");
  }

  function isSpecialMenuItem(item) {
    return item.key === "#" || item.key === "logo";
  }

  function renderWidgets() {
    const widgets = [
      Setting.getItem(<ThemeSelect themeAlgorithm={props.themeAlgorithm} onChange={props.setLogoAndThemeAlgorithm} />, "theme"),
      Setting.getItem(<LanguageSelect languages={props.account.organization.languages} />, "language"),
      Setting.getItem(Conf.AiAssistantUrl?.trim() && (
        <Tooltip title={i18next.t("general:Click to open AI assistant")}>
          <div className="select-box" onClick={props.openAiAssistant}>
            <DeploymentUnitOutlined style={{fontSize: "24px"}} />
          </div>
        </Tooltip>
      ), "ai-assistant"),
      Setting.getItem(<OpenTour />, "tour"),
    ];

    if (widgetItemsIsAll()) {
      return widgets.reverse().map(item => item.label);
    }

    return widgets.filter(item => widgetItems.includes(item.key)).reverse().map(item => item.label);
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
          {Setting.isLocalAdminUser(props.account) && Conf.ShowGithubCorner && !Setting.isMobile() &&
            <a href={"https://casdoor.com"} target="_blank" rel="noreferrer" style={{marginRight: "8px"}}>
              <span className="saas-hosting-btn">
                🚀 SaaS Hosting 🔥
              </span>
            </a>
          }
          {Setting.isAdminUser(props.account) && (props.uri.indexOf("/trees") === -1) &&
            <OrganizationSelect
              initValue={Setting.getOrganization()}
              withAll={true}
              className="org-select"
              style={{display: Setting.isMobile() ? "none" : "flex", backgroundColor: isDark ? "black" : "white", color: textColor}}
              onChange={(value) => {
                Setting.setOrganization(value);
              }}
            />
          }
          {renderWidgets()}
          {renderRightDropdown()}
        </React.Fragment>
      );
    }
  }

  function getMenuItems() {
    const res = [];

    if (props.account === null || props.account === undefined) {
      return [];
    }

    res.push(Setting.getItem(<Link style={{color: textColor}} to="/">{i18next.t("general:Home")}</Link>, "/home", <HomeOutlined />, [
      Setting.getItem(<Link to="/">{i18next.t("general:Dashboard")}</Link>, "/"),
      Setting.getItem(<Link to="/shortcuts">{i18next.t("general:Shortcuts")}</Link>, "/shortcuts"),
      Setting.getItem(<Link to="/apps">{i18next.t("general:Apps")}</Link>, "/apps"),
    ]));

    res.push(Setting.getItem(<Link style={{color: textColor}} to="/organizations">{i18next.t("general:User Management")}</Link>, "/orgs", <AppstoreOutlined />, [
      Setting.getItem(<Link to="/organizations">{i18next.t("general:Organizations")}</Link>, "/organizations"),
      Setting.getItem(<Link to="/groups">{i18next.t("general:Groups")}</Link>, "/groups"),
      Setting.getItem(<Link to="/users">{i18next.t("general:Users")}</Link>, "/users"),
      Setting.getItem(<Link to="/invitations">{i18next.t("general:Invitations")}</Link>, "/invitations"),
    ]));

    res.push(Setting.getItem(<Link style={{color: textColor}} to="/applications">{i18next.t("general:Identity")}</Link>, "/identity", <LockOutlined />, [
      Setting.getItem(<Link to="/applications">{i18next.t("general:Applications")}</Link>, "/applications"),
      Setting.getItem(<Link to="/providers">{i18next.t("application:Providers")}</Link>, "/providers"),
      Setting.getItem(<Link to="/resources">{i18next.t("general:Resources")}</Link>, "/resources"),
      Setting.getItem(<Link to="/certs">{i18next.t("general:Certs")}</Link>, "/certs"),
      Setting.getItem(<Link to="/keys">{i18next.t("general:Keys")}</Link>, "/keys"),
    ]));

    res.push(Setting.getItem(<Link style={{color: textColor}} to="/roles">{i18next.t("general:Authorization")}</Link>, "/auth", <SafetyCertificateOutlined />, [
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

    res.push(Setting.getItem(<Link style={{color: textColor}} to="/sites">{i18next.t("general:LLM AI")}</Link>, "/gateway", <CheckCircleOutlined />, [
      Setting.getItem(<Link to="/agents">{i18next.t("general:Agents")}</Link>, "/agents"),
      Setting.getItem(<Link to="/servers">{i18next.t("general:MCP Servers")}</Link>, "/servers"),
      Setting.getItem(<Link to="/server-store">{i18next.t("general:MCP Store")}</Link>, "/server-store"),
      Setting.getItem(<Link to="/entries">{i18next.t("general:Entries")}</Link>, "/entries"),
      Setting.getItem(<Link to="/sites">{i18next.t("general:Sites")}</Link>, "/sites"),
      Setting.getItem(<Link to="/rules">{i18next.t("general:Rules")}</Link>, "/rules"),
    ]));

    res.push(Setting.getItem(<Link style={{color: textColor}} to="/sessions">{i18next.t("general:Auditing")}</Link>, "/logs", <WalletOutlined />, [
      Setting.getItem(<Link to="/sessions">{i18next.t("general:Sessions")}</Link>, "/sessions"),
      Setting.getItem(<Link to="/records">{i18next.t("general:Records")}</Link>, "/records"),
      Setting.getItem(<Link to="/tokens">{i18next.t("general:Tokens")}</Link>, "/tokens"),
      Setting.getItem(<Link to="/verifications">{i18next.t("general:Verifications")}</Link>, "/verifications"),
    ]));

    res.push(Setting.getItem(<Link style={{color: textColor}} to="/products">{i18next.t("general:Business")}</Link>, "/business", <DollarOutlined />, [
      Setting.getItem(<Link to="/product-store">{i18next.t("general:Product Store")}</Link>, "/product-store"),
      Setting.getItem(<Link to="/products">{i18next.t("general:Products")}</Link>, "/products"),
      Setting.getItem(<Link to="/coupons">{i18next.t("general:Coupons")}</Link>, "/coupons"),
      Setting.getItem(<Link to="/cart">{i18next.t("general:Cart")}</Link>, "/cart"),
      Setting.getItem(<Link to="/orders">{i18next.t("general:Orders")}</Link>, "/orders"),
      Setting.getItem(<Link to="/payments">{i18next.t("general:Payments")}</Link>, "/payments"),
      Setting.getItem(<Link to="/plans">{i18next.t("general:Plans")}</Link>, "/plans"),
      Setting.getItem(<Link to="/pricings">{i18next.t("general:Pricings")}</Link>, "/pricings"),
      Setting.getItem(<Link to="/subscriptions">{i18next.t("general:Subscriptions")}</Link>, "/subscriptions"),
      Setting.getItem(<Link to="/transactions">{i18next.t("general:Transactions")}</Link>, "/transactions"),
    ]));

    if (Setting.isAdminUser(props.account)) {
      res.push(Setting.getItem(<Link style={{color: textColor}} to="/sysinfo">{i18next.t("general:Admin")}</Link>, "/admin", <SettingOutlined />, [
        Setting.getItem(<Link to="/sysinfo">{i18next.t("general:System Info")}</Link>, "/sysinfo"),
        Setting.getItem(<Link to="/forms">{i18next.t("general:Forms")}</Link>, "/forms"),
        Setting.getItem(<Link to="/syncers">{i18next.t("general:Syncers")}</Link>, "/syncers"),
        Setting.getItem(<Link to="/webhooks">{i18next.t("general:Webhooks")}</Link>, "/webhooks"),
        Setting.getItem(<Link to="/webhook-events">{i18next.t("general:Webhook Events")}</Link>, "/webhook-events"),
        Setting.getItem(<Link to="/tickets">{i18next.t("general:Tickets")}</Link>, "/tickets"),
        Setting.getItem(<a target="_blank" rel="noreferrer" href={Setting.isLocalhost() ? `${Setting.ServerUrl}/swagger` : "/swagger"}>{i18next.t("general:Swagger")}</a>, "/swagger")]));
    } else {
      res.push(Setting.getItem(<Link style={{color: textColor}} to="/syncers">{i18next.t("general:Admin")}</Link>, "/admin", <SettingOutlined />, [
        Setting.getItem(<Link to="/forms">{i18next.t("general:Forms")}</Link>, "/forms"),
        Setting.getItem(<Link to="/syncers">{i18next.t("general:Syncers")}</Link>, "/syncers"),
        Setting.getItem(<Link to="/webhooks">{i18next.t("general:Webhooks")}</Link>, "/webhooks"),
        Setting.getItem(<Link to="/webhook-events">{i18next.t("general:Webhook Events")}</Link>, "/webhook-events"),
        Setting.getItem(<Link to="/tickets">{i18next.t("general:Tickets")}</Link>, "/tickets")]));
    }

    if (navItemsIsAll()) {
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

    const filteredResult = resFiltered.filter(item => {
      if (isSpecialMenuItem(item)) {return true;}
      return Array.isArray(item.children) && item.children.length > 0;
    });

    // Count total end items (leaf nodes)
    let totalEndItems = 0;
    filteredResult.forEach(item => {
      if (Array.isArray(item.children)) {
        totalEndItems += item.children.length;
      }
    });

    // If total end items <= MaxItemsForFlatMenu, flatten the menu (show only one level)
    if (totalEndItems <= Conf.MaxItemsForFlatMenu) {
      const flattenedResult = [];
      filteredResult.forEach(item => {
        if (isSpecialMenuItem(item)) {
          flattenedResult.push(item);
        } else if (Array.isArray(item.children)) {
          // Add children directly without parent group
          item.children.forEach(child => {
            flattenedResult.push(child);
          });
        }
      });
      return flattenedResult;
    }

    return filteredResult;
  }

  function renderLoginIfNotLoggedIn(component) {
    if (props.account === null) {
      const lastLoginOrg = localStorage.getItem("lastLoginOrg");
      sessionStorage.setItem("from", window.location.pathname);
      if (lastLoginOrg) {
        return <Redirect to={`/login/${lastLoginOrg}`} />;
      } else {
        return <Redirect to="/login" />;
      }
    } else if (props.account === undefined) {
      return null;
    } else if (props.account.needUpdatePassword) {
      if (window.location.pathname === "/account") {
        return component;
      } else {
        return <Redirect to="/account" />;
      }
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
        <Route exact path="/keys" render={(props) => renderLoginIfNotLoggedIn(<KeyListPage account={account} {...props} />)} />
        <Route exact path="/keys/:organizationName/:keyName" render={(props) => renderLoginIfNotLoggedIn(<KeyEditPage account={account} {...props} />)} />
        <Route exact path="/agents" render={(props) => renderLoginIfNotLoggedIn(<AgentListPage account={account} {...props} />)} />
        <Route exact path="/agents/:organizationName/:agentName" render={(props) => renderLoginIfNotLoggedIn(<AgentEditPage account={account} {...props} />)} />
        <Route exact path="/servers" render={(props) => renderLoginIfNotLoggedIn(<ServerListPage account={account} {...props} />)} />
        <Route exact path="/server-store" render={(props) => renderLoginIfNotLoggedIn(<ServerStorePage account={account} {...props} />)} />
        <Route exact path="/servers/:organizationName/:serverName" render={(props) => renderLoginIfNotLoggedIn(<ServerEditPage account={account} {...props} />)} />
        <Route exact path="/entries" render={(props) => renderLoginIfNotLoggedIn(<EntryListPage account={account} {...props} />)} />
        <Route exact path="/entries/:organizationName/:entryName/transcript" render={(props) => renderLoginIfNotLoggedIn(<OpenClawSessionTranscriptPage account={account} {...props} />)} />
        <Route exact path="/entries/:organizationName/:entryName" render={(props) => renderLoginIfNotLoggedIn(<EntryEditPage account={account} {...props} />)} />
        <Route exact path="/sites" render={(props) => renderLoginIfNotLoggedIn(<SiteListPage account={account} {...props} />)} />
        <Route exact path="/sites/:organizationName/:siteName" render={(props) => renderLoginIfNotLoggedIn(<SiteEditPage account={account} {...props} />)} />
        <Route exact path="/rules" render={(props) => renderLoginIfNotLoggedIn(<RuleListPage account={account} {...props} />)} />
        <Route exact path="/rules/:organizationName/:ruleName" render={(props) => renderLoginIfNotLoggedIn(<RuleEditPage account={account} {...props} />)} />
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
        <Route exact path="/product-store" render={(props) => renderLoginIfNotLoggedIn(<ProductStorePage account={account} {...props} />)} />
        <Route exact path="/products" render={(props) => renderLoginIfNotLoggedIn(<ProductListPage account={account} {...props} />)} />
        <Route exact path="/products/:organizationName/:productName" render={(props) => renderLoginIfNotLoggedIn(<ProductEditPage account={account} {...props} />)} />
        <Route exact path="/products/:organizationName/:productName/buy" render={(props) => renderLoginIfNotLoggedIn(<ProductBuyPage account={account} {...props} />)} />
        <Route exact path="/coupons" render={(props) => renderLoginIfNotLoggedIn(<CouponListPage account={account} {...props} />)} />
        <Route exact path="/coupons/:organizationName/:couponName" render={(props) => renderLoginIfNotLoggedIn(<CouponEditPage account={account} {...props} />)} />
        <Route exact path="/cart" render={(props) => renderLoginIfNotLoggedIn(<CartListPage account={account} {...props} />)} />
        <Route exact path="/orders" render={(props) => renderLoginIfNotLoggedIn(<OrderListPage account={account} {...props} />)} />
        <Route exact path="/orders/:organizationName/:orderName" render={(props) => renderLoginIfNotLoggedIn(<OrderEditPage account={account} {...props} />)} />
        <Route exact path="/orders/:organizationName/:orderName/pay" render={(props) => renderLoginIfNotLoggedIn(<OrderPayPage account={account} {...props} />)} />
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
        <Route exact path="/forms" render={(props) => renderLoginIfNotLoggedIn(<FormListPage account={account} {...props} />)} />
        <Route exact path="/forms/:formName" render={(props) => renderLoginIfNotLoggedIn(<FormEditPage account={account} {...props} />)} />
        <Route exact path="/syncers" render={(props) => renderLoginIfNotLoggedIn(<SyncerListPage account={account} {...props} />)} />
        <Route exact path="/syncers/:organizationName/:syncerName" render={(props) => renderLoginIfNotLoggedIn(<SyncerEditPage account={account} {...props} />)} />
        <Route exact path="/transactions" render={(props) => renderLoginIfNotLoggedIn(<TransactionListPage account={account} {...props} />)} />
        <Route exact path="/transactions/:organizationName/:transactionName" render={(props) => renderLoginIfNotLoggedIn(<TransactionEditPage account={account} {...props} />)} />
        <Route exact path="/webhooks" render={(props) => renderLoginIfNotLoggedIn(<WebhookListPage account={account} {...props} />)} />
        <Route exact path="/webhook-events" render={(props) => renderLoginIfNotLoggedIn(<WebhookEventListPage account={account} {...props} />)} />
        <Route exact path="/webhooks/:webhookName" render={(props) => renderLoginIfNotLoggedIn(<WebhookEditPage account={account} {...props} />)} />
        <Route exact path="/tickets" render={(props) => renderLoginIfNotLoggedIn(<TicketListPage account={account} {...props} />)} />
        <Route exact path="/tickets/:organizationName/:ticketName" render={(props) => renderLoginIfNotLoggedIn(<TicketEditPage account={account} {...props} />)} />
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

  const onClose = () => {
    setMenuVisible(false);
  };

  const showMenu = () => {
    setMenuVisible(true);
  };

  const siderWidth = 256;
  const siderCollapsedWidth = 80;
  const showSider = !Setting.isMobile() && !props.requiredEnableMfa;
  const contentMarginLeft = showSider ? (siderCollapsed ? siderCollapsedWidth : siderWidth) : 0;

  return (
    <React.Fragment>
      <EnableMfaNotification account={props.account} />
      {showSider && (
        <Sider
          collapsed={siderCollapsed}
          collapsedWidth={siderCollapsedWidth}
          width={siderWidth}
          trigger={null}
          theme={isDark ? "dark" : "light"}
          style={{
            height: "100vh",
            position: "fixed",
            left: 0,
            top: 0,
            bottom: 0,
            zIndex: 100,
            boxShadow: "2px 0 8px rgba(0,0,0,0.08)",
            display: "flex",
            flexDirection: "column",
          }}
        >
          <div style={{
            height: 52,
            flexShrink: 0,
            display: "flex",
            alignItems: "center",
            justifyContent: siderCollapsed ? "center" : "flex-start",
            padding: siderCollapsed ? "0" : "0 16px 0 24px",
            overflow: "hidden",
          }}>
            <Link to="/">
              <img
                src={siderCollapsed ? (organization?.favicon || siderLogo || props.logo) : (siderLogo ?? props.logo)}
                alt="logo"
                style={{
                  height: siderCollapsed ? 28 : 40,
                  width: siderCollapsed ? 28 : undefined,
                  maxWidth: siderCollapsed ? 28 : 160,
                  objectFit: "contain",
                  borderRadius: siderCollapsed ? 4 : 0,
                  transition: "max-width 0.2s, height 0.2s, width 0.2s",
                }}
              />
            </Link>
          </div>
          <div className="sider-menu-container" style={{flex: 1, overflow: "auto"}}>
            <Menu
              mode="inline"
              items={getMenuItems()}
              selectedKeys={[selectedLeafKey]}
              openKeys={menuOpenKeys}
              onOpenChange={setMenuOpenKeys}
              theme={isDark ? "dark" : "light"}
              style={{borderRight: 0}}
            />
          </div>
        </Sider>
      )}
      <div style={{marginLeft: contentMarginLeft, transition: "margin-left 0.2s", display: "flex", flexDirection: "column", minHeight: "100vh"}}>
        <Header style={{display: "flex", justifyContent: "space-between", alignItems: "center", padding: "0", marginBottom: "4px", backgroundColor: isDark ? "black" : "white", color: textColor, position: "sticky", top: 0, zIndex: 99, boxShadow: "0 1px 4px rgba(0,0,0,0.08)", height: "52px", lineHeight: "52px"}}>
          <div style={{display: "flex", alignItems: "center"}}>
            {props.requiredEnableMfa ? null : (Setting.isMobile() ? (
              <React.Fragment>
                <Drawer title={i18next.t("general:Close")} placement="left" open={menuVisible} onClose={onClose}>
                  <Menu
                    items={getMenuItems()}
                    mode={"inline"}
                    selectedKeys={[selectedLeafKey]}
                    openKeys={menuOpenKeys}
                    onOpenChange={setMenuOpenKeys}
                    style={{lineHeight: "48px"}}
                    onClick={onClose}
                  />
                </Drawer>
                <Button icon={<BarsOutlined />} onClick={showMenu} type="text">
                  {i18next.t("general:Menu")}
                </Button>
              </React.Fragment>
            ) : (
              <Button
                icon={siderCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                onClick={toggleSider}
                type="text"
                style={{fontSize: 16, width: 40, height: 40, color: textColor}}
              />
            ))}
            <BreadcrumbBar uri={currentUri} />
          </div>
          <div style={{flexShrink: 0, display: "flex", alignItems: "center"}}>
            {renderAccountMenu()}
          </div>
        </Header>
        <Content style={{display: "flex", flexDirection: "column"}}>
          {isWithoutCard() ?
            renderRouter() :
            <Card className="content-warp-card" styles={{body: {padding: 0, margin: 0}}}>
              {renderRouter()}
            </Card>
          }
        </Content>
      </div>
    </React.Fragment>
  );
}

export default withRouter(ManagementPage);
