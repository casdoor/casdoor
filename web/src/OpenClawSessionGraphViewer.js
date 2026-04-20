import React from "react";
import {
  Alert,
  Button,
  Col,
  Descriptions,
  Drawer,
  Row,
  Tag,
  Tooltip,
  Typography
} from "antd";
import {FileTextOutlined, FullscreenExitOutlined, FullscreenOutlined} from "@ant-design/icons";
import i18next from "i18next";
import Loading from "./common/Loading";
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  ReactFlowProvider
} from "reactflow";
import "reactflow/dist/style.css";
import * as EntryBackend from "./backend/EntryBackend";
import * as Setting from "./Setting";
import {
  buildOpenClawFlowElements,
  formatOpenClawSessionGraphTimestamp,
  getOpenClawNodeColor,
  getOpenClawNodeTarget
} from "./OpenClawSessionGraphUtils";

const {Text} = Typography;

function normalizeNodeKey(value) {
  return `${value ?? ""}`.trim();
}

function isToolCallNode(node) {
  return node?.kind === "tool_call";
}

function isToolResultNode(node) {
  return node?.kind === "tool_result";
}

function findLinkedToolCallNode(nodes, toolResultNode) {
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

function findLinkedToolResultNode(nodes, toolCallNode) {
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

function getNodeStatusText(node) {
  if (node?.kind !== "tool_result" || node?.ok === undefined || node?.ok === null) {
    return "";
  }

  return node.ok
    ? i18next.t("general:OK")
    : i18next.t("entry:Failed");
}

function OpenClawNodeHoverCard({node}) {
  if (!node) {
    return null;
  }

  const status = getNodeStatusText(node);
  const target = getOpenClawNodeTarget(node);
  const rows = [
    {
      key: "type",
      label: i18next.t("general:Type"),
      value: node.kind || "-",
    },
    {
      key: "timestamp",
      label: i18next.t("general:Timestamp"),
      value: formatOpenClawSessionGraphTimestamp(node.timestamp),
    },
  ];

  if (node.tool) {
    rows.push({
      key: "tool",
      label: i18next.t("general:Tool"),
      value: node.tool,
    });
  }

  if (target) {
    rows.push({
      key: "target",
      label: i18next.t("entry:Target"),
      value: target,
    });
  }

  if (status) {
    rows.push({
      key: "status",
      label: i18next.t("general:Status"),
      value: status,
    });
  }

  const title = node.summary || i18next.t("entry:Session graph node");

  return (
    <div style={{maxWidth: 760, fontSize: 12, lineHeight: 1.5}}>
      <div style={{fontSize: 13, fontWeight: 600, marginBottom: 7, wordBreak: "break-word"}}>{title}</div>
      <div style={{display: "grid", rowGap: 5}}>
        {rows.map((row) => (
          <div key={row.key} style={{display: "grid", gridTemplateColumns: "88px minmax(0, 1fr)", columnGap: 8}}>
            <span style={{color: "#94a3b8"}}>{row.label}</span>
            <span style={{wordBreak: "break-word", overflowWrap: "anywhere"}}>{row.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function OpenClawNodeLabel({title, subtitle, node}) {
  const titleStyle = {
    fontSize: 13,
    fontWeight: 600,
    lineHeight: 1.35,
    whiteSpace: "normal",
    overflow: "hidden",
    textOverflow: "ellipsis",
    wordBreak: "break-word",
    overflowWrap: "anywhere",
    display: "-webkit-box",
    WebkitLineClamp: 2,
    WebkitBoxOrient: "vertical",
  };
  const subtitleStyle = {
    fontSize: 12,
    color: "#64748b",
    lineHeight: 1.35,
    whiteSpace: "normal",
    overflow: "hidden",
    textOverflow: "ellipsis",
    wordBreak: "break-word",
    overflowWrap: "anywhere",
    display: "-webkit-box",
    WebkitLineClamp: 2,
    WebkitBoxOrient: "vertical",
  };

  return (
    <Tooltip
      placement="top"
      mouseEnterDelay={0.12}
      styles={{
        root: {maxWidth: 800},
        container: {maxWidth: 800},
      }}
      title={<OpenClawNodeHoverCard node={node} />}
      destroyTooltipOnHide
    >
      <div style={{display: "flex", flexDirection: "column", gap: "6px", width: "100%"}}>
        <div style={titleStyle}>
          {title || "-"}
        </div>
        <div style={subtitleStyle}>
          {subtitle || "-"}
        </div>
      </div>
    </Tooltip>
  );
}

function getStatusTag(node) {
  if (
    node?.kind !== "tool_result" ||
    node?.ok === undefined ||
    node?.ok === null
  ) {
    return null;
  }

  return node.ok ? (
    <Tag color="success">{i18next.t("general:OK")}</Tag>
  ) : (
    <Tag color="error">{i18next.t("webhook:Failed")}</Tag>
  );
}

function OpenClawSessionGraphCanvas(props) {
  const {
    graph,
    onNodeSelect,
    height = 640,
    fullscreen = false,
    onEnterFullscreen,
    onExitFullscreen,
    topLeftOverlay = null,
  } = props;
  const [reactFlowInstance, setReactFlowInstance] = React.useState(null);
  const heightCss = typeof height === "number" ? `${height}px` : height;
  const elements = React.useMemo(() => {
    const flowElements = buildOpenClawFlowElements(graph);
    return {
      nodes: flowElements.nodes.map((node) => ({
        ...node,
        data: {
          ...node.data,
          label: (
            <OpenClawNodeLabel
              title={node.data.title}
              subtitle={node.data.subtitle}
              node={node.data.rawNode}
            />
          ),
        },
      })),
      edges: flowElements.edges,
    };
  }, [graph]);

  React.useEffect(() => {
    if (!reactFlowInstance || elements.nodes.length === 0) {
      return;
    }

    reactFlowInstance.fitView({padding: 0.2, duration: 0, minZoom: 0.05});
    const anchorNode = elements.nodes.find((node) => node.data?.isAnchor);
    if (!anchorNode) {
      return;
    }

    window.setTimeout(() => {
      const anchorWidth = Number(anchorNode.style?.width) || 250;
      const anchorHeight = Number(anchorNode.style?.minHeight) || 76;
      reactFlowInstance.setCenter(
        anchorNode.position.x + anchorWidth / 2,
        anchorNode.position.y + anchorHeight / 2,
        {zoom: 1.02, duration: 0}
      );
    }, 0);
  }, [elements.nodes, reactFlowInstance, fullscreen]);

  return (
    <div
      style={{
        position: "relative",
        height: heightCss,
        width: "100%",
        minHeight: typeof height === "number" ? height : 0,
        border: "2px solid #d1d5db",
        borderRadius: 16,
        overflow: "hidden",
      }}
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
        onInit={setReactFlowInstance}
        onNodeClick={(_, node) => onNodeSelect(node.data?.rawNode ?? null)}
        proOptions={{hideAttribution: true}}
      >
        <MiniMap
          pannable
          zoomable
          nodeColor={(node) => getOpenClawNodeColor(node.data?.rawNode)}
        />
        <Controls showInteractive={false} />
        <Background color="#f1f5f9" gap={16} />
      </ReactFlow>
      {topLeftOverlay ? (
        <div
          style={{
            position: "absolute",
            top: 8,
            left: 8,
            right: 48,
            zIndex: 9,
            pointerEvents: "none",
          }}
        >
          {topLeftOverlay}
        </div>
      ) : null}
      <div
        style={{
          position: "absolute",
          top: 8,
          right: 8,
          zIndex: 10,
        }}
      >
        <Tooltip
          title={
            fullscreen
              ? i18next.t("entry:Exit session graph fullscreen")
              : i18next.t("entry:Session graph fullscreen")
          }
        >
          <Button
            type={fullscreen ? "primary" : "default"}
            size="small"
            icon={fullscreen ? <FullscreenExitOutlined /> : <FullscreenOutlined />}
            onClick={fullscreen ? onExitFullscreen : onEnterFullscreen}
          />
        </Tooltip>
      </div>
    </div>
  );
}

class OpenClawSessionGraphViewer extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      loading: false,
      error: "",
      graph: null,
      selectedNode: null,
      graphFullscreen: false,
    };
    this.requestKey = "";
    this.isUnmounted = false;
    this.handleFullscreenEscape = (e) => {
      if (e.key === "Escape" && this.state.graphFullscreen) {
        this.setState({graphFullscreen: false});
      }
    };
  }

  componentDidMount() {
    this.isUnmounted = false;
    this.loadGraph();
    window.addEventListener("keydown", this.handleFullscreenEscape);
  }

  componentDidUpdate(prevProps, prevState) {
    const entryChanged =
      prevProps.entry?.owner !== this.props.entry?.owner ||
      prevProps.entry?.name !== this.props.entry?.name;
    const providerChanged = prevProps.provider !== this.props.provider;

    if (entryChanged || providerChanged) {
      if (entryChanged && this.state.graphFullscreen) {
        this.setState({graphFullscreen: false});
      }
      this.loadGraph();
    }

    if (this.state.graphFullscreen !== prevState.graphFullscreen) {
      document.body.style.overflow = this.state.graphFullscreen ? "hidden" : "";
    }
  }

  componentWillUnmount() {
    this.isUnmounted = true;
    this.requestKey = "";
    window.removeEventListener("keydown", this.handleFullscreenEscape);
    document.body.style.overflow = "";
  }

  getLabelSpan() {
    return this.props.labelSpan ?? (Setting.isMobile() ? 22 : 2);
  }

  getContentSpan() {
    return this.props.contentSpan ?? 22;
  }

  loadGraph() {
    if (!this.props.entry?.owner || !this.props.entry?.name) {
      this.requestKey = "";
      this.setState({
        loading: false,
        error: "",
        graph: null,
        selectedNode: null,
      });
      return;
    }

    const requestKey = `${this.props.entry.owner}/${this.props.entry.name}`;
    this.requestKey = requestKey;
    this.setState({loading: true, error: "", selectedNode: null});

    EntryBackend.getOpenClawSessionGraph(
      this.props.entry.owner,
      this.props.entry.name
    )
      .then((res) => {
        if (this.isUnmounted || this.requestKey !== requestKey) {
          return;
        }

        if (res.status === "ok" && res.data) {
          this.setState({
            loading: false,
            error: "",
            graph: res.data,
          });
        } else if (res.status === "ok") {
          this.setState({
            loading: false,
            error: "",
            graph: null,
          });
        } else {
          this.setState({
            loading: false,
            error: `${i18next.t("entry:Failed to load session graph")}: ${res.msg}`,
            graph: null,
          });
        }
      })
      .catch((error) => {
        if (this.isUnmounted || this.requestKey !== requestKey) {
          return;
        }

        this.setState({
          loading: false,
          error: `${i18next.t("entry:Failed to load session graph")}: ${error}`,
          graph: null,
        });
      });
  }

  renderStats() {
    const stats = this.state.graph?.stats;
    if (!stats) {
      return null;
    }

    return (
      <div
        style={{display: "flex", flexWrap: "wrap", gap: 8, marginBottom: 12, alignItems: "center"}}
      >
        <div style={{display: "flex", flexWrap: "wrap", gap: 8, flex: 1, minWidth: 0}}>
          <Tag color="default">{i18next.t("site:Nodes")}: {stats.totalNodes}</Tag>
          <Tag color="blue">{i18next.t("entry:Tasks")}: {stats.taskCount}</Tag>
          <Tag color="orange">{i18next.t("entry:Tool calls")}: {stats.toolCallCount}</Tag>
          <Tag color="green">{i18next.t("entry:Results")}: {stats.toolResultCount}</Tag>
          <Tag color="purple">{i18next.t("entry:Finals")}: {stats.finalCount}</Tag>
          {stats.failedCount > 0 ? (
            <Tag color="red">{i18next.t("webhook:Failed")}: {stats.failedCount}</Tag>
          ) : null}
        </div>
        {this.renderRawTranscriptButton()}
      </div>
    );
  }

  renderRawTranscriptButton() {
    if (!this.state.graph?.rawTranscript || !this.props.entry?.owner || !this.props.entry?.name) {
      return null;
    }

    return (
      <Tooltip title={i18next.t("entry:Raw JSONL")}>
        <Button
          size="small"
          icon={<FileTextOutlined />}
          onClick={() => {
            window.location.href = `/entries/${this.props.entry.owner}/${encodeURIComponent(this.props.entry.name)}/transcript`;
          }}
        >
          {i18next.t("entry:Raw JSONL")}
        </Button>
      </Tooltip>
    );
  }

  renderNodeText(value, style = {}) {
    if (!value) {
      return "-";
    }

    return (
      <div style={{whiteSpace: "pre-wrap", wordBreak: "break-word", overflowWrap: "anywhere", maxWidth: "100%", ...style}}>
        {value}
      </div>
    );
  }

  getLinkedNodes(node) {
    const nodes = Array.isArray(this.state.graph?.nodes) ? this.state.graph.nodes : [];
    if (!node || nodes.length === 0) {
      return {
        linkedToolCallNode: null,
        linkedToolResultNode: null,
      };
    }

    if (isToolCallNode(node)) {
      return {
        linkedToolCallNode: null,
        linkedToolResultNode: findLinkedToolResultNode(nodes, node),
      };
    }

    if (isToolResultNode(node)) {
      return {
        linkedToolCallNode: findLinkedToolCallNode(nodes, node),
        linkedToolResultNode: null,
      };
    }

    return {
      linkedToolCallNode: null,
      linkedToolResultNode: null,
    };
  }

  getToolPairNodes(node) {
    const {linkedToolCallNode, linkedToolResultNode} = this.getLinkedNodes(node);
    if (isToolCallNode(node)) {
      return {
        callNode: node,
        resultNode: linkedToolResultNode,
      };
    }

    if (isToolResultNode(node)) {
      return {
        callNode: linkedToolCallNode,
        resultNode: node,
      };
    }

    return {
      callNode: null,
      resultNode: null,
    };
  }

  renderToolPairPanel(title, node, currentNode) {
    const isCurrentNode = normalizeNodeKey(node?.id) === normalizeNodeKey(currentNode?.id);
    const target = getOpenClawNodeTarget(node);
    const displayTarget = target && target !== node?.tool ? target : "";
    const panelMaxHeight = Setting.isMobile() ? 360 : 420;

    return (
      <div
        style={{
          display: "grid",
          gridTemplateRows: "auto minmax(0, 1fr)",
          gap: 10,
          maxHeight: panelMaxHeight,
          minHeight: 0,
          minWidth: 0,
          padding: 12,
          border: "1px solid #dbe3ef",
          borderRadius: 12,
          background: "#fafcff",
          overflow: "hidden",
        }}
      >
        <div style={{display: "flex", justifyContent: "space-between", alignItems: "flex-start", gap: 12, minWidth: 0}}>
          <div style={{display: "flex", alignItems: "center", gap: 8, minWidth: 0, flexWrap: "wrap"}}>
            <Text strong>{title}</Text>
            {node ? getStatusTag(node) : null}
          </div>
          {node && !isCurrentNode ? (
            <Button
              size="small"
              style={{flex: "none"}}
              onClick={() => this.setState({selectedNode: node})}
            >
              {i18next.t("entry:Open linked node")}
            </Button>
          ) : null}
        </div>
        {node ? (
          <div style={{display: "grid", rowGap: 8, minHeight: 0, minWidth: 0, overflowY: "auto", overflowX: "hidden", paddingRight: 4}}>
            <div style={{minWidth: 0}}>
              <Text type="secondary">{i18next.t("entry:Summary")}: </Text>
              <Text style={{wordBreak: "break-word", overflowWrap: "anywhere"}}>{node.summary || "-"}</Text>
            </div>
            <div style={{minWidth: 0}}>
              <Text type="secondary">{i18next.t("general:Tool")}: </Text>
              <Text>{node.tool || "-"}</Text>
            </div>
            {displayTarget ? (
              <div style={{minWidth: 0}}>
                <Text type="secondary">{i18next.t("entry:Target")}: </Text>
                <Text style={{wordBreak: "break-word", overflowWrap: "anywhere"}}>{displayTarget}</Text>
              </div>
            ) : null}
            {node.error ? (
              <div style={{minWidth: 0}}>
                <Text type="secondary">{i18next.t("general:Error")}: </Text>
                {this.renderNodeText(node.error)}
              </div>
            ) : null}
            {node.text ? (
              <div style={{minWidth: 0}}>
                <Text type="secondary">{i18next.t("entry:Text")}: </Text>
                {this.renderNodeText(node.text)}
              </div>
            ) : null}
          </div>
        ) : (
          <Text type="secondary">-</Text>
        )}
      </div>
    );
  }

  renderToolPair(node) {
    const {callNode, resultNode} = this.getToolPairNodes(node);
    if (!callNode && !resultNode) {
      return null;
    }

    return (
      <div
        style={{
          display: "grid",
          gridTemplateColumns: Setting.isMobile() ? "1fr" : "repeat(2, minmax(0, 1fr))",
          gap: 12,
          minWidth: 0,
        }}
      >
        {this.renderToolPairPanel(i18next.t("entry:Call"), callNode, node)}
        {this.renderToolPairPanel(i18next.t("entry:Result"), resultNode, node)}
      </div>
    );
  }

  renderNodeDrawer() {
    const node = this.state.selectedNode;
    const toolPair = this.renderToolPair(node);

    return (
      <Drawer
        title={node?.summary || i18next.t("entry:Session graph node")}
        width={Setting.isMobile() ? "100%" : 720}
        placement="right"
        onClose={() => this.setState({selectedNode: null})}
        open={this.state.selectedNode !== null}
        destroyOnClose
      >
        {node ? (
          <Descriptions
            bordered
            size="small"
            column={1}
            layout={Setting.isMobile() ? "vertical" : "horizontal"}
            style={{padding: "12px", height: "100%", overflowY: "auto"}}
          >
            <Descriptions.Item label={i18next.t("general:Type")}>
              <div style={{display: "flex", alignItems: "center", gap: 8}}>
                <Text>{node.kind || "-"}</Text>
                {getStatusTag(node)}
              </div>
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Summary")}>
              {node.summary || "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Timestamp")}>
              {node.timestamp || "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Entry ID")}>
              {node.entryId || "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Tool Call ID")}>
              {node.toolCallId || "-"}
            </Descriptions.Item>
            <Descriptions.Item label={`${i18next.t("general:Parent")} ${i18next.t("general:ID")}`}>
              {node.parentId || "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Original Parent ID")}>
              {node.originalParentId || "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Target")}>
              {getOpenClawNodeTarget(node) || "-"}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Tool")}>
              {node.tool || "-"}
            </Descriptions.Item>
            {toolPair ? (
              <Descriptions.Item label={i18next.t("entry:Call / Result")}>
                {toolPair}
              </Descriptions.Item>
            ) : null}
            <Descriptions.Item label={i18next.t("entry:Query")}>
              {this.renderNodeText(node.query)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:URL")}>
              {this.renderNodeText(node.url)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Path")}>
              {this.renderNodeText(node.path)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("general:Error")}>
              {this.renderNodeText(node.error)}
            </Descriptions.Item>
            <Descriptions.Item label={i18next.t("entry:Text")}>
              {this.renderNodeText(node.text)}
            </Descriptions.Item>
          </Descriptions>
        ) : null}
      </Drawer>
    );
  }

  renderContent() {
    if (this.state.loading) {
      return (
        <Loading />
      );
    }

    if (this.state.error) {
      return <Alert type="warning" showIcon message={this.state.error} />;
    }

    if (!this.state.graph) {
      return null;
    }

    const {graphFullscreen} = this.state;

    return (
      <>
        {!graphFullscreen ? this.renderStats() : null}
        <div
          style={
            graphFullscreen
              ? {
                position: "fixed",
                zIndex: 999,
                inset: 0,
                background: "#fff",
                display: "flex",
                flexDirection: "column",
                padding: 16,
                boxSizing: "border-box",
              }
              : {}
          }
        >
          <div
            style={{
              flex: graphFullscreen ? 1 : undefined,
              minHeight: graphFullscreen ? 0 : undefined,
              position: "relative",
            }}
          >
            <ReactFlowProvider>
              <OpenClawSessionGraphCanvas
                graph={this.state.graph}
                height={graphFullscreen ? "100%" : 640}
                fullscreen={graphFullscreen}
                topLeftOverlay={graphFullscreen ? this.renderStats() : null}
                onEnterFullscreen={() => this.setState({graphFullscreen: true})}
                onExitFullscreen={() => this.setState({graphFullscreen: false})}
                onNodeSelect={(selectedNode) => this.setState({selectedNode})}
              />
            </ReactFlowProvider>
          </div>
        </div>
        {this.renderNodeDrawer()}
      </>
    );
  }

  render() {
    if (!this.state.loading && !this.state.error && !this.state.graph) {
      return null;
    }

    return (
      <Row style={{marginTop: "20px"}}>
        <Col style={{marginTop: "5px"}} span={this.getLabelSpan()}>
          {i18next.t("entry:Session graph")}:
        </Col>
        <Col span={this.getContentSpan()}>
          <div data-testid="openclaw-session-graph">{this.renderContent()}</div>
        </Col>
      </Row>
    );
  }
}

export default OpenClawSessionGraphViewer;
