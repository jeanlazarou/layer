import React, { useEffect, useState } from 'react'
import { Box, Typography, IconButton, Tooltip, ToggleButton, ToggleButtonGroup, CircularProgress, Button } from '@mui/material'
import TerminalIcon from '@mui/icons-material/Terminal'
import OpenInFullIcon from '@mui/icons-material/OpenInFull'
import RestartAltIcon from '@mui/icons-material/RestartAlt'
import UIPreview from './UIPreview'
import type { AppSession, LocatedIssue } from '../types'

export interface OutputState {
  status: 'idle' | 'running' | 'done'
  success: boolean
  result: string
  kind: 'value' | 'ui' | 'app' | 'none'
  logs: string[]
  errors: LocatedIssue[]
  /** e.g. "Opened fibonacci.layer" */
  notice?: string
}

export const IDLE_OUTPUT: OutputState = { status: 'idle', success: true, result: '', kind: 'none', logs: [], errors: [] }

interface OutputConsoleProps {
  output: OutputState
  /** Set while an app is running (the program's main expression is app({...})). */
  app: AppSession | null
  describeLocation: (issue: LocatedIssue) => string
  onIssueClick: (issue: LocatedIssue) => void
  onMessage: (msg: unknown, group?: string) => void
  onTimeTravel: (index: number) => void
  onRestart: () => void
  onMaximize?: () => void
}

const mono = '"Fira Code", "SFMono-Regular", Menlo, monospace'

function logColor(line: string): string {
  if (line.startsWith('[WARN]')) return '#ffb74d'
  if (line.startsWith('[ERROR]')) return '#ef9a9a'
  if (line.startsWith('[TRACE]')) return '#90caf9'
  return '#9e9e9e'
}

/** A plain-data value in Layer syntax (as the interpreter's Show prints it). */
function showData(v: unknown): string {
  if (v === null || v === undefined) return 'nil'
  if (typeof v === 'string') return JSON.stringify(v)
  if (Array.isArray(v)) return `[${v.map(showData).join(', ')}]`
  if (typeof v === 'object') {
    const fields = Object.keys(v as Record<string, unknown>).sort()
    return `{${fields.map(k => `${k}: ${showData((v as Record<string, unknown>)[k])}`).join(', ')}}`
  }
  return String(v)
}

/** Short display of a message for the history list. */
function showMessage(msg: unknown): string {
  return msg === undefined ? 'init' : showData(msg)
}

function Logs({ lines, maxHeight }: { lines: string[]; maxHeight?: number }) {
  if (lines.length === 0) return null
  return (
    <Box sx={{ maxHeight, overflow: 'auto', px: 1.5, py: 0.75, bgcolor: '#263238', fontFamily: mono, fontSize: 12, borderTop: '1px solid #444' }}>
      {lines.map((l, i) => (
        <div key={i} style={{ color: logColor(l), whiteSpace: 'pre-wrap' }}>
          {l}
        </div>
      ))}
    </Box>
  )
}

function Errors({ errors, describeLocation, onIssueClick }: { errors: LocatedIssue[]; describeLocation: (issue: LocatedIssue) => string; onIssueClick: (issue: LocatedIssue) => void }) {
  return (
    <>
      {errors.map((issue, i) => {
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
              fontFamily: mono,
              fontSize: 12,
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
    </>
  )
}

/** The running app's model and the messages that led to it. */
function AppInspector({ app, onTimeTravel, onRestart }: { app: AppSession; onTimeTravel: (i: number) => void; onRestart: () => void }) {
  const current = app.steps[app.index]
  return (
    <Box sx={{ width: 260, flexShrink: 0, borderLeft: '1px solid #ddd', bgcolor: '#fff', display: 'flex', flexDirection: 'column', minHeight: 0 }}>
      <Box sx={{ px: 1.25, py: 0.75, display: 'flex', alignItems: 'center', borderBottom: '1px solid #eee' }}>
        <Typography variant="caption" sx={{ fontWeight: 700, flexGrow: 1 }}>
          Model
        </Typography>
        <Tooltip title="Restart the app from its init">
          <Button size="small" startIcon={<RestartAltIcon sx={{ fontSize: 14 }} />} onClick={onRestart} sx={{ fontSize: '0.7rem', py: 0, textTransform: 'none' }}>
            Restart
          </Button>
        </Tooltip>
      </Box>
      <Box component="pre" aria-label="app model" sx={{ m: 0, px: 1.25, py: 0.75, fontFamily: mono, fontSize: 11, whiteSpace: 'pre-wrap', wordBreak: 'break-word', maxHeight: 120, overflow: 'auto', borderBottom: '1px solid #eee' }}>
        {current.modelText}
      </Box>
      <Typography variant="caption" sx={{ fontWeight: 700, px: 1.25, pt: 0.75 }}>
        Messages <span style={{ fontWeight: 400, color: '#888' }}>— click one to go back to that state</span>
      </Typography>
      <Box component="ol" aria-label="app messages" start={0} sx={{ m: 0, px: 1.25, py: 0.5, pl: 3.5, flex: 1, overflow: 'auto', fontFamily: mono, fontSize: 11 }}>
        {app.steps.map((step, i) => (
          <Box
            component="li"
            key={i}
            onClick={() => onTimeTravel(i)}
            sx={{
              cursor: 'pointer',
              py: 0.15,
              color: i > app.index ? '#bbb' : '#333',
              fontWeight: i === app.index ? 700 : 400,
              bgcolor: i === app.index ? '#e3f2fd' : 'transparent',
              '&:hover': { bgcolor: '#f0f7ff' },
            }}
          >
            {showMessage(step.msg)}
          </Box>
        ))}
      </Box>
    </Box>
  )
}

const OutputConsole: React.FC<OutputConsoleProps> = ({ output, app, describeLocation, onIssueClick, onMessage, onTimeTravel, onRestart, onMaximize }) => {
  const appView = app ? app.steps[app.index].view : null
  const staticUI = !app && output.status === 'done' && output.success && output.kind === 'ui'
  const visual = appView !== null || staticUI
  const [view, setView] = useState<'preview' | 'text'>('preview')

  // A new UI or app result opens in the live preview.
  useEffect(() => {
    if (visual) setView('preview')
  }, [visual])

  const xml = appView ?? output.result

  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column', bgcolor: '#1e1e1e', minHeight: 0 }}>
      <Box sx={{ px: 1.5, py: 0.5, display: 'flex', alignItems: 'center', gap: 1, bgcolor: '#2d2d2d', borderBottom: '1px solid #444' }}>
        <TerminalIcon sx={{ fontSize: 16, color: '#4caf50' }} />
        <Typography variant="subtitle2" sx={{ color: '#fff', fontWeight: 600, fontSize: '0.8rem', flexGrow: 1 }}>
          {app ? 'Output — app running' : 'Output'}
        </Typography>
        {output.status === 'running' && <CircularProgress size={14} sx={{ color: '#4caf50' }} />}
        {visual && (
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

      {visual && view === 'preview' ? (
        <Box sx={{ flex: 1, display: 'flex', minHeight: 0 }}>
          <Box sx={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
            {output.errors.length > 0 && (
              <Box sx={{ px: 1.5, pb: 0.75, bgcolor: '#1e1e1e' }}>
                <Errors errors={output.errors} describeLocation={describeLocation} onIssueClick={onIssueClick} />
              </Box>
            )}
            <Box sx={{ flex: 1, overflow: 'auto', bgcolor: '#f5f5f5' }}>
              <UIPreview xmlContent={xml} onMessage={app ? onMessage : undefined} />
            </Box>
            <Logs lines={output.logs} maxHeight={110} />
          </Box>
          {app && <AppInspector app={app} onTimeTravel={onTimeTravel} onRestart={onRestart} />}
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
            <div key={i} style={{ color: logColor(l), whiteSpace: 'pre-wrap' }}>
              {l}
            </div>
          ))}

          {visual ? (
            <pre style={{ margin: 0, whiteSpace: 'pre-wrap', color: '#fff' }}>{xml}</pre>
          ) : (
            output.status === 'done' &&
            output.success &&
            output.kind !== 'none' && (
              <Box sx={{ mt: output.logs.length ? 1 : 0 }}>
                <span style={{ color: '#4caf50' }}>⇒ </span>
                <span style={{ color: '#fff', whiteSpace: 'pre-wrap' }}>{output.result}</span>
              </Box>
            )
          )}

          <Errors errors={output.errors} describeLocation={describeLocation} onIssueClick={onIssueClick} />
        </Box>
      )}
    </Box>
  )
}

export default OutputConsole
