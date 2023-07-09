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
import * as AuthBackend from "../AuthBackend";
import {EmailMfaType, RecoveryMfaType, SmsMfaType} from "../MfaSetupPage";
import {mfaAuth} from "./MfaVerifyForm";
import MfaVerifySmsForm from "./MfaVerifySmsForm";
import MfaVerifyTotpForm from "./MfaVerifyTotpForm";

export const NextMfa = "NextMfa";
export const RequiredMfa = "RequiredMfa";

export function MfaAuthVerifyForm({formValues, oAuthParams, mfaProps, application, onSuccess, onFail}) {
  formValues.password = "";
  formValues.username = "";
  const [loading, setLoading] = useState(false);
  const [mfaType, setMfaType] = useState(mfaProps.mfaType);
  const [recoveryCode, setRecoveryCode] = useState("");

  const verify = ({passcode}) => {
    setLoading(true);
    const values = {...formValues, passcode, mfaType};
    AuthBackend.login(values, oAuthParams).then((res) => {
      if (res.status === "ok") {
        onSuccess(res);
      } else {
        onFail(res.msg);
      }
    }).catch((res) => {
      onFail(res.message);
    }).finally(() => {
      setLoading(false);
    });
  };

  const recover = () => {
    setLoading(true);
    AuthBackend.login({...formValues, recoveryCode}, oAuthParams).then(res => {
      if (res.status === "ok") {
        onSuccess(res);
      } else {
        onFail(res.msg);
      }
    }).catch((res) => {
      onFail(res.message);
    }).finally(() => {
      setLoading(false);
    });
  };

  if (mfaType !== RecoveryMfaType) {
    return (
      <div style={{width: 300, height: 350}}>
        <div style={{marginBottom: 24, textAlign: "center", fontSize: "24px"}}>
          {i18next.t("mfa:Multi-factor authentication")}
        </div>
        <div style={{marginBottom: 24}}>
          {i18next.t("mfa:Multi-factor authentication description")}
        </div>
        {mfaType === SmsMfaType || mfaType === EmailMfaType ? (
          <MfaVerifySmsForm
            mfaProps={mfaProps}
            method={mfaAuth}
            onFinish={verify}
            application={application}
          />) : (
          <MfaVerifyTotpForm
            mfaProps={mfaProps}
            onFinish={verify}
          />
        )}
        <span style={{float: "right"}}>
          {i18next.t("mfa:Have problems?")}
          <a onClick={() => {
            setMfaType("recovery");
          }}>
            {i18next.t("mfa:Use a recovery code")}
          </a>
        </span>
      </div>
    );
  } else {
    return (
      <div style={{width: 300, height: 350}}>
        <div style={{marginBottom: 24, textAlign: "center", fontSize: "24px"}}>
          {i18next.t("mfa:Multi-factor recover")}
        </div>
        <div style={{marginBottom: 24}}>
          {i18next.t("mfa:Multi-factor recover description")}
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
          }}>{i18next.t("forget:Verify")}
        </Button>
        <span style={{float: "right"}}>
          {i18next.t("mfa:Have problems?")}
          <a onClick={() => {
            setMfaType(mfaProps.mfaType);
          }}>
            {i18next.t("mfa:Use SMS verification code")}
          </a>
        </span>
      </div>
    );
  }
}
