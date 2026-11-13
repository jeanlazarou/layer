import React, { useEffect, useRef } from 'react'
import { Box, Typography, Chip, Tooltip } from '@mui/material'
import Editor, { type Monaco, type OnMount } from '@monaco-editor/react'
import { LAYER_LANGUAGE_ID, registerLayerLanguage } from '../language/layerLanguage'
import type { LocatedIssue } from '../types'

type StandaloneEditor = Parameters<OnMount>[0]

interface CodeEditorProps {
  /** Stable key per edited buffer (layer id or 'main'); keeps undo history separate. */
  bufferId: string
  code: string
  title: string
  color: string
  badge?: string
  issues: LocatedIssue[]
  onChange: (code: string) => void
  onRun: () => void
  /** Line to reveal (e.g. after clicking an error); changes trigger a jump. */
  reveal?: { line: number; column: number; nonce: number }
}

const CodeEditor: React.FC<CodeEditorProps> = ({ bufferId, code, title, color, badge, issues, onChange, onRun, reveal }) => {
  const editorRef = useRef<StandaloneEditor | null>(null)
  const monacoRef = useRef<Monaco | null>(null)
  const onRunRef = useRef(onRun)
  onRunRef.current = onRun

  const handleMount: OnMount = (ed, monaco) => {
    editorRef.current = ed
    monacoRef.current = monaco
    // Ctrl/Cmd+Enter runs the program from anywhere in the editor.
    ed.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => onRunRef.current())
    applyMarkers()
  }

  const applyMarkers = () => {
    const ed = editorRef.current
    const monaco = monacoRef.current
    const model = ed?.getModel()
    if (!ed || !monaco || !model) return
    monaco.editor.setModelMarkers(
      model,
      'layer',
      issues.map(issue => {
        const line = Math.min(Math.max(issue.targetLine, 1), model.getLineCount())
        const word = model.getWordAtPosition({ lineNumber: line, column: issue.targetColumn })
        return {
          severity: monaco.MarkerSeverity.Error,
          message: issue.trace ? `${issue.message}\nin ${issue.trace}` : issue.message,
          source: `${issue.kind} error`,
          startLineNumber: line,
          startColumn: issue.targetColumn,
          endLineNumber: line,
          endColumn: word ? word.endColumn : issue.targetColumn + 1,
        }
      }),
    )
  }

  // Re-apply markers whenever issues or the visible buffer change.
  useEffect(applyMarkers, [issues, bufferId])

  useEffect(() => {
    const ed = editorRef.current
    if (!ed || !reveal) return
    ed.revealLineInCenter(reveal.line)
    ed.setPosition({ lineNumber: reveal.line, column: reveal.column })
    ed.focus()
  }, [reveal])

  return (
    <Box sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <Box
        sx={{
          px: 2,
          py: 1,
          borderBottom: 1,
          borderColor: 'divider',
          display: 'flex',
          alignItems: 'center',
          gap: 1,
          bgcolor: 'background.paper',
        }}
      >
        <Chip label={title} size="small" sx={{ bgcolor: color, color: 'white', fontWeight: 600 }} />
        {badge && (
          <Typography variant="caption" color="text.secondary">
            {badge}
          </Typography>
        )}
        <Box sx={{ flexGrow: 1 }} />
        {issues.length > 0 && (
          <Tooltip title={issues.map(i => `${i.targetLine}:${i.targetColumn} ${i.message}`).join('\n')}>
            <Chip label={`${issues.length} error${issues.length > 1 ? 's' : ''}`} size="small" color="error" variant="outlined" />
          </Tooltip>
        )}
        <Typography variant="caption" color="text.disabled" sx={{ display: { xs: 'none', md: 'block' } }}>
          ⌘/Ctrl + Enter to run
        </Typography>
      </Box>

      <Box sx={{ flex: 1, minHeight: 0 }}>
        <Editor
          height="100%"
          path={`${bufferId}.layer`}
          language={LAYER_LANGUAGE_ID}
          value={code}
          onChange={value => onChange(value ?? '')}
          beforeMount={registerLayerLanguage}
          onMount={handleMount}
          theme="layer-light"
          options={{
            minimap: { enabled: false },
            fontSize: 14,
            lineNumbers: 'on',
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            wordWrap: 'on',
            fixedOverflowWidgets: true,
          }}
        />
      </Box>
    </Box>
  )
}

export default CodeEditor
