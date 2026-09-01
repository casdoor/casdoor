// Small read-only / locked fields the antd pages have. Each is stubbed so the
// spec does not depend on a database happening to hold a matching object.

const ok = (data) => ({status: "ok", msg: "", data});

describe("Edit page details", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("lists a group's members as links, read-only", () => {
    cy.intercept({method: "GET", pathname: "/api/get-group"}, ok({
      owner: "built-in",
      name: "e2e-group",
      displayName: "e2e group",
      type: "Virtual",
      parentId: "built-in",
      users: ["built-in/alice", "built-in/bob"],
    })).as("getGroup");

    cy.visit("/groups/built-in/e2e-group");
    cy.wait("@getGroup");
    cy.assertEditPageLoaded();

    cy.contains("a", "built-in/alice").should("have.attr", "href", "/users/built-in/alice");
    cy.contains("a", "built-in/bob").should("exist");
  });

  it("shows the organization's user balance read-only", () => {
    cy.intercept({method: "GET", pathname: "/api/get-organization"}, ok({
      owner: "admin",
      name: "e2e-org",
      displayName: "e2e org",
      userBalance: 4321,
      orgBalance: 0,
      accountItems: [],
      tags: [],
      languages: [],
    })).as("getOrganization");

    cy.visit("/organizations/e2e-org");
    cy.wait("@getOrganization");
    cy.assertEditPageLoaded();

    // the balance rows live on the Advanced tab, and Radix unmounts inactive ones
    cy.contains("[role=tab]", "Advanced").click();
    cy.get('input[value="4321"]').should("be.disabled");
  });

  it("locks the organization and name of a plan-created invitation", () => {
    const invitation = {
      owner: "built-in",
      name: "e2e-invitation",
      displayName: "e2e invitation",
      code: "abc",
      quota: 1,
      usedCount: 0,
      state: "Active",
    };

    cy.intercept({method: "GET", pathname: "/api/get-invitation"}, ok({...invitation, tag: "plain"})).as("getPlain");
    cy.visit("/invitations/built-in/e2e-invitation");
    cy.wait("@getPlain");
    cy.assertEditPageLoaded();
    cy.get('input[value="e2e-invitation"]').should("not.be.disabled");

    cy.intercept({method: "GET", pathname: "/api/get-invitation"}, ok({
      ...invitation,
      tag: "auto_created_invitation_for_plan",
    })).as("getFromPlan");
    cy.visit("/invitations/built-in/e2e-invitation");
    cy.wait("@getFromPlan");
    cy.assertEditPageLoaded();
    cy.get('input[value="e2e-invitation"]').should("be.disabled");
  });
});
