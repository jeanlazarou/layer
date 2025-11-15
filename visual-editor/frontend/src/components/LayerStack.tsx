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
}

const LayerCard: React.FC<LayerCardProps> = ({
  layer,
  isSelected,
  onSelect,
  onToggle,
  onLockToggle,
  isDragging = false
}) => {
  return (
    <Card
      elevation={isDragging ? 12 : isSelected ? 8 : 1}
      sx={{
        mb: 1,
        cursor: 'pointer',
        opacity: layer.enabled ? 1 : 0.5,
        border: isSelected ? 2 : 0,
        borderColor: 'primary.main',
        transition: 'all 0.2s',
        bgcolor: isDragging ? 'action.hover' : 'background.paper',
        transform: isDragging ? 'rotate(2deg)' : 'none',
        '&:hover': {
          elevation: 4,
          transform: isDragging ? 'rotate(2deg)' : 'translateY(-2px)',
        },
      }}
      onClick={onSelect}
    >
      <CardContent sx={{ p: 2, '&:last-child': { pb: 2 } }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
          <DragIndicatorIcon
            sx={{
              color: 'text.disabled',
              mr: 1,
              cursor: isDragging ? 'grabbing' : 'grab',
              '&:active': {
                cursor: 'grabbing'
              }
            }}
          />
          <Box
            sx={{
              width: 12,
              height: 12,
              borderRadius: '50%',
              bgcolor: layer.color,
              mr: 1,
            }}
          />
          <Typography variant="h6" sx={{ flexGrow: 1, fontWeight: 600 }}>
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
          >
            {layer.locked ? (
              <LockIcon fontSize="small" />
            ) : (
              <LockOpenIcon fontSize="small" />
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
          >
            {layer.enabled ? (
              <VisibilityIcon fontSize="small" />
            ) : (
              <VisibilityOffIcon fontSize="small" />
            )}
          </IconButton>
        </Box>

        <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
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
    <Box sx={{ p: 2 }}>
      <Typography variant="h6" sx={{ mb: 2, fontWeight: 600 }}>
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
                      {...provided.dragHandleProps}
                    >
                      <LayerCard
                        layer={layer}
                        isSelected={selectedLayerId === layer.id}
                        onSelect={() => onLayerSelect(layer.id)}
                        onToggle={onLayerToggle}
                        onLockToggle={onLockToggle}
                        isDragging={snapshot.isDragging}
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

      <Divider sx={{ my: 2 }} />

      <Card
        elevation={selectedLayerId === 'main' ? 8 : 1}
        sx={{
          cursor: 'pointer',
          border: selectedLayerId === 'main' ? 2 : 0,
          borderColor: 'primary.main',
        }}
        onClick={() => onLayerSelect('main')}
      >
        <CardContent sx={{ p: 2, '&:last-child': { pb: 2 } }}>
          <Typography variant="h6" sx={{ fontWeight: 600 }}>
            Main Expression
          </Typography>
          <Typography variant="caption" color="text.secondary">
            Entry point
          </Typography>
        </CardContent>
      </Card>

      <Button
        fullWidth
        startIcon={<AddIcon />}
        sx={{ mt: 2 }}
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
