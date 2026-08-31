// Login-page surfaces the antd frontend has: the organization choice box, the
// "continue as the account you are already signed in as" panel, the device
// approval header, and the translated signin-method tabs.

const ok = (data) => ({status: "ok", msg: "", data});

const baseApplication = {
  owner: "admin",
  name: "e2e-app",
  displayName: "e2e app",
  organization: "built-in",
  enablePassword: true,
  enableSignUp: true,
  signinMethods: [{name: "Password", rule: "All"}],
  signupItems: [],
  providers: [],
  organizationObj: {owner: "admin", name: "built-in", passwordOptions: [], countryCodes: ["US"]},
};

function stubApplication(overrides = {}) {
  const body = ok({...baseApplication, ...overrides});
  // /login/:owner loads the organization's default application, the other
  // shapes go through get-application / get-app-login
  cy.intercept({method: "GET", pathname: "/api/get-default-application"}, body).as("getDefaultApp");
  cy.intercept({method: "GET", pathname: "/api/get-application"}, body).as("getApp");
  cy.intercept({method: "GET", pathname: "/api/get-app-login"}, body).as("getAppLogin");
}

function visitLogin(path = "/login/built-in") {
  cy.visit(path, {
    onBeforeLoad(win) {
      win.localStorage.setItem("isTourVisible", "false");
      win.localStorage.setItem("language", "en");
    },
  });
}

describe("Organization choice", () => {
  it('asks the visitor to pick an organization when orgChoiceMode is "Select"', () => {
    stubApplication({orgChoiceMode: "Select"});
    visitLogin();

    cy.contains("Please select an organization to sign in").should("be.visible");
    // the credential form is not offered until an organization is chosen
    cy.get("#password").should("not.exist");
  });

  it('asks the visitor to type one when orgChoiceMode is "Input"', () => {
    stubApplication({orgChoiceMode: "Input"});
    visitLogin();

    cy.contains("Please type an organization to sign in").should("be.visible");
    cy.get("input").first().type("built-in");
    cy.contains("button", "Confirm").click();
    // choosing one reloads the page with the box turned off
    cy.location("search").should("include", "orgChoiceMode=None");
    cy.location("pathname").should("eq", "/login/built-in");
  });

  it("skips the box once an organization has been chosen", () => {
    stubApplication({orgChoiceMode: "Select"});
    visitLogin("/login/built-in?orgChoiceMode=None");

    cy.get("#password").should("exist");
    cy.contains("Please select an organization to sign in").should("not.exist");
  });

  it("never shows the box on a plain application without the setting", () => {
    stubApplication({orgChoiceMode: "None"});
    visitLogin();
    cy.get("#password").should("exist");
  });
});

describe("Signin method tabs", () => {
  it("translates the LDAP and WebAuthn tab labels", () => {
    stubApplication({
      enableWebAuthn: true,
      signinMethods: [
        {name: "Password", rule: "All"},
        {name: "LDAP", rule: "None"},
        {name: "WebAuthn", rule: "None"},
      ],
      organizationObj: {...baseApplication.organizationObj, ldaps: [{id: "l1"}]},
    });
    visitLogin();

    cy.get("[role=tab]").should("contain.text", "LDAP");
    cy.get("[role=tab]").should("contain.text", "WebAuthn");
  });
});

describe("Continue as the current account", () => {
  it("offers the signed-in account before the form on an OAuth request", () => {
    cy.openConsole();
    stubApplication();
    cy.intercept({method: "POST", pathname: "/api/login"}).as("login");

    // an authorize request is not a plain /login, so the console redirect does not kick in
    visitLogin("/login/oauth/authorize?client_id=e2e&response_type=code&redirect_uri=https%3A%2F%2Fexample.com&scope=read&state=s");

    cy.contains("Continue with").should("be.visible");
    cy.contains("button", "admin").should("be.visible").click();

    // the session cookie identifies the user, so no credentials are posted
    cy.wait("@login").its("request.body").should((body) => {
      expect(body.application).to.equal("e2e-app");
      expect(body.username).to.be.undefined;
      expect(body.password).to.be.undefined;
    });
  });

  it("does not offer it to an anonymous visitor", () => {
    stubApplication();
    visitLogin();
    cy.get("#password").should("exist");
    cy.contains("Continue with").should("not.exist");
  });
});

describe("Device approval", () => {
  it("names the application and the confirmation code", () => {
    stubApplication();
    visitLogin("/login/oauth/device/ABCD-1234");

    cy.contains("Approve sign-in on this device").should("be.visible");
    cy.contains("e2e app").should("be.visible");
    cy.contains("ABCD-1234").should("be.visible");
  });
});
