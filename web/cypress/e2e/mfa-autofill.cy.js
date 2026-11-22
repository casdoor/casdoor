describe("TOTP provider callback", {retries: 0}, () => {
  const state = btoa("application=app-built-in&provider=example&method=signin");
  const callback = `/callback?code=synthetic-provider-code&state=${encodeURIComponent(state)}`;

  beforeEach(() => {
    cy.intercept("GET", "/api/get-account*", {status: "error", data: "Please login first"});
    cy.intercept("GET", /\/api\/get-(user-)?application\?/, {
      status: "ok",
      data: {owner: "admin", name: "app-built-in", displayName: "MFA test", organization: "built-in"},
    });
    cy.intercept("POST", "/api/login*", (req) => {
      if (req.body.provider) {
        req.alias = "providerCallback";
        req.reply({status: "ok", data: "NextMfa", data2: [{mfaType: "app", isPreferred: true, mfaRememberInHours: 0}]});
      } else {
        req.alias = "mfaVerify";
        req.reply({status: "error", msg: "Synthetic verification response"});
      }
    });
    cy.visit(callback, {onBeforeLoad(win) {
      win.localStorage.setItem("language", "en");
      win.localStorage.setItem("isTourVisible", "false");
    }});
    cy.wait("@providerCallback");
  });

  it("offers one named, labelled OTP input and submits all six digits including a leading zero", () => {
    cy.get('input[name="otp"]')
      .should("have.attr", "autocomplete", "one-time-code")
      .and("have.attr", "inputmode", "numeric")
      .and("have.attr", "maxlength", "6");
    cy.get('label[for="totp-code"]').should("be.visible");
    cy.get("#totp-code").type("012345{enter}");
    cy.wait("@mfaVerify").its("request.body").should("include", {passcode: "012345", provider: ""});
  });

  it("does not redeem the provider authorization code again after a reload during MFA", () => {
    cy.contains("Multi-factor authentication").should("be.visible");
    cy.location("search").should("not.contain", "code=");
    cy.reload();
    cy.contains("Sign-in was interrupted").should("be.visible");
    cy.get("@providerCallback.all").should("have.length", 1);
  });

  it("normalizes formatted clipboard codes before applying the six-digit limit", () => {
    cy.window().then((win) => {
      const clipboardData = new win.DataTransfer();
      clipboardData.setData("text", "012 345");
      cy.get("#totp-code").trigger("paste", {clipboardData});
    });
    cy.get("#totp-code").should("have.value", "012345").type("{enter}");
    cy.wait("@mfaVerify").its("request.body.passcode").should("equal", "012345");
  });
});
