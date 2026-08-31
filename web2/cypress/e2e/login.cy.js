describe("Login test", () => {
  const selector = {
    username: "#username",
    password: "#password",
    loginButton: "form button[type=submit]",
  };
  const loginBody = (password) => ({
    application: "app-built-in",
    organization: "built-in",
    username: "admin",
    password,
    autoSignin: true,
    type: "login",
  });

  it("Login succeeded", () => {
    cy.request({method: "POST", url: "/api/login", body: loginBody("123")}).then((response) => {
      expect(response).property("body").property("status").to.equal("ok");
    });
  });

  it("ui Login succeeded", () => {
    cy.visit("/");
    cy.get(selector.username).type("admin");
    cy.get(selector.password).type("123");
    cy.get(selector.loginButton).click();
    cy.location("pathname", {timeout: 20000}).should("eq", "/");
  });

  it("Login failed", () => {
    cy.request({method: "POST", url: "/api/login", body: loginBody("1234"), failOnStatusCode: false}).then((response) => {
      expect(response).property("body").property("status").to.equal("error");
    });
  });

  it("ui Login failed", () => {
    cy.visit("/");
    cy.get(selector.username).type("admin");
    cy.get(selector.password).type("1234");
    cy.get(selector.loginButton).click();
    cy.location("pathname").should("eq", "/login");
  });
});
