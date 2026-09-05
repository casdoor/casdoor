# Code style

- Do not add Go tests (`_test.go`) unless explicitly asked.
- Keep comments sparse: only for genuinely non-obvious mechanics, not one per block.

# Frontend migration (`web-old` → `web`)

`web` is the shadcn/Tailwind rewrite of the antd frontend in `web-old`. Apart from
the UI layer, `web` should behave exactly like `web-old` — same validations, same
field formats, same conditions. When the two differ, `web-old` is the reference.

Not migrated on purpose:

- **AI assistant.** The `aiAssistantUrl` config, its header button and the iframe
  drawer (`Conf.AiAssistantUrl` / `renderAiAssistant()` in `web-old/src/App.js`,
  the `ai-assistant` entry of `WidgetItemTree`) are dropped. Do not port them.
