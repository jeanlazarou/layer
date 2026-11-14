import React, { useEffect, useRef, useState } from 'react'
import { Tooltip } from '@mui/material'

interface MenuTooltipProps {
  title: string
  /** True while the control's own menu/dropdown is open. */
  menuOpen: boolean
  children: React.ReactElement
}

/**
 * A tooltip for a control that opens a menu. Both pop up below the control,
 * so an open tooltip covers the first menu item — and when the menu is a
 * React child of the control (a Select's dropdown, rendered in a portal),
 * hovering the menu items bubbles up to the tooltip and reopens it. This one
 * stays hidden while the menu is open, ignores hovers until it closes, and
 * ignores the focus MUI gives back to the control when the menu closes — it
 * reappears on the next real hover (or keyboard focus) of the control.
 */
const MenuTooltip: React.FC<MenuTooltipProps> = ({ title, menuOpen, children }) => {
  const [hovered, setHovered] = useState(false)
  const closedAt = useRef(0)

  // Reset when the menu opens or closes: hovers seen while it was open came
  // from the menu, not from the control.
  useEffect(() => {
    setHovered(false)
    if (!menuOpen) closedAt.current = Date.now()
  }, [menuOpen])

  return (
    <Tooltip
      title={title}
      open={hovered && !menuOpen}
      onOpen={event => {
        if (menuOpen) return
        // Focus restored to the control as the menu closes is not a request
        // for the tooltip.
        if (event.type === 'focus' && Date.now() - closedAt.current < 1000) return
        setHovered(true)
      }}
      onClose={() => setHovered(false)}
    >
      {children}
    </Tooltip>
  )
}

export default MenuTooltip
