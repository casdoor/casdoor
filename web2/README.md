# web2 — Casdoor frontend on shadcn/ui

`web2` is a port of the Casdoor frontend from **Ant Design** (`../web`) to
**shadcn/ui + Tailwind CSS**, built with Vite. It talks to the same Go backend
over the same REST endpoints, so the two frontends are interchangeable.

```
Vite 5 · React 18 · TypeScript · Tailwind CSS 3 · shadcn/ui (Radix) · react-router 6 · i18next
```

## Getting started

```bash
yarn install
yarn start          # http://localhost:7002, proxies /api to http://localhost:8000
```

Point the dev proxy at a different backend with `CASDOOR_BACKEND`:

```bash
CASDOOR_BACKEND=https://door.casdoor.com yarn start
```

Build:

```bash
yarn build          # emits build-temp/
yarn postbuild      # renames build-temp/ -> build/
```

Other scripts: `yarn lint` / `yarn fix` (ESLint) and `yarn typecheck` (`tsc --noEmit`).

### Serving it from the Go backend

The backend serves `web/build` (see `routers/static_filter.go`). To serve `web2`
instead, either point that path at `web2/build` or copy `web2/build` over
`web/build`. Nothing else changes — the API contract is identical.

## How it maps to `../web`

| `web` (antd) | `web2` (shadcn) |
| --- | --- |
| `src/backend/*.js` | `src/backend/*.ts` — same functions, same endpoints, `any`-typed params |
| `src/Setting.js` | `src/lib/setting.tsx` — the pure logic is ported verbatim; antd renderers replaced |
| `src/auth/Util.js`, `Provider.js`, `Obfuscator.js` | `src/auth/*.ts` — ported verbatim (PKCE, OAuth/CAS/SAML query handling) |
| `src/locales/**` | copied unchanged — same keys, same 11 bundled languages |
| `BaseListPage` | `components/crud/CrudListPage` + `hooks/use-table-data` |
| antd `<Table>` | `components/crud/DataTable` (server-side paging, per-column search, sort) |
| `<Row><Col>` label rows | `components/crud/FormRow` |
| antd `message` | `sonner` toasts (`Setting.showMessage` keeps the same signature) |
| antd `ConfigProvider` dark algorithm | Tailwind `dark` class; the theme is stored under the same `themeAlgorithm` localStorage key, so it survives switching between the two frontends |
| `common/Editor.js` (CodeMirror) | `components/common/CodeEditor` |
| `table/*.js` sub-tables | `components/crud/EditableTable` |

Behaviour deliberately preserved: the `newXxx()` default objects
(`src/pages/defaults.ts`), the "add" flow (list page hands the object to the edit
page through router state, the edit page POSTs `add-x`), the organization
selector broadcasting `storageOrganizationChanged`, and the sign-in payloads
(`signinMethod`, password obfuscation, MFA re-post, PKCE state round-trip).

## What is covered

**Authentication** — sign in (self / OAuth authorize / CAS / SAML / device code),
sign up from `signupItems`, forgot password, `/callback` (OAuth, OIDC, SAML POST
binding, CAS, Telegram, Steam, Web3 token key), MFA second factor + recovery
code, consent, prompt, result.

**Console** — dashboard, apps, shortcuts, my account, system info, and list +
edit pages for: organizations, users, groups (incl. tree), invitations,
applications, providers, resources, certs, keys, roles, permissions, models,
adapters, enforcers, agents, MCP servers, entries, sites, rules, sessions,
records, tokens, verifications, products, coupons, orders, payments, plans,
pricings, subscriptions, transactions, forms, syncers, webhooks, webhook events,
tickets, LDAP (edit + sync).

## Not ported yet

These exist in `../web` and still need work here:

- **Login extras**: WebAuthn, Face ID, Web3/MetaMask, WeChat QR panel, Telegram
  widget, captcha modal/inline widget, device-login panel, `/mfa/setup`.
- **Storefront & checkout**: product store, server (MCP) store, cart, product
  buy, order pay, payment result, `/select-plan` and `/buy-plan` pricing pages,
  `/qrcode`.
- **Editors/viewers**: form editor (`/forms/:name`), Casbin policy editor on the
  adapter page, enforcer tester, OpenClaw session graph & transcript viewers,
  SELinux entry viewer.
- **Edit-page extras**: application import/export, resource upload, organization
  theme editor and navbar/widget item trees, and some sub-tables (token
  attributes, custom scopes, IP/WAF/UA rules, managed & MFA accounts, addresses,
  WebAuthn credentials).
- **Chrome**: the product tour, the AI assistant drawer, and the GitHub corner.

Everything above degrades to a normal 404 inside the console rather than
breaking another page.

## Adding a page

Most list pages are a `CrudListPage` with columns, and most edit pages are a
`SimpleEditPage` with a field list — see `src/pages/RoleListPage.tsx` and
`src/pages/RoleEditPage.tsx` for the shortest example of each. Reach for a
hand-written page (like `UserEditPage`) only when the layout genuinely differs.

## Notes

- `i18n.ts` sets `nsSeparator: ":"` explicitly. Without it, i18next v23 treats a
  key containing spaces as natural language and never splits off the `general:`
  namespace, so every multi-word key would resolve to itself.
- shadcn components live in `src/components/ui`. `components.json` is configured
  for the `new-york` style with the `neutral` base colour, so `npx shadcn@latest
  add <component>` drops new ones straight in.
