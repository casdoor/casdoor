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

describe("Lightweight provider device callback", () => {
  function runCallback(source, response) {
    const query = "?application=app&provider=line&method=signin&userCode=ABCD-EFGH";
    const location = new URL(`https://identity.example.test/callback?state=${encodeURIComponent(btoa(query))}&code=fixture-code`);
    let target;
    location.replace = (url) => { target = url; };
    const values = new Map([["from", "/after-device"]]);
    const storage = {
      getItem: (key) => values.get(key) ?? null,
      setItem: (key, value) => values.set(key, String(value)),
      removeItem: (key) => values.delete(key),
    };
    const fixtureWindow = {location, frameElement: null};
    const requests = [];
    const fetch = async (_url, options) => {
      requests.push(JSON.parse(options.body));
      return {json: async () => response};
    };
    new Function("window", "document", "localStorage", "sessionStorage", "fetch", source)(
      fixtureWindow,
      {querySelector: () => null, getElementById: () => ({style: {}})},
      storage, storage, fetch,
    );
    return fixtureWindow.CasdoorAuthCallback.run().then(() => ({requests, target, storage}));
  }

  it("preserves the device request and finishes without another login", () => {
    cy.request("/AuthCallbackHandler.js").then(({body: source}) => {
      return runCallback(source, {status: "ok"}).then(({requests, target}) => {
        expect(requests).to.have.length(1);
        expect(requests[0]).to.include({
          type: "device",
          userCode: "ABCD-EFGH",
          application: "app",
          provider: "line",
          code: "fixture-code",
        });
        expect(target).to.equal("/after-device");
      });
    });
  });

  it("preserves the device request when handing off to MFA", () => {
    cy.request("/AuthCallbackHandler.js").then(({body: source}) => {
      return runCallback(source, {status: "ok", data: "NextMfa"}).then(({requests, target, storage}) => {
        expect(requests).to.have.length(1);
        expect(new URL(target).searchParams.get("__casdoor_callback_react")).to.equal("1");
        const payload = JSON.parse(storage.getItem("casdoor_callback_react_fallback"));
        expect(payload.responseType).to.equal("device");
        expect(payload.body).to.include({type: "device", userCode: "ABCD-EFGH"});
        expect(payload.res.data).to.equal("NextMfa");
      });
    });
  });
});
