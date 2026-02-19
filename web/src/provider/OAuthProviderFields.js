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

import React from "react";
import {Col, Input, Row, Switch} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";
import i18next from "i18next";

const {TextArea} = Input;

export function renderOAuthProviderFields(provider, updateProviderField, renderUserMappingInput) {
  return (
    <React.Fragment>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Email regex"), i18next.t("provider:Email regex - Tooltip"))} :
        </Col>
        <Col span={22}>
          <TextArea rows={4} value={provider.emailRegex} onChange={e => {
            updateProviderField("emailRegex", e.target.value);
          }} />
        </Col>
      </Row>
      {
        provider.type.startsWith("Custom") ? (
          <React.Fragment>
            <Col>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Auth URL"), i18next.t("provider:Auth URL - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={provider.customAuthUrl} onChange={e => {
                    updateProviderField("customAuthUrl", e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Token URL"), i18next.t("provider:Token URL - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={provider.customTokenUrl} onChange={e => {
                    updateProviderField("customTokenUrl", e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Scope"), i18next.t("provider:Scope - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={provider.scopes} onChange={e => {
                    updateProviderField("scopes", e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:UserInfo URL"), i18next.t("provider:UserInfo URL - Tooltip"))}
                </Col>
                <Col span={22} >
                  <Input value={provider.customUserInfoUrl} onChange={e => {
                    updateProviderField("customUserInfoUrl", e.target.value);
                  }} />
                </Col>
              </Row>
              <Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("provider:Enable PKCE"), i18next.t("provider:Enable PKCE - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Switch checked={provider.enablePkce} onChange={checked => {
                    updateProviderField("enablePkce", checked);
                  }} />
                </Col>
              </Row>
            </Col>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("provider:User mapping"), i18next.t("provider:User mapping - Tooltip"))} :
              </Col>
              <Col span={22} >
                {renderUserMappingInput()}
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("general:Favicon"), i18next.t("general:Favicon - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Row style={{marginTop: "20px"}} >
                  <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                    {Setting.getLabel(i18next.t("general:URL"), i18next.t("general:URL - Tooltip"))} :
                  </Col>
                  <Col span={23} >
                    <Input prefix={<LinkOutlined />} value={provider.customLogo} onChange={e => {
                      updateProviderField("customLogo", e.target.value);
                    }} />
                  </Col>
                </Row>
                <Row style={{marginTop: "20px"}} >
                  <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 1}>
                    {i18next.t("general:Preview")}:
                  </Col>
                  <Col span={23} >
                    <a target="_blank" rel="noreferrer" href={provider.customLogo}>
                      <img src={provider.customLogo} alt={provider.customLogo} height={90} style={{marginBottom: "20px"}} />
                    </a>
                  </Col>
                </Row>
              </Col>
            </Row>
          </React.Fragment>
        ) : null
      }
    </React.Fragment>
  );
}
