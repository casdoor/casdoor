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
import {Col, Input, Row, Select} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";

const {Option} = Select;

export function renderLogProviderFields(provider, updateProviderField) {
  return (
    <React.Fragment>
      {provider.type === "Agent" && provider.subType === "OpenClaw" ? (
        <React.Fragment>
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={provider.host} onChange={e => {
                updateProviderField("host", e.target.value);
              }} />
            </Col>
          </Row>
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("provider:Agent ID"), i18next.t("provider:Agent ID - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Input value={provider.title} onChange={e => {
                updateProviderField("title", e.target.value);
              }} />
            </Col>
          </Row>
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {i18next.t("general:Path")} :
            </Col>
            <Col span={22} >
              <Input value={provider.endpoint} onChange={e => {
                updateProviderField("endpoint", e.target.value);
              }} />
            </Col>
          </Row>
        </React.Fragment>
      ) : null}
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("general:State"), i18next.t("general:State - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Select virtual={false} style={{width: "100%"}} value={provider.state || "Enabled"} onChange={value => {
            updateProviderField("state", value);
          }}>
            <Option value="Enabled">{i18next.t("general:Enabled")}</Option>
            <Option value="Disabled">{i18next.t("general:Disabled")}</Option>
          </Select>
        </Col>
      </Row>
    </React.Fragment>
  );
}
