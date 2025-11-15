import React, { useState } from 'react'
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Box,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
} from '@mui/material'
import type { Layer } from '../types'

interface AddLayerDialogProps {
  open: boolean
  onClose: () => void
  onAdd: (layer: Omit<Layer, 'id'>) => void
  existingLayerNames: string[]
}

const LAYER_COLORS = [
  { name: 'Green', value: '#4caf50' },
  { name: 'Blue', value: '#2196f3' },
  { name: 'Orange', value: '#ff9800' },
  { name: 'Purple', value: '#9c27b0' },
  { name: 'Red', value: '#f44336' },
  { name: 'Teal', value: '#009688' },
  { name: 'Pink', value: '#e91e63' },
  { name: 'Indigo', value: '#3f51b5' },
]

const DEFAULT_LAYER_CODE = `# Custom layer
fn example() =
  42
`

const AddLayerDialog: React.FC<AddLayerDialogProps> = ({
  open,
  onClose,
  onAdd,
  existingLayerNames,
}) => {
  const [layerName, setLayerName] = useState('')
  const [color, setColor] = useState(LAYER_COLORS[3].value) // Purple default
  const [error, setError] = useState('')

  const handleAdd = () => {
    const trimmedName = layerName.trim().toLowerCase()

    // Validation
    if (!trimmedName) {
      setError('Layer name is required')
      return
    }

    if (!/^[a-z][a-z0-9_]*$/.test(trimmedName)) {
      setError('Layer name must start with a letter and contain only letters, numbers, and underscores')
      return
    }

    if (existingLayerNames.includes(trimmedName)) {
      setError('A layer with this name already exists')
      return
    }

    if (['core', 'main'].includes(trimmedName)) {
      setError('This layer name is reserved')
      return
    }

    // Create the new layer
    const newLayer: Omit<Layer, 'id'> = {
      name: trimmedName,
      enabled: true,
      color,
      locked: false,
      code: DEFAULT_LAYER_CODE,
    }

    onAdd(newLayer)
    handleClose()
  }

  const handleClose = () => {
    setLayerName('')
    setColor(LAYER_COLORS[3].value)
    setError('')
    onClose()
  }

  return (
    <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
      <DialogTitle>Add Custom Layer</DialogTitle>
      <DialogContent>
        <Box sx={{ pt: 2, display: 'flex', flexDirection: 'column', gap: 3 }}>
          {error && (
            <Alert severity="error" onClose={() => setError('')}>
              {error}
            </Alert>
          )}

          <TextField
            autoFocus
            label="Layer Name"
            placeholder="mylayer"
            value={layerName}
            onChange={(e) => setLayerName(e.target.value)}
            fullWidth
            helperText="Use lowercase letters, numbers, and underscores"
          />

          <FormControl fullWidth>
            <InputLabel>Color</InputLabel>
            <Select
              value={color}
              label="Color"
              onChange={(e) => setColor(e.target.value)}
            >
              {LAYER_COLORS.map((c) => (
                <MenuItem key={c.value} value={c.value}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <Box
                      sx={{
                        width: 20,
                        height: 20,
                        borderRadius: '50%',
                        bgcolor: c.value,
                        border: 1,
                        borderColor: 'divider',
                      }}
                    />
                    {c.name}
                  </Box>
                </MenuItem>
              ))}
            </Select>
          </FormControl>

          <Alert severity="info">
            The layer will be created with default template code. You can edit it after creation.
          </Alert>
        </Box>
      </DialogContent>
      <DialogActions>
        <Button onClick={handleClose}>Cancel</Button>
        <Button onClick={handleAdd} variant="contained">
          Add Layer
        </Button>
      </DialogActions>
    </Dialog>
  )
}

export default AddLayerDialog
