import type { ReactNode } from 'react'

export interface TextLinkProps {
  /** An app route, or a full URL when `external`. */
  to: string
  /** Opens another site in a new tab. */
  external?: boolean
  className?: string
  children: ReactNode
}
