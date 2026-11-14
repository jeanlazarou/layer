import React from 'react'
import { Dialog, DialogTitle, DialogContent, Typography, Box, IconButton, Link } from '@mui/material'
import CloseIcon from '@mui/icons-material/Close'

interface HelpDialogProps {
  open: boolean
  onClose: () => void
}

const REPO = 'https://github.com/jeanlazarou/layer'

const Code: React.FC<{ children: string }> = ({ children }) => (
  <Box component="pre" sx={{ m: 0, mb: 1.5, p: 1, bgcolor: '#f5f5f5', borderRadius: 1, fontSize: '0.78rem', fontFamily: 'monospace', overflowX: 'auto' }}>
    {children}
  </Box>
)

const Section: React.FC<{ title: string; children: React.ReactNode }> = ({ title, children }) => (
  <Box sx={{ mb: 2 }}>
    <Typography variant="subtitle2" sx={{ fontWeight: 700, mb: 0.5 }}>
      {title}
    </Typography>
    {children}
  </Box>
)

const HelpDialog: React.FC<HelpDialogProps> = ({ open, onClose }) => (
  <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
    <DialogTitle sx={{ display: 'flex', alignItems: 'center' }}>
      <Box sx={{ flexGrow: 1 }}>Layer in two minutes</Box>
      <IconButton onClick={onClose} size="small" aria-label="close">
        <CloseIcon />
      </IconButton>
    </DialogTitle>
    <DialogContent dividers>
      <Section title="The idea">
        <Typography variant="body2" sx={{ mb: 1 }}>
          A program is a stack of layers, like an image editor. An unqualified call such as <code>fib(5)</code> is answered by the
          topmost layer that defines <code>fib</code>; <code>core.fib(5)</code> calls a specific layer. Hide, reorder, or add layers on
          the left to change what the program does, without editing the code inside them.
        </Typography>
        <Code>{`@layer debug {
  fn fib(n) =
    let _ = log("fib({n})") in   # wrap: log, then call the layer below
      core.fib(n)
}`}</Code>
      </Section>

      <Section title="Syntax">
        <Code>{`fn add(x, y) = x + y                  # function (body is one expression)
λ(x) -> x * 2                         # lambda
let r = add(1, 2) in r * 10           # let binding
if n < 0 then "neg" else "pos"        # conditional
match xs with                         # pattern matching
| []            -> 0
| [h, ...rest] when h > 0 -> h + sum(rest)
| {name: n}     -> n                  # record pattern
"fib({n}) = {fib(n)}"                 # string interpolation ('single quotes' don't)
add(1, ?)                             # partial application`}</Code>
      </Section>

      <Section title="Purity">
        <Typography variant="body2">
          <code>core</code> and custom layers are <strong>pure</strong>: they cannot call effects such as <code>log</code> or data I/O.
          <code> ui</code>, <code>debug</code> and <code>data</code> are impure; click a custom layer's badge to mark it{' '}
          <code>@layer&lt;impure&gt;</code>. Higher impure layers may still wrap pure functions — that is how tracing works.
        </Typography>
      </Section>

      <Section title="Blend modes (toolbar)">
        <Typography variant="body2" component="div">
          <ul style={{ margin: 0, paddingLeft: 20 }}>
            <li><strong>override</strong> — the top definition wins (default)</li>
            <li><strong>chain</strong> — bottom runs first; each layer above receives the previous result</li>
            <li><strong>collect</strong> — every layer runs; results come back as a list</li>
            <li><strong>fallback</strong> — try top-down; the first result that is not <code>None</code> wins</li>
          </ul>
          A function defined in several layers can choose its own mode in the “Who answers each call?” panel (in
          source: <code>blend: {'{'}fetch: fallback, total: chain{'}'}</code>). Combining modes apply once per call:
          recursion inside a layer stays in that layer.
        </Typography>
      </Section>

      <Section title="Interactive apps">
        <Typography variant="body2" sx={{ mb: 1 }}>
          A main expression that returns <code>app({'{'}init, update, view{'}'})</code> runs as a live app. The model is
          plain data; buttons and inputs send <strong>messages</strong> (also plain data), and <code>update(msg, model)</code>{' '}
          returns the next model. The Output panel shows the model and every message — click one to go back to that
          state. Running again (or toggling a layer) keeps the model; <strong>Restart</strong> starts from{' '}
          <code>init</code>.
        </Typography>
        <Code>{`fn update(msg, model) =
  match msg with
  | "inc"     -> {model | count: model.count + 1}   # record update
  | {step: s} -> {model | step: s}

fn page(model) = view({title: "Counter", children: [
  text("Count: {model.count}"),
  button({label: "+", on_click: "inc"}),
  input({type: "number", value: model.step, on_input: "step"})  # sends {step: n}
]})

app({init: {count: 0, step: 1}, update: update, view: page})`}</Code>
        <Typography variant="body2">
          Layers wrap apps too: a debug layer can redefine <code>update</code> to log each message, and with{' '}
          <code>blend: {'{'}page: collect{'}'}</code> every layer can add a panel. See the app examples.
        </Typography>
      </Section>

      <Section title="Shortcuts">
        <Typography variant="body2">
          <strong>⌘/Ctrl + Enter</strong> runs the program. Your work is saved in this browser automatically; use Save to download a{' '}
          <code>.layer</code> file you can run with the CLI.
        </Typography>
      </Section>

      <Typography variant="body2">
        More: <Link href={`${REPO}#readme`} target="_blank" rel="noopener">README</Link> ·{' '}
        <Link href={`${REPO}/blob/main/LANGUAGE_SPEC.md`} target="_blank" rel="noopener">Language spec</Link> ·{' '}
        <Link href={`${REPO}/blob/main/docs/TUTORIAL.md`} target="_blank" rel="noopener">Tutorial</Link>
      </Typography>
    </DialogContent>
  </Dialog>
)

export default HelpDialog
