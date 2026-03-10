// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

(function() {
  "use strict";

  function getFallbackUrl() {
    var url = new URL(window.location.href);
    url.searchParams.delete("provider_hint");
    return url.pathname + url.search + url.hash;
  }

  function redirectToFallback() {
    window.location.replace(getFallbackUrl());
  }

  function getAcceptLanguage() {
    return localStorage.getItem("language") || navigator.language || "en";
  }

  function isProviderVisible(providerItem) {
    if (!providerItem || !providerItem.provider) {
      return false;
    }

    if (["OAuth", "SAML", "Web3"].indexOf(providerItem.provider.category) === -1) {
      return false;
    }

    if (providerItem.provider.type === "WeChatMiniProgram") {
      return false;
    }

    return true;
  }

  function isProviderVisibleForSignIn(providerItem) {
    if (providerItem.canSignIn === false) {
      return false;
    }

    return isProviderVisible(providerItem);
  }

  function base64UrlEncode(buffer) {
    var binary = "";
    for (var index = 0; index < buffer.length; index++) {
      binary += String.fromCharCode(buffer[index]);
    }

    return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
  }

  function generateCodeVerifier() {
    var array = new Uint8Array(32);
    window.crypto.getRandomValues(array);
    return base64UrlEncode(array);
  }

  async function generateCodeChallenge(verifier) {
    var data = new TextEncoder().encode(verifier);
    var digest = await window.crypto.subtle.digest("SHA-256", data);
    return base64UrlEncode(new Uint8Array(digest));
  }

  function storeCodeVerifier(state, verifier) {
    localStorage.setItem("pkce_verifier_" + state, verifier);
  }

  var authInfo = {
    Google: {scope: "profile+email", endpoint: "https://accounts.google.com/signin/oauth"},
    GitHub: {scope: "user:email+read:user", endpoint: "https://github.com/login/oauth/authorize"},
    QQ: {scope: "get_user_info", endpoint: "https://graph.qq.com/oauth2.0/authorize"},
    WeChat: {scope: "snsapi_login", endpoint: "https://open.weixin.qq.com/connect/qrconnect", mpScope: "snsapi_userinfo", mpEndpoint: "https://open.weixin.qq.com/connect/oauth2/authorize"},
    WeChatMiniProgram: {endpoint: "https://mp.weixin.qq.com/"},
    Facebook: {scope: "email,public_profile", endpoint: "https://www.facebook.com/dialog/oauth"},
    DingTalk: {scope: "openid", endpoint: "https://login.dingtalk.com/oauth2/auth"},
    Weibo: {scope: "email", endpoint: "https://api.weibo.com/oauth2/authorize"},
    Gitee: {scope: "user_info%20emails", endpoint: "https://gitee.com/oauth/authorize"},
    LinkedIn: {scope: "r_liteprofile%20r_emailaddress", endpoint: "https://www.linkedin.com/oauth/v2/authorization"},
    WeCom: {scope: "snsapi_userinfo", endpoint: "https://login.work.weixin.qq.com/wwlogin/sso/login", silentEndpoint: "https://open.weixin.qq.com/connect/oauth2/authorize", internalEndpoint: "https://login.work.weixin.qq.com/wwlogin/sso/login"},
    Lark: {endpoint: "https://open.feishu.cn/open-apis/authen/v1/index", endpoint2: "https://accounts.larksuite.com/open-apis/authen/v1/authorize"},
    GitLab: {scope: "read_user+profile", endpoint: "https://gitlab.com/oauth/authorize"},
    ADFS: {scope: "openid", endpoint: "http://example.com"},
    Baidu: {scope: "basic", endpoint: "http://openapi.baidu.com/oauth/2.0/authorize"},
    Alipay: {scope: "basic", endpoint: "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm"},
    Casdoor: {scope: "openid%20profile%20email", endpoint: "http://example.com"},
    Infoflow: {endpoint: "https://xpc.im.baidu.com/oauth2/authorize"},
    Apple: {scope: "name%20email", endpoint: "https://appleid.apple.com/auth/authorize"},
    AzureAD: {scope: "user.read", endpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"},
    AzureADB2C: {scope: "openid", endpoint: "https://tenant.b2clogin.com/tenant.onmicrosoft.com/userflow/oauth2/v2.0/authorize"},
    Slack: {scope: "users:read", endpoint: "https://slack.com/oauth/authorize"},
    Steam: {endpoint: "https://steamcommunity.com/openid/login"},
    Okta: {scope: "openid%20profile%20email", endpoint: "http://example.com"},
    Douyin: {scope: "user_info", endpoint: "https://open.douyin.com/platform/oauth/connect"},
    Kwai: {scope: "user_info", endpoint: "https://open.kuaishou.com/oauth2/connect"},
    Custom: {endpoint: "https://example.com/"},
    Bilibili: {endpoint: "https://passport.bilibili.com/register/pc_oauth2.html"},
    Line: {scope: "profile%20openid%20email", endpoint: "https://access.line.me/oauth2/v2.1/authorize"},
    Amazon: {scope: "profile", endpoint: "https://www.amazon.com/ap/oa"},
    Auth0: {scope: "openid%20profile%20email", endpoint: "http://auth0.com/authorize"},
    BattleNet: {scope: "openid", endpoint: "https://oauth.battlenet.com.cn/authorize"},
    Bitbucket: {scope: "account", endpoint: "https://bitbucket.org/site/oauth2/authorize"},
    Box: {scope: "root_readwrite", endpoint: "https://account.box.com/api/oauth2/authorize"},
    CloudFoundry: {scope: "cloud_controller.read", endpoint: "https://login.cloudfoundry.org/oauth/authorize"},
    Dailymotion: {scope: "userinfo", endpoint: "https://api.dailymotion.com/oauth/authorize"},
    Deezer: {scope: "basic_access", endpoint: "https://connect.deezer.com/oauth/auth.php"},
    DigitalOcean: {scope: "read", endpoint: "https://cloud.digitalocean.com/v1/oauth/authorize"},
    Discord: {scope: "identify%20email", endpoint: "https://discord.com/api/oauth2/authorize"},
    Dropbox: {scope: "account_info.read", endpoint: "https://www.dropbox.com/oauth2/authorize"},
    EveOnline: {scope: "publicData", endpoint: "https://login.eveonline.com/oauth/authorize"},
    Fitbit: {scope: "activity%20heartrate%20location%20nutrition%20profile%20settings%20sleep%20social%20weight", endpoint: "https://www.fitbit.com/oauth2/authorize"},
    Gitea: {scope: "user:email", endpoint: "https://gitea.com/login/oauth/authorize"},
    Heroku: {scope: "global", endpoint: "https://id.heroku.com/oauth/authorize"},
    InfluxCloud: {scope: "read:org", endpoint: "https://cloud2.influxdata.com/oauth/authorize"},
    Instagram: {scope: "user_profile", endpoint: "https://api.instagram.com/oauth/authorize"},
    Intercom: {scope: "user.read", endpoint: "https://app.intercom.com/oauth"},
    Kakao: {scope: "account_email", endpoint: "https://kauth.kakao.com/oauth/authorize"},
    Lastfm: {scope: "user_read", endpoint: "https://www.last.fm/api/auth"},
    Mailru: {scope: "userinfo", endpoint: "https://oauth.mail.ru/login"},
    MailRu: {scope: "userinfo", endpoint: "https://oauth.mail.ru/login"},
    Meetup: {scope: "basic", endpoint: "https://secure.meetup.com/oauth2/authorize"},
    MicrosoftOnline: {scope: "openid%20profile%20email", endpoint: "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"},
    Naver: {scope: "profile", endpoint: "https://nid.naver.com/oauth2.0/authorize"},
    Nextcloud: {scope: "openid%20profile%20email", endpoint: "https://cloud.example.org/apps/oauth2/authorize"},
    OneDrive: {scope: "offline_access%20onedrive.readonly", endpoint: "https://login.live.com/oauth20_authorize.srf"},
    Oura: {scope: "personal", endpoint: "https://cloud.ouraring.com/oauth/authorize"},
    Patreon: {scope: "identity", endpoint: "https://www.patreon.com/oauth2/authorize"},
    PayPal: {scope: "openid%20profile%20email", endpoint: "https://www.sandbox.paypal.com/connect"},
    SalesForce: {scope: "openid%20profile%20email", endpoint: "https://login.salesforce.com/services/oauth2/authorize"},
    Shopify: {scope: "read_products", endpoint: "https://myshopify.com/admin/oauth/authorize"},
    Soundcloud: {scope: "non-expiring", endpoint: "https://api.soundcloud.com/connect"},
    SoundCloud: {scope: "non-expiring", endpoint: "https://api.soundcloud.com/connect"},
    Spotify: {scope: "user-read-email", endpoint: "https://accounts.spotify.com/authorize"},
    Strava: {scope: "read", endpoint: "https://www.strava.com/oauth/authorize"},
    Stripe: {scope: "read_only", endpoint: "https://connect.stripe.com/oauth/authorize"},
    TikTok: {scope: "user.info.basic", endpoint: "https://www.tiktok.com/auth/authorize/"},
    Tumblr: {scope: "basic", endpoint: "https://www.tumblr.com/oauth2/authorize"},
    Twitch: {scope: "user_read", endpoint: "https://id.twitch.tv/oauth2/authorize"},
    Twitter: {scope: "users.read%20tweet.read", endpoint: "https://twitter.com/i/oauth2/authorize"},
    Telegram: {scope: "", endpoint: "https://core.telegram.org/widgets/login"},
    Typetalk: {scope: "my", endpoint: "https://typetalk.com/oauth2/authorize"},
    Uber: {scope: "profile", endpoint: "https://login.uber.com/oauth/v2/authorize"},
    VK: {scope: "email", endpoint: "https://oauth.vk.com/authorize"},
    Wepay: {scope: "manage_accounts%20view_user", endpoint: "https://www.wepay.com/v2/oauth2/authorize"},
    Xero: {scope: "openid%20profile%20email", endpoint: "https://login.xero.com/identity/connect/authorize"},
    Yahoo: {scope: "openid%20profile%20email", endpoint: "https://api.login.yahoo.com/oauth2/request_auth"},
    Yammer: {scope: "user", endpoint: "https://www.yammer.com/oauth2/authorize"},
    Yandex: {scope: "login:email", endpoint: "https://oauth.yandex.com/authorize"},
    Zoom: {scope: "user:read", endpoint: "https://zoom.us/oauth/authorize"},
    MetaMask: {scope: "", endpoint: ""},
    Web3Onboard: {scope: "", endpoint: ""}
  };

  function getStateFromQueryParams(applicationName, providerName, method, isShortState) {
    var query = window.location.search;
    query = query + "&application=" + encodeURIComponent(applicationName) + "&provider=" + encodeURIComponent(providerName) + "&method=" + method;
    if (method === "link") {
      query = query + "&from=" + window.location.pathname;
    }

    if (!isShortState) {
      return btoa(query);
    }

    var state = providerName;
    sessionStorage.setItem(state, query);
    return state;
  }

  async function getAuthUrl(application, provider, method, code) {
    if (!application || !provider) {
      return "";
    }

    var normalizedType = provider.type.indexOf("Custom") === 0 ? "Custom" : provider.type;
    var info = authInfo[normalizedType];
    if (!info) {
      return "";
    }

    var endpoint = info.endpoint;
    var redirectOrigin = application.forcedRedirectOrigin ? application.forcedRedirectOrigin : window.location.origin;
    var redirectUri = redirectOrigin + "/callback";
    var scope = info.scope;
    if (provider.scopes && provider.scopes.trim() !== "") {
      scope = provider.scopes;
    }

    var isShortState = (provider.type === "WeChat" && navigator.userAgent.indexOf("MicroMessenger") !== -1) || provider.type === "Twitter";
    var applicationName = application.name;
    if (application.isShared) {
      applicationName = application.name + "-org-" + application.organization;
    }

    var state = getStateFromQueryParams(applicationName, provider.name, method, isShortState);
    var codeVerifier = generateCodeVerifier();
    var codeChallenge = await generateCodeChallenge(codeVerifier);
    storeCodeVerifier(state, codeVerifier);

    if (provider.type === "AzureAD") {
      if (provider.domain !== "") {
        endpoint = endpoint.replace("common", provider.domain);
      }
    } else if (provider.type === "Apple") {
      redirectUri = redirectOrigin + "/api/callback";
    } else if (provider.type === "Google" && provider.disableSsl) {
      scope += "+https://www.googleapis.com/auth/user.phonenumbers.read";
    } else if (provider.type === "Nextcloud") {
      if (provider.domain) {
        endpoint = provider.domain + "/apps/oauth2/authorize";
      }
    } else if (provider.type === "Lark" && provider.disableSsl) {
      endpoint = authInfo[provider.type].endpoint2;
    }

    if (["Google", "GitHub", "Facebook", "Weibo", "Gitee", "LinkedIn", "GitLab", "AzureAD", "Slack", "Line", "Amazon", "Auth0", "BattleNet", "Bitbucket", "Box", "CloudFoundry", "Dailymotion", "DigitalOcean", "Discord", "Dropbox", "EveOnline", "Gitea", "Heroku", "InfluxCloud", "Instagram", "Intercom", "Kakao", "MailRu", "Mailru", "Meetup", "MicrosoftOnline", "Naver", "Nextcloud", "OneDrive", "Oura", "Patreon", "PayPal", "SalesForce", "SoundCloud", "Soundcloud", "Spotify", "Strava", "Stripe", "Tumblr", "Twitch", "Typetalk", "Uber", "VK", "Wepay", "Xero", "Yahoo", "Yammer", "Yandex", "Zoom"].indexOf(provider.type) !== -1) {
      return endpoint + "?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&scope=" + scope + "&response_type=code&state=" + state;
    } else if (provider.type === "QQ") {
      return endpoint + "?response_type=code&client_id=" + provider.clientId + "&redirect_uri=" + encodeURIComponent(redirectUri) + "&state=" + encodeURIComponent(state) + "&scope=" + encodeURIComponent(scope);
    } else if (provider.type === "AzureADB2C") {
      return "https://" + provider.domain + ".b2clogin.com/" + provider.domain + ".onmicrosoft.com/" + provider.appId + "/oauth2/v2.0/authorize?client_id=" + provider.clientId + "&nonce=defaultNonce&redirect_uri=" + encodeURIComponent(redirectUri) + "&scope=" + scope + "&response_type=code&state=" + state + "&prompt=login";
    } else if (provider.type === "DingTalk") {
      return endpoint + "?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&scope=" + scope + "&response_type=code&prompt=login%20consent&state=" + state;
    } else if (provider.type === "WeChat") {
      if (navigator.userAgent.indexOf("MicroMessenger") !== -1) {
        return authInfo[provider.type].mpEndpoint + "?appid=" + provider.clientId2 + "&redirect_uri=" + redirectUri + "&state=" + state + "&scope=" + authInfo[provider.type].mpScope + "&response_type=code#wechat_redirect";
      }

      if (provider.clientId2 && provider.disableSsl && provider.signName === "media") {
        return redirectOrigin + "/callback?state=" + state + "&code=wechat_oa:" + code;
      }

      return endpoint + "?appid=" + provider.clientId + "&redirect_uri=" + redirectUri + "&scope=" + scope + "&response_type=code&state=" + state + "#wechat_redirect";
    } else if (provider.type === "WeCom") {
      if (provider.subType === "Internal") {
        if (provider.method === "Silent") {
          endpoint = authInfo[provider.type].silentEndpoint;
          return endpoint + "?appid=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&scope=" + scope + "&response_type=code#wechat_redirect";
        }

        if (provider.method === "Normal") {
          endpoint = authInfo[provider.type].internalEndpoint;
          return endpoint + "?login_type=CorpApp&appid=" + provider.clientId + "&agentid=" + provider.appId + "&redirect_uri=" + redirectUri + "&state=" + state;
        }

        return "https://error:not-supported-provider-method:" + provider.method;
      }

      if (provider.subType === "Third-party") {
        if (provider.method === "Silent") {
          endpoint = authInfo[provider.type].silentEndpoint;
          return endpoint + "?appid=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&scope=" + scope + "&response_type=code#wechat_redirect";
        }

        if (provider.method === "Normal") {
          endpoint = authInfo[provider.type].endpoint;
          return endpoint + "?login_type=ServiceApp&appid=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state;
        }

        return "https://error:not-supported-provider-method:" + provider.method;
      }

      return "https://error:not-supported-provider-sub-type:" + provider.subType;
    } else if (provider.type === "Lark") {
      if (provider.disableSsl) {
        redirectUri = encodeURIComponent(redirectUri);
      }

      return endpoint + "?app_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state;
    } else if (provider.type === "ADFS") {
      return provider.domain + "/adfs/oauth2/authorize?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code&nonce=casdoor&scope=openid";
    } else if (provider.type === "Baidu") {
      return endpoint + "?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code&scope=" + scope + "&display=popup";
    } else if (provider.type === "Alipay") {
      return endpoint + "?app_id=" + provider.clientId + "&scope=auth_user&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code&scope=" + scope + "&display=popup";
    } else if (provider.type === "Casdoor") {
      return provider.domain + "/login/oauth/authorize?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code&scope=" + scope;
    } else if (provider.type === "Infoflow") {
      return endpoint + "?appid=" + provider.clientId + "&redirect_uri=" + redirectUri + "?state=" + state;
    } else if (provider.type === "Apple") {
      return endpoint + "?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code%20id_token&scope=" + scope + "&response_mode=form_post";
    } else if (provider.type === "Steam") {
      return endpoint + "?openid.claimed_id=http://specs.openid.net/auth/2.0/identifier_select&openid.identity=http://specs.openid.net/auth/2.0/identifier_select&openid.mode=checkid_setup&openid.ns=http://specs.openid.net/auth/2.0&openid.realm=" + redirectOrigin + "&openid.return_to=" + redirectUri + "?state=" + state;
    } else if (provider.type === "Okta") {
      return provider.domain + "/v1/authorize?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code&scope=" + scope;
    } else if (provider.type === "Douyin" || provider.type === "TikTok") {
      return endpoint + "?client_key=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code&scope=" + scope;
    } else if (provider.type === "Kwai") {
      return endpoint + "?app_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code&scope=" + scope;
    } else if (normalizedType === "Custom") {
      var authUrl = provider.customAuthUrl + "?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&scope=" + provider.scopes + "&response_type=code&state=" + state;
      if (provider.enablePkce) {
        authUrl += "&code_challenge=" + codeChallenge + "&code_challenge_method=S256";
      }
      return authUrl;
    } else if (provider.type === "Bilibili") {
      return endpoint + "#/?client_id=" + provider.clientId + "&return_url=" + redirectUri + "&state=" + state + "&response_type=code";
    } else if (provider.type === "Deezer") {
      return endpoint + "?app_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&perms=" + scope;
    } else if (provider.type === "Lastfm") {
      return endpoint + "?api_key=" + provider.clientId + "&cb=" + redirectUri;
    } else if (provider.type === "Shopify") {
      return endpoint + "?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&scope=" + scope + "&state=" + state + "&grant_options[]=per-user";
    } else if (provider.type === "Twitter" || provider.type === "Fitbit") {
      return endpoint + "?client_id=" + provider.clientId + "&redirect_uri=" + redirectUri + "&state=" + state + "&response_type=code&scope=" + scope + "&code_challenge=" + codeChallenge + "&code_challenge_method=S256";
    } else if (provider.type === "Telegram") {
      return redirectOrigin + "/telegram-login?state=" + state;
    } else if (provider.type === "MetaMask" || provider.type === "Web3Onboard") {
      return redirectUri + "?state=" + state;
    }

    return "";
  }

  async function fetchJson(url, options) {
    var response = await fetch(url, options);
    if (!response.ok) {
      throw new Error("Request failed with status " + response.status);
    }

    return response.json();
  }

  async function getApplication(applicationName) {
    return fetchJson("/api/get-application?id=" + encodeURIComponent(applicationName), {
      credentials: "include"
    });
  }

  async function getProviders(applicationName, language) {
    return fetchJson("/api/get-login-providers?application=" + encodeURIComponent(applicationName) + "&acceptLanguage=" + encodeURIComponent(language), {
      credentials: "include"
    });
  }

  async function run() {
    localStorage.setItem("signinUrl", window.location.pathname + window.location.search);

    var providerHint = new URLSearchParams(window.location.search).get("provider_hint");
    if (!providerHint) {
      redirectToFallback();
      return;
    }

    var searchParams = new URLSearchParams();
    var currentSearchParams = new URLSearchParams(window.location.search);
    searchParams.set("clientId", currentSearchParams.get("client_id") || "");
    searchParams.set("responseType", currentSearchParams.get("response_type") || "");
    searchParams.set("redirectUri", currentSearchParams.get("redirect_uri") || "");
    searchParams.set("type", "code");
    searchParams.set("scope", currentSearchParams.get("scope") || "");
    searchParams.set("state", currentSearchParams.get("state") || "");
    searchParams.set("nonce", currentSearchParams.get("nonce") || "");
    searchParams.set("code_challenge_method", currentSearchParams.get("code_challenge_method") || "");
    searchParams.set("code_challenge", currentSearchParams.get("code_challenge") || "");

    var response = await fetch("/api/get-app-login?" + searchParams.toString(), {
      method: "GET",
      credentials: "include",
      headers: {
        "Accept-Language": getAcceptLanguage()
      }
    });
    var payload = await response.json();
    if (payload.status !== "ok" || !payload.data) {
      redirectToFallback();
      return;
    }

    var application = payload.data;
    localStorage.setItem("applicationName", application.name || "");

    var providerItem = (application.providers || []).find(function(item) {
      return item && item.provider && item.provider.name === providerHint && isProviderVisibleForSignIn(item);
    });
    if (!providerItem) {
      redirectToFallback();
      return;
    }

    var authUrl = await getAuthUrl(application, providerItem.provider, "signup");
    if (!authUrl) {
      redirectToFallback();
      return;
    }

    window.location.replace(authUrl);
  }

  window.CasdoorProviderHintRedirect = {
    run: function() {
      return run().catch(function() {
        redirectToFallback();
      });
    }
  };
})();