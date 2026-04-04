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
import {Col, Input, Row} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

export function renderLogProviderFields(provider, updateProviderField) {
  if (provider.type === "Agent" && provider.subType === "OpenClaw") {
    return (
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Input value={provider.host} placeholder="e.g. 192.168.1.100" onChange={e => {
            updateProviderField("host", e.target.value);
          }} />
        </Col>
      </Row>
    );
  }
  return null;
}
