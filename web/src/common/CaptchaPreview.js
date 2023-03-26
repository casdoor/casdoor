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

import {Button} from "antd";
import React from "react";
import i18next from "i18next";
import {CaptchaModal} from "./modal/CaptchaModal";
import * as UserBackend from "../backend/UserBackend";

export const CaptchaPreview = (props) => {
  const {owner, name, provider, captchaType, subType, clientId, clientSecret, clientId2, clientSecret2, providerUrl} = props;
  const [visible, setVisible] = React.useState(false);

  const clickPreview = () => {
    provider.name = name;
    provider.clientId = clientId;
    provider.type = captchaType;
    provider.providerUrl = providerUrl;
    if (clientSecret !== "***") {
      provider.clientSecret = clientSecret;
      // ProviderBackend.updateProvider(owner, providerName, provider).then(() => {
      //   setOpen(true);
      // });
      setVisible(true);
    } else {
      setVisible(true);
    }
  };

  const isButtonDisabled = () => {
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

  const onOk = (captchaType, captchaToken, clientSecret) => {
    UserBackend.verifyCaptcha(captchaType, captchaToken, clientSecret).then(() => {
      setVisible(false);
    });
  };

  const onCancel = () => {
    setVisible(false);
  };

  return (
    <React.Fragment>
      <Button
        style={{fontSize: 14}}
        type={"primary"}
        onClick={clickPreview}
        disabled={isButtonDisabled()}
      >
        {i18next.t("general:Preview")}
      </Button>
      <CaptchaModal
        owner={owner}
        name={name}
        visible={visible}
        onOk={onOk}
        onCancel={onCancel}
        isCurrentProvider={true}
      />
    </React.Fragment>
  );
};
