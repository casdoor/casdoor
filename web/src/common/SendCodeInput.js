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

import {Button, Input} from "antd";
import React from "react";
import i18next from "i18next";
import * as UserBackend from "../backend/UserBackend";
import {SafetyOutlined} from "@ant-design/icons";
import {CaptchaModal} from "./modal/CaptchaModal";

const {Search} = Input;

export const SendCodeInput = React.forwardRef(({value, disabled, textBefore, onChange, onButtonClickArgs, application, method, countryCode, useInlineCaptcha, captchaValues, onCaptchaError, onCodeSent}, ref) => {
  const [visible, setVisible] = React.useState(false);
  const [buttonLeftTime, setButtonLeftTime] = React.useState(0);
  const [buttonLoading, setButtonLoading] = React.useState(false);
  const [codeSent, setCodeSent] = React.useState(false);
  const [captchaDisabled, setCaptchaDisabled] = React.useState(false);
  const [failedAttempts, setFailedAttempts] = React.useState(0);
  const [maxAttempts] = React.useState(5);

  // Expose methods to parent component
  React.useImperativeHandle(ref, () => ({
    handleCodeFailed: handleCodeFailed,
    resetAttempts: () => setFailedAttempts(0),
  }));

  const handleCountDown = (leftTime = 60) => {
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

  const handleOk = (captchaType, captchaToken, clientSecret) => {
    setVisible(false);
    setButtonLoading(true);
    UserBackend.sendCode(captchaType, captchaToken, clientSecret, method, countryCode, ...onButtonClickArgs).then(res => {
      setButtonLoading(false);
      if (res) {
        handleCountDown(60);
        setCodeSent(true);
        // Disable captcha completely after code is sent successfully
        setCaptchaDisabled(true);
        // Notify parent component that code was sent
        if (onCodeSent) {
          onCodeSent();
        }
      } else {
        // If sendCode failed (e.g., wrong captcha), trigger captcha refresh for inline captcha
        if (useInlineCaptcha && onCaptchaError && !captchaDisabled) {
          onCaptchaError();
        }
      }
    });
  };

  const handleCancel = () => {
    setVisible(false);
  };

  const handleCodeFailed = () => {
    const newFailedAttempts = failedAttempts + 1;
    setFailedAttempts(newFailedAttempts);

    // If max attempts reached, disable the input
    if (newFailedAttempts >= maxAttempts) {
      // Disable the input field
    }
  };

  const handleSendCode = () => {
    // If captcha is disabled (after code was sent successfully), resend code without captcha
    if (captchaDisabled) {
      setButtonLoading(true);
      UserBackend.sendCode("none", "", "", method, countryCode, ...onButtonClickArgs).then(res => {
        setButtonLoading(false);
        if (res) {
          handleCountDown(60);
        }
      });
      return;
    }

    // If inline captcha is enabled and we have captcha values, use them directly
    if (useInlineCaptcha) {
      if (captchaValues?.captchaToken) {
        handleOk(captchaValues.captchaType, captchaValues.captchaToken, captchaValues.clientSecret);
      }
      // If inline captcha is enabled but not filled yet, do nothing
    } else {
      // Otherwise, show the captcha modal
      setVisible(true);
    }
  };

  return (
    <React.Fragment>
      {!codeSent ? (
        // Before code is sent: Show only the Send Code button
        <Button
          style={{width: "100%"}}
          type="primary"
          disabled={disabled || buttonLeftTime > 0}
          loading={buttonLoading}
          onClick={handleSendCode}
        >
          {buttonLeftTime > 0 ? `${buttonLeftTime} s` : buttonLoading ? i18next.t("code:Sending") : i18next.t("code:Send Code")}
        </Button>
      ) : (
        // After code is sent: Show the input field with countdown button
        <Search
          addonBefore={textBefore}
          disabled={disabled || failedAttempts >= maxAttempts}
          value={value}
          prefix={<SafetyOutlined />}
          placeholder={failedAttempts >= maxAttempts ? i18next.t("code:Too many failed attempts") : i18next.t("code:Enter your code")}
          className="verification-code-input"
          onChange={e => onChange(e.target.value)}
          enterButton={
            <Button style={{fontSize: 14}} type={"primary"} disabled={disabled || buttonLeftTime > 0 || failedAttempts >= maxAttempts} loading={buttonLoading}>
              {buttonLeftTime > 0 ? `${buttonLeftTime} s` : buttonLoading ? i18next.t("code:Sending") : i18next.t("code:Send Code")}
            </Button>
          }
          onSearch={handleSendCode}
          autoComplete="one-time-code"
        />
      )}
      {!useInlineCaptcha && (
        <CaptchaModal
          owner={application.owner}
          name={application.name}
          visible={visible}
          onOk={handleOk}
          onCancel={handleCancel}
          isCurrentProvider={false}
        />
      )}
    </React.Fragment>
  );
});
