import React, { useState, useMemo } from 'react'
import { Box, Typography, IconButton, Tooltip } from '@mui/material'
import TerminalIcon from '@mui/icons-material/Terminal'
import VisibilityIcon from '@mui/icons-material/Visibility'
import CodeIcon from '@mui/icons-material/Code'
import OpenInFullIcon from '@mui/icons-material/OpenInFull'
import UIPreview from './UIPreview'

interface OutputConsoleProps {
  output: string
  onMaximize?: () => void
}

const OutputConsole: React.FC<OutputConsoleProps> = ({ output, onMaximize }) => {
  const [viewMode, setViewMode] = useState<'text' | 'preview'>('text')

  // Check if output contains UI XML
  const isUIOutput = useMemo(() => {
    return output.includes('<view') || output.includes('<text') || output.includes('<button')
  }, [output])

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
        <Typography variant="subtitle2" sx={{ color: '#fff', fontWeight: 600, fontSize: '0.8rem', flexGrow: 1 }}>
          Output
        </Typography>

        {/* Toggle button - only show if UI output detected */}
        {isUIOutput && (
          <>
            <Tooltip title="Show XML">
              <IconButton
                size="small"
                onClick={() => setViewMode('text')}
                sx={{
                  color: viewMode === 'text' ? '#4caf50' : '#888',
                  p: 0.5,
                }}
              >
                <CodeIcon sx={{ fontSize: 16 }} />
              </IconButton>
            </Tooltip>
            <Tooltip title="Show Preview">
              <IconButton
                size="small"
                onClick={() => setViewMode('preview')}
                sx={{
                  color: viewMode === 'preview' ? '#4caf50' : '#888',
                  p: 0.5,
                }}
              >
                <VisibilityIcon sx={{ fontSize: 16 }} />
              </IconButton>
            </Tooltip>
          </>
        )}

        {/* Maximize button */}
        {onMaximize && (
          <Tooltip title="Maximize">
            <IconButton
              size="small"
              onClick={onMaximize}
              sx={{
                color: '#888',
                p: 0.5,
                ml: 0.5,
              }}
            >
              <OpenInFullIcon sx={{ fontSize: 16 }} />
            </IconButton>
          </Tooltip>
        )}
      </Box>

      {/* Console Content */}
      {viewMode === 'text' || !isUIOutput ? (
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
      ) : (
        <Box
          sx={{
            flex: 1,
            overflow: 'auto',
            bgcolor: '#f5f5f5',
          }}
        >
          <UIPreview xmlContent={output} />
        </Box>
      )}
    </Box>
  )
}

export default OutputConsole
