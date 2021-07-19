// Copyright 2021 The casbin Authors. All Rights Reserved.
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

import * as Util from "./Util";
import {StaticBaseUrl} from "../Setting";

export const GoogleAuthScope  = "profile+email"
export const GoogleAuthUri = "https://accounts.google.com/signin/oauth";
export const GoogleAuthLogo = `${StaticBaseUrl}/img/social_google.png`;

export const GithubAuthScope  = "user:email+read:user"
export const GithubAuthUri = "https://github.com/login/oauth/authorize";
export const GithubAuthLogo = `${StaticBaseUrl}/img/social_github.png`;

export const QqAuthScope  = "get_user_info"
export const QqAuthUri = "https://graph.qq.com/oauth2.0/authorize";
export const QqAuthLogo = `${StaticBaseUrl}/img/social_qq.png`;

export const WeChatAuthScope = "snsapi_login"
export const WeChatAuthUri = "https://open.weixin.qq.com/connect/qrconnect";
export const WeChatAuthLogo = `${StaticBaseUrl}/img/social_wechat.png`;

export const FacebookAuthScope = "email,public_profile";
export const FacebookAuthUri = "https://www.facebook.com/dialog/oauth";
export const FacebookAuthLogo = `${StaticBaseUrl}/img/social_facebook.png`;

// export const DingTalkAuthScope = "email,public_profile";
export const DingTalkAuthUri = "https://oapi.dingtalk.com/connect/oauth2/sns_authorize";
export const DingTalkAuthLogo = `${StaticBaseUrl}/img/social_dingtalk.png`;

export const WeiboAuthScope = "email";
export const WeiboAuthUri = "https://api.weibo.com/oauth2/authorize";
export const WeiboAuthLogo = `${StaticBaseUrl}/img/social_weibo.png`;

export const GiteeAuthScope = "user_info,emails";
export const GiteeAuthUri = "https://gitee.com/oauth/authorize";
export const GiteeAuthLogo = `${StaticBaseUrl}/img/social_gitee.png`;

export const LinkedInAuthScope = "r_liteprofile%20r_emailaddress";
export const LinkedInAuthUri = "https://www.linkedin.com/oauth/v2/authorization";
export const LinkedInAuthLogo = `${StaticBaseUrl}/img/social_linkedin.png`;

export function getAuthLogo(provider) {
  if (provider.type === "Google") {
    return GoogleAuthLogo;
  } else if (provider.type === "GitHub") {
    return GithubAuthLogo;
  } else if (provider.type === "QQ") {
    return QqAuthLogo;
  } else if (provider.type === "WeChat") {
    return WeChatAuthLogo;
  } else if (provider.type === "Facebook") {
    return FacebookAuthLogo;
  } else if (provider.type === "DingTalk") {
    return DingTalkAuthLogo;
  } else if (provider.type === "Weibo") {
    return WeiboAuthLogo;
  } else if (provider.type === "Gitee") {
    return GiteeAuthLogo;
  } else if (provider.type === "LinkedIn") {
    return LinkedInAuthLogo;
  }
}

export function getAuthUrl(application, provider, method) {
  if (application === null || provider === null) {
    return "";
  }

  const redirectUri = `${window.location.origin}/callback`;
  const state = Util.getQueryParamsToState(application.name, provider.name, method);
  if (provider.type === "Google") {
    return `${GoogleAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${GoogleAuthScope}&response_type=code&state=${state}`;
  } else if (provider.type === "GitHub") {
    return `${GithubAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${GithubAuthScope}&response_type=code&state=${state}`;
  } else if (provider.type === "QQ") {
    return `${QqAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${QqAuthScope}&response_type=code&state=${state}`;
  } else if (provider.type === "WeChat") {
    return `${WeChatAuthUri}?appid=${provider.clientId}&redirect_uri=${redirectUri}&scope=${WeChatAuthScope}&response_type=code&state=${state}#wechat_redirect`;
  } else if (provider.type === "Facebook") {
    return `${FacebookAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${FacebookAuthScope}&response_type=code&state=${state}`;
  } else if (provider.type === "DingTalk") {
    return `${DingTalkAuthUri}?appid=${provider.clientId}&redirect_uri=${redirectUri}&scope=snsapi_login&response_type=code&state=${state}`;
  } else if (provider.type === "Weibo") {
    return `${WeiboAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${WeiboAuthScope}&response_type=code&state=${state}`;
  } else if (provider.type === "Gitee") {
    return `${GiteeAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${GiteeAuthScope}&response_type=code&state=${state}`;
  } else if (provider.type === "LinkedIn") {
    return `${LinkedInAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${LinkedInAuthScope}&response_type=code&state=${state}`
  }
}
