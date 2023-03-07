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
import {CaptchaModal} from "./CaptchaModal";

const {Search} = Input;

export const SendCodeInput = (props) => {
  const {disabled, textBefore, onChange, onButtonClickArgs, application, method, countryCode} = props;
  const [visible, setVisible] = React.useState(false);

  const [buttonLeftTime, setButtonLeftTime] = React.useState(0);
  const [buttonLoading, setButtonLoading] = React.useState(false);

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

  const handleOk = (captchaType, captchaToken, clintSecret) => {
    setVisible(false);
    setButtonLoading(true);
    UserBackend.sendCode(captchaType, captchaToken, clintSecret, method, countryCode, ...onButtonClickArgs).then(res => {
      setButtonLoading(false);
      if (res) {
        handleCountDown(60);
      }
    });
  };

  const handleCancel = () => {
    setVisible(false);
  };

  return (
    <React.Fragment>
      <Search
        addonBefore={textBefore}
        disabled={disabled}
        prefix={<SafetyOutlined />}
        placeholder={i18next.t("code:Enter your code")}
        onChange={e => onChange(e.target.value)}
        enterButton={
          <Button style={{fontSize: 14}} type={"primary"} disabled={disabled || buttonLeftTime > 0} loading={buttonLoading}>
            {buttonLeftTime > 0 ? `${buttonLeftTime} s` : buttonLoading ? i18next.t("code:Sending Code") : i18next.t("code:Send Code")}
          </Button>
        }
        onSearch={() => setVisible(true)}
      />
      <CaptchaModal
        owner={application.owner}
        name={application.name}
        visible={visible}
        onOk={handleOk}
        onCancel={handleCancel}
        isCurrentProvider={false}
      />
    </React.Fragment>
  );
};
