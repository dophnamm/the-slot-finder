import type { ReactNode } from 'react'

export type SectionVariant = 'section' | 'page'

export interface SectionProps {
  /** The heading; may hold badges next to the text. */
  title: ReactNode
  description?: ReactNode
  /** Buttons on the right of the heading. */
  action?: ReactNode
  /** The body. Without children or a footer, only the heading is shown. */
  children?: ReactNode
  /** A row under the body, e.g. the pager. */
  footer?: ReactNode
  /** Pads the body for text, numbers or a form. Leave it off for a table, which runs from edge to edge. */
  card?: boolean
  /** `page` is the page's own header (the h1), with any body in a card under it. */
  variant?: SectionVariant
  className?: string
  contentClassName?: string
  /** The body's data is loading again (another page, filter or month): a thin bar runs along its top edge. */
  refreshing?: boolean
}
