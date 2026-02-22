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
import {Button, Col, Input, Row, Select, Switch} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import * as ProviderEditTestSms from "../common/TestSmsWidget";
import {CountryCodeSelect} from "../common/select/CountryCodeSelect";
import HttpHeaderTable from "../table/HttpHeaderTable";
import {LinkOutlined} from "@ant-design/icons";

const {Option} = Select;

const SMS_PROVIDERS_WITHOUT_SIGN_NAME = ["Custom HTTP SMS", "Twilio SMS", "Amazon SNS", "Msg91 SMS", "Infobip SMS"];
const SMS_PROVIDERS_WITHOUT_TEMPLATE_CODE = ["Infobip SMS", "Custom HTTP SMS"];

export function renderSmsProviderFields(provider, updateProviderField, renderSmsMappingInput, account) {
  return (
    <React.Fragment>
      {SMS_PROVIDERS_WITHOUT_SIGN_NAME.includes(provider.type) ?
        null :
        (<Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Sign Name"), i18next.t("provider:Sign Name - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={provider.signName} onChange={e => {
              updateProviderField("signName", e.target.value);
            }} />
          </Col>
        </Row>
        )
      }
      {SMS_PROVIDERS_WITHOUT_TEMPLATE_CODE.includes(provider.type) ?
        null :
        (<Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("provider:Template code"), i18next.t("provider:Template code - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input value={provider.templateCode} onChange={e => {
              updateProviderField("templateCode", e.target.value);
            }} />
          </Col>
        </Row>
        )
      }
      {
        provider.type === "Custom HTTP SMS" ? (
          <React.Fragment>
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
                      {id: "PUT", name: "PUT"},
                      {id: "DELETE", name: "DELETE"},
                    ].map((method, index) => <Option key={index} value={method.id}>{method.name}</Option>)
                  }
                </Select>
              </Col>
            </Row>
            {
              provider.method !== "GET" ? (<Row style={{marginTop: "20px"}} >
                <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                  {Setting.getLabel(i18next.t("webhook:Content type"), i18next.t("webhook:Content type - Tooltip"))} :
                </Col>
                <Col span={22} >
                  <Select virtual={false} style={{width: "100%"}} value={provider.issuerUrl === "" ? "application/x-www-form-urlencoded" : provider.issuerUrl} onChange={value => {
                    updateProviderField("issuerUrl", value);
                  }}>
                    {
                      [
                        {id: "application/json", name: "application/json"},
                        {id: "application/x-www-form-urlencoded", name: "application/x-www-form-urlencoded"},
                      ].map((method, index) => <Option key={index} value={method.id}>{method.name}</Option>)
                    }
                  </Select>
                </Col>
              </Row>) : null
            }
            <Row style={{marginTop: "20px"}} >
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("provider:HTTP header"), i18next.t("provider:HTTP header - Tooltip"))} :
              </Col>
              <Col span={22} >
                <HttpHeaderTable httpHeaders={provider.httpHeaders} onUpdateTable={(value) => {updateProviderField("httpHeaders", value);}} />
              </Col>
            </Row>
            {provider.method !== "GET" ? <Row style={{marginTop: "20px"}}>
              <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
                {Setting.getLabel(i18next.t("provider:HTTP body mapping"), i18next.t("provider:HTTP body mapping - Tooltip"))} :
              </Col>
              <Col span={22}>
                {renderSmsMappingInput()}
              </Col>
            </Row> : null}
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
          </React.Fragment>
        ) : null
      }
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:Enable proxy"), i18next.t("provider:Enable proxy - Tooltip"))} :
        </Col>
        <Col span={1} >
          <Switch checked={provider.enableProxy} onChange={checked => {
            updateProviderField("enableProxy", checked);
          }} />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}} >
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {Setting.getLabel(i18next.t("provider:SMS Test"), i18next.t("provider:SMS Test - Tooltip"))} :
        </Col>
        <Col span={4} >
          <Input.Group compact>
            <CountryCodeSelect
              style={{width: "90px"}}
              initValue={provider.content}
              onChange={(value) => {
                updateProviderField("content", value);
              }}
              countryCodes={account.organization.countryCodes}
            />
            <Input value={provider.receiver}
              style={{width: "150px"}}
              placeholder = {i18next.t("user:Input your phone number")}
              onChange={e => {
                updateProviderField("receiver", e.target.value);
              }} />
          </Input.Group>
        </Col>
        <Col span={2} >
          <Button style={{marginLeft: "10px", marginBottom: "5px"}} type="primary"
            disabled={!Setting.isValidPhone(provider.receiver) || (provider.type === "Custom HTTP SMS" && provider.endpoint === "")}
            onClick={() => ProviderEditTestSms.sendTestSms(provider, "+" + Setting.getCountryCode(provider.content) + provider.receiver)} >
            {i18next.t("provider:Send Testing SMS")}
          </Button>
        </Col>
      </Row>
    </React.Fragment>
  );
}
