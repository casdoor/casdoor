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
import React from "react";
import * as UserBackend from "./backend/UserBackend";
import * as Setting from "./Setting";

export const PasswordModal = (props) => {
  const [visible, setVisible] = React.useState(false);
  const [confirmLoading, setConfirmLoading] = React.useState(false);
  const [oldPassword, setOldPassword] = React.useState("");
  const [newPassword, setNewPassword] = React.useState("");
  const [rePassword, setRePassword] = React.useState("");
  const {user} = props;

  const showModal = () => {
    setVisible(true);
  };

  const handleCancel = () => {
    setVisible(false);
  };

  const handleOk = () => {
    if (newPassword === "" || rePassword === "") {
      Setting.showMessage("error", Setting.I18n("user:Empty input!"));
      return;
    }
    if (newPassword !== rePassword) {
      Setting.showMessage("error", Setting.I18n("user:Two passwords you typed do not match."));
      return;
    }
    setConfirmLoading(true);
    UserBackend.setPassword(user.owner, user.name, oldPassword, newPassword).then((res) => {
      setConfirmLoading(false);
      if (res.status === "ok") {
        Setting.showMessage("success", Setting.I18n("user:Password Set"));
        setVisible(false);
      }
      else Setting.showMessage("error", Setting.I18n(`user:${res.msg}`));
    })
  }

  let hasOldPassword = user.password !== "";

  return (
    <Row>
      <Button type="default" onClick={showModal}>
        { hasOldPassword ? Setting.I18n("user:Modify password...") : Setting.I18n("user:Set password...")}
      </Button>
      <Modal
        maskClosable={false}
        title={Setting.I18n("user:Password")}
        visible={visible}
        okText={Setting.I18n("user:Set Password")}
        cancelText={Setting.I18n("user:Cancel")}
        confirmLoading={confirmLoading}
        onCancel={handleCancel}
        onOk={handleOk}
        width={600}
      >
        <Col style={{margin: "0px auto 40px auto", width: 1000, height: 300}}>
          { hasOldPassword ? (
            <Row style={{width: "100%", marginBottom: "20px"}}>
              <Input.Password addonBefore={Setting.I18n("user:Old Password")} placeholder={Setting.I18n("user:input password")} onChange={(e) => setOldPassword(e.target.value)}/>
            </Row>
          ) : null}
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <Input.Password addonBefore={Setting.I18n("user:New Password")} placeholder={Setting.I18n("user:input password")} onChange={(e) => setNewPassword(e.target.value)}/>
          </Row>
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <Input.Password addonBefore={Setting.I18n("user:Re-enter New")} placeholder={Setting.I18n("user:input password")} onChange={(e) => setRePassword(e.target.value)}/>
          </Row>
        </Col>
      </Modal>
    </Row>
  )
}

export default PasswordModal;