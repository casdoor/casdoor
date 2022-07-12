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
import React from "react";
import i18next from "i18next";
import * as UserBackend from "../backend/UserBackend";
import * as ProviderBackend from "../backend/ProviderBackend";
import {SafetyOutlined} from "@ant-design/icons";
import {CaptchaWidget} from "./CaptchaWidget";

export const CaptchaPreview = ({
  provider,
  providerName,
  clientSecret,
  captchaType,
  subType,
  owner,
  clientId,
  name,
  providerUrl,
  clientId2,
  clientSecret2,
}) => {
  const [visible, setVisible] = React.useState(false);
  const [captchaImg, setCaptchaImg] = React.useState("");
  const [captchaToken, setCaptchaToken] = React.useState("");
  const [secret, setSecret] = React.useState(clientSecret);
  const [secret2, setSecret2] = React.useState(clientSecret2);

  const handleOk = () => {
    UserBackend.verifyCaptcha(captchaType, captchaToken, secret).then(() => {
      setCaptchaToken("");
      setVisible(false);
    });
  };

  const handleCancel = () => {
    setVisible(false);
  };

  const getCaptchaFromBackend = () => {
    UserBackend.getCaptcha(owner, name, true).then((res) => {
      if (captchaType === "Default") {
        setSecret(res.captchaId);
        setCaptchaImg(res.captchaImage);
      } else {
        setSecret(res.clientSecret);
        setSecret2(res.clientSecret2);
      }
    });
  };

  const clickPreview = () => {
    setVisible(true);
    provider.name = name;
    provider.clientId = clientId;
    provider.type = captchaType;
    provider.providerUrl = providerUrl;
    if (clientSecret !== "***") {
      provider.clientSecret = clientSecret;
      ProviderBackend.updateProvider(owner, providerName, provider).then(() => {
        getCaptchaFromBackend();
      });
    } else {
      getCaptchaFromBackend();
    }
  };

  const renderDefaultCaptcha = () => {
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
            marginBottom: 10,
          }}
        />
        <Row>
          <Input
            autoFocus
            value={captchaToken}
            prefix={<SafetyOutlined />}
            placeholder={i18next.t("general:Captcha")}
            onPressEnter={handleOk}
            onChange={(e) => setCaptchaToken(e.target.value)}
          />
        </Row>
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

  const getButtonDisabled = () => {
    if (captchaType !== "Default") {
      if (!clientId || !clientSecret) {
        return true;
      }
      if (captchaType === "Aliyun Captcha") {
        if (!subType || !clientId2 || !clientSecret2) {
          return true;
        }
      }
    }
    return false;
  };

  return (
    <React.Fragment>
      <Button
        style={{fontSize: 14}}
        type={"primary"}
        onClick={clickPreview}
        disabled={getButtonDisabled()}
      >
        {i18next.t("general:Preview")}
      </Button>
      <Modal
        closable={false}
        maskClosable={false}
        destroyOnClose={true}
        title={i18next.t("general:Captcha")}
        visible={visible}
        okText={i18next.t("user:OK")}
        cancelText={i18next.t("user:Cancel")}
        onOk={handleOk}
        onCancel={handleCancel}
        width={348}
      >
        {renderCheck()}
      </Modal>
    </React.Fragment>
  );
};
