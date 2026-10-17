import type { PlainKey } from '@/shared/i18n'

/** One problem of an imported file: the line, the column ('' for the whole row) and its code. */
export interface RowProblem {
  line: number
  column: string
  code: string
}

export interface ProblemTableProps {
  problems: RowProblem[]
  /** What each code means; a code without a label shows as is. */
  labels: Partial<Record<string, PlainKey>>
}
