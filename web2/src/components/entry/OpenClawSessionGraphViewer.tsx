import * as React from "react";
import i18next from "i18next";
import {FileText, Maximize2, Minimize2} from "lucide-react";
import {useNavigate} from "react-router-dom";
import ReactFlow, {Background, Controls, MiniMap, ReactFlowProvider, type ReactFlowInstance} from "reactflow";
import "reactflow/dist/style.css";
import {Alert} from "@/components/ui/alert";
import {Badge} from "@/components/ui/badge";
import {Button} from "@/components/ui/button";
import {Sheet, SheetContent, SheetHeader, SheetTitle} from "@/components/ui/sheet";
import {Tooltip, TooltipContent, TooltipTrigger} from "@/components/ui/tooltip";
import {DescriptionList, type DescriptionItem} from "@/components/common/DescriptionList";
import {Loading} from "@/components/common/Loading";
import {FormRow} from "@/components/crud/FormRow";
import * as EntryBackend from "@/backend/EntryBackend";
import {
  buildOpenClawFlowElements,
  formatOpenClawSessionGraphTimestamp,
  getOpenClawNodeColor,
  getOpenClawNodeTarget,
} from "@/lib/openclaw-graph";

function normalizeNodeKey(value: any) {
  return `${value ?? ""}`.trim();
}

function isToolCallNode(node: any) {
  return node?.kind === "tool_call";
}

function isToolResultNode(node: any) {
  return node?.kind === "tool_result";
}

function findLinkedToolCallNode(nodes: any[], toolResultNode: any) {
  if (!isToolResultNode(toolResultNode)) {
    return null;
  }

  const parentId = normalizeNodeKey(toolResultNode.parentId);
  if (parentId) {
    const directParent = nodes.find((candidate) => {
      return isToolCallNode(candidate) && normalizeNodeKey(candidate.id) === parentId;
    });
    if (directParent) {
      return directParent;
    }
  }

  const toolCallId = normalizeNodeKey(toolResultNode.toolCallId);
  if (!toolCallId) {
    return null;
  }

  return nodes.find((candidate) => {
    return isToolCallNode(candidate) && normalizeNodeKey(candidate.toolCallId) === toolCallId;
  }) || null;
}

function findLinkedToolResultNode(nodes: any[], toolCallNode: any) {
  if (!isToolCallNode(toolCallNode)) {
    return null;
  }

  const toolCallId = normalizeNodeKey(toolCallNode.toolCallId);
  if (toolCallId) {
    const byToolCallId = nodes.find((candidate) => {
      return isToolResultNode(candidate) && normalizeNodeKey(candidate.toolCallId) === toolCallId;
    });
    if (byToolCallId) {
      return byToolCallId;
    }
  }

  const nodeId = normalizeNodeKey(toolCallNode.id);
  if (!nodeId) {
    return null;
  }

  return nodes.find((candidate) => {
    return isToolResultNode(candidate) && normalizeNodeKey(candidate.parentId) === nodeId;
  }) || null;
}

function getNodeStatusText(node: any) {
  if (node?.kind !== "tool_result" || node?.ok === undefined || node?.ok === null) {
    return "";
  }

  return node.ok ? i18next.t("general:OK") : i18next.t("webhook:Failed");
}

function StatusBadge({node}: {node: any}) {
  if (node?.kind !== "tool_result" || node?.ok === undefined || node?.ok === null) {
    return null;
  }

  return node.ok ? (
    <Badge variant="success">{i18next.t("general:OK")}</Badge>
  ) : (
    <Badge variant="destructive">{i18next.t("webhook:Failed")}</Badge>
  );
}

function NodeHoverCard({node}: {node: any}) {
  if (!node) {
    return null;
  }

  const status = getNodeStatusText(node);
  const target = getOpenClawNodeTarget(node);
  const rows: {key: string; label: string; value: string}[] = [
    {key: "type", label: i18next.t("general:Type"), value: node.kind || "-"},
    {key: "timestamp", label: i18next.t("general:Timestamp"), value: formatOpenClawSessionGraphTimestamp(node.timestamp)},
  ];

  if (node.tool) {
    rows.push({key: "tool", label: i18next.t("general:Tool"), value: node.tool});
  }
  if (target) {
    rows.push({key: "target", label: i18next.t("entry:Target"), value: target});
  }
  if (status) {
    rows.push({key: "status", label: i18next.t("general:Status"), value: status});
  }

  return (
    <div className="max-w-[min(90vw,46rem)] text-xs leading-relaxed">
      <div className="mb-2 break-words text-[13px] font-semibold">
        {node.summary || i18next.t("entry:Session graph node")}
      </div>
      <div className="grid gap-1">
        {rows.map((row) => (
          <div key={row.key} className="grid grid-cols-[88px_minmax(0,1fr)] gap-x-2">
            <span className="opacity-70">{row.label}</span>
            <span className="[overflow-wrap:anywhere] break-words">{row.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function NodeLabel({title, subtitle, node}: {title: string; subtitle: string; node: any}) {
  return (
    <Tooltip delayDuration={120}>
      <TooltipTrigger asChild>
        <div className="flex w-full flex-col gap-1.5">
          <div className="line-clamp-2 [overflow-wrap:anywhere] break-words text-[13px] font-semibold leading-snug">
            {title || "-"}
          </div>
          <div className="line-clamp-2 [overflow-wrap:anywhere] break-words text-xs leading-snug text-slate-500">
            {subtitle || "-"}
          </div>
        </div>
      </TooltipTrigger>
      <TooltipContent side="top" className="max-w-[min(90vw,50rem)]">
        <NodeHoverCard node={node} />
      </TooltipContent>
    </Tooltip>
  );
}

interface CanvasProps {
  graph: any;
  onNodeSelect: (node: any) => void;
  height?: number | string;
  fullscreen?: boolean;
  onEnterFullscreen: () => void;
  onExitFullscreen: () => void;
  topLeftOverlay?: React.ReactNode;
}

function GraphCanvas({
  graph,
  onNodeSelect,
  height = 640,
  fullscreen = false,
  onEnterFullscreen,
  onExitFullscreen,
  topLeftOverlay = null,
}: CanvasProps) {
  const [instance, setInstance] = React.useState<ReactFlowInstance | null>(null);
  const heightCss = typeof height === "number" ? `${height}px` : height;

  const elements = React.useMemo(() => {
    const flowElements = buildOpenClawFlowElements(graph);
    return {
      nodes: flowElements.nodes.map((node) => ({
        ...node,
        data: {
          ...node.data,
          label: <NodeLabel title={node.data.title} subtitle={node.data.subtitle} node={node.data.rawNode} />,
        },
      })),
      edges: flowElements.edges,
    };
  }, [graph]);

  React.useEffect(() => {
    if (!instance || elements.nodes.length === 0) {
      return;
    }

    instance.fitView({padding: 0.2, duration: 0, minZoom: 0.05});
    const anchorNode = elements.nodes.find((node) => node.data?.isAnchor);
    if (!anchorNode) {
      return;
    }

    // after fitView has settled, re-center on the entry the user came from
    const timer = window.setTimeout(() => {
      const anchorWidth = Number(anchorNode.style?.width) || 250;
      const anchorHeight = Number(anchorNode.style?.minHeight) || 76;
      instance.setCenter(
        anchorNode.position.x + anchorWidth / 2,
        anchorNode.position.y + anchorHeight / 2,
        {zoom: 1.02, duration: 0},
      );
    }, 0);
    return () => window.clearTimeout(timer);
  }, [elements.nodes, instance, fullscreen]);

  return (
    <div
      className="relative w-full overflow-hidden rounded-2xl border-2"
      style={{height: heightCss, minHeight: typeof height === "number" ? height : 0}}
    >
      <ReactFlow
        style={{width: "100%", height: "100%"}}
        nodes={elements.nodes}
        edges={elements.edges}
        fitView
        fitViewOptions={{padding: 0.2, minZoom: 0.05}}
        minZoom={0.05}
        nodesDraggable={false}
        nodesConnectable={false}
        onInit={setInstance}
        onNodeClick={(_, node) => onNodeSelect(node.data?.rawNode ?? null)}
        proOptions={{hideAttribution: true}}
      >
        <MiniMap pannable zoomable nodeColor={(node) => getOpenClawNodeColor(node.data?.rawNode)} />
        <Controls showInteractive={false} />
        <Background color="#f1f5f9" gap={16} />
      </ReactFlow>
      {topLeftOverlay ? (
        <div className="pointer-events-none absolute left-2 right-12 top-2 z-[9]">{topLeftOverlay}</div>
      ) : null}
      <div className="absolute right-2 top-2 z-10">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant={fullscreen ? "default" : "outline"}
              size="iconSm"
              onClick={fullscreen ? onExitFullscreen : onEnterFullscreen}
              aria-label={
                fullscreen
                  ? i18next.t("entry:Exit session graph fullscreen")
                  : i18next.t("entry:Session graph fullscreen")
              }
            >
              {fullscreen ? <Minimize2 className="h-4 w-4" /> : <Maximize2 className="h-4 w-4" />}
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            {fullscreen
              ? i18next.t("entry:Exit session graph fullscreen")
              : i18next.t("entry:Session graph fullscreen")}
          </TooltipContent>
        </Tooltip>
      </div>
    </div>
  );
}

function ToolPairPanel({
  title,
  node,
  currentNode,
  onOpenLinked,
}: {
  title: string;
  node: any;
  currentNode: any;
  onOpenLinked: (node: any) => void;
}) {
  const isCurrentNode = normalizeNodeKey(node?.id) === normalizeNodeKey(currentNode?.id);
  const target = getOpenClawNodeTarget(node);
  const displayTarget = target && target !== node?.tool ? target : "";

  const line = (label: string, value: React.ReactNode) => (
    <div className="min-w-0">
      <span className="text-muted-foreground">{label}: </span>
      <span className="[overflow-wrap:anywhere] whitespace-pre-wrap break-words">{value}</span>
    </div>
  );

  return (
    <div className="grid max-h-[420px] min-h-0 min-w-0 grid-rows-[auto_minmax(0,1fr)] gap-2.5 overflow-hidden rounded-xl border bg-muted/30 p-3">
      <div className="flex min-w-0 items-start justify-between gap-3">
        <div className="flex min-w-0 flex-wrap items-center gap-2">
          <span className="text-sm font-semibold">{title}</span>
          {node ? <StatusBadge node={node} /> : null}
        </div>
        {node && !isCurrentNode ? (
          <Button variant="outline" size="sm" className="flex-none" onClick={() => onOpenLinked(node)}>
            {i18next.t("entry:Open linked node")}
          </Button>
        ) : null}
      </div>
      {node ? (
        <div className="grid min-h-0 min-w-0 gap-2 overflow-y-auto overflow-x-hidden pr-1 text-sm">
          {line(i18next.t("entry:Summary"), node.summary || "-")}
          {line(i18next.t("general:Tool"), node.tool || "-")}
          {displayTarget ? line(i18next.t("entry:Target"), displayTarget) : null}
          {node.error ? line(i18next.t("general:Error"), node.error) : null}
          {node.text ? line(i18next.t("entry:Text"), node.text) : null}
        </div>
      ) : (
        <span className="text-sm text-muted-foreground">-</span>
      )}
    </div>
  );
}

/** Multi-line node field, `-` when empty. */
function NodeText({value}: {value: any}) {
  if (!value) {
    return <>-</>;
  }
  return <div className="max-w-full [overflow-wrap:anywhere] whitespace-pre-wrap break-words">{value}</div>;
}

interface ViewerProps {
  entry: any;
  provider?: any;
  /** stack the label above the graph (used inside the list-page popover) */
  block?: boolean;
}

/** Port of `web/src/OpenClawSessionGraphViewer.js`. */
export function OpenClawSessionGraphViewer({entry, provider, block}: ViewerProps) {
  const navigate = useNavigate();
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState("");
  const [graph, setGraph] = React.useState<any>(null);
  const [selectedNode, setSelectedNode] = React.useState<any>(null);
  const [fullscreen, setFullscreen] = React.useState(false);

  const owner = entry?.owner;
  const name = entry?.name;

  React.useEffect(() => {
    setFullscreen(false);
    if (!owner || !name) {
      setLoading(false);
      setError("");
      setGraph(null);
      setSelectedNode(null);
      return;
    }

    let cancelled = false;
    setLoading(true);
    setError("");
    setSelectedNode(null);

    EntryBackend.getOpenClawSessionGraph(owner, name)
      .then((res: any) => {
        if (cancelled) {
          return;
        }
        setLoading(false);
        if (res.status === "ok") {
          setError("");
          setGraph(res.data ?? null);
        } else {
          setError(`${i18next.t("entry:Failed to load session graph")}: ${res.msg}`);
          setGraph(null);
        }
      })
      .catch((err: any) => {
        if (cancelled) {
          return;
        }
        setLoading(false);
        setError(`${i18next.t("entry:Failed to load session graph")}: ${err}`);
        setGraph(null);
      });

    return () => {
      cancelled = true;
    };
  }, [owner, name, provider]);

  // Escape leaves fullscreen, and the page behind it must not scroll while it is on.
  React.useEffect(() => {
    if (!fullscreen) {
      return;
    }
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setFullscreen(false);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKeyDown);
      document.body.style.overflow = "";
    };
  }, [fullscreen]);

  const nodes: any[] = Array.isArray(graph?.nodes) ? graph.nodes : [];

  const getToolPairNodes = (node: any) => {
    if (isToolCallNode(node)) {
      return {callNode: node, resultNode: findLinkedToolResultNode(nodes, node)};
    }
    if (isToolResultNode(node)) {
      return {callNode: findLinkedToolCallNode(nodes, node), resultNode: node};
    }
    return {callNode: null, resultNode: null};
  };

  const renderStats = () => {
    const stats = graph?.stats;
    if (!stats) {
      return null;
    }

    return (
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <div className="flex min-w-0 flex-1 flex-wrap gap-2">
          <Badge variant="secondary">{i18next.t("site:Nodes")}: {stats.totalNodes}</Badge>
          <Badge variant="outline" className="border-blue-500/30 bg-blue-500/10 text-blue-600 dark:text-blue-400">
            {i18next.t("entry:Tasks")}: {stats.taskCount}
          </Badge>
          <Badge variant="warning">{i18next.t("entry:Tool calls")}: {stats.toolCallCount}</Badge>
          <Badge variant="success">{i18next.t("entry:Results")}: {stats.toolResultCount}</Badge>
          <Badge variant="outline" className="border-violet-500/30 bg-violet-500/10 text-violet-600 dark:text-violet-400">
            {i18next.t("entry:Finals")}: {stats.finalCount}
          </Badge>
          {stats.failedCount > 0 ? (
            <Badge variant="destructive">{i18next.t("webhook:Failed")}: {stats.failedCount}</Badge>
          ) : null}
        </div>
        {graph?.rawTranscript && owner && name ? (
          <Button
            variant="outline"
            size="sm"
            className="pointer-events-auto"
            onClick={() => navigate(`/entries/${owner}/${encodeURIComponent(name)}/transcript`)}
          >
            <FileText className="mr-1 h-3.5 w-3.5" />
            {i18next.t("entry:Raw JSONL")}
          </Button>
        ) : null}
      </div>
    );
  };

  const renderNodeDrawer = () => {
    const node = selectedNode;
    const {callNode, resultNode} = getToolPairNodes(node);
    const hasToolPair = Boolean(callNode || resultNode);

    const items: DescriptionItem[] = node
      ? [
        {
          label: i18next.t("general:Type"),
          children: (
            <div className="flex items-center gap-2">
              <span>{node.kind || "-"}</span>
              <StatusBadge node={node} />
            </div>
          ),
        },
        {label: i18next.t("entry:Summary"), children: node.summary || "-"},
        {label: i18next.t("general:Timestamp"), children: node.timestamp || "-"},
        {label: i18next.t("entry:Entry ID"), children: node.entryId || "-"},
        {label: i18next.t("entry:Tool Call ID"), children: node.toolCallId || "-"},
        {label: `${i18next.t("general:Parent")} ${i18next.t("general:ID")}`, children: node.parentId || "-"},
        {label: i18next.t("entry:Original Parent ID"), children: node.originalParentId || "-"},
        {label: i18next.t("entry:Target"), children: getOpenClawNodeTarget(node) || "-"},
        {label: i18next.t("general:Tool"), children: node.tool || "-"},
        {
          label: i18next.t("entry:Call / Result"),
          hidden: !hasToolPair,
          children: (
            <div className="grid min-w-0 gap-3 md:grid-cols-2">
              <ToolPairPanel
                title={i18next.t("entry:Call")}
                node={callNode}
                currentNode={node}
                onOpenLinked={setSelectedNode}
              />
              <ToolPairPanel
                title={i18next.t("payment:Result")}
                node={resultNode}
                currentNode={node}
                onOpenLinked={setSelectedNode}
              />
            </div>
          ),
        },
        {label: i18next.t("entry:Query"), children: <NodeText value={node.query} />},
        {label: i18next.t("general:URL"), children: <NodeText value={node.url} />},
        {label: i18next.t("general:Path"), children: <NodeText value={node.path} />},
        {label: i18next.t("general:Error"), children: <NodeText value={node.error} />},
        {label: i18next.t("entry:Text"), children: <NodeText value={node.text} />},
      ]
      : [];

    return (
      <Sheet open={selectedNode !== null} onOpenChange={(open) => !open && setSelectedNode(null)}>
        <SheetContent side="right" className="flex w-full flex-col gap-0 overflow-y-auto p-0 sm:max-w-[720px]">
          <SheetHeader className="border-b p-4 pr-12 text-left">
            <SheetTitle className="text-base [overflow-wrap:anywhere] break-words">
              {node?.summary || i18next.t("entry:Session graph node")}
            </SheetTitle>
          </SheetHeader>
          <div className="p-4">{node ? <DescriptionList items={items} /> : null}</div>
        </SheetContent>
      </Sheet>
    );
  };

  const renderContent = () => {
    if (loading) {
      return <Loading />;
    }
    if (error) {
      return <Alert variant="warning">{error}</Alert>;
    }
    if (!graph) {
      return null;
    }

    return (
      <>
        {!fullscreen ? renderStats() : null}
        <div className={fullscreen ? "fixed inset-0 z-[999] flex flex-col bg-background p-4" : undefined}>
          <div className={fullscreen ? "relative min-h-0 flex-1" : "relative"}>
            <ReactFlowProvider>
              <GraphCanvas
                graph={graph}
                height={fullscreen ? "100%" : 640}
                fullscreen={fullscreen}
                topLeftOverlay={fullscreen ? renderStats() : null}
                onEnterFullscreen={() => setFullscreen(true)}
                onExitFullscreen={() => setFullscreen(false)}
                onNodeSelect={setSelectedNode}
              />
            </ReactFlowProvider>
          </div>
        </div>
        {renderNodeDrawer()}
      </>
    );
  };

  if (!loading && !error && !graph) {
    return null;
  }

  return (
    <FormRow label={`${i18next.t("entry:Session graph")}:`} block={block}>
      <div data-testid="openclaw-session-graph">{renderContent()}</div>
    </FormRow>
  );
}

export default OpenClawSessionGraphViewer;
