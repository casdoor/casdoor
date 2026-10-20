// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

(function() {
  "use strict";

  // the path Casdoor is served under, like "/casdoor", which the server writes into the page (routers/base_path.go)
  var basePath = (function() {
    var meta = document.querySelector("meta[name=\"casdoor-base-path\"]");
    var value = meta ? (meta.getAttribute("content") || "").replace(/^\/+|\/+$/g, "") : "";
    return value ? "/" + value : "";
  })();

  var reactFallbackKey = "__casdoor_callback_react";
  var reactFallbackPayloadKey = "casdoor_callback_react_fallback";

  function getIssuer() {
    var meta = document.querySelector("meta[name=\"casdoor-issuer\"]");
    return (meta && meta.getAttribute("content")) || window.location.origin + basePath;
  }

  function setStatus(message, isError) {
    var statusNode = document.getElementById("callback-status");
    if (!statusNode) {
      return;
    }

    statusNode.textContent = message;
    statusNode.style.color = isError ? "#b42318" : "#1f2937";
  }

  function getReactCallbackOrigin() {
    if (window.location.port === "8000" && ["localhost", "127.0.0.1"].indexOf(window.location.hostname) !== -1) {
      return window.location.protocol + "//" + window.location.hostname + ":7001";
    }

    return window.location.origin;
  }

  function goToReactFallback() {
    var url = new URL(window.location.href);
    url.protocol = window.location.protocol;
    url.host = new URL(getReactCallbackOrigin()).host;
    url.searchParams.set(reactFallbackKey, "1");
    window.location.replace(url.toString());
  }

  function storeReactFallbackPayload(payload) {
    sessionStorage.setItem(reactFallbackPayloadKey, JSON.stringify(payload));
  }

  function getQueryParamsFromState(state) {
    var query = sessionStorage.getItem(state);
    if (query === null) {
      return atob(state);
    }
    return query;
  }

  // Keep in sync with consumeLinkNonce() in src/auth/Util.ts
  function consumeLinkNonce(nonce) {
    var expected = sessionStorage.getItem("casdoor_link_nonce");
    sessionStorage.removeItem("casdoor_link_nonce");
    return !!nonce && nonce === expected;
  }

  function getParameterIgnoreCase(params, key) {
    var target = key.toLowerCase();
    var result = null;
    params.forEach(function(val, name) {
      if (result === null && name.toLowerCase() === target) {
        result = val;
      }
    });
    return result;
  }

  function getInnerParams() {
    var params = new URLSearchParams(window.location.search);
    var state = params.get("state");
    if (!state) {
      return null;
    }

    var queryString = getQueryParamsFromState(state);
    return new URLSearchParams(queryString);
  }

  function getResponseType(innerParams) {
    var method = innerParams.get("method");
    if (method === "signup" || method === "signin") {
      var realRedirectUri = innerParams.get("redirect_uri");
      if (realRedirectUri === null) {
        var samlRequest = getParameterIgnoreCase(innerParams, "SAMLRequest");
        var casService = innerParams.get("service");
        if (samlRequest) {
          return "saml";
        }
        if (casService) {
          return "cas";
        }
        return "login";
      }

      var realRedirectUrl = new URL(realRedirectUri);
      if (window.location.origin === realRedirectUrl.origin && (basePath === "" || realRedirectUrl.pathname === basePath || realRedirectUrl.pathname.indexOf(basePath + "/") === 0)) {
        return "login";
      }

      return innerParams.get("response_type") || "code";
    }

    if (method === "link") {
      return "link";
    }

    return "unknown";
  }

  function getCodeVerifier(state) {
    return localStorage.getItem("pkce_verifier_" + state);
  }

  function clearCodeVerifier(state) {
    localStorage.removeItem("pkce_verifier_" + state);
  }

  function getRefinedValue(value) {
    return value || "";
  }

  function getRawGetParameter(key, source) {
    var token = source.split(key + "=")[1];
    if (!token) {
      return "";
    }

    var result = token.split("&")[0];
    if (!result) {
      return "";
    }

    return decodeURIComponent(result);
  }

  function getOAuthGetParameters(innerParams, queryString) {
    var lowercaseQueries = {};
    innerParams.forEach(function(value, key) {
      lowercaseQueries[key.toLowerCase()] = value;
    });

    var clientId = getRefinedValue(innerParams.get("client_id"));
    var responseType = getRefinedValue(innerParams.get("response_type"));

    var redirectUri = getRawGetParameter("redirect_uri", queryString);
    if (redirectUri === "") {
      redirectUri = getRefinedValue(innerParams.get("redirect_uri"));
    }

    var scope = getRefinedValue(innerParams.get("scope"));
    if (redirectUri.indexOf("#") !== -1 && scope === "") {
      scope = getRawGetParameter("scope", queryString);
    }

    var state = getRefinedValue(innerParams.get("state"));
    if (redirectUri.indexOf("#") !== -1 && state === "") {
      state = getRawGetParameter("state", queryString);
    }

    return {
      clientId: clientId,
      responseType: responseType,
      redirectUri: redirectUri,
      scope: scope,
      state: state,
      nonce: getRefinedValue(innerParams.get("nonce")),
      challengeMethod: getRefinedValue(innerParams.get("code_challenge_method")),
      codeChallenge: getRefinedValue(innerParams.get("code_challenge")),
      responseMode: getRefinedValue(innerParams.get("response_mode")),
      relayState: getRefinedValue(lowercaseQueries["relaystate"]),
      resource: getRefinedValue(innerParams.get("resource")),
      type: "code"
    };
  }

  function oAuthParamsToQuery(oAuthParams) {
    if (!oAuthParams) {
      return "";
    }

    var resourceQuery = oAuthParams.resource
      ? "&resource=" + encodeURIComponent(oAuthParams.resource)
      : "";

    return "?clientId=" + encodeURIComponent(oAuthParams.clientId) +
      "&responseType=" + encodeURIComponent(oAuthParams.responseType) +
      "&redirectUri=" + encodeURIComponent(oAuthParams.redirectUri) +
      "&type=" + encodeURIComponent(oAuthParams.type) +
      "&scope=" + encodeURIComponent(oAuthParams.scope) +
      "&state=" + encodeURIComponent(oAuthParams.state) +
      "&nonce=" + encodeURIComponent(oAuthParams.nonce) +
      "&code_challenge_method=" + encodeURIComponent(oAuthParams.challengeMethod) +
      "&code_challenge=" + encodeURIComponent(oAuthParams.codeChallenge) +
      resourceQuery;
  }

  function createFormAndSubmit(action, params) {
    var form = document.createElement("form");
    form.method = "post";
    form.action = action;

    Object.keys(params).forEach(function(key) {
      if (params[key] === null || params[key] === undefined) {
        return;
      }
      var input = document.createElement("input");
      input.type = "hidden";
      input.name = key;
      input.value = params[key];
      form.appendChild(input);
    });

    document.body.appendChild(form);
    form.submit();
  }

  // Keep in sync with sendOAuthResponse and getTokenResponsePayload in src/auth/Util.ts
  function sendOAuthResponse(oAuthParams, payload) {
    if (oAuthParams.state) {
      payload.state = oAuthParams.state;
    }
    payload.iss = getIssuer();
    if (oAuthParams.responseMode === "form_post") {
      createFormAndSubmit(oAuthParams.redirectUri, payload);
      return;
    }

    var hasToken = "access_token" in payload || "id_token" in payload;
    var inFragment = oAuthParams.responseMode === "fragment" || hasToken ||
      (oAuthParams.responseMode !== "query" && (oAuthParams.responseType || "code") !== "code");
    var concatChar = inFragment ? "#" : oAuthParams.redirectUri.indexOf("?") !== -1 ? "&" : "?";
    var query = new URLSearchParams();
    Object.keys(payload).forEach(function(key) {
      query.append(key, payload[key]);
    });
    window.location.replace(oAuthParams.redirectUri + concatChar + query.toString());
  }

  function getTokenResponsePayload(responseTypes, res) {
    var payload = {};
    if (responseTypes.indexOf("code") !== -1) {
      payload.code = res.data;
    }
    if (responseTypes.indexOf("token") !== -1) {
      payload.access_token = responseTypes.indexOf("code") !== -1 ? res.data2 : res.data;
      payload.token_type = "Bearer";
    }
    if (responseTypes.indexOf("id_token") !== -1) {
      payload.id_token = res.data3;
    }
    return payload;
  }

  function extractCallbackCode(params) {
    var isSteam = params.get("openid.mode");
    var code = params.get("code") || params.get("auth_code") || params.get("authCode");

    if (isSteam !== null && code === null) {
      code = window.location.search;
    }

    var telegramId = params.get("id");
    if (telegramId !== null && (code === null || code === "")) {
      var telegramAuthData = {
        id: parseInt(telegramId, 10)
      };
      var hash = params.get("hash");
      var authDate = params.get("auth_date");
      if (hash) {
        telegramAuthData.hash = hash;
      }
      if (authDate) {
        telegramAuthData.auth_date = authDate;
      }
      ["first_name", "last_name", "username", "photo_url"].forEach(function(field) {
        var value = params.get(field);
        if (value !== null && value !== "") {
          telegramAuthData[field] = value;
        }
      });
      code = JSON.stringify(telegramAuthData);
    }

    return code;
  }

  // The backend answers a login that still needs the user's consent with
  // `data: {required: true}` instead of an authorization code.
  function isConsentRequired(res) {
    return !!(res.data && typeof res.data === "object" && res.data.required === true);
  }

  function goToConsentPage(applicationName, queryString) {
    var url = new URL(getReactCallbackOrigin());
    url.pathname = basePath + "/consent/" + encodeURIComponent(applicationName);
    url.search = queryString || "";
    window.location.replace(url.toString());
  }

  function shouldFallbackToReact(res) {
    return res.data === "RequiredMfa" || res.data === "NextMfa" || res.data === "SelectPlan" || res.data === "BuyPlanResult" || res.data3;
  }

  function getFromLink() {
    return basePath + (sessionStorage.getItem("from") || "/");
  }

  async function loginCas(body, casService) {
    return fetch(window.location.origin + basePath + "/api/login?service=" + encodeURIComponent(casService || ""), {
      method: "POST",
      credentials: "include",
      body: JSON.stringify(body),
      headers: {
        "Accept-Language": localStorage.getItem("language") || navigator.language || "en"
      }
    }).then(function(res) {
      return res.json();
    });
  }

  async function run() {
    setStatus("Signing in...", false);

    var params = new URLSearchParams(window.location.search);
    var innerParams = getInnerParams();
    if (!innerParams) {
      setStatus("Missing callback state.", true);
      return;
    }

    var queryString = getQueryParamsFromState(params.get("state"));
    var applicationName = innerParams.get("application");
    var providerName = innerParams.get("provider");
    var method = innerParams.get("method");
    var samlRequest = getParameterIgnoreCase(innerParams, "SAMLRequest");
    var code = extractCallbackCode(params);
    var responseType = getResponseType(innerParams);
    if (responseType === "link" && !consumeLinkNonce(innerParams.get("linkNonce"))) {
      setStatus("Unauthorized", true);
      return;
    }

    var redirectUri = window.location.origin + basePath + "/callback";
    var codeVerifier = getCodeVerifier(params.get("state"));
    var body = {
      type: responseType,
      application: applicationName,
      provider: providerName,
      code: code,
      samlRequest: samlRequest,
      state: applicationName,
      invitationCode: innerParams.get("invitationCode") || "",
      redirectUri: redirectUri,
      method: method,
      codeVerifier: codeVerifier
    };

    if (codeVerifier) {
      clearCodeVerifier(params.get("state"));
    }

    if (responseType === "cas") {
      var casService = innerParams.get("service") || "";
      var casRes = await loginCas(body, casService);
      if (casRes.status !== "ok") {
        setStatus(casRes.msg || "Failed to sign in.", true);
        return;
      }

      if (shouldFallbackToReact(casRes)) {
        storeReactFallbackPayload({
          search: window.location.search,
          body: body,
          res: casRes,
          flow: "cas",
          casService: casService
        });
        goToReactFallback();
        return;
      }

      if (casService === "") {
        setStatus("Logged in successfully. Now you can visit apps protected by Casdoor.", false);
        return;
      }

      var serviceUrl = new URL(casService);
      serviceUrl.searchParams.append("ticket", casRes.data);
      window.location.replace(serviceUrl.toString());
      return;
    }

    var oAuthParams = getOAuthGetParameters(innerParams, queryString);
    var response = await fetch(window.location.origin + basePath + "/api/login" + oAuthParamsToQuery(oAuthParams), {
      method: "POST",
      credentials: "include",
      body: JSON.stringify(body),
      headers: {
        "Accept-Language": localStorage.getItem("language") || navigator.language || "en"
      }
    });
    var res = await response.json();
    if (res.status !== "ok") {
      setStatus(res.msg || "Failed to sign in.", true);
      return;
    }

    // The consent page issues the real code, so it needs the original
    // authorization request, which is what the state carries.
    if (isConsentRequired(res)) {
      goToConsentPage(applicationName, queryString);
      return;
    }

    if (shouldFallbackToReact(res)) {
      storeReactFallbackPayload({
        search: window.location.search,
        body: body,
        res: res,
        flow: "oauth",
        responseType: responseType,
        queryString: queryString,
        innerParams: queryString,
        oAuthParams: oAuthParams
      });
      goToReactFallback();
      return;
    }

    var responseTypes = responseType.split(" ");

    if (responseType === "login") {
      window.location.replace(getFromLink());
      return;
    }

    if (responseType === "code") {
      sendOAuthResponse(oAuthParams, {code: res.data});
      return;
    }

    if (responseTypes.indexOf("token") !== -1 || responseTypes.indexOf("id_token") !== -1) {
      sendOAuthResponse(oAuthParams, getTokenResponsePayload(responseTypes, res));
      return;
    }

    if (responseType === "link") {
      var from = innerParams.get("from") || "/";
      var oauth = innerParams.get("oauth");
      if (oauth) {
        from += "?oauth=" + oauth;
      }
      window.location.replace(basePath + from);
      return;
    }

    if (responseType === "saml") {
      if (res.data2 && res.data2.method === "POST") {
        createFormAndSubmit(res.data2.redirectUrl, {
          SAMLResponse: res.data,
          RelayState: oAuthParams.relayState
        });
      } else if (res.data2) {
        var samlRedirectUri = res.data2.redirectUrl;
        window.location.replace(samlRedirectUri + (samlRedirectUri.indexOf("?") !== -1 ? "&" : "?") + "SAMLResponse=" + encodeURIComponent(res.data) + "&RelayState=" + oAuthParams.relayState);
      } else {
        setStatus("Unsupported SAML callback response.", true);
      }
      return;
    }

    goToReactFallback();
  }

  window.CasdoorAuthCallback = {
    run: function() {
      return run().catch(function(error) {
        setStatus(error && error.message ? error.message : "Failed to complete callback.", true);
      });
    }
  };
})();
