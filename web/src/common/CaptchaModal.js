// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
import i18next from "i18next";
import React, {useEffect} from "react";
import * as UserBackend from "../backend/UserBackend";
import {CaptchaWidget} from "./CaptchaWidget";
import {SafetyOutlined} from "@ant-design/icons";

export const CaptchaModal = (props) => {
  const {owner, name, visible, onOk, onCancel, isCurrentProvider} = props;

  const [captchaType, setCaptchaType] = React.useState("none");
  const [clientId, setClientId] = React.useState("");
  const [clientSecret, setClientSecret] = React.useState("");
  const [subType, setSubType] = React.useState("");
  const [clientId2, setClientId2] = React.useState("");
  const [clientSecret2, setClientSecret2] = React.useState("");

  const [open, setOpen] = React.useState(false);
  const [captchaImg, setCaptchaImg] = React.useState("");
  const [captchaToken, setCaptchaToken] = React.useState("");

  const defaultInputRef = React.useRef(null);

  useEffect(() => {
    if (visible) {
      loadCaptcha();
    } else {
      handleCancel();
      setOpen(false);
    }
  }, [visible]);

  const handleOk = () => {
    onOk?.(captchaType, captchaToken, clientSecret);
  };

  const handleCancel = () => {
    setCaptchaToken("");
    onCancel?.();
  };

  const loadCaptcha = () => {
    UserBackend.getCaptcha(owner, name, isCurrentProvider).then((res) => {
      if (res.type === "none") {
        handleOk();
      } else if (res.type === "Default") {
        setOpen(true);
        setClientSecret(res.captchaId);
        setCaptchaImg(res.captchaImage);
        setCaptchaType("Default");
      } else {
        setOpen(true);
        setCaptchaType(res.type);
        setClientId(res.clientId);
        setClientSecret(res.clientSecret);
        setSubType(res.subType);
        setClientId2(res.clientId2);
        setClientSecret2(res.clientSecret2);
      }
    });
  };

  const renderDefaultCaptcha = () => {
    return (
      <Col style={{textAlign: "center"}}>
        <div style={{display: "inline-block"}}>
          <Row
            style={{
              backgroundImage: `url('data:image/png;base64,${captchaImg}')`,
              backgroundRepeat: "no-repeat",
              height: "80px",
              width: "200px",
              borderRadius: "5px",
              border: "1px solid #ccc",
              marginBottom: "20px",
            }}
          />
          <Row>
            <Input
              ref={defaultInputRef}
              style={{width: "200px"}}
              value={captchaToken}
              prefix={<SafetyOutlined />}
              placeholder={i18next.t("general:Captcha")}
              onPressEnter={handleOk}
              onChange={(e) => setCaptchaToken(e.target.value)}
            />
          </Row>
        </div>
      </Col>
    );
  };

  const onChange = (token) => {
    setCaptchaToken(token);
  };

  const renderCaptcha = () => {
    if (captchaType === "Default") {
      return renderDefaultCaptcha();
    } else {
      return (
        <Col>
          <Row>
            <CaptchaWidget
              captchaType={captchaType}
              subType={subType}
              siteKey={clientId}
              clientSecret={clientSecret}
              onChange={onChange}
              clientId2={clientId2}
              clientSecret2={clientSecret2}
            />
          </Row>
        </Col>
      );
    }
  };

  const renderFooter = () => {
    let isOkDisabled = false;
    if (captchaType === "Default") {
      const regex = /^\d{5}$/;
      if (!regex.test(captchaToken)) {
        isOkDisabled = true;
      }
    } else if (captchaToken === "") {
      isOkDisabled = true;
    }

    return [
      <Button key="cancel" onClick={handleCancel}>{i18next.t("user:Cancel")}</Button>,
      <Button key="ok" disabled={isOkDisabled} type="primary" onClick={handleOk}>{i18next.t("user:OK")}</Button>,
    ];
  };

  return (
    <Modal
      closable={false}
      maskClosable={false}
      destroyOnClose={true}
      title={i18next.t("general:Captcha")}
      open={open}
      okText={i18next.t("user:OK")}
      cancelText={i18next.t("user:Cancel")}
      width={350}
      footer={renderFooter()}
      onCancel={handleCancel}
      onOk={handleOk}
    >
      <div style={{marginTop: "20px", marginBottom: "50px"}}>
        {
          renderCaptcha()
        }
      </div>
    </Modal>
  );
};
