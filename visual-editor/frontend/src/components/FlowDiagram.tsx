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
    <Box sx={{ p: 1.5, height: '100%', overflow: 'auto' }}>
      <Typography variant="subtitle1" sx={{ mb: 1.5, fontWeight: 600, fontSize: '0.95rem' }}>
        Function Flow
      </Typography>

      <Paper variant="outlined" sx={{ p: 1.25, bgcolor: 'grey.50' }}>
        {/* Layer Stack Visualization */}
        <Box sx={{ display: 'flex', flexDirection: 'column', gap: 0.75 }}>
          {enabledLayers.map((layer, index) => (
            <Paper
              key={layer.id}
              elevation={1}
              sx={{
                p: 1.25,
                borderLeft: 3,
                borderColor: layer.color,
                bgcolor: 'white',
              }}
            >
              <Box sx={{ display: 'flex', alignItems: 'center', mb: 0.75 }}>
                <Box
                  sx={{
                    width: 8,
                    height: 8,
                    borderRadius: '50%',
                    bgcolor: layer.color,
                    mr: 0.75,
                  }}
                />
                <Typography variant="body2" sx={{ fontWeight: 600, fontSize: '0.8rem' }}>
                  {layer.name}
                </Typography>
                <Typography variant="caption" color="text.secondary" sx={{ ml: 0.75, fontSize: '0.65rem' }}>
                  (L{enabledLayers.length - index})
                </Typography>
              </Box>

              {/* Functions in this layer */}
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                {functionsByLayer[layer.name]?.map((func, idx) => (
                  <Chip
                    key={`${func.layer}-${func.name}-${idx}`}
                    label={`${func.name}()`}
                    size="small"
                    sx={{
                      height: 20,
                      fontSize: '0.7rem',
                      bgcolor: `${layer.color}20`,
                      borderColor: layer.color,
                      border: 1,
                      fontFamily: 'monospace',
                      '& .MuiChip-label': {
                        px: 0.75,
                        py: 0,
                      }
                    }}
                  />
                )) || (
                  <Typography variant="caption" color="text.secondary" sx={{ fontStyle: 'italic', fontSize: '0.7rem' }}>
                    No functions
                  </Typography>
                )}
              </Box>

              {/* Show what this layer can access */}
              {index < enabledLayers.length - 1 && (
                <Box sx={{ mt: 0.75, pt: 0.75, borderTop: 1, borderColor: 'divider' }}>
                  <Typography variant="caption" color="text.secondary" sx={{ fontSize: '0.65rem', lineHeight: 1.3 }}>
                    Calls: <strong>{enabledLayers.slice(index + 1).map(l => l.name).join(', ')}</strong>
                  </Typography>
                </Box>
              )}
            </Paper>
          ))}
        </Box>

        {/* Legend */}
        <Box sx={{ mt: 2, pt: 1.5, borderTop: 1, borderColor: 'divider' }}>
          <Typography variant="caption" sx={{ fontWeight: 600, display: 'block', mb: 0.75, fontSize: '0.7rem' }}>
            Rules:
          </Typography>
          <Box sx={{ pl: 1.5 }}>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', fontSize: '0.65rem', lineHeight: 1.4 }}>
              • Bottom-to-top stack
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', fontSize: '0.65rem', lineHeight: 1.4 }}>
              • Higher override lower
            </Typography>
            <Typography variant="caption" color="text.secondary" sx={{ display: 'block', fontSize: '0.65rem', lineHeight: 1.4 }}>
              • <code style={{ background: '#f5f5f5', padding: '1px 3px', fontSize: '0.65rem' }}>core.fn()</code> for original
            </Typography>
          </Box>
        </Box>
      </Paper>
    </Box>
  )
}

export default FlowDiagram
