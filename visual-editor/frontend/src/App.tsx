import React, { useState, useCallback, useRef } from 'react'
import { Box, AppBar, Toolbar, Typography, IconButton, Tooltip, Dialog, DialogContent, DialogTitle } from '@mui/material'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import SaveIcon from '@mui/icons-material/Save'
import FolderOpenIcon from '@mui/icons-material/FolderOpen'
import AccountTreeIcon from '@mui/icons-material/AccountTree'
import HelpOutlineIcon from '@mui/icons-material/HelpOutline'
import OpenInFullIcon from '@mui/icons-material/OpenInFull'
import CloseFullscreenIcon from '@mui/icons-material/CloseFullscreen'
import LayerStack from './components/LayerStack'
import CodeEditor from './components/CodeEditor'
import OutputConsole from './components/OutputConsole'
import FlowDiagram from './components/FlowDiagram'
import { executeLayerCode, parseLayerFile } from './api/layerApi'
import type { Layer, LayerId } from './types'
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels'

const INITIAL_LAYERS: Layer[] = [
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
  const [layers, setLayers] = useState<Layer[]>(INITIAL_LAYERS)
  const [selectedLayerId, setSelectedLayerId] = useState<LayerId>('core')
  const [mainCode, setMainCode] = useState(INITIAL_MAIN_CODE)
  const [output, setOutput] = useState('')
  const [isExecuting, setIsExecuting] = useState(false)
  const [showFlowDiagram, setShowFlowDiagram] = useState(false)
  const [consoleMaximized, setConsoleMaximized] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const selectedLayer = layers.find(l => l.id === selectedLayerId)

  const handleLayerReorder = useCallback((newLayers: Layer[]) => {
    setLayers(newLayers)
  }, [])

  const handleLayerToggle = useCallback((layerId: string) => {
    setLayers(prev => prev.map(layer =>
      layer.id === layerId ? { ...layer, enabled: !layer.enabled } : layer
    ))
  }, [])

  const handleLockToggle = useCallback((layerId: string) => {
    // Core layer cannot be unlocked
    if (layerId === 'core') return

    setLayers(prev => prev.map(layer =>
      layer.id === layerId ? { ...layer, locked: !layer.locked } : layer
    ))
  }, [])

  const handleAddLayer = useCallback((layerData: Omit<Layer, 'id'>) => {
    const newLayer: Layer = {
      ...layerData,
      id: `custom_${Date.now()}`, // Generate unique ID
    }
    setLayers(prev => [newLayer, ...prev]) // Add at top
    setSelectedLayerId(newLayer.id) // Auto-select the new layer
  }, [])

  const handleCodeChange = useCallback((code: string) => {
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
      const errorMessage = error instanceof Error ? error.message : 'Unknown error'
      setOutput(`Error: ${errorMessage}`)
    } finally {
      setIsExecuting(false)
    }
  }

  const handleInteractiveExecute = useCallback(async (expression: string, inputs: Record<string, string>): Promise<string> => {
    try {
      // Build the Layer program with current layers
      const enabledLayers = layers.filter(l => l.enabled).reverse()
      const layerNames = enabledLayers.map(l => l.name)

      let program = `@layers [${layerNames.join(', ')}]\n\n`

      enabledLayers.forEach(layer => {
        program += `@layer ${layer.name} {\n${layer.code}\n}\n\n`
      })

      // Add input bindings as let expressions
      // Input fields are always strings - Layer code should handle conversions
      const inputBindings = Object.entries(inputs)
        .map(([name, value]) => `let ${name} = "${value}" in`)
        .join('\n')

      // Execute the expression with input bindings
      if (inputBindings) {
        program += `${inputBindings}\n${expression}`
      } else {
        program += expression
      }

      const result = await executeLayerCode(program)

      if (result.success) {
        return result.result || ''
      } else {
        return `Error: ${result.error}`
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error'
      return `Error: ${errorMessage}`
    }
  }, [layers])

  const handleSave = useCallback(() => {
    // Build Layer source code format
    const enabledLayers = layers.filter(l => l.enabled).reverse() // Reverse for bottom-to-top
    const layerNames = enabledLayers.map(l => l.name)

    let layerSource = `# Layer program\n`
    layerSource += `# Generated by Layer Visual Editor\n\n`
    layerSource += `@layers [${layerNames.join(', ')}]\n\n`

    enabledLayers.forEach(layer => {
      layerSource += `@layer ${layer.name} {\n`
      // Indent the layer code
      const indentedCode = layer.code.split('\n').map(line => line ? `    ${line}` : '').join('\n')
      layerSource += indentedCode
      layerSource += `\n}\n\n`
    })

    layerSource += `# Main expression\n`
    layerSource += mainCode

    const blob = new Blob([layerSource], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'program.layer'
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)

    setOutput('Project saved to program.layer')
  }, [layers, mainCode])

  const handleLoad = useCallback(() => {
    fileInputRef.current?.click()
  }, [])

  const handleFileChange = useCallback(async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    const reader = new FileReader()
    reader.onload = async (e) => {
      try {
        const content = e.target?.result as string

        // Send to backend for parsing
        const parseResult = await parseLayerFile(content)

        if (!parseResult.success || !parseResult.enabledLayers || !parseResult.layers) {
          setOutput(`Error: ${parseResult.error || 'Failed to parse file'}`)
          return
        }

        const { enabledLayers: enabledLayerNames, layers: parsedLayers, mainExpression } = parseResult

        // Build layer array: start with all standard layers
        const loadedLayers: Layer[] = []
        const standardLayerNames = new Set(['core', 'debug', 'ui'])
        const defaultColors = ['#f44336', '#2196f3', '#4caf50', '#ff9800', '#9c27b0', '#00bcd4']

        // Add all standard layers (from INITIAL_LAYERS)
        INITIAL_LAYERS.forEach(initialLayer => {
          const isEnabled = enabledLayerNames.includes(initialLayer.name)
          const hasCode = parsedLayers[initialLayer.name]

          loadedLayers.push({
            ...initialLayer,
            enabled: isEnabled,
            code: hasCode ? parsedLayers[initialLayer.name] : `# ${initialLayer.name} layer\n`,
          })
        })

        // Add custom layers (not in standard set)
        enabledLayerNames.forEach((name, index) => {
          if (!standardLayerNames.has(name) && parsedLayers[name]) {
            loadedLayers.push({
              id: `custom_${name}`,
              name,
              enabled: true,
              color: defaultColors[index % defaultColors.length],
              locked: false,
              code: parsedLayers[name],
            })
          }
        })

        // INITIAL_LAYERS is already in correct UI order (debug, ui, core top-to-bottom)
        setLayers(loadedLayers)
        setMainCode(mainExpression || '')
        setSelectedLayerId('core')

        const enabledCount = loadedLayers.filter(l => l.enabled).length
        setOutput(`Loaded ${file.name}: ${enabledCount}/${loadedLayers.length} layers enabled`)
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown error'
        setOutput(`Error loading file: ${errorMessage}`)
      }
    }
    reader.readAsText(file)

    // Reset input so the same file can be loaded again
    event.target.value = ''
  }, [])

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
          <Tooltip title="Open Project">
            <IconButton color="inherit" onClick={handleLoad}>
              <FolderOpenIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Save Project">
            <IconButton color="inherit" onClick={handleSave}>
              <SaveIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Toggle Flow Diagram">
            <IconButton
              color="inherit"
              onClick={() => setShowFlowDiagram(!showFlowDiagram)}
              sx={{
                bgcolor: showFlowDiagram ? 'rgba(255,255,255,0.1)' : 'transparent',
              }}
            >
              <AccountTreeIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Help">
            <IconButton color="inherit">
              <HelpOutlineIcon />
            </IconButton>
          </Tooltip>
        </Toolbar>
      </AppBar>

      {/* Hidden file input for loading .layer files */}
      <input
        type="file"
        ref={fileInputRef}
        accept=".layer"
        style={{ display: 'none' }}
        onChange={handleFileChange}
      />

      {/* Main Content */}
      <Box sx={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        {/* Layer Stack - Left Panel */}
        <Box sx={{ width: 220, borderRight: 1, borderColor: 'divider', overflow: 'auto' }}>
          <LayerStack
            layers={layers}
            selectedLayerId={selectedLayerId}
            onLayerSelect={setSelectedLayerId}
            onLayerReorder={handleLayerReorder}
            onLayerToggle={handleLayerToggle}
            onLockToggle={handleLockToggle}
            onAddLayer={handleAddLayer}
          />
        </Box>

        {/* Code Editor and Output - Center Panel with Resizable Splitter */}
        <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
          <PanelGroup direction="vertical">
            {/* Code Editor Panel */}
            <Panel defaultSize={60} minSize={20}>
              <Box sx={{ height: '100%', overflow: 'hidden' }}>
                <CodeEditor
                  code={selectedLayer ? selectedLayer.code : mainCode}
                  layerName={selectedLayer ? selectedLayer.name : 'main'}
                  onChange={handleCodeChange}
                />
              </Box>
            </Panel>

            {/* Resize Handle */}
            <PanelResizeHandle>
              <Box
                sx={{
                  height: '4px',
                  backgroundColor: 'divider',
                  cursor: 'row-resize',
                  '&:hover': {
                    backgroundColor: 'primary.main',
                  },
                  transition: 'background-color 0.2s'
                }}
              />
            </PanelResizeHandle>

            {/* Output Console Panel */}
            <Panel defaultSize={40} minSize={15}>
              <Box sx={{ height: '100%', borderTop: 1, borderColor: 'divider', display: 'flex', flexDirection: 'column' }}>
                <OutputConsole
                  output={output}
                  onMaximize={() => setConsoleMaximized(true)}
                  onExecute={handleInteractiveExecute}
                />
              </Box>
            </Panel>
          </PanelGroup>
        </Box>

        {/* Flow Diagram - Right Panel (Collapsible) */}
        {showFlowDiagram && (
          <Box sx={{ width: 280, borderLeft: 1, borderColor: 'divider', overflow: 'auto' }}>
            <FlowDiagram layers={layers} />
          </Box>
        )}
      </Box>

      {/* Maximized Console Dialog */}
      <Dialog
        open={consoleMaximized}
        onClose={() => setConsoleMaximized(false)}
        maxWidth="lg"
        fullWidth
        PaperProps={{
          sx: { height: '90vh' }
        }}
      >
        <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          Output Console
          <IconButton onClick={() => setConsoleMaximized(false)} size="small">
            <CloseFullscreenIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent sx={{ p: 0 }}>
          <OutputConsole output={output} onExecute={handleInteractiveExecute} />
        </DialogContent>
      </Dialog>
    </Box>
  )
}

export default App
