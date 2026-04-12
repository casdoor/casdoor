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
import {Col, Descriptions, Row, Tag} from "antd";
import * as Setting from "./Setting";
import i18next from "i18next";

class SELinuxEntryViewer extends React.Component {
  getLabelSpan() {
    return this.props.labelSpan ?? (Setting.isMobile() ? 22 : 2);
  }

  getContentSpan() {
    return this.props.contentSpan ?? 22;
  }

  getMessage() {
    return `${this.props.entry?.message ?? ""}`.trim();
  }

  getSeverityColor(severity) {
    switch ((severity || "").toLowerCase()) {
    case "warning":
      return "orange";
    case "error":
      return "red";
    case "info":
      return "blue";
    default:
      return "default";
    }
  }

  escapeRegExp(value) {
    return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }

  extractValue(message, key) {
    const escapedKey = this.escapeRegExp(key);
    const quotedMatch = message.match(new RegExp(`(?:^|\\s)${escapedKey}="([^"]*)"`, "i"));
    if (quotedMatch) {
      return quotedMatch[1];
    }

    const plainMatch = message.match(new RegExp(`(?:^|\\s)${escapedKey}=([^\\s]+)`, "i"));
    return plainMatch ? plainMatch[1] : "";
  }

  parseMessage() {
    const message = this.getMessage();
    const severityMatch = message.match(/^\[([^\]]+)\]\s*/);
    const severity = severityMatch ? severityMatch[1] : "";
    const body = severityMatch ? message.slice(severityMatch[0].length) : message;

    const details = {
      severity,
      auditType: this.extractValue(body, "type"),
      auditStamp: (body.match(/msg=audit\(([^)]+)\)/) || [])[1] || "",
      decision: (body.match(/\bavc:\s+([a-z_]+)/i) || [])[1] || "",
      permission: (body.match(/\{\s*([^}]+?)\s*\}/) || [])[1] || "",
      pid: this.extractValue(body, "pid"),
      command: this.extractValue(body, "comm"),
      executable: this.extractValue(body, "exe"),
      path: this.extractValue(body, "path"),
      device: this.extractValue(body, "dev"),
      inode: this.extractValue(body, "ino"),
      sourceContext: this.extractValue(body, "scontext"),
      targetContext: this.extractValue(body, "tcontext"),
      targetClass: this.extractValue(body, "tclass"),
      permissive: this.extractValue(body, "permissive"),
      rawBody: body,
    };

    return details;
  }

  renderValue(value, render) {
    if (!value) {
      return "-";
    }

    return render ? render(value) : value;
  }

  render() {
    const details = this.parseMessage();

    return (
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={this.getLabelSpan()}>
          {i18next.t("entry:SELinux event")}:
        </Col>
        <Col span={this.getContentSpan()}>
          <Descriptions
            bordered
            size="small"
            column={Setting.isMobile() ? 1 : 2}
            layout={Setting.isMobile() ? "vertical" : "horizontal"}
          >
            <Descriptions.Item label={i18next.t("general:Severity")}>
              {this.renderValue(details.severity, value => <Tag color={this.getSeverityColor(value)}>{value}</Tag>)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Type")}>
              {this.renderValue(details.auditType)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Decision")}>
              {this.renderValue(details.decision)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Permission")}>
              {this.renderValue(details.permission)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Audit stamp")}>
              {this.renderValue(details.auditStamp)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Permissive")}>
              {this.renderValue(details.permissive)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Process ID")}>
              {this.renderValue(details.pid)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Command")}>
              {this.renderValue(details.command)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Executable")}>
              {this.renderValue(details.executable)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Target class")}>
              {this.renderValue(details.targetClass)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Path")}>
              {this.renderValue(details.path)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Device")}>
              {this.renderValue(details.device)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Inode")}>
              {this.renderValue(details.inode)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Source context")} span={Setting.isMobile() ? 1 : 2}>
              {this.renderValue(details.sourceContext)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Target context")} span={Setting.isMobile() ? 1 : 2}>
              {this.renderValue(details.targetContext)}
            </Descriptions.Item>
          </Descriptions>
        </Col>
      </Row>
    );
  }
}

export default SELinuxEntryViewer;
