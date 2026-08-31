// Regression cover for the Casdoor columns that are `map[string]string` on the
// wire but are edited as a two-column table, and for the ones that look like
// lists but are plain strings. Feeding either straight into a list control took
// the whole edit page down with "items.map is not a function".

const ok = (data) => ({status: "ok", msg: "", data});

describe("Map- and string-typed fields", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("edits LDAP customAttributes, a map, as rows", () => {
    cy.intercept({method: "GET", pathname: "/api/get-ldap"}, ok({
      owner: "built-in",
      id: "e2e-ldap",
      serverName: "e2e ldap",
      host: "ldap.example.com",
      port: 389,
      username: "cn=admin",
      password: "",
      baseDn: "ou=people,dc=example,dc=com",
      filter: "",
      filterFields: ["uid", "mail"],
      defaultGroups: [],
      autoSync: 0,
      customAttributes: {mail: "email", telephoneNumber: "phone"},
    })).as("getLdap");

    cy.visit("/ldap/built-in/e2e-ldap");
    cy.wait("@getLdap");
    cy.assertEditPageLoaded();

    // one row per map entry, key on the left and value on the right
    cy.get('input[value="mail"]').should("exist");
    cy.get('input[value="email"]').should("exist");
    cy.get('input[value="telephoneNumber"]').should("exist");
    cy.get('input[value="phone"]').should("exist");
  });

  it("edits provider httpHeaders, a map, as rows", () => {
    cy.intercept({method: "GET", pathname: "/api/get-provider"}, ok({
      owner: "admin",
      name: "e2e-http-sms",
      displayName: "e2e http sms",
      category: "SMS",
      type: "Custom HTTP SMS",
      method: "POST",
      endpoint: "https://sms.example.com/send",
      httpHeaders: {"X-Api-Key": "secret", "Content-Type": "application/json"},
    })).as("getProvider");

    cy.visit("/providers/admin/e2e-http-sms");
    cy.wait("@getProvider");
    cy.assertEditPageLoaded();

    cy.get('input[value="X-Api-Key"]').should("exist");
    cy.get('input[value="secret"]').should("exist");
    cy.get('input[value="Content-Type"]').should("exist");
  });

  it("edits application ipWhitelist, a string, as a text field", () => {
    cy.intercept({method: "GET", pathname: "/api/get-application"}, ok({
      owner: "admin",
      name: "e2e-app",
      displayName: "e2e app",
      organization: "built-in",
      tags: [],
      ipWhitelist: "10.0.0.0/8",
      redirectUris: [],
      scopes: [{name: "files:read", displayName: "Read Files", description: "Allow reading files"}],
      providers: [],
      signinMethods: [],
      signupItems: [],
      signinItems: [],
      grantTypes: [],
    })).as("getApplication");

    cy.visit("/applications/admin/e2e-app");
    cy.wait("@getApplication");
    cy.assertEditPageLoaded();

    // the whole CIDR sits in one input, not split into tag badges
    cy.get('input[value="10.0.0.0/8"]').should("exist");

    // scopes live on the OIDC/OAuth tab, and Radix unmounts the inactive ones
    cy.contains('[role=tab]', "OIDC/OAuth").click();

    // scopes are objects, so they render as the three-column table
    cy.get('input[value="files:read"]').should("exist");
    cy.get('input[value="Read Files"]').should("exist");
    cy.get('input[value="Allow reading files"]').should("exist");
  });
});
