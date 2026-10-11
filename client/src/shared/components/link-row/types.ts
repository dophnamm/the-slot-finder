import type { ReactNode } from 'react'

export interface LinkRowProps {
  /** The detail page the row opens. */
  to: string
  children: ReactNode
}
