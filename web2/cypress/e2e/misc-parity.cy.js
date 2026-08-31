// Smaller surfaces the antd frontend has: the About card, the column-search
// "Filter" button, the MFA recovery wording and the consent footnote.

describe("System info", () => {
  it("shows the About Casdoor card", () => {
    cy.openConsole();
    cy.visitPath("/sysinfo");

    cy.get("#about-card").scrollIntoView().within(() => {
      cy.contains("Identity and Access Management").should("exist");
      cy.get('a[href="https://github.com/casdoor/casdoor"]').should("exist");
      cy.get('a[href="https://casdoor.org"]').should("exist");
      cy.contains("Get in Touch!").should("exist");
    });
  });
});

describe("Column search", () => {
  it('applies the term without closing the popover on "Filter"', () => {
    cy.openConsole();
    cy.intercept({method: "GET", pathname: "/api/get-entries"}).as("getEntries");
    cy.visitListPage("/entries");

    cy.get('th[data-column="clientIp"] button[aria-label=Search]').click();
    cy.focused().type("10.0.0.1");
    cy.contains("button", "Filter").click();

    // the popover stays open so the term can be refined
    cy.contains("button", "Filter").should("be.visible");
    cy.get("@getEntries.all", {timeout: 15000}).should((calls) => {
      expect(calls.some((call) => call.request.url.includes("value=10.0.0.1"))).to.be.true;
    });
  });
});

describe("MFA recovery", () => {
  it("offers the recovery code by name and titles the panel as antd does", () => {
    let calls = 0;
    cy.intercept({method: "POST", pathname: "/api/login"}, (req) => {
      calls += 1;
      req.reply(calls === 1
        ? {status: "ok", msg: "", data: "NextMfa", data2: [{mfaType: "app", mfaRememberInHours: 12, isPreferred: true}]}
        : {status: "ok", msg: "", data: "", data2: null});
    }).as("login");

    cy.visit("/", {
      onBeforeLoad(win) {
        win.localStorage.setItem("isTourVisible", "false");
        win.localStorage.setItem("language", "en");
      },
    });
    cy.get("#username").type("admin");
    cy.get("#password").type("123");
    cy.get("form button[type=submit]").click();
    cy.wait("@login");

    cy.contains("button", "Use a recovery code").click();
    cy.contains("Multi-factor recover").should("be.visible");
    cy.get("#recoveryCode").should("be.visible");
    cy.contains("button", "Use SMS verification code").should("be.visible");
  });
});

describe("Consent page", () => {
  it("explains what Allow means", () => {
    cy.intercept({method: "GET", pathname: "/api/get-application"}, {
      status: "ok",
      msg: "",
      data: {
        owner: "admin",
        name: "e2e-app",
        displayName: "e2e app",
        organization: "built-in",
        signupItems: [],
        signinMethods: [],
        providers: [],
        organizationObj: {owner: "admin", name: "built-in"},
      },
    }).as("getApplication");

    cy.openConsole();
    cy.visit("/consent/e2e-app?client_id=e2e&response_type=code&redirect_uri=https%3A%2F%2Fexample.com&scope=read&state=s");

    cy.contains("you allow this app to use your information").should("be.visible");
  });
});
