/**
 * Type definitions for Layer Visual Editor
 */

export interface Layer {
  id: string
  name: string
  enabled: boolean
  color: string
  locked?: boolean
  code: string
}

export interface ExecuteResponse {
  success: boolean
  result?: string
  logs?: string[]
  error?: string
}

export type LayerId = string | 'main'
