import type { ExecuteResponse } from '../types'
import layerWasm from '../wasm/layerWasm'

export interface ParseResponse {
  success: boolean
  enabledLayers?: string[]
  layers?: Record<string, string>
  mainExpression?: string
  error?: string
}

/**
 * Execute Layer code using WebAssembly
 */
export const executeLayerCode = async (code: string): Promise<ExecuteResponse> => {
  try {
    const result = await layerWasm.execute(code)

    return {
      success: result.success,
      result: result.result,
      logs: result.logs || [],
      error: result.error
    }
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error'
    }
  }
}

/**
 * Parse a Layer file using WebAssembly
 */
export const parseLayerFile = async (content: string): Promise<ParseResponse> => {
  try {
    const result = await layerWasm.parseFile(content)

    return {
      success: result.success,
      enabledLayers: result.enabledLayers,
      layers: result.layers,
      mainExpression: result.mainExpression,
      error: result.error
    }
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error'
    }
  }
}
