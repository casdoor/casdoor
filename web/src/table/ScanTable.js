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
import {Col, Row, Table} from "antd";
import i18next from "i18next";
import {Link} from "react-router-dom";
import * as Setting from "../Setting";

function parseFindings(provider, scanResult) {
  if (Array.isArray(scanResult)) {
    return scanResult;
  }

  const metadata = provider?.metadata;
  if (!metadata) {
    return [];
  }

  try {
    const parsed = JSON.parse(metadata);
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function normalizeCVEs(cves) {
  return Array.isArray(cves) ? cves : [];
}

function getCveLabel(cve) {
  return cve?.code || cve?.name || "-";
}

function getCveLink(cve) {
  const references = Array.isArray(cve?.references) ? cve.references : [];
  return references.find((reference) => {
    if (typeof reference !== "string") {
      return false;
    }
    const value = reference.trim();
    return value.startsWith("http://") || value.startsWith("https://");
  }) || "";
}

function getEntryPath(subType, owner, name) {
  if (!owner || !name) {
    return "";
  }

  if (subType === "Site") {
    return `/sites/${owner}/${name}`;
  }

  if (subType === "Agent") {
    return `/agents/${owner}/${name}`;
  }

  return "";
}

export default function ScanTable({provider, options}) {
  const findings = parseFindings(provider, options.scanResult);
  const subType = options?.subType || provider?.subType;
  const owner = options?.owner || provider?.owner;

  const columns = [
    {
      title: i18next.t("general:Name"),
      dataIndex: "name",
      key: "name",
      width: 160,
      render: (text) => {
        const entryPath = getEntryPath(subType, owner, text);
        if (!entryPath) {
          return text;
        }

        return (
          <Link to={entryPath}>
            {text}
          </Link>
        );
      },
    },
    {
      title: i18next.t("general:Product"),
      dataIndex: "product",
      key: "product",
      width: 160,
    },
    {
      title: i18next.t("general:Vendor"),
      dataIndex: "vendor",
      key: "vendor",
      width: 160,
    },
    {
      title: i18next.t("general:Version"),
      dataIndex: "version",
      key: "version",
      width: 140,
    },
    {
      title: i18next.t("general:Severity"),
      dataIndex: "severity",
      key: "severity",
      width: 120,
    },
    {
      title: "CVEs",
      key: "cves",
      width: 420,
      render: (_, record) => {
        const cves = normalizeCVEs(record?.cves);
        if (cves.length === 0) {
          return "0";
        }

        return (
          <div>
            {cves.map((cve, index) => {
              const label = getCveLabel(cve);
              const link = getCveLink(cve);
              const content = (
                <React.Fragment>
                  <div>
                    {label}
                    {cve?.severity ? ` (${cve.severity})` : ""}
                  </div>
                  {cve?.summary ? <div style={{color: "#8c8c8c"}}>{cve.summary}</div> : null}
                </React.Fragment>
              );

              return (
                <div key={`${label}-${index}`} style={{marginBottom: index === cves.length - 1 ? 0 : 8}}>
                  {link ? (
                    <a target="_blank" rel="noreferrer" href={link} style={{display: "block"}}>
                      {content}
                    </a>
                  ) : content}
                </div>
              );
            })}
          </div>
        );
      },
    },
  ];

  return (
    <React.Fragment>
      {findings.length > 0 ? (
        <Row style={{marginTop: "20px"}}>
          <Col span={22} offset={(Setting.isMobile()) ? 0 : 2}>
            <Table
              scroll={{x: "max-content", y: 800}}
              dataSource={findings}
              columns={columns}
              rowKey={(record, index) => `${record?.targetUrl}-${record?.name}-${index}`}
              pagination={false}
              size="middle"
              bordered
              title={() => `${i18next.t("general:Scan")}: ${findings.length}`}
            />
          </Col>
        </Row>
      ) : null}
    </React.Fragment>
  );
}
