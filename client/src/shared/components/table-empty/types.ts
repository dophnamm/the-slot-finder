import type { EmptyStateProps } from '@/shared/components/empty-state'

export interface TableEmptyProps extends EmptyStateProps {
  /** How many columns the table has, so the message spans them all. */
  columns: number
}
