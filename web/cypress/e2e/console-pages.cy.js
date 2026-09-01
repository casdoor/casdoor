// A structural sweep of the console: every list page renders, and every "Add"
// lands on an edit page that finishes loading.
//
// This is the cheapest way to catch a crashing edit page, because the object
// "Add" builds (`src/pages/defaults.ts`) has every field at its real type — so
// a control bound to the wrong shape blows up here rather than in production.
//
// Nothing is written: none of these list pages POSTs on "Add"; the new object
// travels to the edit page through the router state and is only saved when the
// user presses Save.

const listPages = [
  "/adapters",
  "/agents",
  "/applications",
  "/certs",
  "/coupons",
  "/enforcers",
  "/entries",
  "/forms",
  "/groups",
  "/invitations",
  "/keys",
  "/models",
  "/orders",
  "/organizations",
  "/payments",
  "/permissions",
  "/plans",
  "/pricings",
  "/products",
  "/providers",
  "/roles",
  "/rules",
  "/servers",
  "/sites",
  "/subscriptions",
  "/syncers",
  "/tickets",
  "/tokens",
  "/transactions",
  "/users",
  "/webhooks",
];

// list pages that have no "Add"
const readOnlyPages = ["/records", "/resources", "/sessions", "/verifications", "/webhook-events"];

// pages that are not lists at all
const otherPages = ["/", "/apps", "/shortcuts", "/sysinfo", "/account"];

describe("Console pages", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  otherPages.forEach((path) => {
    it(`renders ${path}`, () => {
      cy.visitPath(path);
      cy.get("h1", {timeout: 20000}).should("be.visible");
    });
  });

  readOnlyPages.forEach((path) => {
    it(`renders the ${path} list`, () => {
      cy.visitListPage(path);
    });
  });

  listPages.forEach((path) => {
    it(`renders the ${path} list and its Add page`, () => {
      cy.visitListPage(path);
      cy.get("#add-button", {timeout: 20000}).click();
      cy.location("pathname").should("not.eq", path);
      cy.assertEditPageLoaded();
    });
  });
});
