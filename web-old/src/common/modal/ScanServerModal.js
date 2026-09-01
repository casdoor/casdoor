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

export const scanColumns = [
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
];

const ScanServerModal = (props) => {
  const scanColumnsWithAction = [...scanColumns, {
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
  }];

  return (
    <Modal
      title={i18next.t("server:Scan server")}
      open={props.open}
      width={960}
      confirmLoading={props.loading}
      onOk={props.onSubmit}
      onCancel={props.onCancel}
      okText={i18next.t("general:Sync")}
    >
      <Select
        style={{width: "100%"}}
        value={props.selectedScanProvider}
        onChange={props.onChangeSelectedProvider}
        options={props.scanProviders.map(provider => ({
          label: `${provider.displayName || provider.name}`,
          value: `${provider.owner}/${provider.name}`,
        }))}
      />

      {props.scanResult !== null ? (
        <Table
          style={{marginTop: "16px"}}
          scroll={{x: "max-content", y: 320}}
          dataSource={props.scanServers}
          columns={scanColumnsWithAction}
          rowKey={(record, index) => `${record.url}-${index}`}
          pagination={false}
          size="middle"
          bordered
          title={() => {
            const scannedHosts = i18next.t("server:Scanned hosts") + `:${props.scanResult?.scannedHosts ?? 0}`;
            const onlineHosts = i18next.t("server:Online hosts") + `:${props.scanResult?.onlineHosts?.length ?? 0}`;
            const foundServers = i18next.t("server:Found servers") + `:${props.scanServers.length}`;
            return `${scannedHosts},${onlineHosts},${foundServers}`;
          }}
        />
      ) : null}
    </Modal>
  );
};

export default ScanServerModal;
