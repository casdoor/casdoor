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
import {Link} from "react-router-dom";
import {Button, Select, Tag, Tooltip, message, theme} from "antd";
import {QuestionCircleTwoTone} from "@ant-design/icons";
import {isMobile as isMobileDevice} from "react-device-detect";
import "./i18n";
import i18next from "i18next";
import copy from "copy-to-clipboard";
import {authConfig} from "./auth/Auth";
import {Helmet} from "react-helmet";
import * as Conf from "./Conf";
import * as phoneNumber from "libphonenumber-js";
import moment from "moment";
import {MfaAuthVerifyForm, NextMfa, RequiredMfa} from "./auth/mfa/MfaAuthVerifyForm";
import {EmailMfaType, SmsMfaType, TotpMfaType} from "./auth/MfaSetupPage";

const {Option} = Select;

export const ServerUrl = "";

export const StaticBaseUrl = "https://cdn.casbin.org";

export const Countries = [
  {label: "English", key: "en", country: "US", alt: "English"},
  {label: "Español", key: "es", country: "ES", alt: "Español"},
  {label: "Français", key: "fr", country: "FR", alt: "Français"},
  {label: "Deutsch", key: "de", country: "DE", alt: "Deutsch"},
  {label: "中文", key: "zh", country: "CN", alt: "中文"},
  {label: "Indonesia", key: "id", country: "ID", alt: "Indonesia"},
  {label: "日本語", key: "ja", country: "JP", alt: "日本語"},
  {label: "한국어", key: "ko", country: "KR", alt: "한국어"},
  {label: "Русский", key: "ru", country: "RU", alt: "Русский"},
  {label: "TiếngViệt", key: "vi", country: "VN", alt: "TiếngViệt"},
  {label: "Português", key: "pt", country: "PT", alt: "Português"},
  {label: "Italiano", key: "it", country: "IT", alt: "Italiano"},
  {label: "Malay", key: "ms", country: "MY", alt: "Malay"},
  {label: "Türkçe", key: "tr", country: "TR", alt: "Türkçe"},
  {label: "لغة عربية", key: "ar", country: "SA", alt: "لغة عربية"},
  {label: "עִבְרִית", key: "he", country: "IL", alt: "עִבְרִית"},
  {label: "Nederlands", key: "nl", country: "NL", alt: "Nederlands"},
  {label: "Polski", key: "pl", country: "PL", alt: "Polski"},
  {label: "Suomi", key: "fi", country: "FI", alt: "Suomi"},
  {label: "Svenska", key: "sv", country: "SE", alt: "Svenska"},
  {label: "Українська", key: "uk", country: "UA", alt: "Українська"},
  {label: "Қазақ", key: "kk", country: "KZ", alt: "Қазақ"},
  {label: "فارسی", key: "fa", country: "IR", alt: "فارسی"},
  {label: "Čeština", key: "cs", country: "CZ", alt: "Čeština"},
  {label: "Slovenčina", key: "sk", country: "SK", alt: "Slovenčina"},
];

export function getThemeData(organization, application) {
  if (application?.themeData?.isEnabled) {
    return application.themeData;
  } else if (organization?.themeData?.isEnabled) {
    return organization.themeData;
  } else {
    return Conf.ThemeDefault;
  }
}

export function getAlgorithm(themeAlgorithmNames) {
  return themeAlgorithmNames.sort().reverse().map((algorithmName) => {
    if (algorithmName === "dark") {
      return theme.darkAlgorithm;
    }
    if (algorithmName === "compact") {
      return theme.compactAlgorithm;
    }
    return theme.defaultAlgorithm;
  });
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
      logo: "https://osonsms.com/images/osonsms-logo.svg",
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
      logo: "https://min.io/resources/img/logo.svg",
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
  },
  "Face ID": {
    "Alibaba Cloud Facebody": {
      logo: `${StaticBaseUrl}/img/social_aliyun.png`,
      url: "https://vision.aliyun.com/facebody",
    },
  },
};

export function initCountries() {
  const countries = require("i18n-iso-countries");
  countries.registerLocale(require("i18n-iso-countries/langs/" + getLanguage() + ".json"));
  return countries;
}

export function getCountryCode(country) {
  if (phoneNumber.isSupportedCountry(country)) {
    return phoneNumber.getCountryCallingCode(country);
  }
  return "";
}

export function getCountryCodeData(countryCodes = phoneNumber.getCountries()) {
  if (countryCodes?.includes("All")) {
    countryCodes = phoneNumber.getCountries();
  }
  return countryCodes?.map((countryCode) => {
    if (phoneNumber.isSupportedCountry(countryCode)) {
      const name = initCountries().getName(countryCode, getLanguage());
      return {
        code: countryCode,
        name: name || "",
        phone: phoneNumber.getCountryCallingCode(countryCode),
      };
    }
  }).filter(item => item.name !== "")
    .sort((a, b) => a.phone - b.phone);
}

export function getCountryCodeOption(country) {
  return (
    <Option key={country.code} value={country.code} label={`+${country.phone}`} text={`${country.name}, ${country.code}, ${country.phone}`} >
      <div style={{display: "flex", justifyContent: "space-between", marginRight: "10px"}}>
        <div>
          {country.code === "All" ? null : getCountryImage(country)}
          {`${country.name}`}
        </div>
        {country.code === "All" ? null : `+${country.phone}`}
      </div>
    </Option>
  );
}

export function getCountryImage(country) {
  return <img src={`${StaticBaseUrl}/flag-icons/${country.code}.svg`} alt={country.name} height={20} style={{marginRight: 10}} />;
}

export function initServerUrl() {
  // const hostname = window.location.hostname;
  // if (hostname === "localhost") {
  //   ServerUrl = `http://${hostname}:8000`;
  // }
}

export function isLocalhost() {
  const hostname = window.location.hostname;
  return hostname === "localhost";
}

export function getFullServerUrl() {
  let fullServerUrl = window.location.origin;
  if (fullServerUrl === "http://localhost:7001") {
    fullServerUrl = "http://localhost:8000";
  }
  return fullServerUrl;
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
    return phoneNumber.isValidPhoneNumber(phone, countryCode);
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
  if (signupItems?.length > 0) {
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

export function isRequiredEnableMfa(user, organization) {
  if (!user || !organization || !organization.mfaItems) {
    return false;
  }
  return getMfaItemsByRules(user, organization, [MfaRuleRequired]).length > 0;
}

export function getMfaItemsByRules(user, organization, mfaRules = []) {
  if (!user || !organization || !organization.mfaItems) {
    return [];
  }

  return organization.mfaItems.filter((mfaItem) => mfaRules.includes(mfaItem.rule))
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

export function goToLinkSoft(ths, link) {
  if (link.startsWith("http")) {
    openLink(link);
    return;
  }

  ths.props.history.push(link);
}

export function goToLinkSoftOrJumpSelf(ths, link) {
  if (link.startsWith("http")) {
    goToLink(link);
    return;
  }

  ths.props.history.push(link);
}

export function showMessage(type, text) {
  if (type === "success") {
    message.success(text);
  } else if (type === "error") {
    message.error(text);
  } else if (type === "info") {
    message.info(text);
  }
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

export function isMobile() {
  // return getIsMobileView();
  return isMobileDevice;
}

export function getFormattedDate(date) {
  if (!date) {
    return null;
  }

  const m = moment(date).local();
  return m.format("YYYY-MM-DD HH:mm:ss");
}

export function getFormattedDateShort(date) {
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

  return (
    <React.Fragment>
      &nbsp;
      {s}
      &nbsp;
      &nbsp;
    </React.Fragment>
  );
}

export function getShortText(s, maxLength = 35) {
  if (s.length > maxLength) {
    return `${s.slice(0, maxLength)}...`;
  } else {
    return s;
  }
}

export function getFriendlyFileSize(size) {
  if (size < 1024) {
    return size + " B";
  }

  const i = Math.floor(Math.log(size) / Math.log(1024));
  let num = (size / Math.pow(1024, i));
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

export function getLanguageText(text) {
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

export function getAcceptLanguage() {
  if (i18next.language === null || i18next.language === "") {
    return "en;q=0.9,en;q=0.8";
  }
  return i18next.language + ";q=0.9,en;q=0.8";
}

export function getClickable(text) {
  return (
    <a onClick={() => {
      copy(text);
      showMessage("success", i18next.t("general:Copied to clipboard successfully"));
    }}>
      {text}
    </a>
  );
}

export function getProviderLogoURL(provider) {
  if (provider.type === "Custom" && provider.customLogo) {
    return provider.customLogo;
  }
  if (provider.category === "OAuth") {
    return `${StaticBaseUrl}/img/social_${provider.type.toLowerCase()}.png`;
  } else {
    const info = OtherProviderInfo[provider.category][provider.type];
    // avoid crash when provider is not found
    if (info) {
      return info.logo;
    }
    return "";
  }
}

export function getProviderLogo(provider) {
  const idp = provider.type.toLowerCase().trim().split(" ")[0];
  const url = getProviderLogoURL(provider);
  return (
    <img width={30} height={30} src={url} alt={idp} />
  );
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
      ]
    );
  } else if (category === "SMS") {
    return (
      [
        {id: "Aliyun SMS", name: "Alibaba Cloud SMS"},
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
    ]);
  } else if (category === "Face ID") {
    return ([
      {id: "Alibaba Cloud Facebody", name: "Alibaba Cloud Facebody"},
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

export function renderLogo(application) {
  if (application === null) {
    return null;
  }

  if (application.homepageUrl !== "") {
    return (
      <>
        <a target="_blank" rel="noreferrer" href={application.homepageUrl}>
          <img className="panel-logo" width={60} src={application.logo} alt={application.displayName} />
        </a>
        <div style={{
          marginBottom: "55px",
          fontSize: "16px",
          fontWeight: 600,
        }}>
          {i18next.t("login:Log in to Zhuge Shenma")}
        </div>
      </>
    );
  } else {
    return (
      <img className="panel-logo" width={60} src={application.logo} alt={application.displayName} />
    );
  }
}

function isSigninMethodEnabled(application, signinMethod) {
  if (application && application.signinMethods) {
    return application.signinMethods.filter(item => item.name === signinMethod && item.rule !== "Hide password").length > 0;
  } else {
    return false;
  }
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

export function redirectToLoginPage(application, history) {
  const loginLink = getLoginLink(application);
  if (loginLink.startsWith("http://") || loginLink.startsWith("https://")) {
    goToLink(loginLink);
  } else {
    history.push(loginLink);
  }
}

function renderLink(url, text, onClick) {
  if (url === null) {
    return null;
  }

  if (url.startsWith("/")) {
    return (
      <Link style={{float: "right"}} to={url} onClick={() => {
        if (onClick !== null) {
          onClick();
        }
      }}>{text}</Link>
    );
  } else if (url.startsWith("http")) {
    return (
      <a style={{float: "right"}} href={url} onClick={() => {
        if (onClick !== null) {
          onClick();
        }
      }}>{text}</a>
    );
  } else {
    return null;
  }
}

export function renderSignupLink(application, text) {
  let url;
  if (application === null) {
    url = null;
  } else if (window.location.pathname.includes("/login/oauth/authorize")) {
    url = window.location.pathname.replace("/login/oauth/authorize", "/signup/oauth/authorize");
  } else if (authConfig.appName === application.name) {
    url = "/signup";
  } else {
    if (application.signupUrl === "") {
      url = `/signup/${application.name}`;
    } else {
      url = application.signupUrl;
    }
  }

  const storeSigninUrl = () => {
    sessionStorage.setItem("signinUrl", window.location.pathname + window.location.search);
  };

  return renderLink(url + window.location.search, text, storeSigninUrl);
}

export function renderForgetLink(application, text) {
  let url;
  if (application === null) {
    url = null;
  } else if (authConfig.appName === application.name) {
    url = "/forget";
  } else {
    if (application.forgetUrl === "") {
      url = `/forget/${application.name}`;
    } else {
      url = application.forgetUrl;
    }
  }

  const storeSigninUrl = () => {
    sessionStorage.setItem("signinUrl", window.location.pathname + window.location.search);
  };

  return renderLink(url, text, storeSigninUrl);
}

export function renderHelmet(application) {
  if (application === undefined || application === null || application.organizationObj === undefined || application.organizationObj === null || application.organizationObj === "") {
    return null;
  }

  return (
    <Helmet>
      <title>{application.organizationObj.displayName}</title>
      <link rel="icon" href={application.organizationObj.favicon} />
    </Helmet>
  );
}

export function getLabel(text, tooltip) {
  return (
    <React.Fragment>
      <span style={{marginRight: 4}}>{text}</span>
      <Tooltip placement="top" title={tooltip}>
        <QuestionCircleTwoTone twoToneColor="rgb(45,120,213)" />
      </Tooltip>
    </React.Fragment>
  );
}

export function getItem(label, key, icon, children, type) {
  return {label: label, key: key, icon: icon, children: children, type: type};
}

export function getOption(label, value) {
  return {
    label,
    value,
  };
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

export function getTagColor(s) {
  return "processing";
}

export function getTags(tags, urlPrefix = null) {
  const res = [];
  if (!tags) {
    return res;
  }

  tags.forEach((tag, i) => {
    if (urlPrefix === null) {
      res.push(
        <Tag color={getTagColor(tag)}>
          {tag}
        </Tag>
      );
    } else {
      res.push(
        <Link to={`/${urlPrefix}/${tag}`}>
          <Tag color={getTagColor(tag)}>
            {tag}
          </Tag>
        </Link>
      );
    }
  });
  return res;
}

export function getTag(color, text, icon) {
  return (
    <Tag color={color} icon={icon}>
      {text}
    </Tag>
  );
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

export function getCurrencySymbol(currency) {
  if (currency === "USD" || currency === "usd") {
    return "$";
  } else if (currency === "CNY" || currency === "cny") {
    return "¥";
  } else {
    return currency;
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

export function getUserCommonFields() {
  return ["Owner", "Name", "CreatedTime", "UpdatedTime", "DeletedTime", "Id", "Type", "Password", "PasswordSalt", "DisplayName", "FirstName", "LastName", "Avatar", "PermanentAvatar",
    "Email", "EmailVerified", "Phone", "Location", "Address", "Affiliation", "Title", "IdCardType", "IdCard", "Homepage", "Bio", "Tag", "Region",
    "Language", "Gender", "Birthday", "Education", "Score", "Ranking", "IsDefaultAvatar", "IsOnline", "IsAdmin", "IsForbidden", "IsDeleted", "CreatedIp",
    "PreferredMfaType", "TotpSecret", "SignupApplication", "RecoveryCodes", "MfaPhoneEnabled", "MfaEmailEnabled"];
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

export function getCurrencyText(product) {
  if (product?.currency === "USD") {
    return i18next.t("currency:USD");
  } else if (product?.currency === "CNY") {
    return i18next.t("currency:CNY");
  } else if (product?.currency === "EUR") {
    return i18next.t("currency:EUR");
  } else if (product?.currency === "JPY") {
    return i18next.t("currency:JPY");
  } else if (product?.currency === "GBP") {
    return i18next.t("currency:GBP");
  } else if (product?.currency === "AUD") {
    return i18next.t("currency:AUD");
  } else if (product?.currency === "CAD") {
    return i18next.t("currency:CAD");
  } else if (product?.currency === "CHF") {
    return i18next.t("currency:CHF");
  } else if (product?.currency === "HKD") {
    return i18next.t("currency:HKD");
  } else if (product?.currency === "SGD") {
    return i18next.t("currency:SGD");
  } else if (product?.currency === "BRL") {
    return i18next.t("currency:BRL");
  } else {
    return "(Unknown currency)";
  }
}

export function isDarkTheme(themeAlgorithm) {
  return themeAlgorithm && themeAlgorithm.includes("dark");
}

function getPreferredMfaProp(mfaProps) {
  for (const i in mfaProps) {
    if (mfaProps[i].isPreferred) {
      return mfaProps[i];
    }
  }
  return mfaProps[0];
}

export function checkLoginMfa(res, body, params, handleLogin, componentThis, requireRedirect = null) {
  if (res.data === RequiredMfa) {
    if (!requireRedirect) {
      componentThis.props.onLoginSuccess(window.location.href);
    } else {
      componentThis.props.onLoginSuccess(requireRedirect);
    }
  } else if (res.data === NextMfa) {
    componentThis.setState({
      mfaProps: res.data2,
      selectedMfaProp: getPreferredMfaProp(res.data2),
    }, () => {
      body["providerBack"] = body["provider"];
      body["provider"] = "";
      componentThis.setState({
        getVerifyTotp: () => renderMfaAuthVerifyForm(body, params, handleLogin, componentThis),
      });
    });
  } else if (res.data === "SelectPlan") {
    // paid-user does not have active or pending subscription, go to application default pricing page to select-plan
    const pricing = res.data2;
    goToLink(`/select-plan/${pricing.owner}/${pricing.name}?user=${body.username}`);
  } else if (res.data === "BuyPlanResult") {
    // paid-user has pending subscription, go to buy-plan/result apge to notify payment result
    const sub = res.data2;
    goToLink(`/buy-plan/${sub.owner}/${sub.pricing}/result?subscription=${sub.name}`);
  } else {
    handleLogin(res);
  }
}

export function getApplicationObj(componentThis) {
  return componentThis.props.application;
}

export function parseOffset(offset) {
  if (offset === 2 || offset === 4 || inIframe() || isMobile()) {
    return "0 auto";
  }
  if (offset === 1) {
    return "0 10%";
  }
  if (offset === 3) {
    return "0 60%";
  }
}

function renderMfaAuthVerifyForm(values, authParams, onSuccess, componentThis) {
  return (
    <div>
      <MfaAuthVerifyForm
        mfaProps={componentThis.state.selectedMfaProp}
        formValues={values}
        authParams={authParams}
        application={getApplicationObj(componentThis)}
        onFail={(errorMessage) => {
          showMessage("error", errorMessage);
        }}
        onSuccess={(res) => onSuccess(res)}
      />
      <div>
        {
          componentThis.state.mfaProps.map((mfa) => {
            if (componentThis.state.selectedMfaProp.mfaType === mfa.mfaType) {return null;}
            let mfaI18n = "";
            switch (mfa.mfaType) {
            case SmsMfaType: mfaI18n = i18next.t("mfa:Use SMS"); break;
            case TotpMfaType: mfaI18n = i18next.t("mfa:Use Authenticator App"); break ;
            case EmailMfaType: mfaI18n = i18next.t("mfa:Use Email") ;break;
            }
            return <div key={mfa.mfaType}><Button type={"link"} onClick={() => {
              componentThis.setState({
                selectedMfaProp: mfa,
              });
            }}>{mfaI18n}</Button></div>;
          })
        }
      </div>
    </div>);
}

export function renderLoginPanel(application, getInnerComponent, componentThis) {
  return (
    <div className="login-content" style={{margin: componentThis.props.preview ?? parseOffset(application.formOffset)}}>
      {inIframe() || isMobile() ? null : <div dangerouslySetInnerHTML={{__html: application.formCss}} />}
      {inIframe() || !isMobile() ? null : <div dangerouslySetInnerHTML={{__html: application.formCssMobile}} />}
      <div className={isDarkTheme(componentThis.props.themeAlgorithm) ? "login-panel-dark" : "login-panel"}>
        <div className="side-image" style={{display: application.formOffset !== 4 ? "none" : null}}>
          <div dangerouslySetInnerHTML={{__html: application.formSideHtml}} />
        </div>
        <div className="login-form">
          <div>
            {
              getInnerComponent()
            }
          </div>
        </div>
      </div>
    </div>
  );
}
