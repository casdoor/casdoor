describe("Test providers", () => {
  beforeEach(() => {
    cy.openConsole();
  });

  it("test providers", () => {
    cy.visitListPage("/providers");
    cy.openFirstRow();
  });
});

describe("OAuth provider state", () => {
  const provider = {
    category: "OAuth",
    type: "Line",
    name: "line",
    clientId: "fixture-channel",
    scopes: "profile openid",
  };
  const application = {name: "app"};

  it("preserves base64 state in provider button authorization URLs", () => {
    cy.visit("/login?state=~~~");
    cy.get("#username").should("be.visible");
    cy.window().then(async (win) => {
      // Load through the app's Vite server so TypeScript and aliases resolve in its window.
      const loadModule = new win.Function("url", "return import(url)");
      const {getAuthUrl, getCodeVerifier} = await loadModule(`${win.location.origin}/src/auth/Provider.ts`);
      const state = win.btoa(`${win.location.search}&application=app&provider=line&method=signin`);
      expect(state).to.include("+");

      const params = new URL(getAuthUrl(application, provider, "signin")).searchParams;
      expect(params.get("state")).to.equal(state);
      expect(params.get("scope")).to.equal("profile openid");
      expect(getCodeVerifier(params.get("state"))).to.match(/^[A-Za-z0-9_-]{43}$/);
    });
  });

  it("preserves base64 state in lightweight provider-hint authorization URLs", () => {
    cy.request("/ProviderHintRedirect.js").then(({body: source}) => {
      const location = new URL("https://identity.example.test/login/oauth/authorize?state=~~~&provider_hint=line");
      let target;
      location.replace = (url) => { target = url; };
      const values = new Map();
      const storage = {
        getItem: (key) => values.get(key) ?? null,
        setItem: (key, value) => values.set(key, String(value)),
        removeItem: (key) => values.delete(key),
      };
      const fixtureWindow = {location, crypto: window.crypto};
      const fetch = async () => ({
        ok: true,
        json: async () => ({
          status: "ok",
          data: {...application, providers: [{provider, canSignIn: true}]},
        }),
      });
      new Function("window", "document", "localStorage", "sessionStorage", "fetch", source)(
        fixtureWindow, {querySelector: () => null}, storage, storage, fetch,
      );

      return fixtureWindow.CasdoorProviderHintRedirect.run().then(() => {
        const state = btoa(`${location.search}&application=app&provider=line&method=signup`);
        expect(state).to.include("+");
        const params = new URL(target).searchParams;
        expect(params.get("state")).to.equal(state);
        expect(params.get("scope")).to.equal("profile openid");
        expect(storage.getItem(`pkce_verifier_${params.get("state")}`)).to.match(/^[A-Za-z0-9_-]{43}$/);
      });
    });
  });
});
