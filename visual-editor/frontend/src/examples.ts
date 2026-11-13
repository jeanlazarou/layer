// The repository's examples/*.layer programs, bundled into the editor so
// they can be opened from the Examples menu.
const files = import.meta.glob('../../../examples/*.layer', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>

export interface Example {
  id: string
  title: string
  description: string
  source: string
}

// Shown first, in this order; the rest follow alphabetically.
const FEATURED = [
  'fibonacci',
  'debug_trace',
  'blend_fallback_cache',
  'blend_collect_logging',
  'simple_calculator',
  'calculator',
]

function describe(source: string): string {
  // First comment line that isn't the title.
  const comments = source
    .split('\n')
    .filter(l => l.startsWith('#'))
    .map(l => l.replace(/^#\s?/, '').trim())
    .filter(Boolean)
  return comments[1] ?? comments[0] ?? ''
}

function titleOf(id: string, source: string): string {
  const first = source.split('\n').find(l => l.startsWith('#'))
  const title = first?.replace(/^#\s?/, '').trim()
  return title || id.replace(/_/g, ' ')
}

export const EXAMPLES: Example[] = Object.entries(files)
  .map(([path, source]) => {
    const id = path.split('/').pop()!.replace(/\.layer$/, '')
    return { id, title: titleOf(id, source), description: describe(source), source }
  })
  .sort((a, b) => {
    const ra = FEATURED.indexOf(a.id)
    const rb = FEATURED.indexOf(b.id)
    if (ra !== -1 || rb !== -1) return (ra === -1 ? 99 : ra) - (rb === -1 ? 99 : rb)
    return a.id.localeCompare(b.id)
  })
