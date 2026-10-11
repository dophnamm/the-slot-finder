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
  /** Puts the body in a card. Leave it off for tables: a bordered table is its own frame. */
  card?: boolean
  /** `page` is the larger heading at the top of a page. */
  variant?: SectionVariant
  className?: string
  contentClassName?: string
}
