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
import {Button, Modal, Select, Table} from "antd";
import i18next from "i18next";
import * as Setting from "../../Setting";

const scanCidrOptions = [
  {label: "127.0.0.1/32", value: "127.0.0.1/32"},
  {label: "10.0.0.0/24", value: "10.0.0.0/24"},
  {label: "172.16.0.0/24", value: "172.16.0.0/24"},
  {label: "192.168.1.0/24", value: "192.168.1.0/24"},
];

const scanPortOptions = [
  {label: "1-65535", value: "1-65535"},
  {label: "80", value: "80"},
  {label: "443", value: "443"},
  {label: "3000", value: "3000"},
  {label: "8080", value: "8080"},
];

const scanPathOptions = [
  {label: "/", value: "/"},
  {label: "/mcp", value: "/mcp"},
  {label: "/sse", value: "/sse"},
  {label: "/mcp/sse", value: "/mcp/sse"},
];

const ScanServerModal = (props) => {
  const scanColumns = [
    {
      title: i18next.t("general:Host"),
      dataIndex: "host",
      key: "host",
      width: "140px",
    },
    {
      title: i18next.t("general:Port"),
      dataIndex: "port",
      key: "port",
      width: "90px",
    },
    {
      title: i18next.t("general:Path"),
      dataIndex: "path",
      key: "path",
      width: "120px",
    },
    {
      title: i18next.t("general:URL"),
      dataIndex: "url",
      key: "url",
      render: (text) => {
        if (!text) {
          return null;
        }

        return (
          <a target="_blank" rel="noreferrer" href={text}>
            {Setting.getShortText(text, 60)}
          </a>
        );
      },
    },
    {
      title: i18next.t("general:Action"),
      dataIndex: "scanOp",
      key: "scanOp",
      width: "120px",
      render: (_, record) => {
        return (
          <Button size="small" type="primary" onClick={() => props.onAddScannedServer(record)}>
            {i18next.t("general:Add")}
          </Button>
        );
      },
    },
  ];

  return (
    <Modal
      title="Scan server"
      open={props.open}
      width={960}
      confirmLoading={props.loading}
      onOk={props.onSubmit}
      onCancel={props.onCancel}
      okText={i18next.t("general:Sync")}
    >
      <div style={{marginBottom: "12px"}}>IP range</div>
      <Select
        mode="tags"
        style={{width: "100%"}}
        value={props.scanFilters.cidrs}
        options={scanCidrOptions}
        onChange={(value) => props.onChangeScanFilters({cidrs: value})}
        placeholder="Select or input CIDR/IP"
      />

      <div style={{marginTop: "16px", marginBottom: "12px"}}>Ports</div>
      <Select
        mode="tags"
        style={{width: "100%"}}
        value={props.scanFilters.ports}
        options={scanPortOptions}
        onChange={(value) => props.onChangeScanFilters({ports: value})}
        placeholder="Select or input ports"
      />

      <div style={{marginTop: "16px", marginBottom: "12px"}}>Paths</div>
      <Select
        mode="tags"
        style={{width: "100%"}}
        value={props.scanFilters.paths}
        options={scanPathOptions}
        onChange={(value) => props.onChangeScanFilters({paths: value})}
        placeholder="Select or input paths"
      />

      {props.scanResult !== null ? (
        <Table
          style={{marginTop: "16px"}}
          scroll={{x: "max-content", y: 320}}
          dataSource={props.scanServers}
          columns={scanColumns}
          rowKey={(record, index) => `${record.url}-${index}`}
          pagination={false}
          size="middle"
          bordered
          title={() => {
            return `Scanned hosts: ${props.scanResult?.scannedHosts ?? 0}, online hosts: ${props.scanResult?.onlineHosts?.length ?? 0}, found servers: ${props.scanServers.length}`;
          }}
        />
      ) : null}
    </Modal>
  );
};

export default ScanServerModal;
