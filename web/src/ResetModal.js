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

import {Button, Col, Modal, Row, Input,} from "antd";
import i18next from "i18next";
import React from "react";
import * as Setting from "./Setting"
import * as UserBackend from "./backend/UserBackend"

export const ResetModal = (props) => {
  const [visible, setVisible] = React.useState(false);
  const [confirmLoading, setConfirmLoading] = React.useState(false);
  const [sendButtonText, setSendButtonText] = React.useState(i18next.t("user:Send Code"));
  const [sendCodeCoolDown, setCoolDown] = React.useState(false);
  const {buttonText, destType, coolDownTime} = props;

  const showModal = () => {
    setVisible(true);
  };

  const handleCancel = () => {
    setVisible(false);
  };

  const handleOk = () => {
    let dest = document.getElementById("dest").value;
    let code = document.getElementById("code").value;
    if (dest === "") {
      Setting.showMessage("error", i18next.t("user:Empty " + destType));
      return;
    }
    if (code === "") {
      Setting.showMessage("error", i18next.t("user:Empty Code"));
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
    })
  }

  const countDown = (second) => {
    if (second <= 0) {
      setSendButtonText(i18next.t("user:Send Code"));
      setCoolDown(false);
      return;
    }
    setSendButtonText(second);
    setTimeout(() => countDown(second - 1), 1000);
  }

  const sendCode = () => {
    if (sendCodeCoolDown) return;
    let dest = document.getElementById("dest").value;
    if (dest === "") {
      Setting.showMessage("error", i18next.t("user:Empty " + destType));
      return;
    }
    UserBackend.sendCode(dest, destType).then(res => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("user:Code Sent"));
        setCoolDown(true);
        countDown(coolDownTime);
      } else {
        Setting.showMessage("error", i18next.t("user:" + res.msg));
      }
    })
  }

  let placeHolder = "";
  if (destType === "email") placeHolder = i18next.t("user:Input your email");
  else if (destType === "phone") placeHolder = i18next.t("user:Input your phone number");

  return (
    <Row>
      <Button style={{marginTop: '22px'}} type="default" onClick={showModal}>
        {buttonText}
      </Button>
      <Modal
        title={buttonText}
        visible={visible}
        okText={buttonText}
        cancelText={i18next.t("user:Cancel")}
        confirmLoading={confirmLoading}
        onCancel={handleCancel}
        onOk={handleOk}
        width={600}
      >
        <Col style={{margin: "0px auto 40px auto", width: 1000, height: 300}}>
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <Input addonBefore={i18next.t("user:New " + destType)} id="dest" placeholder={placeHolder}
                            addonAfter={<button style={{width: "90px", border: "none", backgroundColor: "#fff"}} onClick={sendCode}>{" " + sendButtonText + " "}</button>}
            />

          </Row>
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <Input addonBefore={i18next.t("user:Code You Received")} placeholder={i18next.t("user:Enter your code")} id="code"/>
          </Row>
        </Col>
      </Modal>
    </Row>
  )
}

export default ResetModal;