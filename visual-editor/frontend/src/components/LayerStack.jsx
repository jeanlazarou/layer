import React from 'react'
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
import DragIndicatorIcon from '@mui/icons-material/DragIndicator'
import AddIcon from '@mui/icons-material/Add'

const LayerCard = ({ layer, isSelected, onSelect, onToggle }) => {
  return (
    <Card
      elevation={isSelected ? 8 : 1}
      sx={{
        mb: 1,
        cursor: 'pointer',
        opacity: layer.enabled ? 1 : 0.5,
        border: isSelected ? 2 : 0,
        borderColor: 'primary.main',
        transition: 'all 0.2s',
        '&:hover': {
          elevation: 4,
          transform: 'translateY(-2px)',
        },
      }}
      onClick={onSelect}
    >
      <CardContent sx={{ p: 2, '&:last-child': { pb: 2 } }}>
        <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
          <DragIndicatorIcon
            sx={{ color: 'text.disabled', mr: 1, cursor: 'grab' }}
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

          {layer.locked && (
            <LockIcon sx={{ fontSize: 18, color: 'text.disabled', mr: 1 }} />
          )}

          <IconButton
            size="small"
            onClick={(e) => {
              e.stopPropagation()
              if (!layer.locked) onToggle(layer.id)
            }}
            disabled={layer.locked}
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

function LayerStack({ layers, selectedLayerId, onLayerSelect, onLayerToggle }) {
  return (
    <Box sx={{ p: 2 }}>
      <Typography variant="h6" sx={{ mb: 2, fontWeight: 600 }}>
        Layer Stack
      </Typography>

      {layers.map((layer) => (
        <LayerCard
          key={layer.id}
          layer={layer}
          isSelected={selectedLayerId === layer.id}
          onSelect={() => onLayerSelect(layer.id)}
          onToggle={onLayerToggle}
        />
      ))}

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
      >
        Add Layer
      </Button>
    </Box>
  )
}

export default LayerStack
