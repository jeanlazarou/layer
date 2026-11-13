// Assembling editor state into a Layer program, mapping interpreter errors
// back to the layer they came from, and turning a split .layer file back
// into editor layers.
import type { BlendMode, Issue, Layer, LayerId, LocatedIssue, SplitResult } from './types'
import { BLEND_MODES, isBuiltinLayer } from './types'

export const LAYER_COLORS: Record<string, string> = {
  core: '#4caf50',
  ui: '#2196f3',
  debug: '#ff9800',
  data: '#9c27b0',
}

const CUSTOM_COLORS = ['#f44336', '#009688', '#e91e63', '#3f51b5', '#795548', '#00bcd4']

export function colorFor(name: string, index = 0): string {
  return LAYER_COLORS[name] ?? CUSTOM_COLORS[index % CUSTOM_COLORS.length]
}

/** Where one layer's (or the main expression's) code sits in a program. */
interface Segment {
  target: LayerId
  /** 1-based line of the first line of the target's code. */
  firstLine: number
  lineCount: number
  /** Columns of indentation added in front of every line. */
  indent: number
}

export interface Program {
  source: string
  segments: Segment[]
}

interface BuildOptions {
  /** Main expression to use instead of the editor's (UI button handlers). */
  main?: string
  /**
   * Also emit disabled layers as @layer blocks. They are not listed in
   * @layers, so they stay inactive — this only preserves their code (Save).
   */
  includeDisabled?: boolean
  /** Indent layer bodies (for readable saved files). */
  indent?: number
  header?: string[]
}

/**
 * Keep core at the bottom of the stack: the interpreter always places core
 * first, whatever its position in @layers, so the editor must not suggest
 * otherwise.
 */
export function normalizeOrder(layers: Layer[]): Layer[] {
  const core = layers.filter(l => l.name === 'core')
  return [...layers.filter(l => l.name !== 'core'), ...core]
}

/** Assemble editor layers (top of the stack first) into a program. */
export function buildProgram(layers: Layer[], mainCode: string, blendMode: BlendMode, opts: BuildOptions = {}): Program {
  const indent = opts.indent ?? 0
  const pad = ' '.repeat(indent)
  const lines: string[] = [...(opts.header ?? [])]
  const segments: Segment[] = []

  const bottomUp = [...layers].reverse()
  const enabled = bottomUp.filter(l => l.enabled)
  const blend = blendMode !== 'override' ? ` blend: ${blendMode}` : ''
  lines.push(`@layers [${enabled.map(l => l.name).join(', ')}]${blend}`, '')

  for (const layer of opts.includeDisabled ? bottomUp : enabled) {
    const annotation = !isBuiltinLayer(layer.name) && layer.impure ? '<impure>' : ''
    lines.push(`@layer${annotation} ${layer.name} {`)
    const code = layer.code.split('\n')
    segments.push({ target: layer.id, firstLine: lines.length + 1, lineCount: code.length, indent })
    lines.push(...code.map(line => (line ? pad + line : line)))
    lines.push('}', '')
  }

  const main = (opts.main ?? mainCode).split('\n')
  segments.push({ target: 'main', firstLine: lines.length + 1, lineCount: main.length, indent: 0 })
  lines.push(...main)

  return { source: lines.join('\n'), segments }
}

/** Map an issue's program position back to the layer it belongs to. */
export function locateIssue(issue: Issue, program: Program): LocatedIssue {
  for (const seg of program.segments) {
    if (issue.line >= seg.firstLine && issue.line < seg.firstLine + seg.lineCount) {
      return {
        ...issue,
        target: seg.target,
        targetLine: issue.line - seg.firstLine + 1,
        targetColumn: Math.max(1, issue.column - seg.indent),
      }
    }
  }
  return { ...issue, target: null, targetLine: issue.line, targetColumn: issue.column }
}

export function describeLocation(issue: LocatedIssue, layers: Layer[]): string {
  if (issue.target === null) return issue.line > 0 ? `line ${issue.line}` : ''
  const name = issue.target === 'main' ? 'main' : layers.find(l => l.id === issue.target)?.name ?? '?'
  return `${name}:${issue.targetLine}:${issue.targetColumn}`
}

/** Serialize the whole editor state as a .layer file. */
export function serializeProject(layers: Layer[], mainCode: string, blendMode: BlendMode): string {
  return buildProgram(layers, mainCode, blendMode, {
    includeDisabled: true,
    indent: 4,
    header: ['# Layer program — saved by the Layer Visual Editor', ''],
  }).source + '\n'
}

export interface Project {
  layers: Layer[]
  mainCode: string
  blendMode: BlendMode
}

/**
 * Turn a split .layer file into editor layers. Order (top first): active
 * layers as listed in @layers, then declared-but-inactive layers, then the
 * built-in layers the file doesn't mention (disabled), with core last.
 */
export function projectFromSplit(split: SplitResult): Project {
  const enabled = split.enabledLayers ?? []
  const declared = split.declaredLayers ?? []
  const code = split.layers ?? {}
  const purity = split.purity ?? {}

  const ordered: string[] = []
  const add = (name: string) => {
    if (!ordered.includes(name)) ordered.push(name)
  }
  ;[...enabled].reverse().forEach(add)
  ;[...declared].reverse().forEach(add)
  ;['debug', 'ui', 'data', 'core'].forEach(add)

  let customIndex = 0
  const layers: Layer[] = ordered.map(name => ({
    id: isBuiltinLayer(name) ? name : `custom_${name}`,
    name,
    enabled: name === 'core' || enabled.includes(name),
    color: colorFor(name, isBuiltinLayer(name) ? 0 : customIndex++),
    locked: name === 'core',
    code: code[name] ?? `# ${name} layer\n`,
    impure: purity[name] === 'impure' ? true : undefined,
  }))

  const blend = BLEND_MODES.some(m => m.value === split.blendMode) ? (split.blendMode as BlendMode) : 'override'
  return { layers: normalizeOrder(layers), mainCode: split.mainExpression ?? '', blendMode: blend }
}
