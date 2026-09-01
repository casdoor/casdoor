// The second factor at sign-in. `/api/login` is stubbed so the specs can drive
// each factor without an account that actually has it enabled.

const nextMfa = (factors) => ({status: "ok", msg: "", data: "NextMfa", data2: factors});

const factor = (mfaType, extra = {}) => ({
  mfaType,
  countryCode: "",
  secret: "",
  mfaRememberInHours: 12,
  ...extra,
});

/** First `/api/login` answers "NextMfa"; the re-post with the passcode succeeds. */
function stubLogin(factors) {
  let calls = 0;
  cy.intercept({method: "POST", pathname: "/api/login"}, (req) => {
    calls += 1;
    if (calls === 1) {
      req.reply(nextMfa(factors));
    } else {
      req.reply({status: "ok", msg: "", data: "", data2: null});
    }
  }).as("login");
}

function signIn() {
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
}

describe("Sign-in second factor", () => {
  it("starts on the preferred factor, not merely the first one", () => {
    // SMS comes first on the wire, but the user prefers the authenticator app
    stubLogin([factor("sms", {secret: "13800000000"}), factor("app", {isPreferred: true})]);
    signIn();

    cy.contains("please enter the TOTP code").should("be.visible");
    cy.get("input[inputmode=numeric]").should("be.visible");
  });

  it("lets the user switch to another enabled factor", () => {
    stubLogin([factor("sms", {secret: "13800000000"}), factor("app", {isPreferred: true})]);
    signIn();

    cy.contains("button", "Use SMS").click();
    // the SMS factor shows the "Get Code" field instead of a bare passcode box
    cy.contains("button", "Get Code").should("be.visible");
    cy.contains("please enter the TOTP code").should("not.exist");

    cy.contains("button", "Use Authenticator App").click();
    cy.contains("please enter the TOTP code").should("be.visible");
  });

  it("asks for the RADIUS password, not a TOTP code", () => {
    stubLogin([factor("radius")]);
    signIn();

    cy.contains("please enter the RADIUS password").should("be.visible");
    // a password is not digits, so this factor does not get the numeric keypad
    cy.get("input[inputmode=text]").should("have.attr", "placeholder", "Password");
  });

  it("asks for the code from the push notification", () => {
    stubLogin([factor("push")]);
    signIn();

    cy.contains("please enter the verification code from push notification").should("be.visible");
  });

  it("re-posts the login with the passcode and the chosen factor", () => {
    stubLogin([factor("app", {isPreferred: true})]);
    signIn();

    cy.get("input[inputmode=numeric]").type("123456");
    cy.contains("button", "Verify Code").click();

    cy.wait("@login").its("request.body").should((body) => {
      expect(body.passcode).to.equal("123456");
      expect(body.mfaType).to.equal("app");
      // antd blanks these on the re-post; the session is carried by the first call
      expect(body.password).to.equal("");
      expect(body.username).to.equal("");
    });
  });
});
