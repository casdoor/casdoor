// Per-column search and sorting on the list pages, and the translated enum
// badges — the systemic differences from the antd list pages.

/**
 * Asserts that some request on `alias` matched. React re-fetches on mount more
 * than once in dev, so waiting for "the next call" is not reliable; this checks
 * every call the alias captured instead.
 */
function expectRequest(alias, predicate) {
  cy.get(`${alias}.all`, {timeout: 15000}).should((calls) => {
    expect(calls.some((call) => predicate(call.request.url)), "a matching request was made").to.be.true;
  });
}

describe("List page columns", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("offers a search button on every column the antd page searches", () => {
    cy.visitListPage("/entries");
    // provider was already searchable; type / client IP / user agent were not
    ["provider", "type", "clientIp", "userAgent"].forEach((field) => {
      cy.get(`th[data-column="${field}"] button[aria-label=Search]`).should("exist");
    });
  });

  it("sends the typed term to the backend as field + value", () => {
    cy.intercept({method: "GET", pathname: "/api/get-entries"}).as("getEntries");
    cy.visitListPage("/entries");

    cy.get('th[data-column="clientIp"] button[aria-label=Search]').click();
    cy.focused().type("127.0.0.1{enter}");

    expectRequest("@getEntries", (url) => url.includes("field=clientIp") && url.includes("value=127.0.0.1"));
  });

  it("sorts on a column the antd page sorts but web2 did not", () => {
    cy.intercept({method: "GET", pathname: "/api/get-permissions"}).as("getPermissions");
    cy.visitListPage("/permissions");

    // "users" is a refsColumn, which had neither a sorter nor a search box;
    // the sort control is the first button in the header cell
    cy.get('th[data-column="users"] button').first().click();

    expectRequest("@getPermissions", (url) => url.includes("sortField=users") && url.includes("sortOrder=ascend"));
  });

  it("renders enum values as translated badges, not raw stored text", () => {
    const states = ["Open", "In Progress", "Resolved", "Closed"];
    cy.intercept({method: "GET", pathname: "/api/get-tickets"}, {
      status: "ok",
      msg: "",
      data: states.map((state, i) => ({
        owner: "built-in",
        name: `t${i}`,
        displayName: `t${i}`,
        createdTime: "2026-02-01T10:00:00Z",
        state,
      })),
      data2: states.length,
    }).as("getTickets");

    cy.visitListPage("/tickets");
    cy.wait("@getTickets");
    cy.get("tbody tr").should("have.length", states.length);

    // before, "In Progress" / "Resolved" / "Closed" all rendered the same green
    // badge; each state should now carry its own colour
    const classes = [];
    states.forEach((state) => {
      cy.contains("tbody td div", state)
        .should("be.visible")
        .then(($el) => classes.push($el.attr("class")));
    });
    cy.then(() => {
      expect(new Set(classes).size, "one badge colour per state").to.equal(states.length);
    });
  });
});
