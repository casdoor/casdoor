// The two places the frontend has to cooperate with
// `routers/lightweight_auth_filter.go`: the static callback page's handover back
// to React, and the `provider_hint` jump its redirect page normally makes.

const ok = (data, data2 = null) => ({status: "ok", msg: "", data, data2});

const REACT_FALLBACK_PAYLOAD_KEY = "casdoor_callback_react_fallback";

// what `?state=` decodes to; Util.getQueryParamsFromState falls back to atob()
const INNER_QUERY =
  "application=e2e-app&provider=p-github&method=signin&redirect_uri=http%3A%2F%2Flocalhost%3A7002%2Fcallback";

function callbackUrl({fallbackMarker}) {
  const state = btoa(INNER_QUERY);
  const search = `?code=THE-PROVIDER-CODE&state=${encodeURIComponent(state)}`;
  return {
    search,
    url: `/callback${search}${fallbackMarker ? "&__casdoor_callback_react=1" : ""}`,
  };
}

/** Seeds the handover the static page would have left behind. */
function seedFallback(win, payload) {
  win.sessionStorage.setItem(REACT_FALLBACK_PAYLOAD_KEY, JSON.stringify(payload));
  win.localStorage.setItem("isTourVisible", "false");
  win.localStorage.setItem("language", "en");
}

describe("Static callback page handover", () => {
  it("continues from the stored response instead of spending the code again", () => {
    cy.intercept({method: "POST", pathname: "/api/login"}, ok("should-not-happen")).as("login");
    const {search, url} = callbackUrl({fallbackMarker: true});

    cy.visit(url, {
      onBeforeLoad: (win) =>
        seedFallback(win, {
          search,
          flow: "oauth",
          responseType: "login",
          innerParams: INNER_QUERY,
          res: ok("signed-in"),
        }),
    });

    // the handover already carried the answer, so the login page is reached
    // without re-POSTing the provider's single-use code
    cy.location("pathname", {timeout: 20000}).should("not.eq", "/callback");
    cy.get("@login.all").should("have.length", 0);
  });

  it("clears the handover so it cannot be replayed", () => {
    const {search, url} = callbackUrl({fallbackMarker: true});

    cy.visit(url, {
      onBeforeLoad: (win) =>
        seedFallback(win, {
          search,
          flow: "oauth",
          responseType: "login",
          innerParams: INNER_QUERY,
          res: ok("signed-in"),
        }),
    });

    cy.location("pathname", {timeout: 20000}).should("not.eq", "/callback");
    cy.window().then((win) => {
      expect(win.sessionStorage.getItem(REACT_FALLBACK_PAYLOAD_KEY)).to.eq(null);
    });
  });

  it("ignores a handover left over from a different callback", () => {
    cy.intercept({method: "POST", pathname: "/api/login"}, ok("signed-in")).as("login");
    const {url} = callbackUrl({fallbackMarker: true});

    cy.visit(url, {
      onBeforeLoad: (win) =>
        seedFallback(win, {
          // a stale payload from an earlier sign-in
          search: "?code=AN-OLDER-CODE&state=b3RoZXI=",
          flow: "oauth",
          responseType: "login",
          innerParams: INNER_QUERY,
          res: ok("stale"),
        }),
    });

    // it does not match this URL, so the normal sign-in runs
    cy.wait("@login").its("request.body.code").should("eq", "THE-PROVIDER-CODE");
  });
});

describe("provider_hint", () => {
  it("goes straight to the hinted provider", () => {
    // a Custom OAuth provider so the authorize URL stays on this origin and
    // Cypress can follow it
    const landing = `${Cypress.config("baseUrl")}/provider-hint-landed`;
    const application = {
      owner: "admin",
      name: "e2e-app",
      displayName: "e2e app",
      organization: "built-in",
      enablePassword: true,
      signinMethods: [{name: "Password", rule: "All"}],
      signupItems: [],
      organizationObj: {owner: "admin", name: "built-in", passwordOptions: [], countryCodes: ["US"]},
      providers: [
        {
          name: "p-custom",
          canSignIn: true,
          provider: {
            owner: "admin",
            name: "p-custom",
            category: "OAuth",
            type: "Custom",
            displayName: "Custom",
            clientId: "cid",
            scopes: "openid",
            customAuthUrl: landing,
          },
        },
      ],
    };
    const body = ok(application);
    cy.intercept({method: "GET", pathname: "/api/get-default-application"}, body);
    cy.intercept({method: "GET", pathname: "/api/get-application"}, body);
    cy.intercept({method: "GET", pathname: "/api/get-app-login"}, body);

    cy.visit("/login/built-in?provider_hint=p-custom", {
      onBeforeLoad(win) {
        win.localStorage.setItem("isTourVisible", "false");
        win.localStorage.setItem("language", "en");
      },
    });

    cy.location("pathname", {timeout: 20000}).should("eq", "/provider-hint-landed");
    cy.location("search").should("include", "client_id=cid");
  });

  it("leaves the picker alone when the hint names no visible provider", () => {
    const body = ok({
      owner: "admin",
      name: "e2e-app",
      displayName: "e2e app",
      organization: "built-in",
      enablePassword: true,
      signinMethods: [{name: "Password", rule: "All"}],
      signupItems: [],
      providers: [],
      organizationObj: {owner: "admin", name: "built-in", passwordOptions: [], countryCodes: ["US"]},
    });
    cy.intercept({method: "GET", pathname: "/api/get-default-application"}, body);
    cy.intercept({method: "GET", pathname: "/api/get-application"}, body);
    cy.intercept({method: "GET", pathname: "/api/get-app-login"}, body);

    cy.visit("/login/built-in?provider_hint=nope", {
      onBeforeLoad(win) {
        win.localStorage.setItem("isTourVisible", "false");
        win.localStorage.setItem("language", "en");
      },
    });

    cy.get("#password").should("exist");
    cy.location("pathname").should("eq", "/login/built-in");
  });
});
