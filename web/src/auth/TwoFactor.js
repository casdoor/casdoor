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
import * as Setting from "../Setting";
import {twoFactorAuthTotp, twoFactorRecoverTotp} from "../backend/TwoFactorAuthBackend";

export const NextTwoFactor = "nextTwoFactor";

export function VerityTotp({onSuccess, onFail}) {
  const [loading, setLoading] = useState(false);
  const [currentStep, setCurrentStep] = useState(0);
  const [passcode, setPasscode] = useState("");
  const [recoveryCode, setRecoveryCode] = useState("");
  const verity = () => {
    if (passcode.length === 0) {
      Setting.showMessage("error", i18next.t("two-factor:Enter passcode"));
      return;
    }

    setLoading(true);
    twoFactorAuthTotp({passcode}).then((res) => {
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

  const recover = () => {
    if (recoveryCode.length === 0) {
      Setting.showMessage("error", i18next.t("two-factor:Enter recovery code"));
      return;
    }
    setLoading(true);
    twoFactorRecoverTotp({recoveryCode}).then(res => {
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

  switch (currentStep) {
  case 0:
    return (
      <div style={{width: 300}}>
        <div style={{marginBottom: 24, textAlign: "center", fontSize: "28px"}}>{i18next.t(
          "two-factor:Two-factor authentication")}</div>
        <div style={{marginBottom: 24}}>{i18next.t(
          "two-factor:Two-factor authentication description")}</div>
        <Input placeholder={i18next.t("two-factor:Passcode")}
          style={{marginBottom: 24}} type={"passcode"} size={"large"}
          onChange={event => setPasscode(event.target.value)} />
        <Button style={{width: "100%"}} size={"large"} loading={loading}
          type={"primary"} onClick={() => {
            verity();
          }}>{i18next.t("two-factor:Verity")}</Button>
        <span style={{float: "right"}}>
          {i18next.t("two-factor:Have problems?")}
          <a onClick={() => {
            setCurrentStep(1);
          }}>
            {i18next.t("two-factor:Use a recovery code")}
          </a>
        </span>
      </div>
    );
  case 1:
    return (
      <div style={{width: 300}}>
        <div style={{marginBottom: 24, textAlign: "center", fontSize: "28px"}}>{i18next.t(
          "two-factor:Two-factor recover")}</div>
        <div style={{marginBottom: 24}}>{i18next.t(
          "two-factor:Two-factor recover description")}</div>
        <Input placeholder={i18next.t("two-factor:Recovery code")}
          style={{marginBottom: 24}} type={"passcode"} size={"large"}
          onChange={event => setRecoveryCode(event.target.value)} />
        <Button style={{width: "100%"}} size={"large"} loading={loading}
          type={"primary"} onClick={() => {
            recover();
          }}>{i18next.t("two-factor:Verity")}</Button>
      </div>
    );
  default:
    return null;
  }
}
