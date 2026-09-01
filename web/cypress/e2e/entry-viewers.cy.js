// The entry edit page picks a viewer based on what produced the entry: SELinux
// audit fields, OTLP trace spans, or the OpenClaw session graph. Real databases
// rarely hold one of each, so these specs stub the three endpoints the viewers
// read and assert on the rendered result.

const ok = (data) => ({status: "ok", msg: "", data});

const baseEntry = {
  owner: "built-in",
  name: "e2e-entry",
  createdTime: "2026-02-01T10:00:00Z",
  displayName: "e2e entry",
  application: "",
  clientIp: "127.0.0.1",
  userAgent: "cypress",
};

// Match on `pathname`, not a URL glob: the Casdoor query string is
// `?id=<owner>/<name>`, and the "/" in it makes minimatch treat the object name
// as a further path segment, so `**/api/get-entry*` never matches.
function stubEntry(entry, provider) {
  cy.intercept({method: "GET", pathname: "/api/get-entry"}, ok(entry)).as("getEntry");
  cy.intercept({method: "GET", pathname: "/api/get-provider"}, ok(provider)).as("getProvider");
}

describe("Entry viewers", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("renders the SELinux viewer for a SELinux Log provider", () => {
    stubEntry(
      {
        ...baseEntry,
        provider: "provider-selinux",
        type: "log",
        message:
          '[warning] type=AVC msg=audit(1700000000.123:456): avc:  denied  { read write } for  ' +
          'pid=1234 comm="httpd" exe="/usr/sbin/httpd" path="/var/www/x" dev="dm-0" ino=98765 ' +
          "scontext=system_u:system_r:httpd_t:s0 tcontext=unconfined_u:object_r:default_t:s0 " +
          "tclass=file permissive=0",
      },
      {owner: "built-in", name: "provider-selinux", category: "Log", type: "SELinux Log"},
    );

    cy.visit("/entries/built-in/e2e-entry");
    cy.wait(["@getEntry", "@getProvider"]);

    // the parsed audit fields, not the raw line
    cy.contains("denied").should("be.visible");
    cy.contains("read write").should("be.visible");
    cy.contains("httpd").should("be.visible");
    cy.contains("system_u:system_r:httpd_t:s0").should("be.visible");
    cy.contains("98765").should("be.visible");
  });

  it("renders the trace span table and the span drawer", () => {
    const trace = {
      resourceSpans: [
        {
          schemaUrl: "https://opentelemetry.io/schemas/1.21.0",
          resource: {attributes: [{key: "service.name", value: {stringValue: "casdoor-e2e"}}]},
          scopeSpans: [
            {
              schemaUrl: "scope-schema",
              scope: {name: "go.opentelemetry.io/otel", version: "1.21.0"},
              spans: [
                {
                  name: "GET /api/get-users",
                  spanId: "aabbccdd",
                  traceId: "11223344",
                  kind: "SPAN_KIND_SERVER",
                  startTimeUnixNano: "1770000000000000000",
                  endTimeUnixNano: "1770000001500000000",
                  attributes: [{key: "http.status_code", value: {intValue: "200"}}],
                  status: {code: "OK"},
                },
              ],
            },
          ],
        },
      ],
    };

    stubEntry(
      {...baseEntry, provider: "provider-trace", type: "trace", message: JSON.stringify(trace)},
      {owner: "built-in", name: "provider-trace", category: "Log", type: "Log"},
    );

    cy.visit("/entries/built-in/e2e-entry");
    cy.wait(["@getEntry", "@getProvider"]);

    cy.contains("GET /api/get-users").should("be.visible");
    cy.contains("casdoor-e2e").should("be.visible");
    // 1.5 s, computed by the string-based nanosecond subtraction
    cy.contains("1.500 s").should("be.visible");

    cy.contains("GET /api/get-users").click();
    cy.get("[role=dialog]").within(() => {
      cy.contains("11223344").should("be.visible");
      cy.contains("aabbccdd").should("be.visible");
      cy.contains("go.opentelemetry.io/otel@1.21.0").should("be.visible");
      cy.contains("http.status_code").should("be.visible");
    });
  });

  it("renders the OpenClaw session graph, its node drawer and the transcript link", () => {
    const graph = {
      stats: {totalNodes: 4, taskCount: 1, toolCallCount: 1, toolResultCount: 1, finalCount: 1, failedCount: 1},
      rawTranscript: true,
      nodes: [
        {id: "n1", kind: "task", summary: "Investigate the flaky login test", timestamp: "2026-02-01T10:00:00.000Z"},
        {id: "n2", kind: "tool_call", tool: "search", query: "retry loop", toolCallId: "tc1", parentId: "n1", timestamp: "2026-02-01T10:00:01.000Z"},
        {id: "n3", kind: "tool_result", tool: "search", ok: false, error: "index unavailable", toolCallId: "tc1", parentId: "n2", timestamp: "2026-02-01T10:00:02.000Z"},
        {id: "n4", kind: "final", summary: "The retry loop swallows the 401", isAnchor: true, timestamp: "2026-02-01T10:00:03.000Z"},
      ],
      edges: [
        {source: "n1", target: "n2"},
        {source: "n2", target: "n3"},
        {source: "n3", target: "n4"},
      ],
    };

    stubEntry(
      {
        ...baseEntry,
        provider: "provider-openclaw",
        type: "session",
        message: JSON.stringify({kind: "task", sessionId: "s1", entryId: "e1"}),
      },
      {owner: "built-in", name: "provider-openclaw", category: "Log", type: "Agent", subType: "OpenClaw"},
    );
    cy.intercept({method: "GET", pathname: "/api/get-openclaw-session-graph"}, ok(graph)).as("getGraph");

    cy.visit("/entries/built-in/e2e-entry");
    cy.wait(["@getEntry", "@getProvider", "@getGraph"]);

    cy.get("[data-testid=openclaw-session-graph]").should("be.visible");
    cy.get(".react-flow__node").should("have.length", 4);
    cy.get(".react-flow__edge").should("have.length", 3);

    // the stats row, including the failed tool result
    cy.get("[data-testid=openclaw-session-graph]").within(() => {
      cy.contains("4").should("exist");
    });

    // clicking a node opens the detail drawer, which pairs the call and result
    cy.contains(".react-flow__node", "retry loop").click();
    cy.get("[role=dialog]").within(() => {
      cy.contains("tool_call").should("be.visible");
      cy.contains("tc1").should("be.visible");
      cy.contains("index unavailable").should("be.visible");
    });
    cy.get("body").type("{esc}");

    // the "Raw JSONL" button routes to the transcript page
    cy.intercept({method: "GET", pathname: "/api/get-openclaw-session-transcript"}, ok({
      fileName: "session-s1.jsonl",
      fileSize: 2048,
      loadedSize: 2048,
      truncated: true,
      content: '{"kind":"task","sessionId":"s1"}\n{"kind":"final","sessionId":"s1"}\n',
    })).as("getTranscript");

    cy.contains("button", "JSONL").click();
    cy.location("pathname").should("eq", "/entries/built-in/e2e-entry/transcript");
    cy.wait("@getTranscript");
    cy.contains("session-s1.jsonl").should("be.visible");
    cy.contains("2.00 KB").should("be.visible");
    cy.contains('"sessionId":"s1"').should("be.visible");
  });

  it("falls back to the raw message when nothing specialised matches", () => {
    stubEntry(
      {...baseEntry, provider: "provider-plain", type: "log", message: "just a line of text"},
      {owner: "built-in", name: "provider-plain", category: "Log", type: "Log"},
    );

    cy.visit("/entries/built-in/e2e-entry");
    cy.wait(["@getEntry", "@getProvider"]);

    cy.get("[data-testid=openclaw-session-graph]").should("not.exist");
    cy.contains("just a line of text").should("be.visible");
  });
});
