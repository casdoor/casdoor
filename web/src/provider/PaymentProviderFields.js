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
import {LinkOutlined} from "@ant-design/icons";

const {Option} = Select;

export function renderPaymentProviderFields(provider, updateProviderField, certs) {
  return (
    <React.Fragment>
      {
        (provider.type === "Alipay" || provider.type === "WeChat Pay" || provider.type === "Casdoor") ? (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Cert"), i18next.t("general:Cert - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Select virtual={false} style={{width: "100%"}} value={provider.cert} onChange={(value => {updateProviderField("cert", value);})}>
                {
                  certs.map((cert, index) => <Option key={index} value={cert.name}>{cert.name}</Option>)
                }
              </Select>
            </Col>
          </Row>
        ) : null
      }
      {
        (provider.type === "Alipay") ? (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {Setting.getLabel(i18next.t("general:Root cert"), i18next.t("general:Root cert - Tooltip"))} :
            </Col>
            <Col span={22} >
              <Select virtual={false} style={{width: "100%"}} value={provider.metadata} onChange={(value => {updateProviderField("metadata", value);})}>
                {
                  certs.map((cert, index) => <Option key={index} value={cert.name}>{cert.name}</Option>)
                }
              </Select>
            </Col>
          </Row>
        ) : null
      }
      {(provider.type === "GC" || provider.type === "FastSpring") ? (
        <Row style={{marginTop: "20px"}}>
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Host"), i18next.t("provider:Host - Tooltip"))} :
          </Col>
          <Col span={22}>
            <Input prefix={<LinkOutlined />} value={provider.host} onChange={e => {
              updateProviderField("host", e.target.value);
            }} />
          </Col>
        </Row>
      ) : null}
    </React.Fragment>
  );
}
