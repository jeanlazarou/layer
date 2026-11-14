import type { Layer } from './types'

// The project shown on first visit (and after "New"). It demonstrates the
// core idea: debug wraps core.factorial without touching it — hide the
// debug layer and run again to see the difference. The main expression's
// comment shows how to run the same logic as an interactive app.
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
    code: `# How the app's model looks. Buttons and inputs send messages
# to core.update; the model never changes anywhere else.
fn page(model) =
  view({
    title: "Factorial",
    children: [
      input({name: "n", label: "n", type: "number", value: model.n, on_input: "set_n"}),
      row([
        button({label: "−1", on_click: "dec"}),
        button({label: "+1", on_click: "inc"})
      ]),
      text("{model.n}! = {factorial(model.n)}")
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
    locked: true,
    color: '#4caf50',
    code: `# Pure functions. Unqualified recursive calls go back through the
# whole stack, so a layer above can observe every step.
fn factorial(n) =
  match n with
  | 0 -> 1
  | n -> n * factorial(n - 1)

# The app's model and how messages change it.
fn init() = {n: 5}

fn update(msg, model) =
  match msg with
  | "inc"      -> {model | n: model.n + 1}
  | "dec"      -> if model.n > 0 then {model | n: model.n - 1} else model
  | {set_n: n} -> if n >= 0 then {model | n: n} else model`,
  },
]

export const DEFAULT_MAIN = `# The entry point. To run the interactive app instead, replace it with:
#   app({init: init(), update: update, view: page})
factorial(5)`
