import type { ReactNode } from 'react'
import type { TERM_DEFINITIONS } from './constants'

export type TermName = keyof typeof TERM_DEFINITIONS

export interface TermProps {
  /** The word's definition shows on hover or focus; without one the children show as they are. */
  term?: TermName
  children: ReactNode
}
