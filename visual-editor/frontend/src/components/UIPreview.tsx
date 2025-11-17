import React, { useState, useCallback } from 'react'
import { Box, Typography, Button, TextField } from '@mui/material'

interface UIPreviewProps {
  xmlContent: string
  onExecute?: (expression: string, inputs: Record<string, string>) => Promise<string>
}

interface UIComponent {
  type: string
  props: Record<string, string>
  children: (UIComponent | string)[]
}

const UIPreview: React.FC<UIPreviewProps> = ({ xmlContent, onExecute }) => {
  const [inputs, setInputs] = useState<Record<string, string>>({})
  const [results, setResults] = useState<Record<string, string>>({})

  const handleInputChange = useCallback((name: string, value: string) => {
    setInputs(prev => ({ ...prev, [name]: value }))
  }, [])

  const handleButtonClick = useCallback(async (onclick: string) => {
    if (!onExecute) {
      console.log('Button clicked:', onclick, 'with inputs:', inputs)
      return
    }

    try {
      // Execute the onclick expression with current input values
      const result = await onExecute(onclick, inputs)

      // Store result for display in the result box
      setResults(prev => ({ ...prev, [onclick]: result }))

      // Update the display input field if it exists and result is valid
      // This allows Layer functions to control what appears in the display
      if (xmlContent.includes('name="display"') && result && result !== 'undefined') {
        setInputs(prev => ({ ...prev, display: result }))
      }
    } catch (error) {
      console.error('Error executing onclick:', error)
    }
  }, [onExecute, inputs, xmlContent])
  const parseXML = (xml: string): UIComponent | null => {
    // Extract just the XML part (remove "Result: " prefix and any logs)
    let cleanXml = xml

    // Remove "Result: " prefix if present
    if (cleanXml.includes('Result: ')) {
      cleanXml = cleanXml.substring(cleanXml.indexOf('Result: ') + 8)
    }

    // Find the first < and last > to extract just the XML
    const firstTag = cleanXml.indexOf('<')
    const lastTag = cleanXml.lastIndexOf('>')

    if (firstTag !== -1 && lastTag !== -1 && lastTag > firstTag) {
      cleanXml = cleanXml.substring(firstTag, lastTag + 1).trim()
    }

    // Simple XML parser for UIComponent output
    const parser = new DOMParser()
    const doc = parser.parseFromString(cleanXml, 'text/xml')

    if (doc.documentElement.nodeName === 'parsererror') {
      console.error('XML Parse Error:', cleanXml)
      return null
    }

    const parseNode = (node: Element): UIComponent => {
      const type = node.nodeName
      const props: Record<string, string> = {}

      // Extract attributes
      for (let i = 0; i < node.attributes.length; i++) {
        const attr = node.attributes[i]
        props[attr.name] = attr.value
      }

      // Parse children
      const children: (UIComponent | string)[] = []
      for (let i = 0; i < node.childNodes.length; i++) {
        const child = node.childNodes[i]
        if (child.nodeType === Node.ELEMENT_NODE) {
          children.push(parseNode(child as Element))
        } else if (child.nodeType === Node.TEXT_NODE) {
          const text = child.textContent?.trim()
          if (text) children.push(text)
        }
      }

      return { type, props, children }
    }

    return parseNode(doc.documentElement)
  }

  const renderComponent = (component: UIComponent | string, key: number): React.ReactNode => {
    if (typeof component === 'string') {
      return <span key={key}>{component}</span>
    }

    const { type, props, children } = component

    switch (type) {
      case 'view':
        return (
          <Box
            key={key}
            sx={{
              border: '2px solid #2196f3',
              borderRadius: 2,
              p: 2,
              mb: 2,
              bgcolor: '#f5f9ff',
            }}
          >
            {props.title && (
              <Typography variant="h6" sx={{ mb: 1, color: '#1976d2', fontWeight: 600 }}>
                {props.title}
              </Typography>
            )}
            {children.map((child, i) => renderComponent(child, i))}
          </Box>
        )

      case 'text':
        return (
          <Typography
            key={key}
            variant="body2"
            sx={{ mb: 0.5, color: '#333' }}
          >
            {props.content || children}
          </Typography>
        )

      case 'button':
        return (
          <Button
            key={key}
            variant="contained"
            onClick={() => props.onclick && handleButtonClick(props.onclick)}
            disabled={!props.onclick}
            sx={{
              mr: 1,
              mb: 1,
              textTransform: 'none',
            }}
          >
            {props.label || children}
          </Button>
        )

      case 'input':
        const inputName = props.name || `input_${key}`
        const inputValue = inputs[inputName] || props.value || ''
        return (
          <TextField
            key={key}
            size="small"
            placeholder={props.placeholder}
            value={inputValue}
            onChange={(e) => handleInputChange(inputName, e.target.value)}
            sx={{
              mb: 1,
              mr: 1,
              width: props.width || '200px',
            }}
          />
        )

      case 'container':
        const isHorizontal = props.layout === 'horizontal'
        return (
          <Box
            key={key}
            sx={{
              display: 'flex',
              flexDirection: isHorizontal ? 'row' : 'column',
              gap: isHorizontal ? 2 : 1,
              mb: 1,
              flexWrap: 'wrap',
            }}
          >
            {children.map((child, i) => renderComponent(child, i))}
          </Box>
        )

      default:
        return (
          <Box key={key} sx={{ mb: 1 }}>
            <Typography variant="caption" color="text.secondary">
              &lt;{type}&gt;
            </Typography>
            {children.map((child, i) => renderComponent(child, i))}
          </Box>
        )
    }
  }

  const component = parseXML(xmlContent)

  if (!component) {
    return (
      <Box sx={{ p: 2, color: 'error.main' }}>
        <Typography variant="body2" sx={{ mb: 1, fontWeight: 600 }}>
          Failed to parse UI output
        </Typography>
        <Typography variant="caption" sx={{ display: 'block', mb: 1 }}>
          The output doesn't appear to be valid XML. Check the console for details.
        </Typography>
        <Typography variant="caption" sx={{ display: 'block', fontFamily: 'monospace', bgcolor: '#f5f5f5', p: 1, borderRadius: 1, color: '#333' }}>
          {xmlContent.substring(0, 200)}...
        </Typography>
      </Box>
    )
  }

  // Get the most recent result
  const latestResult = Object.values(results).slice(-1)[0]

  return (
    <Box sx={{ p: 2, bgcolor: 'white', borderRadius: 1, minHeight: 200 }}>
      {renderComponent(component, 0)}

      {/* Display latest result */}
      {latestResult && (
        <Box
          sx={{
            mt: 2,
            p: 2,
            bgcolor: '#e3f2fd',
            borderRadius: 1,
            border: '2px solid #2196f3',
          }}
        >
          <Typography variant="subtitle2" sx={{ fontWeight: 600, color: '#1976d2', mb: 0.5 }}>
            Result:
          </Typography>
          <Typography variant="body1" sx={{ fontFamily: 'monospace', color: '#333' }}>
            {latestResult}
          </Typography>
        </Box>
      )}
    </Box>
  )
}

export default UIPreview
