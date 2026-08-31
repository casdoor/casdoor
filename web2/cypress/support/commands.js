// Custom commands shared by the web2 e2e specs. The antd frontend's equivalent
// lives in `web/cypress/support/commands.js`; the selectors differ because this
// frontend is shadcn/ui, so it targets the ids the app renders rather than antd
// class names.

const ADMIN = {
  application: "app-built-in",
  organization: "built-in",
  username: "admin",
  password: "123",
  autoSignin: true,
  type: "login",
};

const selector = {
  username: "#username",
  password: "#password",
  loginButton: "form button[type=submit]",
};

/**
 * Puts the app in a deterministic state before it boots: no product tour popover
 * covering what a test clicks, and English so specs can assert on copy whatever
 * the machine's browser language is.
 */
function prepareApp(win) {
  win.localStorage.setItem("isTourVisible", "false");
  win.localStorage.setItem("language", "en");
}

/**
 * Signs in through the form. Use this when the sign-in flow itself is under
 * test — it is also the only path that leaves `location.state.from === "/login"`
 * behind, which is what triggers the "enable MFA" prompt.
 */
Cypress.Commands.add("login", () => {
  cy.visit("/", {onBeforeLoad: prepareApp});
  cy.get(selector.username).type(ADMIN.username);
  cy.get(selector.password).type(ADMIN.password);
  cy.get(selector.loginButton).click();
  cy.location("pathname", {timeout: 20000}).should("eq", "/");
});

/**
 * Signs in over the API and caches the session, so a spec with many tests pays
 * for the round trip once. Use this whenever the test is about a console page
 * rather than about signing in.
 */
Cypress.Commands.add("openConsole", () => {
  cy.session(
    "admin",
    () => {
      cy.request({method: "POST", url: "/api/login", body: ADMIN})
        .its("body.status")
        .should("eq", "ok");
    },
    {cacheAcrossSpecs: true},
  );
  // localStorage is cleared between tests, so this has to be set per test
  cy.visit("/", {onBeforeLoad: prepareApp});
  cy.location("pathname", {timeout: 20000}).should("eq", "/");
});

/** Visits a console page and asserts the router landed on it. */
Cypress.Commands.add("visitPath", (path) => {
  cy.visit(path);
  cy.location("pathname").should("eq", path);
});

/**
 * Visits a list page and asserts it actually rendered. Every list page is a
 * `CrudListPage`, so it always has the `PageHeader` title and the `DataTable`,
 * whether or not the organization has any rows — which keeps this portable
 * across databases while still catching a blank page or a crash.
 */
Cypress.Commands.add("visitListPage", (path) => {
  cy.visitPath(path);
  cy.get("h1", {timeout: 20000}).should("be.visible");
  cy.get("table", {timeout: 20000}).should("exist");
});

/**
 * Asserts an edit page finished loading rather than sitting on its spinner.
 * While the object is still being fetched the page is only a `<Loading />`, so
 * the `EditPageShell` heading plus at least one field tells the two apart. Both
 * checks are structural, so they do not depend on the account's language.
 */
Cypress.Commands.add("assertEditPageLoaded", () => {
  cy.get("h1", {timeout: 20000}).should("be.visible");
  cy.get("input, textarea, [role=combobox]").should("have.length.greaterThan", 0);
});

/**
 * Opens the first row of the list currently on screen through its name link and
 * waits for the edit page to render.
 *
 * Following the link instead of hard-coding an object id keeps the spec working
 * on any database — and covers the list page's own links, which a hard-coded
 * URL never would.
 */
Cypress.Commands.add("openFirstRow", () => {
  cy.get("tbody tr", {timeout: 20000}).should("have.length.greaterThan", 0);
  cy.get("tbody tr").first().find("a").first().click();
  cy.assertEditPageLoaded();
});
