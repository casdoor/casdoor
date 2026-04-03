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
import {Alert, Button, Card, Col, Descriptions, Drawer, Input, Row, Select, Table} from "antd";
import {LinkOutlined} from "@ant-design/icons";
import * as EntryBackend from "./backend/EntryBackend";
import * as Setting from "./Setting";
import i18next from "i18next";
import * as OrganizationBackend from "./backend/OrganizationBackend";
import * as ApplicationBackend from "./backend/ApplicationBackend";
import Editor from "./common/Editor";

const {Option} = Select;
class EntryEditPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      classes: props,
      entryName: props.match.params.entryName,
      owner: props.match.params.organizationName,
      entry: null,
      organizations: [],
      applications: [],
      mode: props.location.mode !== undefined ? props.location.mode : "edit",
      traceSpanDrawerVisible: false,
      selectedTraceSpan: null,
    };
  }

  UNSAFE_componentWillMount() {
    this.getEntry();
    this.getOrganizations();
    this.getApplications(this.state.owner);
  }

  getEntry() {
    EntryBackend.getEntry(this.state.entry?.owner || this.state.owner, this.state.entryName)
      .then((res) => {
        if (res.data === null) {
          this.props.history.push("/404");
          return;
        }

        if (res.status === "ok") {
          this.setState({
            entry: res.data,
          });
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to get")}: ${res.msg}`);
        }
      });
  }

  getOrganizations() {
    if (Setting.isAdminUser(this.props.account)) {
      OrganizationBackend.getOrganizations("admin")
        .then((res) => {
          this.setState({
            organizations: res.data || [],
          });
        });
    }
  }

  getApplications(owner) {
    ApplicationBackend.getApplicationsByOrganization("admin", owner)
      .then((res) => {
        this.setState({
          applications: res.data || [],
        });
      });
  }

  updateEntryField(key, value) {
    const entry = this.state.entry;
    if (key === "owner" && entry.owner !== value) {
      entry.application = "";
      this.getApplications(value);
    }

    entry[key] = value;
    this.setState({
      entry: entry,
    });
  }

  submitEntryEdit(willExit) {
    const entry = Setting.deepCopy(this.state.entry);
    EntryBackend.updateEntry(this.state.owner, this.state.entryName, entry)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully modified"));
          if (willExit) {
            this.props.history.push("/entries");
          } else {
            this.setState({
              mode: "edit",
              owner: entry.owner,
              entryName: entry.name,
            }, () => {this.getEntry();});
            this.props.history.push(`/entries/${entry.owner}/${entry.name}`);
          }
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to update")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  deleteEntry() {
    EntryBackend.deleteEntry(this.state.entry)
      .then((res) => {
        if (res.status === "ok") {
          Setting.showMessage("success", i18next.t("general:Successfully deleted"));
          this.props.history.push("/entries");
        } else {
          Setting.showMessage("error", `${i18next.t("general:Failed to delete")}: ${res.msg}`);
        }
      })
      .catch(error => {
        Setting.showMessage("error", `${i18next.t("general:Failed to connect to server")}: ${error}`);
      });
  }

  getEditorMaxWidth() {
    return Setting.isMobile() ? window.innerWidth - 60 : 560;
  }

  formatJsonValue(value) {
    if (value === undefined || value === null || value === "") {
      return "";
    }

    if (typeof value === "string") {
      try {
        return JSON.stringify(JSON.parse(value), null, 2);
      } catch (e) {
        return value;
      }
    }

    return JSON.stringify(value, null, 2);
  }

  formatAnyValue(value) {
    if (value === undefined || value === null) {
      return "";
    }

    if (value.stringValue !== undefined) {
      return value.stringValue;
    }

    if (value.boolValue !== undefined) {
      return `${value.boolValue}`;
    }

    if (value.intValue !== undefined) {
      return `${value.intValue}`;
    }

    if (value.doubleValue !== undefined) {
      return `${value.doubleValue}`;
    }

    if (value.bytesValue !== undefined) {
      return value.bytesValue;
    }

    if (Array.isArray(value.arrayValue?.values)) {
      return value.arrayValue.values.map(item => this.formatAnyValue(item)).join(", ");
    }

    if (Array.isArray(value.kvlistValue?.values)) {
      return value.kvlistValue.values.map(item => `${item?.key || "-"}=${this.formatAnyValue(item?.value)}`).join(", ");
    }

    return this.formatJsonValue(value);
  }

  getAnyValueType(value) {
    if (value === undefined || value === null) {
      return "-";
    }

    if (value.stringValue !== undefined) {
      return "string";
    }

    if (value.boolValue !== undefined) {
      return "bool";
    }

    if (value.intValue !== undefined) {
      return "int";
    }

    if (value.doubleValue !== undefined) {
      return "double";
    }

    if (value.bytesValue !== undefined) {
      return "bytes";
    }

    if (Array.isArray(value.arrayValue?.values)) {
      return "array";
    }

    if (Array.isArray(value.kvlistValue?.values)) {
      return "map";
    }

    return "unknown";
  }

  getAttributeValue(attributes, key) {
    const attribute = attributes.find(item => item?.key === key);
    return attribute ? this.formatAnyValue(attribute.value) : "";
  }

  renderTraceAttributeTable(attributes) {
    const rows = Array.isArray(attributes) ? attributes.map((attribute, index) => ({
      key: `${attribute?.key || "attribute"}-${index}`,
      name: attribute?.key || "-",
      type: this.getAnyValueType(attribute?.value),
      value: this.formatAnyValue(attribute?.value) || "-",
    })) : [];

    if (rows.length === 0) {
      return "-";
    }

    const columns = [
      {
        title: i18next.t("user:Keys"),
        dataIndex: "name",
        key: "name",
        width: 220,
      },
      {
        title: i18next.t("general:Type"),
        dataIndex: "type",
        key: "type",
        width: 120,
      },
      {
        title: i18next.t("user:Values"),
        dataIndex: "value",
        key: "value",
        render: value => (
          <div style={{whiteSpace: "pre-wrap", wordBreak: "break-word"}}>
            {value}
          </div>
        ),
      },
    ];

    return (
      <Table
        scroll={{x: "max-content"}}
        size="small"
        bordered
        columns={columns}
        dataSource={rows}
        rowKey="key"
        pagination={false}
      />
    );
  }

  normalizeIntegerString(value) {
    const text = `${value ?? ""}`.trim();
    if (!/^\d+$/.test(text)) {
      return "";
    }

    return text.replace(/^0+(?=\d)/, "");
  }

  subtractIntegerStrings(minuend, subtrahend) {
    const left = this.normalizeIntegerString(minuend);
    const right = this.normalizeIntegerString(subtrahend);
    if (!left || !right) {
      return "";
    }

    if (left.length < right.length || (left.length === right.length && left < right)) {
      return "";
    }

    let borrow = 0;
    let result = "";

    for (let i = 0; i < left.length; i++) {
      const leftDigit = Number(left[left.length - 1 - i]);
      const rightDigit = Number(right[right.length - 1 - i] || 0);
      let digit = leftDigit - borrow - rightDigit;
      if (digit < 0) {
        digit += 10;
        borrow = 1;
      } else {
        borrow = 0;
      }

      result = `${digit}${result}`;
    }

    return result.replace(/^0+(?=\d)/, "");
  }

  getTraceData() {
    if (this.state.entry?.type !== "trace") {
      return {spans: [], error: ""};
    }

    const message = this.state.entry?.message?.trim();
    if (!message) {
      return {spans: [], error: ""};
    }

    try {
      const trace = JSON.parse(message);
      return {
        spans: this.flattenTraceSpans(trace),
        error: "",
      };
    } catch (e) {
      return {
        spans: [],
        error: e.message,
      };
    }
  }

  flattenTraceSpans(trace) {
    const spans = [];
    const resourceSpans = Array.isArray(trace?.resourceSpans) ? trace.resourceSpans : [];

    resourceSpans.forEach((resourceSpan, resourceIndex) => {
      const resource = resourceSpan?.resource ?? {};
      const resourceAttributes = Array.isArray(resource.attributes) ? resource.attributes : [];
      const serviceName = this.getAttributeValue(resourceAttributes, "service.name");
      const scopeSpans = Array.isArray(resourceSpan?.scopeSpans) ? resourceSpan.scopeSpans : [];

      scopeSpans.forEach((scopeSpan, scopeIndex) => {
        const scope = scopeSpan?.scope ?? {};
        const scopeSchemaUrl = scopeSpan?.schemaUrl ?? "";
        const innerSpans = Array.isArray(scopeSpan?.spans) ? scopeSpan.spans : [];

        innerSpans.forEach((span, spanIndex) => {
          spans.push({
            key: `${resourceIndex}-${scopeIndex}-${spanIndex}-${span?.spanId ?? span?.name ?? "span"}`,
            resource,
            resourceAttributes,
            resourceSchemaUrl: resourceSpan?.schemaUrl ?? "",
            scope,
            scopeSchemaUrl,
            serviceName,
            span,
          });
        });
      });
    });

    return spans;
  }

  formatTraceTimestamp(unixNano) {
    if (!unixNano) {
      return "-";
    }

    const normalized = this.normalizeIntegerString(unixNano);
    if (!normalized) {
      return `${unixNano}`;
    }

    const padded = normalized.padStart(9, "0");
    const milliseconds = Number(padded.slice(0, -6) || "0");
    const nanoseconds = padded.slice(-9);
    const date = new Date(milliseconds);
    if (!Number.isFinite(milliseconds) || Number.isNaN(date.getTime())) {
      return `${unixNano}`;
    }

    return `${Setting.getFormattedDate(date.toISOString())}.${nanoseconds}`;
  }

  getSpanDuration(span) {
    if (!span?.startTimeUnixNano || !span?.endTimeUnixNano) {
      return "-";
    }

    const duration = this.subtractIntegerStrings(span.endTimeUnixNano, span.startTimeUnixNano);
    if (!duration) {
      return "-";
    }

    const durationNumber = Number(duration);
    if (!Number.isFinite(durationNumber)) {
      return `${duration} ns`;
    }

    if (durationNumber >= 1e9) {
      return `${(durationNumber / 1e9).toFixed(3)} s`;
    }

    if (durationNumber >= 1e6) {
      return `${(durationNumber / 1e6).toFixed(3)} ms`;
    }

    if (durationNumber >= 1e3) {
      return `${(durationNumber / 1e3).toFixed(3)} us`;
    }

    return `${durationNumber} ns`;
  }

  getSpanStatus(span) {
    const code = span?.status?.code ?? "";
    const message = span?.status?.message ?? "";

    if (code && message) {
      return `${code}: ${message}`;
    }

    return code || message || "-";
  }

  getScopeName(scope) {
    if (!scope?.name) {
      return "-";
    }

    return scope.version ? `${scope.name}@${scope.version}` : scope.name;
  }

  openTraceSpanDrawer(traceSpan) {
    this.setState({
      traceSpanDrawerVisible: true,
      selectedTraceSpan: traceSpan,
    });
  }

  closeTraceSpanDrawer = () => {
    this.setState({
      traceSpanDrawerVisible: false,
      selectedTraceSpan: null,
    });
  };

  renderJsonEditor(value) {
    const formattedValue = this.formatJsonValue(value);
    if (!formattedValue) {
      return "-";
    }

    return (
      <Editor
        value={formattedValue}
        lang="json"
        fillHeight
        fillWidth
        maxWidth={this.getEditorMaxWidth()}
        dark
        readOnly
      />
    );
  }

  renderMessageEditor() {
    return (
      <Editor
        value={this.formatJsonValue(this.state.entry?.message) || ""}
        lang="json"
        readOnly
      />
    );
  }

  renderTraceSpans() {
    if (this.state.entry?.type !== "trace") {
      return null;
    }

    const {spans, error} = this.getTraceData();
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: ["span", "name"],
        key: "name",
        width: 220,
        render: (text, record) => (
          <Button type="link" style={{padding: 0}} onClick={() => this.openTraceSpanDrawer(record)}>
            {text || record.span?.spanId || "-"}
          </Button>
        ),
      },
      {
        title: i18next.t("entry:Service", {defaultValue: "Service"}),
        dataIndex: "serviceName",
        key: "serviceName",
        width: 180,
        render: value => value || "-",
      },
      {
        title: i18next.t("entry:Span ID", {defaultValue: "Span ID"}),
        dataIndex: ["span", "spanId"],
        key: "spanId",
        width: 180,
        render: value => value || "-",
      },
      {
        title: i18next.t("entry:Start time", {defaultValue: "Start time"}),
        dataIndex: ["span", "startTimeUnixNano"],
        key: "startTimeUnixNano",
        width: 220,
        render: value => this.formatTraceTimestamp(value),
      },
      {
        title: i18next.t("entry:Duration", {defaultValue: "Duration"}),
        key: "duration",
        width: 120,
        render: (_, record) => this.getSpanDuration(record.span),
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        width: 100,
        render: (_, record) => (
          <Button type="link" onClick={() => this.openTraceSpanDrawer(record)}>
            {i18next.t("general:View")}
          </Button>
        ),
      },
    ];

    return (
      <>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            Trace spans:
          </Col>
          <Col span={22} >
            {error ? (
              <Alert
                message={`Failed to parse trace message: ${error}`}
                type="warning"
                showIcon
              />
            ) : (
              <Table
                scroll={{x: "max-content"}}
                size="small"
                bordered
                columns={columns}
                dataSource={spans}
                rowKey="key"
                onRow={record => ({
                  onClick: () => this.openTraceSpanDrawer(record),
                  style: {cursor: "pointer"},
                })}
                pagination={spans.length > 10 ? {pageSize: 10, hideOnSinglePage: true} : false}
                locale={{emptyText: "No spans"}}
              />
            )}
          </Col>
        </Row>
        {this.renderTraceSpanDrawer()}
      </>
    );
  }

  renderTraceSpanDrawer() {
    const traceSpan = this.state.selectedTraceSpan;
    const span = traceSpan?.span;
    if (!traceSpan) {
      return (
        <Drawer
          title="Span detail"
          width={Setting.isMobile() ? "100%" : 760}
          placement="right"
          destroyOnClose
          onClose={this.closeTraceSpanDrawer}
          open={this.state.traceSpanDrawerVisible}
        />
      );
    }

    return (
      <Drawer
        title={`Span detail: ${span?.name || span?.spanId || "-"}`}
        width={Setting.isMobile() ? "100%" : 760}
        placement="right"
        destroyOnClose
        onClose={this.closeTraceSpanDrawer}
        open={this.state.traceSpanDrawerVisible}
      >
        <Descriptions
          bordered
          size="small"
          column={1}
          layout={Setting.isMobile() ? "vertical" : "horizontal"}
          style={{padding: "12px", height: "100%", overflowY: "auto"}}
        >
          <Descriptions.Item label={i18next.t("general:Name")}>
            {span?.name || "-"}
          </Descriptions.Item>
          <Descriptions.Item label="Service">
            {traceSpan.serviceName || "-"}
          </Descriptions.Item>
          <Descriptions.Item label="Scope">
            {this.getScopeName(traceSpan.scope)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("general:Type")}>
            {span?.kind || "-"}
          </Descriptions.Item>
          <Descriptions.Item label="Trace ID">
            {span?.traceId || "-"}
          </Descriptions.Item>
          <Descriptions.Item label="Span ID">
            {span?.spanId || "-"}
          </Descriptions.Item>
          <Descriptions.Item label="Parent Span ID">
            {span?.parentSpanId || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("general:Status")}>
            {this.getSpanStatus(span)}
          </Descriptions.Item>
          <Descriptions.Item label="Start time">
            {this.formatTraceTimestamp(span?.startTimeUnixNano)}
          </Descriptions.Item>
          <Descriptions.Item label="End time">
            {this.formatTraceTimestamp(span?.endTimeUnixNano)}
          </Descriptions.Item>
          <Descriptions.Item label="Duration">
            {this.getSpanDuration(span)}
          </Descriptions.Item>
          <Descriptions.Item label="Resource schema URL">
            {traceSpan.resourceSchemaUrl || "-"}
          </Descriptions.Item>
          <Descriptions.Item label="Scope schema URL">
            {traceSpan.scopeSchemaUrl || "-"}
          </Descriptions.Item>
          <Descriptions.Item label="Resource attributes">
            {this.renderTraceAttributeTable(traceSpan.resourceAttributes)}
          </Descriptions.Item>
          <Descriptions.Item label="Span attributes">
            {this.renderTraceAttributeTable(span?.attributes)}
          </Descriptions.Item>
          <Descriptions.Item label="Events">
            {this.renderJsonEditor(span?.events)}
          </Descriptions.Item>
          <Descriptions.Item label="Links">
            {this.renderJsonEditor(span?.links)}
          </Descriptions.Item>
          <Descriptions.Item label="Raw span">
            {this.renderJsonEditor(span)}
          </Descriptions.Item>
        </Descriptions>
      </Drawer>
    );
  }

  renderEntry() {
    return (
      <Card size="small" title={
        <div>
          {this.state.mode === "add" ? i18next.t("entry:New Entry") : i18next.t("entry:Edit Entry")}&nbsp;&nbsp;&nbsp;&nbsp;
          <Button onClick={() => this.submitEntryEdit(false)}>{i18next.t("general:Save")}</Button>
          <Button style={{marginLeft: "20px"}} type="primary" onClick={() => this.submitEntryEdit(true)}>{i18next.t("general:Save & Exit")}</Button>
          {this.state.mode === "add" ? <Button style={{marginLeft: "20px"}} onClick={() => this.deleteEntry()}>{i18next.t("general:Cancel")}</Button> : null}
        </div>
      } style={(Setting.isMobile()) ? {margin: "5px"} : {}} type="inner">
        <Row style={{marginTop: "10px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Organization"), i18next.t("general:Organization - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} disabled={!Setting.isAdminUser(this.props.account)} value={this.state.entry.owner} onChange={(value => {this.updateEntryField("owner", value);})}>
              {
                this.state.organizations.map((organization, index) => <Option key={index} value={organization.name}>{organization.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.entry.name} onChange={e => {
              this.updateEntryField("name", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("general:Display name")}:
          </Col>
          <Col span={22} >
            <Input value={this.state.entry.displayName} onChange={e => {
              this.updateEntryField("displayName", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Type"), i18next.t("general:Type - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input disabled value={this.state.entry.type ?? ""} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Listening URL"), i18next.t("general:Listening URL - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input prefix={<LinkOutlined />} value={this.state.entry.url} onChange={e => {
              this.updateEntryField("url", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("token:Access token"), i18next.t("token:Access token - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Input.Password placeholder={"***"} value={this.state.entry.token} onChange={e => {
              this.updateEntryField("token", e.target.value);
            }} />
          </Col>
        </Row>
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {Setting.getLabel(i18next.t("general:Application"), i18next.t("general:Application - Tooltip"))} :
          </Col>
          <Col span={22} >
            <Select virtual={false} style={{width: "100%"}} value={this.state.entry.application} onChange={(value => {this.updateEntryField("application", value);})}>
              {
                this.state.applications.map((application, index) => <Option key={index} value={application.name}>{application.name}</Option>)
              }
            </Select>
          </Col>
        </Row>
        {this.renderTraceSpans()}
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={(Setting.isMobile()) ? 22 : 2}>
            {i18next.t("payment:Message")}:
          </Col>
          <Col span={22} >
            {this.renderMessageEditor()}
          </Col>
        </Row>
      </Card>
    );
  }

  render() {
    if (this.state.entry === null) {
      return null;
    }

    return (
      <div>
        {this.renderEntry()}
      </div>
    );
  }
}

export default EntryEditPage;
