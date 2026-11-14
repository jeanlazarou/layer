# Layer Visual Editor

A browser editor for the Layer language where the layer stack is the main interface. It runs entirely
client-side: the Go interpreter is compiled to WebAssembly, so there is no server.

## Running it

```bash
cd layer-go && ./build-wasm.sh          # builds public/layer.wasm + wasm_exec.js (needs Go)
cd ../visual-editor/frontend
pnpm install
pnpm dev                                # http://localhost:5173
```

A prebuilt `public/layer.wasm` is committed, so `pnpm dev` also works without Go as long as you don't change
the interpreter. `pnpm build` produces a static site in `dist/` that can be hosted anywhere (it uses relative
URLs, so a sub-path such as GitHub Pages' `/layer/` works). The repository's
[`pages.yml`](../.github/workflows/pages.yml) workflow publishes it.

## What you can do

- **Layer stack (left).** Top layers answer calls first. Click the eye to hide a layer (it stays in the file
  but is not listed in `@layers`), drag to reorder, add custom layers, and click a custom layer's
  `pure`/`impure` badge to toggle `@layer<impure>`. `core` is pinned at the bottom because the interpreter
  always places it there.
- **Run** with the toolbar button or **⌘/Ctrl + Enter**. After the first run, toggling, reordering or changing
  the blend mode re-runs automatically, so you see the effect of each layer immediately.
- **Blend mode** selector in the toolbar: override, chain, collect, fallback.
- **Live diagnostics.** Syntax errors, undefined names, unknown layers and purity violations are reported as
  you type, as editor markers and as an error count on the layer card. Runtime errors in the Output panel
  show the layer, line and Layer call trace; click one to jump to it.
- **Who answers each call?** (tree icon). For every function, it lists which layers define it, which one
  answers under its blend mode, and which are shadowed. Functions defined in several layers get a mode
  selector, which writes per-function modes (`blend: {fetch: fallback}`) into the program.
- **Examples** menu with every program from [`examples/`](../examples/), plus **Open/Save** of `.layer`
  files (Save keeps hidden layers). Your work is also kept in the browser's localStorage between visits.
- **Interactive apps.** When the main expression returns `app({init, update, view})`, the Output panel runs
  it live with a model/message inspector (see below); a plain component is shown as a static UI.

## Interactive apps

When the main expression returns `app({init, update, view})`, the Output panel runs the app:

- The view is rendered with Material UI (`view`, `text`, `button`, `input`, `row`/`column`). Buttons send
  their `on_click` message; inputs send `{tag: text}` for `on_input: "tag"` (numbers for `type: "number"`).
  Messages and models are plain data, passed to the interpreter as JSON — never pasted into source code.
- The inspector beside it shows the current **model** and every **message**. Click a message to go back to
  the state after it; the next message continues from there. Keystrokes in one input are merged into one
  history entry.
- **Run** again (or toggle, reorder or re-blend a layer) re-renders the app with the new code and keeps its
  model — hot reload. **Restart** starts again from `init`.

Each message is one `executeLayer(source, {model, messages: [msg]})` call: the interpreter keeps no state
between events, the editor keeps the history. A main expression that returns a plain component is shown as a
static UI.

## Code map

```
frontend/src/
  App.tsx                     state, toolbar, running, app sessions (messages, time travel, hot reload)
  program.ts                  editor state ⇄ Layer source; maps errors back to layers
  defaults.ts                 the project shown on first visit
  examples.ts                 bundles ../../examples/*.layer
  storage.ts                  localStorage workspace
  language/layerLanguage.ts   Monaco syntax highlighting + completion for Layer
  components/
    LayerStack.tsx            layer cards, purity badges, diff pills
    CodeEditor.tsx            Monaco wrapper with error markers
    OutputConsole.tsx         logs, results, clickable errors, UI preview, app inspector
    UIPreview.tsx             renders ui-layer XML as live components; buttons/inputs send messages
    ResolutionPanel.tsx       "who answers each call?"
    HelpDialog.tsx
  wasm/
    layerWasm.ts              loads public/layer.wasm; typed run/describe/split calls
    browserFileSystem.ts      data-layer file I/O backed by localStorage
```

The WASM module exports three functions returning JSON (see
[`layer-go/cmd/wasm/main.go`](../layer-go/cmd/wasm/main.go)): `executeLayer(source, appStateJSON?)`,
`describeLayer(source)` and `parseLayerFile(source)`.

## Checks

```bash
pnpm typecheck && pnpm lint && pnpm build
pnpm exec playwright install chromium     # once
pnpm test:e2e                             # browser tests: builds, serves dist/, runs e2e/*.spec.ts
E2E_BASE_URL=https://jeanlazarou.github.io/layer/ pnpm test:e2e   # against a deployed copy
```

The browser tests cover running, hiding layers, live diagnostics, located runtime errors, apps (messages,
inputs, time travel, hot reload, Restart, collected panels, persistence), examples, per-function blend
modes, autosave across reloads, and the help dialog. CI runs them on every
push and again against the live site after each Pages deploy.

## Known limitations

- Monaco is loaded from a CDN by `@monaco-editor/react`, so the editor needs network access on first load.
- Programs run on the main thread; a very long computation blocks the page until it finishes. Runaway
  recursion stops with an error at a depth of 20,000 calls.
