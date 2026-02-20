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
import {LinkOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";
import i18next from "i18next";

export function renderStorageProviderFields(provider, updateProviderField) {
  return (
    <React.Fragment>
      {["Local File System", "MinIO", "Tencent Cloud COS", "Google Cloud Storage", "Qiniu Cloud Kodo", "Synology", "Casdoor"].includes(provider.type) ? null : (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {Setting.getLabel(i18next.t("provider:Endpoint (Intranet)"), i18next.t("provider:Region endpoint for Intranet"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={provider.intranetEndpoint} onChange={e => {
              updateProviderField("intranetEndpoint", e.target.value);
            }} />
          </Col>
        </Row>
      )}
      {["Local File System"].includes(provider.type) ? null : (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {Setting.getLabel(i18next.t("provider:Endpoint"), i18next.t("provider:Region endpoint for Internet"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={provider.endpoint} onChange={e => {
              updateProviderField("endpoint", e.target.value);
            }} />
          </Col>
        </Row>
      )}
      {["Local File System"].includes(provider.type) ? null : (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {["Casdoor"].includes(provider.type) ?
              Setting.getLabel(i18next.t("general:Provider"), i18next.t("general:Provider - Tooltip"))
              : Setting.getLabel(i18next.t("provider:Bucket"), i18next.t("provider:Bucket - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={provider.bucket} onChange={e => {
              updateProviderField("bucket", e.target.value);
            }} />
          </Col>
        </Row>
      )}
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={2}>
          {Setting.getLabel(i18next.t("provider:Path prefix"), i18next.t("provider:Path prefix - Tooltip"))} :
        </Col>
        <Col span={22} >
          <Input value={provider.pathPrefix} onChange={e => {
            updateProviderField("pathPrefix", e.target.value);
          }} />
        </Col>
      </Row>
      {["Synology", "Casdoor"].includes(provider.type) ? null : (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {Setting.getLabel(i18next.t("provider:Domain"), i18next.t("provider:Domain - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={provider.domain} disabled={provider.type === "Local File System"} onChange={e => {
              updateProviderField("domain", e.target.value);
            }} />
          </Col>
        </Row>
      )}
      {["Casdoor"].includes(provider.type) ? (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={provider.content} onChange={e => {
              updateProviderField("content", e.target.value);
            }} />
          </Col>
        </Row>
      ) : null}
      {["AWS S3", "Tencent Cloud COS", "Qiniu Cloud Kodo", "Casdoor", "CUCloud OSS", "MinIO"].includes(provider.type) ? (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={2}>
            {["Casdoor"].includes(provider.type) ?
              Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip")) :
              Setting.getLabel(i18next.t("provider:Region ID"), i18next.t("provider:Region ID - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={provider.regionId} onChange={e => {
              updateProviderField("regionId", e.target.value);
            }} />
          </Col>
        </Row>
      ) : null}
    </React.Fragment>
  );
}
