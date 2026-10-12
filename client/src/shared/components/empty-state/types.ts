import type { LucideIcon } from 'lucide-react'

export interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description?: string
  /** Draws a bordered frame, for an empty state that stands in for a table. */
  framed?: boolean
}
