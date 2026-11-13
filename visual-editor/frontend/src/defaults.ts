import type { BlendMode, Layer } from './types'

// The project shown on first visit (and after "New"). It demonstrates the
// core idea: debug wraps core.factorial without touching it — hide the
// debug layer and run again to see the difference.
export const DEFAULT_LAYERS: Layer[] = [
  {
    id: 'debug',
    name: 'debug',
    enabled: true,
    color: '#ff9800',
    code: `# Wraps core.factorial with logging, without changing core.
# Hide this layer (eye icon) and the logs disappear.
fn factorial(n) =
  let _ = log("factorial({n})") in
    core.factorial(n)`,
  },
  {
    id: 'ui',
    name: 'ui',
    enabled: true,
    color: '#2196f3',
    code: `# Renders an interactive UI. Run app() from the main expression.
fn app() =
  view({
    title: "Factorial",
    children: [
      input({name: "n", placeholder: "Enter a number", type: "number"}),
      button({label: "Compute", onclick: "factorial(n)"}),
      text("5! = {factorial(5)}")
    ]
  })`,
  },
  {
    id: 'data',
    name: 'data',
    enabled: false,
    color: '#9c27b0',
    code: `# Effects with the Option pattern: results are Some(value) or None.
fn save_result(n, result) =
  let _ = write_json("results.json", {input: n, output: result}) in
    result`,
  },
  {
    id: 'core',
    name: 'core',
    enabled: true,
    color: '#4caf50',
    locked: true,
    code: `# Pure functions. Unqualified recursive calls go back through the
# whole stack, so a layer above can observe every step.
fn factorial(n) =
  match n with
  | 0 -> 1
  | n -> n * factorial(n - 1)`,
  },
]

export const DEFAULT_MAIN = `# The entry point. Try: app()
factorial(5)`

export const DEFAULT_BLEND: BlendMode = 'override'
