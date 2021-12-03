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

import React from "react";
import {Tooltip} from "antd";
import * as Util from "./Util";
import {StaticBaseUrl} from "../Setting";

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
  },
  Facebook: {
    scope: "email,public_profile",
    endpoint: "https://www.facebook.com/dialog/oauth",
  },
  DingTalk: {
    scope: "snsapi_login",
    endpoint: "https://oapi.dingtalk.com/connect/oauth2/sns_authorize",
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
  },
  Lark: {
    // scope: "email",
    endpoint: "https://open.feishu.cn/open-apis/authen/v1/index",
  },
  GitLab: {
    scope: "read_user+profile",
    endpoint: "https://gitlab.com/oauth/authorize",
  },
};

const otherProviderInfo = {
  SMS: {
    "Aliyun SMS": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://aliyun.com/product/sms",
    },
    "Tencent Cloud SMS": {
      logo: `${StaticBaseUrl}/img/social_tencent_cloud.jpg`,
      url: "https://cloud.tencent.com/product/sms",
    },
    "Volc Engine SMS": {
      logo: `${StaticBaseUrl}/img/social_volc_engine.jpg`,
      url: "https://www.volcengine.com/products/cloud-sms",
    },
  },
  Email: {
    "Default": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "",
    },
  },
  Storage: {
    "Local File System": {
      logo: `${StaticBaseUrl}/img/social_file.png`,
      url: "",
    },
    "AWS S3": {
      logo: `${StaticBaseUrl}/img/social_aws.png`,
      url: "https://aws.amazon.com/s3",
    },
    "Aliyun OSS": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://aliyun.com/product/oss",
    },
  },
  SAML: {
    "Aliyun IDaaS": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://aliyun.com/product/idaas"
    },
  },
};

export function getProviderLogo(provider) {
  if (provider.category === "OAuth") {
    return `${StaticBaseUrl}/img/social_${provider.type.toLowerCase()}.png`;
  } else {
    return otherProviderInfo[provider.category][provider.type].logo;
  }
}

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
    return otherProviderInfo[provider.category][provider.type].url;
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
          <img width={36} height={36} src={getProviderLogo(provider)} alt={provider.displayName} />
        </a>
      </Tooltip>
    )
  } else {
    return (
      <Tooltip title={provider.type}>
        <img width={36} height={36} src={getProviderLogo(provider)} alt={provider.displayName} />
      </Tooltip>
    )
  }
}

export function getAuthUrl(application, provider, method) {
  if (application === null || provider === null) {
    return "";
  }

  const endpoint = authInfo[provider.type].endpoint;
  const redirectUri = `${window.location.origin}/callback`;
  const scope = authInfo[provider.type].scope;
  const state = Util.getQueryParamsToState(application.name, provider.name, method);

  if (provider.type === "Google") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "GitHub") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "QQ") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "WeChat") {
    return `${endpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}#wechat_redirect`;
  } else if (provider.type === "Facebook") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "DingTalk") {
    return `${endpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "Weibo") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "Gitee") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "LinkedIn") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&scope=${scope}&response_type=code&state=${state}`;
  } else if (provider.type === "WeCom") {
    if (provider.method === "Silent") {
      return `${authInfo[provider.type].silentEndpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&scope=${scope}&response_type=code#wechat_redirect`;
    } else if (provider.method === "Normal") {
      return `${endpoint}?appid=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&usertype=member`;
    } else {
      return `https://error:not-supported-provider-method:${provider.method}`;
    }
  } else if (provider.type === "Lark") {
    return `${endpoint}?app_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}`;
  } else if (provider.type === "GitLab") {
    return `${endpoint}?client_id=${provider.clientId}&redirect_uri=${redirectUri}&state=${state}&response_type=code&scope=${scope}`;
  }
}
