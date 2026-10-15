import type { ReactElement } from 'react'

export interface IconHintProps {
  /** Already translated; shows on a long hover or keyboard focus. */
  label: string
  /** The icon button the label describes. */
  children: ReactElement
}
