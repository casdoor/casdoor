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

import {Form} from "antd";
import i18next from "i18next";
import * as MfaBackend from "../../backend/MfaBackend";
import * as Setting from "../../Setting";
import React from "react";
import {EmailMfaType, SmsMfaType, TotpMfaType} from "../MfaSetupPage";
import MfaVerifySmsForm from "./MfaVerifySmsForm";
import MfaVerifyTotpForm from "./MfaVerifyTotpForm";

export const mfaAuth = "mfaAuth";
export const mfaSetup = "mfaSetup";

export function MfaVerifyForm({mfaProps, application, user, onSuccess, onFail}) {
  const [form] = Form.useForm();
  const onFinish = ({passcode}) => {
    const data = {passcode, mfaType: mfaProps.mfaType, ...user};
    MfaBackend.MfaSetupVerify(data)
      .then((res) => {
        if (res.status === "ok") {
          onSuccess(res);
        } else {
          onFail(res);
        }
      })
      .catch((error) => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      })
      .finally(() => {
        form.setFieldsValue({passcode: ""});
      });
  };

  if (mfaProps === undefined || mfaProps === null || application === undefined || application === null || user === undefined || user === null) {
    return <div></div>;
  }

  if (mfaProps.mfaType === SmsMfaType || mfaProps.mfaType === EmailMfaType) {
    return <MfaVerifySmsForm mfaProps={mfaProps} onFinish={onFinish} application={application} method={mfaSetup} user={user} />;
  } else if (mfaProps.mfaType === TotpMfaType) {
    return <MfaVerifyTotpForm mfaProps={mfaProps} onFinish={onFinish} />;
  } else {
    return <div></div>;
  }
}
