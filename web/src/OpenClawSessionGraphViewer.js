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
import {FullscreenExitOutlined, FullscreenOutlined} from "@ant-design/icons";
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
  getOpenClawNodeColor,
  getOpenClawNodeTarget
} from "./OpenClawSessionGraphUtils";

const {Text} = Typography;

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
      value: node.timestamp || "-",
    },
  ];

  if (node.tool) {
    rows.push({
      key: "tool",
      label: i18next.t("entry:Tool"),
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
    <div style={{maxWidth: 420}}>
      <div style={{fontWeight: 600, marginBottom: 8, wordBreak: "break-word"}}>{title}</div>
      <div style={{display: "grid", rowGap: 6}}>
        {rows.map((row) => (
          <div key={row.key} style={{display: "grid", gridTemplateColumns: "92px minmax(0, 1fr)", columnGap: 8}}>
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
      overlayInnerStyle={{maxWidth: 460}}
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
    height = 460,
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

    reactFlowInstance.fitView({padding: 0.2, duration: 0});
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
        border: "1px solid #e5e7eb",
        borderRadius: 16,
        overflow: "hidden",
      }}
    >
      <ReactFlow
        style={{width: "100%", height: "100%"}}
        nodes={elements.nodes}
        edges={elements.edges}
        fitView
        nodesDraggable={false}
        nodesConnectable={false}
        onInit={setReactFlowInstance}
        onNodeClick={(_, node) => onNodeSelect(node.data?.rawNode ?? null)}
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
        style={{display: "flex", flexWrap: "wrap", gap: 8, marginBottom: 12}}
      >
        <Tag color="default">{i18next.t("site:Nodes")}: {stats.totalNodes}</Tag>
        <Tag color="blue">{i18next.t("entry:Tasks")}: {stats.taskCount}</Tag>
        <Tag color="orange">{i18next.t("entry:Tool calls")}: {stats.toolCallCount}</Tag>
        <Tag color="green">{i18next.t("entry:Results")}: {stats.toolResultCount}</Tag>
        <Tag color="purple">{i18next.t("entry:Finals")}: {stats.finalCount}</Tag>
        {stats.failedCount > 0 ? (
          <Tag color="red">{i18next.t("webhook:Failed")}: {stats.failedCount}</Tag>
        ) : null}
      </div>
    );
  }

  renderNodeText(value) {
    if (!value) {
      return "-";
    }

    return (
      <div style={{whiteSpace: "pre-wrap", wordBreak: "break-word"}}>
        {value}
      </div>
    );
  }

  renderNodeDrawer() {
    const node = this.state.selectedNode;

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
                height={graphFullscreen ? "100%" : 460}
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
