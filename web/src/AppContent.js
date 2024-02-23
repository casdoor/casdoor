import * as Setting from "./Setting";
import {Avatar, Button, Card, Drawer, Dropdown, Menu, Result, Tooltip} from "antd";
import EnableMfaNotification from "./common/notifaction/EnableMfaNotification";
import {Link, Redirect, Route, Switch} from "react-router-dom";
import React, {Suspense} from "react";
import i18next from "i18next";
import {
  AppstoreTwoTone,
  BarsOutlined,
  DeploymentUnitOutlined, DollarTwoTone, DownOutlined,
  HomeTwoTone,
  LockTwoTone, LogoutOutlined,
  SafetyCertificateTwoTone, SettingOutlined, SettingTwoTone, WalletTwoTone
} from "@ant-design/icons";
import {Content, Header} from "antd/es/layout/layout";
import indexRouters from "./routers";
import LanguageSelect from "./common/select/LanguageSelect";
import ThemeSelect from "./common/select/ThemeSelect";
import OpenTour from "./common/OpenTour";
import OrganizationSelect from "./common/select/OrganizationSelect";
import * as Conf from "./Conf";
import AccountAvatar from "./account/AccountAvatar";
import {useDispatch, useSelector} from "react-redux";
import {setLogo, setThemeAlgorithm} from "./store/themeSlice";
import * as AuthBackend from "./auth/AuthBackend";
import {clearWeb3AuthToken} from "./auth/Web3Auth";
import {setAccount} from "./store/accountSlice";

function AppContent(props) {

  const account = useSelector((state) => state.account.value);
  const themeAlgorithm = useSelector((state) => state.theme.themeAlgorithm);
  const logo = useSelector((state) => state.theme.logo);

  const dispatch = useDispatch();

  function isWithoutCard() {
    return Setting.isMobile() || window.location.pathname.startsWith("/trees");
  }

  function renderLoginIfNotLoggedIn(component) {
    if (account === null) {
      sessionStorage.setItem("from", window.location.pathname);
      return <Redirect to="/login" />;
    } else if (account === undefined) {
      return null;
    } else {
      return component;
    }
  }

  function logout() {

    props.setExpiredAndSubmitted();

    AuthBackend.logout()
      .then((res) => {
        if (res.status === "ok") {
          const owner = account.owner;

          dispatch(setThemeAlgorithm(["default"]));
          dispatch(setAccount(null));

          clearWeb3AuthToken();
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

  function renderRouter() {
    return (
      <Suspense fallback={<div></div>}>
        <Switch>
          {indexRouters.map((el, index) => {
            const Element = el.component;
            return (<Route key="" exact={el.exact} path={el.path} render={(props) => {
              if (el.path === "/mfa/setup") {
                props.onfinish = () => {this.props.setRequireMfaFalse();};
                props.account = account;
                return renderLoginIfNotLoggedIn(<Element {...props} />);
              }
              if (el.auth) {
                props.account = account;
                return renderLoginIfNotLoggedIn(<Element {...props} />);
              } else {
                return <Element />;
              }
            }} />);
          })}
          <Route exact path="" render={(props) =>
            <Result status="404" title="404 NOT FOUND" subTitle={i18next.t("general:Sorry, the page you visited does not exist.")} extra={<a href="/web/public"><Button type="primary">{i18next.t("general:Back Home")}</Button></a>} />}
          />
        </Switch>
      </Suspense>
    );
  }

  function renderAvatar() {
    if (account.avatar === "") {
      return (
        <Avatar style={{backgroundColor: Setting.getAvatarColor(account.name), verticalAlign: "middle"}} size="large">
          {Setting.getShortName(account.name)}
        </Avatar>
      );
    } else {
      return (
        <Avatar src={account.avatar} style={{verticalAlign: "middle"}} size="large"
          icon={<AccountAvatar src={account.avatar} style={{verticalAlign: "middle"}} size={40} />}
        >
          {Setting.getShortName(account.name)}
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
          {Setting.isMobile() ? null : Setting.getShortText(Setting.getNameAtLeast(account.displayName), 30)} &nbsp; <DownOutlined />
                  &nbsp;
                  &nbsp;
                  &nbsp;
        </div>
      </Dropdown>
    );
  }

  function renderAccountMenu() {
    if (account === undefined) {
      return null;
    } else if (account === null) {
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
            themeAlgorithm={themeAlgorithm}
            onChange={(nextThemeAlgorithm) => {
              // props.setThemeAlgorithm(nextThemeAlgorithm);
              dispatch(setThemeAlgorithm(nextThemeAlgorithm));
              // props.setLogo(nextThemeAlgorithm);
              dispatch(setLogo(nextThemeAlgorithm));
            }} />
          <LanguageSelect languages={account.organization.languages} />
          <Tooltip title="Click to open AI assitant">
            <div className="select-box" onClick={() => {
              props.OpenAiAssistant();
            }}>
              <DeploymentUnitOutlined style={{fontSize: "24px", color: "rgb(77,77,77)"}} />
            </div>
          </Tooltip>
          <OpenTour />
          {Setting.isAdminUser(account) && !Setting.isMobile() && (props.uri.indexOf("/trees") === -1) &&
                      <OrganizationSelect
                        initValue={Setting.getOrganization()}
                        withAll={true}
                        style={{marginRight: "20px", width: "180px", display: "flex"}}
                        onChange={(value) => {
                          Setting.setOrganization(value);
                        }}
                        className="select-box"
                      />
          }
        </React.Fragment>
      );
    }
  }

  function getMenuItems() {
    const res = [];

    if (account === null || account === undefined) {
      return [];
    }

    res.push(Setting.getItem(<Link to="/">{i18next.t("general:Home")}</Link>, "/home", <HomeTwoTone />, [
      Setting.getItem(<Link to="/">{i18next.t("general:Dashboard")}</Link>, "/"),
      Setting.getItem(<Link to="/shortcuts">{i18next.t("general:Shortcuts")}</Link>, "/shortcuts"),
      Setting.getItem(<Link to="/apps">{i18next.t("general:Apps")}</Link>, "/apps"),
    ].filter(item => {
      return Setting.isLocalAdminUser(account);
    })));

    if (Setting.isLocalAdminUser(account)) {
      if (Conf.ShowGithubCorner) {
        res.push(Setting.getItem(<a href={"https://casdoor.com"}>
          <span style={{fontWeight: "bold", backgroundColor: "rgba(87,52,211,0.4)", marginTop: "12px", paddingLeft: "5px", paddingRight: "5px", display: "flex", alignItems: "center", height: "40px", borderRadius: "5px"}}>
            🚀 SaaS Hosting 🔥
          </span>
        </a>, "#"));
      }

      res.push(Setting.getItem(<Link style={{color: "black"}} to="/organizations">{i18next.t("general:User Management")}</Link>, "/orgs", <AppstoreTwoTone />, [
        Setting.getItem(<Link to="/organizations">{i18next.t("general:Organizations")}</Link>, "/organizations"),
        Setting.getItem(<Link to="/groups">{i18next.t("general:Groups")}</Link>, "/groups"),
        Setting.getItem(<Link to="/users">{i18next.t("general:Users")}</Link>, "/users"),
        Setting.getItem(<Link to="/invitations">{i18next.t("general:Invitations")}</Link>, "/invitations"),
      ]));

      res.push(Setting.getItem(<Link style={{color: "black"}} to="/applications">{i18next.t("general:Identity")}</Link>, "/identity", <LockTwoTone />, [
        Setting.getItem(<Link to="/applications">{i18next.t("general:Applications")}</Link>, "/applications"),
        Setting.getItem(<Link to="/providers">{i18next.t("general:Providers")}</Link>, "/providers"),
        Setting.getItem(<Link to="/resources">{i18next.t("general:Resources")}</Link>, "/resources"),
        Setting.getItem(<Link to="/certs">{i18next.t("general:Certs")}</Link>, "/certs"),
      ]));

      res.push(Setting.getItem(<Link style={{color: "black"}} to="/roles">{i18next.t("general:Authorization")}</Link>, "/auth", <SafetyCertificateTwoTone />, [
        Setting.getItem(<Link to="/roles">{i18next.t("general:Roles")}</Link>, "/roles"),
        Setting.getItem(<Link to="/permissions">{i18next.t("general:Permissions")}</Link>, "/permissions"),
        Setting.getItem(<Link to="/models">{i18next.t("general:Models")}</Link>, "/models"),
        Setting.getItem(<Link to="/adapters">{i18next.t("general:Adapters")}</Link>, "/adapters"),
        Setting.getItem(<Link to="/enforcers">{i18next.t("general:Enforcers")}</Link>, "/enforcers"),
      ].filter(item => {
        if (!Setting.isLocalAdminUser(account) && ["/models", "/adapters", "/enforcers"].includes(item.key)) {
          return false;
        } else {
          return true;
        }
      })));

      res.push(Setting.getItem(<Link style={{color: "black"}} to="/sessions">{i18next.t("general:Logging & Auditing")}</Link>, "/logs", <WalletTwoTone />, [
        Setting.getItem(<Link to="/sessions">{i18next.t("general:Sessions")}</Link>, "/sessions"),
        Conf.CasvisorUrl ? Setting.getItem(<a target="_blank" rel="noreferrer" href={Conf.CasvisorUrl}>{i18next.t("general:Records")}</a>, "/records")
          : Setting.getItem(<Link to="/records">{i18next.t("general:Records")}</Link>, "/records"),
        Setting.getItem(<Link to="/tokens">{i18next.t("general:Tokens")}</Link>, "/tokens"),
      ]));

      res.push(Setting.getItem(<Link style={{color: "black"}} to="/products">{i18next.t("general:Business & Payments")}</Link>, "/business", <DollarTwoTone />, [
        Setting.getItem(<Link to="/products">{i18next.t("general:Products")}</Link>, "/products"),
        Setting.getItem(<Link to="/payments">{i18next.t("general:Payments")}</Link>, "/payments"),
        Setting.getItem(<Link to="/plans">{i18next.t("general:Plans")}</Link>, "/plans"),
        Setting.getItem(<Link to="/pricings">{i18next.t("general:Pricings")}</Link>, "/pricings"),
        Setting.getItem(<Link to="/subscriptions">{i18next.t("general:Subscriptions")}</Link>, "/subscriptions"),
      ]));

      if (Setting.isAdminUser(account)) {
        res.push(Setting.getItem(<Link style={{color: "black"}} to="/sysinfo">{i18next.t("general:Admin")}</Link>, "/admin", <SettingTwoTone />, [
          Setting.getItem(<Link to="/sysinfo">{i18next.t("general:System Info")}</Link>, "/sysinfo"),
          Setting.getItem(<Link to="/syncers">{i18next.t("general:Syncers")}</Link>, "/syncers"),
          Setting.getItem(<Link to="/webhooks">{i18next.t("general:Webhooks")}</Link>, "/webhooks"),
          Setting.getItem(<a target="_blank" rel="noreferrer" href={Setting.isLocalhost() ? `${Setting.ServerUrl}/swagger` : "/swagger"}>{i18next.t("general:Swagger")}</a>, "/swagger")]));
      } else {
        res.push(Setting.getItem(<Link style={{color: "black"}} to="/syncers">{i18next.t("general:Admin")}</Link>, "/admin", <SettingTwoTone />, [
          Setting.getItem(<Link to="/syncers">{i18next.t("general:Syncers")}</Link>, "/syncers"),
          Setting.getItem(<Link to="/webhooks">{i18next.t("general:Webhooks")}</Link>, "/webhooks")]));
      }
    }

    return res;
  }

  const onClick = ({key}) => {
    if (key !== "/swagger" && key !== "/records") {
      if (props.requiredEnableMfa) {
        Setting.showMessage("info", "Please enable MFA first!");
      } else {
        props.history.push(key);
      }
    }
  };
  const menuStyleRight = Setting.isAdminUser(account) && !Setting.isMobile() ? "calc(180px + 280px)" : "280px";

  return (
    <React.Fragment>
      <EnableMfaNotification account={account} />
      <Header style={{padding: "0", marginBottom: "3px", backgroundColor: themeAlgorithm.includes("dark") ? "black" : "white"}} >
        {Setting.isMobile() ? null : (
          <Link to={"/"}>
            <div className="logo" style={{background: `url(${logo})`}} />
          </Link>
        )}
        {props.requiredEnableMfa || (Setting.isMobile() ?
          <React.Fragment>
            <Drawer title={i18next.t("general:Close")} placement="left" visible={props.menuVisible} onClose={props.onClose}>
              <Menu
                items={getMenuItems()}
                mode={"inline"}
                selectedKeys={[props.selectedMenuKey]}
                style={{lineHeight: "64px"}}
                onClick={props.onClose}
              >
              </Menu>
            </Drawer>
            <Button icon={<BarsOutlined />} onClick={props.showMenu} type="text">
              {i18next.t("general:Menu")}
            </Button>
          </React.Fragment> :
          <Menu
            onClick={onClick}
            items={getMenuItems()}
            mode={"horizontal"}
            selectedKeys={[props.selectedMenuKey]}
            style={{position: "absolute", left: "145px", right: menuStyleRight}}
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

export default AppContent;
