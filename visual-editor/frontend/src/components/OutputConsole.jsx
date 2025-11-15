import React from 'react'
import { Box, Typography } from '@mui/material'
import TerminalIcon from '@mui/icons-material/Terminal'

function OutputConsole({ output }) {
  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', bgcolor: '#1e1e1e' }}>
      {/* Console Header */}
      <Box
        sx={{
          p: 1.5,
          display: 'flex',
          alignItems: 'center',
          bgcolor: '#2d2d2d',
          borderBottom: 1,
          borderColor: '#444',
        }}
      >
        <TerminalIcon sx={{ fontSize: 18, mr: 1, color: '#4caf50' }} />
        <Typography variant="subtitle2" sx={{ color: '#fff', fontWeight: 600 }}>
          Output Console
        </Typography>
      </Box>

      {/* Console Content */}
      <Box
        sx={{
          flex: 1,
          p: 2,
          overflow: 'auto',
          fontFamily: '"Fira Code", "Courier New", monospace',
          fontSize: 13,
          color: '#d4d4d4',
        }}
      >
        <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
          {output || '> Ready to execute...\n> Click the play button to run your Layer program.'}
        </pre>
      </Box>
    </Box>
  )
}

export default OutputConsole
