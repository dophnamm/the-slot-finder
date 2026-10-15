import type { TermName } from '@/shared/components/term'

export interface FigureProps {
  /** Already translated. */
  label: string
  /** Shows the word's definition on hover. */
  term?: TermName
  hours: number
  warning?: boolean
}
