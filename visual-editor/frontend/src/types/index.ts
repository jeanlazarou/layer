/**
 * Type definitions for the Layer Visual Editor
 */

export interface Layer {
  id: string
  name: string
  enabled: boolean
  color: string
  locked?: boolean
  code: string
  /** Custom layers only: true when declared as @layer<impure>. */
  impure?: boolean
}

export type LayerId = string | 'main'

export type BlendMode = 'override' | 'chain' | 'collect' | 'fallback'

/** The stack's blend configuration: a default plus per-function modes. */
export interface BlendSettings {
  mode: BlendMode
  perFunction: Record<string, BlendMode>
}

export const DEFAULT_BLEND: BlendSettings = { mode: 'override', perFunction: {} }

export const BLEND_MODES: { value: BlendMode; label: string; help: string }[] = [
  { value: 'override', label: 'Override', help: 'Top layer wins; call lower layers explicitly (core.f)' },
  { value: 'chain', label: 'Chain', help: 'Bottom layer runs first; each higher layer receives the previous result' },
  { value: 'collect', label: 'Collect', help: 'Every layer runs; results are returned as a list (top first)' },
  { value: 'fallback', label: 'Fallback', help: 'Try layers top-down; first result that is not None wins' },
]

/** Built-in layers have fixed purity: core is pure, the others are impure. */
export const BUILTIN_LAYERS = ['core', 'ui', 'debug', 'data'] as const

export function isBuiltinLayer(name: string): boolean {
  return (BUILTIN_LAYERS as readonly string[]).includes(name)
}

export function isLayerPure(layer: Pick<Layer, 'name' | 'impure'>): boolean {
  if (layer.name === 'core') return true
  if (isBuiltinLayer(layer.name)) return false
  return !layer.impure
}

/** An error reported by the interpreter, located in the assembled program. */
export interface Issue {
  kind: 'lexer' | 'parser' | 'check' | 'runtime' | 'internal'
  line: number
  column: number
  message: string
  trace?: string
}

/** An issue mapped back to the layer (or main expression) it came from. */
export interface LocatedIssue extends Issue {
  /** Layer id, 'main', or null when the position is outside user code. */
  target: LayerId | null
  /** Line/column within the target's code (1-based). */
  targetLine: number
  targetColumn: number
}

export interface RunResult {
  success: boolean
  /** The value shown to the user; for 'ui' and 'app', the view as XML. */
  result: string
  kind: 'value' | 'ui' | 'app' | 'none'
  /** App only: the current model (plain data, JSON) and its Layer text. */
  model?: unknown
  modelText?: string
  logs: string[]
  errors: Issue[]
}

/** One state of a running app: the message that led to it and the result. */
export interface AppStep {
  /** undefined for the initial state. */
  msg?: unknown
  model: unknown
  modelText: string
  view: string
  /**
   * Consecutive messages from the same source (keystrokes in one input) are
   * merged into one history entry; this identifies the source.
   */
  group?: string
}

/** A running app: every state it went through, and the one on screen. */
export interface AppSession {
  steps: AppStep[]
  index: number
}

export interface Description {
  blendMode: BlendMode
  layers: { name: string; pure: boolean; defines: string[] }[]
  resolutions: { name: string; layers: string[]; mode: BlendMode }[]
}

export interface DescribeResult {
  success: boolean
  description?: Description
  errors: Issue[]
}

export interface SplitResult {
  success: boolean
  errors: Issue[]
  enabledLayers?: string[]
  declaredLayers?: string[]
  layers?: Record<string, string>
  purity?: Record<string, 'pure' | 'impure'>
  blendMode?: string
  blendModes?: Record<string, string>
  mainExpression?: string
}
