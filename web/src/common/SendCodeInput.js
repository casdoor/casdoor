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

import {Button, Col, Input, Modal, Row} from "antd";
import React from "react";
import i18next from "i18next";
import * as UserBackend from "../backend/UserBackend";
import {SafetyOutlined} from "@ant-design/icons";
import {CaptchaWidget} from "./CaptchaWidget";

const {Search} = Input;

export const SendCodeInput = (props) => {
  const {disabled, textBefore, onChange, onButtonClickArgs, application, method, countryCode} = props;
  const [visible, setVisible] = React.useState(false);
  const [key, setKey] = React.useState("");
  const [captchaImg, setCaptchaImg] = React.useState("");
  const [checkType, setCheckType] = React.useState("");
  const [checkId, setCheckId] = React.useState("");
  const [buttonLeftTime, setButtonLeftTime] = React.useState(0);
  const [buttonLoading, setButtonLoading] = React.useState(false);
  const [buttonDisabled, setButtonDisabled] = React.useState(true);
  const [clientId, setClientId] = React.useState("");
  const [subType, setSubType] = React.useState("");
  const [clientId2, setClientId2] = React.useState("");
  const [clientSecret2, setClientSecret2] = React.useState("");

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

  const handleOk = () => {
    setVisible(false);
    setButtonLoading(true);
    UserBackend.sendCode(checkType, checkId, key, method, countryCode, ...onButtonClickArgs).then(res => {
      setKey("");
      setButtonLoading(false);
      if (res) {
        handleCountDown(60);
      }
    });
  };

  const handleCancel = () => {
    setVisible(false);
    setKey("");
  };

  const loadCaptcha = () => {
    UserBackend.getCaptcha(application.owner, application.name, false).then(res => {
      if (res.type === "none") {
        UserBackend.sendCode("none", "", "", method, countryCode, ...onButtonClickArgs).then(res => {
          if (res) {
            handleCountDown(60);
          }
        });
      } else if (res.type === "Default") {
        setCheckId(res.captchaId);
        setCaptchaImg(res.captchaImage);
        setCheckType("Default");
        setVisible(true);
      } else {
        setCheckType(res.type);
        setClientId(res.clientId);
        setCheckId(res.clientSecret);
        setVisible(true);
        setSubType(res.subType);
        setClientId2(res.clientId2);
        setClientSecret2(res.clientSecret2);
      }
    });
  };

  const renderCaptcha = () => {
    return (
      <Col>
        <Row
          style={{
            backgroundImage: `url('data:image/png;base64,${captchaImg}')`,
            backgroundRepeat: "no-repeat",
            height: "80px",
            width: "200px",
            borderRadius: "5px",
            border: "1px solid #ccc",
            marginBottom: 10,
          }}
        />
        <Row>
          <Input autoFocus value={key} prefix={<SafetyOutlined />} placeholder={i18next.t("general:Captcha")} onPressEnter={handleOk} onChange={e => setKey(e.target.value)} />
        </Row>
      </Col>
    );
  };

  const onSubmit = (token) => {
    setButtonDisabled(false);
    setKey(token);
  };

  const renderCheck = () => {
    if (checkType === "Default") {
      return renderCaptcha();
    } else {
      return (
        <CaptchaWidget
          captchaType={checkType}
          subType={subType}
          siteKey={clientId}
          clientSecret={checkId}
          onChange={onSubmit}
          clientId2={clientId2}
          clientSecret2={clientSecret2}
        />
      );
    }
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
        onSearch={loadCaptcha}
      />
      <Modal
        closable={false}
        maskClosable={false}
        destroyOnClose={true}
        title={i18next.t("general:Captcha")}
        open={visible}
        okText={i18next.t("user:OK")}
        cancelText={i18next.t("user:Cancel")}
        onOk={handleOk}
        onCancel={handleCancel}
        okButtonProps={{disabled: key.length !== 5 && buttonDisabled}}
        width={348}
      >
        {
          renderCheck()
        }
      </Modal>
    </React.Fragment>
  );
};
