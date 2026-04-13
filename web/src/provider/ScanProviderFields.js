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
import {Button, Col, Row, Select, Table} from "antd";
import * as Setting from "../Setting";
import i18next from "i18next";
import {scanColumns} from "../common/modal/ScanServerModal";

const hostOptions = [
  {label: "127.0.0.1/32", value: "127.0.0.1/32"},
  {label: "10.0.0.0/24", value: "10.0.0.0/24"},
  {label: "172.16.0.0/24", value: "172.16.0.0/24"},
  {label: "192.168.1.0/24", value: "192.168.1.0/24"},
];

const portOptions = [
  {label: "80", value: "80"},
  {label: "3000", value: "3000"},
  {label: "8080", value: "8080"},
];

const pathOptions = [
  {label: "/", value: "/"},
  {label: "/mcp", value: "/mcp"},
  {label: "/sse", value: "/sse"},
  {label: "/mcp/sse", value: "/mcp/sse"},
];

function toList(rawValue) {
  return `${rawValue || ""}`.split(",").map(item => item.trim()).filter(item => item !== "");
}

function normalizeAndJoin(values) {
  return (values || []).map(item => `${item}`.trim()).filter(item => item !== "").join(",");
}

export function renderScanProviderFields(provider, updateProviderField, options = {}) {
  if (provider.type !== "MCP Scan" || provider.subType !== "Intranet Scan") {
    return null;
  }

  const canScan = options.mode !== "add";

  return (
    <React.Fragment>
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {i18next.t("general:Host")}:
        </Col>
        <Col span={22}>
          <Select
            mode="tags"
            style={{width: "100%"}}
            value={toList(provider.scopes)}
            options={hostOptions}
            onChange={value => updateProviderField("scopes", normalizeAndJoin(value))}
          />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {i18next.t("general:Port")}:
        </Col>
        <Col span={22}>
          <Select
            mode="tags"
            style={{width: "100%"}}
            value={toList(provider?.content)}
            options={portOptions}
            onChange={value => updateProviderField("content", normalizeAndJoin(value))}
          />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
          {i18next.t("general:Path")}:
        </Col>
        <Col span={22}>
          <Select
            mode="tags"
            style={{width: "100%"}}
            value={toList(provider?.endpoint)}
            options={pathOptions}
            onChange={value => updateProviderField("endpoint", normalizeAndJoin(value))}
          />
        </Col>
      </Row>
      <Row style={{marginTop: "20px"}}>
        <Col span={22} offset={(Setting.isMobile()) ? 0 : 2}>
          <Button type="primary" loading={options.scanLoading} disabled={!canScan} onClick={options.onScan}>
            {i18next.t("server:Scan server")}
          </Button>
        </Col>
      </Row>
      {options.scanResult !== null ? (
        <Row style={{marginTop: "20px"}}>
          <Col span={22} offset={(Setting.isMobile()) ? 0 : 2}>
            <Table
              scroll={{x: "max-content", y: 320}}
              dataSource={options.scanServers || []}
              columns={scanColumns}
              rowKey={(record, index) => `${record.url}-${index}`}
              pagination={false}
              size="middle"
              bordered
              title={() => {
                const scannedHosts = i18next.t("server:Scanned hosts") + `:${options.scanResult?.scannedHosts ?? 0}`;
                const onlineHosts = i18next.t("server:Online hosts") + `:${options.scanResult?.onlineHosts?.length ?? 0}`;
                const foundServers = i18next.t("server:Found servers") + `:${options.scanServers.length}`;
                return `${scannedHosts},${onlineHosts},${foundServers}`;
              }}
            />
          </Col>
        </Row>
      ) : null}
    </React.Fragment>
  );
}
