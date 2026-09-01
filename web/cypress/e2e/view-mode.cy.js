// A non-admin may look at the billing objects but not change them, and the edit
// page titles itself after the mode. Both come from the antd frontend.

const ok = (data) => ({status: "ok", msg: "", data});

const plan = {
  owner: "built-in",
  name: "e2e-plan",
  displayName: "e2e plan",
  price: 10,
  currency: "USD",
  period: "Monthly",
  paymentProviders: [],
  isEnabled: true,
};

/** Rewrites `get-account` so the console believes the signed-in user is not an admin. */
function asNonAdmin() {
  cy.intercept({method: "GET", pathname: "/api/get-account"}, (req) => {
    req.continue((res) => {
      if (res.body?.status === "ok" && res.body.data) {
        // isLocalAdminUser is `isAdmin || owner === "built-in"`, so both have to go
        res.body.data.isAdmin = false;
        res.body.data.owner = "e2e-org";
        res.body.data.type = "normal-user";
      }
    });
  }).as("getAccount");
}

describe("Read-only view mode", () => {
  it("gives an admin the editable page", () => {
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-plans"}, {
      status: "ok", msg: "", data: [plan], data2: 1,
    }).as("getPlans");

    cy.visitListPage("/plans");
    cy.wait("@getPlans");
    cy.get("#add-button").should("not.be.disabled");
    cy.contains("tbody tr button", "Edit").should("exist");
  });

  it("gives a non-admin View, a locked form and no way to save", () => {
    asNonAdmin();
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-plan"}, ok(plan)).as("getPlan");
    cy.intercept({method: "GET", pathname: "/api/get-plans"}, {
      status: "ok", msg: "", data: [plan], data2: 1,
    }).as("getPlans");

    cy.visitListPage("/plans");
    cy.wait("@getPlans");

    // Add and Delete are out of reach
    cy.get("#add-button").should("be.disabled");
    cy.contains("tbody tr button", "Delete").should("be.disabled");

    // the row action reads "View" and opens the read-only page
    cy.contains("tbody tr button", "View").click();
    cy.wait("@getPlan");
    cy.assertEditPageLoaded();

    cy.contains("h1", "View Plan").should("be.visible");
    cy.contains("button", "Save").should("not.exist");
    cy.contains("button", "Save & Exit").should("not.exist");
    cy.get("input").each(($input) => {
      cy.wrap($input).should("be.disabled");
    });
  });
});

describe("Edit page titles", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it('says "New X" while adding', () => {
    cy.visitListPage("/plans");
    cy.get("#add-button").click();
    cy.contains("h1", "New Plan").should("be.visible");
  });

  it('says "Edit X" for an existing object', () => {
    cy.intercept({method: "GET", pathname: "/api/get-plan"}, ok(plan)).as("getPlan");
    cy.visit("/plans/built-in/e2e-plan");
    cy.wait("@getPlan");
    cy.contains("h1", "Edit Plan").should("be.visible");
  });
});
