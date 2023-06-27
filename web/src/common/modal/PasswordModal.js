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
import * as UserBackend from "../../backend/UserBackend";
import * as Setting from "../../Setting";
import * as PasswordChecker from "../PasswordChecker";

export const PasswordModal = (props) => {
  const [visible, setVisible] = React.useState(false);
  const [confirmLoading, setConfirmLoading] = React.useState(false);
  const [oldPassword, setOldPassword] = React.useState("");
  const [newPassword, setNewPassword] = React.useState("");
  const [rePassword, setRePassword] = React.useState("");
  const {user} = props;
  const {organization} = props;
  const {account} = props;

  const [passwordOptions, setPasswordOptions] = React.useState([]);
  const [newPasswordValid, setNewPasswordValid] = React.useState(false);
  const [rePasswordValid, setRePasswordValid] = React.useState(false);
  const [newPasswordErrorMessage, setNewPasswordErrorMessage] = React.useState("");
  const [rePasswordErrorMessage, setRePasswordErrorMessage] = React.useState("");

  React.useEffect(() => {
    if (organization) {
      setPasswordOptions(organization.passwordOptions);
    }
  }, [user.owner]);
  const showModal = () => {
    setVisible(true);
  };

  const handleCancel = () => {
    setVisible(false);
  };
  const handleNewPassword = (value) => {
    setNewPassword(value);

    const errorMessage = PasswordChecker.checkPasswordComplexity(value, passwordOptions);
    setNewPasswordValid(errorMessage === "");
    setNewPasswordErrorMessage(errorMessage);
  };

  const handleRePassword = (value) => {
    setRePassword(value);

    if (value !== newPassword) {
      setRePasswordErrorMessage(i18next.t("signup:Your confirmed password is inconsistent with the password!"));
      setRePasswordValid(false);
    } else {
      setRePasswordValid(true);
    }
  };

  const handleOk = () => {
    if (newPassword === "" || rePassword === "") {
      Setting.showMessage("error", i18next.t("user:Empty input!"));
      return;
    }
    if (newPassword !== rePassword) {
      Setting.showMessage("error", i18next.t("user:Two passwords you typed do not match."));
      return;
    }
    setConfirmLoading(true);

    if (organization === null) {
      Setting.showMessage("error", "organization is null");
      setConfirmLoading(false);
      return;
    }

    const errorMsg = PasswordChecker.checkPasswordComplexity(newPassword, organization.passwordOptions);
    if (errorMsg !== "") {
      Setting.showMessage("error", errorMsg);
      setConfirmLoading(false);
      return;
    }

    const values = {"userOwner": user.owner, "userName": user.name, "currentPassword": oldPassword, "password": newPassword};
    UserBackend.setPassword(values)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("user:Password set successfully"));
          setVisible(false);
        } else {
          Setting.showMessage("error", i18next.t(`user:${res.msg}`));
        }
      })
      .finally(() => {
        setConfirmLoading(false);
      });
  };

  const hasOldPassword = user.password !== "";

  return (
    <Row>
      <Button type="default" disabled={props.disabled} onClick={showModal}>
        {hasOldPassword ? i18next.t("user:Modify password...") : i18next.t("user:Set password...")}
      </Button>
      <Modal
        maskClosable={false}
        title={i18next.t("general:Password")}
        open={visible}
        okText={i18next.t("user:Set Password")}
        cancelText={i18next.t("general:Cancel")}
        confirmLoading={confirmLoading}
        onCancel={handleCancel}
        onOk={handleOk}
        width={600}
      >
        <Col style={{margin: "0px auto 40px auto", width: 1000, height: 300}}>
          {(hasOldPassword && !Setting.isAdminUser(account)) ? (
            <Row style={{width: "100%", marginBottom: "20px"}}>
              <Input.Password addonBefore={i18next.t("user:Old Password")} placeholder={i18next.t("user:input password")} onChange={(e) => setOldPassword(e.target.value)} />
            </Row>
          ) : null}
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <Input.Password
              addonBefore={i18next.t("user:New Password")}
              placeholder={i18next.t("user:input password")}
              onChange={(e) => {handleNewPassword(e.target.value);}}
              status={(!newPasswordValid && newPasswordErrorMessage) ? "error" : undefined}
            />
          </Row>
          {!newPasswordValid && newPasswordErrorMessage && <div style={{color: "red", marginTop: "-20px"}}>{newPasswordErrorMessage}</div>}
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <Input.Password
              addonBefore={i18next.t("user:Re-enter New")}
              placeholder={i18next.t("user:input password")}
              onChange={(e) => handleRePassword(e.target.value)}
              status={(!rePasswordValid && rePasswordErrorMessage) ? "error" : undefined}
            />
          </Row>
          {!rePasswordValid && rePasswordErrorMessage && <div style={{color: "red", marginTop: "-20px"}}>{rePasswordErrorMessage}</div>}
        </Col>
      </Modal>
    </Row>
  );
};

export default PasswordModal;
