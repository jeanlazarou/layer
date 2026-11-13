// Persists the editor workspace in localStorage so a refresh never loses
// work. Storage can be unavailable (private mode, blocked site data), so
// every access is guarded and failures fall back to the defaults.
import type { BlendMode, Layer } from './types'

const KEY = 'layer-editor:workspace:v1'

export interface Workspace {
  layers: Layer[]
  mainCode: string
  blendMode: BlendMode
  projectName: string
}

export function loadWorkspace(): Workspace | null {
  try {
    const raw = window.localStorage.getItem(KEY)
    if (!raw) return null
    const ws = JSON.parse(raw) as Partial<Workspace>
    if (!Array.isArray(ws.layers) || typeof ws.mainCode !== 'string') return null
    return {
      layers: ws.layers,
      mainCode: ws.mainCode,
      blendMode: ws.blendMode ?? 'override',
      projectName: ws.projectName ?? '',
    }
  } catch {
    return null
  }
}

export function saveWorkspace(ws: Workspace): void {
  try {
    window.localStorage.setItem(KEY, JSON.stringify(ws))
  } catch {
    // Storage full or unavailable: the editor keeps working without it.
  }
}
