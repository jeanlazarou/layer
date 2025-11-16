import React from 'react'
import { Box, Typography } from '@mui/material'

interface UIPreviewProps {
  xmlContent: string
}

interface UIComponent {
  type: string
  props: Record<string, string>
  children: (UIComponent | string)[]
}

const UIPreview: React.FC<UIPreviewProps> = ({ xmlContent }) => {
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
          <Box
            key={key}
            component="button"
            sx={{
              px: 2,
              py: 1,
              mr: 1,
              mb: 1,
              bgcolor: '#1976d2',
              color: 'white',
              border: 'none',
              borderRadius: 1,
              cursor: 'pointer',
              fontSize: '0.875rem',
              fontWeight: 500,
              '&:hover': {
                bgcolor: '#1565c0',
              },
            }}
          >
            {props.label || children}
          </Box>
        )

      case 'input':
        return (
          <Box
            key={key}
            component="input"
            sx={{
              px: 1.5,
              py: 0.75,
              mb: 1,
              border: '1px solid #ccc',
              borderRadius: 1,
              fontSize: '0.875rem',
              width: '200px',
            }}
            placeholder={props.placeholder}
            value={props.value || ''}
            readOnly
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

  return (
    <Box sx={{ p: 2, bgcolor: 'white', borderRadius: 1, minHeight: 200 }}>
      {renderComponent(component, 0)}
    </Box>
  )
}

export default UIPreview
