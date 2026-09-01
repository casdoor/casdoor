// The sub-tables of the application and organization edit pages: the columns
// and option lists the antd tables offer.

const ok = (data) => ({status: "ok", msg: "", data});

const application = {
  owner: "admin",
  name: "e2e-app",
  displayName: "e2e app",
  organization: "built-in",
  tags: [],
  redirectUris: [],
  scopes: [],
  signinMethods: [],
  signupItems: [],
  signinItems: [],
  grantTypes: [],
  tokenAttributes: [],
  providers: [],
};

/** the provider list the table resolves each row against */
const providers = [
  {owner: "admin", name: "p-github", category: "OAuth", type: "GitHub"},
  {owner: "admin", name: "p-sms", category: "SMS", type: "Twilio SMS"},
  {owner: "admin", name: "p-captcha", category: "Captcha", type: "Default"},
  {owner: "admin", name: "p-google", category: "OAuth", type: "Google"},
];

function openApplication(overrides = {}) {
  cy.intercept({method: "GET", pathname: "/api/get-providers"}, {
    status: "ok", msg: "", data: providers, data2: providers.length,
  }).as("getProviders");
  cy.intercept({method: "GET", pathname: "/api/get-application"}, ok({...application, ...overrides})).as("getApp");
  cy.visit("/applications/admin/e2e-app");
  cy.wait("@getApp");
  cy.assertEditPageLoaded();
}

describe("Application providers table", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("shows the category, type and the columns each provider kind needs", () => {
    openApplication({
      providers: [
        {name: "p-github", provider: providers[0], canSignUp: true, rule: "None", bindingRule: []},
        {name: "p-sms", provider: providers[1], rule: "None", countryCodes: []},
      ],
    });
    cy.contains('[role=tab]', "Providers").click();

    ["Category", "Country/Region", "Binding rule", "Signup group", "Rule"].forEach((title) => {
      cy.contains("th", title).should("exist");
    });

    // the category links out to the provider itself
    cy.get('a[href="/providers/admin/p-github"]').should("contain.text", "OAuth");
    cy.contains("td", "GitHub").should("exist");
  });

  it("offers the rule options that match the provider", () => {
    openApplication({
      providers: [{name: "p-captcha", provider: providers[2], rule: "None"}],
    });
    cy.contains('[role=tab]', "Providers").click();

    // a captcha provider's rule is None / Dynamic / Always / Internet-Only
    cy.contains("tbody tr", "Captcha").find("[role=combobox]").last().click();
    cy.contains("[role=option]", "Internet-Only").should("be.visible");
    cy.contains("[role=option]", "Dynamic").should("be.visible");
  });
});

describe("Application signup items table", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("has the Type, Custom CSS and Options columns", () => {
    openApplication({signupItems: [{name: "Username", visible: true, required: true, rule: "None", type: "Input"}]});
    cy.contains('[role=tab]', "UI Customization").click();

    ["Type", "Custom CSS", "Options"].forEach((title) => {
      cy.contains("th", title).should("exist");
    });
  });

  it("picks the rule options from the item name", () => {
    openApplication({
      signupItems: [{name: "Display name", visible: true, required: true, rule: "None", type: "Input"}],
    });
    cy.contains('[role=tab]', "UI Customization").click();

    // "Display name" offers None / Real name / First, last — not the old flat list
    cy.contains("tbody tr", "Display name").find("[role=combobox]").eq(2).click();
    cy.contains("[role=option]", "First, last").should("be.visible");
  });
});

describe("Application token attributes table", () => {
  it("turns the value into a user-field picker for an Existing Field attribute", () => {
    cy.openConsole();
    openApplication({
      // the table only shows for a custom JWT
      tokenFormat: "JWT-Custom",
      tokenAttributes: [{name: "roles", category: "Existing Field", value: "Roles", type: "Array"}],
    });
    cy.contains('[role=tab]', "OIDC/OAuth").click();

    // the cells are inputs and selects, so reach the row through its name field
    cy.get('input[value="roles"]').closest("tr").within(() => {
      // category and value are both pickers now, not free text
      cy.get("[role=combobox]").should("have.length", 3);
      cy.contains("Existing Field").should("exist");
      cy.contains("Roles").should("exist");
    });
  });
});

describe("Organization MFA items", () => {
  it("allows only one factor to be Required", () => {
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-organization"}, ok({
      owner: "admin",
      name: "e2e-org",
      displayName: "e2e org",
      accountItems: [],
      tags: [],
      languages: [],
      mfaItems: [
        {name: "Email", rule: "Required"},
        {name: "SMS", rule: "Optional"},
      ],
    })).as("getOrganization");

    cy.visit("/organizations/e2e-org");
    cy.wait("@getOrganization");
    cy.assertEditPageLoaded();
    // the MFA items table lives on the password-type tab
    cy.contains('[role=tab]', "Password type").click();

    // second row is SMS, which is Optional; making it Required too is refused
    cy.get("tbody tr").eq(1).find("[role=combobox]").last().click();
    cy.contains("[role=option]", "Required").click();

    cy.contains("Only 1 MFA method can be required").should("be.visible");
  });
});

describe("Application signin items table", () => {
  it("has a Custom CSS column and can add a custom item", () => {
    cy.openConsole();
    openApplication({
      signinItems: [{name: "Logo", visible: true, label: "", placeholder: "", rule: "None"}],
    });
    cy.contains('[role=tab]', "UI Customization").click();

    cy.contains("th", "Custom CSS").should("exist");

    cy.get("tbody tr").its("length").then((before) => {
      cy.contains("button", "Add custom item").click();
      cy.get("tbody tr").should("have.length", before + 1);
    });
  });
});
