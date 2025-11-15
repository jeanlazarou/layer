import React, { useMemo } from 'react'
import { Box, Paper, Typography, Chip } from '@mui/material'
import type { Layer } from '../types'

interface FunctionInfo {
  name: string
  layer: string
  layerColor: string
}

interface FlowDiagramProps {
  layers: Layer[]
}

const FlowDiagram: React.FC<FlowDiagramProps> = ({ layers }) => {
  const functions = useMemo(() => {
    const funcs: FunctionInfo[] = []

    // Extract function definitions from each layer
    layers.forEach(layer => {
      if (!layer.enabled) return

      // Match function definitions: fn name(...) = ...
      const fnRegex = /fn\s+(\w+)\s*\(/g
      let match

      while ((match = fnRegex.exec(layer.code)) !== null) {
        funcs.push({
          name: match[1],
          layer: layer.name,
          layerColor: layer.color,
        })
      }
    })

    return funcs
  }, [layers])

  // Group functions by layer
  const functionsByLayer = useMemo(() => {
    const grouped: Record<string, FunctionInfo[]> = {}

    functions.forEach(func => {
      if (!grouped[func.layer]) {
        grouped[func.layer] = []
      }
      grouped[func.layer].push(func)
    })

    return grouped
  }, [functions])

  const enabledLayers = layers.filter(l => l.enabled)

  if (enabledLayers.length === 0) {
    return (
      <Box sx={{ p: 3, textAlign: 'center' }}>
        <Typography variant="body2" color="text.secondary">
          No enabled layers to visualize
        </Typography>
      </Box>
    )
  }

  return (
    <Box sx={{ p: 2, height: '100%', overflow: 'auto' }}>
      <Typography variant="h6" sx={{ mb: 2, fontWeight: 600 }}>
        Function Flow Diagram
      </Typography>

      <Paper variant="outlined" sx={{ p: 2, bgcolor: 'grey.50' }}>
        {/* Layer Stack Visualization */}
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 1 }}>
          {enabledLayers.map((layer, index) => (
            <Paper
              key={layer.id}
              elevation={2}
              sx={{
                p: 2,
                borderLeft: 4,
                borderColor: layer.color,
                bgcolor: 'white',
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <Box
                  sx={{
                    width: 12,
                    height: 12,
                    borderRadius: '50%',
                    bgcolor: layer.color,
                    mr: 1,
                  }}
                />
                <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
                  {layer.name}
                </Typography>
                <Typography variant="caption" color="text.secondary" sx={{ ml: 1 }}>
                  (Layer {enabledLayers.length - index})
                </Typography>
              </Box>

              {/* Functions in this layer */}
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                {functionsByLayer[layer.name]?.map((func, idx) => (
                  <Chip
                    key={`${func.layer}-${func.name}-${idx}`}
                    label={`fn ${func.name}()`}
                    size="small"
                    sx={{
                      bgcolor: `${layer.color}20`,
                      borderColor: layer.color,
                      border: 1,
                      fontFamily: 'monospace',
                    }}
                  />
                )) || (
                  <Typography variant="caption" color="text.secondary" sx={{ fontStyle: 'italic' }}>
                    No functions defined
                  </Typography>
                )}
              </Box>

              {/* Show what this layer can access */}
              {index < enabledLayers.length - 1 && (
                <Box sx={{ mt: 1, pt: 1, borderTop: 1, borderColor: 'divider' }}>
                  <Typography variant="caption" color="text.secondary">
                    Can call: functions from{' '}
                    <strong>{enabledLayers.slice(index + 1).map(l => l.name).join(', ')}</strong>
                  </Typography>
                </Box>
              )}
            </Paper>
          ))}
        </Box>

        {/* Legend */}
        <Box sx={{ mt: 3, pt: 2, borderTop: 2, borderColor: 'divider' }}>
          <Typography variant="caption" sx={{ fontWeight: 600, display: 'block', mb: 1 }}>
            Layer Composition Rules:
          </Typography>
          <Box sx={{ pl: 2 }}>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
              • Layers are stacked from bottom (core) to top
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
              • Higher layers can override or wrap functions from lower layers
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
              • Use <code style={{ background: '#f5f5f5', padding: '2px 4px' }}>core.function()</code> to call original implementation
            </Typography>
          </Box>
        </Box>
      </Paper>
    </Box>
  )
}

export default FlowDiagram
