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
import {CaptchaModal} from "./CaptchaModal";

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
  const captchaModal = React.useRef(null);

  const clickPreview = () => {
    captchaModal.current.showCaptcha();
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
      {
        <CaptchaModal
          provider={provider}
          providerName={providerName}
          clientSecret={clientSecret}
          captchaType={captchaType}
          subType={subType}
          owner={owner}
          clientId={clientId}
          name={name}
          providerUrl={providerUrl}
          clientId2={clientId2}
          clientSecret2={clientSecret2}
          preview={true}
          ref={captchaModal}
        />
      }
    </React.Fragment>
  );
};
