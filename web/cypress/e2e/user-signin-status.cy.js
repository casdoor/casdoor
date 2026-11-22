describe("User sign-in status", {retries: 0}, () => {
  const rows = [
    {owner: "example", name: "local", displayName: "Local user", isForbidden: false},
    {owner: "example", name: "directory", displayName: "Directory user", azuread: "immutable-directory-id", isForbidden: false},
  ];

  beforeEach(() => {
    cy.intercept("GET", "/api/**", {status: "ok", data: []});
    cy.intercept("GET", "/api/get-account*", {
      status: "ok",
      data: {owner: "example", name: "administrator", displayName: "Administrator", isAdmin: true},
      data2: {owner: "admin", name: "example", displayName: "Example", enableTour: false},
    });
    cy.intercept("GET", "/api/get-organization?*", {status: "ok", data: {owner: "admin", name: "example"}});
    cy.intercept("GET", "/api/get-users?*", (req) => {
      const query = new URL(req.url).searchParams;
      req.alias = `users_${query.get("signInStatus") || "all"}_${query.get("p")}`;
      req.reply({status: "ok", data: rows, data2: 250});
    }).as("users");
    cy.intercept({method: "POST", pathname: "/api/update-user"}, {status: "ok", data: true}).as("updateUser");
    cy.visit("/users", {onBeforeLoad(win) {
      win.localStorage.setItem("language", "en");
    }});
    cy.wait("@users");
  });

  it("defaults to 100 enabled users and resets pagination for Disabled and All", () => {
    cy.get("@users").its("request.query").should("include", {pageSize: "100", signInStatus: "enabled"});
    cy.get('button[aria-label="Next page"]').click();
    cy.wait("@users_enabled_2").its("request.query").should("include", {p: "2"});
    cy.get('[aria-label="Casdoor sign-in status"]').click();
    cy.get('[role="option"]').contains("Disabled").click();
    cy.wait("@users_disabled_1").its("request.query").should("include", {p: "1", signInStatus: "disabled"});
    cy.get('[aria-label="Casdoor sign-in status"]').click();
    cy.get('[role="option"]').contains("All").click();
    cy.wait("@users_all_1").its("request.query").should("include", {signInStatus: ""});
  });

  it("confirms local status updates without allowing changes to directory users", () => {
    cy.get('button[role="switch"][aria-label="Is forbidden: local"]').click();
    cy.get('[role="alertdialog"]').should("be.visible").contains("Cancel").click();
    cy.get("@updateUser.all").should("have.length", 0);
    cy.get('button[role="switch"][aria-label="Is forbidden: local"]').click();
    cy.get('[role="alertdialog"]').contains("OK").click();
    cy.wait("@updateUser").then(({request}) => {
      expect(request.query.columns).to.equal("IsForbidden");
      expect(request.body).to.deep.equal({isForbidden: true});
    });
    cy.wait("@users");
    cy.contains("tr", "directory").find('[role="switch"]').should("be.disabled");
    cy.get('[aria-label="Is forbidden: directory"]').should("not.exist");
  });

  it("does not change the displayed status when an update fails", () => {
    cy.intercept({method: "POST", pathname: "/api/update-user"}, {
      status: "error", msg: "Synthetic update rejected",
    }).as("failedUpdate");
    cy.get('button[role="switch"][aria-label="Is forbidden: local"]').click();
    cy.get('[role="alertdialog"]').contains("OK").click();
    cy.wait("@failedUpdate");
    cy.contains("Synthetic update rejected").should("be.visible");
    cy.get('[aria-label="Is forbidden: local"]').should("have.attr", "aria-checked", "false");
  });

  it("does not offer administrative switches to a non-administrator", () => {
    cy.intercept("GET", "/api/get-account*", {
      status: "ok", data: {owner: "example", name: "reader"},
      data2: {owner: "admin", name: "example", enableTour: false},
    });
    cy.reload();
    cy.contains("tr", "local").find('[role="switch"]').should("be.disabled");
    cy.get('[aria-label="Is forbidden: local"]').should("not.exist");
  });
});
