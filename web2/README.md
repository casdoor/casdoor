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
| `src/locales/**` | copied from `web`, same 11 bundled languages, plus 9 strings for rows `web` does not render |
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

**Captcha** — the full rule handling of the antd frontend (`Never` / `Always` /
`Dynamic` / `Internet-Only`), as a dialog or inline in the sign-in form, for both
sign-in/sign-up and the "Get Code" button. Default (image), reCAPTCHA v2/v3,
hCaptcha, Aliyun, GEETEST and Cloudflare Turnstile all mount through the same
`CaptchaWidget` port, so the tokens the backend validates are unchanged.

**MFA** — `/mfa/setup` is a three-step wizard (verify password → verify the
factor → enable + recovery code) covering SMS, email, TOTP (QR + secret), RADIUS
and push. An organization that marks a factor `Required` redirects the user there
right after sign-in, and the user page can set the preferred factor or remove
MFA.

**Console** — dashboard, apps, shortcuts, my account, system info, and list +
edit pages for: organizations, users, groups (incl. tree), invitations,
applications, providers, resources, certs, keys, roles, permissions, models,
adapters, enforcers, agents, MCP servers, entries, sites, rules, sessions,
records, tokens, verifications, products, coupons, orders, payments, plans,
pricings, subscriptions, transactions, forms, syncers, webhooks, webhook events,
tickets, LDAP (edit + sync).

**Organization branding** — the sign-in pages apply the organization/application
`themeData` to the shadcn CSS variables, set the tab title and favicon, inject
`headerHtml` / `pageHtml` into `<head>`, and block the whole surface when an
`ipRestriction` is set. Before `get-application` returns, the `organizationTheme`
/ `organizationLogo` / `organizationFootHtml` cookies that
`routers/theme_filter.go` sets stand in, so the first paint is already branded.

**Organization menus** — `navItems` / `userNavItems` hide navbar entries (and
collapse the sidebar into a flat list once few enough are left), and
`widgetItems` hides the header's theme, language, AI assistant and tour buttons.

**Bulk import** — "Download template" and "Upload (.xlsx)" on the user, group,
role and permission list pages, posting to `upload-users` / `upload-groups` /
`upload-roles` / `upload-permissions`.

**Theme** — the organization/application `themeData` editor covers the same four
values the antd one did (theme preset, primary colour, border radius, compact),
rendered as shadcn controls rather than antd-token-previewer. `themeType: "dark"`
forces the dark palette on the branded surfaces and `isCompact` tightens the root
rem size, which is this frontend's equivalent of antd's dark/compact algorithms.
The navbar and widget item pickers are checkable trees (`CheckboxTree`) with
antd's check semantics, so the stored `navItems` / `userNavItems` / `widgetItems`
are byte-compatible with the antd frontend.

## Deliberately not ported

- **Entry viewers**: the rich entry message viewer, the SELinux entry viewer, and
  the OpenClaw session graph and transcript pages. The entry edit page shows the
  raw message instead, and the entry list no longer offers a "Transcript" action.
- **Web3**: `auth/Web3Auth.ts` talks to `window.ethereum` directly rather than
  going through `@web3-onboard`, so the extra wallets that library bundles
  (Coinbase, Phantom, Trust, Gnosis, ...) are not offered. MetaMask works.

## Still to do

- **Deployment**: the Go backend still serves `web/build`
  (`routers/static_filter.go` returns it before it ever looks at
  `frontendBaseDir`), `public/` has none of the standalone scripts
  `routers/lightweight_auth_filter.go` serves, `index.html` is missing the
  boot-failure fallback, and the Dockerfile / Makefile / CI still build `../web`.
  There are no Cypress e2e tests here either.

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
