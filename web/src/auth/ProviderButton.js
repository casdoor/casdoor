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

import React from "react";
import i18next from "i18next";
import * as Provider from "./Provider";
import {getProviderLogoURL} from "../Setting";
import {GithubLoginButton, GoogleLoginButton} from "react-social-login-buttons";
import QqLoginButton from "./QqLoginButton";
import FacebookLoginButton from "./FacebookLoginButton";
import WeiboLoginButton from "./WeiboLoginButton";
import GiteeLoginButton from "./GiteeLoginButton";
import WechatLoginButton from "./WechatLoginButton";
import DingTalkLoginButton from "./DingTalkLoginButton";
import LinkedInLoginButton from "./LinkedInLoginButton";
import WeComLoginButton from "./WeComLoginButton";
import LarkLoginButton from "./LarkLoginButton";
import GitLabLoginButton from "./GitLabLoginButton";
import AdfsLoginButton from "./AdfsLoginButton";
import CasdoorLoginButton from "./CasdoorLoginButton";
import BaiduLoginButton from "./BaiduLoginButton";
import AlipayLoginButton from "./AlipayLoginButton";
import InfoflowLoginButton from "./InfoflowLoginButton";
import AppleLoginButton from "./AppleLoginButton";
import AzureADLoginButton from "./AzureADLoginButton";
import AzureADB2CLoginButton from "./AzureADB2CLoginButton";
import SlackLoginButton from "./SlackLoginButton";
import SteamLoginButton from "./SteamLoginButton";
import BilibiliLoginButton from "./BilibiliLoginButton";
import OktaLoginButton from "./OktaLoginButton";
import DouyinLoginButton from "./DouyinLoginButton";
import KwaiLoginButton from "./KwaiLoginButton";
import LoginButton from "./LoginButton";
import * as AuthBackend from "./AuthBackend";
import {WechatOfficialAccountModal} from "./Util";
import * as Setting from "../Setting";

function getSigninButton(provider) {
  const text = i18next.t("login:Sign in with {type}").replace("{type}", provider.displayName !== "" ? provider.displayName : provider.type);
  if (provider.type === "GitHub") {
    return <GithubLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Google") {
    return <GoogleLoginButton text={text} align={"center"} />;
  } else if (provider.type === "QQ") {
    return <QqLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Facebook") {
    return <FacebookLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Weibo") {
    return <WeiboLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Gitee") {
    return <GiteeLoginButton text={text} align={"center"} />;
  } else if (provider.type === "WeChat") {
    return <WechatLoginButton text={text} align={"center"} />;
  } else if (provider.type === "DingTalk") {
    return <DingTalkLoginButton text={text} align={"center"} />;
  } else if (provider.type === "LinkedIn") {
    return <LinkedInLoginButton text={text} align={"center"} />;
  } else if (provider.type === "WeCom") {
    return <WeComLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Lark") {
    return <LarkLoginButton text={text} align={"center"} />;
  } else if (provider.type === "GitLab") {
    return <GitLabLoginButton text={text} align={"center"} />;
  } else if (provider.type === "ADFS") {
    return <AdfsLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Casdoor") {
    return <CasdoorLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Baidu") {
    return <BaiduLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Alipay") {
    return <AlipayLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Infoflow") {
    return <InfoflowLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Apple") {
    return <AppleLoginButton text={text} align={"center"} />;
  } else if (provider.type === "AzureAD") {
    return <AzureADLoginButton text={text} align={"center"} />;
  } else if (provider.type === "AzureADB2C") {
    return <AzureADB2CLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Slack") {
    return <SlackLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Steam") {
    return <SteamLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Bilibili") {
    return <BilibiliLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Okta") {
    return <OktaLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Douyin") {
    return <DouyinLoginButton text={text} align={"center"} />;
  } else if (provider.type === "Kwai") {
    return <KwaiLoginButton text={text} align={"center"} />;
  } else {
    return <LoginButton key={provider.type} type={provider.type} logoUrl={getProviderLogoURL(provider)} />;
  }
}

function goToSamlUrl(provider, location) {
  const params = new URLSearchParams(location.search);
  const clientId = params.get("client_id") ?? "";
  const state = params.get("state");
  const realRedirectUri = params.get("redirect_uri");
  const redirectUri = `${window.location.origin}/callback/saml`;
  const providerName = provider.name;

  const relayState = `${clientId}&${state}&${providerName}&${realRedirectUri}&${redirectUri}`;
  AuthBackend.getSamlLogin(`${provider.owner}/${providerName}`, btoa(relayState)).then((res) => {
    if (res.status === "ok") {
      if (res.data2 === "POST") {
        document.write(res.data);
      } else {
        window.location.href = res.data;
      }
    } else {
      Setting.showMessage("error", res.msg);
    }
  });
}

export function goToWeb3Url(application, provider, method) {
  if (provider.type === "MetaMask") {
    import("./Web3Auth")
      .then(module => {
        const authViaMetaMask = module.authViaMetaMask;
        authViaMetaMask(application, provider, method);
      });
  } else if (provider.type === "Web3Onboard") {
    import("./Web3Auth")
      .then(module => {
        const authViaWeb3Onboard = module.authViaWeb3Onboard;
        authViaWeb3Onboard(application, provider, method);
      });
  }
}

export function renderProviderLogo(provider, application, width, margin, size, location) {
  if (size === "small") {
    if (provider.category === "OAuth") {
      if (provider.type === "WeChat" && provider.clientId2 !== "" && provider.clientSecret2 !== "" && provider.disableSsl === true && !navigator.userAgent.includes("MicroMessenger")) {
        return (
          <a key={provider.displayName} >
            <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={{margin: margin}} onClick={() => {
              WechatOfficialAccountModal(application, provider, "signup");
            }} />
          </a>
        );
      } else {
        return (
          <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "signup")}>
            <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={{margin: margin}} />
          </a>
        );
      }
    } else if (provider.category === "SAML") {
      return (
        <a key={provider.displayName} onClick={() => goToSamlUrl(provider, location)}>
          <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={{margin: margin}} />
        </a>
      );
    } else if (provider.category === "Web3") {
      return (
        <a key={provider.displayName} onClick={() => goToWeb3Url(application, provider, "signup")}>
          <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={{margin: margin}} />
        </a>
      );
    }
  } else if (provider.type.startsWith("Custom")) {
    // style definition
    const text = i18next.t("login:Sign in with {type}").replace("{type}", provider.displayName);
    const customAStyle = {display: "block", height: "55px", color: "#000"};
    const customButtonStyle = {display: "flex", alignItems: "center", width: "calc(100% - 10px)", height: "50px", margin: "5px", padding: "0 10px", backgroundColor: "transparent", boxShadow: "0px 1px 3px rgba(0,0,0,0.5)", border: "0px", borderRadius: "3px", cursor: "pointer"};
    const customImgStyle = {justfyContent: "space-between"};
    const customSpanStyle = {textAlign: "center", width: "100%", fontSize: "19px"};
    if (provider.category === "OAuth") {
      return (
        <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "signup")} style={customAStyle}>
          <div style={customButtonStyle}>
            <img width={26} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={customImgStyle} />
            <span style={customSpanStyle}>{text}</span>
          </div>
        </a>
      );
    } else if (provider.category === "SAML") {
      return (
        <a key={provider.displayName} onClick={() => goToSamlUrl(provider, location)} style={customAStyle}>
          <div style={customButtonStyle}>
            <img width={26} src={getProviderLogoURL(provider)} alt={provider.displayName} className="provider-img" style={customImgStyle} />
            <span style={customSpanStyle}>{text}</span>
          </div>
        </a>
      );
    }
  } else {
    // big button, for disable password signin
    if (provider.category === "SAML") {
      return (
        <div key={provider.displayName} className="provider-big-img">
          <a onClick={() => goToSamlUrl(provider, location)}>
            {
              getSigninButton(provider)
            }
          </a>
        </div>
      );
    } else if (provider.category === "Web3") {
      return (
        <div key={provider.displayName} className="provider-big-img">
          <a onClick={() => goToWeb3Url(application, provider, "signup")}>
            {
              getSigninButton(provider)
            }
          </a>
        </div>
      );
    } else {
      return (
        <div key={provider.displayName} className="provider-big-img">
          <a href={Provider.getAuthUrl(application, provider, "signup")}>
            {
              getSigninButton(provider)
            }
          </a>
        </div>
      );
    }
  }
}
