import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  Box,
  AppBar,
  Toolbar,
  Typography,
  IconButton,
  Tooltip,
  Dialog,
  DialogContent,
  DialogTitle,
  Button,
  Menu,
  MenuItem,
  ListItemText,
  Select,
  Divider,
} from '@mui/material'
import PlayArrowIcon from '@mui/icons-material/PlayArrow'
import SaveIcon from '@mui/icons-material/Save'
import FolderOpenIcon from '@mui/icons-material/FolderOpen'
import NoteAddIcon from '@mui/icons-material/NoteAdd'
import AccountTreeIcon from '@mui/icons-material/AccountTree'
import HelpOutlineIcon from '@mui/icons-material/HelpOutline'
import MenuBookIcon from '@mui/icons-material/MenuBook'
import CloseFullscreenIcon from '@mui/icons-material/CloseFullscreen'
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels'

import LayerStack from './components/LayerStack'
import CodeEditor from './components/CodeEditor'
import OutputConsole, { IDLE_OUTPUT, type OutputState } from './components/OutputConsole'
import ResolutionPanel from './components/ResolutionPanel'
import HelpDialog from './components/HelpDialog'
import { describeProgram, runProgram, splitLayerFile } from './wasm/layerWasm'
import { buildProgram, describeLocation, locateIssue, normalizeOrder, projectFromSplit, serializeProject, type Program } from './program'
import { loadWorkspace, saveWorkspace } from './storage'
import { EXAMPLES } from './examples'
import { DEFAULT_LAYERS, DEFAULT_MAIN } from './defaults'
import { BLEND_MODES, DEFAULT_BLEND, isLayerPure, type BlendMode, type BlendSettings, type Description, type Layer, type LayerId, type LocatedIssue } from './types'

const saved = loadWorkspace()

/** ", blend: chain, 2 per-function modes" for status messages. */
function describeBlend(blend: BlendSettings): string {
  const n = Object.keys(blend.perFunction).length
  const parts = [
    ...(blend.mode !== 'override' ? [`blend: ${blend.mode}`] : []),
    ...(n ? [`${n} per-function blend mode${n === 1 ? '' : 's'}`] : []),
  ]
  return parts.length ? `, ${parts.join(', ')}` : ''
}

function App() {
  const [layers, setLayers] = useState<Layer[]>(() => normalizeOrder(saved?.layers ?? DEFAULT_LAYERS))
  const [mainCode, setMainCode] = useState(saved?.mainCode ?? DEFAULT_MAIN)
  const [blend, setBlend] = useState<BlendSettings>(saved?.blend ?? DEFAULT_BLEND)
  const [projectName, setProjectName] = useState(saved?.projectName ?? '')
  const [selectedLayerId, setSelectedLayerId] = useState<LayerId>('core')

  const [output, setOutput] = useState<OutputState>(IDLE_OUTPUT)
  const [description, setDescription] = useState<Description | null>(null)
  const [liveIssues, setLiveIssues] = useState<LocatedIssue[]>([])
  const [runIssues, setRunIssues] = useState<{ source: string; issues: LocatedIssue[] }>({ source: '', issues: [] })

  const [showResolution, setShowResolution] = useState(() => window.innerWidth >= 1200)
  const [helpOpen, setHelpOpen] = useState(false)
  const [consoleMaximized, setConsoleMaximized] = useState(false)
  const [examplesAnchor, setExamplesAnchor] = useState<HTMLElement | null>(null)
  const [reveal, setReveal] = useState<{ line: number; column: number; nonce: number }>()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const hasRun = useRef(false)

  const selectedLayer = layers.find(l => l.id === selectedLayerId)

  // The program as it would run right now (enabled layers only).
  const program: Program = useMemo(() => buildProgram(layers, mainCode, blend), [layers, mainCode, blend])

  // ── Persistence ────────────────────────────────────────────────────────────
  useEffect(() => {
    const t = setTimeout(() => saveWorkspace({ layers, mainCode, blend, projectName }), 300)
    return () => clearTimeout(t)
  }, [layers, mainCode, blend, projectName])

  useEffect(() => {
    document.title = projectName ? `${projectName} — Layer` : 'Layer Visual Editor'
  }, [projectName])

  // ── Live diagnostics + resolution view (debounced while typing) ───────────
  useEffect(() => {
    let cancelled = false
    const t = setTimeout(async () => {
      const res = await describeProgram(program.source)
      if (cancelled) return
      setLiveIssues(res.errors.map(e => locateIssue(e, program)))
      setDescription(res.success && res.description ? res.description : null)
    }, 350)
    return () => {
      cancelled = true
      clearTimeout(t)
    }
  }, [program])

  // Runtime errors only stay on screen while the program is unchanged.
  const issues = useMemo(() => {
    const runtime = runIssues.source === program.source ? runIssues.issues.filter(i => i.kind === 'runtime') : []
    return [...liveIssues, ...runtime]
  }, [liveIssues, runIssues, program.source])

  const errorCounts = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const issue of issues) if (issue.target) counts[issue.target] = (counts[issue.target] ?? 0) + 1
    return counts
  }, [issues])

  // ── Running ────────────────────────────────────────────────────────────────
  const handleRun = useCallback(async () => {
    hasRun.current = true
    setOutput(prev => ({ ...prev, status: 'running', notice: undefined }))
    const res = await runProgram(program.source)
    const located = res.errors.map(e => locateIssue(e, program))
    setRunIssues({ source: program.source, issues: located })
    setOutput({
      status: 'done',
      success: res.success,
      result: res.result,
      kind: res.kind,
      logs: res.logs,
      errors: located,
    })
  }, [program])

  // Re-run automatically when the stack's structure changes (toggle,
  // reorder, purity, blend mode) — the core interaction of the editor.
  const runRef = useRef(handleRun)
  runRef.current = handleRun
  const structureKey = JSON.stringify([layers.map(l => [l.id, l.enabled, l.impure ?? false]), blend])
  useEffect(() => {
    if (hasRun.current) void runRef.current()
  }, [structureKey])

  // UI button handlers: run an expression on top of the current stack with
  // the input values bound as variables.
  const handleInteractiveExecute = useCallback(
    async (expression: string, inputs: Record<string, string | number>): Promise<string> => {
      const res = await runProgram(buildProgram(layers, mainCode, blend, { main: expression }).source, inputs)
      if (res.logs.length > 0) {
        setOutput(prev => ({ ...prev, logs: [...prev.logs, ...res.logs] }))
      }
      if (!res.success) return `Error: ${res.errors.map(e => e.message).join('; ')}`
      return res.result
    },
    [layers, mainCode, blend],
  )

  // ── Editing ────────────────────────────────────────────────────────────────
  const handleCodeChange = useCallback(
    (code: string) => {
      if (selectedLayerId === 'main') setMainCode(code)
      else setLayers(prev => prev.map(l => (l.id === selectedLayerId ? { ...l, code } : l)))
    },
    [selectedLayerId],
  )

  const updateLayer = (id: string, change: (l: Layer) => Partial<Layer>) =>
    setLayers(prev => prev.map(l => (l.id === id ? { ...l, ...change(l) } : l)))

  const handleIssueClick = (issue: LocatedIssue) => {
    if (!issue.target) return
    setSelectedLayerId(issue.target)
    // Let the editor switch buffers before revealing the line.
    setTimeout(() => setReveal({ line: issue.targetLine, column: issue.targetColumn, nonce: Date.now() }), 0)
    setConsoleMaximized(false)
  }

  // ── Projects ───────────────────────────────────────────────────────────────
  const loadSource = async (source: string, name: string) => {
    const split = await splitLayerFile(source)
    if (!split.success) {
      setOutput({
        ...IDLE_OUTPUT,
        status: 'done',
        success: false,
        notice: `Could not open ${name}:`,
        errors: split.errors.map(err => ({ ...err, target: null, targetLine: err.line, targetColumn: err.column })),
      })
      return
    }
    const project = projectFromSplit(split)
    setDescription(null) // refreshed (debounced) from the new program
    setLiveIssues([])
    setLayers(project.layers)
    setMainCode(project.mainCode)
    setBlend(project.blend)
    setProjectName(name)
    setSelectedLayerId(project.layers.find(l => l.enabled && l.name !== 'core')?.id ?? 'core')
    setRunIssues({ source: '', issues: [] })
    hasRun.current = false
    const active = project.layers.filter(l => l.enabled).length
    setOutput({ ...IDLE_OUTPUT, notice: `Opened ${name}: ${active} active layer${active === 1 ? '' : 's'}${describeBlend(project.blend)}. Press ▶ to run.` })
  }

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    event.target.value = '' // allow re-opening the same file
    if (!file) return
    await loadSource(await file.text(), file.name.replace(/\.layer$/, ''))
  }

  const handleSave = () => {
    const blob = new Blob([serializeProject(layers, mainCode, blend)], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${projectName || 'program'}.layer`
    a.click()
    URL.revokeObjectURL(url)
    setOutput(prev => ({ ...prev, notice: `Saved ${a.download} (hidden layers are kept, but not listed in @layers)` }))
  }

  const handleNew = () => {
    if (!window.confirm('Start a new project? The current one is only kept if you saved it.')) return
    setLayers(DEFAULT_LAYERS)
    setMainCode(DEFAULT_MAIN)
    setBlend(DEFAULT_BLEND)
    setProjectName('')
    setSelectedLayerId('core')
    setRunIssues({ source: '', issues: [] })
    hasRun.current = false
    setOutput(IDLE_OUTPUT)
  }

  // ── Layout ─────────────────────────────────────────────────────────────────
  const editorTitle = selectedLayer ? selectedLayer.name : 'main'
  const editorBadge = selectedLayer
    ? `${isLayerPure(selectedLayer) ? 'pure' : 'impure'} layer${selectedLayer.enabled ? '' : ' • hidden (not in @layers)'}`
    : 'main expression — runs on top of the stack'
  const bufferId = selectedLayer ? selectedLayer.id : 'main'
  const bufferIssues = issues.filter(i => i.target === bufferId)

  const outputConsole = (maximize?: () => void) => (
    <OutputConsole
      output={output}
      describeLocation={issue => describeLocation(issue, layers)}
      onIssueClick={handleIssueClick}
      onMaximize={maximize}
      onExecute={handleInteractiveExecute}
    />
  )

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      <AppBar position="static" elevation={1}>
        <Toolbar variant="dense" sx={{ gap: 0.5 }}>
          <Typography variant="h6" sx={{ fontWeight: 700, mr: 1 }}>
            Layer
          </Typography>
          <Typography variant="body2" sx={{ opacity: 0.8, flexGrow: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {projectName || 'untitled'}
          </Typography>

          <Tooltip title="Run (⌘/Ctrl + Enter)">
            <span>
              <Button
                variant="contained"
                color="success"
                size="small"
                startIcon={<PlayArrowIcon />}
                onClick={handleRun}
                disabled={output.status === 'running'}
                sx={{ mr: 1 }}
              >
                Run
              </Button>
            </span>
          </Tooltip>

          <Tooltip title="Blend mode: how functions defined in several layers combine">
            <Select
              size="small"
              value={blend.mode}
              onChange={e => setBlend(prev => ({ ...prev, mode: e.target.value as BlendMode }))}
              renderValue={v => {
                const n = Object.keys(blend.perFunction).length
                return `blend: ${v}${n ? ` · ${n} per-function` : ''}`
              }}
              sx={{ color: 'inherit', fontSize: '0.8rem', height: 30, mr: 1, '.MuiOutlinedInput-notchedOutline': { borderColor: 'rgba(255,255,255,0.4)' }, '.MuiSvgIcon-root': { color: 'inherit' } }}
            >
              {BLEND_MODES.map(m => (
                <MenuItem key={m.value} value={m.value}>
                  <ListItemText primary={m.label} secondary={m.help} />
                </MenuItem>
              ))}
            </Select>
          </Tooltip>

          <Tooltip title="Examples">
            <IconButton color="inherit" onClick={e => setExamplesAnchor(e.currentTarget)}>
              <MenuBookIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="New project">
            <IconButton color="inherit" onClick={handleNew}>
              <NoteAddIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Open .layer file">
            <IconButton color="inherit" onClick={() => fileInputRef.current?.click()}>
              <FolderOpenIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Save as .layer file">
            <IconButton color="inherit" onClick={handleSave}>
              <SaveIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Who answers each call?">
            <IconButton
              color="inherit"
              onClick={() => setShowResolution(v => !v)}
              sx={{ bgcolor: showResolution ? 'rgba(255,255,255,0.15)' : 'transparent' }}
            >
              <AccountTreeIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Help">
            <IconButton color="inherit" onClick={() => setHelpOpen(true)}>
              <HelpOutlineIcon />
            </IconButton>
          </Tooltip>
        </Toolbar>
      </AppBar>

      <Menu anchorEl={examplesAnchor} open={!!examplesAnchor} onClose={() => setExamplesAnchor(null)} PaperProps={{ sx: { maxHeight: '70vh', width: 380 } }}>
        <Typography variant="caption" color="text.secondary" sx={{ px: 2, py: 0.5, display: 'block' }}>
          Opening an example replaces the current project.
        </Typography>
        <Divider />
        {EXAMPLES.map(ex => (
          <MenuItem
            key={ex.id}
            onClick={() => {
              setExamplesAnchor(null)
              void loadSource(ex.source, ex.id)
            }}
          >
            <ListItemText primary={ex.title} secondary={ex.description} secondaryTypographyProps={{ noWrap: true }} />
          </MenuItem>
        ))}
      </Menu>

      <input type="file" ref={fileInputRef} accept=".layer" style={{ display: 'none' }} onChange={handleFileChange} />

      <Box sx={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
        <Box sx={{ width: 240, flexShrink: 0, borderRight: 1, borderColor: 'divider', overflow: 'auto', bgcolor: 'background.paper' }}>
          <LayerStack
            layers={layers}
            selectedLayerId={selectedLayerId}
            errorCounts={errorCounts}
            onLayerSelect={setSelectedLayerId}
            onLayerReorder={next => setLayers(normalizeOrder(next))}
            onLayerToggle={id => updateLayer(id, l => ({ enabled: !l.enabled }))}
            onLockToggle={id => id !== 'core' && updateLayer(id, l => ({ locked: !l.locked }))}
            onPurityToggle={id => updateLayer(id, l => ({ impure: l.impure ? undefined : true }))}
            onDeleteLayer={id => {
              const layer = layers.find(l => l.id === id)
              if (!layer || !window.confirm(`Delete layer "${layer.name}" and its code?`)) return
              setLayers(prev => prev.filter(l => l.id !== id))
              if (selectedLayerId === id) setSelectedLayerId('core')
            }}
            onAddLayer={data => {
              const layer: Layer = { ...data, id: `custom_${data.name}_${Date.now()}` }
              setLayers(prev => [layer, ...prev])
              setSelectedLayerId(layer.id)
            }}
          />
        </Box>

        <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
          <PanelGroup direction="vertical">
            <Panel defaultSize={60} minSize={20}>
              <CodeEditor
                bufferId={bufferId}
                code={selectedLayer ? selectedLayer.code : mainCode}
                title={editorTitle}
                color={selectedLayer ? selectedLayer.color : '#455a64'}
                badge={editorBadge}
                issues={bufferIssues}
                onChange={handleCodeChange}
                onRun={handleRun}
                reveal={reveal}
              />
            </Panel>
            <PanelResizeHandle>
              <Box sx={{ height: '4px', bgcolor: 'divider', cursor: 'row-resize', '&:hover': { bgcolor: 'primary.main' } }} />
            </PanelResizeHandle>
            <Panel defaultSize={40} minSize={15}>{outputConsole(() => setConsoleMaximized(true))}</Panel>
          </PanelGroup>
        </Box>

        {showResolution && (
          <Box sx={{ width: 300, flexShrink: 0, borderLeft: 1, borderColor: 'divider', bgcolor: 'background.paper' }}>
            <ResolutionPanel
              description={description}
              layers={layers}
              blend={blend}
              onFunctionModeChange={(name, mode) =>
                setBlend(prev => {
                  const perFunction = { ...prev.perFunction }
                  if (mode) perFunction[name] = mode
                  else delete perFunction[name]
                  return { ...prev, perFunction }
                })
              }
            />
          </Box>
        )}
      </Box>

      <Dialog open={consoleMaximized} onClose={() => setConsoleMaximized(false)} maxWidth="lg" fullWidth PaperProps={{ sx: { height: '90vh' } }}>
        <DialogTitle sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', py: 1 }}>
          Output
          <IconButton onClick={() => setConsoleMaximized(false)} size="small">
            <CloseFullscreenIcon />
          </IconButton>
        </DialogTitle>
        <DialogContent sx={{ p: 0, display: 'flex' }}>
          <Box sx={{ flex: 1 }}>{outputConsole()}</Box>
        </DialogContent>
      </Dialog>

      <HelpDialog open={helpOpen} onClose={() => setHelpOpen(false)} />
    </Box>
  )
}

export default App
