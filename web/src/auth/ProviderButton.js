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
import SlackLoginButton from "./SlackLoginButton";
import SteamLoginButton from "./SteamLoginButton";
import BilibiliLoginButton from "./BilibiliLoginButton";
import OktaLoginButton from "./OktaLoginButton";
import DouyinLoginButton from "./DouyinLoginButton";
import * as AuthBackend from "./AuthBackend";

function getSigninButton(type) {
  const text = i18next.t("login:Sign in with {type}").replace("{type}", type);
  if (type === "GitHub") {
    return <GithubLoginButton text={text} align={"center"} />;
  } else if (type === "Google") {
    return <GoogleLoginButton text={text} align={"center"} />;
  } else if (type === "QQ") {
    return <QqLoginButton text={text} align={"center"} />;
  } else if (type === "Facebook") {
    return <FacebookLoginButton text={text} align={"center"} />;
  } else if (type === "Weibo") {
    return <WeiboLoginButton text={text} align={"center"} />;
  } else if (type === "Gitee") {
    return <GiteeLoginButton text={text} align={"center"} />;
  } else if (type === "WeChat") {
    return <WechatLoginButton text={text} align={"center"} />;
  } else if (type === "DingTalk") {
    return <DingTalkLoginButton text={text} align={"center"} />;
  } else if (type === "LinkedIn") {
    return <LinkedInLoginButton text={text} align={"center"} />;
  } else if (type === "WeCom") {
    return <WeComLoginButton text={text} align={"center"} />;
  } else if (type === "Lark") {
    return <LarkLoginButton text={text} align={"center"} />;
  } else if (type === "GitLab") {
    return <GitLabLoginButton text={text} align={"center"} />;
  } else if (type === "Adfs") {
    return <AdfsLoginButton text={text} align={"center"} />;
  } else if (type === "Casdoor") {
    return <CasdoorLoginButton text={text} align={"center"} />;
  } else if (type === "Baidu") {
    return <BaiduLoginButton text={text} align={"center"} />;
  } else if (type === "Alipay") {
    return <AlipayLoginButton text={text} align={"center"} />;
  } else if (type === "Infoflow") {
    return <InfoflowLoginButton text={text} align={"center"} />;
  } else if (type === "Apple") {
    return <AppleLoginButton text={text} align={"center"} />;
  } else if (type === "AzureAD") {
    return <AzureADLoginButton text={text} align={"center"} />;
  } else if (type === "Slack") {
    return <SlackLoginButton text={text} align={"center"} />;
  } else if (type === "Steam") {
    return <SteamLoginButton text={text} align={"center"} />;
  } else if (type === "Bilibili") {
    return <BilibiliLoginButton text={text} align={"center"} />;
  } else if (type === "Okta") {
    return <OktaLoginButton text={text} align={"center"} />;
  } else if (type === "Douyin") {
    return <DouyinLoginButton text={text} align={"center"} />;
  }

  return text;
}

function getSamlUrl(provider, location) {
  const params = new URLSearchParams(location.search);
  const clientId = params.get("client_id");
  const application = params.get("state");
  const realRedirectUri = params.get("redirect_uri");
  const redirectUri = `${window.location.origin}/callback/saml`;
  const providerName = provider.name;
  const relayState = `${clientId}&${application}&${providerName}&${realRedirectUri}&${redirectUri}`;
  AuthBackend.getSamlLogin(`${provider.owner}/${providerName}`, btoa(relayState)).then((res) => {
    if (res.data2 === "POST") {
      document.write(res.data);
    } else {
      window.location.href = res.data;
    }
  });
}

export function renderProviderLogo(provider, application, width, margin, size, location) {
  if (size === "small") {
    if (provider.category === "OAuth") {
      return (
        <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "signup")}>
          <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} style={{margin: margin}} />
        </a>
      );
    } else if (provider.category === "SAML") {
      return (
        <a key={provider.displayName} onClick={() => getSamlUrl(provider, location)}>
          <img width={width} height={width} src={getProviderLogoURL(provider)} alt={provider.displayName} style={{margin: margin}} />
        </a>
      );
    }

  } else if (provider.type === "Custom") {
    // style definition
    const text = i18next.t("login:Sign in with {type}").replace("{type}", provider.displayName);
    const customAStyle = {display: "block", height: "55px", color: "#000"};
    const customButtonStyle = {display: "flex", alignItems: "center", width: "calc(100% - 10px)", height: "50px", margin: "5px", padding: "0 10px", backgroundColor: "transparent", boxShadow: "0px 1px 3px rgba(0,0,0,0.5)", border: "0px", borderRadius: "3px", cursor: "pointer"};
    const customImgStyle = {justfyContent: "space-between"};
    const customSpanStyle = {textAlign: "center", lineHeight: "50px", width: "100%", fontSize: "19px"};
    if (provider.category === "OAuth") {
      return (
        <a key={provider.displayName} href={Provider.getAuthUrl(application, provider, "signup")} style={customAStyle}>
          <button style={customButtonStyle}>
            <img width={26} src={getProviderLogoURL(provider)} alt={provider.displayName} style={customImgStyle} />
            <span style={customSpanStyle}>{text}</span>
          </button>
        </a>
      );
    } else if (provider.category === "SAML") {
      return (
        <a key={provider.displayName} onClick={() => getSamlUrl(provider, location)} style={customAStyle}>
          <button style={customButtonStyle}>
            <img width={26} src={getProviderLogoURL(provider)} alt={provider.displayName} style={customImgStyle} />
            <span style={customSpanStyle}>{text}</span>
          </button>
        </a>
      );
    }
  } else {
    return (
      <div key={provider.displayName} style={{marginBottom: "10px"}}>
        <a href={Provider.getAuthUrl(application, provider, "signup")}>
          {
            getSigninButton(provider.type)
          }
        </a>
      </div>
    );
  }
}
