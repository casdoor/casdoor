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
import {Alert, Button, Col, Descriptions, Drawer, Row, Table} from "antd";
import * as Setting from "./Setting";
import i18next from "i18next";
import Editor from "./common/Editor";
import SELinuxEntryViewer from "./SELinuxEntryViewer";
import * as ProviderBackend from "./backend/ProviderBackend";
import OpenClawSessionGraphViewer from "./OpenClawSessionGraphViewer";
import {isOpenClawSessionEntry} from "./OpenClawSessionGraphUtils";

class EntryMessageViewer extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      traceSpanDrawerVisible: false,
      selectedTraceSpan: null,
      provider: null,
    };
    this.pendingProviderRequestKey = "";
    this.isUnmounted = false;
  }

  componentDidMount() {
    this.isUnmounted = false;
    this.getProvider();
  }

  componentDidUpdate(prevProps) {
    if (
      prevProps.entry?.provider !== this.props.entry?.provider
      || prevProps.entry?.owner !== this.props.entry?.owner
      || prevProps.provider !== this.props.provider
    ) {
      this.getProvider();
    }
  }

  componentWillUnmount() {
    this.isUnmounted = true;
    this.pendingProviderRequestKey = "";
  }

  getProvider() {
    if (this.props.provider) {
      this.pendingProviderRequestKey = "";
      if (this.state.provider !== null) {
        this.setState({provider: null});
      }
      return;
    }

    const providerName = this.props.entry?.provider;
    const owner = this.props.entry?.owner;

    if (!providerName || !owner) {
      this.pendingProviderRequestKey = "";
      if (this.state.provider !== null) {
        this.setState({provider: null});
      }
      return;
    }

    const requestKey = `${owner}/${providerName}`;
    this.pendingProviderRequestKey = requestKey;

    ProviderBackend.getProvider(owner, providerName)
      .then((res) => {
        if (this.isUnmounted || this.pendingProviderRequestKey !== requestKey) {
          return;
        }

        if (res.status === "ok") {
          this.setState({
            provider: res.data ?? null,
          });
        } else {
          this.setState({
            provider: null,
          });
        }
      })
      .catch(() => {
        if (this.isUnmounted || this.pendingProviderRequestKey !== requestKey) {
          return;
        }

        this.setState({
          provider: null,
        });
      });
  }

  getEditorMaxWidth() {
    return Setting.isMobile() ? window.innerWidth - 60 : 800;
  }

  getLabelSpan() {
    return this.props.labelSpan ?? (Setting.isMobile() ? 22 : 2);
  }

  getContentSpan() {
    return this.props.contentSpan ?? 22;
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
        title: i18next.t("general:Keys"),
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
    if (this.props.entry?.type !== "trace") {
      return {spans: [], error: ""};
    }

    const message = this.props.entry?.message?.trim();
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

  getMessageEditorHeight(text) {
    const lineHeight = 22;
    const lines = (text || "").split("\n").length;
    const visibleRows = Math.min(30, Math.max(10, lines));
    return `${visibleRows * lineHeight}px`;
  }

  getMessageEditorLang(rawMessage) {
    if (rawMessage === undefined || rawMessage === null || rawMessage === "") {
      return undefined;
    }
    const t = typeof rawMessage;
    if (t === "object") {
      return "json";
    }
    if (t === "number" || t === "boolean" || t === "bigint") {
      return "json";
    }
    if (t === "string") {
      try {
        JSON.parse(rawMessage);
        return "json";
      } catch (e) {
        return undefined;
      }
    }
    return undefined;
  }

  renderMessageEditor() {
    const rawMessage = this.props.entry?.message;
    const message = this.formatJsonValue(rawMessage) || "";
    const lang = this.getMessageEditorLang(rawMessage);

    return (
      <Editor
        value={message}
        lang={lang}
        readOnly
        fillWidth
        maxWidth={this.getEditorMaxWidth()}
        dark
        height={this.getMessageEditorHeight(message)}
      />
    );
  }

  shouldRenderTraceViewer() {
    return `${this.props.entry?.type ?? ""}`.trim().toLowerCase() === "trace";
  }

  getProviderViewerType() {
    const provider = this.props.provider ?? this.state.provider;
    if (!provider) {
      return "";
    }

    const category = `${provider.category ?? ""}`.trim();
    const type = `${provider.type ?? ""}`.trim();

    if (category === "Log" && type === "SELinux Log") {
      return "selinux";
    }

    return "";
  }

  renderSpecializedViewer() {
    const provider = this.props.provider ?? this.state.provider;
    switch (this.getProviderViewerType()) {
    case "selinux":
      return <SELinuxEntryViewer entry={this.props.entry} labelSpan={this.getLabelSpan()} contentSpan={this.getContentSpan()} />;
    default:
      if (this.shouldRenderTraceViewer()) {
        return this.renderTraceSpans();
      }
      if (isOpenClawSessionEntry(this.props.entry, provider)) {
        return <OpenClawSessionGraphViewer entry={this.props.entry} provider={provider} labelSpan={this.getLabelSpan()} contentSpan={this.getContentSpan()} />;
      }
      return null;
    }
  }

  renderTraceSpans() {
    if (this.props.entry?.type !== "trace") {
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
        title: i18next.t("entry:Service"),
        dataIndex: "serviceName",
        key: "serviceName",
        width: 180,
        render: value => value || "-",
      },
      {
        title: i18next.t("entry:Span ID"),
        dataIndex: ["span", "spanId"],
        key: "spanId",
        width: 180,
        render: value => value || "-",
      },
      {
        title: i18next.t("subscription:Start time"),
        dataIndex: ["span", "startTimeUnixNano"],
        key: "startTimeUnixNano",
        width: 220,
        render: value => this.formatTraceTimestamp(value),
      },
      {
        title: i18next.t("entry:Duration"),
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
          <Col style={{marginTop: "5px"}} span={this.getLabelSpan()}>
            {i18next.t("entry:Trace spans")}:
          </Col>
          <Col span={this.getContentSpan()} >
            {error ? (
              <Alert
                message={`${i18next.t("entry:Failed to parse trace message")}: ${error}`}
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
                locale={{emptyText: i18next.t("entry:No spans")}}
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
          title={i18next.t("entry:Span detail")}
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
        title={`${i18next.t("entry:Span detail")}: ${span?.name || span?.spanId || "-"}`}
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
          <Descriptions.Item label={i18next.t("entry:Service")}>
            {traceSpan.serviceName || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("provider:Scope")}>
            {this.getScopeName(traceSpan.scope)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("general:Type")}>
            {span?.kind || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Trace ID")}>
            {span?.traceId || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Span ID")}>
            {span?.spanId || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Parent Span ID")}>
            {span?.parentSpanId || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("general:Status")}>
            {this.getSpanStatus(span)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("subscription:Start time")}>
            {this.formatTraceTimestamp(span?.startTimeUnixNano)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("subscription:End time")}>
            {this.formatTraceTimestamp(span?.endTimeUnixNano)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Duration")}>
            {this.getSpanDuration(span)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Resource schema URL")}>
            {traceSpan.resourceSchemaUrl || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Scope schema URL")}>
            {traceSpan.scopeSchemaUrl || "-"}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Resource attributes")}>
            {this.renderTraceAttributeTable(traceSpan.resourceAttributes)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Span attributes")}>
            {this.renderTraceAttributeTable(span?.attributes)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("webhook:Events")}>
            {this.renderJsonEditor(span?.events)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Links")}>
            {this.renderJsonEditor(span?.links)}
          </Descriptions.Item>
          <Descriptions.Item label={i18next.t("entry:Raw span")}>
            {this.renderJsonEditor(span)}
          </Descriptions.Item>
        </Descriptions>
      </Drawer>
    );
  }

  render() {
    return (
      <>
        {this.renderSpecializedViewer()}
        <Row style={{marginTop: "20px"}} >
          <Col style={{marginTop: "5px"}} span={this.getLabelSpan()}>
            {i18next.t("payment:Message")}:
          </Col>
          <Col span={this.getContentSpan()} >
            {this.renderMessageEditor()}
          </Col>
        </Row>
      </>
    );
  }
}

export default EntryMessageViewer;
