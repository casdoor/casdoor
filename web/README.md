# web — the Casdoor console

The Casdoor frontend, on **shadcn/ui + Tailwind CSS** and built with Vite. It
started as a port of the Ant Design console, which is kept for reference at
`../web-old` and is no longer built or served; both talk to the same Go backend
over the same REST endpoints.

```
Vite 5 · React 18 · TypeScript · Tailwind CSS 3 · shadcn/ui (Radix) · react-router 6 · i18next · reactflow
```

## Getting started

```bash
yarn install
yarn start          # http://localhost:7001, proxies /api to http://localhost:8000
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

## End-to-end tests

Cypress specs live in `cypress/e2e`. They need the Go backend on `:8000` and the
dev server on `:7001`, then:

```bash
yarn e2e            # headless
yarn e2e:open       # interactive runner
```

The suite signs in as `admin` / `123` and only reads: every "Add" it presses
hands the new object to the edit page through the router state, which is not
POSTed until the user saves, so no rows are created.

Three kinds of spec, because they catch different things:

- `console-pages.cy.js` sweeps every list page and every "Add" edit page. It is
  broad but works from `pages/defaults.ts`, where optional fields are still
  `undefined`.
- The per-feature specs (`application`, `certs`, `user`, ...) open the **first
  real row** of a list, so the edit page meets values the backend actually
  stores — which is how the `ipWhitelist` and `scopes` mis-bindings surfaced.
- `entry-viewers`, `map-fields`, `edit-page-details`, `mfa-signin`,
  `mfa-notification`, `callback-mfa`, `lightweight-callback`, `list-columns`,
  `list-filters`, `view-mode`, `page-actions`, `signup-validation`,
  `login-page`, `application-tables` and `misc-parity` stub the relevant endpoint with `cy.intercept`, for the states a
  real database rarely holds — an OpenClaw session, a plan-created invitation, an account with two MFA
  factors, a provider login that comes back asking for one, a non-admin account.

`DataTable` puts `data-column="<dataIndex>"` on each header cell, which is how a
spec reaches one column's sort, search or filter control; an open filter menu is
`[data-column-filter="<dataIndex>"]` and its entries are `[role=menuitemradio]`.

Two things worth knowing when writing a spec:

- Match intercepts on `pathname`, not a URL glob. Casdoor's query string is
  `?id=<owner>/<name>`, and the `/` in it makes minimatch read the object name as
  another path segment, so `**/api/get-entry*` never matches.
- On Windows, Vite binds `localhost` to `::1` only, while Cypress's Node-side
  `cy.request` resolves it to `127.0.0.1` and gets `ECONNREFUSED`. Start the dev
  server with `yarn start --host 127.0.0.1` (and point `CYPRESS_BASE_URL` at the
  same address) when that happens.

### Serving it from the Go backend

`routers/static_filter.go` serves `web/build`, which `yarn build` produces here.
`public/` carries the two standalone scripts `routers/lightweight_auth_filter.go`
serves (`AuthCallbackHandler.js`, `ProviderHintRedirect.js`); Vite copies them to
the build root, so a Docker image gets them without needing `public/` at runtime.

## How it maps to `../web-old`

| `web-old` (antd) | `web` (shadcn) |
| --- | --- |
| `src/backend/*.js` | `src/backend/*.ts` — same functions, same endpoints, `any`-typed params |
| `src/Setting.js` | `src/lib/setting.tsx` — the pure logic is ported verbatim; antd renderers replaced |
| `src/auth/Util.js`, `Provider.js`, `Obfuscator.js` | `src/auth/*.ts` — ported verbatim (PKCE, OAuth/CAS/SAML query handling) |
| `src/locales/**` | copied from the antd console, same 11 bundled languages, plus 9 strings for rows it does not render |
| `BaseListPage` | `components/crud/CrudListPage` + `hooks/use-table-data` |
| antd `<Table>` | `components/crud/DataTable` (server-side paging, per-column search, sort, filter menus) |
| a column's `render` that only wraps the cell in a link | `ColumnDef.link` / `linkExternal`, which keeps the search highlight inside the link |
| antd `<Result status="403">` | `components/common/UnauthorizedPage` |
| `<Row><Col>` label rows | `components/crud/FormRow` |
| antd `message` | `sonner` toasts (`Setting.showMessage` keeps the same signature) |
| antd `ConfigProvider` dark algorithm | Tailwind `dark` class; the theme is stored under the same `themeAlgorithm` localStorage key, so it survives switching between the two frontends |
| `common/Editor.js` (CodeMirror) | `components/common/CodeEditor` |
| `table/*.js` sub-tables | `components/crud/EditableTable` |
| antd `<Descriptions bordered>` | `components/common/DescriptionList` |
| antd `<Drawer>` | `components/ui/sheet` |
| `OpenClawSessionGraphUtils.js` | `lib/openclaw-graph.ts` — same layout, still `reactflow` |
| trace helpers inside `EntryMessageViewer.js` | `lib/otlp-trace.ts` |

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

The sign-in page honours `orgChoiceMode`, so an application can ask which
organization the visitor belongs to first; a visitor already signed in to that
organization gets a one-click "continue as" panel above the form, which is how
an OAuth or device request is approved. Sign-up runs the same client-side rules
the antd Form does — each signup item's own `regex` included — and reports them
in Casdoor's own wording rather than the browser's.

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

At sign-in the second factor opens on the factor the user marked preferred and
offers the others as links; each one asks for what it actually needs (a "Get
Code" field for SMS/email, an authenticator code, the code from a push
notification, or the RADIUS password), and a recovery code always works instead.
The same panel appears on `/callback` and `/callback/saml`, because a provider's
authorization code is single-use — sending the user back to `/login` would drop
the pending sign-in.

**The static callback page** — `routers/lightweight_auth_filter.go` answers
`/callback?state=...` with a small page that posts the provider's code itself,
so a returning visitor does not wait for the bundle. When it meets something it
cannot finish it stores what it already got under
`casdoor_callback_react_fallback` and comes here with
`__casdoor_callback_react=1`; `AuthCallback` continues from that stored response
rather than spending the single-use code again, and ignores a payload whose
`search` does not match this callback. `?provider_hint=<name>` likewise jumps
straight to that provider, which is what happens when the static redirect page
is unavailable or hands the request back.

The application edit page keeps the antd page's own shape: the same eight tabs
(Basic, Authentication, OIDC/OAuth, SAML, Providers, UI Customization, Security,
Reverse Proxy) holding the same fields, the open one kept in the URL hash, and
the "Menu mode" switch that lays those tabs across the top or down the side.

The application and organization edit pages carry the same sub-tables the antd
frontend has, down to the option list each row offers: a provider row's category,
type, country codes, binding rule, signup group and the rule its own kind calls
for; a signup item's type, custom CSS and choice options; a token attribute's
"Static Value" / "Existing Field" pair; and the rule that only lets one MFA
factor be `Required`.

**Console** — every list page carries the same per-column search, sorting and
filter menus the antd tables offer. A filter is a canned search: antd turned the
picked value into the `field`/`value` pair of the list API, and so does
`ColumnDef.filters` here, which is why the provider type filter can keep its
two-level "category → type" menu. A read the backend refuses with "Unauthorized
operation" replaces the page with the 403 `UnauthorizedPage`, as `BaseListPage`
did, rather than showing an empty table; the same check drives the read-only
demo site's "go to the writable demo" prompt in `lib/fetch-filter.ts`. Stored
enum values (ticket and subscription states, permission
actions and effects, coupon discount types, ...) render as translated badges from
one map per enum in `lib/enum-labels.tsx`, and the billing objects (coupons,
orders, payments, plans, pricings, products, subscriptions) open read-only for
anyone who is not a local admin — "View" instead of "Edit", no Add, no Delete, a
locked form and no Save. Edit pages title themselves "New X" / "View X" /
"Edit X" after their mode.

The per-page actions are there too: duplicating an application, uploading the
Terms of Use HTML, resetting the footer HTML, previewing a webhook's payload,
copying a pricing page URL, the "Test buy page" link, the save-time checks on
permissions and products, the block on deleting a group that still has
subgroups, "Show all" on the group tree, and the LDAP sync page's synced /
unsynced marks, group ids and per-user results. A record opens in a drawer with
its whole response and object, because the row is far too wide to read in the
table; a payment links to its result page; a transaction can be a "Recharge",
which is POSTed before the edit page opens it under that title; and a Client IP
links to db-ip.com, as in the antd tables.

Dashboard, apps, shortcuts, my account, system info, breadcrumbs,
and list + edit pages for: organizations, users, groups (incl. tree), invitations,
applications, providers, resources, certs, keys, roles, permissions, models,
adapters, enforcers, agents, MCP servers, entries, sites, rules, sessions,
records, tokens, verifications, products, coupons, orders, payments, plans,
pricings, subscriptions, transactions, forms, syncers, webhooks, webhook events,
tickets, LDAP (edit + sync).

**Branding** — the console takes its tab title and favicon from the signed-in
user's organization, brands the sidebar with that organization's logo (its
`logoDark` in dark mode, its favicon when the sidebar is collapsed), and both the
console and the sign-in pages close with the Casdoor wordmark rather than the
word — the same footer image the antd frontend renders, and the one
`Setting.getDefaultFooterContent()` writes into an application's `footerHtml`.
`index.html` ships the `https://cdn.casbin.org/img/favicon.png` literal that
`routers/static_filter.go` rewrites, so the first paint is branded too.

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

**Entry viewers** — an entry renders with a viewer chosen by what produced it:
the SELinux audit fields of a `SELinux Log` provider, the spans of an OTLP
`trace` message (with a per-span drawer), or the OpenClaw session graph — a
react-flow tree with a node drawer that pairs each tool call with its result,
fullscreen, and a link to the raw JSONL transcript at
`/entries/:organizationName/:entryName/transcript`. The list page's message cell
opens the same viewer in a popover.

**MFA reminders** — after sign-in, an organization that marks an MFA item
`Prompted` gets a toast recommending it, with "Later" and "Go to enable"; one
marked `Required` gets a warning and the redirect to `/mfa/setup`.

## Deliberately not ported

- **Web3**: `auth/Web3Auth.ts` talks to `window.ethereum` directly rather than
  going through `@web3-onboard`, so the extra wallets that library bundles
  (Coinbase, Phantom, Trust, Gnosis, ...) are not offered. MetaMask works.

## Still to do

- **i18n**: nine strings this frontend adds (`form:New Form`,
  `organization:Token retention days`, `provider:Advanced`, ...) are English in
  all ten non-English locales.
- **i18n plumbing**: `i18n/generate.go` now walks this directory's `.ts`/`.tsx`
  sources, but the generator has not been re-run against them, so the catalogs
  are still the ones the antd console produced.

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
