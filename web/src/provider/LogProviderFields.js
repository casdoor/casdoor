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
import {Col, Input, InputNumber, Row} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";
import i18next from "i18next";

export function renderLogProviderFields(provider, updateProviderField) {
  if (provider.type === "Casdoor Permission Log") {
    return null;
  }

  return (
    <React.Fragment>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Input prefix={<LinkOutlined />} value={provider.host} placeholder="127.0.0.1" onChange={e => {
            updateProviderField("host", e.target.value);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Port"), i18next.t("provider:Port - Tooltip"))} :
        </Col>
        <Col span={22} >
          <InputNumber value={provider.port} min={0} max={65535} onChange={value => {
            updateProviderField("port", value);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("general:Title"), i18next.t("provider:Log tag - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Input value={provider.title} placeholder="casdoor" onChange={e => {
            updateProviderField("title", e.target.value);
          }} />
        </Col>
      </Row>
    </React.Fragment>
  );
}
