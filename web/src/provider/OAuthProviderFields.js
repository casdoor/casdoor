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
import {Col, Input, Radio, Row, Switch} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";
import i18next from "i18next";

const {TextArea} = Input;

export function renderOAuthProviderFields(provider, updateProviderField, renderUserMappingInput) {
  const getDomainLabel = provider => {
    switch (provider.category) {
    case "OAuth":
      if (provider.type === "AzureAD" || provider.type === "AzureADB2C") {
        return Setting.getLabel(i18next.t("provider:Tenant ID"), i18next.t("provider:Tenant ID - Tooltip"));
      } else {
        return Setting.getLabel(i18next.t("provider:Domain"), i18next.t("provider:Domain - Tooltip"));
      }
    default:
      return Setting.getLabel(i18next.t("provider:Domain"), i18next.t("provider:Domain - Tooltip"));
    }
  };

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
        provider.type !== "WeChat" ? null : (
          <React.Fragment>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("provider:Use WeChat Media Platform in PC"), i18next.t("provider:Use WeChat Media Platform in PC - Tooltip"))} :
              </Col>
              <Col span={1} >
                <Switch disabled={!provider.clientId} checked={provider.disableSsl} onChange={checked => {
                  updateProviderField("disableSsl", checked);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("token:Access token"), i18next.t("token:Access token - Tooltip"))} :
              </Col>
              <Col span={22} >
                <Input value={provider.content} disabled={!provider.disableSsl || !provider.clientId2} onChange={e => {
                  updateProviderField("content", e.target.value);
                }} />
              </Col>
            </Row>
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("provider:Follow-up action"), i18next.t("provider:Follow-up action - Tooltip"))} :
              </Col>
              <Col>
                <Radio.Group value={provider.signName}
                  disabled={!provider.disableSsl || !provider.clientId || !provider.clientId2}
                  buttonStyle="solid"
                  onChange={e => {
                    updateProviderField("signName", e.target.value);
                  }}>
                  <Radio.Button value="open">{i18next.t("provider:Use WeChat Open Platform to login")}</Radio.Button>
                  <Radio.Button value="media">{i18next.t("provider:Use WeChat Media Platform to login")}</Radio.Button>
                </Radio.Group>
              </Col>
            </Row>
          </React.Fragment>
        )
      }
      {
        provider.type !== "ADFS" && provider.type !== "AzureAD"
        && provider.type !== "AzureADB2C" && (provider.type !== "Casdoor" && provider.category !== "Storage")
        && provider.type !== "Okta" && provider.type !== "Nextcloud" ? null : (
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={2}>
                {getDomainLabel(provider)} :
              </Col>
              <Col span={22} >
                <Input prefix={<LinkOutlined />} value={provider.domain} onChange={e => {
                  updateProviderField("domain", e.target.value);
                }} />
              </Col>
            </Row>
          )
      }
      {
        provider.type !== "Google" && provider.type !== "Lark" ? null : (
          <Row style={{marginTop: "20px"}} >
            <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
              {provider.type === "Google" ?
                Setting.getLabel(i18next.t("provider:Get phone number"), i18next.t("provider:Get phone number - Tooltip"))
                : Setting.getLabel(i18next.t("provider:Use global endpoint"), i18next.t("provider:Use global endpoint - Tooltip"))} :
            </Col>
            <Col span={1} >
              <Switch disabled={!provider.clientId} checked={provider.disableSsl} onChange={checked => {
                updateProviderField("disableSsl", checked);
              }} />
            </Col>
          </Row>
        )
      }
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
