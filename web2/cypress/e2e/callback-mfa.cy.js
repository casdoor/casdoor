// Signing in through an external provider when the account has MFA on.
//
// The provider's authorization code is single-use, so the second factor has to
// be collected on the callback page itself. Sending the user to /login instead
// drops the pending login and leaves them unable to finish — this spec is here
// to keep that from coming back.

const factor = {mfaType: "app", countryCode: "", secret: "", mfaRememberInHours: 12, isPreferred: true};

/** First `/api/login` answers "NextMfa"; the re-post with the passcode succeeds. */
function stubProviderLogin() {
  let calls = 0;
  cy.intercept({method: "POST", pathname: "/api/login"}, (req) => {
    calls += 1;
    if (calls === 1) {
      req.reply({status: "ok", msg: "", data: "NextMfa", data2: [factor]});
    } else {
      req.reply({status: "ok", msg: "", data: "", data2: null});
    }
  }).as("login");
}

function visitCallback(win) {
  win.localStorage.setItem("isTourVisible", "false");
  win.localStorage.setItem("language", "en");
}

describe("Provider callback with MFA", () => {
  beforeEach(() => {
    // signed in, so the redirect after the second factor lands on the console
    cy.openConsole();
    stubProviderLogin();
  });

  it("collects the second factor on /callback instead of bouncing to /login", () => {
    // AuthCallback reads the original request out of the base64 `state`
    const state = btoa("?application=app-built-in&provider=github&method=signin");
    cy.visit(`/callback?code=e2e-code&state=${state}`, {onBeforeLoad: visitCallback});
    cy.wait("@login");

    cy.location("pathname").should("eq", "/callback");
    cy.contains("Multi-factor authentication").should("be.visible");
    cy.contains("please enter the TOTP code").should("be.visible");

    cy.get("input[inputmode=numeric]").type("123456");
    cy.contains("button", "Verify Code").click();

    // the re-post keeps the provider's login body, with the provider moved aside
    cy.wait("@login").its("request.body").should((body) => {
      expect(body.passcode).to.equal("123456");
      expect(body.mfaType).to.equal("app");
      expect(body.code).to.equal("e2e-code");
      expect(body.application).to.equal("app-built-in");
      expect(body.providerBack).to.equal("github");
      expect(body.provider).to.equal("");
    });

    // responseType "login" ends on the console
    cy.location("pathname", {timeout: 20000}).should("eq", "/");
  });

  it("collects the second factor on /callback/saml", () => {
    cy.intercept({method: "GET", pathname: "/api/get-application"}, {
      status: "ok",
      msg: "",
      data: {owner: "admin", name: "app-built-in", displayName: "Casdoor", organization: "built-in"},
    }).as("getApplication");

    // relayState is base64 "clientId&state&provider&redirectUri"
    const relayState = btoa("&e2e-state&saml-provider&null");
    cy.visit(`/callback/saml?relayState=${encodeURIComponent(relayState)}&samlResponse=e2e-assertion`, {
      onBeforeLoad: visitCallback,
    });
    cy.wait("@login");

    cy.location("pathname").should("eq", "/callback/saml");
    cy.contains("Multi-factor authentication").should("be.visible");

    cy.get("input[inputmode=numeric]").type("654321");
    cy.contains("button", "Verify Code").click();

    cy.wait("@login").its("request.body").should((body) => {
      expect(body.passcode).to.equal("654321");
      expect(body.samlResponse).to.equal("e2e-assertion");
      expect(body.providerBack).to.equal("saml-provider");
      expect(body.provider).to.equal("");
    });

    cy.location("pathname", {timeout: 20000}).should("eq", "/");
  });
});
