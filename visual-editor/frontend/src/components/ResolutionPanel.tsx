import React from 'react'
import { Box, Paper, Typography, Tooltip, IconButton } from '@mui/material'
import ArrowForwardIcon from '@mui/icons-material/ArrowForward'
import CloseIcon from '@mui/icons-material/Close'
import type { BlendMode, BlendSettings, Description, Layer } from '../types'
import { BLEND_MODES } from '../types'
import { colorFor } from '../program'

interface ResolutionPanelProps {
  description: Description | null
  layers: Layer[]
  blend: BlendSettings
  /** Set (or with null, clear) the blend mode of one function. */
  onFunctionModeChange: (name: string, mode: BlendMode | null) => void
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

/** How a call to `name` resolves through the stack under its blend mode. */
function Chain({ name, layers, mode, colorOf }: { name: string; layers: string[]; mode: BlendMode; colorOf: (l: string) => string }) {
  // layers is top of the stack first.
  if (layers.length === 1 || mode === 'override') {
    return (
      <>
        {layers.map((layer, i) => (
          <LayerChip
            key={layer}
            name={layer}
            color={colorOf(layer)}
            dim={i > 0}
            title={i === 0 ? `${name}() runs ${layer}.${name}` : `Shadowed — still reachable explicitly as ${layer}.${name}()`}
          />
        ))}
      </>
    )
  }
  if (mode === 'chain') {
    // Bottom layer runs first and pipes its result upward.
    return (
      <>
        {[...layers].reverse().map((layer, i) => (
          <React.Fragment key={layer}>
            {i > 0 && <ArrowForwardIcon sx={{ fontSize: 12, color: 'text.disabled' }} />}
            <LayerChip name={layer} color={colorOf(layer)} title={i === 0 ? 'Runs first with the original arguments' : 'Receives the previous result as its first argument'} />
          </React.Fragment>
        ))}
      </>
    )
  }
  return (
    <>
      {layers.map((layer, i) => (
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
      ))}
    </>
  )
}

/** Compact per-function mode selector: "default" follows the stack's mode. */
function ModeSelect({ name, value, stackMode, onChange }: { name: string; value: BlendMode | undefined; stackMode: BlendMode; onChange: (mode: BlendMode | null) => void }) {
  return (
    <Tooltip title={`How the definitions of ${name}() combine. "default" follows the toolbar's blend mode.`}>
      <Box
        component="select"
        aria-label={`blend mode for ${name}`}
        value={value ?? ''}
        onChange={(e: React.ChangeEvent<HTMLSelectElement>) => onChange(e.target.value ? (e.target.value as BlendMode) : null)}
        sx={{
          ml: 'auto',
          fontSize: '0.68rem',
          py: 0.1,
          border: '1px solid',
          borderColor: value ? 'primary.main' : 'divider',
          borderRadius: 1,
          bgcolor: 'background.paper',
          color: value ? 'primary.main' : 'text.secondary',
        }}
      >
        <option value="">default ({stackMode})</option>
        {BLEND_MODES.map(m => (
          <option key={m.value} value={m.value}>
            {m.value}
          </option>
        ))}
      </Box>
    </Tooltip>
  )
}

const ResolutionPanel: React.FC<ResolutionPanelProps> = ({ description, layers, blend, onFunctionModeChange }) => {
  const colorOf = (name: string) => layers.find(l => l.name === name)?.color ?? colorFor(name)
  const stackMode = BLEND_MODES.find(m => m.value === blend.mode)
  const defined = new Set(description?.resolutions.map(r => r.name) ?? [])
  const unused = Object.entries(blend.perFunction).filter(([name]) => !defined.has(name))

  return (
    <Box sx={{ p: 1.5, height: '100%', overflow: 'auto' }}>
      <Typography variant="subtitle1" sx={{ fontWeight: 600, fontSize: '0.95rem' }}>
        Who answers each call?
      </Typography>
      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1.5, fontSize: '0.7rem' }}>
        Stack blend mode <strong>{stackMode?.label}</strong>: {stackMode?.help}. Functions defined in several layers can pick their own mode.
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
              description.resolutions.map(r => (
                <Box key={r.name} sx={{ py: 0.6, borderBottom: '1px solid', borderColor: 'divider', '&:last-child': { borderBottom: 0 } }}>
                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                    <Typography sx={{ fontFamily: 'monospace', fontSize: '0.75rem', fontWeight: 600 }}>{r.name}()</Typography>
                    {r.layers.length > 1 && (
                      <ModeSelect
                        name={r.name}
                        value={blend.perFunction[r.name]}
                        stackMode={blend.mode}
                        onChange={mode => onFunctionModeChange(r.name, mode)}
                      />
                    )}
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'center', flexWrap: 'wrap', gap: 0.5, mt: 0.4 }}>
                    <Chain name={r.name} layers={r.layers} mode={r.mode} colorOf={colorOf} />
                  </Box>
                </Box>
              ))
            )}
          </Paper>

          {unused.length > 0 && (
            <Box sx={{ mb: 2 }}>
              <Typography variant="caption" sx={{ fontWeight: 600, display: 'block' }}>
                Saved modes for functions not defined right now
              </Typography>
              {unused.map(([name, mode]) => (
                <Box key={name} sx={{ display: 'flex', alignItems: 'center', gap: 0.5 }}>
                  <Typography sx={{ fontFamily: 'monospace', fontSize: '0.7rem' }}>
                    {name}: {mode}
                  </Typography>
                  <IconButton size="small" aria-label={`remove blend mode for ${name}`} onClick={() => onFunctionModeChange(name, null)} sx={{ p: 0.25 }}>
                    <CloseIcon sx={{ fontSize: 12 }} />
                  </IconButton>
                </Box>
              ))}
            </Box>
          )}

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
