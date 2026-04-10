const openClawPayloadKinds = new Set(["task", "tool_call", "tool_result", "final"]);

export function isOpenClawSessionEntry(entry, provider) {
  if (!entry || `${entry.type ?? ""}`.trim().toLowerCase() !== "session") {
    return false;
  }

  if (provider?.category === "Log" && provider?.type === "Agent" && provider?.subType === "OpenClaw") {
    return true;
  }

  if (provider) {
    return false;
  }

  const payload = parseOpenClawBehaviorPayload(entry.message);
  return Boolean(payload?.sessionId && payload?.entryId && payload?.kind);
}

function parseOpenClawBehaviorPayload(message) {
  if (!message) {
    return null;
  }

  const source = typeof message === "string" ? message : JSON.stringify(message);
  if (!source) {
    return null;
  }

  try {
    const payload = JSON.parse(source);
    const kind = `${payload?.kind ?? ""}`.trim();
    const sessionId = `${payload?.sessionId ?? ""}`.trim();
    const entryId = `${payload?.entryId ?? ""}`.trim();
    if (!kind || !sessionId || !entryId || !openClawPayloadKinds.has(kind)) {
      return null;
    }
    return payload;
  } catch (e) {
    return null;
  }
}

export function getOpenClawNodeTarget(node) {
  return node?.query || node?.url || node?.path || node?.tool || "";
}

export function getOpenClawNodeColor(node) {
  switch (node?.kind) {
  case "task":
    return "#4c6ef5";
  case "assistant_step":
    return "#0f766e";
  case "tool_call":
    return "#f08c00";
  case "tool_result":
    return node?.ok === false ? "#e03131" : "#2f9e44";
  case "final":
    return "#6c5ce7";
  default:
    return "#868e96";
  }
}

function normalizeText(value) {
  return `${value ?? ""}`.replace(/\s+/g, " ").trim();
}

function stripLeadingPrefix(text, prefix) {
  const normalizedText = normalizeText(text);
  const normalizedPrefix = normalizeText(prefix);
  if (!normalizedText || !normalizedPrefix) {
    return normalizedText;
  }

  if (normalizedText.toLowerCase().startsWith(normalizedPrefix.toLowerCase())) {
    return normalizedText.slice(normalizedPrefix.length).trim();
  }

  return normalizedText;
}

function getAssistantStepTitle(node) {
  const summary = normalizeText(node?.summary);
  const match = summary.match(/^(\d+\s+tool calls?)(?:\s*:\s*.+)?$/i);
  if (match) {
    return match[1];
  }
  return summary || node?.id || "-";
}

function getToolCallTitle(node) {
  const target = normalizeText(getOpenClawNodeTarget(node));
  if (target) {
    return target;
  }

  const prefix = node?.tool ? `${node.tool}:` : "";
  return stripLeadingPrefix(node?.summary, prefix) || normalizeText(node?.summary) || node?.id || "-";
}

function getToolResultTitle(node) {
  const target = normalizeText(getOpenClawNodeTarget(node));
  if (target) {
    return target;
  }
  if (node?.ok === false && node?.error) {
    return normalizeText(node.error);
  }

  const prefix = node?.tool ? `${node.tool} ${node.ok === false ? "failed" : "ok"}:` : "";
  return stripLeadingPrefix(node?.summary, prefix) || normalizeText(node?.summary) || node?.id || "-";
}

function getNodeTitle(node) {
  switch (node?.kind) {
  case "assistant_step":
    return getAssistantStepTitle(node);
  case "tool_call":
    return getToolCallTitle(node);
  case "tool_result":
    return getToolResultTitle(node);
  default:
    return normalizeText(node?.summary) || node?.id || "-";
  }
}

function compareNodes(left, right) {
  const leftTimestamp = `${left?.timestamp ?? ""}`.trim();
  const rightTimestamp = `${right?.timestamp ?? ""}`.trim();
  const leftMillis = parseTimestampMillis(leftTimestamp);
  const rightMillis = parseTimestampMillis(rightTimestamp);
  if (leftMillis !== null && rightMillis !== null) {
    if (leftMillis !== rightMillis) {
      return leftMillis - rightMillis;
    }
  } else if (leftTimestamp !== rightTimestamp) {
    return leftTimestamp.localeCompare(rightTimestamp);
  }

  return `${left?.id ?? ""}`.localeCompare(`${right?.id ?? ""}`);
}

function parseTimestampMillis(timestamp) {
  if (!timestamp) {
    return null;
  }

  const milliseconds = Date.parse(timestamp);
  if (Number.isNaN(milliseconds)) {
    return null;
  }

  return milliseconds;
}

function buildTreeIndexes(graph) {
  const sourceNodes = Array.isArray(graph?.nodes) ? graph.nodes : [];
  const sourceEdges = Array.isArray(graph?.edges) ? graph.edges : [];
  const nodeMap = Object.fromEntries(sourceNodes.map(node => [node.id, node]));
  const childrenMap = new Map();
  const incomingCount = new Map();

  sourceNodes.forEach((node) => {
    childrenMap.set(node.id, []);
    incomingCount.set(node.id, 0);
  });

  sourceEdges.forEach((edge) => {
    if (!nodeMap[edge.source] || !nodeMap[edge.target]) {
      return;
    }

    if (!childrenMap.has(edge.source)) {
      childrenMap.set(edge.source, []);
    }
    childrenMap.get(edge.source).push(edge.target);
    incomingCount.set(edge.target, (incomingCount.get(edge.target) || 0) + 1);
  });

  childrenMap.forEach((childIds) => childIds.sort((left, right) => compareNodes(nodeMap[left], nodeMap[right])));
  const roots = sourceNodes
    .filter(node => !incomingCount.get(node.id))
    .sort(compareNodes)
    .map(node => node.id);

  return {nodeMap, childrenMap, roots};
}

function computeTreeLayout(graph) {
  const {nodeMap, childrenMap, roots} = buildTreeIndexes(graph);
  const positions = new Map();
  const visited = new Set();
  const verticalGap = 160;
  const horizontalGap = 320;
  let cursor = 0;

  function placeNode(nodeId, depth, stack) {
    if (!nodeMap[nodeId]) {
      return {top: cursor * verticalGap, bottom: cursor * verticalGap, center: cursor * verticalGap};
    }
    if (positions.has(nodeId)) {
      const y = positions.get(nodeId).y;
      return {top: y, bottom: y, center: y};
    }
    if (stack.has(nodeId)) {
      const y = cursor * verticalGap;
      cursor += 1;
      positions.set(nodeId, {x: depth * horizontalGap, y});
      visited.add(nodeId);
      return {top: y, bottom: y, center: y};
    }

    stack.add(nodeId);
    const childIds = (childrenMap.get(nodeId) || []).filter(childId => nodeMap[childId]);
    if (childIds.length === 0) {
      const y = cursor * verticalGap;
      cursor += 1;
      positions.set(nodeId, {x: depth * horizontalGap, y});
      visited.add(nodeId);
      stack.delete(nodeId);
      return {top: y, bottom: y, center: y};
    }

    const childBoxes = childIds.map(childId => placeNode(childId, depth + 1, stack));
    const top = childBoxes[0].top;
    const bottom = childBoxes[childBoxes.length - 1].bottom;
    const center = childBoxes.length === 1 ? childBoxes[0].center : (top + bottom) / 2;
    positions.set(nodeId, {x: depth * horizontalGap, y: center});
    visited.add(nodeId);
    stack.delete(nodeId);
    return {top, bottom, center};
  }

  roots.forEach(rootId => placeNode(rootId, 0, new Set()));

  Object.values(nodeMap)
    .filter(node => !visited.has(node.id))
    .sort(compareNodes)
    .forEach((node) => {
      placeNode(node.id, 0, new Set());
    });

  return positions;
}

function getNodeSubtitle(node) {
  switch (node?.kind) {
  case "assistant_step": {
    const summary = normalizeText(node?.summary);
    const parts = summary.split(":");
    const detail = parts.length > 1 ? parts.slice(1).join(":").trim() : "";
    return detail || node?.timestamp || "-";
  }
  case "tool_call":
    return normalizeText(node?.tool) || node?.timestamp || "-";
  case "tool_result":
    if (node?.ok === false) {
      return normalizeText(node?.error) || `${normalizeText(node?.tool) || "tool"} failed`;
    }
    return `${normalizeText(node?.tool) || "tool"} ok`;
  default:
    return getOpenClawNodeTarget(node) || node?.timestamp || "-";
  }
}

function getNodeBackground(node) {
  switch (node?.kind) {
  case "assistant_step":
    return "#f0fdfa";
  case "tool_call":
    return "#fff7ed";
  case "tool_result":
    return node?.ok === false ? "#fff5f5" : "#f3faf4";
  case "final":
    return "#f5f3ff";
  default:
    return "#ffffff";
  }
}

function getEdgeStyle(edge, nodeMap) {
  const targetNode = nodeMap[edge.target];
  if (targetNode?.kind === "tool_result" && targetNode?.ok === false) {
    return {
      stroke: "#e03131",
      strokeWidth: 2.5,
    };
  }

  if (targetNode?.originalParentId && targetNode.originalParentId !== targetNode.parentId) {
    return {
      stroke: "#0f766e",
      strokeWidth: 2.5,
      strokeDasharray: "6 4",
    };
  }

  return {
    stroke: "#94a3b8",
    strokeWidth: 2,
  };
}

export function buildOpenClawFlowElements(graph) {
  const sourceNodes = Array.isArray(graph?.nodes) ? graph.nodes : [];
  const sourceEdges = Array.isArray(graph?.edges) ? graph.edges : [];
  const nodeMap = Object.fromEntries(sourceNodes.map(node => [node.id, node]));
  const positions = computeTreeLayout(graph);

  const flowNodes = sourceNodes
    .slice()
    .sort(compareNodes)
    .map((node) => {
      const color = getOpenClawNodeColor(node);
      const position = positions.get(node.id) || {x: 0, y: 0};
      return {
        id: node.id,
        position,
        data: {
          title: getNodeTitle(node),
          subtitle: getNodeSubtitle(node),
          rawNode: node,
          isAnchor: node.isAnchor,
        },
        draggable: false,
        selectable: true,
        style: {
          width: 250,
          minHeight: 76,
          padding: "12px 14px",
          borderRadius: 14,
          border: node.isAnchor ? `3px solid ${color}` : `1px solid ${color}`,
          boxShadow: node.isAnchor ? "0 8px 24px rgba(0, 0, 0, 0.12)" : "0 4px 14px rgba(0, 0, 0, 0.08)",
          background: getNodeBackground(node),
          color: "#1f2937",
        },
      };
    });

  const flowEdges = sourceEdges.map(edge => ({
    id: `${edge.source}-${edge.target}`,
    source: edge.source,
    target: edge.target,
    type: "smoothstep",
    animated: false,
    style: getEdgeStyle(edge, nodeMap),
  }));

  return {nodes: flowNodes, edges: flowEdges};
}
