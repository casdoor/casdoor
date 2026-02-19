// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

import React from "react";
import {Col, Row} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import {CaptchaPreview} from "../common/CaptchaPreview";

export function renderCaptchaProviderFields(provider, providerName) {
  return (
    <Row style={{marginTop: "20px"}} >
      <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
        {Setting.getLabel(i18next.t("general:Preview"), i18next.t("general:Preview - Tooltip"))} :
      </Col>
      <Col span={22} >
        <CaptchaPreview
          owner={provider.owner}
          name={provider.name}
          provider={provider}
          providerName={providerName}
          captchaType={provider.type}
          subType={provider.subType}
          clientId={provider.clientId}
          clientSecret={provider.clientSecret}
          clientId2={provider.clientId2}
          clientSecret2={provider.clientSecret2}
          providerUrl={provider.providerUrl}
        />
      </Col>
    </Row>
  );
}
