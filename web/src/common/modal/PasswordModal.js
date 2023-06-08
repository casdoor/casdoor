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
import * as OrganizationBackend from "../../backend/OrganizationBackend";

export const PasswordModal = (props) => {
  const [visible, setVisible] = React.useState(false);
  const [confirmLoading, setConfirmLoading] = React.useState(false);
  const [oldPassword, setOldPassword] = React.useState("");
  const [newPassword, setNewPassword] = React.useState("");
  const [rePassword, setRePassword] = React.useState("");
  const {user} = props;
  const {account} = props;
  function checkPasswordComplexOption(password, complexOptions) {
    /*
      AtLeast8: The password length must be greater than 8
      Aa123: The password must contain at least one lowercase letter, one uppercase letter, and one digit
      SpecialChar: The password must contain at least one special character
      NoRepeat: The password must not contain any repeated characters
    */
    if (complexOptions.length < 1) {
      return "";
    }

    function isValidOption_atLeast8(password) {
      if (password.length < 8) {
        return "AtLeast8";
      }
      return "";
    }

    function isValidOption_Aa123(password) {
      const regex = /^(?=.*[a-z])(?=.*[A-Z])(?=.*[0-9]).+$/;
      if (!regex.test(password)) {
        return "Aa123";
      }
      return "";
    }

    function isValidOption_SpecialChar(password) {
      const regex = /^(?=.*[!@#$%^&*]).+$/;
      if (!regex.test(password)) {
        return "SpecialChar";
      }
      return "";
    }

    function isValidOption_noRepeat(password) {
      const regex = /(.)\1+/;
      if (regex.test(password)) {
        return "NoRepeat";
      }
      return "";
    }

    const validators = {
      AtLeast8: isValidOption_atLeast8,
      Aa123: isValidOption_Aa123,
      SpecialChar: isValidOption_SpecialChar,
      NoRepeat: isValidOption_noRepeat,
    };
    for (const option of complexOptions) {
      const validateFunc = validators[option];
      if (validateFunc) {
        const msg = validateFunc(password);
        if (msg !== "") {
          return msg;
        }
      } else {
        // Invalid complex option
        return "InvalidOption";
      }
    }

    return "";
  }
  const showModal = () => {
    setVisible(true);
  };

  const handleCancel = () => {
    setVisible(false);
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

    OrganizationBackend.getOrganizations("admin").then((res) => {
      const organizations = (res.msg === undefined) ? res : [];

      // find the users' corresponding organization
      let organization = null;
      for (let i = 0; i < organizations.length; i++) {
        if (organizations[i].name === user.owner) {
          organization = organizations[i];
          break;
        }
      }

      const pwdComplexOptionList = (organization["passwordComplexOptions"]);

      const pwdCheckRes = checkPasswordComplexOption(newPassword, pwdComplexOptionList);
      if (pwdCheckRes !== "") {
        Setting.showMessage("error", i18next.t(`passwordComplexOptions:${pwdCheckRes}`));
        setConfirmLoading(false);
        return;
      }

      UserBackend.setPassword(user.owner, user.name, oldPassword, newPassword)
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
            <Input.Password addonBefore={i18next.t("user:New Password")} placeholder={i18next.t("user:input password")} onChange={(e) => setNewPassword(e.target.value)} />
          </Row>
          <Row style={{width: "100%", marginBottom: "20px"}}>
            <Input.Password addonBefore={i18next.t("user:Re-enter New")} placeholder={i18next.t("user:input password")} onChange={(e) => setRePassword(e.target.value)} />
          </Row>
        </Col>
      </Modal>
    </Row>
  );
};

export default PasswordModal;
