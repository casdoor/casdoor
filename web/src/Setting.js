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

import {message, Tag, Tooltip} from "antd";
import {QuestionCircleTwoTone} from "@ant-design/icons";
import React from "react";
import {isMobile as isMobileDevice} from "react-device-detect";
import "./i18n";
import i18next from "i18next";
import copy from "copy-to-clipboard";
import {authConfig} from "./auth/Auth";
import {Helmet} from "react-helmet";
import moment from "moment";

export let ServerUrl = "";

// export const StaticBaseUrl = "https://cdn.jsdelivr.net/gh/casbin/static";
export const StaticBaseUrl = "https://cdn.casbin.org";

// https://catamphetamine.gitlab.io/country-flag-icons/3x2/index.html
export const CountryRegionData = [{name: "Ascension Island", code: "AC"},{name: "Andorra", code: "AD"},{name: "United Arab Emirates", code: "AE"},{name: "Afghanistan", code: "AF"},{name: "Antigua and Barbuda", code: "AG"},{name: "Anguilla", code: "AI"},{name: "Albania", code: "AL"},{name: "Armenia", code: "AM"},{name: "Angola", code: "AO"},{name: "Antarctica", code: "AQ"},{name: "Argentina", code: "AR"},{name: "American Samoa", code: "AS"},{name: "Austria", code: "AT"},{name: "Australia", code: "AU"},{name: "Aruba", code: "AW"},{name: "Åland Islands", code: "AX"},{name: "Azerbaijan", code: "AZ"},{name: "Bosnia and Herzegovina", code: "BA"},{name: "Barbados", code: "BB"},{name: "Bangladesh", code: "BD"},{name: "Belgium", code: "BE"},{name: "Burkina Faso", code: "BF"},{name: "Bulgaria", code: "BG"},{name: "Bahrain", code: "BH"},{name: "Burundi", code: "BI"},{name: "Benin", code: "BJ"},{name: "Saint Barthélemy", code: "BL"},{name: "Bermuda", code: "BM"},{name: "Brunei Darussalam", code: "BN"},{name: "Bolivia", code: "BO"},{name: "Bonaire, Sint Eustatius and Saba", code: "BQ"},{name: "Brazil", code: "BR"},{name: "Bahamas", code: "BS"},{name: "Bhutan", code: "BT"},{name: "Bouvet Island", code: "BV"},{name: "Botswana", code: "BW"},{name: "Belarus", code: "BY"},{name: "Belize", code: "BZ"},{name: "Canada", code: "CA"},{name: "Cocos (Keeling) Islands", code: "CC"},{name: "Congo, Democratic Republic of the", code: "CD"},{name: "Central African Republic", code: "CF"},{name: "Congo", code: "CG"},{name: "Switzerland", code: "CH"},{name: "Cote d'Ivoire", code: "CI"},{name: "Cook Islands", code: "CK"},{name: "Chile", code: "CL"},{name: "Cameroon", code: "CM"},{name: "China", code: "CN"},{name: "Colombia", code: "CO"},{name: "Costa Rica", code: "CR"},{name: "Cuba", code: "CU"},{name: "Cape Verde", code: "CV"},{name: "Curaçao", code: "CW"},{name: "Christmas Island", code: "CX"},{name: "Cyprus", code: "CY"},{name: "Czech Republic", code: "CZ"},{name: "Germany", code: "DE"},{name: "Djibouti", code: "DJ"},{name: "Denmark", code: "DK"},{name: "Dominica", code: "DM"},{name: "Dominican Republic", code: "DO"},{name: "Algeria", code: "DZ"},{name: "Ecuador", code: "EC"},{name: "Estonia", code: "EE"},{name: "Egypt", code: "EG"},{name: "Western Sahara", code: "EH"},{name: "Eritrea", code: "ER"},{name: "Spain", code: "ES"},{name: "Ethiopia", code: "ET"},{name: "Finland", code: "FI"},{name: "Fiji", code: "FJ"},{name: "Falkland Islands", code: "FK"},{name: "Federated States of Micronesia", code: "FM"},{name: "Faroe Islands", code: "FO"},{name: "France", code: "FR"},{name: "Gabon", code: "GA"},{name: "United Kingdom", code: "GB"},{name: "Grenada", code: "GD"},{name: "Georgia", code: "GE"},{name: "French Guiana", code: "GF"},{name: "Guernsey", code: "GG"},{name: "Ghana", code: "GH"},{name: "Gibraltar", code: "GI"},{name: "Greenland", code: "GL"},{name: "Gambia", code: "GM"},{name: "Guinea", code: "GN"},{name: "Guadeloupe", code: "GP"},{name: "Equatorial Guinea", code: "GQ"},{name: "Greece", code: "GR"},{name: "South Georgia and the South Sandwich Islands", code: "GS"},{name: "Guatemala", code: "GT"},{name: "Guam", code: "GU"},{name: "Guinea-Bissau", code: "GW"},{name: "Guyana", code: "GY"},{name: "Hong Kong", code: "HK"},{name: "Heard Island and McDonald Islands", code: "HM"},{name: "Honduras", code: "HN"},{name: "Croatia", code: "HR"},{name: "Haiti", code: "HT"},{name: "Hungary", code: "HU"},{name: "Indonesia", code: "ID"},{name: "Ireland", code: "IE"},{name: "Israel", code: "IL"},{name: "Isle of Man", code: "IM"},{name: "India", code: "IN"},{name: "British Indian Ocean Territory", code: "IO"},{name: "Iraq", code: "IQ"},{name: "Iran", code: "IR"},{name: "Iceland", code: "IS"},{name: "Italy", code: "IT"},{name: "Jersey", code: "JE"},{name: "Jamaica", code: "JM"},{name: "Jordan", code: "JO"},{name: "Japan", code: "JP"},{name: "Kenya", code: "KE"},{name: "Kyrgyzstan", code: "KG"},{name: "Cambodia", code: "KH"},{name: "Kiribati", code: "KI"},{name: "Comoros", code: "KM"},{name: "Saint Kitts and Nevis", code: "KN"},{name: "North Korea", code: "KP"},{name: "South Korea", code: "KR"},{name: "Kuwait", code: "KW"},{name: "Cayman Islands", code: "KY"},{name: "Kazakhstan", code: "KZ"},{name: "Laos", code: "LA"},{name: "Lebanon", code: "LB"},{name: "Saint Lucia", code: "LC"},{name: "Liechtenstein", code: "LI"},{name: "Sri Lanka", code: "LK"},{name: "Liberia", code: "LR"},{name: "Lesotho", code: "LS"},{name: "Lithuania", code: "LT"},{name: "Luxembourg", code: "LU"},{name: "Latvia", code: "LV"},{name: "Libya", code: "LY"},{name: "Morocco", code: "MA"},{name: "Monaco", code: "MC"},{name: "Moldova", code: "MD"},{name: "Montenegro", code: "ME"},{name: "Saint Martin (French Part)", code: "MF"},{name: "Madagascar", code: "MG"},{name: "Marshall Islands", code: "MH"},{name: "North Macedonia", code: "MK"},{name: "Mali", code: "ML"},{name: "Burma", code: "MM"},{name: "Mongolia", code: "MN"},{name: "Macao", code: "MO"},{name: "Northern Mariana Islands", code: "MP"},{name: "Martinique", code: "MQ"},{name: "Mauritania", code: "MR"},{name: "Montserrat", code: "MS"},{name: "Malta", code: "MT"},{name: "Mauritius", code: "MU"},{name: "Maldives", code: "MV"},{name: "Malawi", code: "MW"},{name: "Mexico", code: "MX"},{name: "Malaysia", code: "MY"},{name: "Mozambique", code: "MZ"},{name: "Namibia", code: "NA"},{name: "New Caledonia", code: "NC"},{name: "Niger", code: "NE"},{name: "Norfolk Island", code: "NF"},{name: "Nigeria", code: "NG"},{name: "Nicaragua", code: "NI"},{name: "Netherlands", code: "NL"},{name: "Norway", code: "NO"},{name: "Nepal", code: "NP"},{name: "Nauru", code: "NR"},{name: "Niue", code: "NU"},{name: "New Zealand", code: "NZ"},{name: "Oman", code: "OM"},{name: "Panama", code: "PA"},{name: "Peru", code: "PE"},{name: "French Polynesia", code: "PF"},{name: "Papua New Guinea", code: "PG"},{name: "Philippines", code: "PH"},{name: "Pakistan", code: "PK"},{name: "Poland", code: "PL"},{name: "Saint Pierre and Miquelon", code: "PM"},{name: "Pitcairn", code: "PN"},{name: "Puerto Rico", code: "PR"},{name: "Palestine", code: "PS"},{name: "Portugal", code: "PT"},{name: "Palau", code: "PW"},{name: "Paraguay", code: "PY"},{name: "Qatar", code: "QA"},{name: "Reunion", code: "RE"},{name: "Romania", code: "RO"},{name: "Serbia", code: "RS"},{name: "Russia", code: "RU"},{name: "Rwanda", code: "RW"},{name: "Saudi Arabia", code: "SA"},{name: "Solomon Islands", code: "SB"},{name: "Seychelles", code: "SC"},{name: "Sudan", code: "SD"},{name: "Sweden", code: "SE"},{name: "Singapore", code: "SG"},{name: "Saint Helena", code: "SH"},{name: "Slovenia", code: "SI"},{name: "Svalbard and Jan Mayen", code: "SJ"},{name: "Slovakia", code: "SK"},{name: "Sierra Leone", code: "SL"},{name: "San Marino", code: "SM"},{name: "Senegal", code: "SN"},{name: "Somalia", code: "SO"},{name: "Suriname", code: "SR"},{name: "South Sudan", code: "SS"},{name: "Sao Tome and Principe", code: "ST"},{name: "El Salvador", code: "SV"},{name: "Sint Maarten", code: "SX"},{name: "Syria", code: "SY"},{name: "Swaziland", code: "SZ"},{name: "Tristan da Cunha", code: "TA"},{name: "Turks and Caicos Islands", code: "TC"},{name: "Chad", code: "TD"},{name: "French Southern Territories", code: "TF"},{name: "Togo", code: "TG"},{name: "Thailand", code: "TH"},{name: "Tajikistan", code: "TJ"},{name: "Tokelau", code: "TK"},{name: "Timor-Leste", code: "TL"},{name: "Turkmenistan", code: "TM"},{name: "Tunisia", code: "TN"},{name: "Tonga", code: "TO"},{name: "Turkey", code: "TR"},{name: "Trinidad and Tobago", code: "TT"},{name: "Tuvalu", code: "TV"},{name: "Taiwan", code: "TW"},{name: "Tanzania", code: "TZ"},{name: "Ukraine", code: "UA"},{name: "Uganda", code: "UG"},{name: "United States", code: "US"},{name: "Uruguay", code: "UY"},{name: "Uzbekistan", code: "UZ"},{name: "Holy See (Vatican City State)", code: "VA"},{name: "Saint Vincent and the Grenadines", code: "VC"},{name: "Venezuela", code: "VE"},{name: "Virgin Islands, British", code: "VG"},{name: "Virgin Islands, U.S.", code: "VI"},{name: "Vietnam", code: "VN"},{name: "Vanuatu", code: "VU"},{name: "Wallis and Futuna", code: "WF"},{name: "Samoa", code: "WS"},{name: "Kosovo", code: "XK"},{name: "Yemen", code: "YE"},{name: "Mayotte", code: "YT"},{name: "South Africa", code: "ZA"},{name: "Zambia", code: "ZM"},{name: "Zimbabwe", code: "ZW"}];

export function initServerUrl() {
  const hostname = window.location.hostname;
  if (hostname === "localhost") {
    ServerUrl = `http://${hostname}:8000`;
  }
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

export function isValidEmail(email) {
  // https://github.com/yiminghe/async-validator/blob/057b0b047f88fac65457bae691d6cb7c6fe48ce1/src/rule/type.ts#L9
  const emailRegex = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
  return emailRegex.test(email);
}

export function isValidPhone(phone) {
  // https://learnku.com/articles/31543, `^s*$` filter empty email individually.
  const phoneRegex = /^\s*$|^1(3\d|4[5-9]|5[0-35-9]|6[2567]|7[0-8]|8\d|9[0-35-9])\d{8}$/;
  return phoneRegex.test(phone);
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

export function getFriendlyFileSize(size) {
  if (size < 1024) {
    return size + ' B';
  }

  let i = Math.floor(Math.log(size) / Math.log(1024));
  let num = (size / Math.pow(1024, i));
  let round = Math.round(num);
  num = round < 10 ? num.toFixed(2) : round < 100 ? num.toFixed(1) : round;
  return `${num} ${'KMGTPEZY'[i-1]}B`;
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

export function setLanguage(language) {
  localStorage.setItem("language", language);
  changeMomentLanguage(language);
  i18next.changeLanguage(language);
}

export function changeLanguage(language) {
  localStorage.setItem("language", language);
  changeMomentLanguage(language);
  i18next.changeLanguage(language);
  window.location.reload(true);
}

export function changeMomentLanguage(language) {
  return;
  if (language === "zh") {
    moment.locale("zh", {
      relativeTime: {
        future: "%s内",
        past: "%s前",
        s: "几秒",
        ss: "%d秒",
        m: "1分钟",
        mm: "%d分钟",
        h: "1小时",
        hh: "%d小时",
        d: "1天",
        dd: "%d天",
        M: "1个月",
        MM: "%d个月",
        y: "1年",
        yy: "%d年",
      },
    });
  }
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
  const res = array.filter(item => filterArray.filter(filterItem => filterItem[key] === item[key]).length === 0);
  return res;
}

export function getTagColor(s) {
  return "success";
}

export function getTags(tags) {
  let res = [];
  tags.forEach((tag, i) => {
    res.push(
      <Tag color={getTagColor(tag)}>
        {tag}
      </Tag>
    );
  });
  return res;
}
