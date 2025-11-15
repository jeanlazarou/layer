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

interface LayerCardProps {
  layer: Layer
  isSelected: boolean
  onSelect: () => void
  onToggle: (layerId: string) => void
  onLockToggle: (layerId: string) => void
  isDragging?: boolean
  dragHandleProps?: any
}

const LayerCard: React.FC<LayerCardProps> = ({
  layer,
  isSelected,
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
      </CardContent>
    </Card>
  )
}

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
    // Dropped outside the list
    if (!result.destination) {
      return
    }

    // No movement
    if (result.destination.index === result.source.index) {
      return
    }

    // Reorder the layers array
    const reorderedLayers = Array.from(layers)
    const [removed] = reorderedLayers.splice(result.source.index, 1)
    reorderedLayers.splice(result.destination.index, 0, removed)

    onLayerReorder(reorderedLayers)
  }

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
