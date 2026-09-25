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

// Query-string helpers for the sign-in flows (OAuth / CAS / SAML), ported
// verbatim from web/src/auth/Util.js minus its antd renderers.

import * as AuthBackend from "@/backend/AuthBackend";
import * as Provider from "@/auth/Provider";
import * as Setting from "@/lib/setting";

function getRefinedValue(value) {
  return value ?? "";
}

export function getCasParameters(params?: any) {
  const queries = (params !== undefined) ? params : new URLSearchParams(window.location.search);
  const service = getRefinedValue(queries.get("service"));
  const renew = getRefinedValue(queries.get("renew"));
  const gateway = getRefinedValue(queries.get("gateway"));
  return {
    service: service,
    renew: renew,
    gateway: gateway,
  };
}

function getRawGetParameter(key) {
  const fullUrl = window.location.href;
  const token = fullUrl.split(`${key}=`)[1];
  if (!token) {
    return "";
  }

  let res = token.split("&")[0];
  if (!res) {
    return "";
  }

  res = decodeURIComponent(res);
  return res;
}

export function getCasLoginParameters(owner, name) {
  const queries = new URLSearchParams(window.location.search);
  // CAS service
  let service = getRawGetParameter("service");
  if (service === "") {
    service = getRefinedValue(queries.get("service"));
  }
  return {
    id: `${owner}/${encodeURIComponent(name)}`, // application ID,
    service: service,
    type: "cas",
  };
}

// the SAML HTTP-Redirect and HTTP-POST bindings spell "SAMLRequest" differently
export function getParameterIgnoreCase(params: URLSearchParams, key: string): string | null {
  const target = key.toLowerCase();
  let result: string | null = null;
  params.forEach((val, name) => {
    if (result === null && name.toLowerCase() === target) {
      result = val;
    }
  });
  return result;
}

// getRelayState returns the RelayState in the URL, it is used by the SAML IdP-initiated SSO,
// where there is no SAMLRequest and getOAuthGetParameters() returns null
export function getRelayState() {
  const queries = new URLSearchParams(window.location.search);
  const lowercaseQueries = {};
  queries.forEach((val, key) => {lowercaseQueries[key.toLowerCase()] = val;});

  return getRefinedValue(lowercaseQueries["RelayState".toLowerCase()]);
}

export function getOAuthGetParameters(params?: any): any {
  const queries = (params !== undefined) ? params : new URLSearchParams(window.location.search);
  const lowercaseQueries = {};
  queries.forEach((val, key) => {lowercaseQueries[key.toLowerCase()] = val;});

  const clientId = getRefinedValue(queries.get("client_id"));
  const responseType = getRefinedValue(queries.get("response_type"));

  let redirectUri = getRawGetParameter("redirect_uri");
  if (redirectUri === "") {
    redirectUri = getRefinedValue(queries.get("redirect_uri"));
  }

  let scope = getRefinedValue(queries.get("scope"));
  if (redirectUri.includes("#") && scope === "") {
    scope = getRawGetParameter("scope");
  }

  let state = getRefinedValue(queries.get("state"));
  if (redirectUri.includes("#") && state === "") {
    state = getRawGetParameter("state");
  }

  const nonce = getRefinedValue(queries.get("nonce"));
  const challengeMethod = getRefinedValue(queries.get("code_challenge_method"));
  const codeChallenge = getRefinedValue(queries.get("code_challenge"));
  const responseMode = getRefinedValue(queries.get("response_mode"));
  const samlRequest = getRefinedValue(lowercaseQueries["samlRequest".toLowerCase()]);
  const relayState = getRefinedValue(lowercaseQueries["RelayState".toLowerCase()]);
  const noRedirect = getRefinedValue(lowercaseQueries["noRedirect".toLowerCase()]);
  const resource = getRefinedValue(queries.get("resource"));

  if (clientId === "" && samlRequest === "") {
    // login
    return null;
  } else {
    // code
    return {
      clientId: clientId,
      responseType: responseType,
      redirectUri: redirectUri,
      scope: scope,
      state: state,
      nonce: nonce,
      challengeMethod: challengeMethod,
      codeChallenge: codeChallenge,
      responseMode: responseMode,
      samlRequest: samlRequest,
      relayState: relayState,
      noRedirect: noRedirect,
      resource: resource,
      type: "code",
    };
  }
}

export function getStateFromQueryParams(applicationName, providerName, method, isShortState) {
  let query = window.location.search;
  query = `${query}&application=${encodeURIComponent(applicationName)}&provider=${encodeURIComponent(providerName)}&method=${method}`;
  if (method === "link") {
    query = `${query}&from=${window.location.pathname}&linkNonce=${newLinkNonce()}`;
  }

  // Device authorization flow: the userCode lives in the path (/login/oauth/device/:userCode),
  // not in the query string, so carry it through the social login round-trip via the state.
  const deviceMatch = window.location.pathname.match(/\/login\/oauth\/device\/([^/?]+)/);
  if (deviceMatch) {
    query = `${query}&userCode=${encodeURIComponent(deviceMatch[1])}`;
  }

  if (!isShortState) {
    return btoa(query);
  } else {
    const state = providerName;
    sessionStorage.setItem(state, query);
    return state;
  }
}

// Keep in sync with public/AuthCallbackHandler.js
const LINK_NONCE_KEY = "casdoor_link_nonce";

function newLinkNonce(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  const nonce = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
  sessionStorage.setItem(LINK_NONCE_KEY, nonce);
  return nonce;
}

/**
 * Tells whether this tab started the account linking the callback comes back from. The
 * state is not bound to the session, so a callback URL crafted by someone else would
 * otherwise link their provider account to the signed-in user.
 */
export function consumeLinkNonce(nonce: string | null): boolean {
  const expected = sessionStorage.getItem(LINK_NONCE_KEY);
  sessionStorage.removeItem(LINK_NONCE_KEY);
  return !!nonce && nonce === expected;
}

export function getQueryParamsFromState(state) {
  const query = sessionStorage.getItem(state);
  if (query === null) {
    return atob(state);
  } else {
    return query;
  }
}

/**
 * Polls the WeChat official account for the scan event and, once the user has
 * scanned (or subscribed), continues into the normal OAuth redirect with the
 * code the backend handed back. Ported from web/src/auth/Util.js.
 */
export function getEvent(application, provider, ticket, method) {
  return AuthBackend.getWechatMessageEvent(ticket).then((res: any) => {
    if (res.data === "SCAN" || res.data === "subscribe") {
      Setting.goToLink(Provider.getAuthUrl(application, provider, method ?? "signup", res.data2));
    }
  });
}
