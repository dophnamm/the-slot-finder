import type { ComponentProps, ReactNode } from 'react'

/** A table row's own props (e.g. draggable and its drag events), except the click that opens the page. */
export interface LinkRowProps extends Omit<ComponentProps<'tr'>, 'onClick' | 'children'> {
  /** The detail page the row opens. */
  to: string
  children: ReactNode
}
