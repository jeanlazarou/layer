import React, { useEffect, useState } from 'react'
import { Box, Typography, IconButton, Tooltip, ToggleButton, ToggleButtonGroup, CircularProgress } from '@mui/material'
import TerminalIcon from '@mui/icons-material/Terminal'
import OpenInFullIcon from '@mui/icons-material/OpenInFull'
import UIPreview from './UIPreview'
import type { LocatedIssue } from '../types'

export interface OutputState {
  status: 'idle' | 'running' | 'done'
  success: boolean
  result: string
  kind: 'value' | 'ui' | 'none'
  logs: string[]
  errors: LocatedIssue[]
  /** e.g. "Loaded fibonacci.layer" */
  notice?: string
}

export const IDLE_OUTPUT: OutputState = { status: 'idle', success: true, result: '', kind: 'none', logs: [], errors: [] }

interface OutputConsoleProps {
  output: OutputState
  describeLocation: (issue: LocatedIssue) => string
  onIssueClick: (issue: LocatedIssue) => void
  onMaximize?: () => void
  onExecute: (expression: string, inputs: Record<string, string | number>) => Promise<string>
}

const mono = '"Fira Code", "SFMono-Regular", Menlo, monospace'

function logColor(line: string): string {
  if (line.startsWith('[WARN]')) return '#ffb74d'
  if (line.startsWith('[ERROR]')) return '#ef9a9a'
  if (line.startsWith('[TRACE]')) return '#90caf9'
  return '#9e9e9e'
}

const OutputConsole: React.FC<OutputConsoleProps> = ({ output, describeLocation, onIssueClick, onMaximize, onExecute }) => {
  const isUI = output.status === 'done' && output.success && output.kind === 'ui'
  const [view, setView] = useState<'preview' | 'text'>('preview')

  // Every new UI result opens in the live preview.
  useEffect(() => {
    if (isUI) setView('preview')
  }, [output, isUI])

  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', bgcolor: '#1e1e1e', minHeight: 0 }}>
      <Box sx={{ px: 1.5, py: 0.5, display: 'flex', alignItems: 'center', gap: 1, bgcolor: '#2d2d2d', borderBottom: '1px solid #444' }}>
        <TerminalIcon sx={{ fontSize: 16, color: '#4caf50' }} />
        <Typography variant="subtitle2" sx={{ color: '#fff', fontWeight: 600, fontSize: '0.8rem', flexGrow: 1 }}>
          Output
        </Typography>
        {output.status === 'running' && <CircularProgress size={14} sx={{ color: '#4caf50' }} />}
        {isUI && (
          <ToggleButtonGroup
            size="small"
            exclusive
            value={view}
            onChange={(_, v) => v && setView(v)}
            sx={{ '& .MuiToggleButton-root': { color: '#aaa', py: 0.1, px: 1, fontSize: '0.7rem', borderColor: '#555', textTransform: 'none' }, '& .Mui-selected': { color: '#fff !important', bgcolor: '#455a64 !important' } }}
          >
            <ToggleButton value="preview">Preview</ToggleButton>
            <ToggleButton value="text">XML</ToggleButton>
          </ToggleButtonGroup>
        )}
        {onMaximize && (
          <Tooltip title="Maximize">
            <IconButton size="small" onClick={onMaximize} sx={{ color: '#888', p: 0.5 }}>
              <OpenInFullIcon sx={{ fontSize: 16 }} />
            </IconButton>
          </Tooltip>
        )}
      </Box>

      {isUI && view === 'preview' ? (
        <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}>
          <Box sx={{ flex: 1, overflow: 'auto', bgcolor: '#f5f5f5' }}>
            <UIPreview xmlContent={output.result} onExecute={onExecute} />
          </Box>
          {output.logs.length > 0 && (
            <Box sx={{ maxHeight: 110, overflow: 'auto', px: 1.5, py: 0.75, bgcolor: '#263238', fontFamily: mono, fontSize: 12, borderTop: '1px solid #444' }}>
              {output.logs.map((l, i) => (
                <div key={i} style={{ color: logColor(l), whiteSpace: 'pre-wrap' }}>{l}</div>
              ))}
            </Box>
          )}
        </Box>
      ) : (
        <Box sx={{ flex: 1, p: 1.5, overflow: 'auto', fontFamily: mono, fontSize: 12, color: '#d4d4d4' }}>
          {output.status === 'idle' && !output.notice && (
            <Box sx={{ color: '#888' }}>
              <div>&gt; Ready. Press ▶ or ⌘/Ctrl + Enter to run the program.</div>
              <div>&gt; Toggle layers on the left to change what the program does.</div>
            </Box>
          )}
          {output.notice && <Box sx={{ color: '#80cbc4', mb: 1 }}>{output.notice}</Box>}

          {output.logs.map((l, i) => (
            <div key={i} style={{ color: logColor(l), whiteSpace: 'pre-wrap' }}>{l}</div>
          ))}

          {output.status === 'done' && output.success && output.kind !== 'none' && (
            <Box sx={{ mt: output.logs.length ? 1 : 0 }}>
              <span style={{ color: '#4caf50' }}>⇒ </span>
              <span style={{ color: '#fff', whiteSpace: 'pre-wrap' }}>{output.result}</span>
            </Box>
          )}

          {output.errors.map((issue, i) => {
            const where = describeLocation(issue)
            return (
              <Box
                key={i}
                onClick={() => onIssueClick(issue)}
                sx={{
                  mt: 0.75,
                  p: 0.75,
                  borderLeft: '3px solid #f44336',
                  bgcolor: 'rgba(244, 67, 54, 0.08)',
                  cursor: issue.target ? 'pointer' : 'default',
                  '&:hover': issue.target ? { bgcolor: 'rgba(244, 67, 54, 0.16)' } : {},
                }}
              >
                <span style={{ color: '#ef9a9a' }}>{issue.kind} error</span>
                {where && <span style={{ color: '#90caf9' }}> at {where}</span>}
                <span style={{ color: '#eee' }}>: {issue.message}</span>
                {issue.trace && <div style={{ color: '#9e9e9e', marginTop: 2 }}>in {issue.trace}</div>}
              </Box>
            )
          })}
        </Box>
      )}
    </Box>
  )
}

export default OutputConsole
