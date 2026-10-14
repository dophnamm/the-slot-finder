import type { MouseEvent } from 'react'
import { useHref, useNavigate } from 'react-router'
import { TableRow } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import type { LinkRowProps } from './types'
import { isInsideControl, isSelectingText } from './utils'

/**
 * A table row that opens its detail page when clicked anywhere outside its own links and buttons.
 * Ctrl or ⌘ opens it in a new tab. Keep a link on the row's name for keyboard users.
 */
export function LinkRow({ to, children, className, ...props }: LinkRowProps) {
  const navigate = useNavigate()
  const href = useHref(to)

  const handleClick = (event: MouseEvent<HTMLTableRowElement>) => {
    if (isInsideControl(event.target) || isSelectingText()) return
    if (event.metaKey || event.ctrlKey) {
      window.open(href, '_blank', 'noopener')
      return
    }
    navigate(to)
  }

  return (
    <TableRow className={cn('cursor-pointer', className)} onClick={handleClick} {...props}>
      {children}
    </TableRow>
  )
}
