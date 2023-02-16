// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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

import {ExclamationCircleFilled} from "@ant-design/icons";
import {Modal} from "antd";
import i18next from "i18next";
import * as Conf from "../Conf";

const {confirm} = Modal;
const {fetch: originalFetch} = window;

/**
 * When modify data, prompt it's read-only and ask whether to go writable site
 */
const demoModePrompt = async(url, option) => {
  if (option.method === "POST") {
    confirm({
      title: i18next.t("general:This is a read-only demo site!"),
      icon: <ExclamationCircleFilled />,
      content: i18next.t("general:Go Writable demo site?"),
      okText: i18next.t("user:OK"),
      cancelText: i18next.t("general:Cancel"),
      onOk() {
        const fullURL = document.location.toString();
        window.open("https://demo.casdoor.com" + fullURL.substring(fullURL.lastIndexOf("/")) + "?username=built-in/admin&password=123", "_blank");
      },
      onCancel() {},
    });
  }
  return option;
};

const requsetInterceptors = [];
const responseInterceptors = [];

// when it's in DemoMode, demoModePrompt() should run before fetch
if (Conf.IsDemoMode) {
  requsetInterceptors.push(demoModePrompt);
}

/**
 * rewrite fetch to support interceptors
 */
window.fetch = async(url, option = {}) => {
  for (const fn of requsetInterceptors) {
    fn(url, option);
  }

  const response = await originalFetch(url, option);
  responseInterceptors.forEach(fn => (response) => fn(response));
  return response;
};
