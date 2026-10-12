import { Link } from 'react-router'
import { cn } from '@/utils/functions'
import { TEXT_LINK_CLASS } from './constants'
import type { TextLinkProps } from './types'

/** Text that opens a page: an app route, or another site in a new tab. */
export function TextLink({ to, external = false, className, children }: TextLinkProps) {
  if (external) {
    return (
      <a className={cn(TEXT_LINK_CLASS, className)} href={to} target="_blank" rel="noreferrer">
        {children}
      </a>
    )
  }
  return (
    <Link className={cn(TEXT_LINK_CLASS, className)} to={to}>
      {children}
    </Link>
  )
}
