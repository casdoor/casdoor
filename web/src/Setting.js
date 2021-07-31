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

import {message, Tooltip} from "antd";
import {QuestionCircleTwoTone} from "@ant-design/icons";
import React from "react";
import {isMobile as isMobileDevice} from "react-device-detect";
import "./i18n";
import i18next from "i18next";
import copy from "copy-to-clipboard";
import {authConfig} from "./auth/Auth";
import {Helmet} from "react-helmet";

export let ServerUrl = "";

// export const StaticBaseUrl = "https://cdn.jsdelivr.net/gh/casbin/static";
export const StaticBaseUrl = "https://cdn.casbin.org";

// https://github.com/yiminghe/async-validator/blob/057b0b047f88fac65457bae691d6cb7c6fe48ce1/src/rule/type.ts#L9
export const EmailRegEx = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

// https://learnku.com/articles/31543, `^s*$` filter empty email individually.
export const PhoneRegEx = /^\s*$|^1(3\d|4[5-9]|5[0-35-9]|6[2567]|7[0-8]|8\d|9[0-35-9])\d{8}$/;

export function initServerUrl() {
  const hostname = window.location.hostname;
  if (hostname === "localhost") {
    ServerUrl = `http://${hostname}:8000`;
  }
}

function isLocalhost() {
  const hostname = window.location.hostname;
  return hostname === "localhost";
}

export function isProviderVisible(providerItem) {
  if (providerItem.provider === undefined || providerItem.provider === null) {
    return false;
  }

  if (providerItem.provider.category !== "OAuth") {
    return false;
  }

  if (providerItem.provider.type === "GitHub") {
    if (isLocalhost()) {
      return providerItem.provider.name.includes("localhost");
    } else {
      return !providerItem.provider.name.includes("localhost");
    }
  } else {
    return true;
  }
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

export function getAllPromptedProviderItems(application) {
  return application.providers.filter(providerItem => isProviderPrompted(providerItem));
}

export function getSignupItem(application, itemName) {
  const signupItems = application.signupItems?.filter(signupItem => signupItem.name === itemName);
  if (signupItems.length === 0) {
    return null;
  }
  return signupItems[0];
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
  if (providerItems.length !== 0) {
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

export function isPromptAnswered(user, application) {
  if (!isAffiliationAnswered(user, application)) {
    return false;
  }

  const providerItems = getAllPromptedProviderItems(application);
  for (let i = 0; i < providerItems.length; i ++) {
    if (!isProviderItemAnswered(user, application, providerItems[i])) {
      return false;
    }
  }
  return true;
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
  const w = window.open('about:blank');
  w.location.href = link;
}

export function goToLink(link) {
  window.location.href = link;
}

export function goToLinkSoft(ths, link) {
  ths.props.history.push(link);
}

export function showMessage(type, text) {
  if (type === "") {
    return;
  } else if (type === "success") {
    message.success(text);
  } else if (type === "error") {
    message.error(text);
  }
}

export function isAdminUser(account) {
  if (account === undefined || account === null) {
    return false;
  }
  return account.owner === "built-in" || account.isGlobalAdmin === true;
}

export function deepCopy(obj) {
  return Object.assign({}, obj);
}

export function addRow(array, row) {
  return [...array, row];
}

export function prependRow(array, row) {
  return [row, ...array];
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

  while(start < end && str[start] === ch)
    ++start;

  while(end > start && str[end - 1] === ch)
    --end;

  return (start > 0 || end < str.length) ? str.substring(start, end) : str;
}

export function isMobile() {
  // return getIsMobileView();
  return isMobileDevice;
}

export function getFormattedDate(date) {
  if (date === undefined) {
    return null;
  }

  date = date.replace('T', ' ');
  date = date.replace('+08:00', ' ');
  return date;
}

export function getFormattedDateShort(date) {
  return date.slice(0, 10);
}

export function getShortName(s) {
  return s.split('/').slice(-1)[0];
}

export function getShortText(s, maxLength=35) {
  if (s.length > maxLength) {
    return `${s.slice(0, maxLength)}...`;
  } else {
    return s;
  }
}

function getRandomInt(s) {
  let hash = 0;
  if (s.length !== 0) {
    for (let i = 0; i < s.length; i ++) {
      let char = s.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash;
    }
  }

  return hash;
}

export function getAvatarColor(s) {
  const colorList = ['#f56a00', '#7265e6', '#ffbf00', '#00a2ae'];
  let random = getRandomInt(s);
  if (random < 0) {
    random = -random;
  }
  return colorList[random % 4];
}

export function setLanguage() {
  let language = localStorage.getItem('language');
  if (language === undefined) {
    language = "en"
  }
  i18next.changeLanguage(language)
}

export function changeLanguage(language) {
  localStorage.setItem("language", language)
  i18next.changeLanguage(language)
  window.location.reload(true);
}

export function getClickable(text) {
  return (
    // eslint-disable-next-line jsx-a11y/anchor-is-valid
    <a onClick={() => {
      copy(text);
      showMessage("success", `Copied to clipboard`);
    }}>
      {text}
    </a>
  )
}

export function getProviderLogo(provider) {
  const idp = provider.type.toLowerCase();
  const url = `${StaticBaseUrl}/img/social_${idp}.png`;
  return (
    <img width={30} height={30} src={url} alt={idp} />
  )
}

export function renderLogo(application) {
  if (application === null) {
    return null;
  }

  if (application.homepageUrl !== "") {
    return (
      <a target="_blank" rel="noreferrer" href={application.homepageUrl}>
        <img width={250} src={application.logo} alt={application.displayName} style={{marginBottom: '30px'}}/>
      </a>
    )
  } else {
    return (
      <img width={250} src={application.logo} alt={application.displayName} style={{marginBottom: '30px'}}/>
    );
  }
}

export function goToLogin(ths, application) {
  if (application === null) {
    return;
  }

  if (!application.enablePassword && window.location.pathname.includes("/signup/oauth/authorize")) {
    const link = window.location.href.replace("/signup/oauth/authorize", "/login/oauth/authorize");
    goToLink(link);
    return;
  }

  if (authConfig.appName === application.name) {
    goToLinkSoft(ths, "/login");
  } else {
    if (application.signinUrl === "") {
      goToLink(`${application.homepageUrl}/login`);
    } else {
      goToLink(application.signinUrl);
    }
  }
}

export function goToSignup(ths, application) {
  if (application === null) {
    return;
  }

  if (!application.enablePassword && window.location.pathname.includes("/login/oauth/authorize")) {
    const link = window.location.href.replace("/login/oauth/authorize", "/signup/oauth/authorize");
    goToLink(link);
    return;
  }

  if (authConfig.appName === application.name) {
    goToLinkSoft(ths, "/signup");
  } else {
    if (application.signupUrl === "") {
      goToLinkSoft(ths, `/signup/${application.name}`);
    } else {
      goToLink(application.signupUrl);
    }
  }
}

export function goToForget(ths, application) {
  if (application === null) {
    return;
  }

  if (authConfig.appName === application.name) {
    goToLinkSoft(ths, "/forget");
  } else {
    if (application.forgetUrl === "") {
      goToLinkSoft(ths, `/forget/${application.name}`);
    } else {
      goToLink(application.forgetUrl);
    }
  }
}

export function renderHelmet(application) {
  if (application === undefined || application === null || application.organizationObj === undefined || application.organizationObj === null ||application.organizationObj === "") {
    return null;
  }

  return (
    <Helmet>
      <title>{application.organizationObj.displayName}</title>
      <link rel="icon" href={application.organizationObj.favicon} />
    </Helmet>
  )
}

export function getLabel(text, tooltip) {
  return (
    <React.Fragment>
      <span style={{ marginRight: 4 }}>{text}</span>
      <Tooltip placement="top" title={tooltip}>
        <QuestionCircleTwoTone twoToneColor="rgb(45,120,213)" />
      </Tooltip>
    </React.Fragment>
  );
}

function repeat(str, len) {
  while (str.length < len) {
    str += str.substr(0, len - str.length);
  }
  return str;
}

function maskString(s) {
  if (s.length <= 2) {
    return s;
  } else {
    return `${s[0]}${repeat("*", s.length - 2)}${s[s.length - 1]}`;
  }
}

export function maskEmail(email) {
  if (email === "") return;
  const tokens = email.split("@");
  let username = tokens[0];
  username = maskString(username);

  const domain = tokens[1];
  let domainTokens = domain.split(".");
  domainTokens[domainTokens.length - 2] = maskString(domainTokens[domainTokens.length - 2]);

  return `${username}@${domainTokens.join(".")}`;
}

export function getArrayItem(array, key, value) {
  const res = array.filter(item => item[key] === value)[0];
  return res;
}

export function getDeduplicatedArray(array, filterArray, key) {
  const res = array?.filter(item => filterArray?.filter(filterItem => filterItem[key] === item[key]).length === 0);
  return res;
}

export function getDuplicatedArray(array, filterArray, key) {
  const res = array?.filter(item => filterArray?.filter(filterItem => filterItem[key] === item[key]).length !== 0);
  return res;
}