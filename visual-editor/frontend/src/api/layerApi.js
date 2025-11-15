import axios from 'axios';

export const executeLayerCode = async (code) => {
  try {
    const response = await axios.post('/api/execute', {
      code: code
    });
    return response.data;
  } catch (error) {
    if (error.response) {
      return {
        success: false,
        error: error.response.data.error || 'Server error'
      };
    }
    throw new Error('Failed to connect to backend server');
  }
};
