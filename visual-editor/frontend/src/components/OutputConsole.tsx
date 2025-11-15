import React from 'react'
import { Box, Typography } from '@mui/material'
import TerminalIcon from '@mui/icons-material/Terminal'

interface OutputConsoleProps {
  output: string
}

const OutputConsole: React.FC<OutputConsoleProps> = ({ output }) => {
  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', bgcolor: '#1e1e1e' }}>
      {/* Console Header */}
      <Box
        sx={{
          px: 1.5,
          py: 0.75,
          display: 'flex',
          alignItems: 'center',
          bgcolor: '#2d2d2d',
          borderBottom: 1,
          borderColor: '#444',
        }}
      >
        <TerminalIcon sx={{ fontSize: 16, mr: 0.75, color: '#4caf50' }} />
        <Typography variant="subtitle2" sx={{ color: '#fff', fontWeight: 600, fontSize: '0.8rem' }}>
          Output
        </Typography>
      </Box>

      {/* Console Content */}
      <Box
        sx={{
          flex: 1,
          p: 1.5,
          overflow: 'auto',
          fontFamily: '"Fira Code", "Courier New", monospace',
          fontSize: 12,
          color: '#d4d4d4',
        }}
      >
        <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
          {output || '> Ready\n> Click play to execute'}
        </pre>
      </Box>
    </Box>
  )
}

export default OutputConsole
