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

import React, {useState} from "react";
import i18next from "i18next";
import {Button, Input} from "antd";
import * as AuthBackend from "./AuthBackend";
import {SmsMfaType} from "./MfaSetupPage";
import {MfaSmsVerifyForm} from "./MfaVerifyForm";

export const NextMfa = "NextMfa";

export function MfaAuthVerifyForm({formValues, oAuthParams, mfaProps, application, onSuccess, onFail}) {
  formValues.password = "";
  formValues.username = "";
  const [loading, setLoading] = useState(false);
  const [type, setType] = useState(mfaProps.type);
  const [recoveryCode, setRecoveryCode] = useState("");

  const verify = ({passcode}) => {
    setLoading(true);
    const values = {...formValues, passcode, mfaType: type};
    AuthBackend.login(values, oAuthParams).then((res) => {
      if (res.status === "ok") {
        onSuccess(res);
      } else {
        onFail(res.msg);
      }
    }).catch((reason) => {
      onFail(reason.message);
    }).finally(() => {
      setLoading(false);
    });
  };

  const recover = () => {
    setLoading(true);
    AuthBackend.login({...formValues, recoveryCode}, oAuthParams).then(res => {
      if (res.status === "ok") {
        onSuccess();
      } else {
        onFail(res.msg);
      }
    }).catch((reason) => {
      onFail(reason.message);
    }).finally(() => {
      setLoading(false);
    });
  };

  switch (type) {
  case SmsMfaType:
    return (
      <div style={{width: 300, height: 350}}>
        <div style={{marginBottom: 24, textAlign: "center", fontSize: "24px"}}>
          {i18next.t("mfa:Two-factor authentication")}
        </div>
        <div style={{marginBottom: 24}}>
          {i18next.t("mfa:Two-factor authentication description")}
        </div>
        <MfaSmsVerifyForm
          mfaProps={mfaProps}
          onFinish={verify}
          application={application}
        />
        <span style={{float: "right"}}>
          {i18next.t("mfa:Have problems?")}
          <a onClick={() => {
            setType("recovery");
          }}>
            {i18next.t("mfa:Use a recovery code")}
          </a>
        </span>
      </div>
    );
  case "recovery":
    return (
      <div style={{width: 300, height: 350}}>
        <div style={{marginBottom: 24, textAlign: "center", fontSize: "24px"}}>
          {i18next.t("mfa:Two-factor recover")}
        </div>
        <div style={{marginBottom: 24}}>
          {i18next.t("mfa:Two-factor recover description")}
        </div>
        <Input placeholder={i18next.t("mfa:Recovery code")}
          style={{marginBottom: 24}}
          type={"passcode"}
          size={"large"}
          onChange={event => setRecoveryCode(event.target.value)}
        />
        <Button style={{width: "100%", marginBottom: 20}} size={"large"} loading={loading}
          type={"primary"} onClick={() => {
            recover();
          }}>{i18next.t("mfa:Verify")}
        </Button>
        <span style={{float: "right"}}>
          {i18next.t("mfa:Have problems?")}
          <a onClick={() => {
            setType(mfaProps.type);
          }}>
            {i18next.t("mfa:Use SMS verification code")}
          </a>
        </span>
      </div>
    );
  default:
    return null;
  }
}
