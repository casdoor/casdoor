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
import {Button, Col, Input, Row, Switch} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import {authConfig} from "../auth/Auth";
import copy from "copy-to-clipboard";

const {TextArea} = Input;

export function renderSamlProviderFields(provider, updateProviderField, metadataConfig) {
  const {requestUrl, setRequestUrl, metadataLoading, fetchSamlMetadata, parseSamlMetadata} = metadataConfig;
  return (
    <React.Fragment>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Sign request"), i18next.t("provider:Sign request - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Switch checked={provider.enableSignAuthnRequest} onChange={checked => {
            updateProviderField("enableSignAuthnRequest", checked);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Metadata url"), i18next.t("provider:Metadata url - Tooltip"))} :
        </Col>
        <Col span={6} >
          <Input value={requestUrl} onChange={e => {
            setRequestUrl(e.target.value);
          }} />
        </Col>
        <Col span={16} >
          <Button style={{marginLeft: "10px"}} type="primary" loading={metadataLoading} onClick={() => {fetchSamlMetadata();}}>{i18next.t("general:Request")}</Button>
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Metadata"), i18next.t("provider:Metadata - Tooltip"))} :
        </Col>
        <Col span={22}>
          <TextArea rows={4} value={provider.metadata} onChange={e => {
            updateProviderField("metadata", e.target.value);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={2} />
        <Col span={2}>
          <Button type="primary" onClick={() => {parseSamlMetadata();}}>
            {i18next.t("provider:Parse")}
          </Button>
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Endpoint"), i18next.t("provider:SAML 2.0 Endpoint (HTTP)"))} :
        </Col>
        <Col span={22} >
          <Input value={provider.endpoint} onChange={e => {
            updateProviderField("endpoint", e.target.value);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:IdP"), i18next.t("provider:IdP certificate"))} :
        </Col>
        <Col span={22} >
          <Input value={provider.idP} onChange={e => {
            updateProviderField("idP", e.target.value);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Issuer URL"), i18next.t("provider:Issuer URL - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Input value={provider.issuerUrl} onChange={e => {
            updateProviderField("issuerUrl", e.target.value);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:SP ACS URL"), i18next.t("provider:SP ACS URL - Tooltip"))} :
        </Col>
        <Col span={21} >
          <Input value={`${authConfig.serverUrl}/api/acs`} readOnly="readonly" />
        </Col>
        <Col span={1}>
          <Button type="primary" onClick={() => {
            copy(`${authConfig.serverUrl}/api/acs`);
            Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
          }}>
            {i18next.t("general:Copy")}
          </Button>
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:SP Entity ID"), i18next.t("provider:SP Entity ID - Tooltip"))} :
        </Col>
        <Col span={21} >
          <Input value={`${authConfig.serverUrl}/api/acs`} readOnly="readonly" />
        </Col>
        <Col span={1}>
          <Button type="primary" onClick={() => {
            copy(`${authConfig.serverUrl}/api/acs`);
            Setting.showMessage("success", i18next.t("general:Copied to clipboard successfully"));
          }}>
            {i18next.t("general:Copy")}
          </Button>
        </Col>
      </Row>
    </React.Fragment>
  );
}
