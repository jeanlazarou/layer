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
  result: string
  kind: 'value' | 'ui' | 'none'
  logs: string[]
  errors: Issue[]
}

export interface Description {
  blendMode: BlendMode
  layers: { name: string; pure: boolean; defines: string[] }[]
  resolutions: { name: string; layers: string[] }[]
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
  mainExpression?: string
}
