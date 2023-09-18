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
import {Tooltip} from "antd";
import * as Util from "./Util";
import * as Setting from "../Setting";

const authInfo = {
  Google: {
    scope: "profile+email",
    endpoint: "https://accounts.google.com/signin/oauth",
  },
  GitHub: {
    scope: "user:email+read:user",
    endpoint: "https://github.com/login/oauth/authorize",
  },
  QQ: {
    scope: "get_user_info",
    endpoint: "https://graph.qq.com/oauth2.0/authorize",
  },
  WeChat: {
    scope: "snsapi_login",
    endpoint: "https://open.weixin.qq.com/connect/qrconnect",
    mpScope: "snsapi_userinfo",
    mpEndpoint: "https://open.weixin.qq.com/connect/oauth2/authorize",
  },
  WeChatMiniProgram: {
    endpoint: "https://mp.weixin.qq.com/",
  },
  Facebook: {
    scope: "email,public_profile",
    endpoint: "https://www.facebook.com/dialog/oauth",
  },
  DingTalk: {
    scope: "openid",
    endpoint: "https://login.dingtalk.com/oauth2/auth",
  },
  Weibo: {
    scope: "email",
    endpoint: "https://api.weibo.com/oauth2/authorize",
  },
  Gitee: {
    scope: "user_info%20emails",
    endpoint: "https://gitee.com/oauth/authorize",
  },
  LinkedIn: {
    scope: "r_liteprofile%20r_emailaddress",
    endpoint: "https://www.linkedin.com/oauth/v2/authorization",
  },
  WeCom: {
    scope: "snsapi_userinfo",
    endpoint: "https://open.work.weixin.qq.com/wwopen/sso/3rd_qrConnect",
    silentEndpoint: "https://open.weixin.qq.com/connect/oauth2/authorize",
    internalEndpoint: "https://open.work.weixin.qq.com/wwopen/sso/qrConnect",
  },
  Lark: {
    // scope: "email",
    endpoint: "https://open.feishu.cn/open-apis/authen/v1/index",
  },
  GitLab: {
    scope: "read_user+profile",
    endpoint: "https://gitlab.com/oauth/authorize",
  },
  ADFS: {
    scope: "openid",
    endpoint: "http://example.com",
  },
  Baidu: {
    scope: "basic",
    endpoint: "http://openapi.baidu.com/oauth/2.0/authorize",
  },
  Alipay: {
    scope: "basic",
    endpoint: "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm",
  },
  Casdoor: {
    scope: "openid%20profile%20email",
    endpoint: "http://example.com",
  },
  Infoflow: {
    endpoint: "https://xpc.im.baidu.com/oauth2/authorize",
  },
  Apple: {
    scope: "name%20email",
    endpoint: "https://appleid.apple.com/auth/authorize",
  },
  AzureAD: {
    scope: "user.read",
    endpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
  },
  Slack: {
    scope: "users:read",
    endpoint: "https://slack.com/oauth/authorize",
  },
  Steam: {
    endpoint: "https://steamcommunity.com/openid/login",
  },
  Okta: {
    scope: "openid%20profile%20email",
    endpoint: "http://example.com",
  },
  Douyin: {
    scope: "user_info",
    endpoint: "https://open.douyin.com/platform/oauth/connect",
  },
  Custom: {
    endpoint: "https://example.com/",
  },
  Bilibili: {
    endpoint: "https://passport.bilibili.com/register/pc_oauth2.html",
  },
  Line: {
    scope: "profile%20openid%20email",
    endpoint: "https://access.line.me/oauth2/v2.1/authorize",
  },
  Amazon: {
    scope: "profile",
    endpoint: "https://www.amazon.com/ap/oa",
  },
  Auth0: {
    scope: "openid%20profile%20email",
    endpoint: "http://auth0.com/authorize",
  },
  BattleNet: {
    scope: "openid",
    endpoint: "https://oauth.battlenet.com.cn/authorize",
  },
  Bitbucket: {
    scope: "account",
    endpoint: "https://bitbucket.org/site/oauth2/authorize",
  },
  Box: {
    scope: "root_readwrite",
    endpoint: "https://account.box.com/api/oauth2/authorize",
  },
  CloudFoundry: {
    scope: "cloud_controller.read",
    endpoint: "https://login.cloudfoundry.org/oauth/authorize",
  },
  Dailymotion: {
    scope: "userinfo",
    endpoint: "https://api.dailymotion.com/oauth/authorize",
  },
  Deezer: {
    scope: "basic_access",
    endpoint: "https://connect.deezer.com/oauth/auth.php",
  },
  DigitalOcean: {
    scope: "read",
    endpoint: "https://cloud.digitalocean.com/v1/oauth/authorize",
  },
  Discord: {
    scope: "identify%20email",
    endpoint: "https://discord.com/api/oauth2/authorize",
  },
  Dropbox: {
    scope: "account_info.read",
    endpoint: "https://www.dropbox.com/oauth2/authorize",
  },
  EveOnline: {
    scope: "publicData",
    endpoint: "https://login.eveonline.com/oauth/authorize",
  },
  Fitbit: {
    scope: "activity%20heartrate%20location%20nutrition%20profile%20settings%20sleep%20social%20weight",
    endpoint: "https://www.fitbit.com/oauth2/authorize",
  },
  Gitea: {
    scope: "user:email",
    endpoint: "https://gitea.com/login/oauth/authorize",
  },
  Heroku: {
    scope: "global",
    endpoint: "https://id.heroku.com/oauth/authorize",
  },
  InfluxCloud: {
    scope: "read:org",
    endpoint: "https://cloud2.influxdata.com/oauth/authorize",
  },
  Instagram: {
    scope: "user_profile",
    endpoint: "https://api.instagram.com/oauth/authorize",
  },
  Intercom: {
    scope: "user.read",
    endpoint: "https://app.intercom.com/oauth",
  },
  Kakao: {
    scope: "account_email",
    endpoint: "https://kauth.kakao.com/oauth/authorize",
  },
  Lastfm: {
    scope: "user_read",
    endpoint: "https://www.last.fm/api/auth",
  },
  Mailru: {
    scope: "userinfo",
    endpoint: "https://oauth.mail.ru/login",
  },
  Meetup: {
    scope: "basic",
    endpoint: "https://secure.meetup.com/oauth2/authorize",
  },
  MicrosoftOnline: {
    scope: "openid%20profile%20email",
    endpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
  },
  Naver: {
    scope: "profile",
    endpoint: "https://nid.naver.com/oauth2.0/authorize",
  },
  Nextcloud: {
    scope: "openid%20profile%20email",
    endpoint: "https://cloud.example.org/apps/oauth2/authorize",
  },
  OneDrive: {
    scope: "offline_access%20onedrive.readonly",
    endpoint: "https://login.live.com/oauth20_authorize.srf",
  },
  Oura: {
    scope: "personal",
    endpoint: "https://cloud.ouraring.com/oauth/authorize",
  },
  Patreon: {
    scope: "identity",
    endpoint: "https://www.patreon.com/oauth2/authorize",
  },
  PayPal: {
    scope: "openid%20profile%20email",
    endpoint: "https://www.sandbox.paypal.com/connect",
  },
  SalesForce: {
    scope: "openid%20profile%20email",
    endpoint: "https://login.salesforce.com/services/oauth2/authorize",
  },
  Shopify: {
    scope: "read_products",
    endpoint: "https://myshopify.com/admin/oauth/authorize",
  },
  Soundcloud: {
    scope: "non-expiring",
    endpoint: "https://api.soundcloud.com/connect",
  },
  Spotify: {
    scope: "user-read-email",
    endpoint: "https://accounts.spotify.com/authorize",
  },
  Strava: {
    scope: "read",
    endpoint: "https://www.strava.com/oauth/authorize",
  },
  Stripe: {
    scope: "read_only",
    endpoint: "https://connect.stripe.com/oauth/authorize",
  },
  TikTok: {
    scope: "user.info.basic",
    endpoint: "https://www.tiktok.com/auth/authorize/",
  },
  Tumblr: {
    scope: "email",
    endpoint: "https://www.tumblr.com/oauth2/authorize",
  },
  Twitch: {
    scope: "user_read",
    endpoint: "https://id.twitch.tv/oauth2/authorize",
  },
  Twitter: {
    scope: "users.read",
    endpoint: "https://twitter.com/i/oauth2/authorize",
  },
  Typetalk: {
    scope: "my",
    endpoint: "https://typetalk.com/oauth2/authorize",
  },
  Uber: {
    scope: "profile",
    endpoint: "https://login.uber.com/oauth/v2/authorize",
  },
  VK: {
    scope: "email",
    endpoint: "https://oauth.vk.com/authorize",
  },
  Wepay: {
    scope: "manage_accounts%20view_user",
    endpoint: "https://www.wepay.com/v2/oauth2/authorize",
  },
  Xero: {
    scope: "openid%20profile%20email",
    endpoint: "https://login.xero.com/identity/connect/authorize",
  },
  Yahoo: {
    scope: "openid%20profile%20email",
    endpoint: "https://api.login.yahoo.com/oauth2/request_auth",
  },
  Yammer: {
    scope: "user",
    endpoint: "https://www.yammer.com/oauth2/authorize",
  },
  Yandex: {
    scope: "login:email",
    endpoint: "https://oauth.yandex.com/authorize",
  },
  Zoom: {
    scope: "user:read",
    endpoint: "https://zoom.us/oauth/authorize",
  },
  MetaMask: {
    scope: "",
    endpoint: "",
  },
  Web3Onboard: {
    scope: "",
    endpoint: "",
  },
};

export function getProviderUrl(provider) {
  if (provider.category === "OAuth") {
    const endpoint = authInfo[provider.type].endpoint;
    const urlObj = new URL(endpoint);

    let host = urlObj.host;
    let tokens = host.split(".");
    if (tokens.length > 2) {
      tokens = tokens.slice(1);
    }
    host = tokens.join(".");

    return `${urlObj.protocol}//${host}`;
  } else {
    const info = Setting.OtherProviderInfo[provider.category][provider.type];
    // avoid crash when provider is not found
    if (info) {
      return info.url;
    }
    return "";
  }
}

export function getProviderLogoWidget(provider) {
  if (provider === undefined) {
    return null;
  }

  const url = getProviderUrl(provider);
  if (url !== "") {
    return (
      <Tooltip title={provider.type}>
        <a target="_blank" rel="noreferrer" href={getProviderUrl(provider)}>
          <img width={36} height={36} src={Setting.getProviderLogoURL(provider)} alt={provider.displayName} />
        </a>
      </Tooltip>
    );
  } else {
    return (
      <Tooltip title={provider.type}>
        <img width={36} height={36} src={Setting.getProviderLogoURL(provider)} alt={provider.displayName} />
      </Tooltip>
    );
  }
}

export function getAuthUrl(application, provider, method) {
  if (application === null || provider === null) {
    return "";
  }

  let endpoint = authInfo[provider.type].endpoint;
  let redirectUri = `${window.location.origin}/callback`;
  const scope = authInfo[provider.type].scope;

  const isShortState = provider.type === "WeChat" && navigator.userAgent.includes("MicroMessenger");
  const state = Util.getStateFromQueryParams(application.name, provider.name, method, isShortState);
  const codeChallenge = "P3S-a7dr8bgM4bF6vOyiKkKETDl16rcAzao9F8UIL1Y"; // SHA256(Base64-URL-encode("casdoor-verifier"))

  if (provider.type === "AzureAD") {
    if (provider.domain !== "") {
      endpoint = endpoint.replace("common", provider.domain);
    }
  } else if (provider.type === "Apple") {
    redirectUri = `${window.location.origin}/api/callback`;
  }

  if (provider.type === "Google" || provider.type === "GitHub" || provider.type === "QQ" || provider.type === "Facebook"
    || provider.type === "Weibo" || provider.type === "Gitee" || provider.type === "LinkedIn" || provider.type === "GitLab" || provider.type === "AzureAD"
    || provider.type === "Slack" || provider.type === "Line" || provider.type === "Amazon" || provider.type === "Auth0" || provider.type === "BattleNet"
    || provider.type === "Bitbucket" || provider.type === "Box" || provider.type === "CloudFoundry" || provider.type === "Dailymotion"
    || provider.type === "DigitalOcean" || provider.type === "Discord" || provider.type === "Dropbox" || provider.type === "EveOnline" || provider.type === "Gitea"
    || provider.type === "Heroku" || provider.type === "InfluxCloud" || provider.type === "Instagram" || provider.type === "Intercom" || provider.type === "Kakao"
    || provider.type === "MailRu" || provider.type === "Meetup" || provider.type === "MicrosoftOnline" || provider.type === "Naver" || provider.type === "Nextcloud"
    || provider.type === "OneDrive" || provider.type === "Oura" || provider.type === "Patreon" || provider.type === "PayPal" || provider.type === "SalesForce"
    || provider.type === "SoundCloud" || provider.type === "Spotify" || provider.type === "Strava" || provider.type === "Stripe" || provider.type === "Tumblr"
    || provider.type === "Twitch" || provider.type === "Typetalk" || provider.type === "Uber" || provider.type === "VK" || provider.type === "Wepay"
    || provider.type === "Xero" || provider.type === "Yahoo" || provider.type === "Yammer" || provider.type === "Yandex" || provider.type === "Zoom") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "DingTalk") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&prompt=consent&state=${state}`;
  } else if (provider.type === "WeChat") {
    if (navigator.userAgent.includes("MicroMessenger")) {
      return `${authInfo[provider.type].mpEndpoint}?appid=${provider.clientId2}&redirect_uri=${redirectUri}&state=${state}&scope=${authInfo[provider.type].mpScope}&response_type=code#wechat_redirect`;
    } else {
      return `${endpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}#wechat_redirect`;
    }
  } else if (provider.type === "WeCom") {
    if (provider.subType === "Internal") {
      if (provider.method === "Silent") {
        endpoint = authInfo[provider.type].silentEndpoint;
        return `${endpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&scope=${scope}&response_type=code#wechat_redirect`;
      } else if (provider.method === "Normal") {
        endpoint = authInfo[provider.type].internalEndpoint;
        return `${endpoint}?appid=${provider.clientId}&agentid=${provider.appId}&redirect_uri=${redirectUri}&state=${state}&usertype=member`;
      } else {
        return `https://error:not-supported-provider-method:${provider.method}`;
      }
    } else if (provider.subType === "Third-party") {
      if (provider.method === "Silent") {
        endpoint = authInfo[provider.type].silentEndpoint;
        return `${endpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&scope=${scope}&response_type=code#wechat_redirect`;
      } else if (provider.method === "Normal") {
        return `${endpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&usertype=member`;
      } else {
        return `https://error:not-supported-provider-method:${provider.method}`;
      }
    } else {
      return `https://error:not-supported-provider-sub-type:${provider.subType}`;
    }
  } else if (provider.type === "Lark") {
    return `${endpoint}?app_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}`;
  } else if (provider.type === "ADFS") {
    return `${provider.domain}/adfs/oauth2/authorize?client_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&response_type=code&nonce=casdoor&scope=openid`;
  } else if (provider.type === "Baidu") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&response_type=code&scope=${scope}&display=popup`;
  } else if (provider.type === "Alipay") {
    return `${endpoint}?app_id=${provider.clientId}&scope=auth_user&redirect_uri=${redirectUri}&state=${state}&response_type=code&scope=${scope}&display=popup`;
  } else if (provider.type === "Casdoor") {
    return `${provider.domain}/login/oauth/authorize?client_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&response_type=code&scope=${scope}`;
  } else if (provider.type === "Infoflow") {
    return `${endpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}?state=${state}`;
  } else if (provider.type === "Apple") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&response_type=code%20id_token&scope=${scope}&response_mode=form_post`;
  } else if (provider.type === "Steam") {
    return `${endpoint}?openid.claimed_id=http://specs.openid.net/auth/2.0/identifier_select&openid.identity=http://specs.openid.net/auth/2.0/identifier_select&openid.mode=checkid_setup&openid.ns=http://specs.openid.net/auth/2.0&openid.realm=${window.location.origin}&openid.return_to=${redirectUri}?state=${state}`;
  } else if (provider.type === "Okta") {
    return `${provider.domain}/v1/authorize?client_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&response_type=code&scope=${scope}`;
  } else if (provider.type === "Douyin" || provider.type === "TikTok") {
    return `${endpoint}?client_key=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&response_type=code&scope=${scope}`;
  } else if (provider.type === "Custom") {
    return `${provider.customAuthUrl}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${provider.scopes}&response_type=code&state=${state}`;
  } else if (provider.type === "Bilibili") {
    return `${endpoint}#/?client_id=${provider.clientId}&return_url=${redirectUri}&state=${state}&response_type=code`;
  } else if (provider.type === "Deezer") {
    return `${endpoint}?app_id=${provider.clientId}&redirect_uri=${redirectUri}&perms=${scope}`;
  } else if (provider.type === "Lastfm") {
    return `${endpoint}?api_key=${provider.clientId}&cb=${redirectUri}`;
  } else if (provider.type === "Shopify") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&state=${state}&grant_options[]=per-user`;
  } else if (provider.type === "Twitter" || provider.type === "Fitbit") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&response_type=code&scope=${scope}&code_challenge=${codeChallenge}&code_challenge_method=S256`;
  } else if (provider.type === "MetaMask") {
    return `${redirectUri}?state=${state}`;
  } else if (provider.type === "Web3Onboard") {
    return `${redirectUri}?state=${state}`;
  }
}
