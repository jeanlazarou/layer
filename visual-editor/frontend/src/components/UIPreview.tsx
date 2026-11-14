import React, { useEffect, useMemo, useState } from 'react'
import { Box, Typography, Button, TextField } from '@mui/material'

interface UIPreviewProps {
  /** XML produced by the ui layer (an app's current view, or a static UI). */
  xmlContent: string
  /**
   * Sends a message to the running app; group identifies a source whose
   * consecutive messages may be merged in the history (an input). Absent for
   * static UI.
   */
  onMessage?: (msg: unknown, group?: string) => void
}

interface UIComponent {
  type: string
  props: Record<string, string>
  children: (UIComponent | string)[]
}

function parseXML(xml: string): UIComponent | null {
  const doc = new DOMParser().parseFromString(xml.trim(), 'text/xml')
  if (doc.getElementsByTagName('parsererror').length > 0) return null

  const parseNode = (node: Element): UIComponent => {
    const props: Record<string, string> = {}
    for (const attr of Array.from(node.attributes)) props[attr.name] = attr.value
    const children: (UIComponent | string)[] = []
    for (const child of Array.from(node.childNodes)) {
      if (child.nodeType === Node.ELEMENT_NODE) {
        children.push(parseNode(child as Element))
      } else if (child.nodeType === Node.TEXT_NODE) {
        const text = child.textContent?.trim()
        if (text) children.push(text)
      }
    }
    return { type: node.nodeName, props, children }
  }
  return parseNode(doc.documentElement)
}

/** Handlers (on_click, on_input) are rendered by the interpreter as JSON. */
function parseHandler(raw: string | undefined): { ok: true; value: unknown } | { ok: false } {
  if (raw === undefined) return { ok: false }
  try {
    return { ok: true, value: JSON.parse(raw) }
  } catch {
    return { ok: false }
  }
}

/** Two texts denote the same value of an input (so a draft like "1." is kept). */
function sameValue(a: string, b: string, type: string | undefined): boolean {
  if (type === 'number') return a.trim() !== '' && b.trim() !== '' && Number(a) === Number(b)
  return a === b
}

/**
 * An input keeps a local draft while you type: the app re-renders with the
 * model's value after each message, and the draft is only replaced when that
 * value really differs (e.g. update cleared the field).
 */
function AppInput({ props, onMessage }: { props: Record<string, string>; onMessage?: (msg: unknown, group?: string) => void }) {
  const value = props.value ?? ''
  const [text, setText] = useState(value)
  useEffect(() => {
    setText(prev => (sameValue(prev, value, props.type) ? prev : value))
  }, [value, props.type])

  const handler = parseHandler(props.on_input)
  const tag = handler.ok && typeof handler.value === 'string' ? handler.value : undefined

  const handleChange = (next: string) => {
    setText(next)
    if (!tag || !onMessage) return
    if (props.type === 'number') {
      if (next.trim() === '' || !Number.isFinite(Number(next))) return // wait for a number
      onMessage({ [tag]: Number(next) }, `input:${tag}`)
    } else {
      onMessage({ [tag]: next }, `input:${tag}`)
    }
  }

  return (
    <TextField
      size="small"
      type={props.type || 'text'}
      placeholder={props.placeholder}
      label={props.label}
      value={text}
      onChange={e => handleChange(e.target.value)}
      inputProps={{ name: props.name }}
      sx={{ mb: 1, mr: 1, width: props.width || '200px' }}
    />
  )
}

const UIPreview: React.FC<UIPreviewProps> = ({ xmlContent, onMessage }) => {
  const component = useMemo(() => parseXML(xmlContent), [xmlContent])

  const render = (node: UIComponent | string, key: number): React.ReactNode => {
    if (typeof node === 'string') return <span key={key}>{node}</span>
    const { type, props, children } = node
    const kids = children.map((child, i) => render(child, i))

    switch (type) {
      case 'view':
        return (
          <Box key={key} sx={{ border: '2px solid #2196f3', borderRadius: 2, p: 2, mb: 2, bgcolor: '#f5f9ff' }}>
            {props.title && (
              <Typography variant="h6" sx={{ mb: 1, color: '#1976d2', fontWeight: 600 }}>
                {props.title}
              </Typography>
            )}
            {kids}
          </Box>
        )
      case 'text':
        return (
          <Typography key={key} variant="body2" sx={{ mb: 0.5, color: '#333' }}>
            {props.content ?? kids}
          </Typography>
        )
      case 'button': {
        const handler = parseHandler(props.on_click)
        const enabled = handler.ok && !!onMessage
        return (
          <Button
            key={key}
            variant="contained"
            disabled={!enabled}
            title={handler.ok && !onMessage ? 'Buttons send messages to an app — see app({init, update, view})' : undefined}
            onClick={() => handler.ok && onMessage?.(handler.value)}
            sx={{ mr: 1, mb: 1, textTransform: 'none' }}
          >
            {props.label ?? kids}
          </Button>
        )
      }
      case 'input':
        return <AppInput key={props.name ?? key} props={props} onMessage={onMessage} />
      case 'container': {
        const horizontal = props.layout === 'horizontal'
        return (
          <Box key={key} sx={{ display: 'flex', flexDirection: horizontal ? 'row' : 'column', gap: horizontal ? 2 : 1, mb: 1, flexWrap: 'wrap', alignItems: horizontal ? 'center' : 'stretch' }}>
            {kids}
          </Box>
        )
      }
      default:
        return (
          <Box key={key} sx={{ mb: 1 }}>
            <Typography variant="caption" color="text.secondary">
              &lt;{type}&gt;
            </Typography>
            {kids}
          </Box>
        )
    }
  }

  if (!component) {
    return (
      <Box sx={{ p: 2, color: 'error.main' }}>
        <Typography variant="body2" sx={{ mb: 1, fontWeight: 600 }}>
          Could not render the UI output
        </Typography>
        <Typography variant="caption" component="pre" sx={{ display: 'block', fontFamily: 'monospace', bgcolor: '#fff', p: 1, borderRadius: 1, color: '#333', whiteSpace: 'pre-wrap' }}>
          {xmlContent.slice(0, 400)}
        </Typography>
      </Box>
    )
  }

  return <Box sx={{ p: 2 }}>{render(component, 0)}</Box>
}

export default UIPreview
