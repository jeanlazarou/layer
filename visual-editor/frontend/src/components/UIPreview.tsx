import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Box, Typography, Button, TextField } from '@mui/material'

interface UIPreviewProps {
  /** XML produced by the ui layer. */
  xmlContent: string
  /** Runs a Layer expression with the current input values bound as variables. */
  onExecute: (expression: string, inputs: Record<string, string | number>) => Promise<string>
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

/** Input names and types (text, number, ...) declared anywhere in the tree. */
function inputTypes(component: UIComponent | null): Record<string, string> {
  const types: Record<string, string> = {}
  const walk = (c: UIComponent | string) => {
    if (typeof c === 'string') return
    if (c.type === 'input' && c.props.name) types[c.props.name] = c.props.type || 'text'
    c.children.forEach(walk)
  }
  if (component) walk(component)
  return types
}

const UIPreview: React.FC<UIPreviewProps> = ({ xmlContent, onExecute }) => {
  const component = useMemo(() => parseXML(xmlContent), [xmlContent])
  const types = useMemo(() => inputTypes(component), [component])
  const [inputs, setInputs] = useState<Record<string, string>>({})
  const [result, setResult] = useState<{ label: string; value: string } | null>(null)

  // Make sure every declared input has a value, so handlers can always
  // reference it (an empty string until the user types).
  useEffect(() => {
    setInputs(prev => {
      const next = { ...prev }
      for (const name of Object.keys(types)) if (!(name in next)) next[name] = ''
      return next
    })
  }, [types])

  const handleClick = useCallback(
    async (props: Record<string, string>) => {
      const values: Record<string, string | number> = {}
      for (const [name, value] of Object.entries(inputs)) {
        const num = parseFloat(value)
        values[name] = types[name] === 'number' && value.trim() !== '' && !isNaN(num) ? num : value
      }
      const value = await onExecute(props.onclick, values)
      // target: "<input name>" writes the result into that input (e.g. a
      // calculator display); otherwise it is shown below the UI.
      if (props.target && props.target in types && !value.startsWith('Error')) {
        setInputs(prev => ({ ...prev, [props.target]: value }))
      } else {
        setResult({ label: props.label || props.onclick, value })
      }
    },
    [inputs, types, onExecute],
  )

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
      case 'button':
        return (
          <Button
            key={key}
            variant="contained"
            disabled={!props.onclick}
            onClick={() => handleClick(props)}
            sx={{ mr: 1, mb: 1, textTransform: 'none' }}
          >
            {props.label ?? kids}
          </Button>
        )
      case 'input': {
        const name = props.name || `input_${key}`
        return (
          <TextField
            key={key}
            size="small"
            type={props.type || 'text'}
            placeholder={props.placeholder}
            label={props.label}
            value={inputs[name] ?? props.value ?? ''}
            onChange={e => setInputs(prev => ({ ...prev, [name]: e.target.value }))}
            sx={{ mb: 1, mr: 1, width: props.width || '200px' }}
          />
        )
      }
      case 'container': {
        const horizontal = props.layout === 'horizontal'
        return (
          <Box key={key} sx={{ display: 'flex', flexDirection: horizontal ? 'row' : 'column', gap: horizontal ? 2 : 1, mb: 1, flexWrap: 'wrap' }}>
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

  return (
    <Box sx={{ p: 2 }}>
      {render(component, 0)}
      {result && (
        <Box sx={{ mt: 2, p: 2, bgcolor: '#e3f2fd', borderRadius: 1, border: '2px solid #2196f3' }}>
          <Typography variant="subtitle2" sx={{ fontWeight: 600, color: '#1976d2', mb: 0.5 }}>
            {result.label}
          </Typography>
          <Typography variant="body1" sx={{ fontFamily: 'monospace', color: result.value.startsWith('Error') ? 'error.main' : '#333', whiteSpace: 'pre-wrap' }}>
            {result.value}
          </Typography>
        </Box>
      )}
    </Box>
  )
}

export default UIPreview
