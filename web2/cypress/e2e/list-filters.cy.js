// The header filter menus, the 403 page and the record detail drawer — the
// list-page behaviour web2 was missing next to the antd frontend.

/** One query parameter of a captured request URL. */
function param(url, name) {
  return new URL(url, "http://localhost").searchParams.get(name) ?? "";
}

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

describe("List page column filters", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("offers a filter menu on the columns the antd table filters", () => {
    cy.visitListPage("/records");
    cy.get('th[data-column="method"] button[aria-label=Filter]').should("exist");
    // a column the antd table only searches keeps its search button and no filter
    cy.get('th[data-column="requestUri"] button[aria-label=Filter]').should("not.exist");
  });

  it("sends the picked value to the backend as field + value", () => {
    cy.intercept({method: "GET", pathname: "/api/get-records"}).as("getRecords");
    cy.visitListPage("/records");

    cy.get('th[data-column="method"] button[aria-label=Filter]').click();
    cy.get('[data-column-filter=method]').contains("[role=menuitemradio]", "POST").click();

    expectRequest("@getRecords", (url) => param(url, "field") === "method" && param(url, "value") === "POST");
  });

  it("groups the provider types under their category, like antd's two-level menu", () => {
    cy.visitListPage("/providers");
    cy.get('th[data-column="type"] button[aria-label=Filter]').click();
    // the category is a heading over its own types, not a pickable row
    cy.get('[data-column-filter=type] [role=group][aria-label=OAuth]').within(() => {
      cy.contains("[role=menuitemradio]", "GitHub").should("exist");
    });
  });

  it("clears the filter again", () => {
    cy.intercept({method: "GET", pathname: "/api/get-syncers"}).as("getSyncers");
    cy.visitListPage("/syncers");

    cy.get('th[data-column="type"] button[aria-label=Filter]').click();
    cy.get('[data-column-filter=type]').contains("[role=menuitemradio]", "LDAP").click();
    expectRequest("@getSyncers", (url) => param(url, "field") === "type" && param(url, "value") === "LDAP");

    cy.get('th[data-column="type"] button[aria-label=Filter]').click();
    cy.get('[data-column-filter=type]').contains("button", "Reset").click();
    // the last call carries no field/value at all
    cy.get("@getSyncers.all", {timeout: 15000}).should((calls) => {
      const last = calls[calls.length - 1].request.url;
      expect(param(last, "field"), "field").to.eq("");
      expect(param(last, "value"), "value").to.eq("");
    });
  });
});

describe("Unauthorized console pages", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("shows the 403 page when the backend refuses the read", () => {
    cy.intercept({method: "GET", pathname: "/api/get-organizations"}, {
      status: "error",
      msg: "Unauthorized operation",
      data: null,
      data2: null,
    }).as("getOrganizations");

    cy.visitPath("/organizations");
    cy.contains("403").should("be.visible");
    cy.contains("you do not have permission").should("be.visible");
    // the table is replaced, not merely emptied
    cy.get("table").should("not.exist");
  });

  it("shows it on an edit page too", () => {
    cy.intercept({method: "GET", pathname: "/api/get-application"}, {
      status: "error",
      msg: "Unauthorized operation",
      data: null,
      data2: null,
    }).as("getApplication");

    cy.visit("/applications/admin/app-built-in");
    cy.contains("403").should("be.visible");
  });
});

describe("Record detail", () => {
  it("opens the full record in a drawer", () => {
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-records"}, {
      status: "ok",
      msg: "",
      data: [
        {
          id: 42,
          owner: "built-in",
          name: "admin",
          organization: "built-in",
          clientIp: "127.0.0.1",
          createdTime: "2026-02-01T10:00:00Z",
          user: "admin",
          method: "POST",
          requestUri: "/api/update-user",
          language: "en",
          statusCode: 200,
          action: "update-user",
          isTriggered: true,
          response: "{\"status\":\"ok\"}",
          object: "{\"name\":\"admin\"}",
        },
      ],
      data2: 1,
    }).as("getRecords");

    cy.visitListPage("/records");
    cy.contains("tbody tr button", "View").click();

    cy.contains("Request URI").should("be.visible");
    cy.contains("/api/update-user").should("be.visible");
    // the response and the object are read-only editors, not plain cells
    cy.get(".cm-editor").should("have.length", 2);
  });
});

describe("Transaction recharge", () => {
  it("adds the transaction first, then opens it under the Recharge title", () => {
    cy.openConsole();
    cy.intercept({method: "POST", pathname: "/api/add-transaction"}, {
      status: "ok",
      msg: "",
      data: "recharge-1",
      data2: null,
    }).as("addTransaction");
    cy.intercept({method: "GET", pathname: "/api/get-transaction"}, {
      status: "ok",
      msg: "",
      data: {
        owner: "built-in",
        name: "recharge-1",
        createdTime: "2026-02-01T10:00:00Z",
        category: "Recharge",
        user: "admin",
        tag: "User",
        amount: 100,
        currency: "USD",
        state: "Paid",
      },
      data2: null,
    }).as("getTransaction");

    cy.visitListPage("/transactions");
    cy.contains("button", "Recharge").click();

    cy.wait("@addTransaction").its("request.body.category").should("eq", "Recharge");
    cy.location("pathname").should("eq", "/transactions/built-in/recharge-1");
    cy.get("h1").should("contain.text", "Recharge");
  });
});
