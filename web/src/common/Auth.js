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

import * as Setting from "../Setting";

export const GoogleAuthScope  = "profile+email"
export const GoogleAuthUri = "https://accounts.google.com/signin/oauth";
export const GoogleAuthLogo = "https://cdn.jsdelivr.net/gh/casbin/static/img/social_google.png";
export const GithubAuthScope  = "user:email+read:user"
export const GithubAuthUri = "https://github.com/login/oauth/authorize";
export const GithubAuthLogo = "https://cdn.jsdelivr.net/gh/casbin/static/img/social_github.png";
export const QqAuthScope  = "get_user_info"
export const QqAuthUri = "https://graph.qq.com/oauth2.0/authorize";
export const QqAuthLogo = "https://cdn.jsdelivr.net/gh/casbin/static/img/social_qq.png";
export const WeChatAuthScope = "snsapi_login"
export const WeChatAuthUri = "https://open.weixin.qq.com/connect/qrconnect";
export const WeChatAuthLogo = "https://cdn.jsdelivr.net/gh/casbin/static/img/social_wechat.png";

export const AuthState = "casdoor";

export function getAuthLogo(provider) {
  if (provider.type === "google") {
    return GoogleAuthLogo;
  } else if (provider.type === "github") {
    return GithubAuthLogo;
  } else if (provider.type === "qq") {
    return QqAuthLogo;
  } else if (provider.type === "wechat") {
    return WeChatAuthLogo;
  }
}

export function getAuthUrl(provider, method) {
  const redirectUri = `${Setting.ClientUrl}/callback/${provider.type}/${provider.name}/${method}`;
  if (provider.type === "google") {
    return `${GoogleAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${GoogleAuthScope}&response_type=code&state=${AuthState}`;
  } else if (provider.type === "github") {
    return `${GithubAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${GithubAuthScope}&response_type=code&state=${AuthState}`;
  } else if (provider.type === "qq") {
    return `${QqAuthUri}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${QqAuthScope}&response_type=code&state=${AuthState}`;
  } else if (provider.type === "wechat") {
    return `${WeChatAuthUri}?appid=${provider.clientId}&redirect_uri=${redirectUri}&scope=${WeChatAuthScope}&response_type=code&state=${AuthState}#wechat_redirect`;
  }
}
