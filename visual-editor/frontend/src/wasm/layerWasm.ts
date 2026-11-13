// Layer WASM module loader and typed interface.
//
// The Go module (layer-go/cmd/wasm) exports three functions that take source
// text and return JSON strings; this wrapper loads the module once and
// parses the results.
import './browserFileSystem'
import type { DescribeResult, RunResult, SplitResult } from '../types'

declare global {
  interface Window {
    // Provided by wasm_exec.js
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    Go: any
    executeLayer?: (source: string, varsJSON?: string) => string
    describeLayer?: (source: string) => string
    parseLayerFile?: (source: string) => string
  }
}

// Respect Vite's base path so the editor works when deployed under a
// sub-path (e.g. GitHub Pages at /layer/).
const asset = (name: string) => `${import.meta.env.BASE_URL}${name}`

let loading: Promise<void> | null = null

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = src
    script.onload = () => resolve()
    script.onerror = () => reject(new Error(`Failed to load script: ${src}`))
    document.head.appendChild(script)
  })
}

async function loadWasm(): Promise<void> {
  await loadScript(asset('wasm_exec.js'))
  const go = new window.Go()
  const response = await fetch(asset('layer.wasm'))
  if (!response.ok) {
    throw new Error(`Could not fetch layer.wasm (${response.status}). Run layer-go/build-wasm.sh first.`)
  }
  // instantiate (not instantiateStreaming): works even when a static host
  // serves .wasm with the wrong MIME type.
  const { instance } = await WebAssembly.instantiate(await response.arrayBuffer(), go.importObject)
  // go.run resolves only when the Go program exits, which it never does.
  void go.run(instance)
}

/** Load the interpreter (idempotent). */
export function loadLayer(): Promise<void> {
  if (!loading) {
    loading = loadWasm().catch((error) => {
      loading = null
      throw error
    })
  }
  return loading
}

function internalError(message: string) {
  return { success: false, errors: [{ kind: 'internal' as const, line: 0, column: 0, message }] }
}

async function call<T>(fn: 'executeLayer' | 'describeLayer' | 'parseLayerFile', ...args: string[]): Promise<T> {
  try {
    await loadLayer()
  } catch (error) {
    return internalError(`Interpreter failed to load: ${error instanceof Error ? error.message : error}`) as T
  }
  const exported = window[fn] as ((...a: string[]) => string) | undefined
  if (!exported) {
    return internalError(`${fn} is not available — the interpreter may have crashed; reload the page`) as T
  }
  try {
    return JSON.parse(exported(...args)) as T
  } catch (error) {
    return internalError(`${fn} failed: ${error instanceof Error ? error.message : error}`) as T
  }
}

/**
 * Run a program. `vars` are bound as top-level variables (current values of
 * UI inputs) — passed as data, never spliced into the source.
 */
export async function runProgram(source: string, vars?: Record<string, string | number>): Promise<RunResult> {
  const result = await call<Partial<RunResult>>('executeLayer', source, ...(vars ? [JSON.stringify(vars)] : []))
  return { success: false, errors: [], result: '', kind: 'none', logs: [], ...result }
}

/** Load a program without running it and describe its layer stack. */
export function describeProgram(source: string): Promise<DescribeResult> {
  return call<DescribeResult>('describeLayer', source)
}

/** Split a .layer file into its layers, blend mode, and main expression. */
export function splitLayerFile(source: string): Promise<SplitResult> {
  return call<SplitResult>('parseLayerFile', source)
}
