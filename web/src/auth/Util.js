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

import {message} from "antd";

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

export function getQueryParamsToState() {
  const query = window.location.search;
  return btoa(query);
}

export function stateToGetQueryParams(state) {
  return atob(state);
}
