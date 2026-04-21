import React, { useState } from 'react'
import {
  Box,
  Card,
  CardContent,
  Typography,
  IconButton,
  Divider,
  Button,
} from '@mui/material'
import VisibilityIcon from '@mui/icons-material/Visibility'
import VisibilityOffIcon from '@mui/icons-material/VisibilityOff'
import LockIcon from '@mui/icons-material/Lock'
import LockOpenIcon from '@mui/icons-material/LockOpen'
import DragIndicatorIcon from '@mui/icons-material/DragIndicator'
import AddIcon from '@mui/icons-material/Add'
import { DragDropContext, Droppable, Draggable, DropResult } from 'react-beautiful-dnd'
import AddLayerDialog from './AddLayerDialog'
import type { Layer, LayerId } from '../types'

// ── Diff utilities ─────────────────────────────────────────────────────────────

function extractFunctionNames(code: string): Set<string> {
  const names = new Set<string>()
  const re = /\bfn\s+(\w+)\s*\(/g
  let m: RegExpExecArray | null
  while ((m = re.exec(code)) !== null) {
    names.add(m[1])
  }
  return names
}

interface DiffResult {
  added: string[]      // present in this layer, absent in layer below
  wrapped: string[]    // present in both; this layer calls through to the lower one
  overridden: string[] // present in both; this layer does not call through
}

function computeLayerDiff(selected: Layer, layerBelow: Layer): DiffResult {
  const selectedFns = extractFunctionNames(selected.code)
  const belowFns    = extractFunctionNames(layerBelow.code)

  const added: string[]      = []
  const wrapped: string[]    = []
  const overridden: string[] = []

  selectedFns.forEach(fn => {
    if (!belowFns.has(fn)) {
      added.push(fn)
    } else {
      // Wrap detection: the body references <lowerLayer>.<fn>
      const callsThrough = new RegExp(`\\b${layerBelow.name}\\.${fn}\\b`).test(selected.code)
      if (callsThrough) {
        wrapped.push(fn)
      } else {
        overridden.push(fn)
      }
    }
  })

  return { added, wrapped, overridden }
}

// ── DiffPills ──────────────────────────────────────────────────────────────────

const PILL_STYLES = {
  added:      { bg: '#2e7d32', fg: '#fff' }, // green
  wrapped:    { bg: '#0277bd', fg: '#fff' }, // blue
  overridden: { bg: '#e65100', fg: '#fff' }, // amber-orange
} as const

type PillKind = keyof typeof PILL_STYLES

function DiffPill({ name, kind }: { name: string; kind: PillKind }) {
  const { bg, fg } = PILL_STYLES[kind]
  return (
    <Box
      component="span"
      sx={{
        fontSize: '0.58rem',
        px: 0.6,
        py: 0.15,
        bgcolor: bg,
        color: fg,
        borderRadius: 0.5,
        fontFamily: 'monospace',
        lineHeight: 1.4,
      }}
    >
      {name}
    </Box>
  )
}

function DiffSection({ diff, layerBelowName }: { diff: DiffResult; layerBelowName: string }) {
  const hasAny = diff.added.length > 0 || diff.wrapped.length > 0 || diff.overridden.length > 0
  if (!hasAny) return null

  return (
    <Box sx={{ mt: 0.75, pt: 0.5, borderTop: '1px solid', borderColor: 'divider' }}>
      <Typography variant="caption" sx={{ fontSize: '0.6rem', color: 'text.disabled', display: 'block', mb: 0.4 }}>
        vs {layerBelowName}
      </Typography>

      {diff.added.length > 0 && (
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.4, mb: 0.3 }}>
          {diff.added.map(fn => <DiffPill key={fn} name={fn} kind="added" />)}
        </Box>
      )}
      {diff.wrapped.length > 0 && (
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.4, mb: 0.3 }}>
          {diff.wrapped.map(fn => <DiffPill key={fn} name={fn} kind="wrapped" />)}
        </Box>
      )}
      {diff.overridden.length > 0 && (
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.4 }}>
          {diff.overridden.map(fn => <DiffPill key={fn} name={fn} kind="overridden" />)}
        </Box>
      )}
    </Box>
  )
}

// ── LayerCard ──────────────────────────────────────────────────────────────────

interface LayerCardProps {
  layer: Layer
  isSelected: boolean
  diff?: DiffResult
  layerBelowName?: string
  onSelect: () => void
  onToggle: (layerId: string) => void
  onLockToggle: (layerId: string) => void
  isDragging?: boolean
  dragHandleProps?: any
}

const LayerCard: React.FC<LayerCardProps> = ({
  layer,
  isSelected,
  diff,
  layerBelowName,
  onSelect,
  onToggle,
  onLockToggle,
  isDragging = false,
  dragHandleProps
}) => {
  return (
    <Card
      elevation={isDragging ? 12 : isSelected ? 4 : 1}
      sx={{
        mb: 0.75,
        cursor: 'pointer',
        opacity: layer.enabled ? 1 : 0.5,
        border: isSelected ? 1.5 : 0,
        borderColor: 'primary.main',
        transition: 'all 0.2s',
        bgcolor: isDragging ? 'action.hover' : 'background.paper',
        transform: isDragging ? 'rotate(2deg)' : 'none',
        '&:hover': {
          elevation: 3,
          transform: isDragging ? 'rotate(2deg)' : 'translateY(-1px)',
        },
      }}
      onClick={onSelect}
    >
      <CardContent sx={{ p: 1.25, '&:last-child': { pb: 1.25 } }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 0.5 }}>
          <div
            {...dragHandleProps}
            onClick={(e) => e.stopPropagation()}
            style={{
              display: 'flex',
              alignItems: 'center',
              cursor: layer.locked ? 'not-allowed' : (isDragging ? 'grabbing' : 'grab'),
              marginRight: '6px',
            }}
          >
            <DragIndicatorIcon
              sx={{
                color: layer.locked ? 'action.disabled' : 'text.disabled',
                fontSize: 18,
                pointerEvents: 'none',
              }}
            />
          </div>
          <Box
            sx={{
              width: 8,
              height: 8,
              borderRadius: '50%',
              bgcolor: layer.color,
              mr: 0.75,
            }}
          />
          <Typography variant="body2" sx={{ flexGrow: 1, fontWeight: 600, fontSize: '0.875rem' }}>
            {layer.name}
          </Typography>

          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation()
              if (layer.id !== 'core') onLockToggle(layer.id)
            }}
            disabled={layer.id === 'core'}
            title={layer.id === 'core' ? 'Core layer cannot be unlocked' : layer.locked ? 'Unlock layer' : 'Lock layer'}
            sx={{ p: 0.5 }}
          >
            {layer.locked ? (
              <LockIcon sx={{ fontSize: 16 }} />
            ) : (
              <LockOpenIcon sx={{ fontSize: 16 }} />
            )}
          </IconButton>

          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation()
              if (!layer.locked) onToggle(layer.id)
            }}
            disabled={layer.locked}
            title={layer.locked ? 'Locked layers cannot be toggled' : layer.enabled ? 'Disable layer' : 'Enable layer'}
            sx={{ p: 0.5 }}
          >
            {layer.enabled ? (
              <VisibilityIcon sx={{ fontSize: 16 }} />
            ) : (
              <VisibilityOffIcon sx={{ fontSize: 16 }} />
            )}
          </IconButton>
        </Box>

        <Typography variant="caption" color="text.secondary" sx={{ display: 'block', fontSize: '0.7rem' }}>
          {layer.enabled ? 'Active' : 'Disabled'}
          {layer.locked && ' • Locked'}
        </Typography>

        {isSelected && diff && layerBelowName && (
          <DiffSection diff={diff} layerBelowName={layerBelowName} />
        )}
      </CardContent>
    </Card>
  )
}

// ── DiffLegend ─────────────────────────────────────────────────────────────────

function DiffLegend() {
  const entries: { kind: PillKind; label: string }[] = [
    { kind: 'added',      label: 'adds' },
    { kind: 'wrapped',    label: 'wraps' },
    { kind: 'overridden', label: 'overrides' },
  ]
  return (
    <Box sx={{ display: 'flex', gap: 0.75, alignItems: 'center', mt: 0.75 }}>
      {entries.map(({ kind, label }) => (
        <Box key={kind} sx={{ display: 'flex', alignItems: 'center', gap: 0.35 }}>
          <Box sx={{ width: 7, height: 7, borderRadius: '50%', bgcolor: PILL_STYLES[kind].bg }} />
          <Typography variant="caption" sx={{ fontSize: '0.6rem', color: 'text.disabled' }}>
            {label}
          </Typography>
        </Box>
      ))}
    </Box>
  )
}

// ── LayerStack ─────────────────────────────────────────────────────────────────

interface LayerStackProps {
  layers: Layer[]
  selectedLayerId: LayerId
  onLayerSelect: (layerId: LayerId) => void
  onLayerToggle: (layerId: string) => void
  onLockToggle: (layerId: string) => void
  onLayerReorder: (layers: Layer[]) => void
  onAddLayer: (layer: Omit<Layer, 'id'>) => void
}

const LayerStack: React.FC<LayerStackProps> = ({
  layers,
  selectedLayerId,
  onLayerSelect,
  onLayerToggle,
  onLockToggle,
  onLayerReorder,
  onAddLayer,
}) => {
  const [addDialogOpen, setAddDialogOpen] = useState(false)

  const handleDragEnd = (result: DropResult) => {
    if (!result.destination) return
    if (result.destination.index === result.source.index) return

    const reorderedLayers = Array.from(layers)
    const [removed] = reorderedLayers.splice(result.source.index, 1)
    reorderedLayers.splice(result.destination.index, 0, removed)
    onLayerReorder(reorderedLayers)
  }

  // Compute diff for the selected layer against the nearest enabled layer below it
  const selectedIndex = layers.findIndex(l => l.id === selectedLayerId)
  const selectedLayer = selectedIndex >= 0 ? layers[selectedIndex] : undefined
  const layerBelow = selectedIndex >= 0
    ? layers.slice(selectedIndex + 1).find(l => l.enabled)
    : undefined
  const selectedDiff = selectedLayer && layerBelow
    ? computeLayerDiff(selectedLayer, layerBelow)
    : undefined

  return (
    <Box sx={{ p: 1.5 }}>
      <Typography variant="subtitle1" sx={{ mb: 1.5, fontWeight: 600, fontSize: '0.95rem' }}>
        Layer Stack
      </Typography>

      <DragDropContext onDragEnd={handleDragEnd}>
        <Droppable droppableId="layer-stack">
          {(provided, snapshot) => (
            <Box
              {...provided.droppableProps}
              ref={provided.innerRef}
              sx={{
                minHeight: 100,
                bgcolor: snapshot.isDraggingOver ? 'action.hover' : 'transparent',
                borderRadius: 1,
                transition: 'background-color 0.2s',
              }}
            >
              {layers.map((layer, index) => (
                <Draggable
                  key={layer.id}
                  draggableId={layer.id}
                  index={index}
                  isDragDisabled={layer.locked}
                >
                  {(provided, snapshot) => (
                    <div
                      ref={provided.innerRef}
                      {...provided.draggableProps}
                    >
                      <LayerCard
                        layer={layer}
                        isSelected={selectedLayerId === layer.id}
                        diff={selectedLayerId === layer.id ? selectedDiff : undefined}
                        layerBelowName={selectedLayerId === layer.id ? layerBelow?.name : undefined}
                        onSelect={() => onLayerSelect(layer.id)}
                        onToggle={onLayerToggle}
                        onLockToggle={onLockToggle}
                        isDragging={snapshot.isDragging}
                        dragHandleProps={provided.dragHandleProps}
                      />
                    </div>
                  )}
                </Draggable>
              ))}
              {provided.placeholder}
            </Box>
          )}
        </Droppable>
      </DragDropContext>

      <DiffLegend />

      <Divider sx={{ my: 1.5 }} />

      <Card
        elevation={selectedLayerId === 'main' ? 4 : 1}
        sx={{
          cursor: 'pointer',
          border: selectedLayerId === 'main' ? 1.5 : 0,
          borderColor: 'primary.main',
        }}
        onClick={() => onLayerSelect('main')}
      >
        <CardContent sx={{ p: 1.25, '&:last-child': { pb: 1.25 } }}>
          <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.875rem' }}>
            Main Expression
          </Typography>
          <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.7rem' }}>
            Entry point
          </Typography>
        </CardContent>
      </Card>

      <Button
        fullWidth
        size="small"
        startIcon={<AddIcon sx={{ fontSize: 16 }} />}
        sx={{ mt: 1.5, fontSize: '0.8rem' }}
        variant="outlined"
        onClick={() => setAddDialogOpen(true)}
      >
        Add Layer
      </Button>

      <AddLayerDialog
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        onAdd={onAddLayer}
        existingLayerNames={layers.map(l => l.name)}
      />
    </Box>
  )
}

export default LayerStack
