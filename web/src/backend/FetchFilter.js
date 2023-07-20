// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import {Modal} from "antd";
import {ExclamationCircleFilled} from "@ant-design/icons";
import i18next from "i18next";
import * as Conf from "../Conf";
import * as Setting from "../Setting";

const {confirm} = Modal;
const {fetch: originalFetch} = window;

const demoModeFilter = (response) => {
  if (!Conf.IsDemoMode) {
    return;
  }

  response.json().then(res => {
    if (Setting.isResponseDenied(res)) {
      confirm({
        title: i18next.t("general:This is a read-only demo site!"),
        icon: <ExclamationCircleFilled />,
        content: i18next.t("general:Go to writable demo site?"),
        okText: i18next.t("general:OK"),
        cancelText: i18next.t("general:Cancel"),
        onOk() {
          Setting.openLink(`https://demo.casdoor.com${location.pathname}${location.search}?username=built-in/admin&password=123`);
        },
        onCancel() {},
      });
    }
  });
};

const GetResponseFilter = (response) => {
  response.json().then(res => {
    if (res.status === "ok") {
      if (res.data === null) {
        window.location.href = "/404";
        return;
      }
    }
    if (res.status === "error") {
      if (res.code === 403) {
        return;
      }
      Setting.showMessage("error", res.msg);
    }
  });
};

window.fetch = async(url, option = {}) => {
  return new Promise((resolve, reject) => {
    originalFetch(url, option).then(res => {
      if (!url.startsWith("/api/get-organizations")) {
        demoModeFilter(res.clone());
      }
      if (url.startsWith("/api/get-")) {
        GetResponseFilter(res.clone());
      }
      resolve(res);
    });
  });
};
