import React, { useState, useCallback } from 'react'
import { Box, AppBar, Toolbar, Typography, IconButton, Tooltip } from '@mui/material'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import SaveIcon from '@mui/icons-material/Save'
import HelpOutlineIcon from '@mui/icons-material/HelpOutline'
import LayerStack from './components/LayerStack'
import CodeEditor from './components/CodeEditor'
import OutputConsole from './components/OutputConsole'
import { executeLayerCode } from './api/layerApi'

const INITIAL_LAYERS = [
  {
    id: 'debug',
    name: 'debug',
    enabled: true,
    color: '#ff9800',
    code: `# Debug layer wraps core functions with logging
fn factorial(n) =
  let _ = log("Computing factorial of {n}") in
    core.factorial(n)`,
  },
  {
    id: 'ui',
    name: 'ui',
    enabled: true,
    color: '#2196f3',
    code: `# UI layer for rendering results
fn app() =
  view({
    title: "Factorial Calculator",
    children: [
      text({content: "Result: {factorial(5)}"})
    ]
  })`,
  },
  {
    id: 'core',
    name: 'core',
    enabled: true,
    color: '#4caf50',
    locked: true,
    code: `# Core layer - pure functional logic
fn factorial(n) =
  match n with
  | 0 -> 1
  | n -> n * factorial(n - 1)`,
  },
]

const INITIAL_MAIN_CODE = `# Main expression
factorial(5)`

function App() {
  const [layers, setLayers] = useState(INITIAL_LAYERS)
  const [selectedLayerId, setSelectedLayerId] = useState('core')
  const [mainCode, setMainCode] = useState(INITIAL_MAIN_CODE)
  const [output, setOutput] = useState('')
  const [isExecuting, setIsExecuting] = useState(false)

  const selectedLayer = layers.find(l => l.id === selectedLayerId)

  const handleLayerReorder = useCallback((newLayers) => {
    setLayers(newLayers)
  }, [])

  const handleLayerToggle = useCallback((layerId) => {
    setLayers(prev => prev.map(layer =>
      layer.id === layerId ? { ...layer, enabled: !layer.enabled } : layer
    ))
  }, [])

  const handleCodeChange = useCallback((code) => {
    if (selectedLayerId === 'main') {
      setMainCode(code)
    } else {
      setLayers(prev => prev.map(layer =>
        layer.id === selectedLayerId ? { ...layer, code } : layer
      ))
    }
  }, [selectedLayerId])

  const handleExecute = async () => {
    setIsExecuting(true)
    setOutput('Executing...\n')

    try {
      // Build the complete Layer program
      const enabledLayers = layers.filter(l => l.enabled).reverse() // Reverse for bottom-to-top
      const layerNames = enabledLayers.map(l => l.name)

      let program = `@layers [${layerNames.join(', ')}]\n\n`

      enabledLayers.forEach(layer => {
        program += `@layer ${layer.name} {\n${layer.code}\n}\n\n`
      })

      program += mainCode

      const result = await executeLayerCode(program)

      if (result.success) {
        let outputText = ''
        if (result.logs && result.logs.length > 0) {
          outputText += result.logs.join('\n') + '\n\n'
        }
        outputText += `Result: ${result.result}`
        setOutput(outputText)
      } else {
        setOutput(`Error: ${result.error}`)
      }
    } catch (error) {
      setOutput(`Error: ${error.message}`)
    } finally {
      setIsExecuting(false)
    }
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      {/* App Bar */}
      <AppBar position="static" elevation={1}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1, fontWeight: 600 }}>
            Layer Visual Editor
          </Typography>
          <Tooltip title="Execute Code">
            <IconButton
              color="inherit"
              onClick={handleExecute}
              disabled={isExecuting}
            >
              <PlayArrowIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Save">
            <IconButton color="inherit">
              <SaveIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Help">
            <IconButton color="inherit">
              <HelpOutlineIcon />
            </IconButton>
          </Tooltip>
        </Toolbar>
      </AppBar>

      {/* Main Content */}
      <Box sx={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Layer Stack - Left Panel */}
        <Box sx={{ width: 280, borderRight: 1, borderColor: 'divider', overflow: 'auto' }}>
          <LayerStack
            layers={layers}
            selectedLayerId={selectedLayerId}
            onLayerSelect={setSelectedLayerId}
            onLayerReorder={handleLayerReorder}
            onLayerToggle={handleLayerToggle}
          />
        </Box>

        {/* Code Editor and Output - Right Panel */}
        <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
          {/* Code Editor */}
          <Box sx={{ flex: 1, overflow: 'hidden' }}>
            <CodeEditor
              code={selectedLayer ? selectedLayer.code : mainCode}
              layerName={selectedLayer ? selectedLayer.name : 'main'}
              onChange={handleCodeChange}
            />
          </Box>

          {/* Output Console */}
          <Box sx={{ height: 250, borderTop: 1, borderColor: 'divider' }}>
            <OutputConsole output={output} />
          </Box>
        </Box>
      </Box>
    </Box>
  )
}

export default App
