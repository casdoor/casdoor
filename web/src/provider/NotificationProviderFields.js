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
import {Button, Col, Input, Row, Select} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as ProviderNotification from "../common/TestNotificationWidget";

const {Option} = Select;
const {TextArea} = Input;

export function renderNotificationProviderFields(provider, updateProviderField, getReceiverRow) {
  return (
    <React.Fragment>
      {["CUCloud"].includes(provider.type) ? (
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
      {["Custom HTTP"].includes(provider.type) ? (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Method"), i18next.t("provider:Method - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={provider.method} onChange={value => {
              updateProviderField("method", value);
            }}>
              {
                [
                  {id: "GET", name: "GET"},
                  {id: "POST", name: "POST"},
                ].map((method, index) => <Option key={index} value={method.id}>{method.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
      ) : null}
      {["Custom HTTP", "CUCloud"].includes(provider.type) ? (
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Parameter"), i18next.t("provider:Parameter - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={provider.title} onChange={e => {
              updateProviderField("title", e.target.value);
            }} />
          </Col>
        </Row>
      ) : null}
      {["Google Chat", "CUCloud"].includes(provider.type) ? (
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
      ) : null}
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Content"), i18next.t("provider:Content - Tooltip"))} :
        </Col>
        <Col span={22} >
          <TextArea autoSize={{minRows: 3, maxRows: 100}} value={provider.content} onChange={e => {
            updateProviderField("content", e.target.value);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        {getReceiverRow(provider)}
        <Button style={{marginLeft: "10px", marginBottom: "5px"}} type="primary"
          onClick={() => ProviderNotification.sendTestNotification(provider)} >
          {i18next.t("provider:Send Testing Notification")}
        </Button>
      </Row>
    </React.Fragment>
  );
}
