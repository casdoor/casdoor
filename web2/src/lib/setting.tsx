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

// Framework-agnostic helpers shared by the whole app. This is the shadcn port of
// the legacy web/src/Setting.js: the pure logic is kept identical so that the new
// UI talks to the Casdoor backend exactly like the antd frontend did, while every
// antd-specific renderer has been replaced by a shadcn/Tailwind equivalent.

import React from "react";
import i18next from "i18next";
import dayjs from "dayjs";
import copy from "copy-to-clipboard";
import {toast} from "sonner";
import * as phoneNumber from "libphonenumber-js";
import countriesLib from "i18n-iso-countries";
import enCountries from "i18n-iso-countries/langs/en.json";
import * as Conf from "@/Conf";
import {authConfig} from "@/auth/Auth";
import "@/i18n";

export const ServerUrl = "";

export const StaticBaseUrl = Conf.StaticBaseUrl;

export const MAX_PAGE_SIZE = 25;
export const SEARCH_DEBOUNCE_MS = 300;

export function showMessage(type: "success" | "error" | "info" | "warning", text: string) {
  if (type === "success") {
    toast.success(text);
  } else if (type === "error") {
    toast.error(text);
  } else if (type === "warning") {
    toast.warning(text);
  } else {
    toast.info(text);
  }
}

export function isLocalhost() {
  const hostname = window.location.hostname;
  return hostname === "localhost";
}

export function initServerUrl() {
  // The frontend is served by the backend in production, so relative URLs are enough.
  // In development the Vite dev-server proxies /api to the Go backend.
}

export function initWebConfig() {
  Conf.initConfigFromCookie();
}

export function getFullServerUrl() {
  let fullServerUrl = window.location.origin;
  if (fullServerUrl === "http://localhost:7002" || fullServerUrl === "http://localhost:7001") {
    fullServerUrl = "http://localhost:8000";
  }
  return fullServerUrl;
}

export function isMobile() {
  if (typeof window === "undefined") {
    return false;
  }
  return window.matchMedia("(max-width: 767px)").matches;
}

export function getFormattedDate(date) {
  if (!date) {
    return null;
  }
  return dayjs(date).format("YYYY-MM-DD HH:mm:ss");
}

export function getFormattedDateShort(date) {
  if (!date) {
    return "";
  }
  return date.slice(0, 10);
}

export function getShortName(s) {
  return s.split("/").slice(-1)[0];
}

export function getNameAtLeast(s) {
  s = getShortName(s);
  if (s.length >= 6) {
    return s;
  }
  return s.padEnd(6, " ");
}

let countriesInited = false;

export function initCountries() {
  if (!countriesInited) {
    countriesLib.registerLocale(enCountries as any);
    countriesInited = true;
  }
  return countriesLib;
}

export function getCountryImage(country) {
  return (
    <img
      src={`${StaticBaseUrl}/flag-icons/${country.code}.svg`}
      alt={country.name}
      height={20}
      className="mr-2 inline-block h-5"
    />
  );
}

export function getAvatarPlaceholder(name, size = 40) {
  return (
    <div
      className="flex shrink-0 items-center justify-center rounded-full font-bold text-white"
      style={{
        backgroundColor: getAvatarColor(name),
        fontSize: Math.round(size * 0.44),
        height: size,
        width: size,
      }}
    >
      {(name || "?").charAt(0).toUpperCase()}
    </div>
  );
}

export function getProviderLogo(provider) {
  const idp = provider.type.toLowerCase().trim().split(" ")[0];
  const url = getProviderLogoURL(provider);
  return <img width={30} height={30} src={url} alt={idp} className="h-[30px] w-[30px] object-contain" />;
}

export function copyToClipboard(text: string) {
  copy(text);
  showMessage("success", i18next.t("general:Copied to clipboard successfully"));
}

export function getClickable(text) {
  return (
    <button
      type="button"
      className="text-left underline-offset-4 hover:underline"
      onClick={() => copyToClipboard(text)}
    >
      {text}
    </button>
  );
}

// Sort order coming from the data table, translated to what the Casdoor API expects.
export function toApiSortOrder(order?: "asc" | "desc" | null) {
  if (order === "asc") {
    return "ascend";
  }
  if (order === "desc") {
    return "descend";
  }
  return "";
}

export function getThemeData(organization, application) {
  if (application?.themeData?.isEnabled) {
    return application.themeData;
  } else if (organization?.themeData?.isEnabled) {
    return organization.themeData;
  } else {
    return Conf.ThemeDefault;
  }
}


export function getAlgorithmNames(themeData) {
  const algorithms = [themeData?.themeType !== "dark" ? "default" : "dark"];
  if (themeData?.isCompact === true) {
    algorithms.push("compact");
  }

  return algorithms;
}

export function getLogo(themes) {
  if (themes.includes("dark")) {
    return `${StaticBaseUrl}/img/casdoor-logo_1185x256_dark.png`;
  } else {
    return `${StaticBaseUrl}/img/casdoor-logo_1185x256.png`;
  }
}


export const Countries = [
  {label: "English", key: "en", country: "US", alt: "English"},
  {label: "Español", key: "es", country: "ES", alt: "Español"},
  {label: "Français", key: "fr", country: "FR", alt: "Français"},
  {label: "Deutsch", key: "de", country: "DE", alt: "Deutsch"},
  {label: "日本語", key: "ja", country: "JP", alt: "日本語"},
  {label: "中文", key: "zh", country: "CN", alt: "中文"},
  {label: "TiếngViệt", key: "vi", country: "VN", alt: "TiếngViệt"},
  {label: "Português", key: "pt", country: "PT", alt: "Português"},
  {label: "Türkçe", key: "tr", country: "TR", alt: "Türkçe"},
  {label: "Polski", key: "pl", country: "PL", alt: "Polski"},
  {label: "Українська", key: "uk", country: "UA", alt: "Українська"},
];

export const OtherProviderInfo = {
  SMS: {
    "Aliyun SMS": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://aliyun.com/product/sms",
    },
    "Amazon SNS": {
      logo: `${StaticBaseUrl}/img/social_aws.png`,
      url: "https://aws.amazon.com/cn/sns/",
    },
    "Azure ACS": {
      logo: `${StaticBaseUrl}/img/social_azure.png`,
      url: "https://azure.microsoft.com/en-us/products/communication-services",
    },
    "Infobip SMS": {
      logo: `${StaticBaseUrl}/img/social_infobip.png`,
      url: "https://portal.infobip.com/homepage/",
    },
    "Tencent Cloud SMS": {
      logo: `${StaticBaseUrl}/img/social_tencent_cloud.jpg`,
      url: "https://cloud.tencent.com/product/sms",
    },
    "Baidu Cloud SMS": {
      logo: `${StaticBaseUrl}/img/social_baidu_cloud.png`,
      url: "https://cloud.baidu.com/product/sms.html",
    },
    "Volc Engine SMS": {
      logo: `${StaticBaseUrl}/img/social_volc_engine.jpg`,
      url: "https://www.volcengine.com/products/cloud-sms",
    },
    "Huawei Cloud SMS": {
      logo: `${StaticBaseUrl}/img/social_huawei.png`,
      url: "https://www.huaweicloud.com/product/msgsms.html",
    },
    "UCloud SMS": {
      logo: `${StaticBaseUrl}/img/social_ucloud.png`,
      url: "https://www.ucloud.cn/site/product/usms.html",
    },
    "Twilio SMS": {
      logo: `${StaticBaseUrl}/img/social_twilio.svg`,
      url: "https://www.twilio.com/messaging",
    },
    "SmsBao SMS": {
      logo: `${StaticBaseUrl}/img/social_smsbao.png`,
      url: "https://www.smsbao.com/",
    },
    "SUBMAIL SMS": {
      logo: `${StaticBaseUrl}/img/social_submail.svg`,
      url: "https://www.mysubmail.com",
    },
    "Msg91 SMS": {
      logo: `${StaticBaseUrl}/img/social_msg91.ico`,
      url: "https://control.msg91.com/app/",
    },
    "OSON SMS": {
      logo: `${StaticBaseUrl}/img/social_osonsms.svg`,
      url: "https://osonsms.com/",
    },
    "Custom HTTP SMS": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "https://casdoor.org/docs/provider/sms/overview",
    },
    "Mock SMS": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "",
    },
  },
  Email: {
    "Default": {
      logo: `${StaticBaseUrl}/img/email_default.png`,
      url: "",
    },
    "SUBMAIL": {
      logo: `${StaticBaseUrl}/img/social_submail.svg`,
      url: "https://www.mysubmail.com",
    },
    "Mailtrap": {
      logo: `${StaticBaseUrl}/img/email_mailtrap.png`,
      url: "https://mailtrap.io",
    },
    "Azure ACS": {
      logo: `${StaticBaseUrl}/img/social_azure.png`,
      url: "https://learn.microsoft.com/zh-cn/azure/communication-services",
    },
    "SendGrid": {
      logo: `${StaticBaseUrl}/img/email_sendgrid.png`,
      url: "https://sendgrid.com/",
    },
    "Custom HTTP Email": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "https://casdoor.org/docs/provider/email/overview",
    },
    "Resend": {
      logo: `${StaticBaseUrl}/img/email_resend.png`,
      url: "https://resend.com/",
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
    "MinIO": {
      logo: `${StaticBaseUrl}/img/social_minio.png`,
      url: "https://min.io/",
    },
    "Aliyun OSS": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://aliyun.com/product/oss",
    },
    "Tencent Cloud COS": {
      logo: `${StaticBaseUrl}/img/social_tencent_cloud.jpg`,
      url: "https://cloud.tencent.com/product/cos",
    },
    "Azure Blob": {
      logo: `${StaticBaseUrl}/img/social_azure.png`,
      url: "https://azure.microsoft.com/en-us/services/storage/blobs/",
    },
    "Qiniu Cloud Kodo": {
      logo: `${StaticBaseUrl}/img/social_qiniu_cloud.png`,
      url: "https://www.qiniu.com/solutions/storage",
    },
    "Google Cloud Storage": {
      logo: `${StaticBaseUrl}/img/social_google_cloud.png`,
      url: "https://cloud.google.com/storage",
    },
    "Synology": {
      logo: `${StaticBaseUrl}/img/social_synology.png`,
      url: "https://www.synology.com/en-global/dsm/feature/file_sharing",
    },
    "Casdoor": {
      logo: `${StaticBaseUrl}/img/casdoor.png`,
      url: "https://casdoor.org/docs/provider/storage/overview",
    },
    "CUCloud OSS": {
      logo: `${StaticBaseUrl}/img/social_cucloud.png`,
      url: "https://www.cucloud.cn/product/oss.html",
    },
  },
  SAML: {
    "Aliyun IDaaS": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://aliyun.com/product/idaas",
    },
    "Keycloak": {
      logo: `${StaticBaseUrl}/img/social_keycloak.png`,
      url: "https://www.keycloak.org/",
    },
    "Custom": {
      logo: `${StaticBaseUrl}/img/social_custom.png`,
      url: "https://door.casdoor.com/",
    },
    "Custom Flexible": {
      logo: `${StaticBaseUrl}/img/social_custom.png`,
      url: "https://door.casdoor.com/",
    },
  },
  Payment: {
    "Dummy": {
      logo: `${StaticBaseUrl}/img/payment_paypal.png`,
      url: "",
    },
    "Balance": {
      logo: `${StaticBaseUrl}/img/payment_balance.svg`,
      url: "",
    },
    "Alipay": {
      logo: `${StaticBaseUrl}/img/payment_alipay.png`,
      url: "https://www.alipay.com/",
    },
    "WeChat Pay": {
      logo: `${StaticBaseUrl}/img/payment_wechat_pay.png`,
      url: "https://pay.weixin.qq.com/",
    },
    "PayPal": {
      logo: `${StaticBaseUrl}/img/payment_paypal.png`,
      url: "https://www.paypal.com/",
    },
    "Stripe": {
      logo: `${StaticBaseUrl}/img/social_stripe.png`,
      url: "https://stripe.com/",
    },
    "AirWallex": {
      logo: `${StaticBaseUrl}/img/payment_airwallex.svg`,
      url: "https://airwallex.com/",
    },
    "GC": {
      logo: `${StaticBaseUrl}/img/payment_gc.png`,
      url: "https://gc.org",
    },
    "Polar": {
      logo: `${StaticBaseUrl}/img/payment_polar.png`,
      url: "https://polar.sh/",
    },
    "Paddle": {
      logo: `${StaticBaseUrl}/img/payment_paddle.png`,
      url: "https://www.paddle.com/",
    },
    "FastSpring": {
      logo: `${StaticBaseUrl}/img/payment_fastspring.png`,
      url: "https://fastspring.com/",
    },
    "Lemon Squeezy": {
      logo: `${StaticBaseUrl}/img/payment_lemonsqueezy.jpg`,
      url: "https://www.lemonsqueezy.com/",
    },
    "Adyen": {
      logo: `${StaticBaseUrl}/img/payment_adyen.svg`,
      url: "https://www.adyen.com/",
    },
  },
  Captcha: {
    "Default": {
      logo: `${StaticBaseUrl}/img/captcha_default.png`,
      url: "https://pkg.go.dev/github.com/dchest/captcha",
    },
    "reCAPTCHA": {
      logo: `${StaticBaseUrl}/img/social_recaptcha.png`,
      url: "https://www.google.com/recaptcha",
    },
    "reCAPTCHA v2": {
      logo: `${StaticBaseUrl}/img/social_recaptcha.png`,
      url: "https://www.google.com/recaptcha",
    },
    "reCAPTCHA v3": {
      logo: `${StaticBaseUrl}/img/social_recaptcha.png`,
      url: "https://www.google.com/recaptcha",
    },
    "hCaptcha": {
      logo: `${StaticBaseUrl}/img/social_hcaptcha.png`,
      url: "https://www.hcaptcha.com",
    },
    "Aliyun Captcha": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://help.aliyun.com/product/28308.html",
    },
    "GEETEST": {
      logo: `${StaticBaseUrl}/img/social_geetest.png`,
      url: "https://www.geetest.com",
    },
    "Cloudflare Turnstile": {
      logo: `${StaticBaseUrl}/img/social_cloudflare.png`,
      url: "https://www.cloudflare.com/products/turnstile/",
    },
  },
  AI: {
    "OpenAI API - GPT": {
      logo: `${StaticBaseUrl}/img/social_openai.svg`,
      url: "https://platform.openai.com",
    },
  },
  Web3: {
    "MetaMask": {
      logo: `${StaticBaseUrl}/img/social_metamask.svg`,
      url: "https://metamask.io/",
    },
    "Web3Onboard": {
      logo: `${StaticBaseUrl}/img/social_web3onboard.svg`,
      url: "https://onboard.blocknative.com/",
    },
  },
  Notification: {
    "Telegram": {
      logo: `${StaticBaseUrl}/img/social_telegram.png`,
      url: "https://telegram.org/",
    },
    "Custom HTTP": {
      logo: `${StaticBaseUrl}/img/email_default.png`,
      url: "https://casdoor.org/docs/provider/notification/overview",
    },
    "DingTalk": {
      logo: `${StaticBaseUrl}/img/social_dingtalk.png`,
      url: "https://www.dingtalk.com/",
    },
    "Lark": {
      logo: `${StaticBaseUrl}/img/social_lark.png`,
      url: "https://www.larksuite.com/",
    },
    "Microsoft Teams": {
      logo: `${StaticBaseUrl}/img/social_teams.png`,
      url: "https://www.microsoft.com/microsoft-teams",
    },
    "Bark": {
      logo: `${StaticBaseUrl}/img/social_bark.png`,
      url: "https://apps.apple.com/us/app/bark-customed-notifications/id1403753865",
    },
    "Pushover": {
      logo: `${StaticBaseUrl}/img/social_pushover.png`,
      url: "https://pushover.net/",
    },
    "Pushbullet": {
      logo: `${StaticBaseUrl}/img/social_pushbullet.png`,
      url: "https://www.pushbullet.com/",
    },
    "Slack": {
      logo: `${StaticBaseUrl}/img/social_slack.png`,
      url: "https://slack.com/",
    },
    "Webpush": {
      logo: `${StaticBaseUrl}/img/email_default.png`,
      url: "https://developer.mozilla.org/en-US/docs/Web/API/Push_API",
    },
    "Discord": {
      logo: `${StaticBaseUrl}/img/social_discord.png`,
      url: "https://discord.com/",
    },
    "Google Chat": {
      logo: `${StaticBaseUrl}/img/social_google_chat.png`,
      url: "https://workspace.google.com/intl/en/products/chat/",
    },
    "Line": {
      logo: `${StaticBaseUrl}/img/social_line.png`,
      url: "https://line.me/",
    },
    "Matrix": {
      logo: `${StaticBaseUrl}/img/social_matrix.png`,
      url: "https://www.matrix.org/",
    },
    "Twitter": {
      logo: `${StaticBaseUrl}/img/social_twitter.png`,
      url: "https://twitter.com/",
    },
    "Reddit": {
      logo: `${StaticBaseUrl}/img/social_reddit.png`,
      url: "https://www.reddit.com/",
    },
    "Rocket Chat": {
      logo: `${StaticBaseUrl}/img/social_rocket_chat.png`,
      url: "https://rocket.chat/",
    },
    "Viber": {
      logo: `${StaticBaseUrl}/img/social_viber.png`,
      url: "https://www.viber.com/",
    },
    "CUCloud": {
      logo: `${StaticBaseUrl}/img/cucloud.png`,
      url: "https://www.cucloud.cn/",
    },
    "WeCom": {
      logo: `${StaticBaseUrl}/img/social_wecom.png`,
      url: "https://work.weixin.qq.com/",
    },
  },
  "Face ID": {
    "Alibaba Cloud Facebody": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://vision.aliyun.com/facebody",
    },
    "Local UniFace": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "https://github.com/yakhyo/uniface",
    },
  },
  "MFA": {
    "RADIUS": {
      logo: `${StaticBaseUrl}/img/mfa_radius.png`,
      url: "",
    },
  },
  "ID Verification": {
    "Jumio": {
      logo: `${StaticBaseUrl}/img/social_jumio.png`,
      url: "https://www.jumio.com/",
    },
    "Alibaba Cloud": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://www.aliyun.com/product/idverification",
    },
  },
  Log: {
    "Casdoor Permission Log": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "https://casdoor.org",
    },
    "System Log": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "https://en.wikipedia.org/wiki/Syslog",
    },
    "Agent": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "",
    },
    "SELinux Log": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "https://github.com/SELinuxProject/selinux",
    },
  },
  Scan: {
    "Security Scan": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "",
    },
    "MCP Scan": {
      logo: `${StaticBaseUrl}/img/social_default.png`,
      url: "",
    },
  },
};

export const UserFields = ["owner", "name", "password", "display_name", "id", "type", "email", "phone", "country_code",
  "is_admin", "homepage", "birthday", "gender", "password_type", "password_salt", "external_id", "avatar", "first_name", "last_name",
  "avatar_type", "permanent_avatar", "email_verified", "region", "location", "address",
  "affiliation", "title", "id_card_type", "id_card", "real_name", "is_verified", "bio", "tag", "language",
  "education", "score", "karma", "ranking", "balance", "balance_credit", "balance_currency", "currency", "is_default_avatar", "is_online",
  "is_forbidden", "is_deleted", "signup_application", "register_type", "register_source", "hash", "pre_hash", "access_token",
  "created_ip", "last_signin_time", "last_signin_ip", "github", "google", "qq", "wechat", "facebook", "dingtalk",
  "weibo", "gitee", "linkedin", "wecom", "lark", "gitlab", "adfs", "baidu", "alipay", "casdoor", "infoflow", "apple",
  "azuread", "azureadb2c", "slack", "steam", "bilibili", "okta", "douyin", "kwai", "line", "amazon", "auth0",
  "battlenet", "bitbucket", "box", "cloudfoundry", "dailymotion", "deezer", "digitalocean", "discord", "dropbox",
  "eveonline", "fitbit", "gitea", "heroku", "influxcloud", "instagram", "intercom", "kakao", "lastfm", "mailru",
  "meetup", "microsoftonline", "naver", "nextcloud", "onedrive", "oura", "patreon", "paypal", "salesforce", "shopify",
  "soundcloud", "spotify", "strava", "stripe", "tiktok", "tumblr", "twitch", "twitter", "typetalk", "uber", "vk",
  "wepay", "xero", "yahoo", "yammer", "yandex", "zoom", "metamask", "web3onboard", "custom", "webauthnCredentials",
  "preferred_mfa_type", "recovery_codes", "totp_secret", "mfa_phone_enabled", "mfa_email_enabled", "invitation",
  "invitation_code", "face_ids", "ldap", "properties", "roles", "permissions", "groups", "last_change_password_time",
  "last_signin_wrong_time", "signin_wrong_times", "managedAccounts", "mfaAccounts", "mfaItems", "need_update_password",
  "created_time", "updated_time", "deleted_time",
  "ip_whitelist"];

export const GroupFields = ["owner", "name", "created_time", "updated_time", "display_name", "manager",
  "contact_email", "type", "parent_id", "is_top_group", "is_enabled"];

export const RoleFields = ["owner", "name", "created_time", "display_name", "description",
  "users", "groups", "roles", "domains", "is_enabled"];

export const PermissionFields = ["owner", "name", "created_time", "display_name", "description",
  "users", "groups", "roles", "domains", "model", "adapter", "resource_type",
  "resources", "actions", "effect", "is_enabled", "submitter", "approver", "approve_time", "state"];


// The "<translated label>#<field>" column list behind the .xlsx import templates,
// ported verbatim from web/src/Setting.js.
export const GetTranslatedUserItems = () => {
  return [
    {name: "Organization", label: i18next.t("general:Organization")},
    {name: "ID", label: i18next.t("general:ID")},
    {name: "Name", label: i18next.t("general:Name")},
    {name: "Display name", label: i18next.t("general:Display name")},
    {name: "First name", label: i18next.t("general:First name")},
    {name: "Last name", label: i18next.t("general:Last name")},
    {name: "Avatar", label: i18next.t("general:Avatar")},
    {name: "User type", label: i18next.t("general:User type")},
    {name: "Password", label: i18next.t("general:Password")},
    {name: "Email", label: i18next.t("general:Email")},
    {name: "Phone", label: i18next.t("general:Phone")},
    {name: "Country code", label: i18next.t("user:Country code")},
    {name: "Country/Region", label: i18next.t("user:Country/Region")},
    {name: "Location", label: i18next.t("user:Location")},
    {name: "Address", label: i18next.t("user:Address")},
    {name: "Addresses", label: i18next.t("user:Addresses")},
    {name: "Affiliation", label: i18next.t("user:Affiliation")},
    {name: "Title", label: i18next.t("general:Title")},
    {name: "ID card type", label: i18next.t("user:ID card type")},
    {name: "ID card", label: i18next.t("user:ID card")},
    {name: "ID card info", label: i18next.t("user:ID card info")},
    {name: "Real name", label: i18next.t("application:Real name")},
    {name: "ID verification", label: i18next.t("user:ID verification")},
    {name: "Homepage", label: i18next.t("user:Homepage")},
    {name: "Bio", label: i18next.t("user:Bio")},
    {name: "Tag", label: i18next.t("general:Tag")},
    {name: "Language", label: i18next.t("user:Language")},
    {name: "Gender", label: i18next.t("user:Gender")},
    {name: "Birthday", label: i18next.t("user:Birthday")},
    {name: "Education", label: i18next.t("user:Education")},
    {name: "Balance", label: i18next.t("user:Balance")},
    {name: "Balance currency", label: i18next.t("organization:Balance currency")},
    {name: "Balance credit", label: i18next.t("organization:Balance credit")},
    {name: "Cart", label: i18next.t("general:Cart")},
    {name: "Transactions", label: i18next.t("general:Transactions")},
    {name: "UID number", label: i18next.t("general:UID number")},
    {name: "Score", label: i18next.t("user:Score")},
    {name: "Karma", label: i18next.t("user:Karma")},
    {name: "Ranking", label: i18next.t("user:Ranking")},
    {name: "Signup application", label: i18next.t("general:Signup application")},
    {name: "Register type", label: i18next.t("user:Register type")},
    {name: "Register source", label: i18next.t("user:Register source")},
    {name: "API key", label: i18next.t("general:API key")},
    {name: "Groups", label: i18next.t("general:Groups")},
    {name: "Roles", label: i18next.t("general:Roles")},
    {name: "Permissions", label: i18next.t("general:Permissions")},
    {name: "3rd-party logins", label: i18next.t("user:3rd-party logins")},
    {name: "Properties", label: i18next.t("user:Properties")},
    {name: "Is online", label: i18next.t("user:Is online")},
    {name: "Is admin", label: i18next.t("user:Is admin")},
    {name: "Is forbidden", label: i18next.t("user:Is forbidden")},
    {name: "Is deleted", label: i18next.t("user:Is deleted")},
    {name: "Need update password", label: i18next.t("user:Need update password")},
    {name: "IP whitelist", label: i18next.t("general:IP whitelist")},
    {name: "Multi-factor authentication", label: i18next.t("mfa:Multi-factor authentication")},
    {name: "WebAuthn credentials", label: i18next.t("user:WebAuthn credentials")},
    {name: "Last change password time", label: i18next.t("user:Last change password time")},
    {name: "Managed accounts", label: i18next.t("user:Managed accounts")},
    {name: "Face ID", label: i18next.t("login:Face ID")},
    {name: "MFA accounts", label: i18next.t("user:MFA accounts")},
    {name: "MFA items", label: i18next.t("general:MFA items")},
  ];
};

export function getUserColumns() {
  const items = GetTranslatedUserItems();
  return UserFields.map((field: string) => {
    let transField = "";
    if (field === "webauthnCredentials") {
      transField = "WebAuthn credentials";
    } else if (field === "region") {
      transField = "Country/Region";
    } else if (field === "mfaAccounts") {
      transField = "MFA accounts";
    } else if (field === "mfaItems") {
      transField = "MFA items";
    } else if (field === "face_ids") {
      transField = "Face ID";
    } else if (field === "managedAccounts") {
      transField = "Managed accounts";
    } else {
      transField = field.toLowerCase().split("_").join(" ");
      transField = transField.charAt(0).toUpperCase() + transField.slice(1);
      transField = transField.replace("ip", "IP")
        .replace("Ip", "IP")
        .replace("Id", "ID")
        .replace("id", "ID");
    }
    if (transField === "Owner") {
      transField = "Organization";
    }
    const transFieldItem = items.find((item: any) => item.name === transField);
    if (transFieldItem === undefined) {
      const toTranslateList = ["general", "user", "organization"].map((ns: string) => `${ns}:${transField}`);
      const transResult = toTranslateList.map((item: string) => i18next.t(item) === transField ? null : i18next.t(item))
        .find((item: any) => item !== null);
      transField = transResult ? transResult : transField;
    }
    return `${transFieldItem ? transFieldItem.label : transField}#${field}`;
  });
}

export function getGroupColumns() {
  return GroupFields.map((field: string) => {
    let transField = field.toLowerCase().split("_").join(" ");
    transField = transField.charAt(0).toUpperCase() + transField.slice(1);
    transField = transField.replace("Id", "ID");
    if (transField === "Owner") {
      transField = "Organization";
    }
    const toTranslateList = ["general", "group"].map((ns: string) => `${ns}:${transField}`);
    const transResult = toTranslateList.map((item: string) => i18next.t(item) === transField ? null : i18next.t(item))
      .find((item: any) => item !== null);
    transField = transResult ? transResult : transField;
    return `${transField}#${field}`;
  });
}

export function getRoleColumns() {
  return RoleFields.map((field: string) => {
    let transField = field.toLowerCase().split("_").join(" ");
    transField = transField.charAt(0).toUpperCase() + transField.slice(1);
    transField = transField.replace("Id", "ID");
    if (transField === "Owner") {
      transField = "Organization";
    }
    const toTranslateList = ["general", "role"].map((ns: string) => `${ns}:${transField}`);
    const transResult = toTranslateList.map((item: string) => i18next.t(item) === transField ? null : i18next.t(item))
      .find((item: any) => item !== null);
    transField = transResult ? transResult : transField;
    return `${transField}#${field}`;
  });
}

export function getPermissionColumns() {
  return PermissionFields.map((field: string) => {
    let transField = field.toLowerCase().split("_").join(" ");
    transField = transField.charAt(0).toUpperCase() + transField.slice(1);
    transField = transField.replace("Id", "ID");
    if (transField === "Owner") {
      transField = "Organization";
    }
    const toTranslateList = ["general", "permission"].map((ns: string) => `${ns}:${transField}`);
    const transResult = toTranslateList.map((item: string) => i18next.t(item) === transField ? null : i18next.t(item))
      .find((item: any) => item !== null);
    transField = transResult ? transResult : transField;
    return `${transField}#${field}`;
  });
}

export function getCountryCode(country) {
  if (phoneNumber.isSupportedCountry(country)) {
    return phoneNumber.getCountryCallingCode(country as any);
  }
  return "";
}

export function getCountryCodeData(countryCodes: any = phoneNumber.getCountries()): any[] {
  if (countryCodes?.includes("All")) {
    countryCodes = phoneNumber.getCountries();
  }
  return countryCodes?.map((countryCode) => {
    if (phoneNumber.isSupportedCountry(countryCode)) {
      const name = (initCountries() as any).getName(countryCode, getLanguage());
      return {
        code: countryCode,
        name: name || "",
        phone: phoneNumber.getCountryCallingCode(countryCode),
      };
    }
  }).filter(item => item && item.name !== "")
    .sort((a, b) => Number(a.phone) - Number(b.phone));
}

export function isProviderVisible(providerItem) {
  if (providerItem.provider === undefined || providerItem.provider === null) {
    return false;
  }

  if (!["OAuth", "SAML", "Web3"].includes(providerItem.provider.category)) {
    return false;
  }

  if (providerItem.provider.type === "WeChatMiniProgram") {
    return false;
  }

  return true;
}

export function isResponseDenied(data) {
  if (data.msg === "Unauthorized operation" || data.msg === "未授权的操作") {
    return true;
  }
  return false;
}

export function isProviderVisibleForSignUp(providerItem) {
  if (providerItem.canSignUp === false) {
    return false;
  }

  return isProviderVisible(providerItem);
}

export function isProviderVisibleForSignIn(providerItem) {
  if (providerItem.canSignIn === false) {
    return false;
  }

  return isProviderVisible(providerItem);
}

export function isProviderPrompted(providerItem) {
  return isProviderVisible(providerItem) && providerItem.prompted;
}

export function isSignupItemPrompted(signupItem) {
  return signupItem.visible && signupItem.prompted;
}

export function getAllPromptedProviderItems(application) {
  return application.providers?.filter(providerItem => isProviderPrompted(providerItem));
}

export function getAllPromptedSignupItems(application) {
  return application.signupItems?.filter(signupItem => isSignupItemPrompted(signupItem));
}

export function getSignupItem(application, itemName) {
  const signupItems = application.signupItems?.filter(signupItem => signupItem.name === itemName);
  if (signupItems?.length > 0) {
    return signupItems[0];
  }
  return null;
}

export function isValidPersonName(personName) {
  return personName !== "";

  // // https://blog.css8.cn/post/14210975.html
  // const personNameRegex = /^[\u4e00-\u9fa5]{2,6}$/;
  // return personNameRegex.test(personName);
}

export function isValidIdCard(idCard) {
  return idCard !== "";

  // const idCardRegex = /^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9X]$/;
  // return idCardRegex.test(idCard);
}

export function isValidEmail(email) {
  // https://github.com/yiminghe/async-validator/blob/057b0b047f88fac65457bae691d6cb7c6fe48ce1/src/rule/type.ts#L9
  const emailRegex = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
  return emailRegex.test(email);
}

export function isValidPhone(phone, countryCode = "") {
  if (countryCode !== "" && countryCode !== "CN") {
    return phoneNumber.isValidPhoneNumber(phone, countryCode as any);
  }

  // https://learnku.com/articles/31543, `^s*$` filter empty email individually.
  const phoneCnRegex = /^1(3\d|4[5-9]|5[0-35-9]|6[2567]|7[0-8]|8\d|9[0-35-9])\d{8}$/;
  const phoneRegex = /[0-9]{4,15}$/;

  return countryCode === "CN" ? phoneCnRegex.test(phone) : phoneRegex.test(phone);
}

export function isValidInvoiceTitle(invoiceTitle) {
  return invoiceTitle !== "";

  // if (invoiceTitle === "") {
  //   return false;
  // }
  //
  // // https://blog.css8.cn/post/14210975.html
  // const invoiceTitleRegex = /^[()（）\u4e00-\u9fa5]{0,50}$/;
  // return invoiceTitleRegex.test(invoiceTitle);
}

export function isValidTaxId(taxId) {
  return taxId !== "";

  // // https://www.codetd.com/article/8592083
  // const regArr = [/^[\da-z]{10,15}$/i, /^\d{6}[\da-z]{10,12}$/i, /^[a-z]\d{6}[\da-z]{9,11}$/i, /^[a-z]{2}\d{6}[\da-z]{8,10}$/i, /^\d{14}[\dx][\da-z]{4,5}$/i, /^\d{17}[\dx][\da-z]{1,2}$/i, /^[a-z]\d{14}[\dx][\da-z]{3,4}$/i, /^[a-z]\d{17}[\dx][\da-z]{0,1}$/i, /^[\d]{6}[\da-z]{13,14}$/i];
  // for (let i = 0; i < regArr.length; i++) {
  //   if (regArr[i].test(taxId)) {
  //     return true;
  //   }
  // }
  // return false;
}

export function isAffiliationPrompted(application) {
  const signupItem = getSignupItem(application, "Affiliation");
  if (signupItem === null) {
    return false;
  }

  return signupItem.prompted;
}

export function hasPromptPage(application) {
  const providerItems = getAllPromptedProviderItems(application);
  if (providerItems?.length > 0) {
    return true;
  }

  const signupItems = getAllPromptedSignupItems(application);
  if (signupItems?.filter(item => item.name === "Country/Region").length > 0) {
    return true;
  }

  return isAffiliationPrompted(application);
}

function isAffiliationAnswered(user, application) {
  if (!isAffiliationPrompted(application)) {
    return true;
  }

  if (user === null) {
    return false;
  }
  return user.affiliation !== "";
}

function isProviderItemAnswered(user, application, providerItem) {
  if (user === null) {
    return false;
  }

  const provider = providerItem.provider;
  const linkedValue = user[provider.type.toLowerCase()];
  return linkedValue !== undefined && linkedValue !== "";
}

function isSignupItemAnswered(user, signupItem) {
  if (user === null) {
    return false;
  }

  if (signupItem.name !== "Country/Region") {
    return true;
  }

  const value = user["region"];
  return value !== undefined && value !== "";
}

export function isPromptAnswered(user, application) {
  if (!isAffiliationAnswered(user, application)) {
    return false;
  }

  const providerItems = getAllPromptedProviderItems(application);
  for (let i = 0; i < providerItems.length; i++) {
    if (!isProviderItemAnswered(user, application, providerItems[i])) {
      return false;
    }
  }

  const signupItems = getAllPromptedSignupItems(application);
  for (let i = 0; i < signupItems.length; i++) {
    if (!isSignupItemAnswered(user, signupItems[i])) {
      return false;
    }
  }
  return true;
}

export const MfaRuleRequired = "Required";
export const MfaRulePrompted = "Prompted";
export const MfaRuleOptional = "Optional";

export const RequiredUpdatePassword = "RequiredUpdatePassword";

export function goToUpdatePassword() {
  // remember where the login was started from, to go back after the password is updated
  const signinUrl = localStorage.getItem("signinUrl");
  if (signinUrl) {
    sessionStorage.setItem("signinUrl", signinUrl);
  }
  goToLink("/account");
}

export function isRequiredEnableMfa(user, organization) {
  if (!user || !organization || (!organization.mfaItems && !user.mfaItems)) {
    return false;
  }
  return getMfaItemsByRules(user, organization, [MfaRuleRequired]).length > 0;
}

export function getMfaItemsByRules(user, organization, mfaRules = []) {
  if (!user || !organization || (!organization.mfaItems && !user.mfaItems)) {
    return [];
  }

  let mfaItems = organization.mfaItems;
  if (user.mfaItems && user.mfaItems.length !== 0) {
    mfaItems = user.mfaItems;
  }

  if (mfaItems === null) {
    return [];
  }

  return mfaItems.filter((mfaItem) => mfaRules.includes(mfaItem.rule))
    .filter((mfaItem) => user.multiFactorAuths.some((mfa) => mfa.mfaType === mfaItem.name && !mfa.enabled));
}

export function parseObject(s) {
  try {
    return eval("(" + s + ")");
  } catch (e) {
    return null;
  }
}

export function parseJson(s) {
  if (s === "") {
    return null;
  } else {
    return JSON.parse(s);
  }
}

export function myParseInt(i) {
  const res = parseInt(i);
  return isNaN(res) ? 0 : res;
}

export function openLink(link) {
  // this.props.history.push(link);
  const w = window.open("about:blank");
  w.location.href = link;
}

export function openLinkSafe(link) {
  // Javascript window.open issue in safari
  // https://stackoverflow.com/questions/45569893/javascript-window-open-issue-in-safari
  const a = document.createElement("a");
  a.href = link;
  a.setAttribute("target", "_blank");
  a.click();
}

export function goToLink(link) {
  window.location.href = link;
}


export function isAdminUser(account) {
  if (account === undefined || account === null) {
    return false;
  }
  return account.owner === "built-in";
}

export function isLocalAdminUser(account) {
  if (account === undefined || account === null) {
    return false;
  }
  return account.isAdmin === true || isAdminUser(account);
}

export function deepCopy(obj) {
  return Object.assign({}, obj);
}

// In "add" mode the object has already been created on the server before the edit page is
// shown, so the "Cancel" button deletes it again. The deletion must use the owner and name
// that the object was created with, otherwise renaming the object in the form would make
// "Cancel" delete another, already existing object instead.
export function getDeleteObj(obj, owner, name) {
  const res = deepCopy(obj);
  res.owner = owner;
  res.name = name;
  return res;
}

export function addRow(array, row, position = "end") {
  return position === "end" ? [...array, row] : [row, ...array];
}

export function deleteRow(array, i) {
  // return array = array.slice(0, i).concat(array.slice(i + 1));
  return [...array.slice(0, i), ...array.slice(i + 1)];
}

export function swapRow(array, i, j) {
  return [...array.slice(0, i), array[j], ...array.slice(i + 1, j), array[i], ...array.slice(j + 1)];
}

export function trim(str, ch) {
  if (str === undefined) {
    return undefined;
  }

  let start = 0;
  let end = str.length;

  while (start < end && str[start] === ch) {++start;}

  while (end > start && str[end - 1] === ch) {--end;}

  return (start > 0 || end < str.length) ? str.substring(start, end) : str;
}

export function getStyleInnerCss(css) {
  if (!css) {
    return css;
  }
  const match = css.match(/^\s*<style[^>]*>([\s\S]*?)<\/style>\s*$/i);
  return match ? match[1] : css;
}

export function getShortText(s, maxLength = 35) {
  if (s.length > maxLength) {
    return `${s.slice(0, maxLength)}...`;
  } else {
    return s;
  }
}

export function getFriendlyUserName(account) {
  if (account.firstName !== "" && account.lastName !== "") {
    return `${account.firstName}, ${account.lastName}`;
  } else if (account.displayName !== "") {
    return account.displayName;
  } else if (account.name !== "") {
    return account.name;
  } else {
    return account.id;
  }
}

export function getFriendlyFileSize(size) {
  if (size < 1024) {
    return size + " B";
  }

  const i = Math.floor(Math.log(size) / Math.log(1024));
  let num: any = (size / Math.pow(1024, i));
  const round = Math.round(num);
  num = round < 10 ? num.toFixed(2) : round < 100 ? num.toFixed(1) : round;
  return `${num} ${"KMGTPEZY"[i - 1]}B`;
}

function getHashInt(s) {
  let hash = 0;
  if (s.length !== 0) {
    for (let i = 0; i < s.length; i++) {
      const char = s.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
  }

  if (hash < 0) {
    hash = -hash;
  }
  return hash;
}

export function getAvatarColor(s) {
  const colorList = ["#f56a00", "#7265e6", "#ffbf00", "#00a2ae"];
  const hash = getHashInt(s);
  return colorList[hash % 4];
}

export function getEffectiveAvatarUrl(user) {
  return user.avatar || user.permanentAvatar || "";
}


export function getLanguageText(text) {
  if (!text) {
    return "";
  }

  if (!text.includes("|")) {
    return text;
  }

  let res;
  const tokens = text.split("|");
  if (getLanguage() !== "zh") {
    res = trim(tokens[0], "");
  } else {
    res = trim(tokens[1], "");
  }
  return res;
}

export function getLanguage() {
  return (i18next.language !== undefined && i18next.language !== null && i18next.language !== "" && i18next.language !== "null") ? i18next.language : Conf.DefaultLanguage;
}

export function setLanguage(language) {
  localStorage.setItem("language", language);
  i18next.changeLanguage(language);
}

// The language chosen on the signin/signup page, remembered so it survives the
// redirect through a provider and can be saved to the newly created user.
export function setSigninLanguage(language) {
  sessionStorage.setItem("signinLanguage", language);
}

export function getSigninLanguage() {
  return sessionStorage.getItem("signinLanguage") ?? "";
}

export function getAcceptLanguage() {
  if (i18next.language === null || i18next.language === "") {
    return "en;q=0.9,en;q=0.8";
  }
  return i18next.language + ";q=0.9,en;q=0.8";
}


export function getProviderLogoURL(provider) {
  if (provider.type.startsWith("Custom") && provider.customLogo) {
    return provider.customLogo;
  }
  if (provider.category === "OAuth") {
    const type = provider.type.startsWith("Custom") ? "Custom" : provider.type;
    return `${StaticBaseUrl}/img/social_${type.toLowerCase()}.png`;
  } else {
    const info = OtherProviderInfo[provider.category][provider.type];
    // avoid crash when provider is not found
    if (info) {
      return info.logo;
    }
    return "";
  }
}

export function getProviderTypeOptions(category) {
  if (category === "OAuth") {
    return (
      [
        {id: "Google", name: "Google"},
        {id: "GitHub", name: "GitHub"},
        {id: "QQ", name: "QQ"},
        {id: "WeChat", name: "WeChat"},
        {id: "WeChatMiniProgram", name: "WeChat Mini Program"},
        {id: "Facebook", name: "Facebook"},
        {id: "DingTalk", name: "DingTalk"},
        {id: "Weibo", name: "Weibo"},
        {id: "Gitee", name: "Gitee"},
        {id: "LinkedIn", name: "LinkedIn"},
        {id: "WeCom", name: "WeCom"},
        {id: "Lark", name: "Lark"},
        {id: "GitLab", name: "GitLab"},
        {id: "ADFS", name: "ADFS"},
        {id: "Baidu", name: "Baidu"},
        {id: "Alipay", name: "Alipay"},
        {id: "Casdoor", name: "Casdoor"},
        {id: "Infoflow", name: "Infoflow"},
        {id: "Apple", name: "Apple"},
        {id: "AzureAD", name: "Azure AD"},
        {id: "AzureADB2C", name: "Azure AD B2C"},
        {id: "Slack", name: "Slack"},
        {id: "Steam", name: "Steam"},
        {id: "Bilibili", name: "Bilibili"},
        {id: "Okta", name: "Okta"},
        {id: "Douyin", name: "Douyin"},
        {id: "Kwai", name: "Kwai"},
        {id: "Line", name: "Line"},
        {id: "Amazon", name: "Amazon"},
        {id: "Auth0", name: "Auth0"},
        {id: "BattleNet", name: "Battle.net"},
        {id: "Bitbucket", name: "Bitbucket"},
        {id: "Box", name: "Box"},
        {id: "CloudFoundry", name: "Cloud Foundry"},
        {id: "Dailymotion", name: "Dailymotion"},
        {id: "Deezer", name: "Deezer"},
        {id: "DigitalOcean", name: "DigitalOcean"},
        {id: "Discord", name: "Discord"},
        {id: "Dropbox", name: "Dropbox"},
        {id: "EveOnline", name: "Eve Online"},
        {id: "Fitbit", name: "Fitbit"},
        {id: "Gitea", name: "Gitea"},
        {id: "Heroku", name: "Heroku"},
        {id: "InfluxCloud", name: "InfluxCloud"},
        {id: "Instagram", name: "Instagram"},
        {id: "Intercom", name: "Intercom"},
        {id: "Kakao", name: "Kakao"},
        {id: "Lastfm", name: "Lastfm"},
        {id: "Mailru", name: "Mailru"},
        {id: "Meetup", name: "Meetup"},
        {id: "MicrosoftOnline", name: "MicrosoftOnline"},
        {id: "Naver", name: "Naver"},
        {id: "Nextcloud", name: "Nextcloud"},
        {id: "OneDrive", name: "OneDrive"},
        {id: "Oura", name: "Oura"},
        {id: "Patreon", name: "Patreon"},
        {id: "PayPal", name: "PayPal"},
        {id: "SalesForce", name: "SalesForce"},
        {id: "Shopify", name: "Shopify"},
        {id: "Soundcloud", name: "Soundcloud"},
        {id: "Spotify", name: "Spotify"},
        {id: "Strava", name: "Strava"},
        {id: "Stripe", name: "Stripe"},
        {id: "Telegram", name: "Telegram"},
        {id: "TikTok", name: "TikTok"},
        {id: "Tumblr", name: "Tumblr"},
        {id: "Twitch", name: "Twitch"},
        {id: "Twitter", name: "Twitter"},
        {id: "Typetalk", name: "Typetalk"},
        {id: "Uber", name: "Uber"},
        {id: "VK", name: "VK"},
        {id: "Wepay", name: "Wepay"},
        {id: "Xero", name: "Xero"},
        {id: "Yahoo", name: "Yahoo"},
        {id: "Yammer", name: "Yammer"},
        {id: "Yandex", name: "Yandex"},
        {id: "Zoom", name: "Zoom"},
        {id: "Custom", name: "Custom"},
        {id: "Custom2", name: "Custom2"},
        {id: "Custom3", name: "Custom3"},
        {id: "Custom4", name: "Custom4"},
        {id: "Custom5", name: "Custom5"},
        {id: "Custom6", name: "Custom6"},
        {id: "Custom7", name: "Custom7"},
        {id: "Custom8", name: "Custom8"},
        {id: "Custom9", name: "Custom9"},
        {id: "Custom10", name: "Custom10"},
        {id: "Custom Flexible", name: "Custom Flexible"},
      ]
    );
  } else if (category === "Email") {
    return (
      [
        {id: "Default", name: "Default"},
        {id: "SUBMAIL", name: "SUBMAIL"},
        {id: "Mailtrap", name: "Mailtrap"},
        {id: "Azure ACS", name: "Azure ACS"},
        {id: "SendGrid", name: "SendGrid"},
        {id: "Custom HTTP Email", name: "Custom HTTP Email"},
        {id: "Resend", name: "Resend"},
      ]
    );
  } else if (category === "SMS") {
    return (
      [
        {id: "Aliyun SMS", name: "Alibaba Cloud SMS"},
        {id: "Alibaba Cloud PNVS SMS", name: "Alibaba Cloud PNVS SMS"},
        {id: "Amazon SNS", name: "Amazon SNS"},
        {id: "Azure ACS", name: "Azure ACS"},
        {id: "Custom HTTP SMS", name: "Custom HTTP SMS"},
        {id: "Mock SMS", name: "Mock SMS"},
        {id: "OSON SMS", name: "OSON SMS"},
        {id: "Infobip SMS", name: "Infobip SMS"},
        {id: "Tencent Cloud SMS", name: "Tencent Cloud SMS"},
        {id: "Baidu Cloud SMS", name: "Baidu Cloud SMS"},
        {id: "Volc Engine SMS", name: "Volc Engine SMS"},
        {id: "Huawei Cloud SMS", name: "Huawei Cloud SMS"},
        {id: "UCloud SMS", name: "UCloud SMS"},
        {id: "Twilio SMS", name: "Twilio SMS"},
        {id: "SmsBao SMS", name: "SmsBao SMS"},
        {id: "SUBMAIL SMS", name: "SUBMAIL SMS"},
        {id: "Msg91 SMS", name: "Msg91 SMS"},
      ]
    );
  } else if (category === "Storage") {
    return (
      [
        {id: "Local File System", name: "Local File System"},
        {id: "AWS S3", name: "AWS S3"},
        {id: "MinIO", name: "MinIO"},
        {id: "Aliyun OSS", name: "Alibaba Cloud OSS"},
        {id: "Tencent Cloud COS", name: "Tencent Cloud COS"},
        {id: "Azure Blob", name: "Azure Blob"},
        {id: "Qiniu Cloud Kodo", name: "Qiniu Cloud Kodo"},
        {id: "Google Cloud Storage", name: "Google Cloud Storage"},
        {id: "Synology", name: "Synology"},
        {id: "Casdoor", name: "Casdoor"},
        {id: "CUCloud OSS", name: "CUCloud OSS"},
      ]
    );
  } else if (category === "SAML") {
    return ([
      {id: "Aliyun IDaaS", name: "Aliyun IDaaS"},
      {id: "Keycloak", name: "Keycloak"},
      {id: "Custom", name: "Custom"},
      {id: "Custom Flexible", name: "Custom Flexible"},
    ]);
  } else if (category === "Payment") {
    return ([
      {id: "Dummy", name: "Dummy"},
      {id: "Balance", name: "Balance"},
      {id: "Alipay", name: "Alipay"},
      {id: "WeChat Pay", name: "WeChat Pay"},
      {id: "PayPal", name: "PayPal"},
      {id: "Stripe", name: "Stripe"},
      {id: "AirWallex", name: "AirWallex"},
      {id: "GC", name: "GC"},
      {id: "Polar", name: "Polar"},
      {id: "Paddle", name: "Paddle"},
      {id: "FastSpring", name: "FastSpring"},
      {id: "Lemon Squeezy", name: "Lemon Squeezy"},
      {id: "Adyen", name: "Adyen"},
    ]);
  } else if (category === "Captcha") {
    return ([
      {id: "Default", name: "Default"},
      {id: "reCAPTCHA v2", name: "reCAPTCHA v2"},
      {id: "reCAPTCHA v3", name: "reCAPTCHA v3"},
      {id: "hCaptcha", name: "hCaptcha"},
      {id: "Aliyun Captcha", name: "Aliyun Captcha"},
      {id: "GEETEST", name: "GEETEST"},
      {id: "Cloudflare Turnstile", name: "Cloudflare Turnstile"},
    ]);
  } else if (category === "Web3") {
    return ([
      {id: "MetaMask", name: "MetaMask"},
      {id: "Web3Onboard", name: "Web3-Onboard"},
    ]);
  } else if (category === "Notification") {
    return ([
      {id: "Telegram", name: "Telegram"},
      {id: "Custom HTTP", name: "Custom HTTP"},
      {id: "DingTalk", name: "DingTalk"},
      {id: "Lark", name: "Lark"},
      {id: "Microsoft Teams", name: "Microsoft Teams"},
      {id: "Bark", name: "Bark"},
      {id: "Pushover", name: "Pushover"},
      {id: "Pushbullet", name: "Pushbullet"},
      {id: "Slack", name: "Slack"},
      {id: "Webpush", name: "Webpush"},
      {id: "Discord", name: "Discord"},
      {id: "Google Chat", name: "Google Chat"},
      {id: "Line", name: "Line"},
      {id: "Matrix", name: "Matrix"},
      {id: "Twitter", name: "Twitter"},
      {id: "Reddit", name: "Reddit"},
      {id: "Rocket Chat", name: "Rocket Chat"},
      {id: "Viber", name: "Viber"},
      {id: "CUCloud", name: "CUCloud"},
      {id: "WeCom", name: "WeCom"},
    ]);
  } else if (category === "Face ID") {
    return ([
      {id: "Alibaba Cloud Facebody", name: "Alibaba Cloud Facebody"},
      {id: "Local UniFace", name: "Local UniFace"},
    ]);
  } else if (category === "MFA") {
    return ([
      {id: "RADIUS", name: "RADIUS"},
    ]);
  } else if (category === "ID Verification") {
    return ([
      {id: "Jumio", name: "Jumio"},
      {id: "Alibaba Cloud", name: "Alibaba Cloud"},
    ]);
  } else if (category === "Log") {
    return ([
      {id: "Casdoor Permission Log", name: "Casdoor Permission Log"},
      {id: "System Log", name: "System Log"},
      {id: "Agent", name: "Agent"},
      {id: "SELinux Log", name: "SELinux Log"},
    ]);
  } else if (category === "Scan") {
    return ([
      {id: "Security Scan", name: "Security Scan"},
      {id: "MCP Scan", name: "MCP Scan"},
    ]);
  } else {
    return [];
  }
}

export function getCryptoAlgorithmOptions(cryptoAlgorithm) {
  if (cryptoAlgorithm.startsWith("ES")) {
    return [];
  } else {
    return (
      [
        {id: 1024, name: "1024"},
        {id: 2048, name: "2048"},
        {id: 4096, name: "4096"},
      ]
    );
  }
}

export function isSigninMethodHidden(signinMethod) {
  // the "Hide password" rule used to be named "Hide-Password", the applications
  // configured before the rename still store the legacy value
  return signinMethod?.rule === "Hide password" || signinMethod?.rule === "Hide-Password";
}

function isSigninMethodEnabled(application, signinMethod) {
  if (application && application.signinMethods) {
    return application.signinMethods.filter(item => item.name === signinMethod && !isSigninMethodHidden(item)).length > 0;
  } else {
    return false;
  }
}

export const CaptchaRule = {
  Always: "Always",
  Never: "Never",
  Dynamic: "Dynamic",
  InternetOnly: "Internet-Only",
};

export function getCaptchaProviderItems(application) {
  const providers = application?.providers;
  if (!providers) {
    return [];
  }

  return providers.filter(providerItem => providerItem?.provider?.category === "Captcha");
}

export function getCaptchaRule(application) {
  const captchaProviderItems = getCaptchaProviderItems(application);
  if (captchaProviderItems.some(providerItem => providerItem.rule === CaptchaRule.Always)) {
    return CaptchaRule.Always;
  } else if (captchaProviderItems.some(providerItem => providerItem.rule === CaptchaRule.Dynamic)) {
    return CaptchaRule.Dynamic;
  } else if (captchaProviderItems.some(providerItem => providerItem.rule === CaptchaRule.InternetOnly)) {
    return CaptchaRule.InternetOnly;
  }

  return CaptchaRule.Never;
}

export function isInlineCaptchaEnabled(application) {
  return application?.signinItems?.some(signinItem => signinItem.name === "Captcha" && signinItem.rule === "inline") || false;
}

export function isCaptchaEnabled(application) {
  return getCaptchaRule(application) !== CaptchaRule.Never;
}

export function isPasswordEnabled(application) {
  return isSigninMethodEnabled(application, "Password");
}

export function isCodeSigninEnabled(application) {
  return isSigninMethodEnabled(application, "Verification code");
}

export function isWebAuthnEnabled(application) {
  return isSigninMethodEnabled(application, "WebAuthn");
}

export function isLdapEnabled(application) {
  return isSigninMethodEnabled(application, "LDAP");
}

export function isFaceIdEnabled(application) {
  return isSigninMethodEnabled(application, "Face ID");
}


export function getOption(label, value) {
  return {
    label,
    value,
  };
}

// getDisplayNameOption returns an option whose value is "owner/name" and whose label is
// "displayName (owner/name)", so that the item can be recognized by both its display name and its ID.
export function getDisplayNameOption(item) {
  const id = `${item.owner}/${item.name}`;
  return getOption(item.displayName ? `${item.displayName} (${id})` : id, id);
}

export function getArrayItem(array, key, value) {
  const res = array.filter(item => item[key] === value)[0];
  return res;
}

export function getDeduplicatedArray(array, filterArray, key) {
  const res = array.filter(item => !filterArray.some(tableItem => tableItem[key] === item[key]));
  return res;
}

export function getNewRowNameForTable(table, rowName) {
  const emptyCount = table.filter(row => row.name.includes(rowName)).length;
  let res = rowName;
  for (let i = 0; i < emptyCount; i++) {
    res = res + " ";
  }
  return res;
}


export function getTagColor(_s?: string) {
  return "processing";
}


export function getApplicationName(application) {
  let name = `${application?.owner}/${application?.name}`;

  if (application?.isShared && application?.organization) {
    name += `-org-${application.organization}`;
  }

  return name;
}

export function getApplicationDisplayName(application) {
  if (application.isShared) {
    return `${application.name}(Shared)`;
  }
  return application.name;
}

export function getRandomName() {
  return Math.random().toString(36).slice(-6);
}

export function getRandomNumber() {
  return Math.random().toString(10).slice(-11);
}

export function getFromLink() {
  const from = sessionStorage.getItem("from");
  if (from === null) {
    return "/";
  }
  return from;
}

export function scrollToDiv(divId) {
  if (divId) {
    const ele = document.getElementById(divId);
    if (ele) {
      ele.scrollIntoView({behavior: "smooth"});
    }
  }
}

export function inIframe() {
  try {
    return window !== window.parent;
  } catch (e) {
    return true;
  }
}

export function getOrganization() {
  const organization = localStorage.getItem("organization");
  return organization !== null ? organization : "All";
}

export function setOrganization(organization) {
  localStorage.setItem("organization", organization);
  window.dispatchEvent(new Event("storageOrganizationChanged"));
}

export function getRequestOrganization(account) {
  if (isAdminUser(account)) {
    return getOrganization() === "All" ? account.owner : getOrganization();
  }
  return account.owner;
}

export function isDefaultOrganizationSelected(account) {
  if (isAdminUser(account)) {
    return getOrganization() === "All";
  }
  return false;
}

const BuiltInObjects = [
  "api-enforcer-built-in",
  "user-enforcer-built-in",
  "api-model-built-in",
  "user-model-built-in",
  "api-adapter-built-in",
  "user-adapter-built-in",
];

export function builtInObject(obj) {
  if (obj === undefined || obj === null) {
    return false;
  }
  return obj.owner === "built-in" && BuiltInObjects.includes(obj.name);
}


export const CurrencyOptions = [
  {id: "USD", name: "USD"},
  {id: "CNY", name: "CNY"},
  {id: "EUR", name: "EUR"},
  {id: "JPY", name: "JPY"},
  {id: "GBP", name: "GBP"},
  {id: "AUD", name: "AUD"},
  {id: "CAD", name: "CAD"},
  {id: "CHF", name: "CHF"},
  {id: "HKD", name: "HKD"},
  {id: "SGD", name: "SGD"},
  {id: "BRL", name: "BRL"},
  {id: "PLN", name: "PLN"},
  {id: "KRW", name: "KRW"},
  {id: "INR", name: "INR"},
  {id: "RUB", name: "RUB"},
  {id: "MXN", name: "MXN"},
  {id: "ZAR", name: "ZAR"},
  {id: "TRY", name: "TRY"},
  {id: "SEK", name: "SEK"},
  {id: "NOK", name: "NOK"},
  {id: "DKK", name: "DKK"},
  {id: "THB", name: "THB"},
  {id: "MYR", name: "MYR"},
  {id: "TWD", name: "TWD"},
  {id: "CZK", name: "CZK"},
  {id: "HUF", name: "HUF"},
];

export function getCurrencySymbol(currency) {
  if (currency === "USD" || currency === "usd") {
    return "$";
  } else if (currency === "CNY" || currency === "cny") {
    return "¥";
  } else if (currency === "EUR" || currency === "eur") {
    return "€";
  } else if (currency === "JPY" || currency === "jpy") {
    return "¥";
  } else if (currency === "GBP" || currency === "gbp") {
    return "£";
  } else if (currency === "AUD" || currency === "aud") {
    return "A$";
  } else if (currency === "CAD" || currency === "cad") {
    return "C$";
  } else if (currency === "CHF" || currency === "chf") {
    return "CHF";
  } else if (currency === "HKD" || currency === "hkd") {
    return "HK$";
  } else if (currency === "SGD" || currency === "sgd") {
    return "S$";
  } else if (currency === "BRL" || currency === "brl") {
    return "R$";
  } else if (currency === "PLN" || currency === "pln") {
    return "zł";
  } else if (currency === "KRW" || currency === "krw") {
    return "₩";
  } else if (currency === "INR" || currency === "inr") {
    return "₹";
  } else if (currency === "RUB" || currency === "rub") {
    return "₽";
  } else if (currency === "MXN" || currency === "mxn") {
    return "$";
  } else if (currency === "ZAR" || currency === "zar") {
    return "R";
  } else if (currency === "TRY" || currency === "try") {
    return "₺";
  } else if (currency === "SEK" || currency === "sek") {
    return "kr";
  } else if (currency === "NOK" || currency === "nok") {
    return "kr";
  } else if (currency === "DKK" || currency === "dkk") {
    return "kr";
  } else if (currency === "THB" || currency === "thb") {
    return "฿";
  } else if (currency === "MYR" || currency === "myr") {
    return "RM";
  } else if (currency === "TWD" || currency === "twd") {
    return "NT$";
  } else if (currency === "CZK" || currency === "czk") {
    return "Kč";
  } else if (currency === "HUF" || currency === "huf") {
    return "Ft";
  } else {
    return currency;
  }
}

export function getCurrencyCountryCode(currency) {
  const currencyToCountry = {
    USD: "US",
    CNY: "CN",
    EUR: "EU",
    JPY: "JP",
    GBP: "GB",
    AUD: "AU",
    CAD: "CA",
    CHF: "CH",
    HKD: "HK",
    SGD: "SG",
    BRL: "BR",
    PLN: "PL",
    KRW: "KR",
    INR: "IN",
    RUB: "RU",
    MXN: "MX",
    ZAR: "ZA",
    TRY: "TR",
    SEK: "SE",
    NOK: "NO",
    DKK: "DK",
    THB: "TH",
    MYR: "MY",
    TWD: "TW",
    CZK: "CZ",
    HUF: "HU",
  };

  return currencyToCountry[currency?.toUpperCase()] || null;
}

export function getCurrencyFlag(currency) {
  const countryCode = getCurrencyCountryCode(currency);
  if (!countryCode) {
    return null;
  }

  return (
    <img src={`${StaticBaseUrl}/flag-icons/${countryCode}.svg`} alt={`${currency} flag`} height={20} style={{marginRight: 5}} />
  );
}

export function getCurrencyWithFlag(currency) {
  const translationKey = `currency:${currency}`;
  const translatedText = i18next.t(translationKey);
  const currencyText = translatedText === translationKey ? currency : translatedText;

  const countryCode = getCurrencyCountryCode(currency);
  if (!countryCode) {
    return currencyText;
  }

  return (
    <span>
      <img src={`${StaticBaseUrl}/flag-icons/${countryCode}.svg`} alt={`${currency} flag`} height={20} style={{marginRight: 5}} />
      {currencyText}
    </span>
  );
}

export function getPriceDisplay(price, currency) {
  const priceValue = price || 0;
  const currencyValue = currency || "USD";
  return (
    <>
      {getCurrencyFlag(currencyValue)} {getCurrencySymbol(currencyValue)}{priceValue} ({getCurrencyText(currencyValue)})
    </>
  );
}

export function getUserCommonFields() {
  return ["Owner", "Name", "CreatedTime", "UpdatedTime", "DeletedTime", "Id", "ExternalId", "Type", "Password", "PasswordSalt", "PasswordType", "DisplayName", "FirstName", "LastName", "Avatar", "AvatarType", "PermanentAvatar",
    "Email", "EmailVerified", "Phone", "CountryCode", "Location", "Address", "Affiliation", "Title", "IdCardType", "IdCard", "RealName", "IsVerified", "Homepage", "Bio", "Tag", "Region",
    "Language", "Gender", "Birthday", "Education", "UidNumber", "Score", "Karma", "Ranking", "Balance", "BalanceCredit", "Currency", "BalanceCurrency", "IsDefaultAvatar", "IsOnline", "IsAdmin", "IsForbidden", "IsDeleted",
    "SignupApplication", "RegisterType", "RegisterSource", "CreatedIp", "LastSigninTime", "LastSigninIp",
    "PreferredMfaType", "TotpSecret", "RecoveryCodes", "MfaPhoneEnabled", "MfaEmailEnabled", "MfaRadiusEnabled", "MfaRadiusUsername", "MfaRadiusProvider", "MfaPushEnabled", "MfaPushReceiver", "MfaPushProvider",
    "WebauthnCredentials", "FaceIds", "Invitation", "InvitationCode", "Ldap", "Properties", "Groups"];
}

export function getDefaultFooterContent() {
  return `Powered by <a target="_blank" href="https://casdoor.org" rel="noreferrer"><img style="padding-bottom: 3px" height="20" alt="Casdoor" src="${StaticBaseUrl}/img/casdoor-logo_1185x256.png"/></a>`;
}

export function getEmptyFooterContent() {
  return `<style>
    #footer {
        display: none;
    }
<style>
  `;
}

export function getDefaultHtmlEmailContent() {
  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Verification Code Email</title>
<style>
    body { font-family: Arial, sans-serif; }
    .email-container { width: 600px; margin: 0 auto; }
    .header { text-align: center; }
    .code { font-size: 24px; margin: 20px 0; text-align: center; }
    .footer { font-size: 12px; text-align: center; margin-top: 50px; }
    .footer a { color: #000; text-decoration: none; }
</style>
</head>
<body>
<div class="email-container">
  <div class="header">
        <h3>Casbin Organization</h3>
        <img src="${StaticBaseUrl}/img/casdoor-logo_1185x256.png" alt="Casdoor Logo" width="300">
    </div>
    <p><strong>%{user.friendlyName}</strong>, here is your verification code</p>
    <p>Use this code for your transaction. It's valid for 5 minutes</p>
    <div class="code">
        %s
    </div>
    <reset-link>
      <div class="link">
         Or click this <a href="%link">link</a> to reset
      </div>
    </reset-link>
    <p>Thanks</p>
    <p>Casbin Team</p>
    <hr>
    <div class="footer">
        <p>Casdoor is a brand operated by Casbin organization. For more info please refer to <a href="https://casdoor.org">https://casdoor.org</a></p>
    </div>
</div>
</body>
</html>`;
}

export function getDefaultInvitationHtmlEmailContent() {
  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Invitation Code Email</title>
<style>
    body { font-family: Arial, sans-serif; }
    .email-container { width: 600px; margin: 0 auto; }
    .header { text-align: center; }
    .code { font-size: 24px; margin: 20px 0; text-align: center; }
    .footer { font-size: 12px; text-align: center; margin-top: 50px; }
    .footer a { color: #000; text-decoration: none; }
</style>
</head>
<body>
<div class="email-container">
  <div class="header">
        <h3>Casbin Organization</h3>
        <img src="${StaticBaseUrl}/img/casdoor-logo_1185x256.png" alt="Casdoor Logo" width="300">
    </div>
    <p>You have been invited into Casdoor</p>
    <div class="code">
        %code
    </div>
    <reset-link>
      <div class="link">
         Or click this <a href="%link">link</a> to signup
      </div>
    </reset-link>
    <p>Thanks</p>
    <p>Casbin Team</p>
    <hr>
    <div class="footer">
        <p>Casdoor is a brand operated by Casbin organization. For more info please refer to <a href="https://casdoor.org">https://casdoor.org</a></p>
    </div>
</div>
</body>
</html>`;
}

export function getCurrencyText(currency) {
  if (currency === "USD") {
    return i18next.t("currency:USD");
  } else if (currency === "CNY") {
    return i18next.t("currency:CNY");
  } else if (currency === "EUR") {
    return i18next.t("currency:EUR");
  } else if (currency === "JPY") {
    return i18next.t("currency:JPY");
  } else if (currency === "GBP") {
    return i18next.t("currency:GBP");
  } else if (currency === "AUD") {
    return i18next.t("currency:AUD");
  } else if (currency === "CAD") {
    return i18next.t("currency:CAD");
  } else if (currency === "CHF") {
    return i18next.t("currency:CHF");
  } else if (currency === "HKD") {
    return i18next.t("currency:HKD");
  } else if (currency === "SGD") {
    return i18next.t("currency:SGD");
  } else if (currency === "BRL") {
    return i18next.t("currency:BRL");
  } else if (currency === "PLN") {
    return i18next.t("currency:PLN");
  } else if (currency === "KRW") {
    return i18next.t("currency:KRW");
  } else if (currency === "INR") {
    return i18next.t("currency:INR");
  } else if (currency === "RUB") {
    return i18next.t("currency:RUB");
  } else if (currency === "MXN") {
    return i18next.t("currency:MXN");
  } else if (currency === "ZAR") {
    return i18next.t("currency:ZAR");
  } else if (currency === "TRY") {
    return i18next.t("currency:TRY");
  } else if (currency === "SEK") {
    return i18next.t("currency:SEK");
  } else if (currency === "NOK") {
    return i18next.t("currency:NOK");
  } else if (currency === "DKK") {
    return i18next.t("currency:DKK");
  } else if (currency === "THB") {
    return i18next.t("currency:THB");
  } else if (currency === "MYR") {
    return i18next.t("currency:MYR");
  } else if (currency === "TWD") {
    return i18next.t("currency:TWD");
  } else if (currency === "CZK") {
    return i18next.t("currency:CZK");
  } else if (currency === "HUF") {
    return i18next.t("currency:HUF");
  } else {
    return "(Unknown currency)";
  }
}

export function isDarkTheme(themeAlgorithm) {
  return themeAlgorithm && themeAlgorithm.includes("dark");
}

/**
 * Applies a saved Form to a list page: keeps only the visible items, in the
 * order the form lists them. The action column is appended by CrudListPage, so
 * unlike the antd version this only deals with the data columns.
 */
export function filterTableColumns(columns, formItems) {
  if (!formItems || formItems.length === 0) {
    return columns;
  }

  return formItems
    .filter(item => item.visible !== false)
    .map(item => {
      const matchedColumn = columns.find(column => (column.key ?? column.dataIndex) === item.name);
      if (!matchedColumn) {
        return null;
      }
      return {
        ...matchedColumn,
        width: item.width !== undefined ? `${item.width}px` : matchedColumn.width,
        title: item.width !== undefined ? i18next.t(item.label) : matchedColumn.title,
      };
    })
    .filter(column => column !== null);
}

export function getFormTypeOptions() {
  return [
    {id: "users", name: "general:Users"},
    {id: "providers", name: "application:Providers"},
    {id: "applications", name: "general:Applications"},
    {id: "organizations", name: "general:Organizations"},
  ];
}

export function getFormTypeItems(formType) {
  if (formType === "users") {
    return [
      {name: "owner", label: "general:Organization", visible: true, width: "150"},
      {name: "signupApplication", label: "general:Application", visible: true, width: "120"},
      {name: "name", label: "general:Name", visible: true, width: "110"},
      {name: "createdTime", label: "general:Created time", visible: true, width: "160"},
      {name: "displayName", label: "general:Display name", visible: true, width: "150"},
      {name: "avatar", label: "general:Avatar", visible: true, width: "80"},
      {name: "email", label: "general:Email", visible: true, width: "160"},
      {name: "phone", label: "general:Phone", visible: true, width: "120"},
      {name: "affiliation", label: "user:Affiliation", visible: true, width: "140"},
      {name: "region", label: "user:Country/Region", visible: true, width: "140"},
      {name: "type", label: "general:User type", visible: true, width: "120"},
      {name: "tag", label: "general:Tag", visible: true, width: "110"},
      {name: "isAdmin", label: "user:Is admin", visible: true, width: "120"},
      {name: "isForbidden", label: "user:Is forbidden", visible: true, width: "110"},
      {name: "isDeleted", label: "user:Is deleted", visible: true, width: "110"},
    ];
  } else if (formType === "providers") {
    return [
      {name: "name", label: "general:Name", visible: true, width: "120"},
      {name: "owner", label: "general:Organization", visible: true, width: "150"},
      {name: "createdTime", label: "general:Created time", visible: true, width: "180"},
      {name: "displayName", label: "general:Display name", visible: true, width: "150"},
      {name: "category", label: "general:Category", visible: true, width: "110"},
      {name: "type", label: "general:Type", visible: true, width: "110"},
      {name: "clientId", label: "provider:Client ID", visible: true, width: "100"},
      {name: "providerUrl", label: "provider:Provider URL", visible: true, width: "150"},
    ];
  } else if (formType === "applications") {
    return [
      {name: "name", label: "general:Name", visible: true, width: "150"},
      {name: "createdTime", label: "general:Created time", visible: true, width: "160"},
      {name: "displayName", label: "general:Display name", visible: true, width: "150"},
      {name: "logo", label: "Logo", visible: true, width: "200"},
      {name: "organization", label: "general:Organization", visible: true, width: "150"},
      {name: "providers", label: "application:Providers", visible: true, width: "500"},
    ];
  } else if (formType === "organizations") {
    return [
      {name: "name", label: "general:Name", visible: true, width: "120"},
      {name: "createdTime", label: "general:Created time", visible: true, width: "160"},
      {name: "displayName", label: "general:Display name", visible: true, width: "150"},
      {name: "favicon", label: "general:Favicon", visible: true, width: "50"},
      {name: "websiteUrl", label: "organization:Website URL", visible: true, width: "200"},
      {name: "passwordType", label: "general:Password type", visible: true, width: "150"},
      {name: "passwordSalt", label: "general:Password salt", visible: true, width: "150"},
      {name: "defaultAvatar", label: "general:Default avatar", visible: true, width: "120"},
      {name: "enableSoftDeletion", label: "organization:Soft deletion", visible: true, width: "140"},
    ];
  } else {
    return [];
  }
}


export function getApiPaths() {
  const objects = ["organization", "group", "user", "application", "provider", "resource", "cert", "role", "permission", "model", "adapter", "enforcer", "session", "token", "product", "payment", "plan", "pricing", "subscription", "syncer", "webhook", "form", "invitation", "ldap", "order", "ticket", "transaction"];
  const res = [];

  // Auth and user session APIs
  res.push("signup", "login", "logout", "sso-logout", "unlink");
  res.push("new-user"); // Custom event for new user creation
  res.push("new-user-ldap"); // Custom event for new user creation via LDAP sync
  res.push("new-user-syncer"); // Custom event for new user creation via syncer

  // CRUD operations for objects
  objects.forEach(obj => {
    ["add", "update", "delete"].forEach(action => {
      res.push(`${action}-${obj}`);
    });
    if (obj === "payment") {
      res.push("invoice-payment", "notify-payment");
    }
    if (obj === "order") {
      res.push("place-order", "cancel-order", "pay-order");
    }
    if (obj === "user") {
      res.push("remove-user-from-group", "upload-users");
      res.push("check-user-password", "set-password", "reset-email-or-phone");
      res.push("verify-identification");
    }
    if (obj === "group") {
      res.push("upload-groups");
    }
    if (obj === "role") {
      res.push("upload-roles");
    }
    if (obj === "permission") {
      res.push("upload-permissions");
    }
    if (obj === "resource") {
      res.push("upload-resource");
    }
    if (obj === "invitation") {
      res.push("send-invitation", "verify-invitation");
    }
    if (obj === "ticket") {
      res.push("add-ticket-message");
    }
    if (obj === "syncer") {
      res.push("run-syncer", "test-syncer-db");
    }
    if (obj === "ldap") {
      res.push("sync-ldap-users");
    }
    if (obj === "enforcer") {
      res.push("enforce", "batch-enforce");
    }
    if (obj === "session") {
      res.push("is-session-duplicated");
    }
  });

  // Special cases that don't follow the standard pattern
  res.push("add-policy", "update-policy", "remove-policy");
  res.push("add-record");
  res.push("delete-mfa", "set-preferred-mfa");

  // MFA setup APIs
  res.push("mfa/setup/initiate", "mfa/setup/verify", "mfa/setup/enable");

  // WebAuthn APIs
  res.push("webauthn/signup/begin", "webauthn/signup/finish");
  res.push("webauthn/signin/begin", "webauthn/signin/finish");

  // OAuth APIs
  res.push("login/oauth/access_token", "login/oauth/refresh_token", "login/oauth/introspect");

  // Verification and communication APIs
  res.push("send-verification-code", "verify-code", "verify-captcha");
  res.push("send-email", "send-sms", "send-notification");

  // SAML APIs
  res.push("acs", "saml/metadata");

  // Casbin engine APIs
  res.push("run-casbin-command", "refresh-engines");

  // Monitoring and health APIs
  res.push("health", "metrics");

  // Other APIs
  res.push("callback", "device-auth", "faceid-signin-begin");
  res.push("user", "userinfo");

  return res;
}

export function getItemId(item) {
  return item.owner + "/" + item.name;
}

export function getVersionInfo(text, siteName) {
  if (text === "") {
    return null;
  }

  try {
    const versionInfo = JSON.parse(text);
    const link = versionInfo?.version !== "" ? `${getRepoUrl(siteName)}/releases/tag/${versionInfo?.version}` : "";
    let versionText = versionInfo?.version !== "" ? versionInfo?.version : "Unknown version";
    if (versionInfo?.commitOffset > 0) {
      versionText += ` (ahead+${versionInfo?.commitOffset})`;
    }

    return {text: versionText, link: link};
  } catch (e) {
    return {text: "", link: ""};
  }
}

function getOriginalName(name) {
  const tokens = name.split("_");
  if (tokens.length > 0) {
    return tokens[0];
  } else {
    return name;
  }
}

export function getRepoUrl(name) {
  name = getOriginalName(name);
  if (name === "casdoor") {
    return "https://github.com/casdoor/casdoor";
  } else {
    return `https://github.com/casbin/${name}`;
  }
}

export function createFormAndSubmit(url, params) {
  const form = document.createElement("form");
  form.method = "post";
  form.action = url;

  for (const k in params) {
    if (!params[k]) {
      continue;
    }
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = k;
    input.value = params[k];
    form.appendChild(input);
  }

  document.body.appendChild(form);
  form.submit();
  setTimeout(() => {form.remove();}, 500);
}

export function getLoginLink(application) {
  let url;
  if (application === null) {
    url = null;
  } else if (window.location.pathname.includes("/signup/oauth/authorize")) {
    url = window.location.pathname.replace("/signup/oauth/authorize", "/login/oauth/authorize");
  } else if (authConfig.appName === application.name) {
    url = "/login";
  } else if (application.signinUrl === "") {
    url = trim(application.homepageUrl, "/") + "/login";
  } else {
    url = application.signinUrl;
  }
  return url + window.location.search;
}
