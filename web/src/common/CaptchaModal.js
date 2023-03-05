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

export const CaptchaModal = ({
  owner,
  name,
  captchaType,
  subType,
  clientId,
  clientId2,
  clientSecret,
  clientSecret2,
  open,
  onOk,
  onCancel,
  canCancel,
}) => {
  const [visible, setVisible] = React.useState(false);
  const [captchaImg, setCaptchaImg] = React.useState("");
  const [captchaToken, setCaptchaToken] = React.useState("");
  const [secret, setSecret] = React.useState(clientSecret);
  const [secret2, setSecret2] = React.useState(clientSecret2);
  const defaultInputRef = React.useRef(null);

  useEffect(() => {
    setVisible(() => {
      if (open) {
        getCaptchaFromBackend();
      } else {
        cleanUp();
      }
      return open;
    });
  }, [open]);

  const handleOk = () => {
    onOk?.(captchaType, captchaToken, secret);
  };

  const handleCancel = () => {
    onCancel?.();
  };

  const cleanUp = () => {
    setCaptchaToken("");
  };

  const getCaptchaFromBackend = () => {
    UserBackend.getCaptcha(owner, name, true).then((res) => {
      if (captchaType === "Default") {
        setSecret(res.captchaId);
        setCaptchaImg(res.captchaImage);

        defaultInputRef.current?.focus();
      } else {
        setSecret(res.clientSecret);
        setSecret2(res.clientSecret2);
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

  const onSubmit = (token) => {
    setCaptchaToken(token);
  };

  const renderCheck = () => {
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
              clientSecret={secret}
              onChange={onSubmit}
              clientId2={clientId2}
              clientSecret2={secret2}
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
    }

    if (canCancel) {
      return [
        <Button key="cancel" onClick={handleCancel}>{i18next.t("user:Cancel")}</Button>,
        <Button key="ok" disabled={isOkDisabled} type="primary" onClick={handleOk}>{i18next.t("user:OK")}</Button>,
      ];
    } else {
      return [
        <Button key="ok" disabled={isOkDisabled} type="primary" onClick={handleOk}>{i18next.t("user:OK")}</Button>,
      ];
    }
  };

  return (
    <React.Fragment>
      <Modal
        closable={false}
        maskClosable={false}
        destroyOnClose={true}
        title={i18next.t("general:Captcha")}
        open={visible}
        width={348}
        footer={renderFooter()}
      >
        <div style={{marginTop: "20px", marginBottom: "50px"}}>
          {
            renderCheck()
          }
        </div>
      </Modal>
    </React.Fragment>
  );
};
