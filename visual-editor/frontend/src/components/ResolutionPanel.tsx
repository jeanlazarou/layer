import React from 'react'
import { Box, Paper, Typography, Tooltip } from '@mui/material'
import ArrowForwardIcon from '@mui/icons-material/ArrowForward'
import type { BlendMode, Description, Layer } from '../types'
import { BLEND_MODES } from '../types'
import { colorFor } from '../program'

interface ResolutionPanelProps {
  description: Description | null
  layers: Layer[]
  blendMode: BlendMode
}

function LayerChip({ name, color, dim, title }: { name: string; color: string; dim?: boolean; title: string }) {
  return (
    <Tooltip title={title}>
      <Box
        component="span"
        sx={{
          fontFamily: 'monospace',
          fontSize: '0.7rem',
          px: 0.75,
          py: 0.1,
          borderRadius: 1,
          border: `1px solid ${color}`,
          bgcolor: dim ? 'transparent' : `${color}22`,
          color: dim ? 'text.disabled' : 'text.primary',
          textDecoration: dim ? 'line-through' : 'none',
          whiteSpace: 'nowrap',
        }}
      >
        {name}
      </Box>
    </Tooltip>
  )
}

/** How a call to `name` resolves through the stack, per blend mode. */
function ResolutionRow({ name, chain, mode, colorOf }: { name: string; chain: string[]; mode: BlendMode; colorOf: (l: string) => string }) {
  // chain is top of the stack first.
  let items: React.ReactNode
  if (chain.length === 1 || mode === 'override') {
    items = chain.map((layer, i) => (
      <LayerChip
        key={layer}
        name={layer}
        color={colorOf(layer)}
        dim={i > 0}
        title={i === 0 ? `${name}() runs ${layer}.${name}` : `Shadowed — still reachable explicitly as ${layer}.${name}()`}
      />
    ))
  } else if (mode === 'chain') {
    // Bottom layer runs first and pipes its result upward.
    const order = [...chain].reverse()
    items = order.map((layer, i) => (
      <React.Fragment key={layer}>
        {i > 0 && <ArrowForwardIcon sx={{ fontSize: 12, color: 'text.disabled' }} />}
        <LayerChip name={layer} color={colorOf(layer)} title={i === 0 ? 'Runs first with the original arguments' : 'Receives the previous result as its first argument'} />
      </React.Fragment>
    ))
  } else {
    items = chain.map((layer, i) => (
      <React.Fragment key={layer}>
        {i > 0 && (
          <Typography component="span" sx={{ fontSize: '0.65rem', color: 'text.disabled' }}>
            {mode === 'collect' ? '+' : 'else'}
          </Typography>
        )}
        <LayerChip
          name={layer}
          color={colorOf(layer)}
          title={mode === 'collect' ? 'Runs; its result is collected' : i === 0 ? 'Tried first' : 'Tried when the layers above return None'}
        />
      </React.Fragment>
    ))
  }

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: 0.5, py: 0.6, borderBottom: '1px solid', borderColor: 'divider' }}>
      <Typography sx={{ fontFamily: 'monospace', fontSize: '0.75rem', fontWeight: 600, mr: 0.5 }}>{name}()</Typography>
      {items}
    </Box>
  )
}

const ResolutionPanel: React.FC<ResolutionPanelProps> = ({ description, layers, blendMode }) => {
  const colorOf = (name: string) => layers.find(l => l.name === name)?.color ?? colorFor(name)
  const mode = BLEND_MODES.find(m => m.value === blendMode)

  return (
    <Box sx={{ p: 1.5, height: '100%', overflow: 'auto' }}>
      <Typography variant="subtitle1" sx={{ fontWeight: 600, fontSize: '0.95rem' }}>
        Who answers each call?
      </Typography>
      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1.5, fontSize: '0.7rem' }}>
        Blend mode <strong>{mode?.label}</strong>: {mode?.help}
      </Typography>

      {!description ? (
        <Typography variant="caption" color="text.secondary">
          Fix the errors in the program to see how calls resolve.
        </Typography>
      ) : (
        <>
          <Paper variant="outlined" sx={{ p: 1, mb: 2 }}>
            {description.resolutions.length === 0 ? (
              <Typography variant="caption" color="text.secondary">
                No functions defined yet.
              </Typography>
            ) : (
              description.resolutions.map(r => <ResolutionRow key={r.name} name={r.name} chain={r.layers} mode={description.blendMode} colorOf={colorOf} />)
            )}
          </Paper>

          <Typography variant="caption" sx={{ fontWeight: 600, display: 'block', mb: 0.75 }}>
            Active stack (top first)
          </Typography>
          {[...description.layers].reverse().map(layer => (
            <Box key={layer.name} sx={{ display: 'flex', alignItems: 'baseline', gap: 0.75, mb: 0.5, pl: 1, borderLeft: `3px solid ${colorOf(layer.name)}` }}>
              <Typography sx={{ fontSize: '0.75rem', fontWeight: 600 }}>{layer.name}</Typography>
              <Typography sx={{ fontSize: '0.65rem', color: layer.pure ? '#2e7d32' : '#e65100' }}>{layer.pure ? 'pure' : 'impure'}</Typography>
              <Typography sx={{ fontSize: '0.65rem', color: 'text.secondary', fontFamily: 'monospace' }}>
                {layer.defines.join(', ')}
              </Typography>
            </Box>
          ))}
        </>
      )}
    </Box>
  )
}

export default ResolutionPanel
