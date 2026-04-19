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

import {Button, Input, Space} from "antd";
import React from "react";
import i18next from "i18next";
import * as UserBackend from "../backend/UserBackend";
import * as AuthBackend from "../auth/AuthBackend";
import * as Setting from "../Setting";
import {SafetyOutlined} from "@ant-design/icons";
import {CaptchaModal} from "./modal/CaptchaModal";

export const SendCodeInput = ({value, disabled, captchaValue, useInlineCaptcha, textBefore, onChange, onButtonClickArgs, application, method, countryCode, refreshCaptcha}) => {
  const [visible, setVisible] = React.useState(false);
  const [buttonLeftTime, setButtonLeftTime] = React.useState(0);
  const [buttonLoading, setButtonLoading] = React.useState(false);

  const getCodeResendTimeout = () => {
    // Use application's codeResendTimeout if available, otherwise default to 60 seconds
    return (application && application.codeResendTimeout > 0) ? application.codeResendTimeout : 60;
  };

  const handleCountDown = (leftTime = getCodeResendTimeout()) => {
    let leftTimeSecond = leftTime;
    setButtonLeftTime(leftTimeSecond);
    const countDown = () => {
      leftTimeSecond--;
      setButtonLeftTime(leftTimeSecond);
      if (leftTimeSecond === 0) {
        return;
      }
      setTimeout(countDown, 1000);
    };
    setTimeout(countDown, 1000);
  };

  const handleOk = (captchaType, captchaToken, clintSecret) => {
    setVisible(false);
    setButtonLoading(true);
    UserBackend.sendCode(captchaType, captchaToken, clintSecret, method, countryCode, ...onButtonClickArgs).then(res => {
      setButtonLoading(false);
      if (res) {
        handleCountDown(getCodeResendTimeout());
      } else {
        if (useInlineCaptcha) {
          refreshCaptcha?.();
        }
      }
    }).catch(() => {
      setButtonLoading(false);
      if (useInlineCaptcha) {
        refreshCaptcha?.();
      }
    });
  };

  const handleCancel = () => {
    setVisible(false);
  };

  const handleSearch = () => {
    const sendCodeWithoutCaptcha = () => {
      handleOk("none", "", "");
    };

    const sendCodeWithCaptcha = () => {
      if (!useInlineCaptcha) {
        setVisible(true);
        return;
      }

      // client secret is validated in backend
      if (!captchaValue?.captchaType || !captchaValue?.captchaToken) {
        Setting.showMessage("error", i18next.t("general:Please complete the captcha correctly"));
        return;
      }

      handleOk(captchaValue.captchaType, captchaValue.captchaToken, captchaValue.clientSecret);
    };

    const checkCaptchaStatusAndSend = () => {
      if (!onButtonClickArgs?.[3]) {
        return;
      }
      const values = {
        organization: application?.organization,
        username: onButtonClickArgs?.[3],
        application: application?.name,
      };

      AuthBackend.getCaptchaStatus(values)
        .then((res) => {
          if (res.status === "ok" && res.data) {
            sendCodeWithCaptcha();
            return;
          }

          sendCodeWithoutCaptcha();
        })
        .catch(() => {
          sendCodeWithoutCaptcha();
        });
    };

    const captchaRule = Setting.getCaptchaRule(application);
    if (captchaRule === Setting.CaptchaRule.Never) {
      sendCodeWithoutCaptcha();
      return;
    }

    if (captchaRule === Setting.CaptchaRule.Always) {
      sendCodeWithCaptcha();
      return;
    }

    if (captchaRule === Setting.CaptchaRule.Dynamic || captchaRule === Setting.CaptchaRule.InternetOnly) {
      checkCaptchaStatusAndSend();
      return;
    }

    sendCodeWithoutCaptcha();
  };

  return (
    <React.Fragment>
      <Space.Compact style={{width: "100%"}}>
        <Input
          addonBefore={textBefore}
          disabled={disabled}
          value={value}
          prefix={<SafetyOutlined />}
          placeholder={i18next.t("code:Enter your code")}
          className="verification-code-input"
          onChange={e => onChange(e.target.value)}
          autoComplete="one-time-code"
        />
        <Button style={{fontSize: 14}} type={"primary"} disabled={disabled || buttonLeftTime > 0} loading={buttonLoading} onClick={handleSearch}>
          {buttonLeftTime > 0 ? `${buttonLeftTime} s` : buttonLoading ? i18next.t("code:Sending") : i18next.t("code:Send Code")}
        </Button>
      </Space.Compact>
      {
        useInlineCaptcha ? null : (
          <CaptchaModal
            owner={application.owner}
            name={application.name}
            visible={visible}
            onOk={handleOk}
            onCancel={handleCancel}
            isCurrentProvider={false}
          />
        )
      }
    </React.Fragment>
  );
};
