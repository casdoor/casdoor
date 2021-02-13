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

export const GoogleClientId  = "";
export const GithubClientId  = "";
export const QqClientId  = "";
export const WechatClientId  = "";

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

export function getGoogleAuthCode(method) {
  window.location.href = `${GoogleAuthUri}?client_id=${GoogleClientId}&redirect_uri=${Setting.ClientUrl}/callback/google/${method}&scope=${GoogleAuthScope}&response_type=code&state=${AuthState}`;
}

export function getGithubAuthCode(method) {
  window.location.href = `${GithubAuthUri}?client_id=${GithubClientId}&redirect_uri=${Setting.ClientUrl}/callback/github/${method}&scope=${GithubAuthScope}&response_type=code&state=${AuthState}`;
}

export function getQqAuthCode(method) {
  window.location.href = `${QqAuthUri}?client_id=${QqClientId}&redirect_uri=${Setting.ClientUrl}/callback/qq/${method}&scope=${QqAuthScope}&response_type=code&state=${AuthState}`;
}

export function getWeChatAuthCode(method) {
  window.location.href = `${WeChatAuthUri}?appid=${WechatClientId}&redirect_uri=${Setting.ClientUrl}/callback/wechat/${method}&scope=${WeChatAuthScope}&response_type=code&state=${AuthState}#wechat_redirect`;
}
