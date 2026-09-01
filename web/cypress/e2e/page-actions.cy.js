// The per-page buttons, guards and validations the antd frontend has.

const ok = (data) => ({status: "ok", msg: "", data});

const application = {
  owner: "admin",
  name: "e2e-app",
  displayName: "e2e app",
  organization: "built-in",
  termsOfUse: "",
  footerHtml: "",
  tags: [],
  redirectUris: [],
  scopes: [],
  providers: [],
  signinMethods: [],
  signupItems: [],
  signinItems: [],
  grantTypes: [],
};

describe("Application list", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("duplicates an application and opens the copy", () => {
    cy.intercept({method: "GET", pathname: "/api/get-applications"}, {
      status: "ok", msg: "", data: [application], data2: 1,
    }).as("getApplications");
    cy.intercept({method: "POST", pathname: "/api/add-application"}, ok("Affected")).as("addApplication");
    cy.intercept({method: "GET", pathname: "/api/get-application"}, ok(application)).as("getApplication");

    cy.visitListPage("/applications");
    cy.wait("@getApplications");
    cy.contains("tbody tr button", "Duplicate").click();

    cy.wait("@addApplication").its("request.body").should((body) => {
      expect(body.name).to.match(/^e2e-app_/);
      expect(body.displayName).to.match(/^Copy Application - e2e-app_/);
      // the copy has to earn its own credentials
      expect(body.clientId).to.equal("");
      expect(body.clientSecret).to.equal("");
    });
    cy.location("pathname").should("match", /^\/applications\/built-in\/e2e-app_/);
  });
});

describe("Application edit", () => {
  beforeEach(() => {
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-application"}, ok(application)).as("getApplication");
    cy.visit("/applications/admin/e2e-app");
    cy.wait("@getApplication");
    cy.assertEditPageLoaded();
  });

  // FormRow renders <label> and the control as siblings inside one grid cell
  const nameInput = () => cy.contains("label", /^\s*Name\s*$/).parent().find("input").first();

  it("rejects the characters a name may not contain", () => {
    // like the antd page, the check runs per keystroke: the offending character
    // is dropped and the value is left as it was
    nameInput().should("have.value", "e2e-app").type("/");
    cy.contains("Invalid characters in application name").should("be.visible");
    nameInput().should("have.value", "e2e-app");
  });

  it("offers an upload for the Terms of Use HTML", () => {
    // Terms of Use lives on the security tab
    cy.contains('[role=tab]', "Security").click();
    cy.contains("button", "Click to Upload").should("be.visible");
  });

  it("resets the footer HTML to the default and to empty", () => {
    cy.contains('[role=tab]', "UI Customization").click();
    cy.contains("button", "Reset to Default").click();
    cy.contains("Powered by").should("be.visible");
    cy.contains("button", "Reset to Empty").click();
    cy.contains("#footer").should("be.visible");
  });
});

describe("Save-time validation", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("refuses a permission with no users and no roles", () => {
    cy.intercept({method: "GET", pathname: "/api/get-permission"}, ok({
      owner: "built-in",
      name: "e2e-permission",
      displayName: "e2e permission",
      users: [],
      roles: [],
      groups: [],
      domains: [],
      resources: ["app-built-in"],
      actions: ["Read"],
      resourceType: "Application",
      effect: "Allow",
      state: "Approved",
      submitter: "admin",
    })).as("getPermission");
    cy.intercept({method: "POST", pathname: "/api/update-permission"}).as("updatePermission");

    cy.visit("/permissions/built-in/e2e-permission");
    cy.wait("@getPermission");
    cy.assertEditPageLoaded();

    cy.contains("button", "Save").first().click();
    cy.contains("users and roles cannot be empty").should("be.visible");
    cy.get("@updatePermission.all").should("have.length", 0);
  });

  it("refuses a product with no currency", () => {
    cy.intercept({method: "GET", pathname: "/api/get-product"}, ok({
      owner: "built-in",
      name: "e2e-product",
      displayName: "e2e product",
      currency: "",
      providers: [],
      returnUrl: "",
      state: "Published",
    })).as("getProduct");
    cy.intercept({method: "POST", pathname: "/api/update-product"}).as("updateProduct");

    cy.visit("/products/built-in/e2e-product");
    cy.wait("@getProduct");
    cy.assertEditPageLoaded();

    cy.contains("button", "Save").first().click();
    cy.contains("Please select a currency").should("be.visible");
    cy.get("@updateProduct.all").should("have.length", 0);
  });
});

describe("Groups", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("blocks deleting a group that still has subgroups", () => {
    cy.intercept({method: "GET", pathname: "/api/get-groups"}, {
      status: "ok",
      msg: "",
      data: [
        {owner: "built-in", name: "parent", displayName: "parent", type: "Virtual", haveChildren: true},
        {owner: "built-in", name: "leaf", displayName: "leaf", type: "Virtual", haveChildren: false},
      ],
      data2: 2,
    }).as("getGroups");

    cy.visitListPage("/groups");
    cy.wait("@getGroups");

    cy.contains("tbody tr", "parent").contains("button", "Delete").should("be.disabled");
    cy.contains("tbody tr", "leaf").contains("button", "Delete").should("not.be.disabled");
  });

  it('clears the selection from the tree with "Show all"', () => {
    cy.visitPath("/trees/built-in");
    cy.contains("button", "Show all").should("be.visible").click();
    cy.location("pathname").should("eq", "/trees/built-in");
  });
});

describe("Webhook preview", () => {
  it("shows the payload a webhook would post", () => {
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-webhook"}, ok({
      owner: "built-in",
      name: "e2e-webhook",
      url: "https://example.com/hook",
      method: "POST",
      contentType: "application/json",
      events: ["login"],
      isUserExtended: false,
      headers: [],
      tokenFields: [],
      objectFields: [],
      isEnabled: true,
    })).as("getWebhook");

    cy.visit("/webhooks/e2e-webhook");
    cy.wait("@getWebhook");
    cy.assertEditPageLoaded();

    // CodeMirror clips its own content, so assert it rendered rather than that
    // the line happens to be scrolled into view
    cy.contains("/api/add-application").should("exist");
  });
});

describe("LDAP sync", () => {
  beforeEach(() => {
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-ldap-users"}, ok({
      users: [
        {uuid: "u1", cn: "alice", uid: "alice", uidNumber: "1001", groupId: "g-eng", email: "a@e.com", phone: "1", address: "x"},
        {uuid: "u2", cn: "bob", uid: "bob", uidNumber: "1002", groupId: "g-ops", email: "b@e.com", phone: "2", address: "y"},
      ],
      existUuids: ["u1"],
    })).as("getLdapUsers");
    cy.visit("/ldap/sync/built-in/e2e-ldap");
    cy.wait("@getLdapUsers");
  });

  it("marks who is already synced and links them to their user", () => {
    cy.contains("tbody tr", "alice").within(() => {
      cy.contains("synced").should("be.visible");
      cy.get('a[href="/users/built-in/alice"]').should("exist");
    });
    cy.contains("tbody tr", "bob").contains("unsynced").should("be.visible");
  });

  it("shows the Group ID column", () => {
    cy.contains("th", "Group ID").should("be.visible");
    cy.contains("tbody td", "g-eng").should("be.visible");
    cy.contains("tbody td", "g-ops").should("be.visible");
  });

  it("refuses to sync with nothing selected", () => {
    cy.intercept({method: "POST", pathname: "/api/sync-ldap-users"}).as("syncUsers");
    cy.contains("button", "Sync").click();
    cy.contains("Please select at least 1 user first").should("be.visible");
    cy.get("@syncUsers.all").should("have.length", 0);
  });

  it("reports the users the backend skipped or could not add", () => {
    cy.intercept({method: "POST", pathname: "/api/sync-ldap-users"}, ok({
      exist: [{cn: "alice"}],
      failed: [{cn: "bob"}],
    })).as("syncUsers");

    cy.contains("tbody tr", "bob").find('button[role=checkbox]').click();
    cy.contains("button", "Sync").click();
    cy.wait("@syncUsers");

    cy.contains("User already exists").should("be.visible");
    cy.contains("Failed to sync").should("be.visible");
  });
});

describe("Pricing", () => {
  it("copies the public pricing page URL", () => {
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-pricing"}, ok({
      owner: "built-in",
      name: "e2e-pricing",
      displayName: "e2e pricing",
      plans: [],
      isEnabled: true,
    })).as("getPricing");

    cy.visit("/pricings/built-in/e2e-pricing");
    cy.wait("@getPricing");
    cy.assertEditPageLoaded();

    cy.contains("button", "Copy pricing page URL").click();
    cy.contains("Copied to clipboard successfully").should("be.visible");
  });
});
