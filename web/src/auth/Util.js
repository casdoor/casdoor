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
import {Alert, Button, Result, message} from "antd";

export function showMessage(type, text) {
  if (type === "success") {
    message.success(text);
  } else if (type === "error") {
    message.error(text);
  }
}

export function renderMessage(msg) {
  if (msg !== null) {
    return (
      <div style={{display: "inline"}}>
        <Alert
          message="Login Error"
          showIcon
          description={msg}
          type="error"
          action={
            <Button size="small" danger>
              Detail
            </Button>
          }
        />
      </div>
    );
  } else {
    return null;
  }
}

export function renderMessageLarge(ths, msg) {
  if (msg !== null) {
    return (
      <div style={{display: "inline"}}>
        <Result
          status="error"
          title="There was a problem signing you in.."
          subTitle={msg}
          extra={[
            <Button type="primary" key="back" onClick={() => {
              window.history.go(-2);
            }}>
              Back
            </Button>,
            // <Button key="home" onClick={() => Setting.goToLinkSoft(ths, "/")}>
            //   Home
            // </Button>,
            // <Button type="primary" key="signup" onClick={() => Setting.goToLinkSoft(ths, "/signup")}>
            //   Sign Up
            // </Button>,
          ]}
        >
        </Result>
      </div>
    );
  } else {
    return null;
  }
}

function getRefinedValue(value) {
  return (value === null)? "" : value;
}

export function getCasParameters(params) {
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

export function getOAuthGetParameters(params) {
  const queries = (params !== undefined) ? params : new URLSearchParams(window.location.search);
  const clientId = getRefinedValue(queries.get("client_id"));
  const responseType = getRefinedValue(queries.get("response_type"));
  const redirectUri = getRefinedValue(queries.get("redirect_uri"));
  const scope = getRefinedValue(queries.get("scope"));
  const state = getRefinedValue(queries.get("state"));
  const nonce = getRefinedValue(queries.get("nonce"));
  const challengeMethod = getRefinedValue(queries.get("code_challenge_method"));
  const codeChallenge = getRefinedValue(queries.get("code_challenge"));
  const samlRequest = getRefinedValue(queries.get("SAMLRequest"));
  const relayState = getRefinedValue(queries.get("RelayState"));
  const noRedirect = getRefinedValue(queries.get("noRedirect"));

  if ((clientId === undefined || clientId === null || clientId === "") && (samlRequest === "" || samlRequest === undefined)) {
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
      samlRequest: samlRequest,
      relayState: relayState,
      noRedirect: noRedirect,
    };
  }
}

export function getQueryParamsToSessionStorage(applicationName, providerName, method) {
  let query = window.location.search;
  query = `${query}&application=${applicationName}&provider=${providerName}&method=${method}`;
  if (method === "link") {
    query = `${query}&from=${window.location.pathname}`;
  }
  sessionStorage.setItem("query", query);
}

export function getGetQueryParamsFromSessionStorage() {
  return sessionStorage.getItem("query");
}
