/**
 * Reading OTLP trace payloads, ported from the helpers inside
 * `web/src/EntryMessageViewer.js`. Pure, so the entry viewer only renders.
 */

import * as Setting from "@/lib/setting";

export function formatJsonValue(value: any): string {
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

/** Renders an OTLP `AnyValue` (the `stringValue` / `intValue` / ... union) as text. */
export function formatAnyValue(value: any): string {
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
    return value.arrayValue.values.map((item: any) => formatAnyValue(item)).join(", ");
  }
  if (Array.isArray(value.kvlistValue?.values)) {
    return value.kvlistValue.values.map((item: any) => `${item?.key || "-"}=${formatAnyValue(item?.value)}`).join(", ");
  }

  return formatJsonValue(value);
}

export function getAnyValueType(value: any): string {
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

export function getAttributeValue(attributes: any[], key: string): string {
  const attribute = attributes.find((item) => item?.key === key);
  return attribute ? formatAnyValue(attribute.value) : "";
}

function normalizeIntegerString(value: any): string {
  const text = `${value ?? ""}`.trim();
  if (!/^\d+$/.test(text)) {
    return "";
  }

  return text.replace(/^0+(?=\d)/, "");
}

/**
 * OTLP timestamps are nanoseconds as decimal strings, which overflow `Number`.
 * Subtract them digit by digit so span durations stay exact.
 */
export function subtractIntegerStrings(minuend: any, subtrahend: any): string {
  const left = normalizeIntegerString(minuend);
  const right = normalizeIntegerString(subtrahend);
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

export function formatTraceTimestamp(unixNano: any): string {
  if (!unixNano) {
    return "-";
  }

  const normalized = normalizeIntegerString(unixNano);
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

export function getSpanDuration(span: any): string {
  if (!span?.startTimeUnixNano || !span?.endTimeUnixNano) {
    return "-";
  }

  const duration = subtractIntegerStrings(span.endTimeUnixNano, span.startTimeUnixNano);
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

export function getSpanStatus(span: any): string {
  const code = span?.status?.code ?? "";
  const message = span?.status?.message ?? "";

  if (code && message) {
    return `${code}: ${message}`;
  }

  return code || message || "-";
}

export function getScopeName(scope: any): string {
  if (!scope?.name) {
    return "-";
  }

  return scope.version ? `${scope.name}@${scope.version}` : scope.name;
}

export interface TraceSpanRow {
  key: string;
  resource: any;
  resourceAttributes: any[];
  resourceSchemaUrl: string;
  scope: any;
  scopeSchemaUrl: string;
  serviceName: string;
  span: any;
}

/** Flattens `resourceSpans[].scopeSpans[].spans[]` into one list of table rows. */
export function flattenTraceSpans(trace: any): TraceSpanRow[] {
  const spans: TraceSpanRow[] = [];
  const resourceSpans: any[] = Array.isArray(trace?.resourceSpans) ? trace.resourceSpans : [];

  resourceSpans.forEach((resourceSpan, resourceIndex) => {
    const resource = resourceSpan?.resource ?? {};
    const resourceAttributes = Array.isArray(resource.attributes) ? resource.attributes : [];
    const serviceName = getAttributeValue(resourceAttributes, "service.name");
    const scopeSpans: any[] = Array.isArray(resourceSpan?.scopeSpans) ? resourceSpan.scopeSpans : [];

    scopeSpans.forEach((scopeSpan, scopeIndex) => {
      const scope = scopeSpan?.scope ?? {};
      const scopeSchemaUrl = scopeSpan?.schemaUrl ?? "";
      const innerSpans: any[] = Array.isArray(scopeSpan?.spans) ? scopeSpan.spans : [];

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

/** The CodeMirror language for an entry `message`, or undefined for plain text. */
export function getMessageEditorLang(rawMessage: any): "json" | undefined {
  if (rawMessage === undefined || rawMessage === null || rawMessage === "") {
    return undefined;
  }

  const type = typeof rawMessage;
  if (type === "object" || type === "number" || type === "boolean" || type === "bigint") {
    return "json";
  }
  if (type === "string") {
    try {
      JSON.parse(rawMessage);
      return "json";
    } catch (e) {
      return undefined;
    }
  }

  return undefined;
}
