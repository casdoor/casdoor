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
    query = `${query}&from=${window.location.pathname}`;
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

export function getQueryParamsFromState(state) {
  const query = sessionStorage.getItem(state);
  if (query === null) {
    return atob(state);
  } else {
    return query;
  }
}
