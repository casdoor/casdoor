# Code style

- Do not add Go tests (`_test.go`) unless explicitly asked.
- Keep comments sparse: only for genuinely non-obvious mechanics, not one per block.

# Commercial and enterprise code (`ee`)

The commercial and enterprise edition features live in `ee/` (Go) and `web/src/ee/` (frontend), under the Casdoor
Enterprise License in `ee/LICENSE`; everything else is Apache 2.0. New paid features go there, not into `object/`,
`controllers/` or `web/src/pages/`.

- `ee` imports `object`, never the other way: when core code needs paid logic, add a hook to
  `object/edition_hooks.go` (with the open source behaviour as the default) and set it in `ee/hooks.go`.
- Types that are fields of core structs (`SecurityPolicy`, `EnterprisePolicy`, `FormFill`, ...) stay in
  `object/edition_types.go`; tables used only by paid features live in `ee` and are added with `object.RegisterTables`.
- Paid APIs are methods of `ee/controllers.ApiController`, which embeds the core one; route them with
  `&eecontrollers.ApiController{}` in `routers/router.go`.
- Go files in `ee` start with the enterprise license header, not the Apache one.

# Frontend migration (`web-old` → `web`)

`web` is the shadcn/Tailwind rewrite of the antd frontend in `web-old`. Apart from
the UI layer, `web` should behave exactly like `web-old` — same validations, same
field formats, same conditions. When the two differ, `web-old` is the reference.

`web-old` is deprecated and read-only: never modify it (code or locales). New
features and fixes go into `web` only.

Not migrated on purpose:

- **AI assistant.** The `aiAssistantUrl` config, its header button and the iframe
  drawer (`Conf.AiAssistantUrl` / `renderAiAssistant()` in `web-old/src/App.js`,
  the `ai-assistant` entry of `WidgetItemTree`) are dropped. Do not port them.
- **Web3 providers.** The `Web3` category (MetaMask, Web3-Onboard) was removed
  from the backend and `web`. Do not port `web-old`'s wallet sign-in code.
  Wallet sign-in is the paid Marketplace plugin `web3-wallet` (an idp action
  with a sign-in script, see `web/src/ee/lib/plugin-signin.ts`); keep its
  code in the plugin, not in Casdoor.
