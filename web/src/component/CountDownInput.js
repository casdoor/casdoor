// Copyright 2021 The casbin Authors. All Rights Reserved.
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
import * as Setting from "../Setting";
import * as UserBackend from "../backend/UserBackend";
import { AuditOutlined, VerifiedOutlined } from "@ant-design/icons";

const { Search } = Input;

export const CountDownInput = (props) => {
  const {defaultButtonText, disabled, prefix, textBefore, placeHolder, onChange, coolDownTime, onButtonClick, onButtonClickArgs} = props;
  const [buttonText, setButtonText] = React.useState(defaultButtonText);
  const [visible, setVisible] = React.useState(false);
  const [key, setKey] = React.useState("");
  const [captchaImg, setCaptchaImg] = React.useState("");
  const [checkType, setCheckType] = React.useState("");
  const [coolDown, setCoolDown] = React.useState(false);
  const [checkId, setCheckId] = React.useState("");
  const [buttonDisabled, setButtonDisabled] = React.useState(false);

  const countDown = (leftTime) => {
    if (leftTime === 0) {
      setButtonDisabled(false);
      setCoolDown(false);
      setButtonText(defaultButtonText);
      return;
    }
    setButtonText(`${leftTime} s`);
    setTimeout(() => countDown(leftTime - 1), 1000);
  }

  const handleOk = () => {
    setVisible(false);
    onButtonClick(checkType, checkId, key, ...onButtonClickArgs).then(res => {
      setKey("");
      if (res) {
        setCoolDown(true);
        setButtonDisabled(true)
        countDown(coolDownTime);
      }
    })
  }

  const handleCancel = () => {
    setVisible(false);
    setKey("");
  }

  const loadHumanCheck = () => {
    UserBackend.getHumanCheck().then(res => {
      if (res.type === "none") {
        onButtonClick("none", "", "", ...onButtonClickArgs);
      } else if (res.type === "captcha") {
        setCheckId(res.captchaId);
        setCaptchaImg(res.captchaImage);
        setCheckType("captcha");
        setVisible(true);
      } else {
        Setting.showMessage("error", Setting.I18n("signup:Unknown Check Type"));
      }
    })
  }

  const renderCaptcha = () => {
    return (
      <Col>
        <Row
          style={{
            backgroundImage: `url('data:image/png;base64,${captchaImg}')`,
            backgroundRepeat: "no-repeat",
            height: "80px",
            width: "200px",
            borderRadius: "3px",
            border: "1px solid #ccc",
            marginBottom: 10
          }}
        />
        <Row>
          <Input autoFocus value={key} placeholder={Setting.I18n("general:Captcha")} onPressEnter={handleOk} onChange={e => setKey(e.target.value)} />
        </Row>
      </Col>
    )
  }

  const renderCheck = () => {
    if (checkType === "captcha") return renderCaptcha();
    return null;
  }

  const getIcon = (prefix) => {
    switch (prefix) {
      case "VerifiedOutlined":
        return <VerifiedOutlined />;
      case "AuditOutlined":
        return <AuditOutlined />;
    }
  };

  return (
    <div>
      <Search
        addonBefore={textBefore}
        disabled={disabled}
        prefix={prefix !== null ? getIcon(prefix) : null}
        placeholder={placeHolder}
        onChange={e => onChange(e.target.value)}
        enterButton={
          <Button type={"primary"} disabled={disabled || buttonDisabled}>
            <div style={{fontSize: 14}}>
              {buttonText}
            </div>
          </Button>
        }
        onSearch={loadHumanCheck}
      />
      <Modal
        closable={false}
        maskClosable={false}
        destroyOnClose={true}
        title={Setting.I18n("general:Captcha")}
        visible={visible}
        okText={Setting.I18n("user:OK")}
        cancelText={Setting.I18n("user:Cancel")}
        onCancel={handleCancel}
        onOk={handleOk}
        width={248}
      >
        {renderCheck()}
      </Modal>
    </div>
  );
}
