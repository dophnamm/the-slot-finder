import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'

export interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description?: string
  /** What the user can do about it, such as a retry button. */
  action?: ReactNode
  /** Draws a bordered frame, for an empty state that stands in for a table. */
  framed?: boolean
}
