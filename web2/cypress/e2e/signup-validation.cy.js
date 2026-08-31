// The client-side checks the antd signup form runs as Form rules, and the field
// names the payload has to use.

const ok = (data) => ({status: "ok", msg: "", data});

const item = (name, extra = {}) => ({name, visible: true, required: true, rule: "None", ...extra});

function stubApplication(signupItems) {
  // with no OAuth params the page loads the application by name
  const body = ok({
    owner: "admin",
    name: "e2e-app",
    displayName: "e2e app",
    organization: "built-in",
    enableSignUp: true,
    signupItems,
    signinMethods: [{name: "Password", rule: "All"}],
    providers: [],
    organizationObj: {owner: "admin", name: "built-in", passwordOptions: [], countryCodes: ["US"]},
  });
  cy.intercept({method: "GET", pathname: "/api/get-application"}, body).as("getApp");
  cy.intercept({method: "GET", pathname: "/api/get-app-login"}, body).as("getAppLogin");
}

function visitSignup() {
  cy.visit("/signup/e2e-app", {
    onBeforeLoad(win) {
      win.localStorage.setItem("isTourVisible", "false");
      win.localStorage.setItem("language", "en");
    },
  });
  cy.wait("@getApp");
}

describe("Signup validation", () => {
  it("enforces the signup item's own regex", () => {
    stubApplication([
      item("Username", {regex: "^[a-z]{4,}$"}),
      item("Signup button"),
    ]);
    cy.intercept({method: "POST", pathname: "/api/signup"}).as("signup");
    visitSignup();

    cy.get("#username").type("AB");
    cy.contains("button", "Sign Up").click();

    cy.contains("doesn't match the signup item regex").should("be.visible");
    cy.get("@signup.all").should("have.length", 0);

    // a value the regex accepts gets through to the backend
    cy.get("#username").clear().type("alice");
    cy.contains("button", "Sign Up").click();
    cy.wait("@signup").its("request.body.username").should("eq", "alice");
  });

  it("rejects a malformed email before posting", () => {
    stubApplication([item("Email", {rule: "No verification"}), item("Signup button")]);
    cy.intercept({method: "POST", pathname: "/api/signup"}).as("signup");
    visitSignup();

    cy.get("#email").type("not-an-email");
    cy.contains("button", "Sign Up").click();

    cy.contains("not valid Email").should("be.visible");
    cy.get("@signup.all").should("have.length", 0);
  });

  it("rejects an ID card number that fails the checksum shape", () => {
    stubApplication([item("ID card"), item("Signup button")]);
    cy.intercept({method: "POST", pathname: "/api/signup"}).as("signup");
    visitSignup();

    cy.get("#idCard").type("123");
    cy.contains("button", "Sign Up").click();
    cy.contains("correct ID card number").should("be.visible");
    cy.get("@signup.all").should("have.length", 0);
  });

  it("posts the multi-word fields under the names the backend reads", () => {
    stubApplication([
      item("ID card", {required: false}),
      item("First name", {required: false}),
      item("Last name", {required: false}),
      item("Invitation code", {required: false}),
      item("Signup button"),
    ]);
    cy.intercept({method: "POST", pathname: "/api/signup"}, ok("built-in/alice")).as("signup");
    visitSignup();

    cy.get("#idCard").type("110101199003077758");
    cy.get("#firstName").type("Ada");
    cy.get("#lastName").type("Lovelace");
    cy.get("#invitationCode").type("INV-1");
    cy.contains("button", "Sign Up").click();

    cy.wait("@signup").its("request.body").should((body) => {
      // "ID card" used to be posted as `iDcard`, which the backend never reads
      expect(body.idCard).to.equal("110101199003077758");
      expect(body.firstName).to.equal("Ada");
      expect(body.lastName).to.equal("Lovelace");
      expect(body.invitationCode).to.equal("INV-1");
    });
  });

  it("renders a Single Choice item as a picker", () => {
    stubApplication([
      item("Department", {type: "Single Choice", options: ["Engineering", "Sales"]}),
      item("Signup button"),
    ]);
    visitSignup();

    cy.get("[role=combobox]").click();
    cy.contains("Engineering").should("be.visible");
  });
});
