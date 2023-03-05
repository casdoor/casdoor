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
import i18next from "i18next";
import React from "react";
import * as Setting from "./Setting";
import * as UserBackend from "./backend/UserBackend";
import {SendCodeInput} from "./common/SendCodeInput";
import {MailOutlined, PhoneOutlined} from "@ant-design/icons";

export const ResetModal = (props) => {
  const [visible, setVisible] = React.useState(false);
  const [confirmLoading, setConfirmLoading] = React.useState(false);
  const [dest, setDest] = React.useState("");
  const [code, setCode] = React.useState("");
  const {buttonText, destType, application, countryCode} = props;

  const showModal = () => {
    setVisible(true);
  };

  const handleCancel = () => {
    setVisible(false);
  };

  const handleOk = () => {
    if (dest === "") {
      if (destType === "phone") {
        Setting.showMessage("error", i18next.t("user:Phone cannot be empty"));
      } else {
        Setting.showMessage("error", i18next.t("user:Email cannot be empty"));
      }
      return;
    }
    if (code === "") {
      Setting.showMessage("error", i18next.t("code:Empty Code"));
      return;
    }
    setConfirmLoading(true);
    UserBackend.resetEmailOrPhone(dest, destType, code).then(res => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("user:" + destType + " reset"));
        window.location.reload();
      } else {
        Setting.showMessage("error", i18next.t("user:" + res.msg));
        setConfirmLoading(false);
      }
    });
  };

  let placeholder = "";
  if (destType === "email") {
    placeholder = i18next.t("user:Input your email");
  } else if (destType === "phone") {
    placeholder = i18next.t("user:Input your phone number");
  }

  return (
    <Row>
      <Button type="default" onClick={showModal}>
        {buttonText}
      </Button>
      <Modal
        maskClosable={false}
        title={buttonText}
        open={visible}
        okText={buttonText}
        cancelText={i18next.t("user:Cancel")}
        confirmLoading={confirmLoading}
        onCancel={handleCancel}
        onOk={handleOk}
        width={600}
      >
        <Col style={{margin: "0px auto 40px auto", width: 1000, height: 300}}>
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <Input
              addonBefore={destType === "email" ? i18next.t("user:New Email") : i18next.t("user:New phone")}
              prefix={destType === "email" ? <React.Fragment><MailOutlined />&nbsp;&nbsp;</React.Fragment> : (<React.Fragment><PhoneOutlined />&nbsp;&nbsp;{countryCode !== "" ? "+" : null}{Setting.getCountryCode(countryCode)}&nbsp;</React.Fragment>)}
              placeholder={placeholder}
              onChange={e => setDest(e.target.value)}
            />
          </Row>
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <SendCodeInput
              textBefore={i18next.t("code:Code You Received")}
              onChange={setCode}
              method={"reset"}
              onButtonClickArgs={[dest, destType, Setting.getApplicationName(application)]}
              application={application}
            />
          </Row>
        </Col>
      </Modal>
    </Row>
  );
};

export default ResetModal;
