import axios from 'axios'
import type { ExecuteResponse } from '../types'

export const executeLayerCode = async (code: string): Promise<ExecuteResponse> => {
  try {
    const response = await axios.post<ExecuteResponse>('/api/execute', {
      code: code
    })
    return response.data
  } catch (error) {
    if (axios.isAxiosError(error) && error.response) {
      return {
        success: false,
        error: error.response.data.error || 'Server error'
      }
    }
    throw new Error('Failed to connect to backend server')
  }
}
