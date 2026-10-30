// Source-level regressions with the existing TypeScript compiler and Node test runner.
// DOM shims do not establish React rendering, browser compatibility or CSP enforcement.
import test from "node:test";
import assert from "node:assert/strict";
import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import path from "node:path";
import vm from "node:vm";
import ts from "typescript";

const web = path.dirname(fileURLToPath(import.meta.url));
const source = (relative) => readFileSync(path.resolve(web, relative), "utf8");
const catalogs = Object.fromEntries(["en", "zh-TW"].map((language) => [language, JSON.parse(source(`src/locales/${language}/data.json`))]));

function translator(language) {
  return {t(fullKey, values = {}) {
    const colon = fullKey.indexOf(":");
    const namespace = fullKey.slice(0, colon);
    const key = fullKey.slice(colon + 1);
    const value = catalogs[language]?.[namespace]?.[key] ?? catalogs.en[namespace]?.[key];
    assert.equal(typeof value, "string", `missing key ${fullKey}`);
    return value.replace(/\{\{([^{}]+)\}\}/g, (placeholder, name) => values[name] === undefined ? placeholder : String(values[name]));
  }};
}

function parsed(relative) {
  return ts.createSourceFile(relative, source(relative), ts.ScriptTarget.Latest, true, relative.endsWith("tsx") ? ts.ScriptKind.TSX : ts.ScriptKind.TS);
}

function declarations(relative, names, injected = {}) {
  const file = parsed(relative);
  const wanted = new Set(names);
  const selected = file.statements.filter((statement) =>
    ts.isFunctionDeclaration(statement) && wanted.has(statement.name?.text) ||
    ts.isVariableStatement(statement) && statement.declarationList.declarations.some((declaration) => wanted.has(declaration.name.getText(file))),
  );
  assert.equal(selected.length, names.length, `missing declaration in ${relative}`);
  const code = selected.map((statement) => statement.getText(file)).join("\n") + `\nglobalThis.result = {${names.join(",")}};`;
  const compiled = ts.transpileModule(code, {compilerOptions: {module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020}}).outputText;
  const context = {exports: {}, ...injected};
  // Only fixed, checked-in first-party declarations are evaluated; vm is not a security sandbox.
  vm.runInNewContext(compiled, context, {timeout: 1000});
  return context.result;
}

test("language aliases preserve Traditional/Simplified script distinctions", () => {
  const {normalizeLanguage} = declarations("src/i18n.ts", ["SupportedLanguages", "normalizeLanguage"]);
  for (const language of ["zh-TW", "zh_Hant_TW", "zh-Hant", "zh-HK", "zh-MO"]) assert.equal(normalizeLanguage(language), "zh-TW");
  for (const language of ["zh", "zh-CN", "zh-Hans", "zh-Hans-HK"]) assert.equal(normalizeLanguage(language), "zh");
  assert.equal(normalizeLanguage("en-US"), "en");
  assert.equal(normalizeLanguage("invalid-language-value"), "");
});

test("enum captions keep stored IDs and unknown/prototype-like values", () => {
  const {enumLabel, providerTypeLabel} = declarations("src/lib/enum-labels.tsx", ["PROVIDER_TYPE_LABELS", "enumLabel", "providerTypeLabel"], {i18next: translator("zh-TW")});
  assert.equal(providerTypeLabel("Local File System"), "本機檔案系統");
  assert.equal(providerTypeLabel("Custom2"), "自訂 2");
  assert.equal(providerTypeLabel("GitHub"), "GitHub");
  for (const value of ["constructor", "toString", "__proto__"]) assert.equal(enumLabel({}, value), value);
});

test("translated multi-select captions and rich-label keywords remain searchable", () => {
  const file = parsed("src/components/common/MultiSelect.tsx");
  const found = [];
  function visit(node) {
    if (ts.isJsxAttribute(node) && node.name.getText(file) === "filter" && node.initializer?.expression) found.push(node.initializer.expression);
    ts.forEachChild(node, visit);
  }
  visit(file);
  assert.equal(found.length, 1);
  const options = [{value: "authorization_code", label: "授權碼"}, {value: "custom", label: {}, keywords: "彈性自訂"}];
  const compiled = ts.transpileModule(`globalThis.filter = ${found[0].getText(file)};`, {compilerOptions: {target: ts.ScriptTarget.ES2020}}).outputText;
  const context = {options};
  vm.runInNewContext(compiled, context, {timeout: 1000});
  assert.equal(context.filter("authorization_code", "授權"), 1);
  assert.equal(context.filter("authorization_code", "AUTHORIZATION"), 1);
  assert.equal(context.filter("custom", "彈性"), 1);
  assert.equal(context.filter("authorization_code", "不存在"), 0);
});

test("XLSX preview follows stable field suffixes without choosing ambiguous data", () => {
  const {getXlsxPreviewValue} = declarations("src/components/crud/XlsxImport.tsx", ["getXlsxPreviewValue"]);
  const row = {"Name#name": "original-id", "Enabled#enabled": false, "Count#count": 0};
  assert.equal(getXlsxPreviewValue(row, "名稱#name"), "original-id");
  assert.equal(getXlsxPreviewValue(row, "啟用#enabled"), false);
  assert.equal(getXlsxPreviewValue(row, "數量#count"), 0);
  assert.equal(getXlsxPreviewValue({"A#name": "a", "B#name": "b"}, "名稱#name"), undefined);
  assert.equal(getXlsxPreviewValue(row, "unrelated"), undefined);
  assert.deepEqual(row, {"Name#name": "original-id", "Enabled#enabled": false, "Count#count": 0});
});

test("breadcrumbs translate actual route actions but never coinciding entity IDs", () => {
  const {buildBreadcrumbItems} = declarations("src/components/layout/BreadcrumbBar.tsx", ["RESOURCE_LABELS", "buildBreadcrumbItems"], {i18next: translator("zh-TW")});
  assert.equal(buildBreadcrumbItems("/users/acme/groups").at(-1).label, "groups");
  assert.equal(buildBreadcrumbItems("/organizations/users").at(-1).label, "users");
  assert.equal(buildBreadcrumbItems("/organizations/acme/users").at(-1).label, catalogs["zh-TW"].general.Users);
  assert.equal(buildBreadcrumbItems("/products/acme/buy").at(-1).label, "buy");
  assert.equal(buildBreadcrumbItems("/products/acme/example/buy").at(-1).label, catalogs["zh-TW"].product.Buy);
  assert.equal(buildBreadcrumbItems("/mfa/setup").at(-1).label, catalogs["zh-TW"].mfa["Set MFA"]);
});

test("new record names localize only on creation and do not change stable values", () => {
  const Setting = {getRandomName: () => "synthetic", getRequestOrganization: () => "acme", isDefaultOrganizationSelected: () => false};
  const {newProvider, newGroup} = declarations("src/pages/defaults.ts", ["now", "newProvider", "newGroup"], {
    i18next: translator("zh-TW"), Setting, dayjs: () => ({format: () => "2026-01-01T00:00:00Z"}),
  });
  const provider = newProvider({owner: "acme"});
  assert.equal(provider.displayName, `${catalogs["zh-TW"].provider["New Provider"]} - synthetic`);
  assert.equal(provider.name, "provider_synthetic");
  assert.equal(provider.category, "OAuth");
  assert.equal(provider.type, "GitHub");
  assert.equal(newGroup({owner: "acme"}).type, "Virtual");
});

test("retained default form schemas translate captions without rewriting saved custom labels", () => {
  let language = "en";
  const i18next = {t: (key) => key.includes(":") ? translator(language).t(key) : key};
  const {getFormTypeItems, filterTableColumns} = declarations("src/lib/setting.tsx", ["getFormTypeItems", "filterTableColumns"], {i18next});
  const schema = getFormTypeItems("applications");
  const snapshot = JSON.stringify(schema);
  assert.equal(schema.find((item) => item.name === "logo").label, "general:Logo");
  const columns = [{key: "logo", title: "original caption"}];
  assert.equal(filterTableColumns(columns, schema)[0].title, catalogs.en.general.Logo);
  language = "zh-TW";
  assert.equal(filterTableColumns(columns, schema)[0].title, catalogs["zh-TW"].general.Logo);
  assert.equal(JSON.stringify(schema), snapshot);
  const custom = [{name: "logo", label: "Customer-owned Logo", visible: true, width: "200"}];
  assert.equal(filterTableColumns(columns, custom)[0].title, custom[0].label);
  assert.equal(custom[0].label, "Customer-owned Logo");
});

test("new email templates preserve substitution markers and use the chosen document language", () => {
  const {getDefaultHtmlEmailContent, getDefaultInvitationHtmlEmailContent} = declarations("src/lib/setting.tsx", ["getDefaultHtmlEmailContent", "getDefaultInvitationHtmlEmailContent"], {
    i18next: translator("zh-TW"), getLanguage: () => "zh-TW", StaticBaseUrl: "https://example.invalid/assets",
  });
  const verification = getDefaultHtmlEmailContent();
  const invitation = getDefaultInvitationHtmlEmailContent();
  assert.match(verification, /<html lang="zh-TW">/);
  assert.match(verification, /%s/);
  assert.match(verification, /%\{user\.friendlyName\}/);
  assert.match(verification, /<reset-link>[\s\S]*%link[\s\S]*<\/reset-link>/);
  assert.match(invitation, /%code/);
  assert.match(invitation, /%link/);
  assert.doesNotMatch(verification, /5 minutes|5 分鐘/);
});

const goSource = source("../routers/static_filter_util.go");
const resolver = /const nativeTextResolver = `([\s\S]*?)`/.exec(goSource)?.[1];
assert.equal(typeof resolver, "string");
const nativeCatalogs = Object.fromEntries(["en", "zh-TW"].map((language) => [language, JSON.parse(source(`../i18n/locales/${language}/data.json`)).native]));

function element(textContent = "", key = "") {
  const attributes = new Map(key ? [["data-casdoor-native-text", key]] : []);
  return {textContent, style: {}, getAttribute: (name) => attributes.get(name) ?? null,
    setAttribute: (name, value) => attributes.set(name, value), removeAttribute: (name) => attributes.delete(name),
    set innerHTML(_value) { throw new Error("Diagnostic text must not be parsed as HTML"); }};
}

function nativeFixture(savedLanguage, {language = "en", target = "https://example.invalid/callback", storageThrows = false} = {}) {
  const status = element(nativeCatalogs[language]["Signing in..."], "Signing in...");
  const loading = element();
  const nodes = {"callback-spinner": element(), "callback-status": status, "casdoor-boot-fallback": loading,
    "casdoor-boot-title": element(), "casdoor-boot-detail": element()};
  const domEvents = new Map();
  const windowEvents = new Map();
  const session = new Map();
  const local = new Map(savedLanguage === null ? [] : [["language", savedLanguage]]);
  const localStorage = {getItem: (key) => { if (storageThrows) throw new Error("storage unavailable"); return local.get(key) ?? null; },
    setItem: (key, value) => local.set(key, value), removeItem: (key) => local.delete(key)};
  const location = new URL(target);
  location.replace = () => { throw new Error("Unexpected navigation"); };
  const window = {localStorage, location, addEventListener: (name, callback) => windowEvents.set(name, callback), setTimeout() {}};
  const document = {documentElement: {lang: language}, readyState: "loading",
    querySelector: (selector) => selector.includes("casdoor-native-messages") ? {getAttribute: () => JSON.stringify({language, catalogs: nativeCatalogs})} : null,
    querySelectorAll: () => status.getAttribute("data-casdoor-native-text") === null ? [] : [status],
    getElementById: (id) => nodes[id] ?? null, addEventListener: (name, callback) => domEvents.set(name, callback)};
  const context = {window, document, localStorage, navigator: {language}, URL, URLSearchParams,
    sessionStorage: {getItem: (key) => session.get(key) ?? null, setItem: (key, value) => session.set(key, value), removeItem: (key) => session.delete(key)},
    atob: (value) => Buffer.from(value, "base64").toString("binary"), fetch: () => { throw new Error("Unexpected network request"); }};
  vm.createContext(context);
  vm.runInContext(resolver, context, {timeout: 1000});
  return {context, window, document, status, nodes, session, local, domEvents, windowEvents};
}

test("native fallback resolver stays ES5-only and tolerates unavailable storage", () => {
  const tree = ts.createSourceFile("resolver.js", resolver, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS);
  const forbidden = new Set([ts.SyntaxKind.ArrowFunction, ts.SyntaxKind.ClassDeclaration, ts.SyntaxKind.TemplateExpression, ts.SyntaxKind.NoSubstitutionTemplateLiteral]);
  function visit(node) {
    assert.ok(!forbidden.has(node.kind), `non-ES5 syntax: ${ts.SyntaxKind[node.kind]}`);
    if (ts.isVariableDeclarationList(node)) assert.equal(node.flags & (ts.NodeFlags.Let | ts.NodeFlags.Const), 0);
    ts.forEachChild(node, visit);
  }
  visit(tree);
  for (const value of ["zh-TW", "zh_Hant_TW", "zh-HK", "zh-MO"]) {
    const fixture = nativeFixture(value);
    assert.equal(fixture.window.CasdoorNativeText("Signing in..."), "正在登入……");
    fixture.domEvents.get("DOMContentLoaded")();
    assert.equal(fixture.status.textContent, "正在登入……");
    assert.equal(fixture.document.documentElement.lang, "zh-TW");
  }
  assert.equal(nativeFixture(null, {language: "zh-TW", storageThrows: true}).window.CasdoorNativeText("Loading"), "載入中");
  assert.equal(nativeFixture("en", {language: "zh-TW"}).window.CasdoorNativeText("Loading"), "Loading");
});

test("lightweight callback error survives deferred DOM localization without network", async () => {
  const fixture = nativeFixture("zh-TW");
  vm.runInContext(source("public/AuthCallbackHandler.js"), fixture.context, {timeout: 1000});
  await fixture.window.CasdoorAuthCallback.run();
  assert.equal(fixture.status.textContent, "缺少登入回呼狀態。");
  fixture.domEvents.get("DOMContentLoaded")();
  assert.equal(fixture.status.textContent, "缺少登入回呼狀態。");
  assert.equal(fixture.status.getAttribute("data-casdoor-native-text"), null);
});

test("standalone OAuth and CAS requests use the effective native locale without changing auth payloads", async () => {
  for (const cas of [false, true]) {
    for (const savedLanguage of [null, "en", "zh_Hant_HK"]) {
      const fixture = nativeFixture(savedLanguage, {language: "zh-TW", target: "https://example.invalid/callback?state=synthetic&code=synthetic-code"});
      fixture.context.navigator.language = "en-US";
      const params = new URLSearchParams({method: "signin", application: "synthetic-app", provider: "synthetic-provider", invitationCode: "synthetic-invite"});
      if (cas) params.set("service", "https://service.example.invalid/cas");
      fixture.session.set("synthetic", params.toString());
      fixture.local.set("pkce_verifier_synthetic", "synthetic-s256-verifier");
      const calls = [];
      fixture.context.fetch = async (url, options) => {
        calls.push({url, options});
        return {json: async () => ({status: "error", msg: "synthetic backend outcome"})};
      };
      vm.runInContext(source("public/AuthCallbackHandler.js"), fixture.context, {timeout: 1000});
      await fixture.window.CasdoorAuthCallback.run();
      assert.equal(calls.length, 1);
      const {url, options} = calls[0];
      assert.equal(options.headers["Accept-Language"], savedLanguage === "en" ? "en" : "zh-TW");
      assert.equal(options.credentials, "include");
      assert.equal(options.method, "POST");
      assert.equal(new URL(url).pathname, "/api/login");
      const payload = JSON.parse(options.body);
      assert.equal(payload.application, "synthetic-app");
      assert.equal(payload.provider, "synthetic-provider");
      assert.equal(payload.code, "synthetic-code");
      assert.equal(payload.method, "signin");
      assert.equal(payload.invitationCode, "synthetic-invite");
      assert.equal(payload.type, cas ? "cas" : "login");
      assert.equal(payload.codeVerifier, "synthetic-s256-verifier");
      assert.equal(fixture.local.has("pkce_verifier_synthetic"), false);
      assert.equal(fixture.status.textContent, "synthetic backend outcome");
      assert.equal(new URL(fixture.window.location).searchParams.get("state"), "synthetic");
    }
  }
});

test("invalid link nonce remains unauthorized after locale changes", async () => {
  const fixture = nativeFixture("zh-TW", {target: "https://example.invalid/callback?state=synthetic"});
  fixture.session.set("synthetic", "method=link&application=synthetic&provider=synthetic&linkNonce=wrong");
  fixture.session.set("casdoor_link_nonce", "expected");
  vm.runInContext(source("public/AuthCallbackHandler.js"), fixture.context, {timeout: 1000});
  await fixture.window.CasdoorAuthCallback.run();
  assert.equal(fixture.status.textContent, "未經授權");
  assert.equal(fixture.session.has("casdoor_link_nonce"), false);
  fixture.domEvents.get("DOMContentLoaded")();
  assert.equal(fixture.status.textContent, "未經授權");
});

test("missing callback script guidance remains final", () => {
  const html = /const authCallbackHtml = `([\s\S]*?)`/.exec(source("../routers/lightweight_auth_filter.go"))?.[1];
  assert.equal(typeof html, "string");
  const body = [...html.matchAll(/<script(?:\s[^>]*)?>([\s\S]*?)<\/script>/g)].map((match) => match[1]).find((script) => script.includes("CasdoorAuthCallback"));
  assert.equal(typeof body, "string");
  const fixture = nativeFixture("zh-TW");
  vm.runInContext(body, fixture.context, {timeout: 1000});
  assert.equal(fixture.status.textContent, "無法載入登入回呼處理程式。");
  fixture.domEvents.get("DOMContentLoaded")();
  assert.equal(fixture.status.textContent, "無法載入登入回呼處理程式。");
});

test("bundle failure shows localized guidance and opaque diagnostic text", () => {
  const html = source("index.html");
  const body = [...html.matchAll(/<script(?:\s[^>]*)?>([\s\S]*?)<\/script>/g)].map((match) => match[1]).find((script) => script.includes("var shown = false"));
  assert.equal(typeof body, "string");
  const fixture = nativeFixture("zh-TW");
  fixture.nodes["casdoor-boot-fallback"].className = "";
  vm.runInContext(body, fixture.context, {timeout: 1000});
  const diagnostic = '<img src=x onerror="untrusted">';
  fixture.windowEvents.get("error")({message: diagnostic});
  assert.equal(fixture.nodes["casdoor-boot-title"].textContent, "Casdoor 載入失敗");
  assert.equal(fixture.nodes["casdoor-boot-detail"].textContent, diagnostic);
});
