// After sign-in, an organization that marks an MFA item "Prompted" should ask
// the user to turn it on. Rather than editing the organization, this rewrites
// the `get-account` response on its way through, so the rest of the account and
// organization stay exactly as the backend sent them.
//
// The assertions avoid translated strings: the mfa type ("SMS") is a stored
// value, not a translation, and the rest is structural.

function stubMfaRule(rule) {
  cy.intercept({method: "GET", pathname: "/api/get-account"}, (req) => {
    req.continue((res) => {
      if (res.body?.status !== "ok" || !res.body.data) {
        return;
      }
      res.body.data.multiFactorAuths = [{mfaType: "SMS", enabled: false}];
      res.body.data.mfaItems = [{name: "SMS", rule}];
      if (res.body.data2) {
        res.body.data2.mfaItems = [{name: "SMS", rule}];
      }
    });
  }).as("getAccount");
}

describe("Enable MFA notification", () => {
  it('prompts the user when an mfa item is "Prompted"', () => {
    stubMfaRule("Prompted");
    cy.login();

    // sonner re-renders the toast as it animates in, so re-query rather than
    // holding on to an alias of the first node it mounted
    cy.contains("[data-sonner-toast]", "SMS", {timeout: 20000}).should("be.visible");

    // "Go to enable" is the toast's primary action
    cy.get("[data-sonner-toast] button").last().click();
    cy.location("pathname").should("eq", "/mfa/setup");
    cy.location("search").should("include", "mfaType=SMS");
  });

  it('sends the user straight to the setup wizard when an mfa item is "Required"', () => {
    stubMfaRule("Required");
    cy.visit("/", {
      onBeforeLoad(win) {
        win.localStorage.setItem("isTourVisible", "false");
      },
    });
    cy.get("#username").type("admin");
    cy.get("#password").type("123");
    cy.get("form button[type=submit]").click();

    cy.location("pathname", {timeout: 20000}).should("eq", "/mfa/setup");
    cy.location("search").should("include", "mfaType=SMS");
  });

  it("stays quiet when no mfa item asks for anything", () => {
    stubMfaRule("Optional");
    cy.login();

    cy.get("h1", {timeout: 20000}).should("be.visible");
    cy.get("[data-sonner-toast]").should("not.exist");
  });
});
