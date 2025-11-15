import React from 'react'
import { Box, Typography, Chip } from '@mui/material'
import Editor from '@monaco-editor/react'

interface CodeEditorProps {
  code: string
  layerName: string
  onChange: (code: string) => void
}

const CodeEditor: React.FC<CodeEditorProps> = ({ code, layerName, onChange }) => {
  const handleEditorChange = (value: string | undefined) => {
    onChange(value || '')
  }

  const getLayerColor = (name: string): string => {
    const colors: Record<string, string> = {
      core: '#4caf50',
      ui: '#2196f3',
      debug: '#ff9800',
      main: '#9c27b0',
    }
    return colors[name] || '#757575'
  }

  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Editor Header */}
      <Box
        sx={{
          p: 2,
          borderBottom: 1,
          borderColor: 'divider',
          display: 'flex',
          alignItems: 'center',
          bgcolor: 'background.paper',
        }}
      >
        <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 600 }}>
          Code Editor
        </Typography>
        <Chip
          label={layerName}
          size="small"
          sx={{
            bgcolor: getLayerColor(layerName),
            color: 'white',
            fontWeight: 600,
          }}
        />
      </Box>

      {/* Monaco Editor */}
      <Box sx={{ flex: 1 }}>
        <Editor
          height="100%"
          defaultLanguage="python"
          value={code}
          onChange={handleEditorChange}
          theme="vs-light"
          options={{
            minimap: { enabled: false },
            fontSize: 14,
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            wordWrap: 'on',
          }}
        />
      </Box>
    </Box>
  )
}

export default CodeEditor
