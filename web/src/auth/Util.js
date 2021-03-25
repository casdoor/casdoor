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
import {Alert, Button, message, Result} from "antd";

export function goToLink(link) {
  window.location.href = link;
}

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
    )
  } else {
    return null;
  }
}

export function renderMessageLarge(msg) {
  if (msg !== null) {
    return (
      <div style={{display: "inline"}}>
        <Result
          status="error"
          title="Login Error"
          subTitle={msg}
          extra={[
            <Button type="primary" key="details">
              Details
            </Button>,
            <Button key="help">Help</Button>,
          ]}
        >
        </Result>
      </div>
    )
  } else {
    return null;
  }
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

export function getOAuthGetParameters(params) {
  const queries = (params !== undefined) ? params : new URLSearchParams(window.location.search);
  const clientId = queries.get("client_id");
  const responseType = queries.get("response_type");
  const redirectUri = queries.get("redirect_uri");
  const scope = queries.get("scope");
  const state = queries.get("state");
  if (clientId === undefined) {
    return null;
  } else {
    return {
      clientId: clientId,
      responseType: responseType,
      redirectUri: redirectUri,
      scope: scope,
      state: state,
    };
  }
}

export function getQueryParamsToState(applicationName, providerName, method) {
  let query = window.location.search;
  query = `${query}&application=${applicationName}&provider=${providerName}&method=${method}`;
  return btoa(query);
}

export function stateToGetQueryParams(state) {
  return atob(state);
}
