// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import {Button, Modal} from "antd";
import i18next from "i18next";
import React from "react";
import {useEffect, useState} from "react";
import {EmailMfaType} from "../../auth/MfaSetupPage";
import * as MfaBackend from "../../backend/MfaBackend";
import * as Setting from "../../Setting";

const EnableMfaModal = ({user, mfaType, onSuccess}) => {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open || !user) {
      return;
    }
    MfaBackend.MfaSetupInitiate({
      mfaType,
      ...user,
    }).then((res) => {
      if (res.status === "error") {
        Setting.showMessage("error", i18next.t("mfa:Failed to initiate MFA"));
      }
    });
  }, [open]);

  const handleOk = () => {
    setLoading(true);
    MfaBackend.MfaSetupEnable({
      mfaType,
      ...user,
    }).then(res => {
      if (res.status === "ok") {
        Setting.showMessage("success", i18next.t("general:Enabled successfully"));
        setOpen(false);
        onSuccess();
      } else {
        Setting.showMessage("error", `${i18next.t("general:Failed to enable")}: ${res.msg}`);
      }
    }
    ).finally(() => {
      setLoading(false);
    });
  };

  const handleCancel = () => {
    setOpen(false);
  };

  const showModal = () => {
    if (!isValid()) {
      if (mfaType === EmailMfaType) {
        Setting.showMessage("error", i18next.t("signup:Please input your Email!"));
      } else {
        Setting.showMessage("error", i18next.t("signup:Please input your phone number!"));
      }
      return;
    }
    setOpen(true);
  };

  const renderText = () => {
    return (
      <p>{i18next.t("mfa:Please confirm the information below")}<br />
        <b>{i18next.t("general:User")}</b>: {`${user.owner}/${user.name}`}<br />
        {mfaType === EmailMfaType ?
          <><b>{i18next.t("general:Email")}</b> : {user.email}</> :
          <><b>{i18next.t("general:Phone")}</b> : {user.phone}</>}
      </p>
    );
  };

  const isValid = () => {
    if (mfaType === EmailMfaType) {
      return user.email !== "";
    } else {
      return user.phone !== "";
    }
  };

  return (
    <React.Fragment>
      <Button type="primary" onClick={showModal}>
        {i18next.t("general:Enable")}
      </Button>
      <Modal
        title={i18next.t("mfa:Enable multi-factor authentication")}
        open={open}
        onOk={handleOk}
        onCancel={handleCancel}
        confirmLoading={loading}
      >
        {renderText()}
      </Modal>
    </React.Fragment>
  );
};

export default EnableMfaModal;
