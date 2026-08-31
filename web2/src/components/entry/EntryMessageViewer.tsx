import * as React from "react";
import i18next from "i18next";
import {Alert} from "@/components/ui/alert";
import {Button} from "@/components/ui/button";
import {Sheet, SheetContent, SheetHeader, SheetTitle} from "@/components/ui/sheet";
import {Table, TableBody, TableCell, TableHead, TableHeader, TableRow} from "@/components/ui/table";
import {CodeEditor} from "@/components/common/CodeEditor";
import {DescriptionList, type DescriptionItem} from "@/components/common/DescriptionList";
import {FormRow} from "@/components/crud/FormRow";
import {OpenClawSessionGraphViewer} from "@/components/entry/OpenClawSessionGraphViewer";
import {SELinuxEntryViewer} from "@/components/entry/SELinuxEntryViewer";
import * as ProviderBackend from "@/backend/ProviderBackend";
import {isOpenClawSessionEntry} from "@/lib/openclaw-graph";
import {
  flattenTraceSpans,
  formatAnyValue,
  formatJsonValue,
  formatTraceTimestamp,
  getAnyValueType,
  getMessageEditorLang,
  getScopeName,
  getSpanDuration,
  getSpanStatus,
  type TraceSpanRow,
} from "@/lib/otlp-trace";

const LINE_HEIGHT = 22;

function TraceAttributeTable({attributes}: {attributes: any[]}) {
  const rows = Array.isArray(attributes)
    ? attributes.map((attribute, index) => ({
      key: `${attribute?.key || "attribute"}-${index}`,
      name: attribute?.key || "-",
      type: getAnyValueType(attribute?.value),
      value: formatAnyValue(attribute?.value) || "-",
    }))
    : [];

  if (rows.length === 0) {
    return <>-</>;
  }

  return (
    <div className="overflow-x-auto rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[220px]">{i18next.t("general:Keys")}</TableHead>
            <TableHead className="w-[120px]">{i18next.t("general:Type")}</TableHead>
            <TableHead>{i18next.t("user:Values")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.map((row) => (
            <TableRow key={row.key}>
              <TableCell className="align-top font-medium">{row.name}</TableCell>
              <TableCell className="align-top text-muted-foreground">{row.type}</TableCell>
              <TableCell className="align-top">
                <div className="whitespace-pre-wrap break-words">{row.value}</div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

/** Read-only CodeMirror for a JSON blob, `-` when there is nothing to show. */
function JsonView({value}: {value: any}) {
  const formatted = formatJsonValue(value);
  if (!formatted) {
    return <>-</>;
  }

  const lines = formatted.split("\n").length;
  return (
    <CodeEditor
      value={formatted}
      language="json"
      readOnly
      height={Math.min(30, Math.max(6, lines)) * LINE_HEIGHT}
      onChange={() => {}}
    />
  );
}

interface EntryMessageViewerProps {
  entry: any;
  /** already-resolved Log provider; when omitted it is fetched from the entry */
  provider?: any;
  /** stack labels above their content (used inside the list-page popover) */
  block?: boolean;
}

/**
 * Port of `web/src/EntryMessageViewer.js`: renders the entry's `message`, plus a
 * viewer specialised on what produced it — SELinux audit fields, OTLP trace
 * spans, or the OpenClaw session graph.
 */
export function EntryMessageViewer({entry, provider: providerProp, block}: EntryMessageViewerProps) {
  const [fetchedProvider, setFetchedProvider] = React.useState<any>(null);
  const [selectedTraceSpan, setSelectedTraceSpan] = React.useState<TraceSpanRow | null>(null);

  const owner = entry?.owner;
  const providerName = entry?.provider;

  React.useEffect(() => {
    if (providerProp || !owner || !providerName) {
      setFetchedProvider(null);
      return;
    }

    let cancelled = false;
    ProviderBackend.getProvider(owner, providerName)
      .then((res: any) => {
        if (cancelled) {
          return;
        }
        setFetchedProvider(res.status === "ok" ? res.data ?? null : null);
      })
      .catch(() => {
        if (!cancelled) {
          setFetchedProvider(null);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [providerProp, owner, providerName]);

  const provider = providerProp ?? fetchedProvider;
  const isTrace = `${entry?.type ?? ""}`.trim().toLowerCase() === "trace";
  const isSELinux =
    `${provider?.category ?? ""}`.trim() === "Log" && `${provider?.type ?? ""}`.trim() === "SELinux Log";

  const traceData = React.useMemo(() => {
    if (!isTrace) {
      return {spans: [] as TraceSpanRow[], error: ""};
    }
    const message = `${entry?.message ?? ""}`.trim();
    if (!message) {
      return {spans: [] as TraceSpanRow[], error: ""};
    }
    try {
      return {spans: flattenTraceSpans(JSON.parse(message)), error: ""};
    } catch (e: any) {
      return {spans: [] as TraceSpanRow[], error: e.message};
    }
  }, [isTrace, entry?.message]);

  const renderTraceSpanDrawer = () => {
    const traceSpan = selectedTraceSpan;
    const span = traceSpan?.span;

    const items: DescriptionItem[] = traceSpan
      ? [
        {label: i18next.t("general:Name"), children: span?.name || "-"},
        {label: i18next.t("entry:Service"), children: traceSpan.serviceName || "-"},
        {label: i18next.t("provider:Scope"), children: getScopeName(traceSpan.scope)},
        {label: i18next.t("general:Type"), children: span?.kind || "-"},
        {label: i18next.t("entry:Trace ID"), children: span?.traceId || "-"},
        {label: i18next.t("entry:Span ID"), children: span?.spanId || "-"},
        {label: i18next.t("entry:Parent Span ID"), children: span?.parentSpanId || "-"},
        {label: i18next.t("general:Status"), children: getSpanStatus(span)},
        {label: i18next.t("subscription:Start time"), children: formatTraceTimestamp(span?.startTimeUnixNano)},
        {label: i18next.t("subscription:End time"), children: formatTraceTimestamp(span?.endTimeUnixNano)},
        {label: i18next.t("entry:Duration"), children: getSpanDuration(span)},
        {label: i18next.t("entry:Resource schema URL"), children: traceSpan.resourceSchemaUrl || "-"},
        {label: i18next.t("entry:Scope schema URL"), children: traceSpan.scopeSchemaUrl || "-"},
        {
          label: i18next.t("entry:Resource attributes"),
          children: <TraceAttributeTable attributes={traceSpan.resourceAttributes} />,
        },
        {label: i18next.t("entry:Span attributes"), children: <TraceAttributeTable attributes={span?.attributes} />},
        {label: i18next.t("webhook:Events"), children: <JsonView value={span?.events} />},
        {label: i18next.t("entry:Links"), children: <JsonView value={span?.links} />},
        {label: i18next.t("entry:Raw span"), children: <JsonView value={span} />},
      ]
      : [];

    return (
      <Sheet open={selectedTraceSpan !== null} onOpenChange={(open) => !open && setSelectedTraceSpan(null)}>
        <SheetContent side="right" className="flex w-full flex-col gap-0 overflow-y-auto p-0 sm:max-w-[760px]">
          <SheetHeader className="border-b p-4 pr-12 text-left">
            <SheetTitle className="text-base [overflow-wrap:anywhere] break-words">
              {`${i18next.t("entry:Span detail")}${span ? `: ${span.name || span.spanId || "-"}` : ""}`}
            </SheetTitle>
          </SheetHeader>
          <div className="p-4">{traceSpan ? <DescriptionList items={items} /> : null}</div>
        </SheetContent>
      </Sheet>
    );
  };

  const renderTraceSpans = () => {
    const {spans, error} = traceData;

    return (
      <>
        <FormRow label={`${i18next.t("entry:Trace spans")}:`} block={block}>
          {error ? (
            <Alert variant="warning">{`${i18next.t("entry:Failed to parse trace message")}: ${error}`}</Alert>
          ) : (
            // antd paginated at 10 rows; here the whole set scrolls inside the table
            <div className="max-h-[420px] overflow-auto rounded-md border">
              <Table>
                <TableHeader className="sticky top-0 z-10 bg-background">
                  <TableRow>
                    <TableHead className="w-[220px]">{i18next.t("general:Name")}</TableHead>
                    <TableHead className="w-[180px]">{i18next.t("entry:Service")}</TableHead>
                    <TableHead className="w-[180px]">{i18next.t("entry:Span ID")}</TableHead>
                    <TableHead className="w-[220px]">{i18next.t("subscription:Start time")}</TableHead>
                    <TableHead className="w-[120px]">{i18next.t("entry:Duration")}</TableHead>
                    <TableHead className="w-[100px]">{i18next.t("general:Action")}</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {spans.length === 0 ? (
                    <TableRow>
                      <TableCell colSpan={6} className="h-24 text-center text-muted-foreground">
                        {i18next.t("entry:No spans")}
                      </TableCell>
                    </TableRow>
                  ) : (
                    spans.map((record) => (
                      <TableRow
                        key={record.key}
                        className="cursor-pointer"
                        onClick={() => setSelectedTraceSpan(record)}
                      >
                        <TableCell className="font-medium text-primary">
                          {record.span?.name || record.span?.spanId || "-"}
                        </TableCell>
                        <TableCell>{record.serviceName || "-"}</TableCell>
                        <TableCell>{record.span?.spanId || "-"}</TableCell>
                        <TableCell className="whitespace-nowrap tabular-nums">
                          {formatTraceTimestamp(record.span?.startTimeUnixNano)}
                        </TableCell>
                        <TableCell className="whitespace-nowrap tabular-nums">{getSpanDuration(record.span)}</TableCell>
                        <TableCell>
                          <Button
                            variant="link"
                            size="sm"
                            className="h-auto p-0"
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedTraceSpan(record);
                            }}
                          >
                            {i18next.t("general:View")}
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))
                  )}
                </TableBody>
              </Table>
            </div>
          )}
        </FormRow>
        {renderTraceSpanDrawer()}
      </>
    );
  };

  const renderSpecializedViewer = () => {
    if (isSELinux) {
      return <SELinuxEntryViewer entry={entry} block={block} />;
    }
    if (isTrace) {
      return renderTraceSpans();
    }
    if (isOpenClawSessionEntry(entry, provider)) {
      return <OpenClawSessionGraphViewer entry={entry} provider={provider} block={block} />;
    }
    return null;
  };

  const message = formatJsonValue(entry?.message) || "";
  const messageHeight = Math.min(30, Math.max(10, message.split("\n").length)) * LINE_HEIGHT;

  return (
    <>
      {renderSpecializedViewer()}
      <FormRow label={`${i18next.t("payment:Message")}:`} block={block}>
        <CodeEditor
          value={message}
          language={getMessageEditorLang(entry?.message)}
          readOnly
          height={messageHeight}
          onChange={() => {}}
        />
      </FormRow>
    </>
  );
}

export default EntryMessageViewer;
